package sqlddlgenerator

import (
	"fmt"
	"strings"

	"github.com/kwilteam/kwil-db/pkg/engine/dto"
)

func GenerateCreateTableStatement(table *dto.Table) (string, error) {
	var columnsAndKeys []string

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
		columnsAndKeys = append(columnsAndKeys, strings.TrimSpace(columnDef))
	}

	// now add foreign keys
	for _, fk := range table.ForeignKeys {
		fkStmt, err := generateForeignKeyStmt(fk)
		if err != nil {
			return "", err
		}
		columnsAndKeys = append(columnsAndKeys, fkStmt)
	}

	// now build the primary key
	pkColumns, err := table.GetPrimaryKey()
	if err != nil {
		return "", err
	}

	columnsAndKeys = append(columnsAndKeys, fmt.Sprintf("PRIMARY KEY (%s)", strings.Join(wrapIdents(pkColumns), ", ")))

	return fmt.Sprintf("CREATE TABLE %s (  %s) WITHOUT ROWID, STRICT;", wrapIdent(table.Name), strings.Join(columnsAndKeys, ",  ")), nil
}

func wrapIdents(idents []string) []string {
	for i, ident := range idents {
		idents[i] = wrapIdent(ident)
	}
	return idents
}

func generateForeignKeyStmt(fk *dto.ForeignKey) (string, error) {
	err := fk.Clean()
	if err != nil {
		return "", err
	}

	stmt := strings.Builder{}
	stmt.WriteString(` FOREIGN KEY (`)
	writeDelimitedStrings(&stmt, fk.ChildKeys)
	stmt.WriteString(`) REFERENCES `)
	stmt.WriteString(wrapIdent(fk.ParentTable))
	stmt.WriteString("(")
	writeDelimitedStrings(&stmt, fk.ParentKeys)
	stmt.WriteString(") ")

	for _, action := range fk.Actions {
		actionStmt, err := generateForeignKeyActionClause(action)
		if err != nil {
			return "", err
		}
		stmt.WriteString(actionStmt)
	}

	return stmt.String(), nil
}

func writeDelimitedStrings(stmt *strings.Builder, strs []string) {
	for i, str := range strs {
		if i > 0 && i < len(strs) {
			stmt.WriteString(", ")
		}

		stmt.WriteString(wrapIdent(str))
	}
}

func generateForeignKeyActionClause(action *dto.ForeignKeyAction) (string, error) {
	err := action.Clean()
	if err != nil {
		return "", err
	}

	stmt := strings.Builder{}
	stmt.WriteString(" ON ")
	stmt.WriteString(action.On.String())
	stmt.WriteString(" ")
	stmt.WriteString(action.Do.String())
	stmt.WriteString(" ")

	return stmt.String(), nil
}
