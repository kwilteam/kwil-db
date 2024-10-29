package node

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"p2p/node/types"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
)

func (n *Node) txAnnStreamHandler(s network.Stream) {
	defer s.Close()

	s.SetDeadline(time.Now().Add(txGetTimeout))

	var ann txHashAnn
	if _, err := ann.ReadFrom(s); err != nil {
		fmt.Println("bad get tx req:", err)
		return
	}

	txHash := ann.Hash

	if !n.mp.PreFetch(txHash) { // it's in mempool
		return
	}

	var fetched bool
	defer func() {
		if !fetched { // release prefetch
			n.mp.Store(txHash, nil)
		}
	}()

	// not in mempool, check tx index
	if n.bki.HaveTx(txHash) {
		return // we have or are currently fetching it, do nothing, assuming we have already re-announced
	}
	// we don't have it. time to fetch

	// t0 := time.Now(); log.Printf("retrieving new tx: %q", txid)

	// First try to get from this stream.
	rawTx, err := requestTx(s, []byte(getMsg))
	if err != nil {
		n.log.Infof("announcer failed to provide %v, trying other peers", txHash)
		// Since we are aware, ask other peers. we could also put this in a goroutine
		s.Close() // close the announcers stream first
		rawTx, err = n.getTxWithRetry(context.TODO(), txHash, 500*time.Millisecond, 10)
		if err != nil {
			n.log.Infof("unable to retrieve tx %v: %v", txHash, err)
			return
		}
	}

	// n.log.Infof("obtained content for tx %q in %v", txid, time.Since(t0))

	// here we could check tx index again in case a block was mined with it
	// while we were fetching it

	// store in mempool since it was not in tx index and thus not confirmed
	n.mp.Store(txHash, rawTx)
	fetched = true

	// re-announce
	go n.announceTx(context.Background(), txHash, rawTx, s.Conn().RemotePeer())
}

func (n *Node) announceTx(ctx context.Context, txHash types.Hash, rawTx []byte, from peer.ID) {
	peers := n.host.Network().Peers()
	if len(peers) == 0 {
		n.log.Warnf("no peers to advertise tx to")
		return
	}

	for _, peerID := range peers {
		if peerID == from {
			continue
		}
		// n.log.Infof("advertising tx %v (len %d) to peer %v", txid, len(rawTx), peerID)
		err := n.advertiseTxToPeer(ctx, peerID, txHash, rawTx)
		if err != nil {
			n.log.Warn("failed to advertise tx to peer", "peer", peerID, "error", err)
			continue
		}
	}
}

// advertiseTxToPeer sends a lightweight advertisement to a connected peer.
// The stream remains open in case the peer wants to request the content right.
func (n *Node) advertiseTxToPeer(ctx context.Context, peerID peer.ID, txHash types.Hash, rawTx []byte) error {
	s, err := n.host.NewStream(ctx, peerID, ProtocolIDTxAnn)
	if err != nil {
		return fmt.Errorf("failed to open stream to peer: %w", err)
	}

	roundTripDeadline := time.Now().Add(txGetTimeout) // lower for this part?
	s.SetWriteDeadline(roundTripDeadline)

	// Send a lightweight advertisement with the object ID
	_, err = newTxHashAnn(txHash).WriteTo(s)
	if err != nil {
		return fmt.Errorf("txann failed: %w", err)
	}

	// n.log.Infof("advertised tx content %s to peer %s", txid, peerID)

	// Keep the stream open for potential content requests
	go func() {
		defer s.Close()

		s.SetReadDeadline(time.Now().Add(txGetTimeout))

		req := make([]byte, 128)
		nr, err := s.Read(req)
		if err != nil && !errors.Is(err, io.EOF) {
			n.log.Warn("bad get tx req", "error", err)
			return
		}
		if nr == 0 /*&& errors.Is(err, io.EOF)*/ {
			return // they hung up, probably didn't want it
		}
		req = req[:nr]
		req, ok := bytes.CutPrefix(req, []byte(getMsg))
		if !ok {
			n.log.Warnf("advertise wait: bad get tx request %q", string(req))
			return
		}

		s.SetWriteDeadline(time.Now().Add(20 * time.Second))
		s.Write(rawTx)
	}()

	return nil
}

// startTxAnns creates pretend transactions, adds them to the tx index, and
// announces them to peers.
func (n *Node) startTxAnns(ctx context.Context, newPeriod, reannouncePeriod time.Duration, sz int) {
	n.wg.Add(1)
	go func() {
		defer n.wg.Done()

		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(newPeriod):
			}

			txHash := types.Hash(randBytes(32))
			rawTx := randBytes(sz)
			n.mp.Store(txHash, rawTx)

			// n.log.Infof("announcing txid %v", txid)
			n.announceTx(ctx, types.Hash(txHash), rawTx, n.host.ID())
		}
	}()

	n.wg.Add(1)
	go func() {
		defer n.wg.Done()

		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(reannouncePeriod):
			}

			func() {
				ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
				defer cancel()

				const sendN = 20
				txns := n.mp.PeekN(sendN)
				n.log.Infof("re-announcing %d unconfirmed txns", len(txns))

				for _, nt := range txns {
					n.announceTx(ctx, nt.Hash, nt.Tx, n.host.ID()) // response handling is async
					if ctx.Err() != nil {
						n.log.Warn("interrupting long re-broadcast")
						break
					}
				}
			}()
		}
	}()
}
