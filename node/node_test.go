package node

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/kwilteam/kwil-db/config"
	"github.com/kwilteam/kwil-db/core/crypto"
	"github.com/kwilteam/kwil-db/core/log"
	ktypes "github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/node/consensus"
	"github.com/kwilteam/kwil-db/node/store/memstore"
	"github.com/kwilteam/kwil-db/node/types"

	p2pcrypto "github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	mock "github.com/libp2p/go-libp2p/p2p/net/mock"
	ma "github.com/multiformats/go-multiaddr"
)

var blackholeIP6 = net.ParseIP("100::")

func newTestHost(t *testing.T, mn mock.Mocknet) ([]byte, host.Host, error) {
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
	return pkBytes, host, nil
}

func fakeAppHash(height int64) types.Hash {
	return types.HashBytes(binary.LittleEndian.AppendUint64(nil, uint64(height)))
}

func createTestBlock(height int64, numTxns int) (*types.Block, types.Hash) {
	txns := make([][]byte, numTxns)
	for i := range numTxns {
		txns[i] = []byte(strconv.FormatInt(height, 10) + strconv.Itoa(i) +
			strings.Repeat("data", 1000))
	}
	return types.NewBlock(height, types.Hash{2, 3, 4}, types.Hash{6, 7, 8}, types.Hash{5, 5, 5},
		time.Unix(1729723553+height, 0), txns), fakeAppHash(height)
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
		Validators: []ktypes.Validator{},
	}
	for _, priv := range privKeys {
		genCfg.Validators = append(genCfg.Validators, ktypes.Validator{
			PubKey: priv.Public().Bytes(),
			Power:  1,
		})
	}
	return privKeys, &genCfg
}

func setHupStreamHandlers(t *testing.T, h host.Host) {
	for _, proto := range neededProtocols {
		h.SetStreamHandler(proto, func(s network.Stream) {
			t.Log("handling incoming stream for", proto)
			s.Close()
		})
	}
}

// func hupStreamHandler(s network.Stream) { s.Close() }

var _ ConsensusEngine = &dummyCE{}

// dummyCE is a dummy consensus engine for testing the Node. The zero value is ready to
// use. Use the Fake() method to manipulate the behavior of the dummyCE.
type dummyCE struct {
	rejectProp   bool
	rejectCommit bool

	ackHandler         func(validatorPK []byte, ack types.AckRes)
	blockCommitHandler func(blk *types.Block, appHash types.Hash)
	blockPropHandler   func(blk *types.Block)
	resetStateHandler  func(height int64)

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

func (ce *dummyCE) NotifyBlockCommit(blk *types.Block, appHash types.Hash) {
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

func (ce *dummyCE) NotifyResetState(height int64) {
	if ce.resetStateHandler != nil {
		ce.resetStateHandler(height)
		return
	}
}

func (ce *dummyCE) NotifyBlockProposal(blk *types.Block) {
	if ce.blockPropHandler != nil {
		ce.blockPropHandler(blk)
		return
	}
}

func (ce *dummyCE) Start(ctx context.Context, proposerBroadcaster consensus.ProposalBroadcaster,
	blkAnnouncer consensus.BlkAnnouncer, ackBroadcaster consensus.AckBroadcaster,
	blkRequester consensus.BlkRequester, stateResetter consensus.ResetStateBroadcaster) {
	ce.proposerBroadcaster = proposerBroadcaster
	ce.blkAnnouncer = blkAnnouncer
	ce.ackBroadcaster = ackBroadcaster
	ce.blkRequester = blkRequester
	ce.stateResetter = stateResetter
}

// Fake gets the methods to talk back to the Node, dictating CE logic manually.
// These could just be methods on the CE, but this makes their relationship clearer.
func (ce *dummyCE) Fake() *faker {
	return (*faker)(ce)
}

type faker dummyCE

func (f *faker) Propose(ctx context.Context, blk *types.Block) {
	f.proposerBroadcaster(ctx, blk)
}

func (f *faker) ACK(ack bool, height int64, blkID types.Hash, appHash *types.Hash) error {
	return f.ackBroadcaster(ack, height, blkID, appHash)
}

func (f *faker) ResetState(height int64) {
	f.stateResetter(height)
}

func (f *faker) RequestBlock(ctx context.Context, height int64) {
	f.blkRequester(ctx, height)
}

func (f *faker) AnnounceBlock(ctx context.Context, blk *types.Block, appHash types.Hash) {
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

func (f *faker) SetBlockCommitHandler(blockCommitHandler func(blk *types.Block, appHash types.Hash)) {
	f.blockCommitHandler = blockCommitHandler
}

func (f *faker) SetBlockPropHandler(blockPropHandler func(blk *types.Block)) {
	f.blockPropHandler = blockPropHandler
}

func (f *faker) SetResetStateHandler(resetStateHandler func(height int64)) {
	f.resetStateHandler = resetStateHandler
}

func TestDualNodeMocknet(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	mn := mock.New()

	pk1, h1, err := newTestHost(t, mn)
	if err != nil {
		t.Fatalf("Failed to add peer to mocknet: %v", err)
	}
	bs1 := memstore.NewMemBS()

	pk2, h2, err := newTestHost(t, mn)
	if err != nil {
		t.Fatalf("Failed to add peer to mocknet: %v", err)
	}
	bs2 := memstore.NewMemBS()

	root1 := t.TempDir()
	root2 := t.TempDir()

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	t.Cleanup(func() {
		cancel()
		wg.Wait()
	})

	privKeys, genCfg := newGenesis(t, [][]byte{pk1, pk2})

	defaultConfigSet := config.DefaultConfig()

	log1 := log.New(log.WithName("NODE1"), log.WithWriter(os.Stdout), log.WithLevel(log.LevelDebug), log.WithFormat(log.FormatUnstructured))
	cfg1 := &Config{
		RootDir:   root1,
		PrivKey:   privKeys[0],
		Logger:    log1,
		Genesis:   *genCfg,
		Consensus: defaultConfigSet.Consensus,
		P2P:       defaultConfigSet.P2P, // mostly ignored as we are using WithHost below
	}
	node1, err := NewNode(cfg1, WithHost(h1), WithBlockStore(bs1))
	if err != nil {
		t.Fatalf("Failed to create Node 1: %v", err)
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer os.RemoveAll(node1.Dir())
		node1.Start(ctx)
	}()

	// time.Sleep(200 * time.Millisecond) // !!!! apparently, needs this if block store does not have latency, so there is a race condition somewhere in CE
	time.Sleep(20 * time.Millisecond)

	log2 := log.New(log.WithName("NODE2"), log.WithWriter(os.Stdout), log.WithLevel(log.LevelDebug), log.WithFormat(log.FormatUnstructured))
	cfg2 := &Config{
		RootDir:   root2,
		PrivKey:   privKeys[1],
		Logger:    log2,
		Genesis:   *genCfg,
		Consensus: defaultConfigSet.Consensus,
		P2P:       defaultConfigSet.P2P,
	}
	node2, err := NewNode(cfg2, WithHost(h2), WithBlockStore(bs2))
	if err != nil {
		t.Fatalf("Failed to create Node 2: %v", err)
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer os.RemoveAll(node2.Dir())
		node2.Start(ctx)
	}()

	// Link and connect the hosts
	if err := mn.LinkAll(); err != nil {
		t.Fatalf("Failed to link hosts: %v", err)
	}
	if err := mn.ConnectAllButSelf(); err != nil {
		t.Fatalf("Failed to connect hosts: %v", err)
	}

	// n1 := mn.Net(h1.ID())
	// links := mn.LinksBetweenPeers(h1.ID(), h2.ID())
	// ln := links[0]
	// net := ln.Networks()[0]
	// peers := net.Peers()
	// t.Log(peers)

	// run for a bit, checks stuff, do tests, like ensure blocks mine (TODO)...
	time.Sleep(4 * time.Second)
	cancel()
	wg.Wait()

}

func TestStreamsBlockFetch(t *testing.T) {
	mn := mock.New()

	pk1, h1, err := newTestHost(t, mn)
	if err != nil {
		t.Fatalf("Failed to add peer to mocknet: %v", err)
	}

	// t.Logf("node host is %v", h1.ID())

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	rootDir := t.TempDir()
	// t.Logf("node root dir: %s", rootDir)

	// memory block store
	bs := memstore.NewMemBS()
	// one block at height 1 with 2 txns
	blk1, appHash1 := createTestBlock(1, 2)
	bs.Store(blk1, appHash1)

	// dummy CE
	ce := &dummyCE{}

	t.Cleanup(func() {
		cancel()
		wg.Wait()
		mn.Close()
	})

	privKeys, genCfg := newGenesis(t, [][]byte{pk1})

	defaultConfigSet := config.DefaultConfig()
	defaultConfigSet.Consensus.ProposeTimeout = 5 * time.Minute

	// log1 := log.New(log.WithName("NODE1"), log.WithWriter(os.Stdout), log.WithLevel(log.LevelDebug), log.WithFormat(log.FormatUnstructured))
	cfg1 := &Config{
		RootDir:   rootDir,
		PrivKey:   privKeys[0],
		Logger:    log.DiscardLogger, // log1,
		Genesis:   *genCfg,
		Consensus: defaultConfigSet.Consensus,
		P2P:       defaultConfigSet.P2P, // mostly ignored as we are using WithHost below
	}
	node1, err := NewNode(cfg1, WithHost(h1), WithBlockStore(bs),
		WithConsensusEngine(ce))
	if err != nil {
		t.Fatalf("Failed to create Node 1: %v", err)
	}

	// now the test host.Host
	_, h2, err := newTestHost(t, mn)
	if err != nil {
		t.Fatalf("Failed to add peer to mocknet: %v", err)
	}
	setHupStreamHandlers(t, h2)
	// t.Logf("test host is %v", h2.ID())

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer os.RemoveAll(node1.Dir())
		node1.Start(ctx)
	}()

	time.Sleep(100 * time.Millisecond)

	// Link and connect the hosts
	if err := mn.LinkAll(); err != nil {
		t.Fatalf("Failed to link hosts: %v", err)
	}
	if err := mn.ConnectAllButSelf(); err != nil {
		t.Fatalf("Failed to connect hosts: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	testCases := []struct {
		name string
		fn   func(t *testing.T)
	}{
		{
			name: "request unknown hash manually",
			fn: func(t *testing.T) {
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
			},
		},
		{
			name: "request by hash using requestFrom, unknown block",
			fn: func(t *testing.T) {
				// t.Parallel()
				unknownHash := types.Hash{1}
				req, _ := blockHashReq{unknownHash}.MarshalBinary() // knownHash[:]
				_, err := requestFrom(ctx, h2, h1.ID(), req, ProtocolIDBlock, 1e4)
				if err == nil {
					t.Errorf("expected error but got none")
				} else if !errors.Is(err, ErrNotFound) {
					t.Errorf("unexpected error: %v", err)
				}
			},
		},
		{
			name: "request by hash manually, known",
			fn: func(t *testing.T) {
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
			},
		},
		{
			name: "request by hash using requestFrom, known block",
			fn: func(t *testing.T) {
				// t.Parallel()
				knownHash := blk1.Hash()
				req, _ := blockHashReq{knownHash}.MarshalBinary() // knownHash[:]
				resp, err := requestFrom(ctx, h2, h1.ID(), req, ProtocolIDBlock, 1e4)
				if err != nil {
					t.Errorf("ReadAll: %v", err)
				} else if bytes.Equal(resp, noData) {
					t.Error("expected data, got", resp)
				}
			},
		},
		{
			name: "request by height manually, unknown",
			fn: func(t *testing.T) {
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
			},
		},
		{
			name: "request by height using requestFrom, unknown",
			fn: func(t *testing.T) {
				// t.Parallel()
				var height int64
				req, _ := blockHeightReq{height}.MarshalBinary()
				_, err := requestFrom(ctx, h2, h1.ID(), req, ProtocolIDBlockHeight, 1e4)
				if err == nil {
					t.Errorf("expected error but got none")
				} else if !errors.Is(err, ErrNotFound) {
					t.Errorf("unexpected error: %v", err)
				}
			},
		},
		{
			name: "request by height manually, known",
			fn: func(t *testing.T) {
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
			},
		},
		{
			name: "request by height using requestFrom, known",
			fn: func(t *testing.T) {
				// t.Parallel()
				var height int64 = 1
				req, _ := blockHeightReq{height}.MarshalBinary()
				resp, err := requestFrom(ctx, h2, h1.ID(), req, ProtocolIDBlockHeight, 1e4)
				if err != nil {
					t.Errorf("ReadAll: %v", err)
				} else if bytes.Equal(resp, noData) {
					t.Error("expected data, got", resp)
				}
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, tt.fn)
	}
}
