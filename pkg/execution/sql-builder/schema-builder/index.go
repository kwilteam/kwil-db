package schemabuilder

import (
	ddlb "kwil/pkg/execution/sql-builder/ddl"
	"kwil/pkg/types/databases"
)

func GenerateIndexDDL(index *databases.Index, schemaName string) string {
	return ddlb.NewIndexBuilder().Name(index.Name).Schema(schemaName).Table(index.Table).Columns(index.Columns...).Using(index.Using).Build()
}
