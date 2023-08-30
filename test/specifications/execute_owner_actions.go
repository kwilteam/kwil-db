package specifications

import (
	"context"
	"testing"

	"github.com/kwilteam/kwil-db/pkg/client"
	"github.com/stretchr/testify/assert"
)

const (
	ownerOnlyActionName = "owner_only"
)

type ExecuteOwnerActionsDsl interface {
	ExecuteQueryDsl
	ExecuteCallDsl
}

func ExecuteOwnerActionSpecification(ctx context.Context, t *testing.T, execute ExecuteOwnerActionsDsl) {
	t.Logf("Executing owner action specification")

	db := SchemaLoader.Load(t, schemaTestDB)
	dbID := execute.DBID(db.Name)

	actionInputs := []any{}
	txHash, err := execute.ExecuteAction(ctx, dbID, ownerOnlyActionName, actionInputs)
	assert.NoError(t, err, "error executing owner action")

	expectTxSuccess(t, execute, ctx, txHash, defaultTxQueryTimeout)()
}

func ExecuteOwnerActionFailSpecification(ctx context.Context, t *testing.T, execute ExecuteOwnerActionsDsl) {
	t.Logf("Executing owner action fail specification")

	db := SchemaLoader.Load(t, schemaTestDB)
	dbID := execute.DBID(db.Name)

	actionInputs := []any{}

	txHash, err := execute.ExecuteAction(ctx, dbID, ownerOnlyActionName, actionInputs)
	assert.NoError(t, err, "error executing owner action")

	expectTxFail(t, execute, ctx, txHash, defaultTxQueryTimeout)()

	// call authenticated, should fail
	_, err = execute.Call(ctx, dbID, ownerOnlyActionName, actionInputs, client.Authenticated(true))
	assert.Error(t, err, "expected error calling owner only action with authentication")
}
