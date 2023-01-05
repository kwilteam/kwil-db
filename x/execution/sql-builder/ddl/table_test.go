package ddlbuilder_test

import (
	"fmt"
	"kwil/x/execution"
	ddlb "kwil/x/execution/sql-builder/ddl"
	"testing"
)

func Test_BuildTable(t *testing.T) {

	cb := ddlb.NewColumnBuilder()
	col1 := cb.Name("id").Type("INT8").WithAttribute(execution.PRIMARY_KEY, true)

	cb = ddlb.NewColumnBuilder()
	col2 := cb.Name("name").Type("TEXT").WithAttribute(execution.UNIQUE, true)

	tb := ddlb.NewTableBuilder()
	strs, err := tb.Schema("kwil").Name("test").AddColumn(col1).AddColumn(col2).Build()
	if err != nil {
		t.Fatal(err)
	}

	for _, str := range strs {
		fmt.Println(str)
	}
}
