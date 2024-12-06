package node

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/core/crypto"
	"github.com/kwilteam/kwil-db/core/crypto/auth"
	"github.com/kwilteam/kwil-db/core/types"
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
	"github.com/kwilteam/kwil-db/node/types/sql"
	"github.com/kwilteam/kwil-db/node/voting"

	rpcserver "github.com/kwilteam/kwil-db/node/services/jsonrpc"
	"github.com/kwilteam/kwil-db/node/services/jsonrpc/adminsvc"
	"github.com/kwilteam/kwil-db/node/services/jsonrpc/funcsvc"
	"github.com/kwilteam/kwil-db/node/services/jsonrpc/usersvc"
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
	rpcSvcLogger := d.logger.New("USER")
	jsonRPCTxSvc := usersvc.NewService(db, e, node, txApp, vs, rpcSvcLogger,
		usersvc.WithReadTxTimeout(time.Duration(d.cfg.DB.ReadTxTimeout)),
		usersvc.WithPrivateMode(d.cfg.RPC.Private),
		usersvc.WithChallengeExpiry(d.cfg.RPC.ChallengeExpiry),
		usersvc.WithChallengeRateLimit(d.cfg.RPC.ChallengeRateLimit),
	// usersvc.WithBlockAgeHealth(6*totalConsensusTimeouts.Dur()),
	)

	rpcServerLogger := d.logger.New("RPC")
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

	var jsonRPCAdminServer *rpcserver.Server
	if d.cfg.Admin.Enable {
		// admin service and server
		adminServerLogger := d.logger.New("ADMIN")
		// The admin service uses a client-style signer rather than just a private
		// key because it is used to sign transactions and provide an Identity for
		// account information (nonce and balance).
		appIface := &mysteryThing{txApp, ce}
		txSigner := &auth.EthPersonalSigner{Key: *d.privKey.(*crypto.Secp256k1PrivateKey)}
		jsonAdminSvc := adminsvc.NewService(db, node, appIface, nil, txSigner, d.cfg,
			d.genesisCfg.ChainID, adminServerLogger)
		jsonRPCAdminServer = buildJRPCAdminServer(d)
		jsonRPCAdminServer.RegisterSvc(jsonAdminSvc)
		jsonRPCAdminServer.RegisterSvc(jsonRPCTxSvc)
		jsonRPCAdminServer.RegisterSvc(&funcsvc.Service{})
	}

	s := &server{
		cfg:                d.cfg,
		closers:            closers,
		node:               node,
		ce:                 ce,
		jsonRPCServer:      jsonRPCServer,
		jsonRPCAdminServer: jsonRPCAdminServer,
		dbCtx:              db,
		log:                d.logger,
	}

	return s
}

var _ adminsvc.App = (*mysteryThing)(nil)

type mysteryThing struct {
	txApp *txapp.TxApp
	ce    *consensus.ConsensusEngine
}

func (mt *mysteryThing) AccountInfo(ctx context.Context, db sql.DB, identifier []byte, pending bool) (balance *big.Int, nonce int64, err error) {
	return mt.txApp.AccountInfo(ctx, db, identifier, pending)
}

func (mt *mysteryThing) Price(ctx context.Context, db sql.DB, tx *types.Transaction) (*big.Int, error) {
	return mt.ce.Price(ctx, db, tx)
}

func buildDB(ctx context.Context, d *coreDependencies, closers *closeFuncs) *pg.DB {
	pg.UseLogger(d.logger.New("PG"))

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

func buildConsensusEngine(_ context.Context, d *coreDependencies, db *pg.DB, accounts *accounts.Accounts, vs *voting.VoteStore,
	mempool *mempool.Mempool, bs *store.BlockStore, txapp *txapp.TxApp, valSet map[string]ktypes.Validator,
	ss *snapshotter.SnapshotStore) *consensus.ConsensusEngine {
	leaderPubKey, err := crypto.UnmarshalSecp256k1PublicKey(d.genesisCfg.Leader)
	if err != nil {
		failBuild(err, "failed to parse leader public key")
	}

	genHash := d.genesisCfg.ComputeGenesisHash()

	ceCfg := &consensus.Config{
		PrivateKey: d.privKey,
		Leader:     leaderPubKey,
		GenesisParams: &consensus.GenesisParams{
			ChainID: d.genesisCfg.ChainID,
			Params: &consensus.NetworkParams{
				MaxBlockSize:     d.genesisCfg.MaxBlockSize,
				JoinExpiry:       d.genesisCfg.JoinExpiry,
				VoteExpiry:       d.genesisCfg.VoteExpiry,
				DisabledGasCosts: d.genesisCfg.DisabledGasCosts,
				MaxVotesPerTx:    d.genesisCfg.MaxVotesPerTx,
			},
		},
		GenesisHash:    genHash,
		DB:             db,
		Accounts:       accounts,
		BlockStore:     bs,
		Mempool:        mempool,
		ValidatorStore: vs,
		TxApp:          txapp,
		ValidatorSet:   valSet,
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
