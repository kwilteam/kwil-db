package specifications

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

type NoticeDsl interface {
	ExecuteCallDsl
	ExecuteActionsDsl
	DatabaseDeployDsl
	TxInfoer
}

func ExecuteNoticeSpecification(ctx context.Context, t *testing.T, caller NoticeDsl) {
	t.Logf("Executing notice specification")

	// Given a valid database schema
	db := SchemaLoader.Load(t, LogDB)

	// When i deploy the database
	txHash, err := caller.DeployDatabase(ctx, db)
	require.NoError(t, err, "failed to send deploy database tx")

	// Then i expect success
	expectTxSuccess(t, caller, ctx, txHash, defaultTxQueryTimeout)

	// And i expect database should exist
	err = caller.DatabaseExists(ctx, caller.DBID(db.Name))
	require.NoError(t, err)

	procedure := "make_logs"
	args := []any{[]any{"a", "b", "c"}}

	// we now test logs with both consensus and non-consensus (execute and call)
	res, err := caller.Execute(ctx, caller.DBID(db.Name), procedure, args)
	require.NoError(t, err)

	ExpectTxSuccess(t, caller, ctx, res)

	// now we read the logs from consensus.
	info, err := caller.TxInfo(ctx, res)
	require.NoError(t, err)

	require.Contains(t, info.TxResult.Log, "a\nb\nc")

	// now we read the logs from non-consensus.
	callRes, err := caller.Call(ctx, caller.DBID(db.Name), procedure, args)
	require.NoError(t, err)

	require.EqualValues(t, callRes.Logs, []string{"a", "b", "c"})
}
