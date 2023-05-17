package sqlddlgenerator

import (
	"fmt"
	"strings"

	"github.com/kwilteam/kwil-db/pkg/engine2/dto"
)

func GenerateCreateTableStatement(table *dto.Table) (string, error) {
	var columns []string

	for _, column := range table.Columns {
		colName := wrapIdent(column.Name)
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

	return fmt.Sprintf("CREATE TABLE %s (  %s) WITHOUT ROWID;", wrapIdent(table.Name), strings.Join(columns, ",  ")), nil
}
