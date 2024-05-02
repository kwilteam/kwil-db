package actparser

import (
	"github.com/kwilteam/kwil-db/parse/sql/tree"
	parseTypes "github.com/kwilteam/kwil-db/parse/types"
)

type ActionStmt interface {
	StmtType() string
}

type ExtensionCallStmt struct {
	parseTypes.Node
	Extension string
	Method    string
	Args      []tree.Expression
	Receivers []string
}

type ActionCallStmt struct {
	parseTypes.Node
	Database  string // for future use, e.g. call an action from another kuneiform
	Method    string
	Args      []tree.Expression
	Receivers []string
}

// DMLStmt is a DML statement, we leave the parsing to sqlparser
type DMLStmt struct {
	parseTypes.Node
	Statement string
}

func (s *ExtensionCallStmt) StmtType() string {
	return "extension_call"
}

func (s *ActionCallStmt) StmtType() string {
	return "action_call"
}

func (s *DMLStmt) StmtType() string {
	return "dml"
}
