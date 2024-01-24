package main

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/kwilteam/kwil-db/cmd/kwil-cli/cmds/common"
	"github.com/kwilteam/kwil-db/core/client"
	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/core/types/transactions"
)

// timedClient is a wrapper around a common.Client
type timedClient struct {
	common.Client
	showReqDur bool
	logger     *log.Logger
	//cl         common.Client
}

func (tc *timedClient) ChainID() string {
	if tc.showReqDur {
		defer tc.printDur(time.Now(), "ChainID")
	}
	return tc.Client.ChainID()
}

func (tc *timedClient) DeployDatabase(ctx context.Context, payload *transactions.Schema, opts ...client.TxOpt) (transactions.TxHash, error) {
	if tc.showReqDur {
		defer tc.printDur(time.Now(), "DeployDatabase")
	}
	return tc.Client.DeployDatabase(ctx, payload, opts...)
}

func (tc *timedClient) DropDatabase(ctx context.Context, name string, opts ...client.TxOpt) (transactions.TxHash, error) {
	if tc.showReqDur {
		defer tc.printDur(time.Now(), "DropDatabase")
	}
	return tc.Client.DropDatabase(ctx, name, opts...)
}

func (tc *timedClient) DropDatabaseID(ctx context.Context, dbid string, opts ...client.TxOpt) (transactions.TxHash, error) {
	if tc.showReqDur {
		defer tc.printDur(time.Now(), "DropDatabaseID")
	}
	return tc.Client.DropDatabaseID(ctx, dbid, opts...)
}

func (tc *timedClient) ExecuteAction(ctx context.Context, dbid string, action string, tuples [][]any, opts ...client.TxOpt) (transactions.TxHash, error) {
	if tc.showReqDur {
		defer tc.printDur(time.Now(), "ExecuteAction")
	}
	return tc.Client.ExecuteAction(ctx, dbid, action, tuples, opts...)
}

func (tc *timedClient) Query(ctx context.Context, dbid string, query string) (*client.Records, error) {
	if tc.showReqDur {
		defer tc.printDur(time.Now(), "Query")
	}
	return tc.Client.Query(ctx, dbid, query)
}

func (tc *timedClient) WaitTx(ctx context.Context, txHash []byte, interval time.Duration) (*transactions.TcTxQueryResponse, error) {
	if tc.showReqDur {
		defer tc.printDur(time.Now(), "WaitTx")
	}
	return tc.Client.WaitTx(ctx, txHash, interval)
}

func (tc *timedClient) Transfer(ctx context.Context, to []byte, amount *big.Int, opts ...client.TxOpt) (transactions.TxHash, error) {
	if tc.showReqDur {
		defer tc.printDur(time.Now(), "Transfer")
	}
	return tc.Client.Transfer(ctx, to, amount, opts...)
}

func (tc *timedClient) printDur(t time.Time, method string) {
	tc.logger.Info(fmt.Sprintf("%s took %vms", method, float64(time.Since(t).Microseconds())/1e3)) // not using zap Fields so it is legible
}
