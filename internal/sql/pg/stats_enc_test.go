package pg

import (
	"bytes"
	"testing"

	gocmp "github.com/google/go-cmp/cmp"

	"github.com/kwilteam/kwil-db/common/sql"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/types/decimal"
	"github.com/stretchr/testify/require"
)

func TestStatsEncoding(t *testing.T) {
	var buf bytes.Buffer

	// tables slice so the test loop is ordered
	var tblRefs []sql.TableRef
	tblRefs = append(tblRefs, sql.TableRef{ // native Go type, int64
		Namespace: "ns",
		Table:     "int64",
	})
	tblRefs = append(tblRefs, sql.TableRef{ // Kwil type Decimal
		Namespace: "ns",
		Table:     "decimal",
	})
	tblRefs = append(tblRefs, sql.TableRef{ // Kwil type Uint256
		Namespace: "ns",
		Table:     "uint256",
	})
	tblRefs = append(tblRefs, sql.TableRef{ // Kwil type UUID
		Namespace: "ns",
		Table:     "uuid",
	})
	tblRefs = append(tblRefs, sql.TableRef{ // byte slice
		Namespace: "ns",
		Table:     "[]byte",
	})

	in := map[sql.TableRef]*sql.Statistics{
		tblRefs[0]: {
			RowCount: 1234,
			ColumnStatistics: []sql.ColumnStatistics{
				{
					NullCount: 42,
					Min:       int64(1),
					MinCount:  2,
					Max:       int64(6),
					MaxCount:  1,
					MCVals:    []any{int64(1), int64(1), int64(4), int64(5), int64(6)},
					MCFreqs:   []int{2, 9, 3, 1, 1},
					Histogram: newHisto([]int64{0, 12}),
				},
			},
		},
		tblRefs[1]: {
			RowCount: 1234,
			ColumnStatistics: []sql.ColumnStatistics{
				{
					NullCount: 42,
					Min:       mustDecimal("1"),
					MinCount:  2,
					Max:       mustDecimal("6"),
					MaxCount:  1,
					MCVals:    []any{mustDecimal("1"), mustDecimal("1"), mustDecimal("4"), mustDecimal("5"), mustDecimal("6")},
					MCFreqs:   []int{2, 9, 3, 1, 1},
					Histogram: newHisto([]*decimal.Decimal{mustDecimal("2"), mustDecimal("88")}),
				},
			},
		},
		tblRefs[2]: {
			RowCount: 1234,
			ColumnStatistics: []sql.ColumnStatistics{
				{
					NullCount: 42,
					Min:       mustUint256("1"),
					MinCount:  2,
					Max:       mustUint256("6"),
					MaxCount:  1,
					MCVals:    []any{mustUint256("1"), mustUint256("1"), mustUint256("4"), mustUint256("5"), mustUint256("6")},
					MCFreqs:   []int{2, 9, 3, 1, 1},
					Histogram: newHisto([]*types.Uint256{mustUint256("12"), mustUint256("22")}),
				},
			},
		},
		tblRefs[3]: { // uuid
			RowCount: 1645,
			ColumnStatistics: []sql.ColumnStatistics{
				{
					NullCount: 8,
					Min:       mustParseUUID("0000857c-8671-4f4e-99bd-fcc621f9d3d1"),
					MinCount:  6,
					Max:       mustParseUUID("9000857c-8671-4f4e-99bd-fcc621f9d3d1"),
					MaxCount:  789,
					MCVals: []any{mustParseUUID("0000857c-8671-4f4e-99bd-fcc621f9d3d1"),
						mustParseUUID("1000857c-8671-4f4e-99bd-fcc621f9d3d1"),
						mustParseUUID("2000857c-8671-4f4e-99bd-fcc621f9d3d1"),
						mustParseUUID("3000857c-8671-4f4e-99bd-fcc621f9d3d1"),
						mustParseUUID("9000857c-8671-4f4e-99bd-fcc621f9d3d1"),
					},
					MCFreqs: []int{6, 9, 3, 1, 789},
					Histogram: newHisto([]*types.UUID{mustParseUUID("2900857c-8671-4f4e-99bd-fcc621f9d3d1"),
						mustParseUUID("3900857c-8671-4f4e-99bd-fcc621f9d3d1")}),
				},
			},
		},
		tblRefs[4]: { // []byte
			RowCount: 88,
			ColumnStatistics: []sql.ColumnStatistics{
				{
					NullCount: 42,
					Min:       []byte{}, // important distinction with non-nil empty slice
					MinCount:  2,
					Max:       []byte{0xff, 0xff, 0xff},
					MaxCount:  1,
					MCVals:    []any{[]byte{0}, []byte{1}, []byte{2}, []byte{0xff, 0xff, 0xff}},
					MCFreqs:   []int{2, 9, 3, 1},
					Histogram: newHisto([][]byte{{1}, {8}}),
				},
			},
		},
	}

	err := EncStats(&buf, in)
	if err != nil {
		t.Fatal(err)
	}

	bts := buf.Bytes() // t.Logf("encoding length = %d", len(bts))

	rd := bytes.NewReader(bts)
	out, err := DecStats(rd)
	if err != nil {
		t.Fatal(err)
	}

	if len(out) != len(in) {
		t.Fatal("maps length not the same")
	}

	hcomp := []gocmp.Option{
		// gocmp.Comparer(func(a, b histo[int64]) bool {
		// 	return gocmp.Equal(a.bounds, b.bounds) && gocmp.Equal(a.freqs, b.freqs)
		// }), etc.
		//
		// Instead of gocmp.Comparer ^ for each histo instance, we define the
		// Equal method for histo[T].
	}

	for _, tblRef := range tblRefs {
		outStat, have := out[tblRef]
		if !have {
			t.Fatalf("output stats lack table %v", tblRef)
		}
		stat := in[tblRef]

		require.True(t, gocmp.Equal(stat, outStat, hcomp...), tblRef.String()+gocmp.Diff(stat, outStat))
	}
}
