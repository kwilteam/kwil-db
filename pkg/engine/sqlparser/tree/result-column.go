package tree

import sqlwriter "github.com/kwilteam/kwil-db/pkg/engine/sqlparser/tree/sql-writer"

type ResultColumn interface {
	resultColumn()
	ToSQL() string
}

type ResultColumnStar struct{}

func (r ResultColumnStar) resultColumn() {}
func (r ResultColumnStar) ToSQL() string {
	return "*"
}

type ResultColumnExpression struct {
	Expression Expression
	Alias      string
}

func (r ResultColumnExpression) resultColumn() {}
func (r ResultColumnExpression) ToSQL() string {
	stmt := sqlwriter.NewWriter()
	stmt.WriteString(r.Expression.ToSQL())
	if r.Alias != "" {
		stmt.Token.As()
		stmt.WriteIdent(r.Alias)
	}
	return stmt.String()
}

type ResultColumnTable struct {
	TableName string
}

func (r ResultColumnTable) resultColumn() {}
func (r ResultColumnTable) ToSQL() string {
	stmt := sqlwriter.NewWriter()
	stmt.WriteIdent(r.TableName)
	stmt.Token.Period()
	stmt.Token.Asterisk()
	return stmt.String()
}
