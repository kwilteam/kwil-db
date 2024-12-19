package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/kwilteam/kwil-db/core/types"
	clientType "github.com/kwilteam/kwil-db/core/types/client"
	"github.com/kwilteam/kwil-db/core/types/transactions"
	"github.com/kwilteam/kwil-db/core/utils"
	"github.com/kwilteam/kwil-db/core/utils/random"
	"go.uber.org/zap"
)

// The harness methods in this file pertain to the embedded dataset schema,
// testScheme.

type asyncResp struct {
	err        error
	res        *transactions.TransactionResult
	expectFail bool
}

func (ar *asyncResp) error() error {
	if ar.err != nil {
		return ar.err
	}
	if ar.res.Code != 0 {
		return fmt.Errorf("execution failed with code %d, log: %q",
			ar.res.Code, ar.res.Log)
	}
	return nil
}
func (ar *asyncResp) Error() error {
	err := ar.error()
	if err != nil {
		if ar.expectFail {
			err = errors.Join(err, ErrExpected)
		}
		return err
	}
	if ar.expectFail {
		return errors.New("UNEXPECTEDLY succeeded when it should have failed")
	}
	return nil
}

func (h *harness) dropDB(ctx context.Context, dbid string) error {
	var txHash transactions.TxHash
	err := h.underNonceLock(ctx, func(nonce int64) error {
		var err error
		txHash, err = h.DropDatabaseID(ctx, dbid, clientType.WithNonce(nonce))
		return err
	})
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	txResp, err := h.WaitTx(ctx, txHash, txPollInterval)
	if err != nil {
		err = errors.Join(err, h.recoverNonce(ctx))
		return fmt.Errorf("WaitTx (drop): %w", err)
	}
	if code := txResp.TxResult.Code; code != 0 {
		return fmt.Errorf("drop tx failed (%d): %v", code, txResp.TxResult.Log)
	}
	return nil
}

func (h *harness) deployDBAsync(ctx context.Context, schema *types.Schema) (string, <-chan asyncResp, error) {
	schema.Name = random.String(12)

	var txHash transactions.TxHash
	err := h.underNonceLock(ctx, func(nonce int64) error {
		var err error
		txHash, err = h.DeployDatabase(ctx, schema, clientType.WithNonce(nonce))
		return err
	})
	if err != nil {
		return "", nil, err
	}

	dbid := utils.GenerateDBID(schema.Name, h.signer.Identity())
	// fmt.Println("deployDBAsync", dbid)
	promise := make(chan asyncResp, 1)
	go func() {
		ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
		resp, err := h.WaitTx(ctx, txHash, txPollInterval)
		if err != nil {
			h.logger.Error("WaitTx", zap.Error(err))
			err = errors.Join(err, h.recoverNonce(ctx))
			promise <- asyncResp{err: err}
			return
		}
		promise <- asyncResp{res: &resp.TxResult}
		h.logger.Info(fmt.Sprintf("database %q deployed in block %d", dbid, resp.Height))
	}()

	return dbid, promise, nil
}

func (h *harness) deployDB(ctx context.Context, schema *types.Schema) (string, error) {
	dbid, promise, err := h.deployDBAsync(ctx, schema)
	if err != nil {
		return "", err
	}
	res := <-promise
	if res.err != nil {
		return "", res.err
	}
	txRes := res.res
	if code := txRes.Code; code != 0 {
		return "", fmt.Errorf("failed to deploy database, code = %d, log = %q", code, txRes.Log)
	}
	return dbid, nil
}
