package schema

import (
	"fmt"

	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
)

type DeleteDef struct {
	name  string
	table string
	where []where_predicate
}

func (q *DeleteDef) Where() []where_predicate {
	return q.where
}

func (q *DeleteDef) Name() string {
	return q.name
}

func (q *DeleteDef) Type() QueryType {
	return Delete
}

func (q *DeleteDef) Prepare(db *Database) (*executableQuery, error) {
	// Create a preparedStatement value and initialize its fields
	tbl, ok := db.Tables[q.table]
	if !ok {
		return nil, fmt.Errorf("table %s not found", q.table)
	}

	qry := executableQuery{
		Statement:  "",
		Args:       make(map[int]arg),
		UserInputs: make([]requiredInputs, 0),
	}
	statement := DeleteBuilder(db.SchemaName() + "." + q.table)
	i := 1
	for _, where := range q.where {
		statement.Where(&where)

		fillable := false
		if where.Default == "" {
			fillable = true

			qry.UserInputs = append(qry.UserInputs, requiredInputs{
				Column: where.Column,
				Type:   tbl.Columns[where.Column].Type.String(),
			})
		}

		pgType, ok := Types[tbl.Columns[where.Column].Type]
		if !ok {
			return nil, fmt.Errorf("unsupported type: %s", tbl.Columns[where.Column].Type)
		}

		qry.Args[i] = arg{
			Column:   where.Column,
			Default:  where.Default,
			Type:     pgType.String(),
			Fillable: fillable,
		}

		i++
	}

	stmt, err := statement.ToSQL()
	if err != nil {
		return nil, err
	}

	qry.Statement = stmt

	return &qry, nil
}

type deleteBuilder struct {
	stmt   *goqu.DeleteDataset
	wheres []goqu.Expression
}

func DeleteBuilder(table string) *deleteBuilder {
	return &deleteBuilder{
		stmt: goqu.Dialect("postgres").Delete(table).Prepared((true)),
	}
}

func (b *deleteBuilder) Where(where *where_predicate) *deleteBuilder {
	gexp := WhereConversions.StringToPredicate(where.Operator, where.Column)
	if gexp == nil {
		return b
	}
	b.wheres = append(b.wheres, gexp) // passing empty interface doesn't work with goqu here for some reason so I pass empty string
	return b
}

func (b *deleteBuilder) ToSQL() (string, error) {
	b.stmt = b.stmt.Where(b.wheres...)
	stmt, _, err := b.stmt.ToSQL()
	return stmt, err
}
