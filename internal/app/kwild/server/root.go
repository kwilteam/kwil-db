package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"path/filepath"
	"sync"
	"time"

	// kwil-db
	txpb "github.com/kwilteam/kwil-db/api/protobuf/tx/v1"
	"github.com/kwilteam/kwil-db/internal/app/kwild/config"
	"github.com/kwilteam/kwil-db/internal/controller/grpc/healthsvc/v0"
	txSvc "github.com/kwilteam/kwil-db/internal/controller/grpc/txsvc/v1"
	"github.com/kwilteam/kwil-db/internal/pkg/healthcheck"
	simple_checker "github.com/kwilteam/kwil-db/internal/pkg/healthcheck/simple-checker"
	"github.com/kwilteam/kwil-db/pkg/abci"
	"github.com/kwilteam/kwil-db/pkg/abci/cometbft"
	"github.com/kwilteam/kwil-db/pkg/balances"
	"github.com/kwilteam/kwil-db/pkg/engine"
	"github.com/kwilteam/kwil-db/pkg/grpc/gateway"
	"github.com/kwilteam/kwil-db/pkg/grpc/gateway/middleware/cors"
	grpc "github.com/kwilteam/kwil-db/pkg/grpc/server"
	"github.com/kwilteam/kwil-db/pkg/kv/atomic"
	"github.com/kwilteam/kwil-db/pkg/kv/badger"
	"github.com/kwilteam/kwil-db/pkg/log"
	"github.com/kwilteam/kwil-db/pkg/modules/datasets"
	"github.com/kwilteam/kwil-db/pkg/modules/snapshots"
	"github.com/kwilteam/kwil-db/pkg/modules/validators"
	"github.com/kwilteam/kwil-db/pkg/sessions"
	"github.com/kwilteam/kwil-db/pkg/sessions/wal"
	snapshotPkg "github.com/kwilteam/kwil-db/pkg/snapshots"
	"github.com/kwilteam/kwil-db/pkg/sql"
	vmgr "github.com/kwilteam/kwil-db/pkg/validators"

	abciTypes "github.com/cometbft/cometbft/abci/types"
	cmtlocal "github.com/cometbft/cometbft/rpc/client/local"

	"google.golang.org/grpc/health/grpc_health_v1"
)

// BuildKwildServer builds the kwild server
func BuildKwildServer(ctx context.Context) (svr *Server, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic while building kwild: %v", r)
		}
	}()

	cfg, err := config.LoadKwildConfig()
	if err != nil {
		return nil, err
	}

	logger := log.New(cfg.Log)
	logger = *logger.Named("kwild")

	deps := &coreDependencies{
		ctx:    ctx,
		cfg:    cfg,
		log:    logger,
		opener: newSqliteOpener(cfg.SqliteFilePath),
	}

	closers := &closeFuncs{
		closers: make([]func() error, 0),
	}

	return buildServer(deps, closers), nil
}

func buildServer(d *coreDependencies, closers *closeFuncs) *Server {
	// atomic committer
	ac := buildAtomicCommitter(d)
	closers.addCloser(ac.Close)

	// engine
	e := buildEngine(d, ac)
	closers.addCloser(e.Close)

	// account store
	accs := buildAccountRepository(d, closers, ac)

	// datasets module
	datasetsModule := buildDatasetsModule(d, e, accs)

	// validator updater and store
	vstore := buildValidatorManager(d, closers, ac)

	// validator module
	validatorModule := buildValidatorModule(d, accs, vstore)

	snapshotModule := buildSnapshotModule(d)

	bootstrapperModule := buildBootstrapModule(d)

	abciApp := buildAbci(d, closers, datasetsModule, validatorModule, ac)

	cometBftNode := buildCometNode(d, closers, abciApp)

	cometBftClient := buildCometBftClient(cometBftNode)

	// tx service
	txsvc := buildTxSvc(d, datasetsModule, accs, &wrappedCometBFTClient{cometBftClient})

	// grpc server
	grpcServer := buildGrpcServer(d, txsvc)

	return &Server{
		grpcServer:   grpcServer,
		gateway:      buildGatewayServer(d),
		cometBftNode: cometBftNode,
		log:          *d.log.Named("kwild-server"),
		closers:      closers,
		cfg:          d.cfg,
	}
}

// coreDependencies holds dependencies that are widely used
type coreDependencies struct {
	ctx    context.Context
	cfg    *config.KwildConfig
	log    log.Logger
	opener sql.Opener
}

// closeFuncs holds a list of closers
// it is used to close all resources on shutdown
type closeFuncs struct {
	closers []func() error
}

func (c *closeFuncs) addCloser(f func() error) {
	c.closers = append(c.closers, f)
}

// closeAll concurrently closes all closers
func (c *closeFuncs) closeAll() error {
	errs := make([]error, 0)
	errCh := make(chan error, len(c.closers))
	wg := sync.WaitGroup{}

	for _, f := range c.closers {
		wg.Add(1)
		go func(f func() error) {
			err := f()
			if err != nil {
				errCh <- err
			}
			wg.Done()
		}(f)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

func buildAbci(d *coreDependencies, closer *closeFuncs, datasetsModule abci.DatasetsModule, validatorModule abci.ValidatorModule,
	atomicCommitter *sessions.AtomicCommitter) *abci.AbciApp {
	badgerKv, err := badger.NewBadgerDB(d.ctx, filepath.Join(d.cfg.RootDir, "abci/info"), &badger.Options{
		GuaranteeFSync: true,
		Logger:         *d.log.Named("abci-kv-store"),
	})
	if err != nil {
		failBuild(err, "failed to open badger")
	}
	closer.addCloser(badgerKv.Close)

	atomicKv, err := atomic.NewAtomicKV(badgerKv)
	if err != nil {
		failBuild(err, "failed to open atomic kv")
	}

	err = atomicCommitter.Register(d.ctx, "blockchain_kv", atomicKv)
	if err != nil {
		failBuild(err, "failed to register atomic kv")
	}

	return abci.NewAbciApp(
		datasetsModule,
		validatorModule,
		atomicKv,
		atomicCommitter,
		snapshotter,
		bootstrapper,
		abci.WithLogger(*d.log.Named("abci")),
	)
}

func buildTxSvc(d *coreDependencies, txsvc txSvc.EngineReader, accs txSvc.AccountReader,
	cometBftClient txSvc.BlockchainBroadcaster) *txSvc.Service {
	return txSvc.NewService(txsvc, accs, cometBftClient,
		txSvc.WithLogger(*d.log.Named("tx-service")),
	)
}

func buildDatasetsModule(d *coreDependencies, eng datasets.Engine, accs datasets.AccountStore) *datasets.DatasetModule {
	feeMultiplier := 1
	if d.cfg.WithoutGasCosts {
		feeMultiplier = 0
	}

	return datasets.NewDatasetModule(eng, accs,
		datasets.WithLogger(*d.log.Named("dataset-module")),
		datasets.WithFeeMultiplier(int64(feeMultiplier)),
	)
}

func buildEngine(d *coreDependencies, a *sessions.AtomicCommitter) *engine.Engine {
	extensions, err := connectExtensions(d.ctx, d.cfg.ExtensionEndpoints)
	if err != nil {
		failBuild(err, "failed to connect to extensions")
	}

	sqlCommitRegister := &sqlCommittableRegister{
		commiter: a,
		log:      *d.log.Named("sqlite-committable"),
	}

	e, err := engine.Open(d.ctx, d.opener,
		sqlCommitRegister,
		engine.WithLogger(*d.log.Named("engine")),
		engine.WithExtensions(adaptExtensions(extensions)),
	)
	if err != nil {
		failBuild(err, "failed to open engine")
	}

	return e
}

func buildAccountRepository(d *coreDependencies, closer *closeFuncs, ac *sessions.AtomicCommitter) *balances.AccountStore {
	db, err := d.opener.Open("accounts_db", *d.log.Named("account-store"))
	if err != nil {
		failBuild(err, "failed to open accounts db")
	}
	closer.addCloser(db.Close)

	err = registerSQL(d.ctx, ac, db, "accounts_db", d.log)
	if err != nil {
		failBuild(err, "failed to register accounts db")
	}

	b, err := balances.NewAccountStore(d.ctx, db,
		balances.WithLogger(*d.log.Named("accountStore")),
		balances.WithNonces(!d.cfg.WithoutNonces),
		balances.WithGasCosts(!d.cfg.WithoutGasCosts),
	)
	if err != nil {
		failBuild(err, "failed to build account store")
	}

	return b
}

func buildValidatorManager(d *coreDependencies, closer *closeFuncs, ac *sessions.AtomicCommitter) *vmgr.ValidatorMgr {
	db, err := d.opener.Open("validator_db", *d.log.Named("validator-store"))
	if err != nil {
		failBuild(err, "failed to open validator db")
	}
	closer.addCloser(db.Close)

	err = registerSQL(d.ctx, ac, db, "validator_db", d.log)
	if err != nil {
		failBuild(err, "failed to register validator db")
	}

	v, err := vmgr.NewValidatorMgr(d.ctx, db,
		vmgr.WithLogger(*d.log.Named("validatorStore")),
	)
	if err != nil {
		failBuild(err, "failed to build validator store")
	}

	return v
}

func buildValidatorModule(d *coreDependencies, accs datasets.AccountStore,
	vals validators.ValidatorMgr) *validators.ValidatorModule {
	return validators.NewValidatorModule(vals, accs, abci.Addresser,
		validators.WithLogger(*d.log.Named("validator-module")))
}

func buildSnapshotModule(d *coreDependencies) *snapshots.SnapshotStore {
	if !d.cfg.SnapshotConfig.Enabled {
		return nil
	}

	return snapshots.NewSnapshotStore(d.cfg.SqliteFilePath,
		d.cfg.SnapshotConfig.SnapshotDir,
		d.cfg.SnapshotConfig.RecurringHeight,
		d.cfg.SnapshotConfig.MaxSnapshots,
		snapshots.WithLogger(*d.log.Named("snapshotStore")),
	)
}

func buildBootstrapModule(d *coreDependencies) *snapshotPkg.Bootstrapper {
	bootstrapper, err := snapshotPkg.NewBootstrapper(d.cfg.SqliteFilePath)
	if err != nil {
		failBuild(err, "Bootstrap module initialization failure")
	}
	return bootstrapper
}

func buildGrpcServer(d *coreDependencies, txsvc txpb.TxServiceServer) *grpc.Server {
	lis, err := net.Listen("tcp", d.cfg.GrpcListenAddress)
	if err != nil {
		failBuild(err, "failed to build grpc server")
	}

	grpcServer := grpc.New(*d.log.Named("grpc-server"), lis)
	txpb.RegisterTxServiceServer(grpcServer, txsvc)
	grpc_health_v1.RegisterHealthServer(grpcServer, buildHealthSvc(d))

	return grpcServer
}

func buildHealthSvc(d *coreDependencies) *healthsvc.Server {
	// health service
	registrar := healthcheck.NewRegistrar(*d.log.Named("healthcheck"))
	registrar.RegisterAsyncCheck(10*time.Second, 3*time.Second, healthcheck.Check{
		Name: "dummy",
		Check: func(ctx context.Context) error {
			// error make this check fail, nil will make it succeed
			return nil
		},
	})
	ck := registrar.BuildChecker(simple_checker.New(d.log))
	return healthsvc.NewServer(ck)
}

func buildGatewayServer(d *coreDependencies) *gateway.GatewayServer {
	gw, err := gateway.NewGateway(d.ctx, d.cfg.HttpListenAddress,
		gateway.WithLogger(d.log),
		gateway.WithMiddleware(cors.MCors([]string{})),
		gateway.WithGrpcService(d.cfg.GrpcListenAddress, txpb.RegisterTxServiceHandlerFromEndpoint),
	)
	if err != nil {
		failBuild(err, "failed to build gateway server")
	}

	return gw
}

func buildCometBftClient(cometBftNode *cometbft.CometBftNode) *cmtlocal.Local {
	return cmtlocal.New(cometBftNode.Node)
}

func buildCometNode(d *coreDependencies, closer *closeFuncs, abciApp abciTypes.Application) *cometbft.CometBftNode {
	// TODO: a lot of the filepaths, as well as cometbft logging level, are hardcoded.  This should be cleaned up with a config

	// for now, I'm just using a KV store for my atomic commit.  This probably is not ideal; a file may be better
	// I'm simply using this because we know it fsyncs the data to disk
	db, err := badger.NewBadgerDB(d.ctx, filepath.Join(d.cfg.RootDir, "signing"), &badger.Options{
		GuaranteeFSync: true,
		Logger:         *d.log.Named("private-validator-signature-store"),
	})
	if err != nil {
		failBuild(err, "failed to build comet node")
	}
	closer.addCloser(db.Close)

	readWriter := &atomicReadWriter{
		kv:  db,
		key: []byte("az"), // any key here will work
	}

	node, err := cometbft.NewCometBftNode(abciApp, d.cfg.PrivateKey.Bytes(), readWriter, filepath.Join(d.cfg.RootDir, "abci"), "debug")
	if err != nil {
		failBuild(err, "failed to build comet node")
	}

	return node
}

func buildAtomicCommitter(d *coreDependencies) *sessions.AtomicCommitter {
	twoPCWal, err := wal.OpenWal(filepath.Join(d.cfg.RootDir, "application/wal"))
	if err != nil {
		failBuild(err, "failed to open 2pc wal")
	}

	// we are actually registering all committables ad-hoc, so we can pass nil here
	return sessions.NewAtomicCommitter(d.ctx, nil, twoPCWal, sessions.WithLogger(*d.log.Named("atomic-committer")))
}

func failBuild(err error, msg string) {
	if err != nil {
		panic(fmt.Sprintf("%s: %s", msg, err.Error()))
	}
}
