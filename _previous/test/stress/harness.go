package main

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sync"

	"github.com/kwilteam/kwil-db/core/crypto/auth"
	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/core/types"
	clientType "github.com/kwilteam/kwil-db/core/types/client"
	"github.com/kwilteam/kwil-db/core/types/transactions"
)

// harness is a Client driver designed around an embedded dataset Kuneiform
// schema. It has methods to execute actions asynchronously, while track it's
// used nonces so it can correctly make multiple unconfirmed transactions with
// increasing nonces. See underNonceLock and recoverNonce.
type harness struct {
	clientType.Client
	logger *log.Logger
	acctID []byte

	signer auth.Signer

	nonceMtx sync.Mutex
	nonce    int64 // atomic.Int64

	concurrentBroadcast bool // broadcast many before confirm, still coordinate nonces
	nonceChaos          int  // apply random nonce-jitter every 1/n times

	nestedLogger *log.Logger

	quiet bool

	// asc actSchemaClient
}

// for about 1 in every f times, produce a non-zero nonce jitter:
// {-2, 1, 1, 2, 3, 4}
func randNonceJitter(f int) int64 {
	if f == 0 {
		return 0
	}
	if rand.Intn(f) > 0 {
		return 0
	}
	n := rand.Int63n(6) - 2
	if n >= 0 { // 0-3 => 1-4
		return n + 1
	}
	return n
}

func (h *harness) underNonceLock(ctx context.Context, fn func(int64) error) error {
	if h.concurrentBroadcast {
		// Grab the next nonce in a thread-safe manner, but do not wait for
		// broadcast to complete to release the lock. If there is a nonce error,
		// there will be more chaos with concurrent broadcasting.
		h.nonceMtx.Lock()
		h.nonce++
		nonce := h.nonce + randNonceJitter(h.nonceChaos)
		h.nonceMtx.Unlock()
		if err := fn(nonce); err != nil {
			if errors.Is(err, transactions.ErrInvalidNonce) {
				// Note: several goroutines may all try to do this if they all hit the nonce error
				h.recoverNonce(ctx)
				h.printf("error, nonce %d was wrong, reverting to %d\n", nonce, h.nonce)
			}
			return err
		}
		return nil
	}

	// -sb
	h.nonceMtx.Lock()
	defer h.nonceMtx.Unlock()
	h.nonce++
	h.printf("using next nonce %d", h.nonce)

	if err := fn(h.nonce); err != nil {
		if errors.Is(err, transactions.ErrInvalidNonce) { // this alone should not happen
			// NOTE: if GetAccount returns only the confirmed nonce, we'll error
			// again shortly if there are others already in mempool.
			acct, err := h.GetAccount(ctx, h.acctID, types.AccountStatusPending)
			if err != nil {
				return err
			}
			h.nonce = acct.Nonce
			h.printf("RESET NONCE TO LATEST REPORTED (underNonceLock): %d", h.nonce)
		}
		return err
	}
	return nil
}

func (h *harness) recoverNonce(ctx context.Context) error {
	h.nonceMtx.Lock()
	defer h.nonceMtx.Unlock()

	acct, err := h.GetAccount(ctx, h.acctID, types.AccountStatusPending)
	if err != nil {
		return err
	}
	h.nonce = acct.Nonce
	h.printf("RESET NONCE TO LATEST CONFIRMED (recoverNonce): %d", h.nonce)
	return nil
}

func (h *harness) printf(msg string, args ...any) {
	var hasErr bool
	for _, arg := range args {
		if err, isErr := arg.(error); isErr && !errors.Is(err, ErrExpected) {
			hasErr = true
			break
		}
	}

	fun := h.nestedLogger.Info
	if hasErr {
		fun = h.nestedLogger.Error
	} else if h.quiet {
		return
	}
	fun(fmt.Sprintf(msg, args...))
}

func (h *harness) printRecs(ctx context.Context, recs *clientType.Records) {
	for recs.Next() {
		if ctx.Err() != nil {
			return
		}
		h.logger.Info(fmt.Sprintf("%#v\n", recs.Record().String()))
	}
}

func (h *harness) executeAsync(ctx context.Context, dbid, action string,
	inputs [][]any) (transactions.TxHash, error) {
	var txHash transactions.TxHash
	err := h.underNonceLock(ctx, func(nonce int64) error {
		var err error
		txHash, err = h.Execute(ctx, dbid, action, inputs,
			clientType.WithNonce(nonce) /*, clientType.WithFee(&big.Int{})*/) // TODO: badFee mode
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", action, err)
	}
	return txHash, nil
}

func (h *harness) execute(ctx context.Context, dbid string, action string,
	inputs [][]any) error {
	txHash, err := h.executeAsync(ctx, dbid, action, inputs)
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
