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
		` (a int8 not null, b int4 default 42, c text, d bytea, e numeric(20,5))`)
	if err != nil {
		t.Fatal(err)
	}

	cols, err := ColumnInfo(ctx, tx, tbl)
	if err != nil {
		t.Fatal(err)
	}

	wantCols := []ColInfo{
		{Pos: 1, Name: "a", DataType: "bigint", Nullable: false, Default: nil},
		{Pos: 2, Name: "b", DataType: "integer", Nullable: true, Default: "42"},
		{Pos: 3, Name: "c", DataType: "text", Nullable: true, Default: nil},
		{Pos: 4, Name: "d", DataType: "bytea", Nullable: true, Default: nil},
		{Pos: 5, Name: "e", DataType: "numeric", Nullable: true, Default: nil},
	}

	assert.Equal(t, wantCols, cols)
	// t.Logf("%#v", cols)

	_, err = tx.Execute(ctx, `insert into `+tbl+` values `+
		`(5, null, '', '\xabab', 12.6), `+
		`(-1, 0, 'B', '\x01', -7), `+
		`(3, 1, null, '\x', 8.1), `+
		`(0, 0, 'Q', NULL, NULL), `+
		`(7, -4, 'c', '\x0001', 0.3333)`)
	if err != nil {
		t.Fatal(err)
	}

	stats, err := TableStats(ctx, tbl, tx)
	require.NoError(t, err)

	// spew.Config.DisableMethods = true
	// defer func() { spew.Config.DisableMethods = false }()
	t.Log(stats)

	fmt.Println(stats.ColumnStatistics[4].Min)
	fmt.Println(stats.ColumnStatistics[4].Max)
}

/*func TestScanBig(t *testing.T) {
	ctx := context.Background()

	cfg := *cfg
	cfg.User = "dcrdata"
	cfg.Pass = "dcrdata"
	cfg.DBName = "dcrdata"

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

	tbl := `addresses`
	cols, err := ColumnInfo(ctx, tx, tbl)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%#v", cols)

	// scanner, ok :=  tx.(QueryScanner)
	// if !ok {
	// 	t.Fatal("tx not a QueryScanner")
	// }
	// scanner.QueryScanFn(ctx, fmt.Sprintf(`SELECT * FROM %s`, tlb))
	stats, err := TableStats(ctx, tbl, tx)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(stats)
}*/
