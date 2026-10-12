package usersvc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/trufnetwork/kwil-db/core/log"
	jsonrpc "github.com/trufnetwork/kwil-db/core/rpc/json"
	userjson "github.com/trufnetwork/kwil-db/core/rpc/json/user"
	"github.com/trufnetwork/kwil-db/core/types"
)

func TestBroadcastRejectsMissingTx(t *testing.T) {
	svc := &Service{log: log.DiscardLogger}

	_, rpcErr := svc.Broadcast(context.Background(), &userjson.BroadcastRequest{})
	require.NotNil(t, rpcErr)
	require.Equal(t, jsonrpc.ErrorInvalidParams, rpcErr.Code)
}

func TestEstimatePriceRejectsTxWithoutBody(t *testing.T) {
	svc := &Service{log: log.DiscardLogger}

	for _, req := range []*userjson.EstimatePriceRequest{
		{},
		{Tx: &types.Transaction{}},
	} {
		_, rpcErr := svc.EstimatePrice(context.Background(), req)
		require.NotNil(t, rpcErr)
		require.Equal(t, jsonrpc.ErrorInvalidParams, rpcErr.Code)
	}
}
