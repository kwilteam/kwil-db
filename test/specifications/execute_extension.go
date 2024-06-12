package specifications

import (
	"context"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	divideActionName = "divide"
)

func ExecuteExtensionSpecification(ctx context.Context, t *testing.T, execute ExecuteExtensionDsl) {
	t.Logf("Executing extension specification")

	db := SchemaLoader.Load(t, SchemaTestDB)
	dbID := execute.DBID(db.Name)

	// try executing extension
	txHash, err := execute.Execute(ctx, dbID, divideActionName, []any{2, 1, 2})
	assert.NoError(t, err)

	expectTxSuccess(t, execute, ctx, txHash, defaultTxQueryTimeout)()

	records, err := execute.Call(ctx, dbID, divideActionName, []any{2, 1, 2})
	assert.NoError(t, err)

	results := records.Export()

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	upper, ok := results[0]["upper_value"]
	assert.True(t, ok)

	upperText, ok := upper.(string)
	require.Truef(t, ok, "expected a string, got a %T", upper)
	upperInt, err := strconv.ParseInt(upperText, 10, 64)
	require.NoError(t, err)

	lower, ok := results[0]["lower_value"]
	assert.True(t, ok)

	lowerText, ok := lower.(string)
	require.Truef(t, ok, "expected a string, got a %T", lower)
	lowerInt, err := strconv.ParseInt(lowerText, 10, 64)
	require.NoError(t, err)

	if upperInt != 2 {
		t.Fatalf("expected upper_value to be 2, got %d", upperInt)
	}

	if lowerInt != 1 {
		t.Fatalf("expected lower_value to be 1, got %d", lowerInt)
	}

	// TODO: try calling an extension in an execution, and having that execution fail

	// try calling extension

}
