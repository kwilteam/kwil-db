package schemabuilder

import (
	"fmt"
	"kwil/x/execution"
	ddlb "kwil/x/execution/sql-builder/ddl"
	"kwil/x/types/databases"
)

func GenerateTableDDL(t *databases.Table, schemaName string) ([]string, error) {
	tbl := ddlb.NewTableBuilder().Schema(schemaName).Name(t.Name)
	for _, col := range t.Columns {
		cb := ddlb.NewColumnBuilder()

		// convert column type to Postgres type
		pgtype, err := execution.DataTypes.KwilToPgType(col.Type)
		if err != nil {
			return nil, fmt.Errorf("failed to convert kwil type to postgres type: %w", err)
		}

		column := cb.Name(col.Name).Type(pgtype)

		// generate attribute alter statements
		for _, attr := range col.Attributes {
			/*if attr.Value == nil {
				continue
			}*/

			column.WithAttribute(attr.Type, attr.Value)
		}

		tbl.AddColumn(column)
	}

	return tbl.Build()
}
