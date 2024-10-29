package node

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	mrand2 "math/rand/v2"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"p2p/node/consensus"
	"p2p/node/mempool"
	"p2p/node/store"
	"p2p/node/types"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/libp2p/go-libp2p"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/crypto"
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
	AcceptProposal(height int64, prevHash types.Hash) bool
	NotifyBlockProposal(blk *types.Block)

	AcceptCommit(height int64, blkID types.Hash, appHash types.Hash) bool
	NotifyBlockCommit(blk *types.Block, appHash types.Hash)

	NotifyACK(validatorPK []byte, ack types.AckRes)

	// Gonna remove this once we have the commit results such as app hash and the tx results stored in the block store.

	Start(ctx context.Context, proposerBroadcaster consensus.ProposalBroadcaster, blkAnnouncer consensus.BlkAnnouncer, ackBroadcaster consensus.AckBroadcaster, blkRequester consensus.BlkRequester)

	// Note: Not sure if these are needed here, just for seperate of concerns:
	// p2p stream handlers role is to downlaod the messages and pass it to the
	// respective modules to process it and we probably should not be triggering any consensus
	// affecting methods.

	// ProcessProposal(blk *types.Block, cb func(ack bool, appHash types.Hash) error)
	// ProcessACK(validatorPK []byte, ack types.AckRes)
	// CommitBlock(blk *types.Block, appHash types.Hash) error
}

// parallel CE dev is leading node and p2p dev, so this "Node" currently uses an
// old interface for a working mock CE. Node code should be updated with final interface

type ConsensusEngineForToyNode interface { // quick and dirty for toy node prototype
	Start(context.Context) error

	AcceptProposalID(height int64, prevHash types.Hash) bool
	ProcessProposal(blk *types.Block, cb func(ack bool, appHash *types.Hash) error)

	ProcessACK(validatorPK []byte, ack types.AckRes)

	AcceptCommit(height int64, blkID types.Hash, appHash types.Hash) bool
	CommitBlock(blk *types.Block, appHash types.Hash) error

	BlockLeaderStream() <-chan *types.QualifiedBlock
}

type Node struct {
	bki types.BlockStore
	mp  types.MemPool
	ce  ConsensusEngine

	pm *peerMan
	// pf *prefetch

	ackChan chan AckRes // from consensus engine, to gossip to leader

	host   host.Host
	pex    bool
	leader atomic.Bool
	dir    string
	wg     sync.WaitGroup
	close  func() error

	role   types.Role
	valSet map[string]types.Validator
}

func addClose(close, top func() error) func() error {
	return func() error { err := top(); return errors.Join(err, close()) }
}

// Config is the configuration for a node. Although everything is presently set
// with functional options, some settings may be required, such as identity, and
// thus are best provided by a config struct.
type Config struct {
	PrivKey []byte
	// Port    uint16
	// Dir     string
}

// NewNode creates a new node. For now we are using functional options, but this
// may be better suited by a config struct, or a hybrid where some settings are
// required, such as identity.
func NewNode(dir string, opts ...Option) (*Node, error) {
	options := &options{}
	for _, opt := range opts {
		opt(options)
	}

	close := func() error { return nil }

	host := options.host
	if host == nil {
		var err error
		host, err = newHost(options.port, options.privKey)
		if err != nil {
			return nil, err
		}
	}

	mp := options.mp
	if options.mp == nil {
		mp = mempool.New()
	}

	pm := &peerMan{h: host}
	if options.pex {
		host.SetStreamHandler(ProtocolIDDiscover, pm.discoveryStreamHandler)
	} else {
		host.SetStreamHandler(ProtocolIDDiscover, func(s network.Stream) {
			s.Close()
		})
	}

	bs := options.bs
	var err error
	if bs == nil {
		blkStrDir := filepath.Join(dir, "blockstore")
		bs, err = store.NewBlockStore(blkStrDir)
		if err != nil {
			return nil, err
		}
	}
	close = addClose(close, bs.Close) //close db after stopping p2p
	close = addClose(close, host.Close)

	ce := consensus.New(options.role, host.ID(), dir, mp, bs, options.valSet)
	if ce == nil {
		return nil, errors.New("failed to create consensus engine")
	}
	// ce.BeLeader(leader)

	node := &Node{
		host:    host,
		pm:      pm,
		pex:     options.pex,
		mp:      mp,
		bki:     bs,
		ce:      ce,
		dir:     dir,
		ackChan: make(chan AckRes, 1),
		close:   close,
		role:    options.role,
		valSet:  options.valSet,
	}

	node.leader.Store(options.role == types.RoleLeader)

	host.SetStreamHandler(ProtocolIDTxAnn, node.txAnnStreamHandler)
	host.SetStreamHandler(ProtocolIDBlkAnn, node.blkAnnStreamHandler)
	host.SetStreamHandler(ProtocolIDBlock, node.blkGetStreamHandler)
	host.SetStreamHandler(ProtocolIDBlockHeight, node.blkGetHeightStreamHandler)
	host.SetStreamHandler(ProtocolIDTx, node.txGetStreamHandler)

	host.SetStreamHandler(ProtocolIDBlockPropose, node.blkPropStreamHandler)
	// host.SetStreamHandler(ProtocolIDACKProposal, node.blkAckStreamHandler)

	return node, nil
}

func (n *Node) Addr() string {
	hosts, ports := hostPort(n.host)
	id := n.host.ID()
	if len(hosts) == 0 {
		return ""
	}
	return fmt.Sprintf("/ip4/%s/tcp/%d/p2p/%s\n", hosts[0], ports[0], id)
}

func (n *Node) Dir() string {
	return n.dir
}

func (n *Node) ID() string {
	return n.host.ID().String()
}

// Start begins tx and block gossip, connects to any bootstrap peers, and begins
// peer discovery.
func (n *Node) Start(ctx context.Context, peers ...string) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	ps, err := pubsub.NewGossipSub(ctx, n.host)
	if err != nil {
		return err
	}
	if err := n.startAckGossip(ctx, ps); err != nil { // gossip.go
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
		n.ce.Start(ctx, n.announceBlkProp, n.announceBlk, n.sendACK, n.getBlkHeight)
	}()

	// mine is our block anns goroutine, which must be only for leader
	// n.wg.Add(1)
	// go func() {
	// 	defer n.wg.Done()
	// 	defer cancel()
	// 	n.announceBlocks(ctx)
	// }()

	// connect to bootstrap peers, if any
	for _, peer := range peers {
		peerInfo, err := connectPeer(ctx, peer, n.host)
		if err != nil {
			log.Printf("failed to connect to %v: %v", peer, err)
			continue
		}
		log.Println("connected to ", peerInfo)
		if err = n.checkPeerProtos(ctx, peerInfo.ID); err != nil {
			log.Printf("WARNING: peer does not support required protocols %v: %v", peer, err)
			if err = n.host.Network().ClosePeer(peerInfo.ID); err != nil {
				log.Printf("failed to disconnect from %v: %v", peer, err)
			}
			// n.host.Peerstore().RemovePeer()
			continue
		}
		// n.host.ConnManager().TagPeer(peerID, "validatorish", 1)
	} // else would use persistent peer store (address book)

	n.wg.Add(1)
	go func() {
		defer n.wg.Done()
		if !n.pex {
			return
		}
		for {
			// discover for this node
			peerChan, err := n.pm.FindPeers(ctx, "kwil_namespace")
			if err != nil {
				log.Println("FindPeers:", err)
			} else {
				go func() {
					for peer := range peerChan {
						addPeerToPeerStore(n.host.Peerstore(), peer)
					}
				}()
			}

			select {
			case <-ctx.Done():
				return
			case <-time.After(20 * time.Second):
			}
		}
	}()

	log.Println("node tx/block gossip started, and peer discovery enabled")

	<-ctx.Done()

	n.wg.Wait()

	return n.close()
}

func (n *Node) checkPeerProtos(ctx context.Context, peer peer.ID) error {
	for _, pid := range []protocol.ID{
		ProtocolIDDiscover,
		ProtocolIDTx,
		ProtocolIDTxAnn,
		ProtocolIDBlock,
		ProtocolIDBlkAnn,
		ProtocolIDBlockPropose,
		// ProtocolIDACKProposal,
		// pubsub.GossipSubID_v12,
	} {
		ok, err := checkProtocolSupport(ctx, n.host, peer, pid)
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("protocol not supported: %v", pid)
		}
	}
	return nil
}

// announceBlocks announces new committed and proposed blocks when the leader.
// func (n *Node) announceBlocks(ctx context.Context) {
// 	blkChan := n.ce.BlockLeaderStream()

// 	for {
// 		var blk *types.QualifiedBlock
// 		select {
// 		case <-ctx.Done():
// 			return
// 		case blk = <-blkChan:
// 		}

// 		if blk.Proposed {
// 			go n.announceBlkProp(ctx, blk.Block, n.host.ID())
// 		} else {
// 			go n.announceBlk(ctx, blk.Block, blk.AppHash, n.host.ID())
// 		}
// 	}
// }

func (n *Node) txGetStreamHandler(s network.Stream) {
	defer s.Close()

	var req txHashReq
	if _, err := req.ReadFrom(s); err != nil {
		fmt.Println("bad get tx req:", err)
		return
	}

	// first check mempool
	rawTx := n.mp.Get(req.Hash)
	if rawTx != nil {
		s.Write(rawTx)
		return
	}

	// this is racy, and should be different in product

	// then confirmed tx index
	_, rawTx, err := n.bki.GetTx(req.Hash)
	if err != nil {
		if !errors.Is(err, types.ErrNotFound) {
			log.Println("unexpected GetTx error:", err)
		}
		s.Write(noData) // don't have it
	} else {
		s.Write(rawTx)
	}

	// NOTE: response could also include conf/unconf or block height (-1 or N)
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

func NewKey(r io.Reader) crypto.PrivKey {
	// priv, _ := secp256k1.GeneratePrivateKeyFromRand(r)
	// privECDSA := priv.ToECDSA()
	// privECDSA.Curve = elliptic.P256()
	// privKey, _, _ := crypto.ECDSAKeyPairFromKey(privECDSA)
	//privKey, _, err := crypto.GenerateECDSAKeyPair(r)
	pk, err := secp256k1.GeneratePrivateKeyFromRand(r)
	if err != nil {
		panic(err)
	}
	privKey := (*crypto.Secp256k1PrivateKey)(pk)
	// privKey, _, err := crypto.GenerateSecp256k1Key(r)
	return privKey
}

func newHost(port uint64, privKey []byte) (host.Host, error) {
	privKeyP2P, err := crypto.UnmarshalSecp256k1PrivateKey(privKey) // rypto.UnmarshalECDSAPrivateKey(privKey)
	if err != nil {
		return nil, err
	}
	sourceMultiAddr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port))

	// listenAddrs := libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/0", "/ip4/0.0.0.0/tcp/0/ws")

	// TODO: use persistent peerstore

	return libp2p.New(
		libp2p.Transport(tcp.NewTCPTransport),
		libp2p.Security(noise.ID, noise.New), // modified TLS based on node-ID
		libp2p.ListenAddrs(sourceMultiAddr),
		// listenAddrs,
		libp2p.Identity(privKeyP2P),
	) // libp2p.RandomIdentity, in-mem peer store, ...
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

		log.Println("Listening on", addr)
	}

	return addrStr, ports
}

func connectPeer(ctx context.Context, addr string, host host.Host) (*peer.AddrInfo, error) {
	// Turn the destination into a multiaddr.
	maddr, err := multiaddr.NewMultiaddr(addr)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	// Extract the peer ID from the multiaddr.
	info, err := peer.AddrInfoFromP2pAddr(maddr)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	// Add the destination's peer multiaddress in the peerstore.
	// This will be used during connection and stream creation by libp2p.
	// host.Peerstore().AddAddrs(info.ID, info.Addrs, peerstore.PermanentAddrTTL)

	return info, host.Connect(ctx, *info)
}

func checkProtocolSupport(_ context.Context, h host.Host, peerID peer.ID, protoIDs ...protocol.ID) (bool, error) {
	supported, err := h.Peerstore().SupportsProtocols(peerID, protoIDs...)
	if err != nil {
		return false, fmt.Errorf("Failed to check protocols for peer %v: %w", peerID, err)
	}
	return len(protoIDs) == len(supported), nil

	// supportedProtos, err := h.Peerstore().GetProtocols(peerID)
	// if err != nil {
	// 	return false, err
	// }
	// log.Printf("protos supported by %v: %v\n", peerID, supportedProtos)

	// for _, protoID := range protoIDs {
	// 	if !slices.Contains(supportedProtos, protoID) {
	// 		return false, nil
	// 	}
	// }
	// return true, nil
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
