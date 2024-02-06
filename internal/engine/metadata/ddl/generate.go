package ddl

import (
	"fmt"

	"github.com/kwilteam/kwil-db/internal/engine/types"
)

// GenerateDDL generates the necessary table and index ddl statements for the given table
func GenerateDDL(dbid string, table *types.Table) ([]string, error) {
	// Derive the PostgreSQL "schema" (not a Kwil schema) from the dbid.
	pgSchema := types.DBIDSchema(dbid)

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

	return statements, nil
}

func wrapIdent(str string) string {
	return fmt.Sprintf(`"%s"`, str)
}
