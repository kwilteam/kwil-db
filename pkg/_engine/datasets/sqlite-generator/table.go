package sqlitegenerator

import (
	"fmt"
	"github.com/kwilteam/kwil-db/pkg/engine/models"
	"strings"
)

func GenerateCreateTableStatement(table *models.Table) (string, error) {
	var columns []string

	for _, column := range table.Columns {
		colName := column.Name
		colType := columnTypeToSQLiteType(column.Type)
		var colAttributes []string

		for _, attr := range column.Attributes {
			attrStr, err := attributeToSQLiteString(column.Name, attr)
			if err != nil {
				return "", err
			}
			if attrStr != "" {
				colAttributes = append(colAttributes, attrStr)
			}
		}

		columnDef := fmt.Sprintf("%s %s %s", colName, colType, strings.Join(colAttributes, " "))
		columns = append(columns, strings.TrimSpace(columnDef))
	}

	return fmt.Sprintf("CREATE TABLE %s (\n  %s\n);", table.Name, strings.Join(columns, ",\n  ")), nil
}
