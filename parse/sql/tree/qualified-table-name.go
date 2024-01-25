package tree

import sqlwriter "github.com/kwilteam/kwil-db/parse/sql/tree/sql-writer"

type QualifiedTableName struct {
	node

	TableName  string
	TableAlias string
	IndexedBy  string
	NotIndexed bool
}

func (q *QualifiedTableName) Accept(v AstVisitor) any {
	return v.VisitQualifiedTableName(q)
}

func (q *QualifiedTableName) Walk(w AstWalker) error {
	return run(
		w.EnterQualifiedTableName(q),
		w.ExitQualifiedTableName(q),
	)
}

func (q *QualifiedTableName) ToSQL() string {
	q.check()

	stmt := sqlwriter.NewWriter()
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
