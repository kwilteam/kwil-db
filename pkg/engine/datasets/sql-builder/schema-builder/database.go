// schemabuilder uses the ddl package to build a valid schema string from a
// database.
package schemabuilder

import (
	"kwil/pkg/engine/models"
	"strings"
)

func GenerateDDL(db *models.Dataset) (string, error) {
	schemaName := db.GetSchemaName()

	stmts := []string{}
	for _, table := range db.Tables {
		stmt, err := GenerateTableDDL(table, schemaName)
		if err != nil {
			return "", err
		}
		stmts = append(stmts, stmt...)
	}

	// TODO: build indexes

	return strings.Join(stmts, "\n "), nil
}
