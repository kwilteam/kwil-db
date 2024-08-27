//go:build pglive

package pg

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kwilteam/kwil-db/common/sql"
	"github.com/kwilteam/kwil-db/core/types"
)

func mkTestTableDB(t *testing.T) *DB {
	ctx := context.Background()
	db, err := NewDB(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		defer db.Close()
	})
	return db
}

func mkStatsTestTableTx(t *testing.T, db *DB) sql.PreparedTx {
	ctx := context.Background()
	tx, err := db.BeginPreparedTx(ctx)
	// tx, err := db.BeginTx(ctx)
	if err != nil {
		t.Fatal(err)
	}
	tbl := "colcheck"
	t.Cleanup(func() {
		tx.Rollback(ctx)
		if t.Failed() {
			db.AutoCommit(true)
			db.Execute(ctx, `drop table if exists `+tbl)
		}
	})

	_, err = tx.Execute(ctx, `drop table if exists `+tbl)
	if err != nil {
		t.Fatal(err)
	}
	_, err = tx.Execute(ctx, `create table if not exists `+tbl+
		` (a int8 primary key, b int8 default 42, c text, d bytea, 
		e numeric(20,5), f int4[], g uint256, h uint256[], i uuid)`)
	if err != nil {
		t.Fatal(err)
	}

	return tx
}

func TestTableStats(t *testing.T) {
	ctx := context.Background()
	db := mkTestTableDB(t)
	tx := mkStatsTestTableTx(t, db)

	tbl := "colcheck"

	cols, err := ColumnInfo(ctx, tx, "", tbl)
	if err != nil {
		t.Fatal(err)
	}

	wantCols := []ColInfo{
		{Pos: 1, Name: "a", DataType: "bigint", Nullable: false},
		{Pos: 2, Name: "b", DataType: "bigint", Nullable: true, defaultVal: "42"},
		{Pos: 3, Name: "c", DataType: "text", Nullable: true},
		{Pos: 4, Name: "d", DataType: "bytea", Nullable: true},
		{Pos: 5, Name: "e", DataType: "numeric", Nullable: true},
		{Pos: 6, Name: "f", DataType: "integer", Array: true, Nullable: true},
		{Pos: 7, Name: "g", DataType: "uint256", Nullable: true},
		{Pos: 8, Name: "h", DataType: "uint256", Array: true, Nullable: true},
		{Pos: 9, Name: "i", DataType: "uuid", Nullable: true},
	}

	assert.Equal(t, wantCols, cols)

	/* equivalent insert:
	_, err = tx.Execute(ctx, `insert into `+tbl+` values `+
		`(5, null, '', '\xabab', 12.6, '{99}', 30, '{}', '0000857c-8671-4f4e-99bd-fcc621f9d3d1'), `+
		`(-1, 0, 'B', '\x01', -7, '{1, 2}', 20, '{184467440737095516150}', '9000857c-8671-4f4e-99bd-fcc621f9d3d1'), `+
		`(3, 1, null, '\x', 8.1, NULL, NULL, NULL, NULL), `+
		`(0, 0, 'Q', NULL, NULL, NULL, NULL, NULL, NULL), `+
		`(7, -4, 'c', '\x0001', 0.3333, '{2,3,4}', 40, '{5,4,3}', '0000157c-8671-4f4e-99bd-fcc621f9d3d1')`)
	if err != nil {
		t.Fatal(err)
	}*/
	rows := [][]any{
		{int64(5), nil, "", []byte{0xab, 0xab}, mustDecimal("12.6"), []int64{99}, mustUint256("30"),
			types.Uint256Array{}, mustParseUUID("0000857c-8671-4f4e-99bd-fcc621f9d3d1")},
		{int64(-1), int64(0), "B", []byte{0x1}, mustDecimal("-7"), []int64{1, 2}, mustUint256("20"),
			types.Uint256Array{mustUint256("184467440737095516150")}, mustParseUUID("9000857c-8671-4f4e-99bd-fcc621f9d3d1")},
		{int64(3), int64(1), nil, []byte{}, mustDecimal("8.1"), nil, nil,
			nil, nil},
		{int64(0), int64(0), "Q", nil, nil, nil, nil,
			nil, nil},
		{int64(7), int64(-4), "c", []byte{0x0, 0x1}, mustDecimal("0.333"), []int64{2, 3, 4}, mustUint256("40"),
			types.Uint256Array{mustUint256("5"), mustUint256("4"), mustUint256("3")}, mustParseUUID("0000157c-8671-4f4e-99bd-fcc621f9d3d1")},
	}
	for _, row := range rows {
		_, err = tx.Execute(ctx, `insert into `+tbl+` values ($1, $2, $3, $4, $5, $6, $7, $8, $9)`, row...)
		if err != nil {
			t.Fatal(err)
		}
	}

	wantStats := []struct {
		Min any
		Max any
	}{
		{ // a
			Min: int64(-1),
			Max: int64(7),
		},
		{ // b
			Min: int64(-4),
			Max: int64(1),
		},
		{ // c
			Min: "",
			Max: "c",
		},
		{ // d
			Min: []byte{},           // []
			Max: []byte{0xab, 0xab}, // [171 171]
		},
		{ // e
			Min: mustDecimal("-7.00000"), // we don't control the precision we get back, only scale
			Max: mustDecimal("12.60000"),
		},
		{ // f -- TODO array min/max!
			Min: nil,
			Max: nil,
		},
		{ // g
			Min: mustUint256("20"),
			Max: mustUint256("40"),
		},
		{ // h -- TODO array min/max!
			Min: nil,
			Max: nil,
		},
		{ // i -- UUID
			Min: mustParseUUID("0000157c-8671-4f4e-99bd-fcc621f9d3d1"),
			Max: mustParseUUID("9000857c-8671-4f4e-99bd-fcc621f9d3d1"),
		},
	}
	//t.Log(wantStats)

	stats, err := TableStats(ctx, "", tbl, tx)
	require.NoError(t, err)

	for i := range stats.ColumnStatistics {
		cs := &stats.ColumnStatistics[i]

		require.Equal(t, wantStats[i].Min, cs.Min)
		require.Equal(t, wantStats[i].Max, cs.Max)
	}
}

// Test_scanSineBig is similar to Test_updates_demo, but actually uses a DB,
// inserting data into a table and testing the TableStats function.
func Test_scanSineBig(t *testing.T) {
	// Build the full set of values
	// sine wave with 100 samples per periods, 100 periods
	const numUpdates = 40000
	const samplesPerPeriod = 100
	const ampl = 200.0   // larger => more integer discretization
	const amplSteps = 10 // "noise" with small ampl variations between periods
	const amplInc = 2.0  // each step adds a multiple of this to the amplitude
	vals := makeTestVals(numUpdates, samplesPerPeriod, amplSteps, ampl, amplInc)

	ctx := context.Background()

	db := mkTestTableDB(t)
	tx := mkStatsTestTableTx(t, db)
	tbl := `colcheck`

	for i, val := range vals {
		_, err := tx.Execute(ctx, `INSERT INTO `+tbl+` VALUES($1,$2,$3);`,
			i, val, strconv.FormatInt(val, 10))
		require.NoError(t, err)
	}

	stats, err := TableStats(ctx, "", tbl, tx)
	require.NoError(t, err)

	require.True(t, stats.RowCount == numUpdates)

	// check the MCVs for the int8 column
	col := stats.ColumnStatistics[1]

	require.Equal(t, len(col.MCFreqs), statsCap)
	require.Equal(t, len(col.MCVals), statsCap)

	_, ok := col.MCVals[0].(int64)
	require.True(t, ok, "wrong value type")

	valsT := convSliceAsserted[int64](col.MCVals)
	require.True(t, slices.IsSorted(valsT))

	t.Log(valsT)
	t.Log(col.MCFreqs)

	var totalFreqMCVs int
	for _, f := range col.MCFreqs {
		totalFreqMCVs += f
	}
	fracMCVs := float64(totalFreqMCVs) / numUpdates
	t.Log(fracMCVs)

	require.Greater(t, totalFreqMCVs, statsCap) // not just all ones
	require.LessOrEqual(t, totalFreqMCVs, numUpdates)

	hist := col.Histogram.(histo[int64])
	t.Log(hist)
	var totalFreqHist int
	for _, f := range hist.freqs {
		totalFreqHist += f
	}
	fracHists := float64(totalFreqHist) / numUpdates
	t.Log(fracHists)

	t.Log(fracMCVs + fracHists)

	t.Log(col.Min.(int64), col.Max.(int64))
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
	cols, err := ColumnInfo(ctx, tx, "", tbl)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%#v", cols)

	stats, err := TableStats(ctx, "", tbl, tx)
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
