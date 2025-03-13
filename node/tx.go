package node

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/kwilteam/kwil-db/node/types"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
)

var (
	ErrNotFound        = errors.New("resource not available")
	ErrTxNotFound      = types.ErrTxNotFound
	ErrTxAlreadyExists = types.ErrTxAlreadyExists
	ErrBlkNotFound     = types.ErrBlkNotFound
	ErrNoResponse      = types.ErrNoResponse
)

const (
	txReadLimit      = 30_000_000
	txAnnTimeout     = 5 * time.Second // time to Write tx ann to peer
	txAnnRespTimeout = txAnnTimeout    // time to wait for get response or a hangup
	txGetTimeout     = 20 * time.Second
)

func getTx(ctx context.Context, txHash types.Hash, peer peer.ID, host host.Host) ([]byte, error) {
	resID, _ := newTxHashReq(txHash).MarshalBinary()
	rawTx, err := requestFrom(ctx, host, peer, resID, ProtocolIDTx, txReadLimit)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, errors.Join(err, ErrTxNotFound)
		}
		return nil, fmt.Errorf("tx get request failed: %v", err)
	}
	return rawTx, nil
}

func requestCurrentResource(s network.Stream, readLimit int64) ([]byte, error) {
	// ask to send it
	_, err := s.Write([]byte(getMsg))
	if err != nil {
		return nil, fmt.Errorf("resource get request failed: %w", err)
	}

	// read the uint64 length
	var lenBuf [8]byte
	_, err = io.ReadFull(s, lenBuf[:])
	if err != nil {
		return nil, fmt.Errorf("failed to read resource length response: %w", err)
	}
	resLen := int64(binary.BigEndian.Uint64(lenBuf[:]))
	if resLen == 0 {
		return nil, ErrTxNotFound
	}
	if resLen > readLimit {
		return nil, fmt.Errorf("tx too large: %d", resLen)
	}

	// read the contents
	resource := make([]byte, resLen)
	if n, err := io.ReadFull(s, resource); err != nil {
		if n == 0 {
			return nil, ErrNoResponse
		}
		return nil, fmt.Errorf("failed to read resource content response: %w", err)
	}

	// ack receipt
	s.Write([]byte(gotItMsg)) // courtesy "got it" message, doesn't affect our retrieval though

	return resource, nil
}

// func requestTx(rw io.ReadWriter, reqMsg []byte) ([]byte, error) {
// 	content, err := request(rw, reqMsg, txReadLimit)
// 	if err != nil {
// 		if errors.Is(err, ErrNotFound) {
// 			return nil, ErrTxNotFound
// 		}
// 		return nil, fmt.Errorf("tx get request failed: %v", err)
// 	}
// 	return content, nil
// }

func (n *Node) getTx(ctx context.Context, txHash types.Hash) ([]byte, error) {
	for _, peer := range n.peers() {
		n.log.Info("requesting tx", "hash", txHash, "peer", peer)
		raw, err := getTx(ctx, txHash, peer, n.host)
		if errors.Is(err, ErrTxNotFound) {
			n.log.Warnf("transaction not available on %v", peer)
			continue
		}
		if errors.Is(err, ErrNoResponse) {
			n.log.Warnf("no response to tx request to %v", peer)
			continue
		}
		if err != nil {
			n.log.Warnf("unexpected error from %v: %v", peer, err)
			continue
		}
		return raw, nil
	}
	return nil, ErrTxNotFound
}

func (n *Node) getTxWithRetry(ctx context.Context, txHash types.Hash, baseDelay time.Duration,
	maxAttempts int) ([]byte, error) {
	var attempts int
	for {
		raw, err := n.getTx(ctx, txHash)
		if err == nil {
			return raw, nil
		}
		n.log.Warnf("unable to retrieve tx %v (%v), waiting to retry", txHash, err)
		select {
		case <-ctx.Done():
		case <-time.After(baseDelay):
		}
		baseDelay *= 2
		attempts++
		if attempts >= maxAttempts {
			return nil, ErrTxNotFound
		}
	}
}

func (n *Node) txGetStreamHandler(s network.Stream) {
	defer s.Close()

	var req txHashReq
	if _, err := req.ReadFrom(s); err != nil {
		n.log.Warn("bad get tx req", "error", err)
		return
	}

	// first check mempool
	ntx := n.mp.Get(req.Hash)
	if ntx != nil {
		ntx.Transaction.WriteTo(s)
		return
	}

	// then confirmed tx index
	tx, _, _, _, err := n.bki.GetTx(req.Hash)
	if err != nil {
		if !errors.Is(err, types.ErrNotFound) {
			n.log.Errorf("unexpected GetTx error: %v", err)
		}
		s.Write(noData) // don't have it
	} else {
		tx.WriteTo(s)
	}

	// NOTE: response could also include conf/unconf or block height (-1 or N)
}
