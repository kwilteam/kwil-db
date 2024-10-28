package node

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"time"

	"p2p/node/types"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
)

var (
	ErrNotFound    = errors.New("resource not available")
	ErrTxNotFound  = errors.New("tx not available")
	ErrBlkNotFound = errors.New("block not available")
	ErrNoResponse  = errors.New("stream closed without response")
)

const (
	txReadLimit  = 30_000_000
	txGetTimeout = 20 * time.Second
)

func readTxResp(rd io.Reader) ([]byte, error) {
	rd = io.LimitReader(rd, txReadLimit)
	resp, err := io.ReadAll(rd)
	if err != nil {
		return nil, err
	}
	if len(resp) == 0 {
		return nil, ErrNoResponse
	}
	if bytes.Equal(resp, noData) {
		return nil, ErrTxNotFound
	}
	return resp, nil
}

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

func requestTx(rw io.ReadWriter, reqMsg []byte) ([]byte, error) {
	content, err := request(rw, reqMsg, txReadLimit)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, ErrTxNotFound
		}
		return nil, fmt.Errorf("tx get request failed: %v", err)
	}
	return content, nil
}

func (n *Node) getTx(ctx context.Context, txHash types.Hash) ([]byte, error) {
	for _, peer := range n.peers() {
		log.Printf("requesting tx %v from %v", txHash, peer)
		raw, err := getTx(ctx, txHash, peer, n.host)
		if errors.Is(err, ErrTxNotFound) {
			log.Printf("transaction not available on %v", peer)
			continue
		}
		if errors.Is(err, ErrNoResponse) {
			log.Printf("no response to tx request to %v", peer)
			continue
		}
		if err != nil {
			log.Printf("unexpected error from %v: %v", peer, err)
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
		log.Printf("unable to retrieve tx %v (%v), waiting to retry", txHash, err)
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
