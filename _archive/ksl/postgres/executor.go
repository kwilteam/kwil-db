package postgres

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"ksl"
	query "ksl/models"
	"ksl/sqldriver"
	"ksl/sqlmigrate"
	"ksl/sqlschema"
	"reflect"
	"strings"

	"github.com/cstockton/go-conv"
	"github.com/doug-martin/goqu/v9"
	"golang.org/x/exp/slices"
)

var (
	ErrInvalidQuery = errors.New("invalid query")
)

type Executor struct {
	sqlmigrate.Describer
	Conn sqldriver.ExecQuerier
}

func (c *Executor) ExecuteInsert(ctx context.Context, stmt sqldriver.InsertStatement) error {
	db, err := c.Describe(stmt.Database)
	if err != nil {
		return err
	}

	table, ok := db.FindTable(stmt.Table).Get()
	if !ok {
		return fmt.Errorf("unknown table %q", stmt.Table)
	}

	resolved := make(goqu.Record)
	for _, column := range table.Columns() {
		input, ok := stmt.Input[column.Name()]
		if !ok {
			if column.IsRequired() {
				return fmt.Errorf("missing required column %q", column.Name())
			}
			continue
		}
		value, err := convert(db, column.Name(), input, column.Type(), column.Arity())
		if err != nil {
			return err
		}
		resolved[column.Name()] = value
	}

	sql, args, e := goqu.Dialect("postgres").Insert(table.Name()).Rows(resolved).ToSQL()
	if e != nil {
		return e
	}
	_, err = c.Conn.ExecContext(ctx, sql, args...)

	return err
}

func (c *Executor) ExecuteUpdate(ctx context.Context, stmt sqldriver.UpdateStatement) error {
	db, err := c.Describe(stmt.Database)
	if err != nil {
		return err
	}

	table, ok := db.FindTable(stmt.Table).Get()
	if !ok {
		return fmt.Errorf("unknown table %q", stmt.Table)
	}

	pk, ok := table.PrimaryKey().Get()
	if !ok {
		return fmt.Errorf("table %q has no primary key", stmt.Table)
	}

	expr := make(goqu.Ex)
	for _, field := range pk.Columns() {
		column := field.Column()

		input, ok := stmt.Where[column.Name()]
		if !ok {
			return fmt.Errorf("missing required column %q in where filter", column.Name())
		}
		value, err := convert(db, column.Name(), input, column.Type(), column.Arity())
		if err != nil {
			return err
		}
		expr[column.Name()] = value
	}

	record := make(goqu.Record)
	for _, column := range table.Columns() {
		if pk.ContainsColumn(column.ID) {
			continue
		}

		input, ok := stmt.Input[column.Name()]
		if !ok {
			continue
		}

		value, err := convert(db, column.Name(), input, column.Type(), column.Arity())
		if err != nil {
			return err
		}
		record[column.Name()] = value
	}

	sql, args, err := goqu.Update(table.Name()).Set(record).Where(expr).ToSQL()
	if err != nil {
		return err
	}
	_, err = c.Conn.ExecContext(ctx, sql, args...)
	return err
}

func (c *Executor) ExecuteDelete(ctx context.Context, stmt sqldriver.DeleteStatement) error {
	db, err := c.Describe(stmt.Database)
	if err != nil {
		return err
	}

	table, ok := db.FindTable(stmt.Table).Get()
	if !ok {
		return fmt.Errorf("unknown table %q", stmt.Table)
	}

	pk, ok := table.PrimaryKey().Get()
	if !ok {
		return fmt.Errorf("table %q has no primary key", stmt.Table)
	}

	expr := make(goqu.Ex)
	for _, field := range pk.Columns() {
		column := field.Column()

		input, ok := stmt.Where[column.Name()]
		if !ok {
			return fmt.Errorf("missing required column %q in where filter", column.Name())
		}
		value, err := convert(db, column.Name(), input, column.Type(), column.Arity())
		if err != nil {
			return err
		}
		expr[column.Name()] = value
	}
	sql, args, err := goqu.Dialect("postgres").Delete(table.Name()).Where(expr).ToSQL()
	if err != nil {
		return err
	}
	_, err = c.Conn.ExecContext(ctx, sql, args...)
	return err
}

func (c *Executor) ExecuteSelect(ctx context.Context, stmt sqldriver.SelectStatement) (*query.Result, error) {
	db, err := c.Describe(stmt.Database)
	if err != nil {
		return nil, err
	}

	table, ok := db.FindTable(stmt.Table).Get()
	if !ok {
		return nil, fmt.Errorf("unknown table %q", stmt.Table)
	}

	expr := make(goqu.Ex)
	for name, input := range stmt.Where {
		column, ok := table.Column(name).Get()
		if !ok {
			return nil, fmt.Errorf("unknown column %q", name)
		}
		value, err := convert(db, column.Name(), input, column.Type(), column.Arity())
		if err != nil {
			return nil, err
		}
		expr[name] = value
	}

	columnNames := make([]any, len(table.Columns()))
	for i, column := range table.Columns() {
		columnNames[i] = column.Name()
	}

	sql, args, err := goqu.Dialect("postgres").Select(columnNames...).From(table.Name()).Where(expr).ToSQL()
	if err != nil {
		return nil, err
	}
	rows, err := c.Conn.QueryContext(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results query.Result

	err = results.Load(rows)
	if err != nil {
		return nil, err
	}

	return &results, nil

	/*

		for rows.Next() {
			columns := make([]any, len(columnNames))
			columnPointers := make([]any, len(columnNames))
			for i := range columns {
				columnPointers[i] = &columns[i]
			}
			if err := rows.Scan(columnPointers...); err != nil {
				return nil, err
			}
			m := make(map[string]string)
			for i, colName := range columnNames {
				val := columnPointers[i].(*string)
				m[colName.(string)] = *val
			}
			results = append(results, m)
		}
		return results, err*/
}

func convert(db sqlschema.Database, name string, input any, typ sqlschema.ColumnType, arity sqlschema.ColumnArity) (any, error) {
	if arity == sqlschema.List {
		switch v := input.(type) {
		case string:
			var list []any
			if err := json.Unmarshal([]byte(v), &list); err != nil {
				return nil, fmt.Errorf("invalid list value: %w", err)
			}
			input = list
		}

		rv := reflect.ValueOf(input)
		if rv.Kind() != reflect.Slice {
			return nil, fmt.Errorf("invalid list value: %T", input)
		}
		var values []any
		for i := 0; i < rv.Len(); i++ {
			v, err := convert(db, name, rv.Index(i).Interface(), typ, sqlschema.Required)
			if err != nil {
				return nil, err
			}
			values = append(values, v)
		}
		return values, nil
	}

	switch typ := typ.Type.(type) {
	case ksl.BuiltInScalar:
		switch typ {
		case ksl.BuiltIns.Int:
			return conv.Int(input)
		case ksl.BuiltIns.BigInt:
			return conv.Int64(input)
		case ksl.BuiltIns.Float:
			return conv.Float64(input)
		case ksl.BuiltIns.String:
			return conv.String(input)
		case ksl.BuiltIns.Bool:
			return conv.Bool(input)
		case ksl.BuiltIns.Date:
			return conv.Time(input)
		case ksl.BuiltIns.DateTime:
			return conv.Time(input)
		case ksl.BuiltIns.Time:
			return conv.Time(input)
		case ksl.BuiltIns.Bytes:
			str, err := conv.String(input)
			if err != nil {
				return nil, err
			}
			return base64.RawStdEncoding.DecodeString(str)
		case ksl.BuiltIns.Decimal:
			return conv.Float64(input)
		}
	case sqlschema.EnumType:
		enum, ok := db.FindEnum(typ.Name).Get()
		if !ok {
			return nil, fmt.Errorf("unknown enum %q", typ.Name)
		}

		value, err := conv.String(input)
		if err != nil {
			return nil, err
		}

		if !slices.Contains(enum.Values(), value) {
			return nil, fmt.Errorf("invalid value for argument %q, expected one of %s: ", name, strings.Join(enum.Values(), ", "))
		}
		return value, nil
	}

	return conv.String(input)
}
