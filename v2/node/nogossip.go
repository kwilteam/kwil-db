package node

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"p2p/node/types"
	"time"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
)

func (n *Node) txAnnStreamHandler(s network.Stream) {
	defer s.Close()

	s.SetDeadline(time.Now().Add(txGetTimeout))

	req := make([]byte, 128)
	nr, err := s.Read(req)
	if err != nil && err != io.EOF {
		log.Println("bad get tx req:", err)
		return
	}
	req, ok := bytes.CutPrefix(req[:nr], []byte(annTxMsgPrefix))
	if !ok {
		log.Println("txAnnStreamHandler: bad get tx request", string(req))
		return
	}
	txid := string(req)
	// log.Printf("tx announcement received: %q", txid)
	txHash, err := types.NewHashFromString(txid)
	if err != nil {
		fmt.Println("bad txid:", err)
		return
	}

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
	if n.txi.Get(txHash) != nil {
		return // we have or are currently fetching it, do nothing, assuming we have already re-announced
	}
	// we don't have it. time to fetch

	t0 := time.Now()
	// log.Printf("retrieving new tx: %q", txid)

	// First try to get from this stream.
	rawTx, err := requestTx(s, []byte(getMsg))
	if err != nil {
		log.Printf("announcer failed to provide %v, trying other peers", txid)
		// Since we are aware, ask other peers. we could also put this in a goroutine
		s.Close() // close the announcers stream first
		rawTx, err = n.getTxWithRetry(context.TODO(), txid, 500*time.Millisecond, 10)
		if err != nil {
			log.Printf("unable to retrieve tx %v: %v", txid, err)
			return
		}
	}

	log.Printf("obtained content for tx %q in %v", txid, time.Since(t0))

	// here we could check tx index again in case a block was mined with it
	// while we were fetching it

	// store in mempool since it was not in tx index and thus not confirmed
	n.mp.Store(txHash, rawTx)
	fetched = true

	// re-announce
	go n.announceTx(context.Background(), txid, rawTx, s.Conn().RemotePeer())
}

func (n *Node) announceTx(ctx context.Context, txid string, rawTx []byte, from peer.ID) {
	peers := n.host.Network().Peers()
	if len(peers) == 0 {
		log.Println("no peers to advertise tx to")
		return
	}

	for _, peerID := range peers {
		if peerID == from {
			continue
		}
		// log.Printf("advertising tx %v (len %d) to peer %v", txid, len(rawTx), peerID)
		err := n.advertiseTxToPeer(ctx, peerID, txid, rawTx)
		if err != nil {
			log.Println(err)
			continue
		}
	}
}

// advertiseTxToPeer sends a lightweight advertisement to a connected peer.
// The stream remains open in case the peer wants to request the content right.
func (n *Node) advertiseTxToPeer(ctx context.Context, peerID peer.ID, txid string, rawTx []byte) error {
	s, err := n.host.NewStream(ctx, peerID, ProtocolIDTxAnn)
	if err != nil {
		return fmt.Errorf("failed to open stream to peer: %w", err)
	}

	roundTripDeadline := time.Now().Add(txGetTimeout) // lower for this part?
	s.SetWriteDeadline(roundTripDeadline)

	// Send a lightweight advertisement with the object ID
	_, err = s.Write([]byte(annTxMsgPrefix + txid))
	if err != nil {
		return fmt.Errorf("txann failed: %w", err)
	}

	// log.Printf("advertised tx content %s to peer %s", txid, peerID)

	// Keep the stream open for potential content requests
	go func() {
		defer s.Close()

		s.SetReadDeadline(time.Now().Add(txGetTimeout))

		req := make([]byte, 128)
		nr, err := s.Read(req)
		if err != nil && !errors.Is(err, io.EOF) {
			log.Println("bad get tx req:", err)
			return
		}
		if nr == 0 /*&& errors.Is(err, io.EOF)*/ {
			return // they hung up, probably didn't want it
		}
		req = req[:nr]
		req, ok := bytes.CutPrefix(req, []byte(getMsg))
		if !ok {
			log.Printf("advertise wait: bad get tx request %q", string(req))
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

			txHash := randBytes(32)
			txid := hex.EncodeToString(txHash)
			rawTx := randBytes(sz)
			n.mp.Store(types.Hash(txHash), rawTx)

			// log.Printf("announcing txid %v", txid)
			n.announceTx(ctx, txid, rawTx, n.host.ID())
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
				log.Printf("re-announcing up to %d unconfirmed txns", sendN)
				for nt := range n.mp.FeedN(sendN) {
					n.announceTx(ctx, nt.ID.String(), nt.Tx, n.host.ID()) // response handling is async
					if ctx.Err() != nil {
						log.Println("interrupting long re-broadcast")
						break
					}
				}
			}()
		}
	}()
}
