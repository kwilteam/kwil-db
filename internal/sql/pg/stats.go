package pg

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math/big"
	"slices"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/kwilteam/kwil-db/common/sql"
	"github.com/kwilteam/kwil-db/core/types/decimal"
	costtypes "github.com/kwilteam/kwil-db/internal/engine/cost/datatypes"
)

func TableStats(ctx context.Context, table string, db sql.Executor) (*costtypes.Statistics, error) {
	// table stats:
	//  1. row count
	//  2. per-column stats
	//		a. min and max
	//		b. null count
	//		c. unique value count ?
	//		d. average record size ?
	//		e. ???

	qualifiedTable := table // pgSchema + "." + table

	// row count
	res, err := db.Execute(ctx, fmt.Sprintf(`SELECT count(*) FROM %s`, qualifiedTable))
	if err != nil {
		return nil, err
	}
	count, ok := sql.Int64(res.Rows[0][0])
	if !ok {
		return nil, fmt.Errorf("no row count for %s", qualifiedTable)
	}
	// TODO: We needs a schema-table stats database so we don't ever have to do
	// a full table scan for column stats.

	colInfo, err := ColumnInfo(ctx, db, qualifiedTable)
	if err != nil {
		return nil, err
	}
	numCols := len(colInfo)
	colTypes := make([]ColType, numCols)
	for i := range colInfo {
		colTypes[i] = colInfo[i].Type()
	}
	// fmt.Println(colTypes)
	colStats := make([]costtypes.ColumnStatistics, numCols)

	// iterate over all rows (select *)
	var scans []any
	for _, col := range colInfo {
		scans = append(scans, col.ScanVal()) // for QueryRowFunc
	}
	err = QueryRowFunc(ctx, db, `SELECT * FROM `+qualifiedTable, scans,
		// func(_ []FieldDesc, vals []any) error { // for QueryRowFuncAny
		func() error {
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
					valStr, invalid, ok := TextValue(val) // val.(string)
					if !ok {
						return fmt.Errorf("not string: %T", val)
					}
					if invalid {
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
						// fmt.Println("A", dec)

					case *decimal.Decimal:
						dec = v
					case decimal.Decimal:
						v2 := v
						dec = &v2
						// fmt.Println("B", dec)
					}

					// fmt.Println(dec)

					if stat.Min == nil {
						stat.Min = dec
						stat.Max = dec
						continue
					}

					// we may need to worry about NaNs here, not sure
					cm, _ := dec.Cmp(stat.Min.(*decimal.Decimal))
					if cm == -1 {
						stat.Min = dec
						fmt.Println("min", dec, stat.Min)
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
					fallthrough
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

	return &costtypes.Statistics{
		RowCount:         count,
		ColumnStatistics: colStats,
	}, nil
}

var ErrNaN = errors.New("NaN")

func pgNumericToDecimal(num pgtype.Numeric) (*decimal.Decimal, error) {
	if num.NaN {
		return nil, ErrNaN
	}

	i, e := num.Int, num.Exp

	// Kwil's decimal semantics do not allow negative scale (only shift decimal
	// left), so if the exponent is positive we need to apply it to the integer.
	if e > 0 {
		// i * 10^e
		z := new(big.Int)
		z.Exp(big.NewInt(10), big.NewInt(int64(e)), nil)
		z.Mul(z, i)
		i, e = z, 0
	}

	// Really this could be uint256, which is same underlying type (a domain) as
	// Numeric. If the caller needs to know, that has to happen differently.
	return decimal.NewFromBigInt(i, e)
}
