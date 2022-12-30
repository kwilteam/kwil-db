package schemabuilder

import (
	"kwil/x/execution/dto"
	ddlb "kwil/x/execution/sql-builder/ddl"
)

func GenerateIndexDDL(index *dto.Index, schemaName string) string {
	return ddlb.NewIndexBuilder().Name(index.Name).Schema(schemaName).Table(index.Table).Columns(index.Columns...).Using(index.Using).Build()
}
