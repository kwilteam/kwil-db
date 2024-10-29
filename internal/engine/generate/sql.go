package generate

import (
	"fmt"

	"github.com/kwilteam/kwil-db/parse"
)

// WriteSQL converts a SQL node to a string.
// It can optionally rewrite named parameters to numbered parameters.
// If so, it returns the order of the parameters in the order they appear in the statement.
func WriteSQL(node *parse.SQLStatement, orderParams bool, pgSchema string) (stmt string, params []string, err error) {
	if node == nil {
		return "", nil, fmt.Errorf("SQL parse node is nil")
	}

	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("panic: %v", e)
		}
	}()

	sqlGen := &sqlGenerator{
		pgSchema: pgSchema,
	}
	if orderParams {
		sqlGen.numberParameters = true
	}
	stmt = node.Accept(sqlGen).(string)

	return stmt + ";", sqlGen.orderedParams, nil
}

// WriteDDL converts a DDL node to a string.
// NOTE: this is temporary so that I can write tests, after we shift from *parse.SQLStatement
// to parse.SQLStmt, we can delete this.
func WriteDDL(node parse.SQLStmt) (stmt string, err error) {
	if node == nil {
		return "", fmt.Errorf("SQL parse node is nil")
	}

	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("panic: %v", e)
		}
	}()

	sqlGen := &sqlGenerator{}

	stmt = node.Accept(sqlGen).(string)
	return stmt + ";", nil
}
