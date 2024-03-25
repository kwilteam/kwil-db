package tree

import sqlwriter "github.com/kwilteam/kwil-db/parse/sql/tree/sql-writer"

type QualifiedTableName struct {
	node

	schema     string
	TableName  string
	TableAlias string
}

func (q *QualifiedTableName) Accept(v AstVisitor) any {
	return v.VisitQualifiedTableName(q)
}

func (q *QualifiedTableName) Walk(w AstListener) error {
	return run(
		w.EnterQualifiedTableName(q),
		w.ExitQualifiedTableName(q),
	)
}

func (q *QualifiedTableName) ToSQL() string {
	q.check()

	stmt := sqlwriter.NewWriter()

	if q.schema != "" {
		stmt.Token.Space()
		stmt.WriteIdentNoSpace(q.schema)
		stmt.Token.Period()
		stmt.WriteIdentNoSpace(q.TableName)
		stmt.Token.Space()
	} else {
		stmt.WriteIdent(q.TableName)
	}

	if q.TableAlias != "" {
		stmt.Token.As()
		stmt.WriteIdent(q.TableAlias)
	}

	return stmt.String()
}

func (q *QualifiedTableName) check() {
	if q.TableName == "" {
		panic("table name is empty")
	}
}

// SetSchema sets the schema of the table.
// It should not be called by the parser, and is meant to be called
// by processes after parsing.
func (q *QualifiedTableName) SetSchema(schema string) {
	q.schema = schema
}
