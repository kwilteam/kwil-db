package server

import (
	"context"
	"fmt"
	"net"
	"os"
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
	"github.com/kwilteam/kwil-db/pkg/kv/badger"
	"github.com/kwilteam/kwil-db/pkg/log"
	"github.com/kwilteam/kwil-db/pkg/modules/datasets"
	"github.com/kwilteam/kwil-db/pkg/modules/snapshots"
	"github.com/kwilteam/kwil-db/pkg/modules/validators"
	snapshotPkg "github.com/kwilteam/kwil-db/pkg/snapshots"
	"github.com/kwilteam/kwil-db/pkg/sql"
	vmgr "github.com/kwilteam/kwil-db/pkg/validators"

	abciTypes "github.com/cometbft/cometbft/abci/types"
	cmtcfg "github.com/cometbft/cometbft/config"
	cmtflags "github.com/cometbft/cometbft/libs/cli/flags"
	cmtlog "github.com/cometbft/cometbft/libs/log"
	nm "github.com/cometbft/cometbft/node"
	"github.com/cometbft/cometbft/p2p"
	"github.com/cometbft/cometbft/privval"
	"github.com/cometbft/cometbft/proxy"
	cmtlocal "github.com/cometbft/cometbft/rpc/client/local"

	"github.com/spf13/viper"
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

	return buildServer(deps), nil
}

type nodeWrapper struct {
	n *nm.Node
}

func (nw *nodeWrapper) Start() error {
	return nw.n.Start()
}

func (nw *nodeWrapper) Stop() error {
	if err := nw.n.Stop(); err != nil {
		return err
	}
	nw.n.Wait()
	return nil
}

var _ (startStopper) = (*nodeWrapper)(nil)

func buildServer(d *coreDependencies) *Server {
	// engine
	e := buildEngine(d)

	// account store
	accs := buildAccountRepository(d)

	// datasets module
	datasetsModule := buildDatasetsModule(d, e, accs)

	// validator updater and store
	vstore := buildValidatorManager(d)

	// validator module
	validatorModule := buildValidatorModule(d, accs, vstore)

	snapshotModule := buildSnapshotModule(d)

	bootstrapperModule := buildBootstrapModule(d)

	abciApp := buildAbci(d, datasetsModule, validatorModule, nil, snapshotModule, bootstrapperModule)

	cometBftNode, err := newCometNode(abciApp, d.cfg)
	if err != nil {
		failBuild(err, "failed to create cometbft node")
	}

	cometBftClient := buildCometBftClient(cometBftNode)

	// tx service
	txsvc := buildTxSvc(d, datasetsModule, accs, &wrappedCometBFTClient{cometBftClient})

	// grpc server
	grpcServer := buildGrpcServer(d, txsvc)

	return &Server{
		grpcServer:   grpcServer,
		gateway:      buildGatewayServer(d),
		cometBftNode: &nodeWrapper{cometBftNode},
	}
}

// coreDependencies holds dependencies that are widely used
type coreDependencies struct {
	ctx    context.Context
	cfg    *config.KwildConfig
	log    log.Logger
	opener sql.Opener
}

func buildAbci(d *coreDependencies, datasetsModule abci.DatasetsModule, validatorModule abci.ValidatorModule,
	atomicCommitter abci.AtomicCommitter, snapshotter abci.SnapshotModule, bootstrapper abci.DBBootstrapModule) *abci.AbciApp {
	return abci.NewAbciApp(
		datasetsModule,
		validatorModule,
		nil, // TODO: add kv store
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

func buildEngine(d *coreDependencies) *engine.Engine {
	extensions, err := connectExtensions(d.ctx, d.cfg.ExtensionEndpoints)
	if err != nil {
		failBuild(err, "failed to connect to extensions")
	}

	e, err := engine.Open(d.ctx, d.opener,
		engine.WithLogger(*d.log.Named("engine")),
		engine.WithExtensions(adaptExtensions(extensions)),
	)
	if err != nil {
		failBuild(err, "failed to open engine")
	}

	return e
}

func buildAccountRepository(d *coreDependencies) *balances.AccountStore {
	db, err := d.opener.Open("accounts_db", *d.log.Named("account-store"))
	if err != nil {
		failBuild(err, "failed to open accounts db")
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

func buildValidatorManager(d *coreDependencies) *vmgr.ValidatorMgr {
	db, err := d.opener.Open("validator_db", *d.log.Named("validator-store"))
	if err != nil {
		failBuild(err, "failed to open validator db")
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

func buildCometBftClient(cometBftNode *nm.Node) *cmtlocal.Local {
	return cmtlocal.New(cometBftNode)
}

func buildCometNode(d *coreDependencies, abciApp abciTypes.Application) *cometbft.CometBftNode {
	// TODO: a lot of the filepaths and logging here are hardcoded.  This should be cleaned up
	// with a config

	// for now, I'm just using a KV store for my atomic commit.  This probably is not ideal; a file may be better
	// I'm simply using this because we know it fsyncs the data to disk
	db, err := badger.NewBadgerDB("abci/signing", &badger.Options{
		GuaranteeFSync: true,
	})
	if err != nil {
		failBuild(err, "failed to build comet node")
	}

	readWriter := &atomicReadWriter{
		kv:  db,
		key: []byte("az"), // any key here will work
	}

	node, err := cometbft.NewCometBftNode(abciApp, d.cfg.PrivateKey.Bytes(), readWriter, "abci/data", "debug")
	if err != nil {
		failBuild(err, "failed to build comet node")
	}

	return node
}

// TODO: clean this up --> @jchappelow
// it seems some of this should be handled in ABCI package if we do not provide it as a package
func newCometNode(app *abci.AbciApp, cfg *config.KwildConfig) (*nm.Node, error) {
	config := cmtcfg.DefaultConfig()
	CometHomeDir := os.Getenv("COMET_BFT_HOME")
	fmt.Printf("Home Directory: %v", CometHomeDir)
	config.SetRoot(CometHomeDir)

	viper.SetConfigFile(fmt.Sprintf("%s/%s", CometHomeDir, "config/config.toml"))
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("reading config: %v", err)
	}
	if err := viper.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("decoding config: %v", err)
	}
	if err := config.ValidateBasic(); err != nil {
		return nil, fmt.Errorf("invalid configuration data: %v", err)
	}

	pv := privval.LoadFilePV(
		config.PrivValidatorKeyFile(),
		config.PrivValidatorStateFile(),
	)
	fmt.Println("PrivateKey: ", pv.Key.PrivKey)

	fmt.Println("PrivateValidator: ", string(pv.Key.PrivKey.Bytes()))
	nodeKey, err := p2p.LoadNodeKey(config.NodeKeyFile())
	if err != nil {
		return nil, fmt.Errorf("failed to load node's key: %v", err)
	}

	logger := cmtlog.NewTMLogger(cmtlog.NewSyncWriter(os.Stdout))
	logger, err = cmtflags.ParseLogLevel(config.LogLevel, logger, cfg.Log.Level)
	if err != nil {
		return nil, fmt.Errorf("failed to parse log level: %v", err)
	}

	// TODO: Move this to Application init and maybe use some kind of kvstore to store the validators info
	fmt.Println("Pre RPC Config: ", config.RPC.ListenAddress, " ", cfg.BcRpcUrl)
	cfg.BcRpcUrl = config.RPC.ListenAddress
	fmt.Println("Post RPC Config: ", config.RPC.ListenAddress, " ", cfg.BcRpcUrl)

	node, err := nm.NewNode(
		config,
		pv,
		nodeKey,
		proxy.NewLocalClientCreator(app), // TODO: NewUnsyncedLocalClientCreator(app) is other option which doesn't take a lock on all the connections to the app
		nm.DefaultGenesisDocProviderFunc(config),
		nm.DefaultDBProvider,
		nm.DefaultMetricsProvider(config.Instrumentation),
		logger,
	)

	if err != nil {
		return nil, fmt.Errorf("creating node: %v", err)
	}

	return node, nil
}

func failBuild(err error, msg string) {
	if err != nil {
		panic(fmt.Sprintf("%s: %s", msg, err.Error()))
	}
}
