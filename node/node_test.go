package node

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net"
	"os"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/kwilteam/kwil-db/config"
	"github.com/kwilteam/kwil-db/core/crypto"
	"github.com/kwilteam/kwil-db/core/crypto/auth"
	"github.com/kwilteam/kwil-db/core/log"
	ktypes "github.com/kwilteam/kwil-db/core/types"
	blockprocessor "github.com/kwilteam/kwil-db/node/block_processor"
	"github.com/kwilteam/kwil-db/node/consensus"
	"github.com/kwilteam/kwil-db/node/mempool"
	"github.com/kwilteam/kwil-db/node/store/memstore"
	"github.com/kwilteam/kwil-db/node/types"

	p2pcrypto "github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	mock "github.com/libp2p/go-libp2p/p2p/net/mock"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/require"
)

var blackholeIP6 = net.ParseIP("100::")

func newTestHost(t *testing.T, mn mock.Mocknet) ([]byte, host.Host) {
	privKey, _, err := p2pcrypto.GenerateSecp256k1Key(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate private key: %v", err)
	}
	id, err := peer.IDFromPrivateKey(privKey)
	if err != nil {
		t.Fatalf("Failed to get private key: %v", err)
	}
	suffix := id
	if len(id) > 8 {
		suffix = id[len(id)-8:]
	}
	ip := append(net.IP{}, blackholeIP6...)
	copy(ip[net.IPv6len-len(suffix):], suffix)
	addr, err := ma.NewMultiaddr(fmt.Sprintf("/ip6/%s/tcp/4242", ip))
	if err != nil {
		t.Fatalf("Failed to create multiaddress: %v", err)
	}
	// t.Log(addr) // e.g. /ip6/100::1bb1:760e:df55:9ed1/tcp/4242
	host, err := mn.AddPeer(privKey, addr)
	if err != nil {
		t.Fatalf("Failed to add peer to mocknet: %v", err)
	}

	pkBytes, err := privKey.Raw()
	if err != nil {
		t.Fatalf("Failed to get private key bytes: %v", err)
	}
	return pkBytes, host
}

func makeTestHosts(t *testing.T, nNodes, nExtraHosts int, blockInterval time.Duration) ([]*Node, []host.Host, mock.Mocknet) {
	mn := mock.New()
	t.Cleanup(func() {
		mn.Close()
	})

	defaultConfigSet := config.DefaultConfig()
	defaultConfigSet.Consensus.ProposeTimeout = config.Duration(blockInterval)

	var nodes []*Node
	var hosts []host.Host
	// var privKeys []*crypto.Secp256k1PrivateKey

	for range nNodes {
		pk, h := newTestHost(t, mn)
		t.Logf("node host is %v", h.ID())

		priv, err := crypto.UnmarshalSecp256k1PrivateKey(pk)
		if err != nil {
			t.Fatalf("Failed to unmarshal private key: %v", err)
		}

		// memory block store
		bs := memstore.NewMemBS()
		// dummy CE
		ce := &dummyCE{}

		rootDir := t.TempDir()
		t.Logf("node root dir: %s", rootDir)

		cfg := &Config{
			RootDir: rootDir,
			PrivKey: priv,
			Logger:  log.DiscardLogger,
			P2P:     &defaultConfigSet.P2P,
			// DB unused
			DBConfig:    &defaultConfigSet.DB,
			Statesync:   &defaultConfigSet.StateSync,
			Mempool:     mempool.New(),
			BlockStore:  bs,
			Snapshotter: newSnapshotStore(),
			Consensus:   ce,
			BlockProc:   &dummyBP{},
		}
		node, err := NewNode(cfg, WithHost(h))
		if err != nil {
			t.Fatalf("Failed to create Node 1: %v", err)
		}

		// privKeys = append(privKeys, priv)
		nodes = append(nodes, node)
	}

	for range nExtraHosts {
		_, h := newTestHost(t, mn)
		setupStreamHandlers(t, h)
		hosts = append(hosts, h)
	}

	time.Sleep(50 * time.Millisecond)

	return nodes, hosts, mn
}

func linkAll(t *testing.T, mn mock.Mocknet) {
	if err := mn.LinkAll(); err != nil {
		t.Fatalf("Failed to link hosts: %v", err)
	}
	if err := mn.ConnectAllButSelf(); err != nil {
		t.Fatalf("Failed to connect hosts: %v", err)
	}
	time.Sleep(100 * time.Millisecond)
}

func linkPeers(t *testing.T, mn mock.Mocknet, p1, p2 peer.ID) {
	if _, err := mn.LinkPeers(p1, p2); err != nil {
		t.Fatalf("Failed to link hosts: %v", err)
	}
	if _, err := mn.ConnectPeers(p1, p2); err != nil {
		t.Fatalf("Failed to connect hosts: %v", err)
	}
	time.Sleep(100 * time.Millisecond)
}

func startNodes(t *testing.T, nodes []*Node) {
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	t.Cleanup(func() {
		cancel()
		wg.Wait()
	})

	for _, n := range nodes {
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer os.RemoveAll(n.Dir())
			n.Start(ctx)
		}()
	}
}

func fakeAppHash(height int64) types.Hash {
	return types.HashBytes(binary.LittleEndian.AppendUint64(nil, uint64(height)))
}

func newTx(nonce uint64, sender, payload string) *ktypes.Transaction {
	return &ktypes.Transaction{
		Signature: &auth.Signature{},
		Body: &ktypes.TransactionBody{
			Description: "test",
			Payload:     []byte(payload),
			Fee:         big.NewInt(0),
			Nonce:       nonce,
		},
		Sender: []byte(sender),
	}
}

func createTestBlock(t *testing.T, height int64, numTxns int) (*ktypes.Block, types.Hash) {
	txns := make([]*ktypes.Transaction, numTxns)
	for i := range numTxns {
		txns[i] = newTx(uint64(i), "bob", strconv.FormatInt(height, 10)+strconv.Itoa(i)+
			strings.Repeat("data", 1000))
	}
	blk := ktypes.NewBlock(height, types.Hash{2, 3, 4}, types.Hash{6, 7, 8}, types.Hash{5, 5, 5},
		time.Unix(1729723553+height, 0), txns)
	return blk, fakeAppHash(height)
}

func newGenesis(t *testing.T, nodekeys [][]byte) ([]crypto.PrivateKey, *config.GenesisConfig) {
	var privKeys []crypto.PrivateKey
	for _, nodekey := range nodekeys {
		priv, err := crypto.UnmarshalSecp256k1PrivateKey(nodekey)
		if err != nil {
			t.Fatalf("Failed to unmarshal private key: %v", err)
		}
		privKeys = append(privKeys, priv)
	}

	genCfg := config.GenesisConfig{
		Leader:     privKeys[0].Public().Bytes(),
		Validators: []*ktypes.Validator{},
	}
	for _, priv := range privKeys {
		genCfg.Validators = append(genCfg.Validators, &ktypes.Validator{
			PubKey: priv.Public().Bytes(),
			Power:  1,
		})
	}
	return privKeys, &genCfg
}

func setupStreamHandlers(t *testing.T, h host.Host) {
	for _, proto := range RequiredStreamProtocols {
		h.SetStreamHandler(proto, func(s network.Stream) {
			t.Log("handling incoming stream for", proto)
			s.Close()
		})
	}
}

type dummyBP struct {
	vals []*ktypes.Validator
}

func (bp *dummyBP) GetValidators() []*ktypes.Validator { return bp.vals }

func (bp *dummyBP) SubscribeValidators() <-chan []*ktypes.Validator {
	return make(<-chan []*ktypes.Validator, 1)
}

// func hupStreamHandler(s network.Stream) { s.Close() }

var _ ConsensusEngine = &dummyCE{}

// dummyCE is a dummy consensus engine for testing the Node. The zero value is ready to
// use. Use the Fake() method to manipulate the behavior of the dummyCE.
type dummyCE struct {
	rejectProp   bool
	rejectCommit bool
	rejectACK    bool

	ackHandler         func(validatorPK []byte, ack types.AckRes)
	blockCommitHandler func(blk *ktypes.Block, appHash types.Hash)
	blockPropHandler   func(blk *ktypes.Block)
	resetStateHandler  func(height int64, txIDs []types.Hash)

	// mtx     sync.Mutex
	// gotACKs map[string]types.AckRes // from NotifyACK: string(validatorPK) -> AckRes

	// set by start
	proposerBroadcaster consensus.ProposalBroadcaster
	blkAnnouncer        consensus.BlkAnnouncer
	ackBroadcaster      consensus.AckBroadcaster
	blkRequester        consensus.BlkRequester
	stateResetter       consensus.ResetStateBroadcaster
}

func (ce *dummyCE) AcceptProposal(height int64, blkID, prevBlkID types.Hash, leaderSig []byte, timestamp int64) bool {
	return !ce.rejectProp
}

func (ce *dummyCE) AcceptCommit(height int64, blkID, appHash types.Hash, leaderSig []byte) bool {
	return !ce.rejectCommit
}

func (ce *dummyCE) NotifyBlockCommit(blk *ktypes.Block, appHash types.Hash) {
	if ce.blockCommitHandler != nil {
		ce.blockCommitHandler(blk, appHash)
		return
	}
}

func (ce *dummyCE) NotifyACK(validatorPK []byte, ack types.AckRes) {
	if ce.ackHandler != nil {
		ce.ackHandler(validatorPK, ack)
		return
	}
	// ce.mtx.Lock()
	// defer ce.mtx.Unlock()
	// ce.gotACKs[string(validatorPK)] = ack
}

func (ce *dummyCE) AcceptACK() bool {
	return !ce.rejectACK
}

func (ce *dummyCE) NotifyResetState(height int64, txIDs []types.Hash) {
	if ce.resetStateHandler != nil {
		ce.resetStateHandler(height, txIDs)
		return
	}
}

func (ce *dummyCE) NotifyBlockProposal(blk *ktypes.Block) {
	if ce.blockPropHandler != nil {
		ce.blockPropHandler(blk)
		return
	}
}

func (ce *dummyCE) NotifyDiscoveryMessage(validatorPK []byte, height int64) {}

func (ce *dummyCE) Role() types.Role {
	return types.RoleLeader
}

func (ce *dummyCE) CheckTx(ctx context.Context, tx *ktypes.Transaction) error {
	return nil
}

func (ce *dummyCE) ConsensusParams() *ktypes.ConsensusParams {
	return nil
}

func (ce *dummyCE) InCatchup() bool {
	return false
}

func (ce *dummyCE) CancelBlockExecution(height int64, txIDs []types.Hash) error {
	return nil
}

func (ce *dummyCE) Start(ctx context.Context, proposerBroadcaster consensus.ProposalBroadcaster,
	blkAnnouncer consensus.BlkAnnouncer, ackBroadcaster consensus.AckBroadcaster,
	blkRequester consensus.BlkRequester, stateResetter consensus.ResetStateBroadcaster,
	discReqBroadcaster consensus.DiscoveryReqBroadcaster, txBroadcaster blockprocessor.BroadcastTxFn) error {
	ce.proposerBroadcaster = proposerBroadcaster
	ce.blkAnnouncer = blkAnnouncer
	ce.ackBroadcaster = ackBroadcaster
	ce.blkRequester = blkRequester
	ce.stateResetter = stateResetter
	return nil
}

// Fake gets the methods to talk back to the Node, dictating CE logic manually.
// These could just be methods on the CE, but this makes their relationship clearer.
func (ce *dummyCE) Fake() *faker {
	return (*faker)(ce)
}

type faker dummyCE

func (f *faker) Propose(ctx context.Context, blk *ktypes.Block) {
	f.proposerBroadcaster(ctx, blk)
}

func (f *faker) ACK(ack bool, height int64, blkID types.Hash, appHash *types.Hash) error {
	return f.ackBroadcaster(ack, height, blkID, appHash)
}

func (f *faker) ResetState(height int64, txIDs []types.Hash) {
	f.stateResetter(height, txIDs)
}

func (f *faker) RequestBlock(ctx context.Context, height int64) {
	f.blkRequester(ctx, height)
}

func (f *faker) AnnounceBlock(ctx context.Context, blk *ktypes.Block, appHash types.Hash) {
	f.blkAnnouncer(ctx, blk, appHash)
}

func (f *faker) RejectNextProposal() {
	f.rejectProp = true
}

func (f *faker) RejectNextCommit() {
	f.rejectCommit = true
}

func (f *faker) SetACKHandler(ackHandler func(validatorPK []byte, ack types.AckRes)) {
	f.ackHandler = ackHandler
}

func (f *faker) SetBlockCommitHandler(blockCommitHandler func(blk *ktypes.Block, appHash types.Hash)) {
	f.blockCommitHandler = blockCommitHandler
}

func (f *faker) SetBlockPropHandler(blockPropHandler func(blk *ktypes.Block)) {
	f.blockPropHandler = blockPropHandler
}

func (f *faker) SetResetStateHandler(resetStateHandler func(height int64, txIDs []types.Hash)) {
	f.resetStateHandler = resetStateHandler
}

func TestPeerDiscoverStream(t *testing.T) {
	nodes, testHosts, mn := makeTestHosts(t, 2, 1, 5*time.Hour)
	// linkAll(t, mn)

	n1, n2 := nodes[0], nodes[1]
	h1, h2 := n1.host, n2.host
	pid1, pid2 := h1.ID(), h2.ID()

	// no need to startNodes to test this stream

	// pm1 := n1.pm

	th1 := testHosts[0]

	// connect h1 and test host (not h2)
	linkPeers(t, mn, pid1, th1.ID())

	ctx := context.Background()

	t.Run("discover myself w/ requestPeersProto", func(t *testing.T) {
		s, err := th1.NewStream(ctx, pid1, ProtocolIDDiscover)
		if err != nil {
			t.Fatalf("Failed create new stream: %v", err)
		}

		addrs, err := requestPeersProto(s)
		if err != nil {
			t.Fatalf("failed to read peer discover response: %v", err)
		}

		// They know me and only me
		if len(addrs) != 1 {
			t.Fatalf("expected one address, got %d", len(addrs))
		}

		require.Equal(t, addrs[0].ID, th1.ID())
	})

	t.Run("discover myself w/ requestPeers", func(t *testing.T) {
		addrs, err := requestPeers(ctx, pid1, th1, log.DiscardLogger)
		if err != nil {
			t.Fatalf("failed to read peer discover response: %v", err)
		}

		// They know me and only me
		if len(addrs) != 1 {
			t.Fatalf("expected one address, got %d", len(addrs))
		}

		require.Equal(t, addrs[0].ID, th1.ID())
	})

	t.Run("discover myself and h2 w/ requestPeers", func(t *testing.T) {
		// Connect h1 to h2
		linkPeers(t, mn, pid1, pid2)
		defer mn.UnlinkPeers(pid1, pid2)
		defer mn.DisconnectPeers(pid1, pid2)

		addrs, err := requestPeers(ctx, h1.ID(), th1, log.DiscardLogger)
		if err != nil {
			t.Fatalf("failed to read peer discover response: %v", err)
		}

		// They know me *and* h2 now
		if len(addrs) != 2 {
			t.Fatalf("expected 2 addresses, got %d", len(addrs))
		}

		if !slices.ContainsFunc(addrs, func(addr peer.AddrInfo) bool {
			return addr.ID == th1.ID()
		}) {
			t.Errorf("self was not included in returned addresses")
		}

		if !slices.ContainsFunc(addrs, func(addr peer.AddrInfo) bool {
			return addr.ID == pid2
		}) {
			t.Errorf("h2 was not included in returned addresses")
		}
	})
}

func TestStreamsBlockFetch(t *testing.T) {
	nodes, extraHosts, mn := makeTestHosts(t, 1, 1, 5*time.Hour)
	linkAll(t, mn)

	n1 := nodes[0]
	h1 := n1.host

	// to n1's block store, one block at height 1 with 2 txns
	blk1, appHash1 := createTestBlock(t, 1, 2)
	n1.bki.Store(blk1, appHash1)

	startNodes(t, nodes)

	h2 := extraHosts[0]

	time.Sleep(100 * time.Millisecond)

	// Link and connect the hosts (was here)
	// time.Sleep(100 * time.Millisecond)

	ctx := context.Background()

	t.Run("request unknown hash manually", func(t *testing.T) {
		// t.Parallel()
		s, err := h2.NewStream(ctx, h1.ID(), ProtocolIDBlock)
		if err != nil {
			t.Fatalf("Failed create new stream: %v", err)
		}
		defer s.Close()

		unknownHash := types.Hash{1}
		_, err = s.Write(unknownHash[:])
		if err != nil {
			t.Fatalf("Failed write to stream: %v", err)
		}

		// (*blockHashReq).ReadFrom should not hang, but should timeout (and error), and close stream on us

		b, err := io.ReadAll(s) // expect EOF (no error)
		if err != nil {
			t.Errorf("ReadAll: %v", err)
		} else if !bytes.Equal(b, noData) {
			t.Error("expected a no-data response, got", b)
		}
	})

	t.Run("request by hash using requestFrom, unknown block", func(t *testing.T) {
		// t.Parallel()
		unknownHash := types.Hash{1}
		req, _ := blockHashReq{unknownHash}.MarshalBinary() // knownHash[:]
		_, err := requestFrom(ctx, h2, h1.ID(), req, ProtocolIDBlock, 1e4)
		if err == nil {
			t.Errorf("expected error but got none")
		} else if !errors.Is(err, ErrNotFound) {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("request by hash manually, known", func(t *testing.T) {
		// t.Parallel()
		s, err := h2.NewStream(ctx, h1.ID(), ProtocolIDBlock)
		if err != nil {
			t.Fatalf("Failed create new stream: %v", err)
		}
		defer s.Close()

		knownHash := blk1.Hash()
		_, err = s.Write(knownHash[:])
		if err != nil {
			t.Fatalf("Failed write to stream: %v", err)
		}

		// (*blockHashReq).ReadFrom should not hang, but should timeout (and error), and close stream on us

		b, err := io.ReadAll(s) // expect EOF (no error)
		if err != nil {
			t.Errorf("ReadAll: %v", err)
		} else if bytes.Equal(b, noData) {
			t.Error("expected data, got", b)
		}
	})

	t.Run("request by hash using requestFrom, known block", func(t *testing.T) {
		// t.Parallel()
		knownHash := blk1.Hash()
		req, _ := blockHashReq{knownHash}.MarshalBinary() // knownHash[:]
		resp, err := requestFrom(ctx, h2, h1.ID(), req, ProtocolIDBlock, 1e4)
		if err != nil {
			t.Errorf("ReadAll: %v", err)
		} else if bytes.Equal(resp, noData) {
			t.Error("expected data, got", resp)
		}
	})

	t.Run("request by height manually, unknown", func(t *testing.T) {
		// t.Parallel()
		s, err := h2.NewStream(ctx, h1.ID(), ProtocolIDBlockHeight)
		if err != nil {
			t.Fatalf("Failed create new stream: %v", err)
		}
		defer s.Close()

		var height int64
		err = binary.Write(s, binary.LittleEndian, height)
		if err != nil {
			t.Fatalf("Failed write to stream: %v", err)
		}

		b, err := io.ReadAll(s)
		if err != nil {
			t.Errorf("ReadAll: %v", err)
		} else if !bytes.Equal(b, noData) {
			t.Error("expected a no-data response, got", b)
		}
	})

	t.Run("request by height using requestFrom, unknown", func(t *testing.T) {
		// t.Parallel()
		var height int64
		req, _ := blockHeightReq{height}.MarshalBinary()
		_, err := requestFrom(ctx, h2, h1.ID(), req, ProtocolIDBlockHeight, 1e4)
		if err == nil {
			t.Errorf("expected error but got none")
		} else if !errors.Is(err, ErrNotFound) {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("request by height manually, known", func(t *testing.T) {
		// t.Parallel()
		s, err := h2.NewStream(ctx, h1.ID(), ProtocolIDBlockHeight)
		if err != nil {
			t.Fatalf("Failed create new stream: %v", err)
		}
		defer s.Close()

		var height int64 = 1
		err = binary.Write(s, binary.LittleEndian, height)
		if err != nil {
			t.Fatalf("Failed write to stream: %v", err)
		}

		b, err := io.ReadAll(s)
		if err != nil {
			t.Errorf("ReadAll: %v", err)
		} else if bytes.Equal(b, noData) {
			t.Error("expected a no-data response, got", b)
		} // else { t.Log(len(b)) }
	})

	t.Run("request by height using requestFrom, known", func(t *testing.T) {
		// t.Parallel()
		var height int64 = 1
		req, _ := blockHeightReq{height}.MarshalBinary()
		resp, err := requestFrom(ctx, h2, h1.ID(), req, ProtocolIDBlockHeight, 1e4)
		if err != nil {
			t.Errorf("ReadAll: %v", err)
		} else if bytes.Equal(resp, noData) {
			t.Error("expected data, got", resp)
		}
	})
}
