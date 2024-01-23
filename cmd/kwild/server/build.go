package server

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/kwilteam/kwil-db/cmd/kwild/config"
	"github.com/kwilteam/kwil-db/internal/abci"
	"github.com/kwilteam/kwil-db/internal/abci/cometbft"
	"github.com/kwilteam/kwil-db/internal/abci/snapshots"
	"github.com/kwilteam/kwil-db/internal/accounts"
	"github.com/kwilteam/kwil-db/internal/engine/execution"
	"github.com/kwilteam/kwil-db/internal/kv/badger"
	acctmod "github.com/kwilteam/kwil-db/internal/modules/accounts"
	"github.com/kwilteam/kwil-db/internal/modules/datasets"
	"github.com/kwilteam/kwil-db/internal/modules/validators"
	admSvc "github.com/kwilteam/kwil-db/internal/services/grpc/admin/v0"
	functionSvc "github.com/kwilteam/kwil-db/internal/services/grpc/function/v0"
	"github.com/kwilteam/kwil-db/internal/services/grpc/healthsvc/v0"
	txSvc "github.com/kwilteam/kwil-db/internal/services/grpc/txsvc/v1"
	gateway "github.com/kwilteam/kwil-db/internal/services/grpc_gateway"
	"github.com/kwilteam/kwil-db/internal/services/grpc_gateway/middleware/cors"
	kwilgrpc "github.com/kwilteam/kwil-db/internal/services/grpc_server"
	healthcheck "github.com/kwilteam/kwil-db/internal/services/health"
	simple_checker "github.com/kwilteam/kwil-db/internal/services/health/simple-checker"
	"github.com/kwilteam/kwil-db/internal/sessions"
	"github.com/kwilteam/kwil-db/internal/sessions/committable"
	"github.com/kwilteam/kwil-db/internal/sql/adapter"
	"github.com/kwilteam/kwil-db/internal/sql/registry"
	"github.com/kwilteam/kwil-db/internal/sql/sqlite"
	vmgr "github.com/kwilteam/kwil-db/internal/validators"

	"github.com/kwilteam/kwil-db/core/crypto"
	"github.com/kwilteam/kwil-db/core/crypto/auth"
	"github.com/kwilteam/kwil-db/core/log"
	admpb "github.com/kwilteam/kwil-db/core/rpc/protobuf/admin/v0"
	functionpb "github.com/kwilteam/kwil-db/core/rpc/protobuf/function/v0"
	txpb "github.com/kwilteam/kwil-db/core/rpc/protobuf/tx/v1"
	"github.com/kwilteam/kwil-db/core/rpc/transport"
	"github.com/kwilteam/kwil-db/core/utils/url"

	abciTypes "github.com/cometbft/cometbft/abci/types"
	cmtEd "github.com/cometbft/cometbft/crypto/ed25519"
	cmtlocal "github.com/cometbft/cometbft/rpc/client/local"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func buildServer(d *coreDependencies, closers *closeFuncs) *Server {
	// atomic committer
	ac := buildCommitter(d, closers)

	// engine
	e := buildEngine(d, closers, ac)

	// account store
	accs := buildAccountRepository(d, closers, ac)
	accsMod := buildAccountsModule(accs) // abci accounts module

	// datasets module
	datasetsModule := buildDatasetsModule(d, e, accs)

	// validator updater and store
	vstore := buildValidatorManager(d, closers, ac)

	// validator module
	validatorModule := buildValidatorModule(d, accs, vstore)

	snapshotModule := buildSnapshotter(d)

	bootstrapperModule := buildBootstrapper(d)

	abciApp := buildAbci(d, closers, accsMod, datasetsModule, validatorModule,
		ac, snapshotModule, bootstrapperModule)

	cometBftNode := buildCometNode(d, closers, abciApp)

	cometBftClient := buildCometBftClient(cometBftNode)

	// tx service and grpc server
	txsvc := buildTxSvc(d, datasetsModule, accsMod, vstore,
		&wrappedCometBFTClient{cometBftClient}, abciApp)
	grpcServer := buildGrpcServer(d, txsvc)

	// admin service and server
	admsvc := buildAdminSvc(d, &wrappedCometBFTClient{cometBftClient}, abciApp, vstore)
	adminTCPServer := buildAdminService(d, closers, admsvc, txsvc)

	return &Server{
		grpcServer:     grpcServer,
		adminTPCServer: adminTCPServer,
		gateway:        buildGatewayServer(d),
		cometBftNode:   cometBftNode,
		log:            *d.log.Named("server"),
		closers:        closers,
		cfg:            d.cfg,
	}
}

// db / committable names, to prevent breaking changes
const (
	accountsDBName  = "accounts"
	validatorDBName = "validators"
	engineName      = "engine"
)

// coreDependencies holds dependencies that are widely used
type coreDependencies struct {
	ctx        context.Context
	autogen    bool
	cfg        *config.KwildConfig
	genesisCfg *config.GenesisConfig
	privKey    cmtEd.PrivKey
	log        log.Logger
	opener     func(ctx context.Context, dbName string, persistentReaders, maximumReaders int, create bool) (*sqlite.Pool, error)
	keypair    *tls.Certificate
}

// closeFuncs holds a list of closers
// it is used to close all resources on shutdown
type closeFuncs struct {
	closers []func() error
}

func (c *closeFuncs) addCloser(f func() error) {
	c.closers = append([]func() error{f}, c.closers...) // slices.Insert(c.closers, 0, f)
}

// closeAll closes all closers
func (c *closeFuncs) closeAll() error {
	var err error
	for _, closer := range c.closers {
		err = errors.Join(closer())
	}

	return err
}

func buildAbci(d *coreDependencies, closer *closeFuncs, accountsModule abci.AccountsModule, datasetsModule abci.DatasetsModule, validatorModule abci.ValidatorModule,
	committer *sessions.MultiCommitter, snapshotter *snapshots.SnapshotStore, bootstrapper *snapshots.Bootstrapper) *abci.AbciApp {
	badgerPath := filepath.Join(d.cfg.RootDir, abciDirName, config.ABCIInfoSubDirName)
	badgerKv, err := badger.NewBadgerDB(d.ctx, badgerPath, &badger.Options{
		GuaranteeFSync: true,
		Logger:         *d.log.Named("abci-kv-store"),
	})
	d.log.Info(fmt.Sprintf("created ABCI kv DB in %v", badgerPath))
	if err != nil {
		failBuild(err, "failed to open badger")
	}
	closer.addCloser(badgerKv.Close)

	var sh abci.SnapshotModule
	if snapshotter != nil {
		sh = snapshotter
	}

	cfg := &abci.AbciConfig{
		GenesisAppHash:     d.genesisCfg.ComputeGenesisHash(),
		ChainID:            d.genesisCfg.ChainID,
		ApplicationVersion: d.genesisCfg.ConsensusParams.Version.App,
		GenesisAllocs:      d.genesisCfg.Alloc,
	}
	return abci.NewAbciApp(cfg,
		accountsModule,
		datasetsModule,
		validatorModule,
		badgerKv,
		committer,
		sh,
		bootstrapper,
		abci.WithLogger(*d.log.Named("abci")),
	)
}

func buildTxSvc(d *coreDependencies, txsvc txSvc.EngineReader, accs txSvc.AccountReader,
	vstore *vmgr.ValidatorMgr, cometBftClient txSvc.BlockchainTransactor, nodeApp txSvc.NodeApplication) *txSvc.Service {
	return txSvc.NewService(txsvc, accs, vstore, cometBftClient, nodeApp,
		txSvc.WithLogger(*d.log.Named("tx-service")),
	)
}

func buildAdminSvc(d *coreDependencies, transactor admSvc.BlockchainTransactor, node admSvc.NodeApplication, validatorStore admSvc.ValidatorReader) *admSvc.Service {
	pk, err := crypto.Ed25519PrivateKeyFromBytes(d.privKey.Bytes())
	if err != nil {
		failBuild(err, "failed to build admin service")
	}

	signer := auth.Ed25519Signer{Ed25519PrivateKey: *pk}

	return admSvc.NewService(transactor, node, validatorStore, &signer, d.cfg,
		admSvc.WithLogger(*d.log.Named("admin-service")),
	)
}

func buildDatasetsModule(d *coreDependencies, eng datasets.Engine, accs datasets.AccountStore) *datasets.DatasetModule {
	feeMultiplier := 1
	if d.genesisCfg.ConsensusParams.WithoutGasCosts {
		feeMultiplier = 0
	}

	return datasets.NewDatasetModule(eng, accs,
		datasets.WithLogger(*d.log.Named("dataset-module")),
		datasets.WithFeeMultiplier(int64(feeMultiplier)),
	)
}

func buildEngine(d *coreDependencies, closer *closeFuncs, a *sessions.MultiCommitter) *execution.GlobalContext {
	extensions, err := getExtensions(d.ctx, d.cfg.AppCfg.ExtensionEndpoints)
	if err != nil {
		failBuild(err, "failed to get extensions")
	}

	for name := range extensions {
		d.log.Info("registered extension", zap.String("name", name))
	}

	reg, err := registry.NewRegistry(d.ctx, func(ctx context.Context, dbid string, create bool) (registry.Pool, error) {
		return sqlite.NewPool(ctx, dbid, 1, 2, true)
	}, d.cfg.AppCfg.SqliteFilePath, registry.WithReaderWaitTimeout(time.Millisecond*100), registry.WithLogger(
		*d.log.Named("registry"),
	))
	if err != nil {
		failBuild(err, "failed to build registry")
	}

	eng, err := execution.NewGlobalContext(d.ctx, reg, extensions)
	if err != nil {
		failBuild(err, "failed to build engine")
	}
	err = a.Register(engineName, reg)
	if err != nil {
		failBuild(err, "failed to register engine")
	}

	closer.addCloser(reg.Close)

	return eng
}

func buildAccountsModule(store *accounts.AccountStore) *acctmod.AccountsModule {
	return acctmod.NewAccountsModule(store)
}

func buildAccountRepository(d *coreDependencies, closer *closeFuncs, ac *sessions.MultiCommitter) *accounts.AccountStore {

	db, err := d.opener(d.ctx, filepath.Join(d.cfg.RootDir, applicationDirName, accountsDBName), 1, 2, true)
	if err != nil {
		failBuild(err, "failed to open accounts db")
	}
	closer.addCloser(db.Close)

	adapted := adapter.PoolAdapater{Pool: db}

	com := committable.New(adapted)

	genCfg := d.genesisCfg
	b, err := accounts.NewAccountStore(d.ctx,
		&adapted,
		com,
		accounts.WithLogger(*d.log.Named("accountStore")),
		accounts.WithNonces(!genCfg.ConsensusParams.WithoutNonces),
		accounts.WithGasCosts(!genCfg.ConsensusParams.WithoutGasCosts),
	)
	if err != nil {
		failBuild(err, "failed to build account store")
	}
	closer.addCloser(b.Close)

	err = ac.Register(accountsDBName, com)
	if err != nil {
		failBuild(err, "failed to register account store")
	}

	return b
}

func buildValidatorManager(d *coreDependencies, closer *closeFuncs, ac *sessions.MultiCommitter) *vmgr.ValidatorMgr {
	db, err := d.opener(d.ctx, filepath.Join(d.cfg.RootDir, applicationDirName, validatorDBName), 1, 2, true)
	if err != nil {
		failBuild(err, "failed to open validator db")
	}
	closer.addCloser(db.Close)

	joinExpiry := d.genesisCfg.ConsensusParams.Validator.JoinExpiry
	feeMultiplier := 1
	if d.genesisCfg.ConsensusParams.WithoutGasCosts {
		feeMultiplier = 0
	}

	adapted := adapter.PoolAdapater{Pool: db}

	com := committable.New(adapted)

	v, err := vmgr.NewValidatorMgr(d.ctx,
		&adapted,
		com,
		vmgr.WithLogger(*d.log.Named("validatorStore")),
		vmgr.WithJoinExpiry(joinExpiry),
		vmgr.WithFeeMultiplier(int64(feeMultiplier)),
	)
	if err != nil {
		failBuild(err, "failed to build validator store")
	}

	err = ac.Register(validatorDBName, com)
	if err != nil {
		failBuild(err, "failed to register validator store")
	}

	return v
}

func buildValidatorModule(d *coreDependencies, accs datasets.AccountStore,
	vals validators.ValidatorMgr) *validators.ValidatorModule {
	return validators.NewValidatorModule(vals, accs,
		validators.WithLogger(*d.log.Named("validator-module")))
}

func buildSnapshotter(d *coreDependencies) *snapshots.SnapshotStore {
	return nil
	// TODO: Uncomment when we have statesync ready
	// cfg := d.cfg.AppCfg
	// if !cfg.SnapshotConfig.Enabled {
	// 	return nil
	// }

	// return snapshots.NewSnapshotStore(cfg.SqliteFilePath,
	// 	cfg.SnapshotConfig.SnapshotDir,
	// 	cfg.SnapshotConfig.RecurringHeight,
	// 	cfg.SnapshotConfig.MaxSnapshots,
	// 	snapshots.WithLogger(*d.log.Named("snapshotStore")),
	// )
}

func buildBootstrapper(d *coreDependencies) *snapshots.Bootstrapper {
	return nil
	// TODO: Uncomment when we have statesync ready
	// rcvdSnapsDir := filepath.Join(d.cfg.RootDir, rcvdSnapsDirName)
	// bootstrapper, err := snapshots.NewBootstrapper(d.cfg.AppCfg.SqliteFilePath, rcvdSnapsDir)
	// if err != nil {
	// 	failBuild(err, "Bootstrap module initialization failure")
	// }
	// return bootstrapper
}

func fileExists(name string) bool {
	_, err := os.Stat(name)
	return !os.IsNotExist(err)
}

func loadTLSCertificate(keyFile, certFile, hostname string) (*tls.Certificate, error) {
	keyExists, certExists := fileExists(keyFile), fileExists(certFile)
	if certExists != keyExists { // one but not both
		return nil, fmt.Errorf("missing a key/cert pair file")

	}
	if !keyExists {
		// Auto-generate a new key/cert pair using any provided host name in the
		// "Subject Alternate Name" section of the certificate (either IP or a
		// hostname like kwild23.applicationX.org).
		if err := genCertPair(certFile, keyFile, []string{hostname}); err != nil {
			return nil, fmt.Errorf("failed to generate TLS key pair: %v", err)
		}
		// TODO: generate a separate CA certificate. Browsers don't like that
		// the site certificate is also a CA, but Go clients are fine with it.
	}
	keyPair, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load TLS key pair: %v", err)
	}
	return &keyPair, nil
}

func buildGrpcServer(d *coreDependencies, txsvc txpb.TxServiceServer) *kwilgrpc.Server {
	lis, err := net.Listen("tcp", d.cfg.AppCfg.GrpcListenAddress)
	if err != nil {
		failBuild(err, "failed to build grpc server")
	}
	if d.cfg.AppCfg.EnableRPCTLS {
		lis = tls.NewListener(lis, &tls.Config{
			Certificates: []tls.Certificate{*d.keypair},
			MinVersion:   tls.VersionTLS12,
		})
	}

	// Increase the maximum message size to the largest allowable transaction
	// size plus a very generous 16KiB buffer for message overhead.
	const msgOverHeadBuffer = 16384
	recvLimit := d.cfg.ChainCfg.Mempool.MaxTxBytes + msgOverHeadBuffer
	grpcServer := kwilgrpc.New(*d.log.Named("grpc-server"), lis, kwilgrpc.WithSrvOpt(grpc.MaxRecvMsgSize(recvLimit)))
	txpb.RegisterTxServiceServer(grpcServer, txsvc)

	// right now, the function service is just registered to the public tx service
	functionsvc := functionSvc.FunctionService{}
	functionpb.RegisterFunctionServiceServer(grpcServer, &functionsvc)

	grpc_health_v1.RegisterHealthServer(grpcServer, buildHealthSvc(d))

	return grpcServer
}

func buildHealthSvc(d *coreDependencies) *healthsvc.Server {
	// health service
	registrar := healthcheck.NewRegistrar(*d.log.Named("auth-healthcheck"))
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

func buildAdminService(d *coreDependencies, closer *closeFuncs, admsvc admpb.AdminServiceServer, txsvc txpb.TxServiceServer) *kwilgrpc.Server {
	u, err := url.ParseURL(d.cfg.AppCfg.AdminListenAddress)
	if err != nil {
		failBuild(err, "failed to build admin service")
	}

	switch u.Scheme {
	default:
		failBuild(err, "unknown admin service protocol "+u.Scheme.String())
	case url.TCP:

		// if tcp, we need to set up TLS
		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", u.Port))
		if err != nil {
			failBuild(err, "failed to build grpc server")
		}
		closer.addCloser(lis.Close)

		// client certs
		caCertPool := x509.NewCertPool()
		var clientsCerts []byte
		if clientsFile := filepath.Join(d.cfg.RootDir, defaultAdminClients); fileExists(clientsFile) {
			clientsCerts, err = os.ReadFile(clientsFile)
			if err != nil {
				failBuild(err, "failed to load client CAs file")
			}
		} else if d.autogen {
			clientCredsFileBase := filepath.Join(d.cfg.RootDir, "auth")
			clientCertFile, clientKeyFile := clientCredsFileBase+".cert", clientCredsFileBase+".key"
			err = transport.GenTLSKeyPair(clientCertFile, clientKeyFile, "kwild CA", nil)
			if err != nil {
				failBuild(err, "failed to generate admin client credentials")
			}
			d.log.Info("generated admin service client key pair", zap.String("cert", clientCertFile), zap.String("key", clientKeyFile))
			if clientsCerts, err = os.ReadFile(clientCertFile); err != nil {
				failBuild(err, "failed to read auto-generate client certificate")
			}
			if err = os.WriteFile(clientsFile, clientsCerts, 0644); err != nil {
				failBuild(err, "failed to write client CAs file")
			}
			d.log.Info("generated admin service client CAs file", zap.String("file", clientsFile))
		} else {
			d.log.Info("No admin client CAs file. Use kwil-admin's node gen-auth-key command to generate")
		}
		if len(clientsCerts) > 0 && !caCertPool.AppendCertsFromPEM(clientsCerts) {
			failBuild(err, "invalid client CAs file")
		}

		// TLS configuration for mTLS (mutual TLS) protocol-level authentication
		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{*d.keypair},
			ClientAuth:   tls.RequireAndVerifyClientCert,
			ClientCAs:    caCertPool,
		}

		creds := grpc.Creds(credentials.NewTLS(tlsConfig))
		opts := kwilgrpc.WithSrvOpt(creds)

		grpcServer := kwilgrpc.New(*d.log.Named("auth-grpc-server"), lis, opts)
		admpb.RegisterAdminServiceServer(grpcServer, admsvc)
		txpb.RegisterTxServiceServer(grpcServer, txsvc)
		grpc_health_v1.RegisterHealthServer(grpcServer, buildHealthSvc(d))

		return grpcServer
	case url.UNIX:
		// if unix, we need to set up unix socket
		err = os.MkdirAll(filepath.Dir(u.Target), 0700) // ensure parent dir exists
		if err != nil {
			failBuild(err, "failed to create admin service unix socket directory at "+filepath.Dir(u.Target))
		}

		d.log.Info("creating admin service unix socket at " + u.Target)

		// unix sockets will remain "open" unless lis.Close is called
		// In case of crash, this creates very bad ux for developers.
		// suggested approach here (not sure about this for obvious reasons):
		// https://gist.github.com/hakobe/6f70d69b8c5243117787fd488ae7fbf2

		err = syscall.Unlink(u.Target)
		if err != nil && !os.IsNotExist(err) {
			failBuild(err, "failed to build grpc server")
		}

		lis, err := net.Listen("unix", u.Target)
		if err != nil {
			failBuild(err, "failed to listen to unix socket")
		}

		closer.addCloser(lis.Close)

		err = os.Chmod(u.Target, 0777) // TODO: probably want this to be more restrictive
		if err != nil {
			failBuild(err, "failed to build grpc server")
		}

		grpcServer := kwilgrpc.New(*d.log.Named("auth-grpc-server"), lis)
		admpb.RegisterAdminServiceServer(grpcServer, admsvc)
		txpb.RegisterTxServiceServer(grpcServer, txsvc)
		grpc_health_v1.RegisterHealthServer(grpcServer, buildHealthSvc(d))

		return grpcServer
	}

	failBuild(nil, "unknown error building admin service") // should never get here
	return nil
}

func buildGatewayServer(d *coreDependencies) *gateway.GatewayServer {
	gw, err := gateway.NewGateway(d.ctx, d.cfg.AppCfg.HTTPListenAddress,
		gateway.WithLogger(*d.log.Named("gateway")),
		gateway.WithMiddleware(cors.MCors([]string{})),
		gateway.WithGrpcService(d.cfg.AppCfg.GrpcListenAddress, txpb.RegisterTxServiceHandlerFromEndpoint),
		gateway.WithGrpcService(d.cfg.AppCfg.GrpcListenAddress, functionpb.RegisterFunctionServiceHandlerFromEndpoint),
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
	// for now, I'm just using a KV store for my atomic commit.  This probably is not ideal; a file may be better
	// I'm simply using this because we know it fsyncs the data to disk
	db, err := badger.NewBadgerDB(d.ctx, filepath.Join(d.cfg.RootDir, signingDirName), &badger.Options{
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

	genDoc, err := extractGenesisDoc(d.genesisCfg)
	if err != nil {
		failBuild(err, "failed to generate cometbft genesis configuration")
	}

	nodeCfg := newCometConfig(d.cfg)
	if nodeCfg.P2P.SeedMode {
		d.log.Info("Seed mode enabled.")
		if !nodeCfg.P2P.PexReactor {
			d.log.Warn("Enabling peer exchange to run in seed mode.")
			nodeCfg.P2P.PexReactor = true
		}
	}

	node, err := cometbft.NewCometBftNode(abciApp, nodeCfg, genDoc, d.privKey,
		readWriter, &d.log)
	if err != nil {
		failBuild(err, "failed to build comet node")
	}

	return node
}

func buildCommitter(d *coreDependencies, closers *closeFuncs) *sessions.MultiCommitter {
	kv, err := badger.NewBadgerDB(d.ctx, filepath.Join(d.cfg.RootDir, applicationDirName, "committer"),
		&badger.Options{
			GuaranteeFSync:                true,
			GarbageCollectionInterval:     5 * time.Minute,
			Logger:                        *d.log.Named("atomic-committer-kv-store"),
			GarbageCollectionDiscardRatio: 0.5,
		},
	)
	if err != nil {
		failBuild(err, "failed to open atomic committer kv")
	}

	closers.addCloser(kv.Close)

	return sessions.NewCommitter(kv, make(map[string]sessions.Committable), sessions.WithLogger(*d.log.Named("atomic-committer")))
}

func failBuild(err error, msg string) {
	if err != nil {
		panic(fmt.Sprintf("%s: %s", msg, err.Error()))
	}
}
