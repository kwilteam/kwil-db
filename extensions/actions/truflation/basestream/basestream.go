// package basestream implements the base stream extension.
// it is meant to be used for a Truflation primitive stream
// that tracks some time series data.
package basestream

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/kwilteam/kwil-db/internal/engine/execution"
	"github.com/kwilteam/kwil-db/internal/engine/types"
	"github.com/kwilteam/kwil-db/internal/sql"
)

// InitializeBasestream initializes the basestream extension.
// It takes 3 configs: table, date_column, and value_column.
// The table is the table that the data is stored in.
// The date_column is the column that the date is stored in, stored as "YYYY-MM-DD".
// The value_column is the column that the value is stored in. It must be an integer.
func InitializeBasestream(ctx *execution.DeploymentContext, metadata map[string]string) (execution.ExtensionNamespace, error) {
	var table, dateColumn, valueColumn string
	var ok bool
	table, ok = metadata["table_name"]
	if !ok {
		return nil, errors.New("missing table config")
	}
	dateColumn, ok = metadata["date_column"]
	if !ok {
		return nil, errors.New("missing date_column config")
	}
	valueColumn, ok = metadata["value_column"]
	if !ok {
		return nil, errors.New("missing value_column config")
	}

	foundTable := false
	foundDateColumn := false
	foundValueColumn := false
	// now we validate that the table and columns exist
	for _, tbl := range ctx.Schema.Tables {
		if strings.EqualFold(tbl.Name, table) {
			foundTable = true
			for _, col := range tbl.Columns {
				if strings.EqualFold(col.Name, dateColumn) {
					foundDateColumn = true
					if col.Type != types.TEXT {
						return nil, fmt.Errorf("date column %s must be of type TEXT", dateColumn)
					}
				}
				if strings.EqualFold(col.Name, valueColumn) {
					foundValueColumn = true
					if col.Type != types.INT {
						return nil, fmt.Errorf("value column %s must be of type INTEGER", valueColumn)
					}
				}
			}
		}
	}

	if !foundTable {
		return nil, fmt.Errorf("table %s not found", table)
	}
	if !foundDateColumn {
		return nil, fmt.Errorf("date column %s not found", dateColumn)
	}
	if !foundValueColumn {
		return nil, fmt.Errorf("value column %s not found", valueColumn)
	}

	return &BaseStreamExt{
		table:       table,
		dateColumn:  dateColumn,
		valueColumn: valueColumn,
	}, nil
}

var _ = execution.ExtensionInitializer(InitializeBasestream)

type BaseStreamExt struct {
	table       string
	dateColumn  string
	valueColumn string
}

func (b *BaseStreamExt) Call(scope *execution.ProcedureContext, method string, args []any) ([]any, error) {
	switch strings.ToLower(method) {
	default:
		return nil, fmt.Errorf("unknown method: %s", method)
	case "get_index":
		return getValue(scope, b.index, args...)
	case "get_value":
		return getValue(scope, b.value, args...)
	}
}

const (
	// getBaseValue gets the base value from a base stream, to be used in index calculation.
	sqlGetBaseValue     = `select %s from %s order by %s ASC LIMIT 1;`
	sqlGetLatestValue   = `select %s from %s order by %s DESC LIMIT 1;`
	sqlGetSpecificValue = `select %s from %s where %s = $date;`
	zeroDate            = "0000-00-00"
)

func (b *BaseStreamExt) sqlGetBaseValue() string {
	return fmt.Sprintf(sqlGetBaseValue, b.valueColumn, b.table, b.dateColumn)
}

func (b *BaseStreamExt) sqlGetLatestValue() string {
	return fmt.Sprintf(sqlGetLatestValue, b.valueColumn, b.table, b.dateColumn)
}

func (b *BaseStreamExt) sqlGetSpecificValue() string {
	return fmt.Sprintf(sqlGetSpecificValue, b.valueColumn, b.table, b.dateColumn)
}

// getValue gets the value for the specified function.
func getValue(scope *execution.ProcedureContext, fn func(context.Context, Querier, string) (int64, error), args ...any) ([]any, error) {
	dataset, err := scope.Dataset(scope.DBID)
	if err != nil {
		return nil, err
	}

	if len(args) != 1 {
		return nil, fmt.Errorf("expected one argument")
	}

	date, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("expected string for date argument")
	}

	val, err := fn(scope.Ctx, dataset, date)
	if err != nil {
		return nil, err
	}

	return []any{val}, nil
}

// index returns the inflation index for a given date.
// this follows Truflation function of ((current_value/first_value)*100).
// It will multiplty the returned result by an additional 1000, since Kwil
// cannot handle decimals.
func (b *BaseStreamExt) index(ctx context.Context, dataset Querier, date string) (int64, error) {

	// we will first get the first ever value
	res, err := dataset.Query(ctx, b.sqlGetBaseValue(), nil)
	if err != nil {
		return 0, err
	}

	scalar, err := getScalar(res)
	if err != nil {
		return 0, err
	}

	baseValue, ok := scalar.(int64)
	if !ok {
		return 0, errors.New("expected int64 for base value")
	}
	if baseValue == 0 {
		return 0, errors.New("base value cannot be zero")
	}

	// now we will get the value for the requested date
	if date == zeroDate || date == "" {
		res, err = dataset.Query(ctx, b.sqlGetLatestValue(), nil)
	} else {
		res, err = dataset.Query(ctx, b.sqlGetSpecificValue(), map[string]any{
			"$date": date,
		})
	}
	if err != nil {
		return 0, err
	}

	scalar, err = getScalar(res)
	if err != nil {
		return 0, err
	}

	currentValue, ok := scalar.(int64)
	if !ok {
		return 0, errors.New("expected int64 for current value")
	}

	// we can't do floating point division, but Truflation normally tracks
	// index precision to the thousandth, so we will multiply by 1000 before
	// performing integer division. This will round the result down (golang truncates
	// integer division results).
	// Truflations calculation is ((current_value/first_value)*100).
	// Therefore, we will alter the equation to ((current_value*100000)/first_value).
	// This essentially gives us the same result, but with an extra 3 digits of precision.
	index := (currentValue * 100000) / baseValue
	return index, nil
}

// value returns the value for a given date.
// if no date or the zero date is given, it will return the latest value.
func (b *BaseStreamExt) value(ctx context.Context, dataset Querier, date string) (int64, error) {
	var res *sql.ResultSet
	var err error
	if date == zeroDate || date == "" {
		res, err = dataset.Query(ctx, b.sqlGetLatestValue(), nil)
	} else {
		res, err = dataset.Query(ctx, b.sqlGetSpecificValue(), map[string]any{
			"$date": date,
		})
	}
	if err != nil {
		return 0, err
	}

	scalar, err := getScalar(res)
	if err != nil {
		return 0, err
	}

	value, ok := scalar.(int64)
	if !ok {
		return 0, errors.New("expected int64 for current value")
	}

	return value, nil
}

// getScalar gets a scalar value from a query result.
// It is expecting a result that has one row and one column.
// If it does not have one row and one column, it will return an error.
func getScalar(res *sql.ResultSet) (any, error) {
	if len(res.ReturnedColumns) != 1 {
		return nil, fmt.Errorf("stream expected one column, got %d", len(res.ReturnedColumns))
	}
	if len(res.Rows) == 0 {
		return nil, fmt.Errorf("stream has no data")
	}
	if len(res.Rows) != 1 {
		return nil, fmt.Errorf("stream expected one row, got %d", len(res.Rows))
	}

	return res.Rows[0][0], nil
}

type Querier interface {
	Query(ctx context.Context, stmt string, params map[string]any) (*sql.ResultSet, error)
}
