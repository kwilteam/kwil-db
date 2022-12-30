package schemabuilder

import (
	"kwil/x/execution/dto"
	"strings"
)

// the schemabuilder package uses the ddl package to build a schema from a
// database.

func GenerateDDL(db *dto.Database) (string, error) {
	schemaName := db.GetSchemaName()

	stmts := []string{}
	for _, table := range db.Tables {
		stmt, err := GenerateTableDDL(table, schemaName)
		if err != nil {
			return "", err
		}
		stmts = append(stmts, stmt...)
	}

	for _, index := range db.Indexes {
		stmt := GenerateIndexDDL(index, schemaName)
		stmts = append(stmts, stmt)
	}

	return strings.Join(stmts, "\n "), nil
}
