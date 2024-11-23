package app

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"kwil/crypto"
	"kwil/crypto/auth"
	"kwil/node"
	"kwil/node/consensus"
	"kwil/node/mempool"
	"kwil/node/meta"
	"kwil/node/peers"
	"kwil/node/pg"
	"kwil/node/store"
	"kwil/node/txapp"
	ktypes "kwil/types"
	"path/filepath"

	"kwil/node/accounts"
	"kwil/node/voting"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
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

	// BlockStore
	bs := buildBlockStore(d, closers)

	// Mempool
	mp := mempool.New()

	host := newHost(d, closers)

	// PeerManager
	pm := buildPeerManager(d, host)

	// accounts
	accounts := buildAccountStore(ctx, db)

	// eventstore, votestore
	_, vs := buildVoteStore(ctx, d, closers) // ev, vs

	// TxAPP
	txapp := buildTxApp(ctx, d, db, accounts, vs)

	// Consensus
	ce := buildConsensusEngine(ctx, d, db, accounts, vs, mp, bs, txapp, valSet)

	// Node
	node := buildNode(d, host, pm, mp, bs, ce, valSet)

	// RPC Services

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

func newHost(d *coreDependencies, closers *closeFuncs) host.Host {
	host, err := node.NewHost(d.cfg.P2P.IP, d.cfg.P2P.Port, d.privKey)
	if err != nil {
		failBuild(err, "failed to create libp2p host")
	}
	closers.addCloser(host.Close, "closing libp2p host")

	return host
}

func buildPeerManager(d *coreDependencies, host host.Host) *peers.PeerMan {
	addrBookPath := filepath.Join(d.rootDir, "addrbook.json")

	pm, err := peers.NewPeerMan(d.cfg.P2P.Pex, addrBookPath,
		d.logger.New("PEERS"),
		host, // tooo much, become minimal interface
		func(ctx context.Context, peerID peer.ID) ([]peer.AddrInfo, error) {
			return node.RequestPeers(ctx, host.ID(), host, d.logger)
		}, node.RequiredStreamProtocols)
	if err != nil {
		failBuild(err, "failed to create peer manager")
	}

	return pm
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

func buildTxApp(ctx context.Context, d *coreDependencies, db *pg.DB, accounts *accounts.Accounts, votestore *voting.VoteStore) *txapp.TxApp {
	signer := auth.GetSigner(d.privKey)
	service := &ktypes.Service{
		Logger:   d.logger.New("TXAPP"),
		Identity: signer.Identity(),
		// TODO: pass extension configs
		// ExtensionConfigs: make(map[string]map[string]string),
	}

	txapp, err := txapp.NewTxApp(ctx, db, nil, signer, nil, service, accounts, votestore)
	if err != nil {
		failBuild(err, "failed to create txapp")
	}

	return txapp
}

func buildConsensusEngine(_ context.Context, d *coreDependencies, db *pg.DB, accounts *accounts.Accounts, vs *voting.VoteStore, mempool *mempool.Mempool, bs *store.BlockStore, txapp *txapp.TxApp, valSet map[string]ktypes.Validator) *consensus.ConsensusEngine {
	leader := d.genesisCfg.Validators[0].PubKey
	leaderPubKey, err := crypto.UnmarshalSecp256k1PublicKey(leader)
	if err != nil {
		failBuild(err, "failed to parse leader public key")
	}

	ceCfg := &consensus.Config{
		PrivateKey: d.privKey,
		Leader:     leaderPubKey,
		// Leader:    d.cfg.Consensus.Leader,
		DB:             db,
		Accounts:       accounts,
		BlockStore:     bs,
		Mempool:        mempool,
		ValidatorStore: vs,
		TxApp:          txapp,
		ValidatorSet:   valSet,
		Logger:         d.logger.New("CONS"),
		ProposeTimeout: d.cfg.Consensus.ProposeTimeout,
	}

	ce := consensus.New(ceCfg)
	if ce == nil {
		failBuild(nil, "failed to create consensus engine")
	}

	return ce
}

func buildNode(d *coreDependencies, h host.Host, pm *peers.PeerMan, mp *mempool.Mempool, bs *store.BlockStore, ce *consensus.ConsensusEngine, valSet map[string]ktypes.Validator) *node.Node {
	logger := d.logger.New("NODE")
	nc := &node.Config{
		RootDir:    d.rootDir,
		PrivKey:    d.privKey,
		Cfg:        d.cfg,
		Genesis:    d.genesisCfg,
		Host:       h,
		PeerMgr:    pm,
		Mempool:    mp,
		BlockStore: bs,
		Consensus:  ce,
		ValSet:     valSet,
		Logger:     logger,
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
