package tree

import sqlwriter "github.com/kwilteam/kwil-db/parse/sql/tree/sql-writer"

type QualifiedTableName struct {
	schema     string
	TableName  string
	TableAlias string
	IndexedBy  string
	NotIndexed bool
}

func (q *QualifiedTableName) Accept(w Walker) error {
	return run(
		w.EnterQualifiedTableName(q),
		w.ExitQualifiedTableName(q),
	)
}

func (q *QualifiedTableName) ToSQL() string {
	q.check()

	stmt := sqlwriter.NewWriter()

	if q.schema != "" {
		stmt.WriteIdentNoSpace(q.schema)
		stmt.Token.Period()
	}

	stmt.WriteIdent(q.TableName)

	if q.TableAlias != "" {
		stmt.Token.As()
		stmt.WriteIdent(q.TableAlias)
	}

	if q.IndexedBy != "" {
		stmt.Token.Indexed().By()
		stmt.WriteIdent(q.IndexedBy)
	}

	if q.NotIndexed {
		stmt.Token.Not().Indexed()
	}

	return stmt.String()
}

func (q *QualifiedTableName) check() {
	if q.TableName == "" {
		panic("table name is empty")
	}

	if q.IndexedBy != "" && q.NotIndexed {
		panic("indexed by and not indexed cannot be set at the same time")
	}
}

// SetSchema sets the schema of the table.
// It should not be called by the parser, and is meant to be called
// by processes after parsing.
func (q *QualifiedTableName) SetSchema(schema string) {
	q.schema = schema
}
