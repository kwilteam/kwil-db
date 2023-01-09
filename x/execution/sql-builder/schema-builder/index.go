package schemabuilder

import (
	ddlb "kwil/x/execution/sql-builder/ddl"
	"kwil/x/types/databases"
)

func GenerateIndexDDL(index *databases.Index, schemaName string) string {
	return ddlb.NewIndexBuilder().Name(index.Name).Schema(schemaName).Table(index.Table).Columns(index.Columns...).Using(index.Using).Build()
}
