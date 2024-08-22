package pg

import (
	"bytes"
	"cmp"
	"context"
	"errors"
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/kwilteam/kwil-db/common/sql"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/types/decimal"
)

// statsCap is the limit on the number of MCVs and histogram bins when working
// with column statistics. Fixed for now, but we should consider making a stats
// field or settable another way.
const statsCap = 100

// RowCount gets a precise row count for the named fully qualified table.
func RowCount(ctx context.Context, qualifiedTable string, db sql.Executor) (int64, error) {
	stmt := fmt.Sprintf(`SELECT count(1) FROM %s`, qualifiedTable)
	res, err := db.Execute(ctx, stmt)
	if err != nil {
		return 0, fmt.Errorf("unable to count rows: %w", err)
	}
	if len(res.Rows) != 1 || len(res.Rows[0]) != 1 {
		return 0, errors.New("exactly one value not returned by row count query")
	}
	count, ok := sql.Int64(res.Rows[0][0])
	if !ok {
		return 0, fmt.Errorf("no row count for %s", qualifiedTable)
	}
	return count, nil
}

// TableStatser is an interface that the implementation of a sql.Executor may
// implement.
type TableStatser interface {
	TableStats(ctx context.Context, schema, table string) (*sql.Statistics, error)
}

// TableStats collects deterministic statistics for a table. If schema is empty,
// the "public" schema is assumed. This method is used to obtain the ground
// truth statistics for a table; incremental statistics updates should be
// preferred when possible. If the sql.Executor implementation is a
// TableStatser, it's method is used directly. This is primarily to allow a stub
// DB for testing.
func TableStats(ctx context.Context, schema, table string, db sql.Executor) (*sql.Statistics, error) {
	if ts, ok := db.(TableStatser); ok {
		return ts.TableStats(ctx, schema, table)
	}

	if schema == "" {
		schema = "public"
	}
	qualifiedTable := schema + "." + table

	count, err := RowCount(ctx, qualifiedTable, db)
	if err != nil {
		return nil, err
	}
	// TODO: We needs a schema-table stats database so we don't ever have to do
	// a full table scan for column stats.

	colInfo, err := ColumnInfo(ctx, db, schema, table)
	if err != nil {
		return nil, err
	}

	// Column statistics
	colStats, err := colStats(ctx, qualifiedTable, colInfo, db)
	if err != nil {
		return nil, err
	}

	return &sql.Statistics{
		RowCount:         count,
		ColumnStatistics: colStats,
	}, nil
}

// rough outline for postgresql extension w/ a full stats function:
//
//  - function: collect_stats(tablename)
//  - iterate over each row, perform computations defined in the extension code
//  - SPI_connect() -> SPI_cursor_open(... query ...) -> SPI_cursor_fetch ->
//    SPI_processed -> SPI_tuptable -> SPI_getbinval

// colStats collects column-wise statistics for the specified table, using the
// provided column definitions to instantiate scan values used by the full scan
// that iterates over all rows of the table.
func colStats(ctx context.Context, qualifiedTable string, colInfo []ColInfo, db sql.Executor) ([]sql.ColumnStatistics, error) {
	// rowCount is unused now, and can seemingly be computed via the scan
	// itself, but I intend to use it in for more complex statistics building algos.

	// https://wiki.postgresql.org/wiki/Retrieve_primary_key_columns
	getIndBase := `SELECT a.attname::text, i.indexrelid::int8
		FROM pg_index i
		JOIN pg_attribute a ON a.attnum = ANY(i.indkey) AND a.attrelid = i.indrelid
		WHERE i.indrelid = '` + qualifiedTable + `'::regclass`
	// use primary key columns first
	getPK := getIndBase + ` AND i.indisprimary;`
	// then unique+not-null index cols?
	// getUniqueInds := getIndBase + ` AND i.indisunique;`
	res, err := db.Execute(ctx, getPK)
	if err != nil {
		return nil, err
	}
	// IMPORTANT NOTE: if the iteration over all rows of the table involves *no*
	// ORDER BY clause, the scan order is not guaranteed. This should be an
	// error for tables where stats must be deterministic.
	//
	// if len(res.Rows) == 0 {
	// 	return nil, errors.New("no suitable orderby column")
	// }
	pkCols := make([]string, len(res.Rows))
	for i, row := range res.Rows {
		pkCols[i] = row[0].(string)
	}

	stmt := `SELECT * FROM ` + qualifiedTable
	if len(pkCols) > 0 {
		stmt += ` ORDER BY ` + strings.Join(pkCols, ",")
	}

	colStats, err := colStatsInternal(ctx, nil, stmt, colInfo, db)
	if err != nil {
		return nil, err
	}

	// I sunk a bunch of time into a two-pass experiment to more accurate MCVs
	// and better histogram bounds, but this is quite costly.  May remove.

	return colStats, nil // single ASC pass
	// second DESC pass:
	// return colStatsInternal(ctx, colStats, stmt, colInfo, db)
}

func cmpBool(a, b bool) int {
	if b {
		if a { // true == true
			return 0
		}
		return -1 // false < true
	}
	if a {
		return 1 // true > false
	}
	return 0 // false == false
}

func cmpDecimal(val, mm *decimal.Decimal) int {
	d, err := val.Cmp(mm)
	if err != nil {
		panic(fmt.Sprintf("%s: (nan decimal?) %v or %v", err, val, mm))
	}
	return d
}

func colStatsInternal(ctx context.Context, firstPass []sql.ColumnStatistics,
	stmt string, colInfo []ColInfo, db sql.Executor) ([]sql.ColumnStatistics, error) {

	// The idea I'm testing with a two-pass scan is to collect mcvs from a
	// reverse-order scan. This is pointless if the MCVs is not at capacity as
	// nothing new can be added.
	if firstPass != nil {
		stmt += ` DESC` // otherwise ASC is implied with ORDER BY
	}

	getLast := func(i int) *sql.ColumnStatistics {
		if len(firstPass) != 0 {
			return &firstPass[i]
		}
		return nil
	}

	// Iterate over all rows (select *), scan into NULLable values like
	// pgtype.Int8, then make statistics with Go native or Kwil types.

	var scans []any
	colTypes := make([]ColType, len(colInfo))
	for i, col := range colInfo {
		colTypes[i] = colInfo[i].Type()
		scans = append(scans, col.scanVal())
	}

	colStats := make([]sql.ColumnStatistics, len(colInfo))

	err := QueryRowFunc(ctx, db, stmt, scans,
		func() error {
			var err error
			for i, val := range scans {
				stat := &colStats[i]
				if val == nil { // with QueryRowFuncAny and vals []any, or with QueryRowFunc where scans are native type pointers
					stat.NullCount++
					continue
				}

				// TODO: do something with array types (num elements stats????)

				switch colTypes[i] {
				case ColTypeInt: // use int64 in stats
					var valInt int64
					switch it := val.(type) {
					case interface{ Int64Value() (pgtype.Int8, error) }: // several of the pgtypes int types
						i8, err := it.Int64Value()
						if err != nil {
							return fmt.Errorf("bad int64: %T", val)
						}
						if !i8.Valid {
							stat.NullCount++
							continue
						}
						valInt = i8.Int64

					default:
						var ok bool
						valInt, ok = sql.Int64(val)
						if !ok {
							return fmt.Errorf("not int: %T", val)
						}
					}

					ins(stat, getLast(i), valInt, cmp.Compare[int64], interpNum)

				case ColTypeText: // use string in stats
					valStr, null, ok := TextValue(val) // val.(string)
					if !ok {
						return fmt.Errorf("not string: %T", val)
					}
					if null {
						stat.NullCount++
						continue
					}

					ins(stat, getLast(i), valStr, strings.Compare, interpString)

				case ColTypeByteA: // use []byte in stats
					var valBytea []byte
					switch vt := val.(type) {
					// Presently we're just using []byte, not pgtype.Array, but
					// might need to for NULL...

					// case *pgtype.Array[byte]:
					// 	if !vt.Valid {
					// 		stat.NullCount++
					// 		continue
					// 	}
					// 	valBytea = vt.Elements
					// case pgtype.Array[byte]:
					// 	if !vt.Valid {
					// 		stat.NullCount++
					// 		continue
					// 	}
					// 	valBytea = vt.Elements
					case *[]byte:
						if vt == nil || *vt == nil {
							stat.NullCount++
							continue
						}
						valBytea = slices.Clone(*vt)
					case []byte:
						if vt == nil {
							stat.NullCount++
							continue
						}
						valBytea = slices.Clone(vt)
					default:
						return fmt.Errorf("not bytea: %T", val)
					}

					ins(stat, getLast(i), valBytea, bytes.Compare, interpBts)

				case ColTypeBool: // use bool in stats
					var b bool
					switch v := val.(type) {
					case *pgtype.Bool:
						if !v.Valid {
							stat.NullCount++
							continue
						}
						b = v.Bool
					case pgtype.Bool:
						if !v.Valid {
							stat.NullCount++
							continue
						}
						b = v.Bool
					case *bool:
						b = *v
					case bool:
						b = v

					default:
						return fmt.Errorf("invalid bool (%T)", val)
					}

					ins(stat, getLast(i), b, cmpBool, interpBool) // there should never be a boolean histogram!

				case ColTypeNumeric: // use *decimal.Decimal in stats
					var dec *decimal.Decimal
					switch v := val.(type) {
					case *pgtype.Numeric:
						if !v.Valid {
							stat.NullCount++
							continue
						}
						if v.NaN {
							continue
						}

						dec, err = pgNumericToDecimal(*v)
						if err != nil {
							continue
						}

					case pgtype.Numeric:
						if !v.Valid {
							stat.NullCount++
							continue
						}
						if v.NaN {
							continue
						}

						dec, err = pgNumericToDecimal(v)
						if err != nil {
							continue
						}

					case *decimal.Decimal:
						if v.NaN() { // we're pretending this is NULL by our sql.Scanner's convetion
							stat.NullCount++
							continue
						}
						if v != nil {
							v2 := *v // clone!
							v = &v2
						}
						dec = v
					case decimal.Decimal:
						if v.NaN() { // we're pretending this is NULL by our sql.Scanner's convetion
							stat.NullCount++
							continue
						}
						v2 := v
						dec = &v2
					}

					ins(stat, getLast(i), dec, cmpDecimal, interpDec)

				case ColTypeUINT256:
					v, ok := val.(*types.Uint256)
					if !ok {
						return fmt.Errorf("not a *types.Uint256: %T", val)
					}

					if v.Null {
						stat.NullCount++
						continue
					}

					ins(stat, getLast(i), v.Clone(), types.CmpUint256, interpUint256)

				case ColTypeFloat: // we don't want, don't have
					var varFloat float64
					switch v := val.(type) {
					case *pgtype.Float8:
						if !v.Valid {
							stat.NullCount++
							continue
						}
						varFloat = v.Float64
					case *pgtype.Float4:
						if !v.Valid {
							stat.NullCount++
							continue
						}
						varFloat = float64(v.Float32)
					case pgtype.Float8:
						if !v.Valid {
							stat.NullCount++
							continue
						}
						varFloat = v.Float64
					case pgtype.Float4:
						if !v.Valid {
							stat.NullCount++
							continue
						}
						varFloat = float64(v.Float32)
					case float32:
						varFloat = float64(v)
					case float64:
						varFloat = v
					case *float32:
						varFloat = float64(*v)
					case *float64:
						varFloat = *v

					default:
						return fmt.Errorf("invalid float (%T)", val)
					}

					ins(stat, getLast(i), varFloat, cmp.Compare[float64], interpNum)

				case ColTypeUUID:
					fallthrough // TODO
				default: // arrays and such
					// fmt.Println("unknown", colTypes[i])
				}
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	// If this is a second pass, merge. See other comments about the two-pass
	// approach. It is costly, complex, and hard to quantify benefit. Likely to remove!
	for i, p := range firstPass {
		if len(colStats[i].MCFreqs) == 0 { // nothing new was recorded for this column
			colStats[i].MCFreqs = p.MCFreqs
			colStats[i].MCVals = p.MCVals
			continue
		}

		// merge up the last mcvs.  This is horribly inefficient for now:
		// concat, build slice of sortable struct, sort by frequency, extract
		// up to statsCap, re-sort by value.
		freqs := append(colStats[i].MCFreqs, p.MCFreqs...)
		vals := append(colStats[i].MCVals, p.MCVals...)

		type mcv struct {
			freq int
			val  any
		}
		mcvs := make([]mcv, len(freqs))
		for i := range freqs {
			mcvs[i] = mcv{freqs[i], vals[i]}
		}
		if len(mcvs) > statsCap { // drop the lowest frequency values
			slices.SortFunc(mcvs, func(a, b mcv) int {
				return cmp.Compare(b.freq, a.freq) // descending freq
			})
			mcvs = mcvs[:statsCap] // mcvs = slices.Delete(mcvs, statsCap, len(mcvs))
		}

		valCompFun := compFun(vals[0]) // based on prototype value
		slices.SortFunc(mcvs, func(a, b mcv) int {
			return valCompFun(b.val, a.val) // ascending value
		})

		// extract the values and frequencies slices
		freqs, vals = nil, nil // clear and preserve type
		for i := range mcvs {
			freqs = append(freqs, mcvs[i].freq)
			vals = append(vals, mcvs[i].val)
			if i == statsCap {
				break
			}
		}

		/* ALT in-place with sort.Interface
		if len(mcvs) > statsCap { // drop the lowest frequency values
			sort.Sort(mcvDescendingFreq{
				vals:  vals,
				freqs: freqs,
			})
			vals = vals[:statsCap]
			freqs = freqs[:statsCap]
		}
		sort.Sort(mcvAscendingValue{
			vals:  vals,
			freqs: freqs,
			comp:  compFun(vals[0]),
		}) */
		colStats[i].MCFreqs = freqs
		colStats[i].MCVals = vals
	}

	return colStats, nil
}

/* xx will remove if we decide a two-pass scan isn't worth it
type mcvDescendingFreq struct {
	vals  []any
	freqs []int
}

func (m mcvDescendingFreq) Len() int { return len(m.vals) }

func (m mcvDescendingFreq) Less(i int, j int) bool {
	return m.freqs[i] < m.freqs[j]
}

func (m mcvDescendingFreq) Swap(i int, j int) {
	m.vals[i], m.vals[j] = m.vals[j], m.vals[i]
	m.freqs[i], m.freqs[j] = m.freqs[j], m.freqs[i]
}

type mcvAscendingValue struct {
	vals  []any
	freqs []int
	comp  func(a, b any) int
}

func (m mcvAscendingValue) Len() int { return len(m.vals) }

func (m mcvAscendingValue) Less(i int, j int) bool {
	return m.comp(m.vals[i], m.vals[j]) == 0
}

func (m mcvAscendingValue) Swap(i int, j int) {
	m.vals[i], m.vals[j] = m.vals[j], m.vals[i]
	m.freqs[i], m.freqs[j] = m.freqs[j], m.freqs[i]
}
*/

// the pain of []any vs. []T (in an any)
func wrapCompFun[T any](f func(a, b T) int) func(a, b any) int {
	return func(a, b any) int { // must not be nil
		return f(a.(T), b.(T))
	}
}

// vs func compFun[T any]() func(a, b T) int
func compFun(val any) func(a, b any) int {
	switch val.(type) {
	case []byte:
		return wrapCompFun(bytes.Compare)
	case int64:
		return wrapCompFun(cmp.Compare[int64])
	case float64:
		return wrapCompFun(cmp.Compare[float64])
	case string:
		return wrapCompFun(strings.Compare)
	case bool:
		return wrapCompFun(cmpBool)
	case *decimal.Decimal:
		return wrapCompFun(cmpDecimal)
	case *types.Uint256:
		return wrapCompFun(types.CmpUint256)
	case types.UUID:
		return wrapCompFun(types.CmpUint256)

	case decimal.DecimalArray: // TODO
	case types.Uint256Array: // TODO
	case []string:
	case []int64:
	}

	panic(fmt.Sprintf("no comp fun for type %T", val))
}

// maybe we do this instead a comp field of histo[T]. It's simpler, but slower.
func compareStatsVal(a, b any) int { //nolint:unused
	switch at := a.(type) {
	case int64:
		return cmp.Compare(at, b.(int64))
	case bool:
		return cmpBool(at, b.(bool))
	case []bool:
		return slices.CompareFunc(at, b.([]bool), cmpBool)
	case string:
		return strings.Compare(at, b.(string))
	case []string:
		return slices.Compare(at, b.([]string))
	case []int64:
		return slices.Compare(at, b.([]int64))
	case float64:
		return cmp.Compare(at, b.(float64))
	case []byte:
		return bytes.Compare(at, b.([]byte))
	case [][]byte:
		return slices.CompareFunc(at, b.([][]byte), bytes.Compare)
	case []*decimal.Decimal:
		return slices.CompareFunc(at, b.([]*decimal.Decimal), cmpDecimal)
	case decimal.DecimalArray:
		return slices.CompareFunc(at, b.(decimal.DecimalArray), cmpDecimal)
	case *decimal.Decimal:
		return cmpDecimal(at, b.(*decimal.Decimal))
	case *types.Uint256:
		return at.Cmp(b.(*types.Uint256))
	case types.Uint256Array:
		return slices.CompareFunc(at, b.(types.Uint256Array), types.CmpUint256)
	case []*types.Uint256:
		return slices.CompareFunc(at, b.([]*types.Uint256), types.CmpUint256)
	case *types.UUID:
		return types.CmpUUID(*at, *(b.(*types.UUID)))
	case types.UUID:
		return types.CmpUUID(at, b.(types.UUID))
	case []*types.UUID:
		return slices.CompareFunc(at, b.([]*types.UUID), func(a, b *types.UUID) int {
			return types.CmpUUID(*a, *b)
		})
	default:
		panic(fmt.Sprintf("unrecognized type %T", a))
	}
}

// The following functions perform a type switch to correctly handle null values
// and then dispatch to the generic ins/up functions with the appropriate
// comparison function for the underlying type: upColStatsWithInsert,
// upColStatsWithDelete, and upColStatsWithUpdate.

// upColStatsWithInsert expects a value of the type created with a
// (*datatype).DeserializeChangeset method.
func upColStatsWithInsert(stats *sql.ColumnStatistics, val any) error {
	if val == nil {
		stats.NullCount++
		return nil
	}
	// INSERT
	switch nt := val.(type) {
	case []byte:
		return ins(stats, nil, nt, bytes.Compare, interpBts)
	case int64:
		return ins(stats, nil, nt, cmp.Compare[int64], interpNum[int64])
	case float64:
		return ins(stats, nil, nt, cmp.Compare[float64], interpNum[float64])
	case string:
		return ins(stats, nil, nt, strings.Compare, interpString)
	case bool:
		return ins(stats, nil, nt, cmpBool, interpBool)

	case *decimal.Decimal:
		if nt.NaN() {
			stats.NullCount++
			return nil // ignore, don't put in stats
		}
		nt2 := *nt
		return ins(stats, nil, &nt2, cmpDecimal, interpDec)

	case *types.Uint256:
		if nt.Null {
			stats.NullCount++
			return nil
		}

		return ins(stats, nil, nt.Clone(), types.CmpUint256, interpUint256)

	case types.UUID:

		return ins(stats, nil, nt, types.CmpUUID, interpUUID)

	case decimal.DecimalArray: // TODO
	case types.Uint256Array: // TODO
	case []string:
	case []int64:

	default:
		return fmt.Errorf("unrecognized tuple column type %T", val)
	}

	fmt.Printf("unhandled %T", val)

	return nil // known type, just no stats handling
}

func upColStatsWithDelete(stats *sql.ColumnStatistics, old any) error {
	// DELETE:
	// - unset min max if removing it. reference mincount and maxcount to know
	// - update null count
	// - adjust mcvs / histogram

	if old == nil {
		stats.NullCount--
		return nil
	}

	switch nt := old.(type) {
	case int64:
		del(stats, nt, cmp.Compare[int64])
	case string:
		del(stats, nt, strings.Compare)
	case float64:
		del(stats, nt, cmp.Compare[float64])
	case bool:
		del(stats, nt, cmpBool)
	case []byte:
		del(stats, nt, bytes.Compare)

	case *decimal.Decimal:
		if nt.NaN() {
			stats.NullCount--
			return nil // ignore, don't put in stats
		}
		nt2 := *nt

		del(stats, &nt2, cmpDecimal)

	case *types.Uint256:
		if nt.Null {
			stats.NullCount--
			return nil
		}

		return del(stats, nt.Clone(), types.CmpUint256)

	case types.UUID:

		return del(stats, nt, types.CmpUUID)

	case decimal.DecimalArray: // TODO
	case types.Uint256Array: // TODO
	case []string:
	case []int64:

	default:
		return fmt.Errorf("unrecognized tuple column type %T", old)
	}

	return nil
}

func upColStatsWithUpdate(stats *sql.ColumnStatistics, old, up any) error { //nolint:unused
	if compareStatsVal(old, up) == 0 {
		// With replica identity full, any update to the row creates a tuple
		// update for the full row. Ignore unchanged columns.
		return nil
	}
	// update may or may not affect null count
	err := upColStatsWithDelete(stats, old)
	if err != nil {
		return err
	}
	return upColStatsWithInsert(stats, up)
}

// The following are the generic functions that handle the re-typed and non-NULL
// values from the upColStatsWith* functions above.

// insMCVs attempts to insert a new value into the MCV set. If the set already
// includes the value, its frequency is incremented. If the value is new and the
// set is not yet at capacity, it is inserted at the appropriate location in the
// slices to keep them sorted by value The sorting is needed later when
// computing cumulative frequency with inequality conditions, and it allows
// locating known values in log time.
func insMCVs[T any](vals []any, freqs []int, val T, comp func(a, b T) int) ([]any, bool, []int) {
	var spill bool
	// sort.Search is much harder to use than slices.BinarySearchFunc but here we are
	loc := sort.Search(len(vals), func(i int) bool {
		v := vals[i].(T)
		return comp(v, val) != -1 // v[i] >= val
	})
	found := loc != len(vals) && comp(vals[loc].(T), val) == 0
	if found {
		if comp(vals[loc].(T), val) != 0 {
			panic("wrong loc")
		}
		freqs[loc]++
	} else if len(vals) < statsCap {
		vals = slices.Insert(vals, loc, any(val))
		freqs = slices.Insert(freqs, loc, 1)
	} else {
		spill = true
	}
	return vals, spill, freqs
}

// this is SOOO much easier with vals as a []T rather than []any.
// alas, that created pains in other places... might switch back.

/*func insMCVs[T any](vals []T, freqs []int, val T, comp func(a, b T) int) ([]T, []int) {
	loc, found := slices.BinarySearchFunc(vals, val, comp)
	if found {
		freqs[loc]++
	} else if len(vals) < statsCap {
		vals = slices.Insert(vals, loc, val)
		freqs = slices.Insert(freqs, loc, 1)
	}
	return vals, freqs
}*/

// ins is used to insert a non-NULL value, but it can be used in different contexts:
// (1) performing a full scan, where mcvSpill may be non-nil on a second pass,
// (2) maintaining stats estimate on an insert. del and update only apply in the
// latter context.
func ins[T any](stats, prev *sql.ColumnStatistics, val T, comp func(v, m T) int,
	interp func(f float64, a, b T) T) error {

	switch mn := stats.Min.(type) {
	case nil: // first observation
		stats.Min = val
		stats.MinCount = 1
	case T:
		switch comp(val, mn) {
		case -1: // new MINimum
			stats.Min = val
			stats.MinCount = 1
		case 0: // another of the same
			stats.MinCount++
		}
	case unknown: // it was deleted, only full (re)scan can figure it out
	default:
		return fmt.Errorf("invalid stats value type %T for tuple of type %T", val, stats.Min)
	}

	switch mn := stats.Max.(type) {
	case nil: // first observation (also would have set Min above)
		stats.Max = val
		stats.MaxCount = 1
	case T:
		switch comp(val, mn) {
		case 1: // new MAXimum
			stats.Max = val
			stats.MaxCount = 1
		case 0: // another of the same
			stats.MaxCount++
		}
	case unknown: // it was deleted, only full (re)scan can figure it out
	default:
		return fmt.Errorf("invalid stats value type %T for tuple of type %T", val, stats.Min)
	}

	if stats.MCVals == nil {
		stats.MCVals = []any{} // insMCVs: = []T{val} ; stats.MCFreqs = []int{1}
	}

	var missed bool
	if prev == nil || len(prev.MCVals) < statsCap { // first pass of full scan, pointless second pass, or no-context insert
		stats.MCVals, missed, stats.MCFreqs = insMCVs(stats.MCVals, stats.MCFreqs, val, comp)
		// vals, freqs, _ := insMCVs(convSlice[T](stats.MCVals), stats.MCFreqs, val, comp)
		// stats.MCVals, stats.MCFreqs = convSlice2(vals), freqs
	} else { // a second pass in a full scan
		// The freqs for MCVals are complete, representing the entire table.
		// Fill the spills struct instead, but EXCLUDING vals in MCVals,
		// allowing possibly higher counts in later rows. Caller should merge
		// back to the MCVals/MCFreqs after the complete second pass.

		// When MCVals was an any (underlying []T) rather than a []any, we were
		// able to simply assert to []T, but I switched to []any for other reasons.
		// _, found := slices.BinarySearchFunc(stats.MCVals.([]T), val, comp)

		loc := sort.Search(len(prev.MCVals), func(i int) bool {
			v := prev.MCVals[i].(T)
			return comp(v, val) != -1 // v[i] >= val
		})
		found := loc != len(prev.MCVals) && comp(prev.MCVals[loc].(T), val) == 0
		// _, found := slices.BinarySearchFunc(convSlice[T](prev.MCVals), val, comp)
		if !found { // not in previous scan
			stats.MCVals, missed, stats.MCFreqs = insMCVs(stats.MCVals, stats.MCFreqs, val, comp)
		} // else ignore this value, already counted in prev pass
	}

	// If the value was not included in the MCVs, it spills into the histogram.
	if missed {
		// we create the histogram only *after* MCVs have been collected so that
		// the bounds can be chosen based on at least some observed values.
		if stats.Histogram == nil {
			var left, right T
			if mn, ok := stats.Min.(T); ok {
				left = mn
			} else {
				left = stats.MCVals[0].(T)
			}
			if mn, ok := stats.Max.(T); ok {
				right = mn
			} else {
				right = stats.MCVals[len(stats.MCVals)-1].(T)
			}
			bounds := makeBounds(statsCap, left, right, comp, interp)
			stats.Histogram = makeHisto(bounds, comp)
		}

		h := stats.Histogram.(histo[T])
		h.ins(val)
	}

	return nil
}

// If we've eliminated all of a value via delete, then the real Min/Max is
// unknown, and it just cannot be used to affect selectivity. Further inserts
// also cannot be compared with an unknown min/max. Rescan is needed to identify
// the actual value. This type signals this case.
type unknown struct{}

// del is used to delete a non-NULL value.
func del[T any](stats *sql.ColumnStatistics, val T, comp func(v, m T) int) error {

	switch mn := stats.Min.(type) {
	case nil: // should not happen
		return errors.New("nil Min on del")
	case T:
		if comp(val, mn) == 0 {
			stats.MinCount--
			if stats.MinCount == 0 {
				stats.Min = unknown{}
			}
		}
	case unknown: // it was deleted, only full (re)scan can figure it out
	default:
		return fmt.Errorf("invalid stats value type %T for tuple of type %T", val, stats.Min)
	}

	switch mn := stats.Max.(type) {
	case nil: // should not happen
		return errors.New("nil Max on del")
	case T:
		if comp(val, mn) == 0 {
			stats.MaxCount--
			if stats.MaxCount == 0 {
				stats.Max = unknown{}
			}
		}
	case unknown: // it was deleted, only full (re)scan can figure it out
	default:
		return fmt.Errorf("invalid stats value type %T for tuple of type %T", val, stats.Max)
	}

	// Look for it in the MCVs
	loc := sort.Search(len(stats.MCVals), func(i int) bool {
		v := stats.MCVals[i].(T)
		return comp(v, val) != -1 // v[i] >= val
	})
	found := loc != len(stats.MCVals) && comp(stats.MCVals[loc].(T), val) == 0
	// loc, found := slices.BinarySearchFunc(convSlice[T](stats.MCVals), val, comp)
	if found {
		if stats.MCFreqs[loc] == 1 {
			stats.MCVals = slices.Delete(stats.MCVals, loc, loc+1)
			stats.MCFreqs = slices.Delete(stats.MCFreqs, loc, loc+1)
		} else {
			stats.MCFreqs[loc]--
		}
	} else {
		// adjust histogram freq--
		hist, ok := stats.Histogram.(histo[T])
		if !ok {
			fmt.Println("mcvs:", convSliceAsserted[T](stats.MCVals))
			fmt.Println("val:", val)
			panic("nil histogram on delete but mcv missed!")
		}
		hist.rm(val)
	}

	return nil
}
