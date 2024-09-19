package server

import (
	"bytes"
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

	abciTypes "github.com/cometbft/cometbft/abci/types"
	cmtEd "github.com/cometbft/cometbft/crypto/ed25519"
	cmtlocal "github.com/cometbft/cometbft/rpc/client/local"

	"github.com/kwilteam/kwil-db/cmd"
	kwildcfg "github.com/kwilteam/kwil-db/cmd/kwild/config"
	"github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/common/chain"
	config "github.com/kwilteam/kwil-db/common/config"
	"github.com/kwilteam/kwil-db/common/sql"
	"github.com/kwilteam/kwil-db/core/crypto"
	"github.com/kwilteam/kwil-db/core/crypto/auth"
	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/core/rpc/transport"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/internal/abci"
	"github.com/kwilteam/kwil-db/internal/abci/cometbft"
	"github.com/kwilteam/kwil-db/internal/abci/meta"
	"github.com/kwilteam/kwil-db/internal/accounts"
	"github.com/kwilteam/kwil-db/internal/engine/execution"
	"github.com/kwilteam/kwil-db/internal/kv"
	"github.com/kwilteam/kwil-db/internal/kv/badger"
	"github.com/kwilteam/kwil-db/internal/listeners"
	"github.com/kwilteam/kwil-db/internal/migrations"
	rpcserver "github.com/kwilteam/kwil-db/internal/services/jsonrpc"
	"github.com/kwilteam/kwil-db/internal/services/jsonrpc/adminsvc"
	"github.com/kwilteam/kwil-db/internal/services/jsonrpc/funcsvc"
	usersvc "github.com/kwilteam/kwil-db/internal/services/jsonrpc/usersvc"
	"github.com/kwilteam/kwil-db/internal/sql/pg"
	"github.com/kwilteam/kwil-db/internal/statesync"
	"github.com/kwilteam/kwil-db/internal/txapp"
	"github.com/kwilteam/kwil-db/internal/voting"
	"github.com/kwilteam/kwil-db/internal/voting/broadcast"
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
	badgerPath := filepath.Join(d.cfg.RootDir, abciDirName, kwildcfg.ABCIInfoSubDirName)
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

	snapshotter := buildSnapshotter(d)
	statesyncer := buildStatesyncer(d, db)

	p2p := buildPeers(d, closers)

	// this is a hack
	// we need the cometbft client to broadcast txs.
	// in order to get this, we need the comet node
	// to get the comet node, we need the abci app
	// to get the abci app, we need the tx router
	// but the tx router needs the cometbft client
	txApp := buildTxApp(d, db, e, ev)

	// migrator
	migrator := buildMigrator(d, db, txApp)
	abciApp := buildAbci(d, db, txApp, snapshotter, statesyncer, p2p, migrator, closers)

	// NOTE: buildCometNode immediately starts talking to the abciApp and
	// replaying blocks (and using atomic db tx commits), i.e. calling
	// FinalizeBlock+Commit. This is not just a constructor, sadly.
	cometBftNode := buildCometNode(d, closers, abciApp)

	// Give abci p2p module access to removing peers
	p2p.SetRemovePeerFn(cometBftNode.RemovePeer)

	// Give migrator access to the consensus params getter
	migrator.SetConsensusParamsGetter(cometBftNode.ConsensusParams)

	cometBftClient := buildCometBftClient(cometBftNode)
	wrappedCmtClient := &wrappedCometBFTClient{
		cl:    cometBftClient,
		cache: abciApp,
	}
	abciApp.SetReplayStatusChecker(cometBftNode.IsCatchup)

	eventBroadcaster := buildEventBroadcaster(d, ev, wrappedCmtClient, txApp)
	abciApp.SetEventBroadcaster(eventBroadcaster.RunBroadcast)

	// listener manager
	listeners := buildListenerManager(d, ev, cometBftNode, txApp, db)

	// user service and server
	rpcSvcLogger := increaseLogLevel("user-json-svc", &d.log, d.cfg.Logging.RPCLevel)
	rpcServerLogger := increaseLogLevel("user-jsonrpc-server", &d.log, d.cfg.Logging.RPCLevel)

	if d.cfg.AppConfig.RPCMaxReqSize < d.cfg.ChainConfig.Mempool.MaxTxBytes {
		d.log.Warnf("RPC request size limit (%d) is less than maximium transaction size (%d)",
			d.cfg.AppConfig.RPCMaxReqSize, d.cfg.ChainConfig.Mempool.MaxTxBytes)
	}

	// Base a long block delay on the configured consensus timeouts
	//  e.g. 6 + 2 + 2 + 3 = 13 sec longest single round
	//  multiple by excessive consensus round count, like 6 => 78 sec
	totalConsensusTimeouts := d.cfg.ChainConfig.Consensus.TimeoutCommit + d.cfg.ChainConfig.Consensus.TimeoutPrecommit +
		d.cfg.ChainConfig.Consensus.TimeoutPrevote + d.cfg.ChainConfig.Consensus.TimeoutPropose

	jsonRPCTxSvc := usersvc.NewService(db, e, wrappedCmtClient, txApp, abciApp, migrator,
		*rpcSvcLogger, usersvc.WithReadTxTimeout(time.Duration(d.cfg.AppConfig.ReadTxTimeout)),
		usersvc.WithPrivateMode(d.cfg.AppConfig.PrivateRPC),
		usersvc.WithChallengeExpiry(time.Duration(d.cfg.AppConfig.ChallengeExpiry)),
		usersvc.WithChallengeRateLimit(d.cfg.AppConfig.ChallengeRateLimit),
		usersvc.WithBlockAgeHealth(6*totalConsensusTimeouts.Dur()))

	jsonRPCServer, err := rpcserver.NewServer(d.cfg.AppConfig.JSONRPCListenAddress,
		*rpcServerLogger, rpcserver.WithTimeout(time.Duration(d.cfg.AppConfig.RPCTimeout)),
		rpcserver.WithReqSizeLimit(d.cfg.AppConfig.RPCMaxReqSize),
		rpcserver.WithCORS(), rpcserver.WithServerInfo(&usersvc.SpecInfo))
	if err != nil {
		failBuild(err, "unable to create json-rpc server")
	}
	jsonRPCServer.RegisterSvc(jsonRPCTxSvc)
	jsonRPCServer.RegisterSvc(&funcsvc.Service{})

	// admin service and server
	signer := buildSigner(d)
	jsonAdminSvc := adminsvc.NewService(db, wrappedCmtClient, txApp, abciApp, p2p, migrator, signer, d.cfg,
		d.genesisCfg.ChainID, *d.log.Named("admin-json-svc"))
	jsonRPCAdminServer := buildJRPCAdminServer(d)
	jsonRPCAdminServer.RegisterSvc(jsonAdminSvc)
	jsonRPCAdminServer.RegisterSvc(jsonRPCTxSvc)
	jsonRPCAdminServer.RegisterSvc(&funcsvc.Service{})

	return &Server{
		jsonRPCServer:      jsonRPCServer,
		jsonRPCAdminServer: jsonRPCAdminServer,
		cometBftNode:       cometBftNode,
		listenerManager:    listeners,
		log:                *d.log.Named("server"),
		closers:            closers,
		cfg:                d.cfg,
		dbCtx:              db,
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
	genesisCfg *chain.GenesisConfig
	privKey    cmtEd.PrivKey
	log        log.Logger
	dbOpener   dbOpener
	poolOpener poolOpener
	keypair    *tls.Certificate
}

// service returns a common.Service with the given logger name
func (c *coreDependencies) service(loggerName string) *common.Service {
	return &common.Service{
		Logger:           c.log.Named(loggerName).Sugar(),
		GenesisConfig:    c.genesisCfg,
		LocalConfig:      c.cfg,
		Identity:         c.privKey.PubKey().Bytes(),
		ExtensionConfigs: make(map[string]map[string]string),
	}
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

func buildTxApp(d *coreDependencies, db *pg.DB, engine *execution.GlobalContext, ev *voting.EventStore) *txapp.TxApp {

	txApp, err := txapp.NewTxApp(d.ctx, db, engine, buildSigner(d), ev, d.service("txapp"))
	if err != nil {
		failBuild(err, "failed to build new TxApp")
	}
	return txApp
}

func buildPeers(d *coreDependencies, closers *closeFuncs) *cometbft.PeerWhiteList {
	var whitelistPeers []string

	db, err := d.poolOpener(d.ctx, d.cfg.AppConfig.DBName, 10)
	if err != nil {
		failBuild(err, "failed to build event store")
	}
	closers.addCloser(db.Close, "closing event store")

	if d.cfg.ChainConfig.P2P.WhitelistPeers != "" {
		whitelistPeers = strings.Split(d.cfg.ChainConfig.P2P.WhitelistPeers, ",")
	}

	// Load the validators from the database if the database is already initialized
	// If the database is not initialized, the validators will be loaded from the genesis file
	vals, err := voting.GetValidators(d.ctx, db)
	if err != nil {
		failBuild(err, "failed to load validators")
	}
	if len(vals) > 0 {
		for _, v := range vals {
			addr, err := cometbft.PubkeyToAddr(v.PubKey)
			if err != nil {
				failBuild(err, "failed to convert pubkey to address")
			}
			whitelistPeers = append(whitelistPeers, addr)
		}
	} else {
		// Load the validators from the genesis file
		for _, v := range d.genesisCfg.Validators {
			addr, err := cometbft.PubkeyToAddr(v.PubKey)
			if err != nil {
				failBuild(err, "failed to convert pubkey to address")
			}
			whitelistPeers = append(whitelistPeers, addr)
		}
	}

	nodePubKey := d.privKey.PubKey().Bytes()
	nodeID, err := cometbft.PubkeyToAddr(nodePubKey)
	if err != nil {
		failBuild(err, "failed to convert pubkey to address")
	}

	// Add the nodes whose validator join requests have been approved by the node
	// to the whitelist
	approvedValidators, err := getPendingValidatorsApprovedByNode(d.ctx, db, d.privKey.PubKey().Bytes())
	if err != nil {
		failBuild(err, "failed to get approved validators")
	}
	for _, v := range approvedValidators {
		addr, err := cometbft.PubkeyToAddr(v.PubKey)
		if err != nil {
			failBuild(err, "failed to convert pubkey to address")
		}
		whitelistPeers = append(whitelistPeers, addr)
	}

	// Add seeds and persistent peers to the whitelist
	if d.cfg.ChainConfig.P2P.PersistentPeers != "" {
		persistentPeers := strings.Split(d.cfg.ChainConfig.P2P.PersistentPeers, ",")
		for _, p := range persistentPeers {
			// split the persistent peer string into node ID and host:port
			parts := strings.Split(p, "@")
			if len(parts) != 2 {
				failBuild(nil, "invalid persistent peer format")
			}
			whitelistPeers = append(whitelistPeers, parts[0])
		}
	}

	if d.cfg.ChainConfig.P2P.Seeds != "" {
		seeds := strings.Split(d.cfg.ChainConfig.P2P.Seeds, ",")
		for _, s := range seeds {
			// split the seed string into node ID and host:port
			parts := strings.Split(s, "@")
			if len(parts) != 2 {
				failBuild(nil, "invalid seed format")
			}
			whitelistPeers = append(whitelistPeers, parts[0])
		}
	}

	// Initialize the Peers with the whitelist peers.
	peers, err := cometbft.P2PInit(d.ctx, db, d.cfg.ChainConfig.P2P.PrivateMode, whitelistPeers, nodeID)
	if err != nil {
		failBuild(err, "failed to initialize P2P store")
	}

	return peers
}

func getPendingValidatorsApprovedByNode(ctx context.Context, db sql.ReadTxMaker, pubKey []byte) ([]*types.Validator, error) {
	readTx, err := db.BeginReadTx(ctx)
	if err != nil {
		return nil, err
	}
	defer readTx.Rollback(ctx)

	// get the pending validators join resolutions
	resolutions, err := voting.GetResolutionsByType(ctx, readTx, voting.ValidatorJoinEventType)
	if err != nil {
		return nil, err
	}

	var validators []*types.Validator

	// get the pending validators remove resolutions
	for _, res := range resolutions {
		for _, voter := range res.Voters {
			if bytes.Equal(voter.PubKey, pubKey) {
				validators = append(validators, voter)
			}
		}
	}

	return validators, nil
}

func buildMigrator(d *coreDependencies, db *pg.DB, txApp *txapp.TxApp) *migrations.Migrator {
	cfg := d.cfg.AppConfig
	migrationsDir := filepath.Join(d.cfg.RootDir, kwildcfg.MigrationsDirName)

	err := os.MkdirAll(filepath.Join(migrationsDir, kwildcfg.ChangesetsDirName), 0755)
	if err != nil {
		failBuild(err, "failed to create changesets directory")
	}

	err = os.MkdirAll(filepath.Join(migrationsDir, cmd.DefaultConfig().AppConfig.Snapshots.SnapshotDir), 0755)
	if err != nil {
		failBuild(err, "failed to create migrations snapshots directory")
	}

	// snapshot store
	dbCfg := &statesync.DBConfig{
		DBUser: cfg.DBUser,
		DBPass: cfg.DBPass,
		DBHost: cfg.DBHost,
		DBPort: cfg.DBPort,
		DBName: cfg.DBName,
	}

	snapshotCfg := &statesync.SnapshotConfig{
		SnapshotDir:     filepath.Join(migrationsDir, cmd.DefaultConfig().AppConfig.Snapshots.SnapshotDir),
		RecurringHeight: 0,
		MaxSnapshots:    1, // only one snapshot is needed for network migrations, taken at the activation height
		MaxRowSize:      cfg.Snapshots.MaxRowSize,
	}

	ss, err := statesync.NewSnapshotStore(snapshotCfg, dbCfg, *d.log.Named("migrations-snapshots"))
	if err != nil {
		failBuild(err, "failed to build snapshot store for migrations")
	}

	migrator, err := migrations.SetupMigrator(d.ctx, db, ss, txApp, migrationsDir, *d.log.Named("migrator"))
	if err != nil {
		failBuild(err, "failed to build migrator")
	}

	return migrator
}

func buildAbci(d *coreDependencies, db *pg.DB, txApp abci.TxApp, snapshotter *statesync.SnapshotStore, statesyncer *statesync.StateSyncer, p2p *cometbft.PeerWhiteList, migrator *migrations.Migrator, closers *closeFuncs) *abci.AbciApp {
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
		ForkHeights:        d.genesisCfg.ForkHeights,
	}
	app, err := abci.NewAbciApp(d.ctx, cfg, sh, ss, txApp,
		d.genesisCfg.ConsensusParams, p2p, migrator, db, *d.log.Named("abci"))
	if err != nil {
		failBuild(err, "failed to build ABCI application")
	}

	closers.addCloser(app.Close, "closing ABCI app")

	return app
}

func buildEventBroadcaster(d *coreDependencies, ev broadcast.EventStore, b broadcast.Broadcaster, txapp *txapp.TxApp) *broadcast.EventBroadcaster {
	return broadcast.NewEventBroadcaster(ev, b, txapp, buildSigner(d), d.genesisCfg.ChainID, d.genesisCfg.ConsensusParams.Votes.MaxVotesPerTx, *d.log.Named("event-broadcaster"))
}

func buildEventStore(d *coreDependencies, closers *closeFuncs) *voting.EventStore {
	// NOTE: we're using the same postgresql database, but isolated pg schema.
	db, err := d.poolOpener(d.ctx, d.cfg.AppConfig.DBName, 10)
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
	fromSnapshot := restoreDB(d)

	db, err := d.dbOpener(d.ctx, d.cfg.AppConfig.DBName, 24)
	if err != nil {
		failBuild(err, "kwild database open failed")
	}
	closer.addCloser(db.Close, "closing main DB")

	if fromSnapshot {
		// readjust the expiry heights of all the pending resolutions after snapshot restore for Zero-downtime migrations
		// snapshot tool handles the migration expiry height readjustment for offline migrations
		adjustExpiration := false
		startHeight := d.genesisCfg.ConsensusParams.Migration.StartHeight
		if d.cfg.AppConfig.MigrateFrom != "" && startHeight != 0 {
			adjustExpiration = true
		}

		err = migrations.CleanupResolutionsAfterMigration(d.ctx, db, adjustExpiration, startHeight)
		if err != nil {
			failBuild(err, "failed to cleanup resolutions after snapshot restore")
		}
	}
	return db
}

// restoreDB restores the database from a snapshot if the genesis apphash is specified.
// Genesis apphash ensures that all the nodes in the network start from the same state.
// Genesis apphash should match the hash of the snapshot file.
// Snapshot file can be compressed or uncompressed represented by .gz extension.
// DB restoration from snapshot is skipped in the following scenarios:
//   - If the DB is already initialized (i.e this is not a new node)
//   - If the genesis apphash is not specified
//   - If statesync is enabled. Statesync will take care of syncing the database
//     to the network state using statesync snapshots.
func restoreDB(d *coreDependencies) bool {
	if d.cfg.ChainConfig.StateSync.Enable || len(d.genesisCfg.DataAppHash) == 0 || isDbInitialized(d) {
		return false
	}

	genCfg := d.genesisCfg
	appCfg := d.cfg.AppConfig
	// DB is uninitialized and genesis apphash is specified.
	// DB is supposed to be restored from the snapshot.
	// Ensure that the snapshot file exists and the snapshot hash matches the genesis apphash.

	// Ensure that the snapshot file exists, if node is supposed to start with a snapshot state
	if genCfg.DataAppHash != nil && appCfg.GenesisState == "" {
		failBuild(nil, "snapshot file not provided")
	}

	// Snapshot file exists
	snapFile, err := os.Open(appCfg.GenesisState)
	if err != nil {
		failBuild(err, "failed to open snapshot file")
	}

	// Check if the snapshot file is compressed, if yes decompress it
	var reader io.Reader
	if strings.HasSuffix(appCfg.GenesisState, ".gz") {
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
	return true
}

// isDbInitialized checks if the database is already initialized.
func isDbInitialized(d *coreDependencies) bool {
	db, err := d.poolOpener(d.ctx, d.cfg.AppConfig.DBName, 3)
	if err != nil {
		failBuild(err, "kwild database open failed")
	}
	defer db.Close()

	// Check if the database is empty or initialized previously
	// If the database is empty, we need to restore the database from the snapshot
	vals, _ := voting.GetValidators(d.ctx, db)
	// ERROR: relation "kwild_voting.voters" does not exist
	// assumption that error is due to the missing table and schema.
	return len(vals) > 0
}

func buildEngine(d *coreDependencies, db *pg.DB) *execution.GlobalContext {
	extensions, err := getExtensions(d.ctx, d.cfg.AppConfig.ExtensionEndpoints)
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

	eng, err := execution.NewGlobalContext(d.ctx, tx, extensions, d.service("engine"))
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
	cfg := d.cfg.AppConfig
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
		MaxRowSize:      cfg.Snapshots.MaxRowSize,
	}

	ss, err := statesync.NewSnapshotStore(snapshotCfg, dbCfg, *d.log.Named("snapshot-store"))
	if err != nil {
		failBuild(err, "failed to build snapshot store")
	}
	return ss
}

func buildStatesyncer(d *coreDependencies, db sql.ReadTxMaker) *statesync.StateSyncer {
	if !d.cfg.ChainConfig.StateSync.Enable {
		return nil
	}

	cfg := d.cfg.AppConfig

	dbCfg := &statesync.DBConfig{
		DBUser: cfg.DBUser,
		DBPass: cfg.DBPass,
		DBHost: cfg.DBHost,
		DBPort: cfg.DBPort,
		DBName: cfg.DBName,
	}

	providers := strings.Split(d.cfg.ChainConfig.StateSync.RPCServers, ",")

	if len(providers) == 0 {
		failBuild(nil, "failed to configure state syncer, no remote servers provided.")
	}

	if len(providers) == 1 {
		// Duplicating the same provider to satisfy cometbft statesync requirement of having at least 2 providers.
		// Statesynce module doesn't have the same requirements and
		// can work with a single provider (providers are passed as is)
		d.cfg.ChainConfig.StateSync.RPCServers += "," + providers[0]
	}

	configDone := false
	for _, p := range providers {
		clt, err := statesync.ChainRPCClient(p)
		if err != nil {
			d.log.Warnf("failed to make chain RPC client to snap provider: %v", err)
			continue
		}

		// Try to fetch the status of the remote server. Set a timeout on the
		// initial RPC so this doesn't hang for a very long time. Although
		// arbitrary, 10s is a reasonable time out for a http server regardless
		// of the location and network routes.
		ctx, cancel := context.WithTimeout(d.ctx, 10*time.Second)

		// we will first get the latest snapshot height that the trusted node has
		latestSnapshot, err := statesync.GetLatestSnapshotInfo(ctx, clt)
		if err != nil {
			cancel()
			d.log.Warnf("failed to get latest snapshot from snap provider: %v", err)
			continue
		}

		latestHeight := int64(latestSnapshot.Height)
		res, err := clt.Header(ctx, &latestHeight)
		if err != nil {
			cancel()
			d.log.Warnf("failed to get header from snap provider: %v", err)
			continue
		}
		cancel()

		// If the remote server is in the same chain, we can trust it.
		if res.Header.ChainID != d.genesisCfg.ChainID {
			d.log.Warnf("snap provider has wrong chain ID: want %v, got %v", d.genesisCfg.ChainID, res.Header.ChainID)
			continue
		}

		if res.Header.Height == 0 {
			d.log.Warnf("zero height from provider %v", p)
			continue
		}

		// Get the trust height and trust hash from the remote server
		d.cfg.ChainConfig.StateSync.TrustHeight = res.Header.Height
		d.cfg.ChainConfig.StateSync.TrustHash = res.Header.Hash().String()

		d.log.Infof("Provider %q: trust height %v, hash %v", p, d.cfg.ChainConfig.StateSync.TrustHeight, d.cfg.ChainConfig.StateSync.TrustHash)

		configDone = true

		break
	}

	if !configDone {
		failBuild(nil, "failed to configure state syncer, failed to fetch trust options from the remote server.")
	}

	// create state syncer
	return statesync.NewStateSyncer(d.ctx, dbCfg, d.cfg.ChainConfig.StateSync.SnapshotDir,
		providers, db, *d.log.Named("state-syncer"))
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
	addr := d.cfg.AppConfig.AdminListenAddress
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

	opts := []rpcserver.Opt{rpcserver.WithTimeout(10 * time.Minute)} // this is an administrator

	adminPass := d.cfg.AppConfig.AdminRPCPass
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
			if d.cfg.AppConfig.NoTLS {
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
	opts = append(opts, rpcserver.WithServerInfo(&adminsvc.SpecInfo))
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

func buildCometBftClient(cometBftNode *cometbft.CometBftNode) *cmtlocal.Local {
	return cmtlocal.New(cometBftNode.Node)
}

func buildCometNode(d *coreDependencies, closer *closeFuncs, abciApp abciTypes.Application) *cometbft.CometBftNode {
	// for now, I'm just using a KV store for my atomic commit.  This probably is not ideal; a file may be better
	// I'm simply using this because we know it fsyncs the data to disk
	db, err := badger.NewBadgerDB(d.ctx, filepath.Join(d.cfg.RootDir, signingDirName), &badger.Options{
		GuaranteeFSync: true,
		Logger:         *increaseLogLevel("private-validator-signature-store", &d.log, log.WarnLevel.String()), // badger is too noisy for an internal component
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

	nodeLogger := increaseLogLevel("cometbft", &d.log, d.cfg.Logging.ConsensusLevel)
	node, err := cometbft.NewCometBftNode(d.ctx, abciApp, nodeCfg, genDoc, d.privKey,
		readWriter, nodeLogger)
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
	if err == nil {
		panic(panicErr{
			err: errors.New(msg),
			msg: msg,
		})
	}

	panic(panicErr{
		err: err,
		msg: fmt.Sprintf("%s: %s", msg, err),
	})
}

func buildListenerManager(d *coreDependencies, ev *voting.EventStore, node *cometbft.CometBftNode, txapp *txapp.TxApp, db sql.ReadTxMaker) *listeners.ListenerManager {
	vr := &validatorReader{db: db, txApp: txapp}
	return listeners.NewListenerManager(d.service("listener-manager"), ev, node, vr)
}

// validatorReader reads the validator set from the chain state.
type validatorReader struct {
	db    sql.ReadTxMaker
	txApp *txapp.TxApp
}

func (v *validatorReader) GetValidators(ctx context.Context) ([]*types.Validator, error) {
	cached, ok := v.txApp.CachedValidators()
	if ok {
		return cached, nil
	}

	// if we don't have a cached validator set, we need to fetch it from the db
	readTx, err := v.db.BeginReadTx(ctx)
	if err != nil {
		return nil, err
	}
	defer readTx.Rollback(ctx)

	return v.txApp.GetValidators(ctx, readTx)
}

func (v *validatorReader) SubscribeValidators() <-chan []*types.Validator {
	return v.txApp.SubscribeValidators()
}
