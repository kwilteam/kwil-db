package node

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	ktypes "github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/node/peers"
	"github.com/kwilteam/kwil-db/node/types"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
)

func (n *Node) txAnnStreamHandler(s network.Stream) {
	defer s.Close()

	s.SetDeadline(time.Now().Add(txGetTimeout))

	var ann txHashAnn
	if _, err := ann.ReadFrom(s); err != nil {
		n.log.Warnf("bad tx ann: %v", err)
		return
	}

	txHash := ann.Hash

	ok, done := n.mp.PreFetch(txHash)
	if !ok { // it's in mempool or being fetched already
		return
	}
	defer done()

	// not in mempool, check tx index
	if n.bki.HaveTx(txHash) {
		return // we have or are currently fetching it, do nothing, assuming we have already re-announced
	}
	// we don't have it. time to fetch

	// t0 := time.Now(); log.Printf("retrieving new tx: %q", txid)

	// First try to get from this stream.
	rawTx, err := requestTx(s, []byte(getMsg))
	if err != nil {
		n.log.Warnf("announcer failed to provide %v due to error %v, trying other peers", txHash, err)
		// Since we are aware, ask other peers. we could also put this in a goroutine
		s.Close() // close the announcers stream first
		rawTx, err = n.getTxWithRetry(context.TODO(), txHash, 500*time.Millisecond, 10)
		if err != nil {
			n.log.Errorf("unable to retrieve tx %v: %v", txHash, err)
			return
		}
	}

	var tx ktypes.Transaction
	if err = tx.UnmarshalBinary(rawTx); err != nil {
		n.log.Errorf("invalid transaction received %v: %v", txHash, err)
		return
	}

	// Ensure the received transaction is the one requested by hash.
	ntx := types.NewTx(&tx) // the immutable tx for CE with Hash stored
	if txHash != ntx.Hash() {
		n.log.Errorf("tx hash mismatch: %v != %v", txHash, ntx.Hash())
		return
	}

	// n.log.Infof("obtained content for tx %q in %v", txid, time.Since(t0))

	// here we could check tx index again in case a block was mined with it
	// while we were fetching it

	ctx := context.Background()
	if err := n.ce.QueueTx(ctx, ntx); err != nil {
		n.log.Warnf("tx %v failed check: %v", txHash, err)
		return
	}

	// re-announce
	n.queueTxn(txHash, rawTx, s.Conn().RemotePeer())
}

func (n *Node) queueTxn(txID types.Hash, rawTx []byte, from peer.ID) {
	tx := orderedTxn{txID: txID, rawtx: rawTx, from: from}

	select {
	case n.txQueue <- tx:
	default:
		n.log.Warnf("tx queue full, dropping tx %v", txID)
	}
}

func (n *Node) announceRawTx(ctx context.Context, txHash types.Hash, rawTx []byte, from peer.ID) {
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

func (n *Node) announceTx(_ context.Context, _ *ktypes.Transaction, txID types.Hash, from peer.ID) {
	// Storing nil for the raw transaction as it is already in the mempool and can be retrieved
	// when dequeued. This helps in reducing the memory footprint of the queue.
	n.queueTxn(txID, nil, from)
}

// advertiseTxToPeer sends a lightweight advertisement to a connected peer.
// The stream remains open in case the peer wants to request the content right.
func (n *Node) advertiseTxToPeer(ctx context.Context, peerID peer.ID, txHash types.Hash, rawTx []byte) error {
	s, err := n.host.NewStream(ctx, peerID, ProtocolIDTxAnn)
	if err != nil {
		return fmt.Errorf("failed to open stream to peer: %w", peers.CompressDialError(err))
	}

	roundTripDeadline := time.Now().Add(txAnnTimeout)
	s.SetWriteDeadline(roundTripDeadline)

	// Send a lightweight advertisement with the object ID
	_, err = newTxHashAnn(txHash).WriteTo(s)
	if err != nil {
		return fmt.Errorf("txann failed to peer %s: %w", peerID, err)
	}

	mets.Advertised(ctx, string(ProtocolIDTxAnn))

	// n.log.Infof("advertised tx content %s to peer %s", txid, peerID)

	// Keep the stream open for potential content requests
	go func() {
		defer s.Close()

		s.SetReadDeadline(time.Now().Add(txAnnRespTimeout))

		req := make([]byte, len(getMsg))
		nr, err := s.Read(req)
		if err != nil && !errors.Is(err, io.EOF) {
			n.log.Warn("bad get tx req", "error", err)
			return
		}
		if nr == 0 /*&& errors.Is(err, io.EOF)*/ {
			mets.AdvertiseRejected(ctx, string(ProtocolIDTxAnn))
			return // they hung up, probably didn't want it
		}
		if getMsg != string(req) {
			n.log.Warnf("advertise wait: bad get tx request %q", string(req))
			return
		}

		s.SetWriteDeadline(time.Now().Add(txGetTimeout))
		s.Write(rawTx)

		mets.AdvertiseServed(ctx, string(ProtocolIDTxAnn), int64(len(rawTx)))
	}()

	return nil
}

// startTxAnns handles periodic reannouncement. It can also be modified to
// regularly create dummy transactions.
func (n *Node) startTxAnns(ctx context.Context, reannouncePeriod time.Duration) {
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
				const sendBtsLimit = 8_000_000
				txns := n.mp.PeekN(sendN, sendBtsLimit)
				if len(txns) == 0 {
					return // from this anon func, not the goroutine!
				}
				n.log.Infof("re-announcing %d unconfirmed txns", len(txns))

				var numSent, bytesSent int64
				for _, tx := range txns {
					rawTx := tx.Bytes()
					n.announceRawTx(ctx, tx.Hash(), rawTx, n.host.ID()) // response handling is async
					if ctx.Err() != nil {
						n.log.Warn("interrupting long re-broadcast")
						break
					}
					numSent++
					bytesSent += int64(len(rawTx))
				}

				mets.TxnsReannounced(ctx, numSent, bytesSent)
			}()
		}
	}()
}

// startOrderedTxQueueAnns ensures that transaction announcements are broadcasted
// in the order they are received, maintaining FIFO order for nonce consistency.
func (n *Node) startOrderedTxQueueAnns(ctx context.Context) {
	n.wg.Add(1)
	go func() {
		defer n.wg.Done()

		for {
			select {
			case <-ctx.Done():
				return
			case txn := <-n.txQueue:
				rawTx := txn.rawtx
				if txn.rawtx == nil {
					// fetch the raw tx from the mempool
					tx := n.mp.Get(txn.txID)
					if tx == nil {
						continue // tx was removed from mempool
					}
					rawTx = tx.Bytes()
				}

				n.announceRawTx(ctx, txn.txID, rawTx, txn.from)
			}
		}
	}()
}
