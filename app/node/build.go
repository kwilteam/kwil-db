package node

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/core/crypto"
	"github.com/kwilteam/kwil-db/core/crypto/auth"
	ktypes "github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/extensions/precompiles"
	"github.com/kwilteam/kwil-db/node"
	"github.com/kwilteam/kwil-db/node/accounts"
	"github.com/kwilteam/kwil-db/node/consensus"
	"github.com/kwilteam/kwil-db/node/engine/execution"
	"github.com/kwilteam/kwil-db/node/mempool"
	"github.com/kwilteam/kwil-db/node/meta"
	"github.com/kwilteam/kwil-db/node/pg"
	"github.com/kwilteam/kwil-db/node/snapshotter"
	"github.com/kwilteam/kwil-db/node/store"
	"github.com/kwilteam/kwil-db/node/txapp"
	"github.com/kwilteam/kwil-db/node/voting"

	rpcserver "github.com/kwilteam/kwil-db/node/services/jsonrpc"
	"github.com/kwilteam/kwil-db/node/services/jsonrpc/funcsvc"
	usersvc "github.com/kwilteam/kwil-db/node/services/jsonrpc/usersvc"
)

func buildServer(ctx context.Context, d *coreDependencies) *server {
	closers := &closeFuncs{
		closers: []func() error{}, // logger.Close is not in here; do it in a defer in Start
		logger:  d.logger,
	}

	valSet := make(map[string]ktypes.Validator)
	for _, v := range d.genesisCfg.Validators {
		valSet[hex.EncodeToString(v.PubKey)] = v
	}

	// Initialize DB
	db := buildDB(ctx, d, closers)

	// metastore
	buildMetaStore(ctx, db)

	e := buildEngine(d, db)

	// BlockStore
	bs := buildBlockStore(d, closers)

	// Mempool
	mp := mempool.New()

	// accounts
	accounts := buildAccountStore(ctx, db)

	// eventstore, votestore
	_, vs := buildVoteStore(ctx, d, closers) // ev, vs

	// TxAPP
	txApp := buildTxApp(ctx, d, db, accounts, vs, e)

	// Snapshot Store
	ss := buildSnapshotStore(d)

	// Consensus
	ce := buildConsensusEngine(ctx, d, db, accounts, vs, mp, bs, txApp, valSet, ss)

	// Node
	node := buildNode(d, mp, bs, ce, ss, db)

	// RPC Services
	rpcSvcLogger := d.logger.New("user-json-svc")
	jsonRPCTxSvc := usersvc.NewService(db, e, node, txApp, vs, rpcSvcLogger,
		// usersvc.WithReadTxTimeout(time.Duration(d.cfg.AppConfig.ReadTxTimeout)),
		usersvc.WithPrivateMode(d.cfg.RPC.Private),
	// usersvc.WithChallengeExpiry(time.Duration(d.cfg.AppConfig.ChallengeExpiry)),
	// usersvc.WithChallengeRateLimit(d.cfg.AppConfig.ChallengeRateLimit),
	// usersvc.WithBlockAgeHealth(6*totalConsensusTimeouts.Dur()),
	)

	rpcServerLogger := d.logger.New("user-jsonrprc-server")
	jsonRPCServer, err := rpcserver.NewServer(d.cfg.RPC.ListenAddress,
		rpcServerLogger, rpcserver.WithTimeout(d.cfg.RPC.Timeout),
		rpcserver.WithReqSizeLimit(d.cfg.RPC.MaxReqSize),
		rpcserver.WithCORS(), rpcserver.WithServerInfo(&usersvc.SpecInfo),
		rpcserver.WithMetricsNamespace("kwil_json_rpc_user_server"))
	if err != nil {
		failBuild(err, "unable to create json-rpc server")
	}
	jsonRPCServer.RegisterSvc(jsonRPCTxSvc)
	jsonRPCServer.RegisterSvc(&funcsvc.Service{})

	// admin service and server
	// signer := buildSigner(d)
	// jsonAdminSvc := adminsvc.NewService(db, wrappedCmtClient, txApp, abciApp, p2p, nil, d.cfg,
	// 	d.genesisCfg.ChainID, *d.log.Named("admin-json-svc"))
	// jsonRPCAdminServer := buildJRPCAdminServer(d)
	// jsonRPCAdminServer.RegisterSvc(jsonAdminSvc)
	// jsonRPCAdminServer.RegisterSvc(jsonRPCTxSvc)
	// jsonRPCAdminServer.RegisterSvc(&funcsvc.Service{})

	s := &server{
		cfg:     d.cfg,
		closers: closers,
		node:    node,
		ce:      ce,
		dbCtx:   db,
		log:     d.logger,
	}

	return s
}

func buildDB(ctx context.Context, d *coreDependencies, closers *closeFuncs) *pg.DB {
	// TODO: restore from snapshots

	db, err := d.dbOpener(ctx, d.cfg.DB.DBName, d.cfg.DB.MaxConns)
	if err != nil {
		failBuild(err, "failed to open kwild postgres database")
	}
	closers.addCloser(db.Close, "closing main DB")

	// TODO: bring back the prev functionality
	return db
}

func buildBlockStore(d *coreDependencies, closers *closeFuncs) *store.BlockStore {
	blkStrDir := filepath.Join(d.rootDir, "blockstore")
	bs, err := store.NewBlockStore(blkStrDir)
	if err != nil {
		failBuild(err, "failed to open blockstore")
	}
	closers.addCloser(bs.Close, "closing blockstore") // Close DB after stopping p2p

	return bs
}

func buildAccountStore(ctx context.Context, db *pg.DB) *accounts.Accounts {
	accounts, err := accounts.InitializeAccountStore(ctx, db)
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

func buildTxApp(ctx context.Context, d *coreDependencies, db *pg.DB, accounts *accounts.Accounts,
	votestore *voting.VoteStore, engine *execution.GlobalContext) *txapp.TxApp {
	signer := auth.GetSigner(d.privKey)
	service := &common.Service{
		Logger:   d.logger.New("TXAPP"),
		Identity: signer.Identity(),
		// TODO: pass extension configs
		// ExtensionConfigs: make(map[string]map[string]string),
	}

	txapp, err := txapp.NewTxApp(ctx, db, engine, signer, nil, service, accounts, votestore)
	if err != nil {
		failBuild(err, "failed to create txapp")
	}

	return txapp
}

func buildConsensusEngine(_ context.Context, d *coreDependencies, db *pg.DB, accounts *accounts.Accounts, vs *voting.VoteStore, mempool *mempool.Mempool, bs *store.BlockStore, txapp *txapp.TxApp, valSet map[string]ktypes.Validator, ss *snapshotter.SnapshotStore) *consensus.ConsensusEngine {
	leaderPubKey, err := crypto.UnmarshalSecp256k1PublicKey(d.genesisCfg.Leader)
	if err != nil {
		failBuild(err, "failed to parse leader public key")
	}

	ceCfg := &consensus.Config{
		PrivateKey:     d.privKey,
		Leader:         leaderPubKey,
		DB:             db,
		Accounts:       accounts,
		BlockStore:     bs,
		Mempool:        mempool,
		ValidatorStore: vs,
		TxApp:          txapp,
		ValidatorSet:   valSet, // TODO: Where to set this validator set? in the constructor or after the ce is caughtup?
		Logger:         d.logger.New("CONS"),
		ProposeTimeout: d.cfg.Consensus.ProposeTimeout,
		Snapshots:      ss,
	}

	ce := consensus.New(ceCfg)
	if ce == nil {
		failBuild(nil, "failed to create consensus engine")
	}

	return ce
}

func buildNode(d *coreDependencies, mp *mempool.Mempool, bs *store.BlockStore, ce *consensus.ConsensusEngine, ss *snapshotter.SnapshotStore, db *pg.DB) *node.Node {
	logger := d.logger.New("NODE")
	nc := &node.Config{
		RootDir:     d.rootDir,
		PrivKey:     d.privKey,
		DB:          db,
		P2P:         &d.cfg.P2P,
		Mempool:     mp,
		BlockStore:  bs,
		Consensus:   ce,
		Statesync:   &d.cfg.StateSync,
		Snapshotter: ss,
		Snapshots:   &d.cfg.Snapshots,
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

func buildEngine(d *coreDependencies, db *pg.DB) *execution.GlobalContext {
	extensions := precompiles.RegisteredPrecompiles()
	for name := range extensions {
		d.logger.Info("registered extension", "name", name)
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

	eng, err := execution.NewGlobalContext(d.ctx, tx,
		extensions, d.newService("engine"))
	if err != nil {
		failBuild(err, "failed to build engine")
	}

	err = tx.Commit(d.ctx)
	if err != nil {
		failBuild(err, "failed to commit engine init db txn")
	}

	return eng
}

func buildSnapshotStore(d *coreDependencies) *snapshotter.SnapshotStore {
	snapshotDir := filepath.Join(d.rootDir, "snapshots")
	cfg := &snapshotter.SnapshotConfig{
		SnapshotDir:     snapshotDir,
		MaxSnapshots:    int(d.cfg.Snapshots.MaxSnapshots),
		RecurringHeight: d.cfg.Snapshots.RecurringHeight,
		Enable:          d.cfg.Snapshots.Enable,
	}

	dbCfg := &snapshotter.DBConfig{
		DBHost: d.cfg.DB.Host,
		DBPort: d.cfg.DB.Port,
		DBUser: d.cfg.DB.User,
		DBPass: d.cfg.DB.Pass,
		DBName: d.cfg.DB.DBName,
	}

	if err := os.MkdirAll(snapshotDir, 0755); err != nil {
		failBuild(err, "failed to create snapshot directory")
	}

	ss, err := snapshotter.NewSnapshotStore(cfg, dbCfg, d.logger.New("SNAP"))
	if err != nil {
		failBuild(err, "failed to create snapshot store")
	}

	return ss
}
