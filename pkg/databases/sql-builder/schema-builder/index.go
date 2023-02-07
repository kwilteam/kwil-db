package schemabuilder

import (
	"kwil/pkg/databases"
	ddlb "kwil/pkg/databases/sql-builder/ddl"
)

func GenerateIndexDDL(index *databases.Index, schemaName string) string {
	return ddlb.NewIndexBuilder().Name(index.Name).Schema(schemaName).Table(index.Table).Columns(index.Columns...).Using(index.Using).Build()
}
