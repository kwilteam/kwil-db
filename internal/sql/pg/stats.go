package pg

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math/big"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/kwilteam/kwil-db/core/types/decimal"
	costtypes "github.com/kwilteam/kwil-db/internal/engine/cost/datatypes"

	"github.com/kwilteam/kwil-db/common/sql"
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
	colStats := make([]costtypes.ColumnStatistics, numCols)

	// iterate over all rows (select *)
	// var scans []any
	// for _, col := range colInfo {
	// 	scans = append(scans, col.ScanVal()) // for QueryRowFunc
	// }
	err = QueryRowFuncAny(ctx, db, `SELECT * FROM `+qualifiedTable,
		func(_ []FieldDesc, vals []any) error {
			for i, val := range vals {
				stat := &colStats[i]
				if val == nil {
					stat.NullCount++
					continue
				}

				switch colTypes[i] {
				case ColTypeInt:
					valInt, ok := sql.Int64(val)
					if !ok {
						return errors.New("not int")
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
					valStr, ok := TextValue(val) // val.(string)
					if !ok {
						return errors.New("not string")
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
					valBytea, ok := val.([]byte)
					if !ok {
						return errors.New("not bytea")
					}
					if stat.Min == nil {
						stat.Min = valBytea
						stat.Max = valBytea
						continue
					}

					if bytes.Compare(valBytea, stat.Min.([]byte)) == -1 {
						stat.Min = valBytea
					} else if bytes.Compare(valBytea, stat.Max.([]byte)) == 1 {
						stat.Max = valBytea
					}
					continue

				case ColTypeNumeric:
					var dec *decimal.Decimal
					switch v := val.(type) {
					case pgtype.Numeric:
						if v.NaN {
							continue
						}

						dec, err = pgNumericToDecimal(v)
						if err != nil {
							continue
						}
						// fmt.Println("A", dec)

					case decimal.Decimal:
						v2 := v
						dec = &v2
						// fmt.Println("B", dec)
					}

					fmt.Println(dec)

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
					case float32:
						varFloat = float64(v)
					case float64:
						varFloat = v
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
