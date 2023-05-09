package schemabuilder

import (
	"github.com/kwilteam/kwil-db/pkg/databases"
	ddlb "github.com/kwilteam/kwil-db/pkg/databases/sql-builder/ddl"
)

func GenerateIndexDDL(index *databases.Index, schemaName string) string {
	return ddlb.NewIndexBuilder().Name(index.Name).Schema(schemaName).Table(index.Table).Columns(index.Columns...).Using(index.Using).Build()
}
