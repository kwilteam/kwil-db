package main

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/kwilteam/kwil-db/core/client"
	"github.com/kwilteam/kwil-db/core/log"
	rpcClient "github.com/kwilteam/kwil-db/core/rpc/client"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/types/transactions"
)

type timedClient struct {
	showReqDur bool
	logger     *log.Logger
	cl         client.TxClient
}

func (tc *timedClient) printDur(t time.Time, method string) {
	tc.logger.Info(fmt.Sprintf("%s took %vms", method, float64(time.Since(t).Microseconds())/1e3)) // not using zap Fields so it is legible
}

func (tc *timedClient) Call(ctx context.Context, req *transactions.CallMessage, _ ...rpcClient.ActionCallOption) ([]map[string]any, error) {
	if tc.showReqDur {
		defer tc.printDur(time.Now(), "Call")
	}
	return tc.cl.Call(ctx, req)
}

func (tc *timedClient) TxQuery(ctx context.Context, txHash []byte) (*transactions.TcTxQueryResponse, error) {
	if tc.showReqDur {
		defer tc.printDur(time.Now(), "TxQuery")
	}
	return tc.cl.TxQuery(ctx, txHash)
}

func (tc *timedClient) GetSchema(ctx context.Context, dbid string) (*transactions.Schema, error) {
	if tc.showReqDur {
		defer tc.printDur(time.Now(), "GetSchema")
	}
	return tc.cl.GetSchema(ctx, dbid)
}

func (tc *timedClient) Query(ctx context.Context, dbid string, query string) ([]map[string]any, error) {
	if tc.showReqDur {
		defer tc.printDur(time.Now(), "Query")
	}
	return tc.cl.Query(ctx, dbid, query)
}

func (tc *timedClient) ListDatabases(ctx context.Context, ownerIdentifier []byte) ([]string, error) {
	if tc.showReqDur {
		defer tc.printDur(time.Now(), "ListDatabases")
	}
	return tc.cl.ListDatabases(ctx, ownerIdentifier)
}

func (tc *timedClient) GetAccount(ctx context.Context, acctID []byte, status types.AccountStatus) (*types.Account, error) {
	if tc.showReqDur {
		defer tc.printDur(time.Now(), "GetAccount")
	}
	return tc.cl.GetAccount(ctx, acctID, status)
}

func (tc *timedClient) Broadcast(ctx context.Context, tx *transactions.Transaction) ([]byte, error) {
	if tc.showReqDur {
		defer tc.printDur(time.Now(), "Broadcast")
	}
	return tc.cl.Broadcast(ctx, tx)
}

func (tc *timedClient) Ping(ctx context.Context) (string, error) {
	if tc.showReqDur {
		defer tc.printDur(time.Now(), "Ping")
	}
	return tc.cl.Ping(ctx)
}

func (tc *timedClient) ChainInfo(ctx context.Context) (*types.ChainInfo, error) {
	if tc.showReqDur {
		defer tc.printDur(time.Now(), "ChainInfo")
	}
	return tc.cl.ChainInfo(ctx)
}

func (tc *timedClient) EstimateCost(ctx context.Context, tx *transactions.Transaction) (*big.Int, error) {
	if tc.showReqDur {
		defer tc.printDur(time.Now(), "EstimateCost")
	}
	return tc.cl.EstimateCost(ctx, tx)
}
