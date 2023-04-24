package sqlitegenerator

import "kwil/pkg/engine/models"

func GenerateDDL(table *models.Table) ([]string, error) {
	var statements []string

	createTableStatement, err := GenerateCreateTableStatement(table)
	if err != nil {
		return nil, err
	}
	statements = append(statements, createTableStatement)

	createIndexStatements, err := GenerateCreateIndexStatements(table.Name, table.Indexes)
	if err != nil {
		return nil, err
	}
	statements = append(statements, createIndexStatements...)

	return statements, nil
}
