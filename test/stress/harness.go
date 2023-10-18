package main

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"sync"

	"github.com/kwilteam/kwil-db/core/client"
	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/types/transactions"
)

// harness is a Client driver designed around an embedded dataset Kuniform
// schema. It has methods to execute actions asynchronously, while track it's
// used nonces so it can correctly make multiple unconfirmed transactions with
// increasing nonces. See underNonceLock and recoverNonce.
type harness struct {
	*client.Client
	logger *log.Logger
	pub    []byte

	nonceMtx sync.Mutex
	nonce    int64 // atomic.Int64

	concurrentBroadcast bool // be wreckless with nonces

	nestedLogger *log.Logger
}

func (h *harness) underNonceLock(ctx context.Context, fn func(int64) error) error {
	if h.concurrentBroadcast {
		// Grab the next nonce in a thread-safe manner, but do not wait for
		// broadcast to complete to release the lock. If there is a nonce error,
		// there will be more chaos with concurrent broadcasting.
		h.nonceMtx.Lock()
		h.nonce++
		nonce := h.nonce // chaos: + int64(rand.Intn(2))
		h.nonceMtx.Unlock()
		if err := fn(nonce); err != nil {
			if errors.Is(err, transactions.ErrInvalidNonce) {
				// Note: several goroutines may all try to do this if they all hit the nonce error
				h.recoverNonce(ctx)
				h.printf("error, nonce reverting to %d\n", h.nonce)
			}
			return err
		}
		return nil
	}

	h.nonceMtx.Lock()
	defer h.nonceMtx.Unlock()
	h.nonce++
	h.printf("using next nonce %d", h.nonce)

	if err := fn(h.nonce); err != nil {
		if errors.Is(err, transactions.ErrInvalidNonce) { // this alone should not happen
			// NOTE: if GetAccount returns only the confirmed nonce, we'll error
			// again shortly if there are others already in mempool.
			acct, err := h.GetAccount(ctx, h.pub, types.AccountStatusPending)
			if err != nil {
				return err
			}
			h.nonce = acct.Nonce
			h.printf("RESET NONCE TO LATEST REPORTED: %d", h.nonce)
		}
		return err
	}
	return nil
}

func (h *harness) recoverNonce(ctx context.Context) error {
	h.nonceMtx.Lock()
	defer h.nonceMtx.Unlock()

	acct, err := h.GetAccount(ctx, h.pub, types.AccountStatusPending)
	if err != nil {
		return err
	}
	h.nonce = acct.Nonce
	h.printf("RESET NONCE TO LATEST CONFIRMED: %d", h.nonce)
	return nil
}

func (h *harness) printf(msg string, args ...any) {
	h.nestedLogger.Info(fmt.Sprintf(msg, args...))
}

func (h *harness) printRecs(ctx context.Context, recs *client.Records) {
	for recs.Next() {
		if ctx.Err() != nil {
			return
		}
		h.logger.Info(fmt.Sprintf("%#v\n", recs.Record().String()))
	}
}

func (h *harness) executeActionAsync(ctx context.Context, dbid string, action string,
	inputs [][]any) (transactions.TxHash, error) {
	var txHash transactions.TxHash
	err := h.underNonceLock(ctx, func(nonce int64) error {
		var err error
		txHash, err = h.ExecuteAction(ctx, dbid, action, inputs,
			client.WithNonce(nonce), client.WithFee(&big.Int{}))
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", action, err)
	}
	return txHash, nil
}

func (h *harness) executeAction(ctx context.Context, dbid string, action string,
	inputs [][]any) error {
	txHash, err := h.executeActionAsync(ctx, dbid, action, inputs)
	if err != nil {
		return err
	}

	// !!!!! if we got a txhash back from the node, mempool should have it
	// time.Sleep(time.Millisecond * 200)
	// !!!!! but it doesn't so WaitTx will spew warnings at first

	txResp, err := h.WaitTx(ctx, txHash, txPollInterval)
	if err != nil {
		err = errors.Join(err, h.recoverNonce(ctx))
		return fmt.Errorf("WaitTx (%v): %w", action, err)
	}
	if code := txResp.TxResult.Code; code != 0 {
		return fmt.Errorf("%s tx failed (%d): %v", action, code, txResp.TxResult.Log)
	}
	return nil
}
