package test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/test/setup"
	"github.com/stretchr/testify/require"
)

var txTimeout = 5 * time.Second

func ExpectTxSuccess(t *testing.T, spec setup.JSONRPCClient, ctx context.Context, txHash types.Hash) {
	res, err := spec.WaitTx(ctx, txHash, 300*time.Millisecond)
	require.NoError(t, err)

	require.True(t, res.Result.Code == 0, "tx failed: %s", res.Result.Log)
}

func ExpectTxError(t *testing.T, spec setup.JSONRPCClient, ctx context.Context, txHash types.Hash, msg string) {
	res, err := spec.WaitTx(ctx, txHash, 300*time.Millisecond)
	require.NoError(t, err)

	require.True(t, res.Result.Code != 0, "tx succeeded")
	require.True(t, strings.Contains(res.Result.Log, msg), "unexpected error message")
}
