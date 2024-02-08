package main

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/core/types"
	clientType "github.com/kwilteam/kwil-db/core/types/client"
	"github.com/kwilteam/kwil-db/core/types/transactions"
)

// timedClient is a wrapper around a common.Client
type timedClient struct {
	clientType.Client
	showReqDur bool
	logger     *log.Logger
}

var _ clientType.Client = (*timedClient)(nil)

func (tc *timedClient) CallAction(ctx context.Context, dbid string, action string, inputs []any) (*clientType.Records, error) {
	if tc.showReqDur {
		defer tc.printDur(time.Now(), "CallAction")
	}
	return tc.Client.CallAction(ctx, dbid, action, inputs)
}

func (tc *timedClient) ChainID() string {
	if tc.showReqDur {
		defer tc.printDur(time.Now(), "ChainID")
	}
	return tc.Client.ChainID()
}

func (tc *timedClient) ChainInfo(ctx context.Context) (*types.ChainInfo, error) {
	if tc.showReqDur {
		defer tc.printDur(time.Now(), "ChainInfo")
	}
	return tc.Client.ChainInfo(ctx)
}

func (tc *timedClient) DeployDatabase(ctx context.Context, payload *transactions.Schema, opts ...clientType.TxOpt) (transactions.TxHash, error) {
	if tc.showReqDur {
		defer tc.printDur(time.Now(), "DeployDatabase")
	}
	return tc.Client.DeployDatabase(ctx, payload, opts...)
}

func (tc *timedClient) DropDatabase(ctx context.Context, name string, opts ...clientType.TxOpt) (transactions.TxHash, error) {
	if tc.showReqDur {
		defer tc.printDur(time.Now(), "DropDatabase")
	}
	return tc.Client.DropDatabase(ctx, name, opts...)
}

func (tc *timedClient) DropDatabaseID(ctx context.Context, dbid string, opts ...clientType.TxOpt) (transactions.TxHash, error) {
	if tc.showReqDur {
		defer tc.printDur(time.Now(), "DropDatabaseID")
	}
	return tc.Client.DropDatabaseID(ctx, dbid, opts...)
}

func (tc *timedClient) ExecuteAction(ctx context.Context, dbid string, action string, tuples [][]any, opts ...clientType.TxOpt) (transactions.TxHash, error) {
	if tc.showReqDur {
		defer tc.printDur(time.Now(), "ExecuteAction")
	}
	return tc.Client.ExecuteAction(ctx, dbid, action, tuples, opts...)
}

func (tc *timedClient) GetAccount(ctx context.Context, pubKey []byte, status types.AccountStatus) (*types.Account, error) {
	if tc.showReqDur {
		defer tc.printDur(time.Now(), "GetAccount")
	}
	return tc.Client.GetAccount(ctx, pubKey, status)
}

func (tc *timedClient) GetSchema(ctx context.Context, dbid string) (*transactions.Schema, error) {
	if tc.showReqDur {
		defer tc.printDur(time.Now(), "GetSchema")
	}
	return tc.Client.GetSchema(ctx, dbid)
}

func (tc *timedClient) ListDatabases(ctx context.Context, owner []byte) ([]*types.DatasetIdentifier, error) {
	if tc.showReqDur {
		defer tc.printDur(time.Now(), "ListDatabases")
	}
	return tc.Client.ListDatabases(ctx, owner)
}

func (tc *timedClient) Ping(ctx context.Context) (string, error) {
	if tc.showReqDur {
		defer tc.printDur(time.Now(), "Ping")
	}
	return tc.Client.Ping(ctx)
}

func (tc *timedClient) Query(ctx context.Context, dbid string, query string) (*clientType.Records, error) {
	if tc.showReqDur {
		defer tc.printDur(time.Now(), "Query")
	}
	return tc.Client.Query(ctx, dbid, query)
}

func (tc *timedClient) TxQuery(ctx context.Context, txHash []byte) (*transactions.TcTxQueryResponse, error) {
	if tc.showReqDur {
		defer tc.printDur(time.Now(), "TxQuery")
	}
	return tc.Client.TxQuery(ctx, txHash)
}

func (tc *timedClient) WaitTx(ctx context.Context, txHash []byte, interval time.Duration) (*transactions.TcTxQueryResponse, error) {
	if tc.showReqDur {
		defer tc.printDur(time.Now(), "WaitTx")
	}
	return tc.Client.WaitTx(ctx, txHash, interval)
}

func (tc *timedClient) Transfer(ctx context.Context, to []byte, amount *big.Int, opts ...clientType.TxOpt) (transactions.TxHash, error) {
	if tc.showReqDur {
		defer tc.printDur(time.Now(), "Transfer")
	}
	return tc.Client.Transfer(ctx, to, amount, opts...)
}

func (tc *timedClient) printDur(t time.Time, method string) {
	tc.logger.Info(fmt.Sprintf("%s took %vms", method, float64(time.Since(t).Microseconds())/1e3)) // not using zap Fields so it is legible
}
