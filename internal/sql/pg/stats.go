package pg

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/kwilteam/kwil-db/common/sql"
	"github.com/kwilteam/kwil-db/core/types/decimal"
)

// RowCounter represents a type that is able to return an exact row count for a
// given table. This is primarily used for testing with a mock DB.
type RowCounter interface {
	RowCount(ctx context.Context, qualifiedTable string) (int64, error)
}

// RowCount gets a precise row count for the named fully qualified table. If the
// Executor satisfies the RowCounter interface, that method will be used
// directly. Otherwise a simple select query is used.
func RowCount(ctx context.Context, qualifiedTable string, db sql.Executor) (int64, error) {
	if rc, ok := db.(RowCounter); ok {
		return rc.RowCount(ctx, qualifiedTable)
	}

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

// TableStats collects deterministic statistics for a table. If schema is empty,
// the "public" schema is assumed. This method is used to obtain the ground
// truth statistics for a table; incremental statistics updates should be
// preferred when possible.
func TableStats(ctx context.Context, schema, table string, db sql.Executor) (*sql.Statistics, error) {
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

type ColStatser interface {
	ColStats(ctx context.Context, qualifiedTable string, colInfo []ColInfo) ([]sql.ColumnStatistics, error)
}

// colStats collects column-wise statistics for the specified table, using the
// provided column definitions to instantiate scan values used by the full scan
// that iterates over all rows of the table.
func colStats(ctx context.Context, qualifiedTable string, colInfo []ColInfo, db sql.Executor) ([]sql.ColumnStatistics, error) {
	if cs, ok := db.(ColStatser); ok {
		return cs.ColStats(ctx, qualifiedTable, colInfo)
	}

	numCols := len(colInfo)
	colTypes := make([]ColType, numCols)
	for i := range colInfo {
		colTypes[i] = colInfo[i].Type()
	}

	colStats := make([]sql.ColumnStatistics, numCols)

	// iterate over all rows (select *)
	var scans []any
	for _, col := range colInfo {
		scans = append(scans, col.ScanVal()) // for QueryRowFunc
	}
	// IMPORTANT NOTE: the following iteration over all rows of the table
	// involves *no* ORDER BY clause. As such, the scan order is not guaranteed
	// to be deterministic, and the any aggregation code should be commutative.
	// For example, we can't naively perform summation of float64 (double
	// precision floating point), but we can with integer or NUMERIC types,
	// issues of overflow aside.
	err := QueryRowFunc(ctx, db, `SELECT * FROM `+qualifiedTable, scans,
		// func(_ []FieldDesc, vals []any) error { // for QueryRowFuncAny
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
				case ColTypeInt:
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

					if stat.Min == nil {
						stat.Min = valInt
						stat.Max = valInt
						continue
					}
					if valInt < stat.Min.(int64) {
						stat.Min = valInt
					} else if valInt > stat.Max.(int64) {
						stat.Max = valInt
					}
					continue

				case ColTypeText:
					valStr, null, ok := TextValue(val) // val.(string)
					if !ok {
						return fmt.Errorf("not string: %T", val)
					}
					if null {
						stat.NullCount++
						continue
					}
					if stat.Min == nil {
						stat.Min = valStr
						stat.Max = valStr
						continue
					}
					if valStr < stat.Min.(string) {
						stat.Min = valStr
					} else if valStr > stat.Max.(string) {
						stat.Max = valStr
					}
					continue

				case ColTypeByteA:
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
						valBytea = *vt
					case []byte:
						if vt == nil {
							stat.NullCount++
							continue
						}
						valBytea = vt
					default:
						return fmt.Errorf("not bytea: %T", val)
					}

					if stat.Min == nil {
						valBytea = slices.Clone(valBytea)
						stat.Min = valBytea
						stat.Max = valBytea
						continue
					}

					if bytes.Compare(valBytea, stat.Min.([]byte)) == -1 {
						stat.Min = slices.Clone(valBytea)
					} else if bytes.Compare(valBytea, stat.Max.([]byte)) == 1 {
						stat.Max = slices.Clone(valBytea)
					}
					continue

				case ColTypeBool:
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

					if stat.Min == nil {
						stat.Min = b
						stat.Max = b
						continue
					}

					if b && !stat.Max.(bool) {
						stat.Max = b // true
					}
					if !b && stat.Min.(bool) {
						stat.Min = b // false
					}

				case ColTypeNumeric:
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
						dec = v
					case decimal.Decimal:
						v2 := v
						dec = &v2
					}

					if stat.Min == nil {
						stat.Min = dec
						stat.Max = dec
						continue
					}

					// we may need to worry about NaNs here, not sure
					cm, _ := dec.Cmp(stat.Min.(*decimal.Decimal))
					if cm == -1 {
						stat.Min = dec
						continue
					}
					cm, _ = dec.Cmp(stat.Max.(*decimal.Decimal))
					if cm == 1 {
						stat.Max = dec
					}

					continue

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

					if stat.Min == nil {
						stat.Min = varFloat
						stat.Max = varFloat
						continue
					}
					if varFloat < stat.Min.(float64) {
						stat.Min = varFloat
					} else if varFloat > stat.Max.(float64) {
						stat.Max = varFloat
					}

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

	return colStats, nil
}
