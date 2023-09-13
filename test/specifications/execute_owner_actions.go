package specifications

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	ownerOnlyActionName = "owner_only"
)

func ExecuteOwnerActionSpecification(ctx context.Context, t *testing.T, execute ExecuteOwnerActionsDsl) {
	t.Logf("Executing owner action specification")

	db := SchemaLoader.Load(t, SchemaTestDB)
	dbID := execute.DBID(db.Name)

	actionInputs := []any{}
	txHash, err := execute.ExecuteAction(ctx, dbID, ownerOnlyActionName, actionInputs)
	require.NoError(t, err, "error executing owner action")

	expectTxSuccess(t, execute, ctx, txHash, defaultTxQueryTimeout)()
}

func ExecuteOwnerActionFailSpecification(ctx context.Context, t *testing.T, execute ExecuteOwnerActionsDsl, dbID string) {
	t.Logf("Executing owner action fail specification")

	actionInputs := []any{}

	txHash, err := execute.ExecuteAction(ctx, dbID, ownerOnlyActionName, actionInputs)
	require.NoError(t, err, "error executing owner action")

	expectTxFail(t, execute, ctx, txHash, defaultTxQueryTimeout)()

	// call authenticated, should fail
	_, err = execute.Call(ctx, dbID, ownerOnlyActionName, actionInputs, true)
	require.Error(t, err, "expected error calling owner only action with authentication")
}
