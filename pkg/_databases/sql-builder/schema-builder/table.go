package schemabuilder

import (
	"fmt"
	"github.com/kwilteam/kwil-db/pkg/databases"
	"github.com/kwilteam/kwil-db/pkg/databases/spec"
	ddlbuilder "github.com/kwilteam/kwil-db/pkg/databases/sql-builder/ddl"
)

func GenerateTableDDL(t *databases.Table[*spec.KwilAny], schemaName string) ([]string, error) {
	tbl := ddlbuilder.NewTableBuilder().Schema(schemaName).Name(t.Name)
	for _, col := range t.Columns {
		cb := ddlbuilder.NewColumnBuilder()

		// convert column type to Postgres type
		pgtype, err := spec.DataTypeConversions.KwilToPgType(col.Type)
		if err != nil {
			return nil, fmt.Errorf("failed to convert kwil type to postgres type: %w", err)
		}

		column := cb.Name(col.Name).Type(pgtype)

		// generate attribute alter statements
		for _, attr := range col.Attributes {
			/*if attr.Value == nil {
				continue
			}*/

			column.WithAttribute(attr.Type, attr.Value.Value())
		}

		tbl.AddColumn(column)
	}

	return tbl.Build()
}
