//go:build pglive

package pg

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTableStats(t *testing.T) {
	ctx := context.Background()

	db, err := NewDB(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	tx, err := db.BeginTx(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)

	tbl := "colcheck"
	_, err = tx.Execute(ctx, `drop table if exists `+tbl)
	if err != nil {
		t.Fatal(err)
	}
	_, err = tx.Execute(ctx, `create table if not exists `+tbl+
		` (a int8 primary key, b int4 default 42, c text, d bytea, e numeric(20,5), f int8[], g uint256, h uint256[])`)
	if err != nil {
		t.Fatal(err)
	}

	cols, err := ColumnInfo(ctx, tx, "", tbl)
	if err != nil {
		t.Fatal(err)
	}

	wantCols := []ColInfo{
		{Pos: 1, Name: "a", DataType: "bigint", Nullable: false},
		{Pos: 2, Name: "b", DataType: "integer", Nullable: true, defaultVal: "42"},
		{Pos: 3, Name: "c", DataType: "text", Nullable: true},
		{Pos: 4, Name: "d", DataType: "bytea", Nullable: true},
		{Pos: 5, Name: "e", DataType: "numeric", Nullable: true},
		{Pos: 6, Name: "f", DataType: "bigint", Array: true, Nullable: true},
		{Pos: 7, Name: "g", DataType: "uint256", Nullable: true},
		{Pos: 8, Name: "h", DataType: "uint256", Array: true, Nullable: true},
	}

	assert.Equal(t, wantCols, cols)
	// t.Logf("%#v", cols)

	_, err = tx.Execute(ctx, `insert into `+tbl+` values `+
		`(5, null, '', '\xabab', 12.6, '{99}', 30, '{}'), `+
		`(-1, 0, 'B', '\x01', -7, '{1, 2}', 20, '{184467440737095516150}'), `+
		`(3, 1, null, '\x', 8.1, NULL, NULL, NULL), `+
		`(0, 0, 'Q', NULL, NULL, NULL, NULL, NULL), `+
		`(7, -4, 'c', '\x0001', 0.3333, '{2,3,4}', 40, '{5,4,3}')`)
	if err != nil {
		t.Fatal(err)
	}

	stats, err := TableStats(ctx, "", tbl, tx)
	require.NoError(t, err)

	t.Log(stats)

	fmt.Println(stats.ColumnStatistics[4].Min)
	fmt.Println(stats.ColumnStatistics[4].Max)
}

/*func TestScanBig(t *testing.T) {
// This test is commented, but helpful for benchmarking performance with a large table.
	ctx := context.Background()

	cfg := *cfg
	cfg.User = "kwild"
	cfg.Pass = "kwild"
	cfg.DBName = "kwil_test_db"

	db, err := NewPool(ctx, &cfg.PoolConfig)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	tx, err := db.BeginTx(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)

	tbl := `giant`
	cols, err := ColumnInfo(ctx, tx, tbl)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%#v", cols)

	stats, err := TableStats(ctx, tbl, tx)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(stats)
}*/

func TestCmpBool(t *testing.T) {
	tests := []struct {
		name     string
		a        bool
		b        bool
		expected int
	}{
		{"true_true", true, true, 0},
		{"false_false", false, false, 0},
		{"true_false", true, false, 1},
		{"false_true", false, true, -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cmpBool(tt.a, tt.b)
			assert.Equal(t, tt.expected, result, "cmpBool(%v, %v) = %v; want %v", tt.a, tt.b, result, tt.expected)
		})
	}
}

func TestCmpBoolSymmetry(t *testing.T) {
	booleans := []bool{true, false}

	for _, a := range booleans {
		for _, b := range booleans {
			t.Run(fmt.Sprintf("a=%v,b=%v", a, b), func(t *testing.T) {
				result1 := cmpBool(a, b)
				result2 := cmpBool(b, a)
				assert.Equal(t, -result2, result1, "cmpBool(%v, %v) and cmpBool(%v, %v) are not symmetric", a, b, b, a)
			})
		}
	}
}

func TestCmpBoolTransitivity(t *testing.T) {
	a, b, c := false, true, true

	ab := cmpBool(a, b)
	bc := cmpBool(b, c)
	ac := cmpBool(a, c)

	assert.True(t, (ab < 0 && bc <= 0) == (ac < 0), "cmpBool lacks transitivity")
}
