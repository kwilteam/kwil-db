package sqlddlgenerator

import (
	"fmt"
	"strings"

	"github.com/kwilteam/kwil-db/pkg/engine/dto"
)

func GenerateCreateTableStatement(table *dto.Table) (string, error) {
	var columnsAndPrimaryKey []string

	for _, column := range table.Columns {
		colName := wrapIdent(column.Name)
		colType, err := columnTypeToSQLiteType(column.Type)
		if err != nil {
			return "", err
		}

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
		columnsAndPrimaryKey = append(columnsAndPrimaryKey, strings.TrimSpace(columnDef))
	}

	// now build the primary key
	pkColumns, err := table.GetPrimaryKey()
	if err != nil {
		return "", err
	}

	columnsAndPrimaryKey = append(columnsAndPrimaryKey, fmt.Sprintf("PRIMARY KEY (%s)", strings.Join(wrapIdents(pkColumns), ", ")))

	return fmt.Sprintf("CREATE TABLE %s (  %s) WITHOUT ROWID, STRICT;", wrapIdent(table.Name), strings.Join(columnsAndPrimaryKey, ",  ")), nil
}

func wrapIdents(idents []string) []string {
	for i, ident := range idents {
		idents[i] = wrapIdent(ident)
	}
	return idents
}
