package server

import (
	"compress/gzip"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	neturl "net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/kwilteam/kwil-db/cmd/kwild/config"
	"github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/common/sql"
	"github.com/kwilteam/kwil-db/internal/abci"
	"github.com/kwilteam/kwil-db/internal/abci/cometbft"
	"github.com/kwilteam/kwil-db/internal/abci/meta"
	"github.com/kwilteam/kwil-db/internal/accounts"
	"github.com/kwilteam/kwil-db/internal/engine/execution"
	"github.com/kwilteam/kwil-db/internal/kv"
	"github.com/kwilteam/kwil-db/internal/kv/badger"
	"github.com/kwilteam/kwil-db/internal/listeners"
	functionSvc "github.com/kwilteam/kwil-db/internal/services/grpc/function/v0"
	"github.com/kwilteam/kwil-db/internal/services/grpc/healthsvc/v0"
	txSvc "github.com/kwilteam/kwil-db/internal/services/grpc/txsvc/v1"
	gateway "github.com/kwilteam/kwil-db/internal/services/grpc_gateway"
	"github.com/kwilteam/kwil-db/internal/services/grpc_gateway/middleware/cors"
	kwilgrpc "github.com/kwilteam/kwil-db/internal/services/grpc_server"
	healthcheck "github.com/kwilteam/kwil-db/internal/services/health"
	simple_checker "github.com/kwilteam/kwil-db/internal/services/health/simple-checker"
	rpcserver "github.com/kwilteam/kwil-db/internal/services/jsonrpc"
	"github.com/kwilteam/kwil-db/internal/services/jsonrpc/adminsvc"
	"github.com/kwilteam/kwil-db/internal/services/jsonrpc/funcsvc"
	usersvc "github.com/kwilteam/kwil-db/internal/services/jsonrpc/usersvc"
	"github.com/kwilteam/kwil-db/internal/sql/pg"
	"github.com/kwilteam/kwil-db/internal/statesync"
	"github.com/kwilteam/kwil-db/internal/txapp"
	"github.com/kwilteam/kwil-db/internal/voting"
	"github.com/kwilteam/kwil-db/internal/voting/broadcast"

	"github.com/kwilteam/kwil-db/core/crypto"
	"github.com/kwilteam/kwil-db/core/crypto/auth"
	"github.com/kwilteam/kwil-db/core/log"
	functionpb "github.com/kwilteam/kwil-db/core/rpc/protobuf/function/v0"
	txpb "github.com/kwilteam/kwil-db/core/rpc/protobuf/tx/v1"
	"github.com/kwilteam/kwil-db/core/rpc/transport"

	abciTypes "github.com/cometbft/cometbft/abci/types"
	cmtEd "github.com/cometbft/cometbft/crypto/ed25519"
	cmtlocal "github.com/cometbft/cometbft/rpc/client/local"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
)

// initStores prepares the datastores with an atomic DB transaction. These
// actions are performed outside of ABCI's DB sessions. The stores should not
// keep the db after initialization. Their functions accept a DB connection.
func initStores(d *coreDependencies, db *pg.DB) error {
	initTx, err := db.BeginTx(d.ctx)
	if err != nil {
		return fmt.Errorf("could not start app initialization DB transaction: %w", err)
	}
	defer initTx.Rollback(d.ctx)

	// chain meta data for abci and txApp
	initChainMetadata(d, initTx)
	// upgrade v0.7.0 that had ABCI meta in badgerDB and nowhere else...
	if err = migrateOldChainState(d, initTx); err != nil {
		return err
	}

	// account store
	initAccountRepository(d, initTx)

	if err = initTx.Commit(d.ctx); err != nil {
		return fmt.Errorf("failed to commit the app initialization DB transaction: %w", err)
	}

	return nil
}

// migrateOldChainState detects if the old badger-based chain metadata store
// needs to be migrated into the new postgres metadata store and does the update.
func migrateOldChainState(d *coreDependencies, initTx sql.Tx) error {
	height, appHash, err := meta.GetChainState(d.ctx, initTx)
	if err != nil {
		return fmt.Errorf("failed to get chain state from metadata store: %w", err)
	}
	if height != -1 { // have already updated the meta chain tables
		d.log.Infof("Application's chain metadata store loaded at height %d / app hash %x",
			height, appHash)
		return nil
	} // else we are either at genesis or we need to migrate from badger

	// detect old badger kv DB for ABCI's metadata
	badgerPath := filepath.Join(d.cfg.RootDir, abciDirName, config.ABCIInfoSubDirName)
	height, appHash, err = getOldChainState(d, badgerPath)
	if err != nil {
		return fmt.Errorf("unable to read old metadata store: %w", err)
	}
	if height < 1 { // badger hadn't been used, and we're just at genesis
		if height == 0 { // files existed, but still at genesis
			if err = os.RemoveAll(badgerPath); err != nil {
				d.log.Errorf("failed to remove old badger db file (%s): %v", badgerPath, err)
			}
		}
		return nil
	}

	d.log.Infof("Migrating from badger DB chain metadata to postgresql: height %d, apphash %x",
		height, appHash)
	err = meta.SetChainState(d.ctx, initTx, height, appHash)
	if err != nil {
		return fmt.Errorf("failed to migrate height and app hash: %w", err)
	}

	if err = os.RemoveAll(badgerPath); err != nil {
		d.log.Errorf("failed to remove old badger db file (%s): %v", badgerPath, err)
	}

	return nil
}

// getOldChainState attempts to retrieve the height and app hash from any legacy
// badger metadata store. If the folder does not exists, it is not an error, but
// height -1 is returned. If the DB exists but there are no entries, it returns
// 0 and no error.
func getOldChainState(d *coreDependencies, badgerPath string) (int64, []byte, error) {
	if _, err := os.Stat(badgerPath); errors.Is(err, os.ErrNotExist) {
		return -1, nil, nil // would hit the kv.ErrKeyNotFound case, but we don't want artifacts
	}
	badgerKv, err := badger.NewBadgerDB(d.ctx, badgerPath, &badger.Options{
		Logger: *d.log.Named("old-abci-kv-store"),
	})
	if err != nil {
		return 0, nil, fmt.Errorf("failed to open badger: %w", err)
	}

	defer badgerKv.Close()

	appHash, err := badgerKv.Get([]byte("a"))
	if err == kv.ErrKeyNotFound {
		return 0, nil, nil
	}
	if err != nil {
		return 0, nil, err
	}
	height, err := badgerKv.Get([]byte("b"))
	if err == kv.ErrKeyNotFound {
		return 0, nil, nil
	}
	if err != nil {
		return 0, nil, err
	}
	return int64(binary.BigEndian.Uint64(height)), slices.Clone(appHash), nil
}

func buildServer(d *coreDependencies, closers *closeFuncs) *Server {
	// main postgres db
	db := buildDB(d, closers)

	if err := initStores(d, db); err != nil {
		failBuild(err, "initStores failed")
	}

	// engine
	e := buildEngine(d, db)

	// Initialize the events and voting data stores
	ev := buildEventStore(d, closers) // makes own DB connection

	// these are dummies, but they might need init in the future.
	snapshotter := buildSnapshotter(d)
	statesyncer := buildStatesyncer(d)

	// this is a hack
	// we need the cometbft client to broadcast txs.
	// in order to get this, we need the comet node
	// to get the comet node, we need the abci app
	// to get the abci app, we need the tx router
	// but the tx router needs the cometbft client
	txApp := buildTxApp(d, db, e, ev, snapshotter)

	abciApp := buildAbci(d, txApp, snapshotter, statesyncer)

	// NOTE: buildCometNode immediately starts talking to the abciApp and
	// replaying blocks (and using atomic db tx commits), i.e. calling
	// FinalizeBlock+Commit. This is not just a constructor, sadly.
	cometBftNode := buildCometNode(d, closers, abciApp)

	cometBftClient := buildCometBftClient(cometBftNode)
	wrappedCmtClient := &wrappedCometBFTClient{
		cl:    cometBftClient,
		cache: abciApp,
	}
	txApp.SetReplayStatusChecker(cometBftNode.IsCatchup)

	eventBroadcaster := buildEventBroadcaster(d, ev, wrappedCmtClient, txApp)
	abciApp.SetEventBroadcaster(eventBroadcaster.RunBroadcast)

	// listener manager
	listeners := buildListenerManager(d, ev, cometBftNode, txApp)

	// user service and server
	jsonRPCTxSvc := usersvc.NewService(db, e, wrappedCmtClient, txApp,
		*d.log.Named("user-json-svc"))
	jsonRPCServer, err := rpcserver.NewServer(d.cfg.AppCfg.JSONRPCListenAddress,
		*d.log.Named("user-jsonrpc-server"))
	if err != nil {
		failBuild(err, "unable to create json-rpc server")
	}
	jsonRPCServer.RegisterSvc(jsonRPCTxSvc)
	jsonRPCServer.RegisterSvc(&funcsvc.Service{})

	// admin service and server
	signer := buildSigner(d)
	jsonAdminSvc := adminsvc.NewService(db, wrappedCmtClient, txApp, signer, d.cfg,
		d.genesisCfg.ChainID, *d.log.Named("admin-json-svc"))
	jsonRPCAdminServer := buildJRPCAdminServer(d)
	jsonRPCAdminServer.RegisterSvc(jsonAdminSvc)
	jsonRPCAdminServer.RegisterSvc(jsonRPCTxSvc)
	jsonRPCAdminServer.RegisterSvc(&funcsvc.Service{})

	// legacy tx service and grpc server
	txsvc := buildTxSvc(d, db, e, wrappedCmtClient, txApp)
	grpcServer := buildGrpcServer(d, txsvc)

	return &Server{
		grpcServer:         grpcServer,
		jsonRPCServer:      jsonRPCServer,
		jsonRPCAdminServer: jsonRPCAdminServer,
		gateway:            buildGatewayServer(d, grpcServer.Addr()),
		cometBftNode:       cometBftNode,
		listenerManager:    listeners,
		log:                *d.log.Named("server"),
		closers:            closers,
		cfg:                d.cfg,
	}
}

// dbOpener opens a sessioned database connection.  Note that in this function the
// dbName is not a Kwil dataset, but a database that can contain multiple
// datasets in different postgresql "schema".
type dbOpener func(ctx context.Context, dbName string, maxConns uint32) (*pg.DB, error)

func newDBOpener(host, port, user, pass string) dbOpener {
	return func(ctx context.Context, dbName string, maxConns uint32) (*pg.DB, error) {
		cfg := &pg.DBConfig{
			PoolConfig: pg.PoolConfig{
				ConnConfig: pg.ConnConfig{
					Host:   host,
					Port:   port,
					User:   user,
					Pass:   pass,
					DBName: dbName,
				},
				MaxConns: maxConns,
			},
			SchemaFilter: func(s string) bool {
				return strings.HasPrefix(s, pg.DefaultSchemaFilterPrefix)
			},
		}
		return pg.NewDB(ctx, cfg)
	}
}

// poolOpener opens a basic database connection pool.
type poolOpener func(ctx context.Context, dbName string, maxConns uint32) (*pg.Pool, error)

func newPoolBOpener(host, port, user, pass string) poolOpener {
	return func(ctx context.Context, dbName string, maxConns uint32) (*pg.Pool, error) {
		cfg := &pg.PoolConfig{
			ConnConfig: pg.ConnConfig{
				Host:   host,
				Port:   port,
				User:   user,
				Pass:   pass,
				DBName: dbName,
			},
			MaxConns: maxConns,
		}
		return pg.NewPool(ctx, cfg)
	}
}

// coreDependencies holds dependencies that are widely used
type coreDependencies struct {
	ctx        context.Context
	autogen    bool
	cfg        *config.KwildConfig
	genesisCfg *config.GenesisConfig
	privKey    cmtEd.PrivKey
	log        log.Logger
	dbOpener   dbOpener
	poolOpener poolOpener
	keypair    *tls.Certificate
}

// closeFuncs holds a list of closers
// it is used to close all resources on shutdown
type closeFuncs struct {
	closers []func() error
	logger  log.Logger
}

func (c *closeFuncs) addCloser(f func() error, msg string) {
	// push to top of stack
	c.closers = slices.Insert(c.closers, 0, func() error {
		c.logger.Info(msg)
		return f()
	})
}

// closeAll closes all closers
func (c *closeFuncs) closeAll() error {
	var err error
	for _, closer := range c.closers {
		err = errors.Join(closer())
	}

	return err
}

func buildTxApp(d *coreDependencies, db *pg.DB, engine *execution.GlobalContext, ev *voting.EventStore, snapshotter *statesync.SnapshotStore) *txapp.TxApp {
	var sh txapp.Snapshotter
	if snapshotter != nil {
		sh = snapshotter
	}

	txApp, err := txapp.NewTxApp(db, engine, buildSigner(d), ev, sh, d.genesisCfg.ChainID,
		!d.genesisCfg.ConsensusParams.WithoutGasCosts, d.cfg.AppCfg.Extensions, *d.log.Named("tx-router"))
	if err != nil {
		failBuild(err, "failed to build new TxApp")
	}
	return txApp
}

func buildAbci(d *coreDependencies, txApp abci.TxApp, snapshotter *statesync.SnapshotStore, statesyncer *statesync.StateSyncer) *abci.AbciApp {
	var sh abci.SnapshotModule
	if snapshotter != nil {
		sh = snapshotter
	}

	var ss abci.StateSyncModule
	if statesyncer != nil {
		ss = statesyncer
	}

	cfg := &abci.AbciConfig{
		GenesisAppHash:     d.genesisCfg.ComputeGenesisHash(),
		ChainID:            d.genesisCfg.ChainID,
		ApplicationVersion: d.genesisCfg.ConsensusParams.Version.App,
		GenesisAllocs:      d.genesisCfg.Alloc,
		GasEnabled:         !d.genesisCfg.ConsensusParams.WithoutGasCosts,
	}
	return abci.NewAbciApp(cfg, sh, ss, txApp,
		&txapp.ConsensusParams{
			VotingPeriod:       d.genesisCfg.ConsensusParams.Votes.VoteExpiry,
			JoinVoteExpiration: d.genesisCfg.ConsensusParams.Validator.JoinExpiry,
		},
		*d.log.Named("abci"),
	)
}

func buildEventBroadcaster(d *coreDependencies, ev broadcast.EventStore, b broadcast.Broadcaster, txapp *txapp.TxApp) *broadcast.EventBroadcaster {
	return broadcast.NewEventBroadcaster(ev, b, txapp, txapp, buildSigner(d), d.genesisCfg.ChainID)
}

func buildEventStore(d *coreDependencies, closers *closeFuncs) *voting.EventStore {
	// NOTE: we're using the same postgresql database, but isolated pg schema.
	db, err := d.poolOpener(d.ctx, d.cfg.AppCfg.DBName, 10)
	if err != nil {
		failBuild(err, "failed to build event store")
	}
	closers.addCloser(db.Close, "closing event store")

	e, err := voting.NewEventStore(d.ctx, db)
	if err != nil {
		failBuild(err, "failed to build event store")
	}

	return e
}

func buildTxSvc(d *coreDependencies, db *pg.DB, txsvc txSvc.EngineReader, cometBftClient txSvc.BlockchainTransactor, nodeApp txSvc.NodeApplication) *txSvc.Service {
	return txSvc.NewService(db, txsvc, cometBftClient, nodeApp,
		txSvc.WithLogger(*d.log.Named("tx-service")),
	)
}

func buildSigner(d *coreDependencies) *auth.Ed25519Signer {
	pk, err := crypto.Ed25519PrivateKeyFromBytes(d.privKey.Bytes())
	if err != nil {
		failBuild(err, "failed to build admin service")
	}

	return &auth.Ed25519Signer{Ed25519PrivateKey: *pk}
}

func buildDB(d *coreDependencies, closer *closeFuncs) *pg.DB {
	// Check if the database is supposed to be restored from the snapshot
	// If yes, restore the database from the snapshot
	restoreDB(d)

	db, err := d.dbOpener(d.ctx, d.cfg.AppCfg.DBName, 11)
	if err != nil {
		failBuild(err, "kwild database open failed")
	}
	closer.addCloser(db.Close, "closing main DB")

	return db
}

// restoreDB restores the database from a snapshot if the genesis apphash is specified.
// Genesis apphash ensures that all the nodes in the network start from the same state.
// Genesis apphash should match the hash of the snapshot file.
// Snapshot file can be compressed or uncompressed represented by .gz extension.
// DB restoration from snapshot is skipped in the following scenarios:
//   - If the DB is already initialized (i.e this is not a new node)
//   - If the genesis apphash is not specified
//   - If statesync is enabled. Statesync will take care of rapildly syncing the database
//     to the network state using statesync snapshots.
func restoreDB(d *coreDependencies) {
	if isDbInitialized(d) || d.genesisCfg.DataAppHash == nil || d.cfg.ChainCfg.StateSync.Enable {
		return
	}

	genCfg := d.genesisCfg
	appCfg := d.cfg.AppCfg
	// DB is uninitialized and genesis apphash is specified.
	// DB is supposed to be restored from the snapshot.
	// Ensure that the snapshot file exists and the snapshot hash matches the genesis apphash.

	// Ensure that the snapshot file exists, if node is supposed to start with a snapshot state
	if genCfg.DataAppHash != nil && appCfg.SnapshotFile == "" {
		failBuild(nil, "snapshot file not provided")
	}

	// Snapshot file exists
	snapFile, err := os.Open(appCfg.SnapshotFile)
	if err != nil {
		failBuild(err, "failed to open snapshot file")
	}

	// Check if the snapshot file is compressed, if yes decompress it
	var reader io.Reader
	if strings.HasSuffix(appCfg.SnapshotFile, ".gz") {
		// Decompress the snapshot file
		gzipReader, err := gzip.NewReader(snapFile)
		if err != nil {
			failBuild(err, "failed to create gzip reader")
		}
		defer gzipReader.Close()
		reader = gzipReader
	} else {
		reader = snapFile
	}

	// Restore DB from the snapshot if snapshot matches.
	err = statesync.RestoreDB(d.ctx, reader, appCfg.DBName, appCfg.DBUser, appCfg.DBPass,
		appCfg.DBHost, appCfg.DBPort, genCfg.DataAppHash, d.log)
	if err != nil {
		failBuild(err, "failed to restore DB from snapshot")
	}
}

// isDbInitialized checks if the database is already initialized.
func isDbInitialized(d *coreDependencies) bool {
	db, err := d.dbOpener(d.ctx, d.cfg.AppCfg.DBName, 11)
	if err != nil {
		failBuild(err, "kwild database open failed")
	}
	defer db.Close()

	// Check if the database is empty or initialized previously
	// If the database is empty, we need to restore the database from the snapshot
	initTx, err := db.BeginTx(d.ctx)
	if err != nil {
		failBuild(err, "could not start app initialization DB transaction")
	}
	defer initTx.Rollback(d.ctx)

	_, err = voting.GetValidators(d.ctx, initTx)
	// ERROR: relation "kwild_voting.voters" does not exist
	// assumption that error is due to the missing table and schema.
	return err == nil
}

func buildEngine(d *coreDependencies, db *pg.DB) *execution.GlobalContext {
	extensions, err := getExtensions(d.ctx, d.cfg.AppCfg.ExtensionEndpoints)
	if err != nil {
		failBuild(err, "failed to get extensions")
	}

	for name := range extensions {
		d.log.Info("registered extension", log.String("name", name))
	}

	tx, err := db.BeginTx(d.ctx)
	if err != nil {
		failBuild(err, "failed to start transaction")
	}
	defer tx.Rollback(d.ctx)

	err = execution.InitializeEngine(d.ctx, tx)
	if err != nil {
		failBuild(err, "failed to initialize engine")
	}

	eng, err := execution.NewGlobalContext(d.ctx, tx, extensions, &common.Service{
		Logger:           d.log.Named("engine").Sugar(),
		ExtensionConfigs: d.cfg.AppCfg.Extensions,
	})
	if err != nil {
		failBuild(err, "failed to build engine")
	}

	err = tx.Commit(d.ctx)
	if err != nil {
		failBuild(err, "failed to commit engine init db txn")
	}

	return eng
}

func initChainMetadata(d *coreDependencies, tx sql.Tx) {
	err := meta.InitializeMetaStore(d.ctx, tx)
	if err != nil {
		failBuild(err, "failed to initialize chain metadata store")
	}
}

func initAccountRepository(d *coreDependencies, tx sql.Tx) {
	err := accounts.InitializeAccountStore(d.ctx, tx)
	if err != nil {
		failBuild(err, "failed to initialize account store")
	}
}

func buildSnapshotter(d *coreDependencies) *statesync.SnapshotStore {
	cfg := d.cfg.AppCfg
	if !cfg.Snapshots.Enabled {
		return nil
	}

	dbCfg := &statesync.DBConfig{
		DBUser: cfg.DBUser,
		DBPass: cfg.DBPass,
		DBHost: cfg.DBHost,
		DBPort: cfg.DBPort,
		DBName: cfg.DBName,
	}

	snapshotCfg := &statesync.SnapshotConfig{
		SnapshotDir:     cfg.Snapshots.SnapshotDir,
		RecurringHeight: cfg.Snapshots.RecurringHeight,
		MaxSnapshots:    int(cfg.Snapshots.MaxSnapshots),
	}

	ss, err := statesync.NewSnapshotStore(snapshotCfg, dbCfg, *d.log.Named("snapshot-store"))
	if err != nil {
		failBuild(err, "failed to build snapshot store")
	}
	return ss
}

func buildStatesyncer(d *coreDependencies) *statesync.StateSyncer {
	if !d.cfg.ChainCfg.StateSync.Enable {
		return nil
	}

	cfg := d.cfg.AppCfg

	dbCfg := &statesync.DBConfig{
		DBUser: cfg.DBUser,
		DBPass: cfg.DBPass,
		DBHost: cfg.DBHost,
		DBPort: cfg.DBPort,
		DBName: cfg.DBName,
	}

	providers := strings.Split(d.cfg.ChainCfg.StateSync.RPCServers, ",")

	if len(providers) == 0 {
		failBuild(nil, "failed to configure state syncer, no remote servers provided.")
	}

	if len(providers) == 1 {
		// Duplicating the same provider to satisfy cometbft statesync requirement of having at least 2 providers.
		// Statesynce module doesn't have the same requirements and
		// can work with a single provider (providers are passed as is)
		d.cfg.ChainCfg.StateSync.RPCServers += "," + providers[0]
	}

	configDone := false
	for _, p := range providers {
		clt, err := statesync.ChainRPCClient(p)
		if err != nil {
			continue
		}

		// Try to fetch the status of the remote server.
		res, err := clt.Header(d.ctx, nil)
		if err != nil {
			continue
		}

		// If the remote server is in the same chain, we can trust it.
		if res.Header.ChainID != d.genesisCfg.ChainID {
			continue
		}

		// Get the trust height and trust hash from the remote server
		d.cfg.ChainCfg.StateSync.TrustHeight = res.Header.Height
		d.cfg.ChainCfg.StateSync.TrustHash = res.Header.Hash().String()
		configDone = true
		break
	}

	if !configDone {
		failBuild(nil, "failed to configure state syncer, failed to fetch trust options from the remote server.")
	}

	// create state syncer
	return statesync.NewStateSyncer(d.ctx, dbCfg, d.cfg.ChainCfg.StateSync.SnapshotDir,
		providers, *d.log.Named("state-syncer"))
}

// tlsConfig returns a tls.Config to be used with the admin RPC service. If
// withClientAuth is true, the config will require client authentication (mutual
// TLS), otherwise it is standard TLS for encryption and server authentication.
func tlsConfig(d *coreDependencies, withClientAuth bool) *tls.Config {
	if d.keypair == nil {
		return nil
	}
	if !withClientAuth {
		// TLS only for encryption and authentication of server to client.
		return &tls.Config{
			Certificates: []tls.Certificate{*d.keypair},
		}
	} // else try to load authorized client certs/pubkeys

	var err error
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
		err = transport.GenTLSKeyPair(clientCertFile, clientKeyFile, "local kwild CA", nil)
		if err != nil {
			failBuild(err, "failed to generate admin client credentials")
		}
		d.log.Info("generated admin service client key pair", log.String("cert", clientCertFile), log.String("key", clientKeyFile))
		if clientsCerts, err = os.ReadFile(clientCertFile); err != nil {
			failBuild(err, "failed to read auto-generate client certificate")
		}
		if err = os.WriteFile(clientsFile, clientsCerts, 0644); err != nil {
			failBuild(err, "failed to write client CAs file")
		}
		d.log.Info("generated admin service client CAs file", log.String("file", clientsFile))
	} else {
		d.log.Info("No admin client CAs file. Use kwil-admin's node gen-auth-key command to generate")
	}

	if len(clientsCerts) > 0 && !caCertPool.AppendCertsFromPEM(clientsCerts) {
		failBuild(err, "invalid client CAs file")
	}

	// TLS configuration for mTLS (mutual TLS) protocol-level authentication
	return &tls.Config{
		Certificates: []tls.Certificate{*d.keypair},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    caCertPool,
	}
}

func buildJRPCAdminServer(d *coreDependencies) *rpcserver.Server {
	var wantTLS bool
	addr := d.cfg.AppCfg.AdminListenAddress
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		if strings.Contains(err.Error(), "missing port in address") {
			host = addr
			port = "8485"
		} else if strings.Contains(err.Error(), "too many colons in address") {
			u, err := neturl.Parse(addr)
			if err != nil {
				failBuild(err, "unknown admin service address "+addr)
			}
			host, port = u.Hostname(), u.Port()
			wantTLS = u.Scheme == "https"
		} else {
			failBuild(err, "unknown admin service address "+addr)
		}
	}

	var opts []rpcserver.Opt

	adminPass := d.cfg.AppCfg.AdminRPCPass
	if adminPass != "" {
		opts = append(opts, rpcserver.WithPass(adminPass))
	}

	// Require TLS only if not UNIX or not loopback TCP interface.
	if isUNIX := strings.HasPrefix(host, "/"); isUNIX {
		addr = host
		// no port and no TLS
		if wantTLS {
			failBuild(errors.New("unix socket with TLS is not supported"), "")
		}
	} else { // TCP
		addr = net.JoinHostPort(host, port)

		var loopback bool
		if netAddr, err := net.ResolveIPAddr("ip", host); err != nil {
			d.log.Warn("unresolvable host, assuming not loopback, but will likely fail to listen",
				log.String("host", host), log.Error(err))
		} else { // e.g. "localhost" usually resolves to a loopback IP address
			loopback = netAddr.IP.IsLoopback()
		}
		if !loopback || wantTLS { // use TLS for encryption, maybe also client auth
			if d.cfg.AppCfg.NoTLS {
				d.log.Warn("disabling TLS on non-loopback admin service listen address",
					log.String("addr", addr), log.Bool("with_password", adminPass != ""))
			} else {
				withClientAuth := adminPass == "" // no basic http auth => use transport layer auth
				opts = append(opts, rpcserver.WithTLS(tlsConfig(d, withClientAuth)))
			}
		}
	}

	// Note that rpcserver.WithPass is not mutually exclusive with TLS in
	// general, only mutual TLS. It could be a simpler alternative to mutual
	// TLS, or just coupled with TLS termination on a local reverse proxy.

	jsonRPCAdminServer, err := rpcserver.NewServer(addr, *d.log.Named("admin-jsonrpc-server"),
		opts...)
	if err != nil {
		failBuild(err, "unable to create json-rpc server")
	}

	return jsonRPCAdminServer
}

func fileExists(name string) bool {
	_, err := os.Stat(name)
	return err == nil
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
	lis, err := net.Listen("tcp", "127.0.0.1:0") // listen on random available port
	if err != nil {
		failBuild(err, "failed to build grpc server")
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

func buildGatewayServer(d *coreDependencies, gRPCAddr string) *gateway.GatewayServer {
	gw, err := gateway.NewGateway(d.ctx, d.cfg.AppCfg.HTTPListenAddress,
		gateway.WithLogger(*d.log.Named("gateway")),
		gateway.WithMiddleware(cors.MCors([]string{})),
		gateway.WithGrpcService(gRPCAddr, txpb.RegisterTxServiceHandlerFromEndpoint),
		gateway.WithGrpcService(gRPCAddr, functionpb.RegisterFunctionServiceHandlerFromEndpoint),
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
		failBuild(err, "failed to open comet node KV store")
	}
	closer.addCloser(db.Close, "closing signing store")

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

	node, err := cometbft.NewCometBftNode(d.ctx, abciApp, nodeCfg, genDoc, d.privKey,
		readWriter, &d.log)
	if err != nil {
		failBuild(err, "failed to build comet node")
	}

	return node
}

// panicErr is the type given to panic from failBuild so that the wrapped error
// may be type-inspected.
type panicErr struct {
	err error
	msg string
}

func (pe panicErr) String() string {
	return pe.msg
}

func (pe panicErr) Error() string { // error interface
	return pe.msg
}

func (pe panicErr) Unwrap() error {
	return pe.err
}

func failBuild(err error, msg string) {
	if err != nil {
		panic(panicErr{
			err: err,
			msg: fmt.Sprintf("%s: %s", msg, err),
		})
	}
}

func buildListenerManager(d *coreDependencies, ev *voting.EventStore, node *cometbft.CometBftNode, txapp *txapp.TxApp) *listeners.ListenerManager {
	return listeners.NewListenerManager(d.cfg.AppCfg.Extensions, ev, node, d.privKey.PubKey().Bytes(), txapp, *d.log.Named("listener-manager"))
}
