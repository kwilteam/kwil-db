package node

import (
	"compress/gzip"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/config"
	"github.com/kwilteam/kwil-db/core/crypto"
	"github.com/kwilteam/kwil-db/core/crypto/auth"
	ktypes "github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/extensions/precompiles"
	"github.com/kwilteam/kwil-db/node"
	"github.com/kwilteam/kwil-db/node/accounts"
	blockprocessor "github.com/kwilteam/kwil-db/node/block_processor"
	"github.com/kwilteam/kwil-db/node/consensus"
	"github.com/kwilteam/kwil-db/node/engine/interpreter"
	"github.com/kwilteam/kwil-db/node/listeners"
	"github.com/kwilteam/kwil-db/node/mempool"
	"github.com/kwilteam/kwil-db/node/meta"
	"github.com/kwilteam/kwil-db/node/migrations"
	"github.com/kwilteam/kwil-db/node/pg"
	"github.com/kwilteam/kwil-db/node/snapshotter"
	"github.com/kwilteam/kwil-db/node/store"
	"github.com/kwilteam/kwil-db/node/txapp"
	"github.com/kwilteam/kwil-db/node/types/sql"
	"github.com/kwilteam/kwil-db/node/voting"

	rpcserver "github.com/kwilteam/kwil-db/node/services/jsonrpc"
	"github.com/kwilteam/kwil-db/node/services/jsonrpc/adminsvc"
	"github.com/kwilteam/kwil-db/node/services/jsonrpc/chainsvc"
	"github.com/kwilteam/kwil-db/node/services/jsonrpc/funcsvc"
	"github.com/kwilteam/kwil-db/node/services/jsonrpc/usersvc"
)

func buildServer(ctx context.Context, d *coreDependencies) *server {
	closers := &closeFuncs{
		closers: []func() error{}, // logger.Close is not in here; do it in a defer in Start
		logger:  d.logger,
	}
	d.closers = closers

	valSet := make(map[string]ktypes.Validator)
	for _, v := range d.genesisCfg.Validators {
		valSet[hex.EncodeToString(v.PubKey)] = ktypes.Validator{
			PubKey: v.PubKey,
			Power:  v.Power,
		}
	}

	// Initialize DB
	db := buildDB(ctx, d, closers)

	// metastore
	buildMetaStore(ctx, db)

	// BlockStore
	bs := buildBlockStore(d, closers)

	e := buildEngine(d, db)

	// Mempool
	mp := mempool.New()

	// accounts
	accounts := buildAccountStore(ctx, d, db)

	// eventstore, votestore
	es, vs := buildVoteStore(ctx, d, closers) // ev, vs

	// TxAPP
	txApp := buildTxApp(ctx, d, db, accounts, vs, e)

	// Snapshot Store
	ss := buildSnapshotStore(d)

	// Migrator
	migrator := buildMigrator(d, db, accounts, vs)

	// BlockProcessor
	bp := buildBlockProcessor(ctx, d, db, txApp, accounts, vs, ss, es, migrator, bs)

	// Consensus
	ce := buildConsensusEngine(ctx, d, db, mp, bs, bp, valSet)

	// Node
	node := buildNode(d, mp, bs, ce, ss, db, bp)

	// listeners
	lm := buildListenerManager(d, es, bp, node)

	// RPC Services
	rpcSvcLogger := d.logger.New("USER")
	jsonRPCTxSvc := usersvc.NewService(db, e, node, bp, vs, migrator, rpcSvcLogger,
		usersvc.WithReadTxTimeout(time.Duration(d.cfg.DB.ReadTxTimeout)),
		usersvc.WithPrivateMode(d.cfg.RPC.Private),
		usersvc.WithChallengeExpiry(time.Duration(d.cfg.RPC.ChallengeExpiry)),
		usersvc.WithChallengeRateLimit(d.cfg.RPC.ChallengeRateLimit),
		// usersvc.WithBlockAgeHealth(6*totalConsensusTimeouts.Dur()),
	)

	rpcServerLogger := d.logger.New("RPC")
	jsonRPCServer, err := rpcserver.NewServer(d.cfg.RPC.ListenAddress,
		rpcServerLogger, rpcserver.WithTimeout(time.Duration(d.cfg.RPC.Timeout)),
		rpcserver.WithReqSizeLimit(d.cfg.RPC.MaxReqSize),
		rpcserver.WithCORS(), rpcserver.WithServerInfo(&usersvc.SpecInfo),
		rpcserver.WithMetricsNamespace("kwil_json_rpc_user_server"))
	if err != nil {
		failBuild(err, "unable to create json-rpc server")
	}
	jsonRPCServer.RegisterSvc(jsonRPCTxSvc)
	jsonRPCServer.RegisterSvc(&funcsvc.Service{})

	var jsonRPCAdminServer *rpcserver.Server
	if d.cfg.Admin.Enable {
		// admin service and server
		adminServerLogger := d.logger.New("ADMIN")
		// The admin service uses a client-style signer rather than just a private
		// key because it is used to sign transactions and provide an Identity for
		// account information (nonce and balance).
		txSigner := auth.GetNodeSigner(d.privKey)
		jsonAdminSvc := adminsvc.NewService(db, node, bp, vs, node.Whitelister(),
			txSigner, d.cfg, d.genesisCfg.ChainID, adminServerLogger)
		jsonRPCAdminServer = buildJRPCAdminServer(d)
		jsonRPCAdminServer.RegisterSvc(jsonAdminSvc)
		jsonRPCAdminServer.RegisterSvc(jsonRPCTxSvc)
		jsonRPCAdminServer.RegisterSvc(&funcsvc.Service{})
	}

	chainRpcSvcLogger := d.logger.New("CHAIN")
	jsonChainSvc := chainsvc.NewService(chainRpcSvcLogger, node, vs, d.genesisCfg)
	jsonRPCServer.RegisterSvc(jsonChainSvc)

	s := &server{
		cfg:                d.cfg,
		closers:            closers,
		node:               node,
		ce:                 ce,
		listeners:          lm,
		jsonRPCServer:      jsonRPCServer,
		jsonRPCAdminServer: jsonRPCAdminServer,
		dbCtx:              db,
		log:                d.logger,
	}

	return s
}

func buildDB(ctx context.Context, d *coreDependencies, closers *closeFuncs) *pg.DB {
	pg.UseLogger(d.logger.New("PG"))

	// TODO: restore from snapshots
	fromSnapshot := restoreDB(d)

	db, err := d.dbOpener(ctx, d.cfg.DB.DBName, d.cfg.DB.MaxConns)
	if err != nil {
		failBuild(err, "failed to open kwild postgres database")
	}
	closers.addCloser(db.Close, "Closing application DB")

	if fromSnapshot {
		d.logger.Info("DB restored from snapshot", "snapshot", d.cfg.GenesisState)
		// readjust the expiry heights of all the pending resolutions after snapshot restore for Zero-downtime migrations
		// snapshot tool handles the migration expiry height readjustment for offline migrations
		// adjustExpiration := false
		// startHeight := d.genesisCfg.ConsensusParams.Migration.StartHeight
		// if d.cfg.MigrationConfig.Enable && startHeight != 0 {
		// 	adjustExpiration = true
		// }

		// err = migrations.CleanupResolutionsAfterMigration(d.ctx, db, adjustExpiration, startHeight)
		// if err != nil {
		// 	failBuild(err, "failed to cleanup resolutions after snapshot restore")
		// }

		if err = db.EnsureFullReplicaIdentityDatasets(d.ctx); err != nil {
			failBuild(err, "failed enable full replica identity on user datasets")
		}
	}
	return db
}

// restoreDB restores the database from a snapshot if the genesis apphash is specified.
// StateHash in the genesis config ensures that all the nodes in the network start from the same state.
// StateHash in the genesis config should match the hash of the snapshot file.
// Snapshot file can be compressed or uncompressed represented by .gz extension.
// DB restoration from snapshot is skipped in the following scenarios:
//   - If the DB is already initialized (i.e this is not a new node)
//   - If the StateHash is not set in the genesis config
//   - If statesync is enabled. Statesync will take care of syncing the database
//     to the network state using statesync snapshots.
//
// returns true if the DB was restored from snapshot, false otherwise.
func restoreDB(d *coreDependencies) bool {
	if d.cfg.StateSync.Enable || len(d.genesisCfg.StateHash) == 0 || isDbInitialized(d) {
		return false
	}

	genCfg := d.genesisCfg
	appCfg := d.cfg

	// DB is uninitialized and genesis statehash is set, so db should be restored from snapshot.
	// Ensure that the snapshot file exists and the snapshot hash matches the genesis apphash.

	if genCfg.StateHash != nil && appCfg.GenesisState == "" {
		failBuild(nil, "snapshot file not provided")
	}

	// Snapshot file exists
	genFileName, err := node.ExpandPath(appCfg.GenesisState)
	if err != nil {
		failBuild(err, "failed to expand genesis state path")
	}

	snapFile, err := os.Open(genFileName)
	if err != nil {
		failBuild(err, "failed to open genesis state file")
	}
	defer snapFile.Close()

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
	err = node.RestoreDB(d.ctx, reader, &appCfg.DB, genCfg.StateHash, d.logger)
	if err != nil {
		failBuild(err, "failed to restore DB from snapshot")
	}
	return true
}

// isDbInitialized checks if the database is already initialized.
func isDbInitialized(d *coreDependencies) bool {
	db, err := d.poolOpener(d.ctx, d.cfg.DB.DBName, 3)
	if err != nil {
		failBuild(err, "kwild database open failed")
	}
	defer db.Close()

	// Check if the kwild_voting schema exists
	exists, err := schemaExists(d.ctx, db, "kwild_voting")
	if err != nil {
		failBuild(err, "failed to check if schema exists")
	}

	// If the schema exists, the database is already initialized
	// If the schema does not exist, the database is not initialized
	return exists
}

// schemaExists checks if the schema with the given name exists in the database
func schemaExists(ctx context.Context, db sql.Executor, schema string) (bool, error) {
	query := fmt.Sprintf("SELECT 1 FROM information_schema.schemata WHERE schema_name = '%s'", schema)
	res, err := db.Execute(ctx, query)
	if err != nil {
		return false, err
	}

	if len(res.Rows) == 0 {
		return false, nil
	}

	if len(res.Rows) > 1 {
		return false, fmt.Errorf("more than one schema found with name %s", schema)
	}

	return true, nil
}

func buildBlockStore(d *coreDependencies, closers *closeFuncs) *store.BlockStore {
	blkStrDir := filepath.Join(d.rootDir, "blockstore")
	bs, err := store.NewBlockStore(blkStrDir)
	if err != nil {
		failBuild(err, "failed to open blockstore")
	}
	closers.addCloser(bs.Close, "Closing blockstore") // Close DB after stopping p2p

	return bs
}

func buildAccountStore(ctx context.Context, d *coreDependencies, db *pg.DB) *accounts.Accounts {
	logger := d.logger.New("ACCOUNTS")
	accounts, err := accounts.InitializeAccountStore(ctx, db, logger)
	if err != nil {
		failBuild(err, "failed to initialize account store")
	}

	return accounts
}

func buildVoteStore(ctx context.Context, d *coreDependencies, closers *closeFuncs) (*voting.EventStore, *voting.VoteStore) {
	poolDB, err := d.poolOpener(ctx, d.cfg.DB.DBName, d.cfg.DB.MaxConns)
	if err != nil {
		failBuild(err, "failed to open kwild postgres database for eventstore")
	}
	closers.addCloser(poolDB.Close, "Closing Eventstore DB")

	ev, vs, err := voting.NewResolutionStore(ctx, poolDB)
	if err != nil {
		failBuild(err, "failed to create vote store")
	}

	return ev, vs
}

func buildMetaStore(ctx context.Context, db *pg.DB) {
	err := meta.InitializeMetaStore(ctx, db)
	if err != nil {
		failBuild(err, "failed to initialize meta store")
	}
}

// service returns a common.Service with the given logger name
func (c *coreDependencies) service(loggerName string) *common.Service {
	signer := auth.GetNodeSigner(c.privKey)

	return &common.Service{
		Logger:        c.logger.New(loggerName),
		GenesisConfig: c.genesisCfg,
		LocalConfig:   c.cfg,
		Identity:      signer.Identity(),
	}
}

func buildTxApp(ctx context.Context, d *coreDependencies, db *pg.DB, accounts *accounts.Accounts,
	votestore *voting.VoteStore, engine common.Engine) *txapp.TxApp {
	signer := auth.GetNodeSigner(d.privKey)

	txapp, err := txapp.NewTxApp(ctx, db, engine, signer, nil, d.service("TxAPP"), accounts, votestore)
	if err != nil {
		failBuild(err, "failed to create txapp")
	}

	return txapp
}

func buildBlockProcessor(ctx context.Context, d *coreDependencies, db *pg.DB, txapp *txapp.TxApp, accounts *accounts.Accounts, vs *voting.VoteStore, ss *snapshotter.SnapshotStore, es *voting.EventStore, migrator *migrations.Migrator, bs *store.BlockStore) *blockprocessor.BlockProcessor {
	signer := auth.GetNodeSigner(d.privKey)

	bp, err := blockprocessor.NewBlockProcessor(ctx, db, txapp, accounts, vs, ss, es, migrator, bs, d.genesisCfg, signer, d.logger.New("BP"))
	if err != nil {
		failBuild(err, "failed to create block processor")
	}

	return bp
}

func buildMigrator(d *coreDependencies, db *pg.DB, accounts *accounts.Accounts, vs *voting.VoteStore) *migrations.Migrator {
	migrationsDir := config.MigrationDir(d.rootDir)

	err := os.MkdirAll(migrations.ChangesetsDir(migrationsDir), 0755)
	if err != nil {
		failBuild(err, "failed to create changesets directory")
	}

	snapshotDir := migrations.SnapshotDir(migrationsDir)
	err = os.MkdirAll(snapshotDir, 0755)
	if err != nil {
		failBuild(err, "failed to create migrations snapshots directory")
	}

	ss, err := snapshotter.NewSnapshotStore(&snapshotter.SnapshotConfig{
		SnapshotDir:     snapshotDir,
		MaxSnapshots:    int(d.cfg.Snapshots.MaxSnapshots),
		RecurringHeight: d.cfg.Snapshots.RecurringHeight,
		Enable:          d.cfg.Snapshots.Enable,
		DBConfig:        &d.cfg.DB,
	}, d.logger.New("SNAP"))
	if err != nil {
		failBuild(err, "failed to create migration's snapshot store")
	}

	migrator, err := migrations.SetupMigrator(d.ctx, db, ss, accounts, migrationsDir, d.genesisCfg.Migration, vs, d.logger.New(`MIGRATOR`))
	if err != nil {
		failBuild(err, "failed to create migrator")
	}

	return migrator
}

func buildConsensusEngine(_ context.Context, d *coreDependencies, db *pg.DB,
	mempool *mempool.Mempool, bs *store.BlockStore, bp *blockprocessor.BlockProcessor, valSet map[string]ktypes.Validator) *consensus.ConsensusEngine {
	leaderPubKey, err := crypto.UnmarshalSecp256k1PublicKey(d.genesisCfg.Leader)
	if err != nil {
		failBuild(err, "failed to parse leader public key")
	}

	ceCfg := &consensus.Config{
		PrivateKey:            d.privKey,
		Leader:                leaderPubKey,
		DB:                    db,
		BlockStore:            bs,
		BlockProcessor:        bp,
		Mempool:               mempool,
		ValidatorSet:          valSet,
		Logger:                d.logger.New("CONS"),
		ProposeTimeout:        time.Duration(d.cfg.Consensus.ProposeTimeout),
		BlockProposalInterval: time.Duration(d.cfg.Consensus.BlockProposalInterval),
		BlockAnnInterval:      time.Duration(d.cfg.Consensus.BlockAnnInterval),
		GenesisHeight:         d.genesisCfg.InitialHeight,
	}

	ce := consensus.New(ceCfg)
	if ce == nil {
		failBuild(nil, "failed to create consensus engine")
	}

	return ce
}

func buildNode(d *coreDependencies, mp *mempool.Mempool, bs *store.BlockStore,
	ce *consensus.ConsensusEngine, ss *snapshotter.SnapshotStore, db *pg.DB,
	bp *blockprocessor.BlockProcessor) *node.Node {
	logger := d.logger.New("NODE")
	nc := &node.Config{
		ChainID:     d.genesisCfg.ChainID,
		RootDir:     d.rootDir,
		PrivKey:     d.privKey,
		DB:          db,
		P2P:         &d.cfg.P2P,
		Mempool:     mp,
		BlockStore:  bs,
		Consensus:   ce,
		Statesync:   &d.cfg.StateSync,
		Snapshotter: ss,
		BlockProc:   bp,
		Logger:      logger,
		DBConfig:    &d.cfg.DB,
	}

	node, err := node.NewNode(nc)
	if err != nil {
		failBuild(err, "failed to create node")
	}

	logger.Infof("This node is %s", node.Addrs())
	return node
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

func buildEngine(d *coreDependencies, db *pg.DB) *interpreter.ThreadSafeInterpreter {
	extensions := precompiles.RegisteredPrecompiles()
	for name := range extensions {
		d.logger.Info("registered extension", "name", name)
	}

	tx, err := db.BeginTx(d.ctx)
	if err != nil {
		failBuild(err, "failed to start transaction")
	}
	defer tx.Rollback(d.ctx)

	interp, err := interpreter.NewInterpreter(d.ctx, tx, d.service("engine"))
	if err != nil {
		failBuild(err, "failed to initialize engine")
	}

	err = tx.Commit(d.ctx)
	if err != nil {
		failBuild(err, "failed to commit engine init db txn")
	}

	return interp
}

func buildSnapshotStore(d *coreDependencies) *snapshotter.SnapshotStore {
	snapshotDir := filepath.Join(d.rootDir, "snapshots")
	cfg := &snapshotter.SnapshotConfig{
		SnapshotDir:     snapshotDir,
		MaxSnapshots:    int(d.cfg.Snapshots.MaxSnapshots),
		RecurringHeight: d.cfg.Snapshots.RecurringHeight,
		Enable:          d.cfg.Snapshots.Enable,
		DBConfig:        &d.cfg.DB,
	}

	if err := os.MkdirAll(snapshotDir, 0755); err != nil {
		failBuild(err, "failed to create snapshot directory")
	}

	ss, err := snapshotter.NewSnapshotStore(cfg, d.logger.New("SNAP"))
	if err != nil {
		failBuild(err, "failed to create snapshot store")
	}

	return ss
}

func buildListenerManager(d *coreDependencies, ev *voting.EventStore, bp *blockprocessor.BlockProcessor, node *node.Node) *listeners.ListenerManager {
	return listeners.NewListenerManager(d.service("ListenerManager"), ev, bp, node)
}

func buildJRPCAdminServer(d *coreDependencies) *rpcserver.Server {
	var wantTLS bool
	addr := d.cfg.Admin.ListenAddress
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		if strings.Contains(err.Error(), "missing port in address") {
			host = addr
			port = "8484"
		} else if strings.Contains(err.Error(), "too many colons in address") {
			u, err := url.Parse(addr)
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

	adminPass := d.cfg.Admin.Pass
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
			d.logger.Warn("unresolvable host, assuming not loopback, but will likely fail to listen",
				"host", host, "error", err)
		} else { // e.g. "localhost" usually resolves to a loopback IP address
			loopback = netAddr.IP.IsLoopback()
		}
		if !loopback || wantTLS { // use TLS for encryption, maybe also client auth
			if d.cfg.Admin.NoTLS {
				d.logger.Warn("disabling TLS on non-loopback admin service listen address",
					"addr", addr, "with_password", adminPass != "")
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
	svcLogger := d.logger.New("ADMINRPC")
	jsonRPCAdminServer, err := rpcserver.NewServer(addr, svcLogger, opts...)
	if err != nil {
		failBuild(err, "unable to create json-rpc server")
	}

	return jsonRPCAdminServer
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
		var extraHosts []string
		if hostname != "" {
			extraHosts = []string{hostname}
		}
		if err := genCertPair(certFile, keyFile, extraHosts); err != nil {
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

// tlsConfig returns a tls.Config to be used with the admin RPC service. If
// withClientAuth is true, the config will require client authentication (mutual
// TLS), otherwise it is standard TLS for encryption and server authentication.
func tlsConfig(d *coreDependencies, withClientAuth bool) *tls.Config {
	if d.adminKey == nil {
		return nil
	}
	if !withClientAuth {
		// TLS only for encryption and authentication of server to client.
		return &tls.Config{
			Certificates: []tls.Certificate{*d.adminKey},
		}
	} // else try to load authorized client certs/pubkeys

	var err error
	// client certs
	caCertPool := x509.NewCertPool()
	var clientsCerts []byte
	if clientsFile := filepath.Join(d.rootDir, defaultAdminClients); fileExists(clientsFile) {
		clientsCerts, err = os.ReadFile(clientsFile)
		if err != nil {
			failBuild(err, "failed to load client CAs file")
		}
	} else /*else if d.autogen {
		clientCredsFileBase := filepath.Join(d.rootDir, "auth")
		clientCertFile, clientKeyFile := clientCredsFileBase+".cert", clientCredsFileBase+".key"
		err = transport.GenTLSKeyPair(clientCertFile, clientKeyFile, "local kwild CA", nil)
		if err != nil {
			failBuild(err, "failed to generate admin client credentials")
		}
		d.logger.Info("generated admin service client key pair", log.String("cert", clientCertFile), log.String("key", clientKeyFile))
		if clientsCerts, err = os.ReadFile(clientCertFile); err != nil {
			failBuild(err, "failed to read auto-generate client certificate")
		}
		if err = os.WriteFile(clientsFile, clientsCerts, 0644); err != nil {
			failBuild(err, "failed to write client CAs file")
		}
		d.logger.Info("generated admin service client CAs file", log.String("file", clientsFile))
	} */
	{
		d.logger.Info("No admin client CAs file. Use kwil-admin's node gen-auth-key command to generate")
	}

	if len(clientsCerts) > 0 && !caCertPool.AppendCertsFromPEM(clientsCerts) {
		failBuild(err, "invalid client CAs file")
	}

	// TLS configuration for mTLS (mutual TLS) protocol-level authentication
	return &tls.Config{
		Certificates: []tls.Certificate{*d.adminKey},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    caCertPool,
	}
}

func fileExists(file string) bool {
	fi, err := os.Stat(file)
	if err != nil {
		return false
	}
	return !fi.IsDir()
}
