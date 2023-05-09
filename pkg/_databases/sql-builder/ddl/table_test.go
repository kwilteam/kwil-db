package ddlbuilder_test

import (
	"fmt"
	"github.com/kwilteam/kwil-db/pkg/databases/spec"
	ddlbuilder "github.com/kwilteam/kwil-db/pkg/databases/sql-builder/ddl"
	"testing"
)

func Test_BuildTable(t *testing.T) {

	cb := ddlbuilder.NewColumnBuilder()
	col1 := cb.Name("id").Type("INT8").WithAttribute(spec.PRIMARY_KEY, true)

	cb = ddlbuilder.NewColumnBuilder()
	col2 := cb.Name("name").Type("TEXT").WithAttribute(spec.UNIQUE, true)

	tb := ddlbuilder.NewTableBuilder()
	strs, err := tb.Schema("kwil").Name("test").AddColumn(col1).AddColumn(col2).Build()
	if err != nil {
		t.Fatal(err)
	}

	for _, str := range strs {
		fmt.Println(str)
	}
}
