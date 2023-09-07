package specifications

import (
	"context"
	"testing"

	"github.com/cstockton/go-conv"
	"github.com/stretchr/testify/assert"
)

const (
	divideActionName = "divide"
)

type ExecuteExtensionDsl interface {
	DatabaseIdentifier
	TxQueryDsl
	ExecuteCallDsl
	ExecuteAction(ctx context.Context, dbid string, actionName string, actionInputs ...[]any) ([]byte, error)
}

func ExecuteExtensionSpecification(ctx context.Context, t *testing.T, execute ExecuteExtensionDsl) {
	t.Logf("Executing insert action specification")

	db := SchemaLoader.Load(t, schemaTestDB)
	dbID := execute.DBID(db.Name)

	// try executing extension
	txHash, err := execute.ExecuteAction(ctx, dbID, divideActionName, []any{3, 2})
	assert.NoError(t, err)

	expectTxSuccess(t, execute, ctx, txHash, defaultTxQueryTimeout)()

	records, err := execute.Call(ctx, dbID, divideActionName, []any{3, 2})
	assert.NoError(t, err)

	results := records.Export()

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	upper, ok := results[0]["upper_value"]
	assert.True(t, ok)

	upperInt, err := conv.Int64(upper)
	assert.NoError(t, err)

	lower, ok := results[0]["lower_value"]
	assert.True(t, ok)

	lowerInt, err := conv.Int64(lower)
	assert.NoError(t, err)

	if upperInt != 2 {
		t.Fatalf("expected upper_value to be 2, got %d", upperInt)
	}

	if lowerInt != 1 {
		t.Fatalf("expected lower_value to be 1, got %d", lowerInt)
	}

	// TODO: try calling an extension in an execution, and having that execution fail

	// try calling extension

}
