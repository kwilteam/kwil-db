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
	"slices"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"kwil/crypto"
	"kwil/log"
	"kwil/node/consensus"
	"kwil/node/mempool"
	"kwil/node/peers"
	"kwil/node/store"
	"kwil/node/types"

	"github.com/libp2p/go-libp2p"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	p2pcrypto "github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	noise "github.com/libp2p/go-libp2p/p2p/security/noise"
	"github.com/libp2p/go-libp2p/p2p/transport/tcp"
	"github.com/multiformats/go-multiaddr"
	//libp2ptls "github.com/libp2p/go-libp2p/p2p/security/tls"
)

const (
	blockTxCount    = 50              // for "mining"
	dummyTxSize     = 123_000         // for broadcast
	dummyTxInterval = 1 * time.Second // broadcast freq
)

type ConsensusEngine interface {
	AcceptProposal(height int64, blkID, prevBlkID types.Hash, leaderSig []byte, timestamp int64) bool
	NotifyBlockProposal(blk *types.Block)

	AcceptCommit(height int64, blkID types.Hash, appHash types.Hash, leaderSig []byte) bool
	NotifyBlockCommit(blk *types.Block, appHash types.Hash)

	NotifyACK(validatorPK []byte, ack types.AckRes)
	NotifyResetState(height int64)

	// Gonna remove this once we have the commit results such as app hash and the tx results stored in the block store.

	Start(ctx context.Context, proposerBroadcaster consensus.ProposalBroadcaster,
		blkAnnouncer consensus.BlkAnnouncer, ackBroadcaster consensus.AckBroadcaster,
		blkRequester consensus.BlkRequester, stateResetter consensus.ResetStateBroadcaster)

	// Note: Not sure if these are needed here, just for separate of concerns:
	// p2p stream handlers role is to download the messages and pass it to the
	// respective modules to process it and we probably should not be triggering any consensus
	// affecting methods.

	// ProcessProposal(blk *types.Block, cb func(ack bool, appHash types.Hash) error)
	// ProcessACK(validatorPK []byte, ack types.AckRes)
	// CommitBlock(blk *types.Block, appHash types.Hash) error
}

type PeerManager interface {
	network.Notifiee
	Start(context.Context) error
	ConnectedPeers() []types.PeerInfo
	KnownPeers() []types.PeerInfo
}

type Node struct {
	bki types.BlockStore
	mp  types.MemPool
	ce  ConsensusEngine
	log log.Logger

	pm PeerManager // *peers.PeerMan
	// pf *prefetch

	ackChan  chan AckRes         // from consensus engine, to gossip to leader
	resetMsg chan ConsensusReset // gossiped in from peers, to consensus engine

	host   host.Host
	pex    bool
	pubkey crypto.PublicKey
	leader atomic.Bool
	dir    string
	wg     sync.WaitGroup
	close  func() error

	// TODO: get both rol and valSet from CE, if really needed by Node
	role   types.Role
	valSet map[string]types.Validator
}

func addClose(close, top func() error) func() error {
	return func() error { err := top(); return errors.Join(err, close()) }
}

// NewNode creates a new node. For now we are using functional options, but this
// may be better suited by a config struct, or a hybrid where some settings are
// required, such as identity.
func NewNode(cfg *Config, opts ...Option) (*Node, error) {
	options := &options{}
	for _, opt := range opts {
		opt(options)
	}

	logger := cfg.Logger
	if logger == nil {
		// logger = log.DiscardLogger // prod
		logger = log.New(log.WithWriter(os.Stdout), log.WithLevel(log.LevelDebug),
			log.WithFormat(log.FormatUnstructured)) // dev
	}

	close := func() error { return nil }

	host := options.host
	if host == nil {
		var err error
		host, err = newHost(cfg.P2P.IP, cfg.P2P.Port, cfg.PrivKey)
		if err != nil {
			return nil, err
		}
	}

	mp := options.mp
	if options.mp == nil {
		mp = mempool.New()
	}

	addrBookPath := filepath.Join(cfg.RootDir, "addrbook.json")
	pm, err := peers.NewPeerMan(cfg.P2P.Pex, addrBookPath,
		logger.New("PEERS"),
		host, // tooo much, become minimal interface
		func(ctx context.Context, peerID peer.ID) ([]peer.AddrInfo, error) {
			return requestPeers(ctx, peerID, host, logger)
		}, neededProtocols)
	if err != nil {
		return nil, err
	}

	bs := options.bs
	if bs == nil {
		blkStrDir := filepath.Join(cfg.RootDir, "blockstore")
		var err error
		bs, err = store.NewBlockStore(blkStrDir)
		if err != nil {
			return nil, err
		}
	}
	close = addClose(close, bs.Close) //close db after stopping p2p
	close = addClose(close, host.Close)

	privKey := cfg.PrivKey
	pubkey := privKey.Public()

	leader := cfg.Genesis.Validators[0].PubKey
	valSet := make(map[string]types.Validator)
	for _, val := range cfg.Genesis.Validators {
		valSet[hex.EncodeToString(val.PubKey)] = val
	}

	leaderPubKey, err := crypto.UnmarshalSecp256k1PublicKey(leader)
	if err != nil {
		return nil, err
	}

	role := types.RoleValidator
	if pubkey.Equals(leaderPubKey) {
		role = types.RoleLeader
	}

	ceCfg := &consensus.Config{
		Signer:         privKey,
		Dir:            cfg.RootDir,
		Leader:         leaderPubKey,
		Mempool:        mp,
		BlockStore:     bs,
		ValidatorSet:   valSet,
		Logger:         logger.New("CONS"),
		ProposeTimeout: cfg.Consensus.ProposeTimeout,
	}
	ce := consensus.New(ceCfg)
	if ce == nil {
		return nil, errors.New("failed to create consensus engine")
	}

	node := &Node{
		log:      logger,
		host:     host,
		pm:       pm,
		pubkey:   pubkey,
		pex:      cfg.P2P.Pex,
		mp:       mp,
		bki:      bs,
		ce:       ce,
		dir:      cfg.RootDir,
		ackChan:  make(chan AckRes, 1),
		resetMsg: make(chan ConsensusReset, 1),
		close:    close,
		role:     role,
		valSet:   valSet,
	}

	node.leader.Store(role == types.RoleLeader)

	host.SetStreamHandler(ProtocolIDTxAnn, node.txAnnStreamHandler)
	host.SetStreamHandler(ProtocolIDBlkAnn, node.blkAnnStreamHandler)
	host.SetStreamHandler(ProtocolIDBlock, node.blkGetStreamHandler)
	host.SetStreamHandler(ProtocolIDBlockHeight, node.blkGetHeightStreamHandler)
	host.SetStreamHandler(ProtocolIDTx, node.txGetStreamHandler)

	host.SetStreamHandler(ProtocolIDBlockPropose, node.blkPropStreamHandler)
	// host.SetStreamHandler(ProtocolIDACKProposal, node.blkAckStreamHandler)

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
	hosts, ports := hostPort(n.host)
	if len(hosts) == 0 {
		return nil
	}

	pubkeyType := n.pubkey.Type()
	addrs := make([]string, len(hosts))
	for i, h := range hosts {
		addrs[i] = FormatPeerString(n.pubkey.Bytes(), pubkeyType, h, ports[i])
	}
	return addrs
}

func (n *Node) MultiAddrs() []string {
	hosts, ports := hostPort(n.host)
	id := n.host.ID()
	if len(hosts) == 0 {
		return nil
	}
	addrs := make([]string, len(hosts))
	for i, h := range hosts {
		addrs[i] = fmt.Sprintf("/ip4/%s/tcp/%d/p2p/%s", h, ports[i], id)
	}
	return addrs
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

	bootpeers, err = peers.ConvertPeersToMultiAddr(bootpeers)
	if err != nil {
		return err
	}

	// connect to bootstrap peers, if any
	for _, peer := range bootpeers {
		peerInfo, err := connectPeer(ctx, peer, n.host)
		if err != nil {
			n.log.Errorf("failed to connect to %v: %v", peer, err)
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

	// Connect to peers in peer store.
	bootedPeers := n.pm.ConnectedPeers()
	for _, peerID := range n.host.Peerstore().Peers() {
		if slices.ContainsFunc(bootedPeers, func(p types.PeerInfo) bool {
			return p.ID == peerID
		}) {
			continue // already connected
		}
		if n.host.ID() == peerID {
			continue
		}
		peerAddrs := n.host.Peerstore().Addrs(peerID)
		if len(peerAddrs) == 0 {
			n.log.Warnf("No addresses found for peer %s, skipping.", peerID)
			continue
		}

		// Create AddrInfo from peerID and known addresses
		if err := n.host.Connect(ctx, peer.AddrInfo{
			ID:    peerID,
			Addrs: peerAddrs,
		}); err != nil {
			n.log.Warnf("Unable to connect to peer %s: %v", peerID, peers.CompressDialError(err))
		}
		n.log.Infof("Connected to address book peer %v", peerID)
	}

	if err := n.startAckGossip(ctx, ps); err != nil {
		cancel()
		return err
	}
	if err := n.startConsensusResetGossip(ctx, ps); err != nil {
		cancel()
		return err
	}

	// custom stream-based gossip uses txAnnStreamHandler and announceTx.
	// This dummy method will make create+announce new pretend transactions.
	// It also periodically rebroadcasts txns.
	n.startTxAnns(ctx, dummyTxInterval, 30*time.Second, dummyTxSize) // nogossip.go

	// mine is our block anns goroutine, which must be only for leader
	n.wg.Add(1)
	go func() {
		defer n.wg.Done()
		defer cancel()
		n.ce.Start(ctx, n.announceBlkProp, n.announceBlk, n.sendACK, n.getBlkHeight, n.sendReset)
	}()

	n.wg.Add(1)
	go func() {
		defer n.wg.Done()
		n.pm.Start(ctx)
	}()

	n.log.Info("Node started.")

	<-ctx.Done()

	n.wg.Wait()

	return n.close()
}

var neededProtocols = []protocol.ID{
	ProtocolIDDiscover,
	ProtocolIDTx,
	ProtocolIDTxAnn,
	ProtocolIDBlockHeight,
	ProtocolIDBlock,
	ProtocolIDBlkAnn,
	ProtocolIDBlockPropose,
	pubsub.GossipSubID_v12,
}

func (n *Node) checkPeerProtos(ctx context.Context, peer peer.ID) error {
	return peers.RequirePeerProtos(ctx, n.host.Peerstore(), peer, neededProtocols...)
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

func newHost(ip string, port uint64, privKey crypto.PrivateKey) (host.Host, error) {
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

	// cg := peers.NewProtocolGater(neededProtocols)

	h, err := libp2p.New(
		libp2p.Transport(tcp.NewTCPTransport),
		libp2p.Security(noise.ID, noise.New), // modified TLS based on node-ID
		libp2p.ListenAddrs(sourceMultiAddr),
		// listenAddrs,
		libp2p.Identity(privKeyP2P),
		// libp2p.ConnectionGater(cg),
	) // libp2p.RandomIdentity, in-mem peer store, ...
	if err != nil {
		return nil, err
	}

	// cg.SetPeerStore(h.Peerstore())

	return h, nil
}

func hostPort(host host.Host) ([]string, []int) {
	var addrStr []string
	var ports []int
	for _, addr := range host.Addrs() { // host.Network().ListenAddresses()
		ps, _ := addr.ValueForProtocol(multiaddr.P_TCP)
		port, _ := strconv.Atoi(ps)
		ports = append(ports, port)
		as, _ := addr.ValueForProtocol(multiaddr.P_IP4)
		if as == "" {
			as, _ = addr.ValueForProtocol(multiaddr.P_IP6)
		}
		addrStr = append(addrStr, as)
	}

	return addrStr, ports
}

func connectPeer(ctx context.Context, addr string, host host.Host) (*peer.AddrInfo, error) {
	// Turn the destination into a multiaddr.
	maddr, err := multiaddr.NewMultiaddr(addr)
	if err != nil {
		return nil, err
	}

	// Extract the peer ID from the multiaddr.
	info, err := peer.AddrInfoFromP2pAddr(maddr)
	if err != nil {
		return nil, err
	}

	// Add the destination's peer multiaddress in the peerstore.
	// This will be used during connection and stream creation by libp2p.
	// host.Peerstore().AddAddrs(info.ID, info.Addrs, peerstore.PermanentAddrTTL)

	return info, host.Connect(ctx, *info)
}

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
