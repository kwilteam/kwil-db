// schemabuilder uses the ddl package to build a valid schema string from a
// database.
package schemabuilder

import (
	"kwil/pkg/databases"
	"kwil/pkg/databases/spec"
	"strings"
)

func GenerateDDL(db *databases.Database[*spec.KwilAny]) (string, error) {
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
