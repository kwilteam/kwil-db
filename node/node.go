package node

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	mrand2 "math/rand/v2"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/kwilteam/kwil-db/core/crypto"
	"github.com/kwilteam/kwil-db/core/log"
	ktypes "github.com/kwilteam/kwil-db/core/types"
	adminTypes "github.com/kwilteam/kwil-db/core/types/admin"
	chainTypes "github.com/kwilteam/kwil-db/core/types/chain"
	"github.com/kwilteam/kwil-db/node/peers"
	"github.com/kwilteam/kwil-db/node/types"

	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/connmgr"
	p2pcrypto "github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/core/protocol"
	noise "github.com/libp2p/go-libp2p/p2p/security/noise"
	"github.com/libp2p/go-libp2p/p2p/transport/tcp"
	"github.com/multiformats/go-multiaddr"
)

const (
	blockTxCount    = 50 // for "mining"
	txReAnnInterval = 30 * time.Second
)

type peerManager interface {
	network.Notifiee
	Start(context.Context) error
	ConnectedPeers() []peers.PeerInfo
	KnownPeers() ([]peers.PeerInfo, []peers.PeerInfo, []peers.PeerInfo)
	Connect(ctx context.Context, info peers.AddrInfo) error

	// Whitelist methods
	Allow(p peer.ID)
	AllowPersistent(p peer.ID)
	Disallow(p peer.ID)
	// Allowed() []peer.ID
	AllowedPersistent() []peer.ID
	// IsAllowed(p peer.ID) bool
}

type WhitelistMgr struct {
	pm     peerManager
	logger log.Logger
}

// Whitelister is a shim between the a Kwil consumer like RPC service and the
// p2p layer (PeerMan) which manages the persistent and effective white list in
// terms of libp2p types.
func (n *Node) Whitelister() *WhitelistMgr {
	return &WhitelistMgr{pm: n.pm, logger: n.log.New("WHITELIST")}
}

func (wl *WhitelistMgr) AddPeer(nodeID string) error {
	wl.logger.Info(nodeID)
	peerID, err := nodeIDToPeerID(nodeID)
	if err != nil {
		return err
	}
	wl.logger.Info(peerID.String())
	wl.pm.AllowPersistent(peerID)
	return nil
}

func (wl *WhitelistMgr) RemovePeer(nodeID string) error {
	peerID, err := nodeIDToPeerID(nodeID)
	if err != nil {
		return err
	}
	wl.pm.Disallow(peerID)
	return nil
}

func (wl *WhitelistMgr) List() []string {
	var list []string
	for _, peerID := range wl.pm.AllowedPersistent() {
		nodeID, err := peers.NodeIDFromPeerID(peerID.String())
		if err != nil { // this shouldn't happen
			wl.logger.Errorf("invalid peer ID in whitelist: %v", err)
			continue // return nil, err
		}
		list = append(list, nodeID)
	}
	return list
}

type Node struct {
	// cfg
	pex    bool
	pubkey crypto.PublicKey
	dir    string
	// pf *prefetch
	chainID string

	// interfaces
	bki         types.BlockStore
	mp          types.MemPool
	ce          ConsensusEngine
	pm          peerManager // *peers.PeerMan
	ss          SnapshotStore
	host        host.Host
	statesyncer *StateSyncService
	bp          BlockProcessor

	// broadcast channels
	ackChan  chan AckRes                  // from consensus engine, to gossip to leader
	resetMsg chan ConsensusReset          // gossiped in from peers, to consensus engine
	discReq  chan types.DiscoveryRequest  // from consensus engine, to gossip to leader for calculating best height of the validators during blocksync.
	discResp chan types.DiscoveryResponse // from gossip, to consensus engine for calculating best height of the validators during blocksync.

	wg        sync.WaitGroup
	log       log.Logger
	dhtCloser func() error
}

// NewNode creates a new node. The config struct is for required configuration,
// and the functional options for optional settings, like dependency overrides.
func NewNode(cfg *Config, opts ...Option) (*Node, error) {
	options := &options{}
	for _, opt := range opts {
		opt(options)
	}

	logger := cfg.Logger
	if logger == nil {
		logger = log.DiscardLogger
	}

	pubkey := cfg.PrivKey.Public()

	// This connection gater is logically be part of PeerMan, but the libp2p
	// Host constructor needs it, and PeerMan needs Host for its peerstore
	// and connect method. For now we create it here and give it to both.
	var cg *peers.WhitelistGater
	if cfg.P2P.PrivateMode {
		logger.Infof("Private P2P mode enabled")
		var peerWhitelist []peer.ID
		for _, nodeID := range cfg.P2P.Whitelist {
			peerID, err := nodeIDToPeerID(nodeID)
			if err != nil {
				return nil, fmt.Errorf("invalid whitelist node ID: %w", err)
			}
			peerWhitelist = append(peerWhitelist, peerID)
			logger.Infof("Adding peer to whitelist: %v", nodeID)
		}
		cg = peers.NewWhitelistGater(peerWhitelist, peers.WithLogger(logger.New("PEERFILT")))
		// PeerMan adds more from address book.
	}

	var err error
	host := options.host
	if host == nil {
		host, err = newHost(cfg.P2P.IP, cfg.P2P.Port, cfg.PrivKey, cg)
		if err != nil {
			return nil, fmt.Errorf("cannot create host: %w", err)
		}
	}

	addrBookPath := filepath.Join(cfg.RootDir, "addrbook.json")

	pm, err := peers.NewPeerMan(cfg.P2P.Pex, addrBookPath,
		logger.New("PEERS"), cg, host,
		func(ctx context.Context, peerID peer.ID) ([]peer.AddrInfo, error) {
			return requestPeers(ctx, host.ID(), host, logger)
		}, RequiredStreamProtocols)
	if err != nil {
		return nil, fmt.Errorf("failed to create peer manager: %w", err)
	}

	mode := dht.ModeServer
	ctx := context.Background()
	dht, err := makeDHT(ctx, host, nil, mode)
	if err != nil {
		return nil, fmt.Errorf("failed to create DHT: %w", err)
	}
	discoverer := makeDiscovery(dht)

	// statesyncer
	rcvdSnapsDir := filepath.Join(cfg.RootDir, "rcvd_snaps")
	ssCfg := &statesyncConfig{
		RcvdSnapsDir:  rcvdSnapsDir,
		StateSyncCfg:  cfg.Statesync,
		DBConfig:      cfg.DBConfig,
		Logger:        logger.New("STATESYNC"),
		DB:            cfg.DB,
		Host:          host,
		Discoverer:    discoverer,
		SnapshotStore: cfg.Snapshotter,
		BlockStore:    cfg.BlockStore,
	}
	ss, err := NewStateSyncService(ctx, ssCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create state sync service: %w", err)
	}

	node := &Node{
		log:         logger,
		pubkey:      pubkey,
		pex:         cfg.P2P.Pex,
		host:        host,
		pm:          pm,
		statesyncer: ss,
		dhtCloser:   dht.Close,
		mp:          cfg.Mempool,
		bki:         cfg.BlockStore,
		ce:          cfg.Consensus,
		dir:         cfg.RootDir,
		chainID:     cfg.ChainID,
		ss:          cfg.Snapshotter,
		bp:          cfg.BlockProc,

		ackChan:  make(chan AckRes, 1),
		resetMsg: make(chan ConsensusReset, 1),
		discReq:  make(chan types.DiscoveryRequest, 1),
		discResp: make(chan types.DiscoveryResponse, 1),
	}

	host.SetStreamHandler(ProtocolIDTxAnn, node.txAnnStreamHandler)
	host.SetStreamHandler(ProtocolIDBlkAnn, node.blkAnnStreamHandler)
	host.SetStreamHandler(ProtocolIDBlock, node.blkGetStreamHandler)
	host.SetStreamHandler(ProtocolIDBlockHeight, node.blkGetHeightStreamHandler)
	host.SetStreamHandler(ProtocolIDTx, node.txGetStreamHandler)

	host.SetStreamHandler(ProtocolIDBlockPropose, node.blkPropStreamHandler)

	if cfg.P2P.Pex {
		host.SetStreamHandler(ProtocolIDDiscover, node.peerDiscoveryStreamHandler)
	} else {
		host.SetStreamHandler(ProtocolIDDiscover, func(s network.Stream) {
			s.Close()
		})
	}

	return node, nil
}

func FormatPeerString(rawPubKey []byte, keyType crypto.KeyType, ip string, port int) string {
	return fmt.Sprintf("%s#%d@%s", hex.EncodeToString(rawPubKey), keyType,
		net.JoinHostPort(ip, strconv.Itoa(port)))
}

func (n *Node) Addrs() []string {
	return addrs(n.host)
}

func addrs(h host.Host) []string {
	hosts, ports, _ := hostPort(h)
	if len(hosts) == 0 {
		return nil
	}

	addrs := make([]string, len(hosts))
	for i, h := range hosts {
		addrs[i] = fmt.Sprintf("%s:%d", h, ports[i])
	}
	return addrs
}

func maddrs(h host.Host) []string {
	hosts, ports, protocols := hostPort(h)
	id := h.ID()
	if len(hosts) == 0 {
		return nil
	}
	addrs := make([]string, len(hosts))
	for i, host := range hosts {
		addrs[i] = fmt.Sprintf("/%s/%s/tcp/%d/p2p/%s", protocols[i], host, ports[i], id)
	}
	return addrs
}

func (n *Node) MultiAddrs() []string {
	return maddrs(n.host)
}

func (n *Node) Dir() string {
	return n.dir
}

func (n *Node) ID() string {
	return n.host.ID().String()
}

// Start begins tx and block gossip, connects to any bootstrap peers, and begins
// peer discovery.
func (n *Node) Start(ctx context.Context, bootpeers ...string) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	n.host.Network().Notify(n.pm)
	defer n.host.Network().StopNotify(n.pm)

	ps, err := pubsub.NewGossipSub(ctx, n.host)
	if err != nil {
		return err
	}

	bootpeersMA, err := peers.ConvertPeersToMultiAddr(bootpeers)
	if err != nil {
		return err
	}

	// connect to bootstrap peers, if any.
	//
	// NOTE: it may be preferable to simply add to Host's peer store here and
	// let PeerMan manage connections.
	for i, peer := range bootpeersMA {
		peerInfo, err := makePeerAddrInfo(peer)
		if err != nil {
			n.log.Warnf("invalid bootnode address %v from setting %v", peer, bootpeers[i])
			continue
		}

		n.pm.Allow(peerInfo.ID)

		err = n.pm.Connect(ctx, peers.AddrInfo(*peerInfo))
		if err != nil {
			n.log.Errorf("failed to connect to %v: %v", peer, err)
			// Add it to the peer store anyway since this was specified as a
			// bootnode, which is supposed to be persistent, so we should try to
			// connect again later.
			n.host.Peerstore().AddAddrs(peerInfo.ID, peerInfo.Addrs, peerstore.PermanentAddrTTL)
			continue
		}
		if err = n.checkPeerProtos(ctx, peerInfo.ID); err != nil {
			n.log.Warnf("WARNING: peer does not support required protocols %v: %v", peer, err)
			if err = n.host.Network().ClosePeer(peerInfo.ID); err != nil {
				n.log.Errorf("failed to disconnect from %v: %v", peer, err)
			}
			// n.host.Peerstore().RemovePeer()
			continue
		}
		n.log.Infof("Connected to bootstrap peer %v", peer)
		// n.host.ConnManager().TagPeer(peerID, "validatorish", 1)
	} // else would use persistent peer store (address book)

	for _, val := range n.bp.GetValidators() {
		n.log.Infof("Adding validator %v to peer whitelist", val.PubKey)
		peerID, err := peerIDForValidator(val.PubKey)
		if err != nil {
			n.log.Errorf("cannot get peerID for validator (%v): %v", val.PubKey, err)
			continue
		}
		n.pm.Allow(peerID)
	}

	valSub := n.bp.SubscribeValidators()
	n.wg.Add(1)
	go func() {
		defer n.wg.Done()
		for {
			select {
			case valUpdates, open := <-valSub:
				if !open {
					return
				}
				// update peer filter
				for _, up := range valUpdates {
					peerID, err := peerIDForValidator(up.PubKey)
					if err != nil {
						n.log.Errorf("cannot get peerID for validator (%v): %v", up.PubKey, err)
						continue
					}
					if up.Power > 0 {
						n.pm.Allow(peerID)
					} else {
						n.pm.Disallow(peerID)
					}
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	// Advertise the snapshotcatalog service if snapshots are enabled
	// umm, but gotcha, if a node has previous snapshots but snapshots are disabled, these snapshots will be unusable.
	if n.ss.Enabled() {
		advertise(ctx, snapshotCatalogNS, n.statesyncer.discoverer)
	}

	if err := n.statesyncer.Bootstrap(ctx); err != nil {
		return fmt.Errorf("failed to bootstrap DHT service with the trusted snapshot providers: %w", err)
	}

	// Attempt statesync if enabled
	if err := n.doStatesync(ctx); err != nil {
		cancel()
		return err
	}

	if err := n.startAckGossip(ctx, ps); err != nil {
		cancel()
		return err
	}

	if err := n.startConsensusResetGossip(ctx, ps); err != nil {
		cancel()
		return err
	}

	if err := n.startDiscoveryRequestGossip(ctx, ps); err != nil {
		cancel()
		return err
	}

	if err := n.startDiscoveryResponseGossip(ctx, ps); err != nil {
		cancel()
		return err
	}

	// custom stream-based gossip uses txAnnStreamHandler and announceTx.
	// This dummy method will make create+announce new pretend transactions.
	// It also periodically rebroadcasts txns.
	n.startTxAnns(ctx, txReAnnInterval)

	// mine is our block anns goroutine, which must be only for leader
	n.wg.Add(1)
	var nodeErr error

	go func() {
		defer n.wg.Done()
		defer cancel()
		// TODO: umm, should node bringup the consensus engine? or server?
		nodeErr = n.ce.Start(ctx, n.announceBlkProp, n.announceBlk, n.sendACK, n.getBlkHeight, n.sendReset, n.sendDiscoveryRequest, n.BroadcastTx)
		if err != nil {
			n.log.Errorf("Consensus engine failed: %v", nodeErr)
			return // cancel context
		}
	}()

	n.wg.Add(1)
	go func() {
		defer n.wg.Done()
		n.pm.Start(ctx)
	}()

	n.log.Info("Node started.")

	<-ctx.Done()
	n.log.Info("Stopping Node protocol handlers...")
	n.wg.Wait()

	n.log.Info("Stopping P2P services...")

	if err = n.dhtCloser(); err != nil {
		n.log.Warn("Failed to cleanly stop the DHT service: %v", err)
	}
	if err = n.host.Close(); err != nil {
		n.log.Warn("Failed to cleanly stop P2P host: %v", err)
	}

	n.log.Info("Node stopped.")
	return nodeErr
}

func peerIDForValidator(pubkeyBts []byte) (peer.ID, error) {
	// We only support secp256k1 keys for validators presently.
	pk, err := crypto.UnmarshalSecp256k1PublicKey(pubkeyBts)
	if err != nil {
		return "", fmt.Errorf("Invalid validator pubkey: %vw", err)
	}
	peerID, err := peer.IDFromPublicKey((*p2pcrypto.Secp256k1PublicKey)(pk))
	if err != nil {
		return "", fmt.Errorf("Invalid validator pubkey: %w", err)
	}
	return peerID, nil
}

func nodeIDToPeerID(nodeID string) (peer.ID, error) {
	pubKey, err := peers.NodeIDToPubKey(nodeID)
	if err != nil {
		return "", err
	}
	return pubkeyToPeerID(pubKey)
}

func pubkeyToPeerID(pubkey crypto.PublicKey) (peer.ID, error) {
	var p2pPub p2pcrypto.PubKey
	var err error
	switch pt := pubkey.(type) {
	case *crypto.Secp256k1PublicKey:
		p2pPub = (*p2pcrypto.Secp256k1PublicKey)(pt)
	case *crypto.Ed25519PublicKey:
		// no shortcuts for ed25519
		rawPub := pubkey.Bytes()
		p2pPub, err = p2pcrypto.UnmarshalEd25519PublicKey(rawPub)
	default:
		return "", fmt.Errorf("unsupported pubkey type: %T", pubkey)
	}
	if err != nil {
		return "", err
	}
	p2pAddr, err := peer.IDFromPublicKey(p2pPub)
	if err != nil {
		return "", err
	}
	return p2pAddr, nil
}

// doStatesync attempts to perform statesync if the db is uninitialized.
// It also initializes the blockstore with the initial block data at the
// height of the discovered snapshot.
func (n *Node) doStatesync(ctx context.Context) error {
	// If statesync is enabled and the db is uninitialized, discover snapshots
	if !n.statesyncer.cfg.Enable {
		return nil
	}

	// Check if the Block store and DB are initialized
	h, _, _ := n.bki.Best()
	if h != 0 {
		return nil
	}

	// check if the db is uninitialized
	height, err := n.statesyncer.DiscoverSnapshots(ctx)
	if err != nil {
		return fmt.Errorf("failed to attempt statesync: %w", err)
	}

	if height <= 0 { // no snapshots found, or statesync failed
		return nil
	}

	// request and commit the block to the blockstore
	_, appHash, rawBlk, err := n.getBlkHeight(ctx, height)
	if err != nil {
		return fmt.Errorf("failed to get statesync block %d: %w", height, err)
	}
	blk, err := ktypes.DecodeBlock(rawBlk)
	if err != nil {
		return fmt.Errorf("failed to decode statesync block %d: %w", height, err)
	}
	// store block
	if err := n.bki.Store(blk, appHash); err != nil {
		return fmt.Errorf("failed to store statesync block to the blockstore %d: %w", height, err)
	}

	return nil
}

func (n *Node) Peers(context.Context) ([]*adminTypes.PeerInfo, error) {
	peers := n.pm.ConnectedPeers()
	var peersInfo []*adminTypes.PeerInfo
	for _, peer := range peers {
		conns := n.host.Network().ConnsToPeer(peer.ID)
		if len(conns) == 0 { // should be at least one
			continue
		}

		var addr string
		if len(peer.Addrs) > 0 {
			host, port, _ := maHostPort(peer.Addrs[0])
			addr = net.JoinHostPort(host, port)
		}

		peersInfo = append(peersInfo, &adminTypes.PeerInfo{
			NodeInfo:   &adminTypes.NodeInfo{},
			Inbound:    conns[0].Stat().Direction == network.DirInbound,
			RemoteAddr: addr,
		})
	}
	return peersInfo, nil
}

func (n *Node) Status(ctx context.Context) (*adminTypes.Status, error) {
	height, blkHash, appHash := n.bki.Best()
	var addr string
	if addrs := n.Addrs(); len(addrs) > 0 {
		addr = addrs[0]
	}
	pkBytes := n.pubkey.Bytes()
	return &adminTypes.Status{
		Node: &adminTypes.NodeInfo{
			ChainID:    n.chainID,
			NodeID:     hex.EncodeToString(pkBytes),
			ListenAddr: addr,
		},
		Sync: &adminTypes.SyncInfo{
			AppHash:         appHash,
			BestBlockHash:   blkHash,
			BestBlockHeight: height,
			// BestBlockTime: ,
			Syncing: n.ce.InCatchup(),
		},
		Validator: &adminTypes.ValidatorInfo{
			PubKey: pkBytes,
			// Power: 1,
		},
	}, nil
}

func (n *Node) TxQuery(ctx context.Context, hash types.Hash, prove bool) (*ktypes.TxQueryResponse, error) {
	tx, height, blkHash, blkIdx, err := n.bki.GetTx(hash)
	if err != nil {
		return nil, err
	}
	blkResults, err := n.bki.Results(blkHash)
	if err != nil {
		return nil, err
	}
	if int(blkIdx) >= len(blkResults) {
		return nil, errors.New("invalid block index")
	}
	res := blkResults[blkIdx]
	return &ktypes.TxQueryResponse{
		Tx:     tx,
		Hash:   hash,
		Height: height,
		Result: &res,
	}, nil
}

func (n *Node) BroadcastTx(ctx context.Context, tx *ktypes.Transaction, _ /*sync TODO*/ uint8) (*ktypes.ResultBroadcastTx, error) {
	rawTx := tx.Bytes()
	txHash := types.HashBytes(rawTx)

	if err := n.ce.CheckTx(ctx, tx); err != nil {
		return nil, err
	}

	n.mp.Store(txHash, tx)

	n.log.Infof("broadcasting new tx %v", txHash)
	n.announceTx(ctx, txHash, rawTx, n.host.ID())

	return &ktypes.ResultBroadcastTx{
		Hash: txHash,
		// Log and Code just for sync?
	}, nil
}

// ChainTx return tx info that is used in Chain rpc.
func (n *Node) ChainTx(hash types.Hash) (*chainTypes.Tx, error) {
	tx, height, blkHash, blkIdx, err := n.bki.GetTx(hash)
	if err != nil {
		return nil, err
	}
	blkResults, err := n.bki.Results(blkHash)
	if err != nil {
		return nil, err
	}
	if int(blkIdx) >= len(blkResults) {
		return nil, errors.New("invalid block index")
	}
	res := blkResults[blkIdx]
	return &chainTypes.Tx{
		Hash:     hash,
		Height:   height,
		Index:    blkIdx,
		Tx:       tx,
		TxResult: &res,
	}, nil
}

// ChainUnconfirmedTx return unconfirmed tx info that is used in Chain rpc.
func (n *Node) ChainUnconfirmedTx(limit int) (int, []types.NamedTx) {
	total := n.mp.Size()
	if limit <= 0 {
		return total, nil
	}
	return n.mp.Size(), n.mp.PeekN(limit)
}

func (n *Node) BlockHeight() int64 {
	height, _, _ := n.bki.Best()
	return height
}

func (n *Node) ConsensusParams() *ktypes.ConsensusParams {
	return n.ce.ConsensusParams()
}

func (n *Node) AbortBlockExecution(height int64, txIDs []types.Hash) error {
	return n.ce.CancelBlockExecution(height, txIDs)
}

func (n *Node) Role() types.Role {
	return n.ce.Role()
}

var RequiredStreamProtocols = []protocol.ID{
	ProtocolIDDiscover,
	ProtocolIDTx,
	ProtocolIDTxAnn,
	ProtocolIDBlockHeight,
	ProtocolIDBlock,
	ProtocolIDBlkAnn,
	ProtocolIDBlockPropose,
	pubsub.GossipSubID_v12,
	ProtocolIDSnapshotCatalog,
	ProtocolIDSnapshotChunk,
	ProtocolIDSnapshotMeta,
}

func (n *Node) checkPeerProtos(ctx context.Context, peer peer.ID) error {
	return peers.RequirePeerProtos(ctx, n.host.Peerstore(), peer, RequiredStreamProtocols...)
}

type randSrc struct{}

func (randSrc) Uint64() uint64 {
	var b [8]byte
	rand.Read(b[:])
	return binary.LittleEndian.Uint64(b[:])
}

var rng = mrand2.New(randSrc{})

func (n *Node) peers() []peer.ID {
	peers := n.host.Network().Peers()
	rng.Shuffle(len(peers), func(i, j int) {
		peers[i], peers[j] = peers[j], peers[i]
	})
	return peers
}

// NewKey generates a new private key from a reader, which should provide random data.
func NewKey(r io.Reader) crypto.PrivateKey {
	privKey, _, err := crypto.GenerateSecp256k1Key(r)
	if err != nil {
		panic(err)
	}

	return privKey
}

func newHost(ip string, port uint64, privKey crypto.PrivateKey, wl *peers.WhitelistGater) (host.Host, error) {
	// convert to the libp2p crypto key type
	var privKeyP2P p2pcrypto.PrivKey
	var err error
	switch kt := privKey.(type) {
	case *crypto.Secp256k1PrivateKey:
		privKeyP2P, err = p2pcrypto.UnmarshalSecp256k1PrivateKey(privKey.Bytes())
	case *crypto.Ed25519PrivateKey: // TODO
		privKeyP2P, err = p2pcrypto.UnmarshalEd25519PrivateKey(privKey.Bytes())
	default:
		err = fmt.Errorf("unknown private key type %T", kt)
	}
	if err != nil {
		return nil, err
	}

	sourceMultiAddr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%d", ip, port))

	// listenAddrs := libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/0", "/ip4/0.0.0.0/tcp/0/ws")

	// The BasicConnManager can keep connections below an upper limit, dropping
	// down to some lower limit (presumably to keep a dynamic peer list), but it
	// won't try to initiation new connections to reach the min. Maybe this will
	// be helpful later, so leaving as a comment:

	// cm, err := connmgr.NewConnManager(60, 100, connmgr.WithGracePeriod(20*time.Minute)) // TODO: absorb this into peerman
	// if err != nil {
	// 	return nil, nil, err
	// }

	// libp2p.New is fine with a nil interface (not an non-nil interface to a
	// nil concrete instance).
	var cg connmgr.ConnectionGater
	if wl != nil { // cfg.P2P.PrivateMode
		cg = wl
	}

	h, err := libp2p.New(
		libp2p.Transport(tcp.NewTCPTransport),
		libp2p.Security(noise.ID, noise.New), // modified TLS based on node-ID
		libp2p.ListenAddrs(sourceMultiAddr),
		// listenAddrs,
		libp2p.Identity(privKeyP2P),
		libp2p.ConnectionGater(cg),
		// libp2p.ConnectionManager(cm),
	) // libp2p.RandomIdentity, in-mem peer store, ...
	if err != nil {
		return nil, err
	}

	// cg.SetPeerStore(h.Peerstore())

	return h, nil
}

func maHostPort(addr multiaddr.Multiaddr) (host, port, protocol string) {
	port, _ = addr.ValueForProtocol(multiaddr.P_TCP)
	protocol = "ip4"
	host, _ = addr.ValueForProtocol(multiaddr.P_IP4)
	if host == "" {
		host, _ = addr.ValueForProtocol(multiaddr.P_IP6)
		protocol = "ip6"
	}
	return
}

func hostPort(host host.Host) ([]string, []int, []string) {
	var addrStr []string
	var ports []int
	var protocols []string              // ip4 or ip6
	for _, addr := range host.Addrs() { // host.Network().ListenAddresses()
		host, portStr, protocol := maHostPort(addr)
		port, _ := strconv.Atoi(portStr)
		ports = append(ports, port)
		addrStr = append(addrStr, host)
		protocols = append(protocols, protocol)
	}

	return addrStr, ports, protocols
}

func makePeerAddrInfo(addr string) (*peer.AddrInfo, error) {
	// Turn the destination into a multiaddr.
	maddr, err := multiaddr.NewMultiaddr(addr)
	if err != nil {
		return nil, err
	}

	// Extract the peer ID from the multiaddr.
	return peer.AddrInfoFromP2pAddr(maddr)
}

func connectPeer(ctx context.Context, addr string, host host.Host) (*peer.AddrInfo, error) {
	// Extract the peer ID and address info from the multiaddr.
	info, err := makePeerAddrInfo(addr)
	if err != nil {
		return nil, err
	}

	// Add the destination's peer multiaddress in the peerstore.
	// This will be used during connection and stream creation by libp2p.
	// host.Peerstore().AddAddrs(info.ID, info.Addrs, peerstore.PermanentAddrTTL)

	return info, host.Connect(ctx, *info)
}

// TODO: this is WRONG considering paths like ~user. We should rewrite this
// correctly, for both ~/ and ~user/ and without assuming a platform separator.
func ExpandPath(path string) (string, error) {
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		path = filepath.Join(home, path[2:])
	}
	return filepath.Abs(path)
}

func (n *Node) InCatchup() bool {
	return n.ce.InCatchup()
}
