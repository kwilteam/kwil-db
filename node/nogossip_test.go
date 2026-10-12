package node

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/stretchr/testify/require"

	"github.com/trufnetwork/kwil-db/core/crypto"
	ktypes "github.com/trufnetwork/kwil-db/core/types"
	"github.com/trufnetwork/kwil-db/node/peers"
	"github.com/trufnetwork/kwil-db/node/types"
)

// joinChain makes h advertise the chain's protocol, so that a node on that
// chain keeps the connection after identifying h. Call it before connecting
// h. It waits briefly so that h's identify service lists the new protocol, as
// makeTestHosts does after adding its handlers.
func joinChain(h host.Host, chainID string) {
	h.SetStreamHandler(peers.ProtocolIDPrefixChainID+protocol.ID(chainID), func(s network.Stream) {
		s.Close()
	})
	time.Sleep(100 * time.Millisecond)
}

// announceTx plays the announcing peer of the tx announcement protocol: it
// sends the hash, waits for the get request, sends the transaction and waits
// for the receiving node to close the stream when it is done with it.
func announceTx(t *testing.T, from host.Host, to peer.ID, tx *ktypes.Transaction) {
	t.Helper()

	s, err := from.NewStream(context.Background(), to, ProtocolIDTxAnn)
	require.NoError(t, err)
	defer s.Close()

	_, err = newTxHashAnn(tx.Hash()).WriteTo(s)
	require.NoError(t, err)

	req := make([]byte, len(getMsg))
	_, err = io.ReadFull(s, req)
	require.NoError(t, err)
	require.Equal(t, getMsg, string(req))

	_, err = s.Write(tx.Bytes())
	require.NoError(t, err)
	require.NoError(t, s.CloseWrite())

	closed := make(chan error, 1)
	go func() {
		_, err := io.ReadAll(s)
		closed <- err
	}()
	select {
	case err := <-closed:
		require.NoError(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("the node did not close the announcement stream")
	}
}

// TestTxAnnStreamHandlerSurvivesPanickingQueue checks that a panic while
// queuing an announced transaction ends only that stream, not the process,
// and that the node goes on handling announcements.
func TestTxAnnStreamHandlerSurvivesPanickingQueue(t *testing.T) {
	nodes, extraHosts, _, mn := makeTestHosts(t, 1, 1, 5*time.Hour, crypto.KeyTypeSecp256k1)
	joinChain(extraHosts[0], nodes[0].chainID)
	linkAll(t, mn)

	n1 := nodes[0]
	queued := make(chan types.Hash, 2)
	n1.ce.(*dummyCE).queueTxHandler = func(_ context.Context, tx *types.Tx) error {
		queued <- tx.Hash()
		if tx.Body == nil {
			panic("no body")
		}
		return nil
	}

	startNodes(t, nodes)
	time.Sleep(100 * time.Millisecond)

	announcer := extraHosts[0]
	bodyless := &ktypes.Transaction{}
	complete := &ktypes.Transaction{
		Body: &ktypes.TransactionBody{PayloadType: ktypes.PayloadTypeExecute, Nonce: 1},
	}

	for _, tx := range []*ktypes.Transaction{bodyless, complete} {
		announceTx(t, announcer, n1.host.ID(), tx)
		select {
		case got := <-queued:
			require.Equal(t, tx.Hash(), got)
		case <-time.After(5 * time.Second):
			t.Fatalf("announced transaction %s was not queued", tx.Hash())
		}
	}
}
