package generate

import (
	"fmt"

	"github.com/kwilteam/kwil-db/core/types"
)

// GenerateDDL generates the necessary table and index ddl statements for the given table
func GenerateDDL(pgSchema string, table *types.Table) ([]string, error) {
	var statements []string

	createTableStatement, err := GenerateCreateTableStatement(pgSchema, table)
	if err != nil {
		return nil, err
	}
	statements = append(statements, createTableStatement)

	createIndexStatements, err := GenerateCreateIndexStatements(pgSchema, table.Name, table.Indexes)
	if err != nil {
		return nil, err
	}
	statements = append(statements, createIndexStatements...)

	for _, stmt := range statements {
		if containsDisallowedDelimiter(stmt) {
			return nil, fmt.Errorf("statement contains disallowed delimiter: %s", stmt)
		}
	}

	return statements, nil
}

func wrapIdent(str string) string {
	return fmt.Sprintf(`"%s"`, str)
}
