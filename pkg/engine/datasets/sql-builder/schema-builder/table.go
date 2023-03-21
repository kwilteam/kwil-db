package schemabuilder

import (
	"fmt"
	ddlbuilder "kwil/pkg/engine/datasets/sql-builder/ddl"
	"kwil/pkg/engine/models"
	"kwil/pkg/engine/types"
)

func GenerateTableDDL(t *models.Table, schemaName string) ([]string, error) {
	tbl := ddlbuilder.NewTableBuilder().Schema(schemaName).Name(t.Name)
	for _, col := range t.Columns {
		cb := ddlbuilder.NewColumnBuilder()

		// convert column type to Postgres type
		pgtype, err := types.DataTypeConversions.KwilToPgType(col.Type)
		if err != nil {
			return nil, fmt.Errorf("failed to convert kwil type to postgres type: %w", err)
		}

		column := cb.Name(col.Name).Type(pgtype)

		// generate attribute alter statements
		for _, attr := range col.Attributes {
			attributeVal, err := types.NewFromSerial(attr.Value)
			if err != nil {
				return nil, fmt.Errorf("failed to convert attribute value to kwil type: %w", err)
			}

			column.WithAttribute(attr.Type, attributeVal)
		}

		tbl.AddColumn(column)
	}

	return tbl.Build()
}
