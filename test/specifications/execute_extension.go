package specifications

import (
	"context"
	"testing"

	"github.com/cstockton/go-conv"
	kTx "github.com/kwilteam/kwil-db/pkg/tx"
	"github.com/stretchr/testify/assert"
)

const (
	divideActionName = "divide"
)

type ExecuteExtensionDsl interface {
	ExecuteAction(ctx context.Context, dbid string, actionName string, actionInputs []map[string]any) (*kTx.Receipt, []map[string]any, error)
}

func ExecuteExtensionSpecification(ctx context.Context, t *testing.T, execute ExecuteExtensionDsl) {
	t.Logf("Executing insert action specification")

	db := SchemaLoader.Load(t, schema_testdb)
	dbID := GenerateSchemaId(db.Owner, db.Name)

	receipt, results, err := execute.ExecuteAction(ctx, dbID, divideActionName, []map[string]any{
		{
			"$numerator":   3,
			"$denominator": 2,
		},
	})
	assert.NoError(t, err)
	assert.NotNil(t, receipt)

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	upperValue, ok := results[0]["upper_value"]
	if !ok {
		t.Fatalf("expected upper_value to be present")
	}
	upperValueInt, err := conv.Int64(upperValue)
	if err != nil {
		t.Fatalf("expected upper_value to be an int")
	}

	lowerValue, ok := results[0]["lower_value"]
	if !ok {
		t.Fatalf("expected lower_value to be present")
	}
	lowerValueInt, err := conv.Int64(lowerValue)
	if err != nil {
		t.Fatalf("expected lower_value to be an int")
	}

	if upperValueInt != 2 {
		t.Fatalf("expected upper_value to be 2, got %d", upperValueInt)
	}

	if lowerValueInt != 1 {
		t.Fatalf("expected lower_value to be 1, got %d", lowerValueInt)
	}
}
