package ddl

import (
	"fmt"
	"strings"

	"github.com/kwilteam/kwil-db/core/types"
)

func GenerateCreateTableStatement(pgSchema string, table *types.Table) (string, error) {
	var columnsAndKeys []string

	for _, column := range table.Columns {
		colName := wrapIdent(column.Name)
		colType, err := column.Type.PGString()
		if err != nil {
			return "", err
		}

		var colAttributes []string

		for _, attr := range column.Attributes {
			attrStr, err := attributeToSQLString(column.Name, attr)
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
		fkStmt, err := generateForeignKeyStmt(pgSchema, fk) // for now assume that schema for all foreign key tables is same
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

	return fmt.Sprintf("CREATE TABLE %s.%s (  %s) ;", wrapIdent(pgSchema),
		wrapIdent(table.Name), strings.Join(columnsAndKeys, ",  ")), nil
}

func wrapIdents(idents []string) []string {
	for i, ident := range idents {
		idents[i] = wrapIdent(ident)
	}
	return idents
}

func generateForeignKeyStmt(pgSchema string, fk *types.ForeignKey) (string, error) {
	stmt := strings.Builder{}
	stmt.WriteString(` FOREIGN KEY (`)
	writeDelimitedStrings(&stmt, fk.ChildKeys)
	stmt.WriteString(`) REFERENCES `)
	if pgSchema != "" {
		stmt.WriteString(wrapIdent(pgSchema)) // fk.ParentSchema maybe
		stmt.WriteString(".")
	}
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

func generateForeignKeyActionClause(action *types.ForeignKeyAction) (string, error) {
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
