package tree

import (
	sqlwriter "github.com/kwilteam/kwil-db/internal/parse/sql/tree/sql-writer"
)

// Relation is one of:
//   - RelationTable
//   - RelationSubquery
//   - RelationJoin
type Relation interface {
	AstNode

	relation()
}

type RelationTable struct {
	node

	schema string
	Name   string
	Alias  string
}

func (t *RelationTable) Accept(v AstVisitor) any {
	return v.VisitRelationTable(t)
}

func (t *RelationTable) Walk(w AstListener) error {
	return run(
		w.EnterRelationTable(t),
		w.ExitRelationTable(t),
	)
}

func (t *RelationTable) ToSQL() string {
	if t.Name == "" {
		panic("table name is empty")
	}

	stmt := sqlwriter.NewWriter()

	if t.schema != "" {
		stmt.Token.Space()
		stmt.WriteIdentNoSpace(t.schema)
		stmt.Token.Period()
		stmt.WriteIdentNoSpace(t.Name)
		stmt.Token.Space()
	} else {
		stmt.WriteIdentNoSpace(t.Name)
	}

	if t.Alias != "" {
		stmt.Token.As()
		stmt.WriteIdentNoSpace(t.Alias)

	}

	return stmt.String()
}
func (t *RelationTable) relation() {}

// SetSchema sets the schema of the table.
// It should not be called by the parser, and is meant to be called
// by processes after parsing.
func (t *RelationTable) SetSchema(schema string) {
	t.schema = schema
}

// RelationFunc is a relation that is a function call.
// This can be used it a user has a function that returns a table.
type RelationFunction struct {
	node

	Function *ExpressionFunction
	Alias    string
}

func (t *RelationFunction) Accept(v AstVisitor) any {
	return v.VisitRelationFunction(t)
}

func (t *RelationFunction) Walk(w AstListener) error {
	return run(
		w.EnterRelationFunction(t),
		walk(w, t.Function),
		w.ExitRelationFunction(t),
	)
}

func (t *RelationFunction) relation() {}

func (t *RelationFunction) ToSQL() string {
	if t.Function == nil {
		panic("function is nil")
	}

	stmt := sqlwriter.NewWriter()

	stmt.WriteString(t.Function.ToSQL())
	if t.Alias != "" {
		stmt.Token.As()
		stmt.WriteIdent(t.Alias)

	}

	return stmt.String()
}

type RelationSubquery struct {
	node

	Select *SelectCore
	Alias  string
}

func (t *RelationSubquery) Accept(v AstVisitor) any {
	return v.VisitRelationSubquery(t)
}

func (t *RelationSubquery) Walk(w AstListener) error {
	return run(
		w.EnterRelationSubquery(t),
		walk(w, t.Select),
		w.ExitRelationSubquery(t),
	)
}

func (t *RelationSubquery) ToSQL() string {
	if t.Select == nil {
		panic("select is nil")
	}

	stmt := sqlwriter.NewWriter()

	stmt.Token.Lparen()

	selectString := t.Select.ToSQL()
	stmt.WriteString(selectString)
	stmt.Token.Rparen()

	if t.Alias != "" {
		stmt.Token.As()
		stmt.WriteIdent(t.Alias)

	}

	return stmt.String()
}
func (t *RelationSubquery) relation() {}

type RelationJoin struct {
	node

	Relation Relation
	Joins    []*JoinPredicate
}

func (t *RelationJoin) Accept(v AstVisitor) any {
	return v.VisitRelationJoin(t)
}

func (t *RelationJoin) Walk(w AstListener) error {
	return run(
		w.EnterRelationJoin(t),
		walk(w, t.Relation),
		walkMany(w, t.Joins),
		w.ExitRelationJoin(t),
	)
}

func (t *RelationJoin) relation() {}

func (t *RelationJoin) ToSQL() string {
	if t.Relation == nil {
		panic("join table or subquery cannot be nil")
	}

	stmt := sqlwriter.NewWriter()

	stmt.WriteString(t.Relation.ToSQL())
	for _, join := range t.Joins {
		stmt.WriteString(join.ToSQL())
	}

	return stmt.String()
}
