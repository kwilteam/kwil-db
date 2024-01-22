package basestream

import (
	"context"
	"fmt"
	"testing"

	"github.com/kwilteam/kwil-db/internal/sql"
	"github.com/stretchr/testify/assert"
)

func Test_Index(t *testing.T) {
	ctx := context.Background()
	b := &BaseStreamExt{
		table:       "price",
		dateColumn:  "date",
		valueColumn: "value",
	}

	// when checking these values, we know that we are adding extra "precision" by returning values
	// as having been multiplied by 1000. This is because Kwil cannot handle decimals.
	// The number 1500 will be identified by Truflation stream clients as 1.500
	mockQ := &mockQuerier{
		stmts: map[string]*sql.ResultSet{
			b.sqlGetBaseValue():     mockScalar("value", int64(75000)),  // 75.000
			b.sqlGetLatestValue():   mockScalar("value", int64(200000)), // 200.000
			b.sqlGetSpecificValue(): mockScalar("value", int64(150000)), // 150.000
		},
	}

	returned, err := b.index(ctx, mockQ, "2024-01-01")
	assert.NoError(t, err)
	assert.Equal(t, int64(200000), returned) // 200.000 * 1000

	returned, err = b.index(ctx, mockQ, "") // this should return the latest value
	assert.NoError(t, err)
	returned2, err := b.index(ctx, mockQ, zeroDate) // this should return the latest value
	assert.NoError(t, err)
	assert.Equal(t, int64(266666), returned) // 266.666 * 1000
	assert.Equal(t, int64(266666), returned2)
}

func Test_Value(t *testing.T) {
	ctx := context.Background()
	b := &BaseStreamExt{
		table:       "price",
		dateColumn:  "date",
		valueColumn: "value",
	}

	mockQ := &mockQuerier{
		stmts: map[string]*sql.ResultSet{
			b.sqlGetLatestValue():   mockScalar("value", int64(200000)), // 200.000
			b.sqlGetSpecificValue(): mockScalar("value", int64(150000)), // 150.000
		},
	}

	returned, err := b.value(ctx, mockQ, "2024-01-01")
	assert.NoError(t, err)
	assert.Equal(t, int64(150000), returned) // 150.000 * 1000

	returned, err = b.value(ctx, mockQ, "") // this should return the latest value
	assert.NoError(t, err)
	returned2, err := b.value(ctx, mockQ, zeroDate) // this should return the latest value
	assert.NoError(t, err)
	assert.Equal(t, int64(200000), returned) // 200.000 * 1000
	assert.Equal(t, int64(200000), returned2)
}

// mockScalar is a helper function that creates a new actions.Result that
// returns the given value as a single row and column result.
func mockScalar(column string, v any) *sql.ResultSet {
	return &sql.ResultSet{
		ReturnedColumns: []string{column},
		Rows: [][]any{
			{v},
		},
	}
}

type mockQuerier struct {
	stmts map[string]*sql.ResultSet
}

func (m *mockQuerier) Query(ctx context.Context, stmt string, params map[string]any) (*sql.ResultSet, error) {
	res, ok := m.stmts[stmt]
	if !ok {
		return nil, fmt.Errorf("unexpected statement: %s", stmt)
	}
	return res, nil
}
