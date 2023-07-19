package sqlanalyzer

import "github.com/kwilteam/kwil-db/pkg/engine/sqlparser/tree"

type StatementContext struct {
	StatementType statementType
	MainTable     string
}

type statementType string

const (
	StatementTypeSelect statementType = "SELECT"
	StatementTypeInsert statementType = "INSERT"
	StatementTypeUpdate statementType = "UPDATE"
	StatementTypeDelete statementType = "DELETE"
)

type SelectStatementContext struct {
	StatementContext
	JoinedTables []string
}

type statementContextVisitor struct {
	*tree.BaseVisitor
}
