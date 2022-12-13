package schema

import (
	"fmt"

	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
)

type InsertDef struct {
	Name    string
	Table   string
	Columns ColumnMap
}

func (q *InsertDef) Type() QueryType {
	return Create
}

func (q *InsertDef) Prepare(db *Database) (*ExecutableQuery, error) {
	// Create a preparedStatement value and initialize its fields
	tbl, ok := db.Tables[q.Table]
	if !ok {
		return nil, fmt.Errorf("table %s not found", q.Table)
	}
	qry := ExecutableQuery{
		Statement:  "",
		Args:       make(map[int]arg),
		UserInputs: make([]*requiredInput, 0),
	}

	// Create a statementInput value for each column and add it to the preparedStatement's inputs slice
	cols := []interface{}{}
	defaultCols := []interface{}{} // I define this separately since we want all default to be at the end
	i := 1
	for name, val := range q.Columns {

		fillable := false
		if val == "" {
			fillable = true
			qry.UserInputs = append(qry.UserInputs, &requiredInput{
				Column: name,
				Type:   tbl.Columns[name].Type.String(),
			})
			cols = append(cols, name)
		} else {
			defaultCols = append(defaultCols, name)
		}

		pgType, ok := Types[tbl.Columns[name].Type]
		if !ok {
			return nil, fmt.Errorf("unsupported type: %s", tbl.Columns[name].Type)
		}

		qry.Args[i] = arg{
			Column:   name,
			Default:  val,
			Type:     pgType.String(),
			Fillable: fillable,
		}
		i++
	}

	cols = append(cols, defaultCols...)
	// Create the SQL statement
	stmt, err := InsertBuilder(db.SchemaName() + "." + q.Table).Columns(cols).ToSQL()
	if err != nil {
		return nil, err
	}

	qry.Statement = stmt

	// Return a pointer to the preparedStatement value
	return &qry, nil
}

type insertBuilder struct {
	stmt *goqu.InsertDataset
}

func InsertBuilder(table string) *insertBuilder {
	return &insertBuilder{
		stmt: goqu.Dialect("postgres").Insert(table).Prepared(true),
	}
}

func (b *insertBuilder) Columns(cols []interface{}) *insertBuilder {
	var vals goqu.Vals
	for i := range cols {
		vals = append(vals, i)
	}

	b.stmt = b.stmt.Cols(cols...).Vals(vals)
	return b
}

func (b *insertBuilder) ToSQL() (string, error) {
	stmt, _, err := b.stmt.ToSQL()
	return stmt, err
}
