package ddlbuilder_test

import (
	"kwil/pkg/databases/spec"
	ddlb "kwil/pkg/databases/sql-builder/ddl"
	"testing"
)

func Test_BuildIndex(t *testing.T) {
	ib := ddlb.NewIndexBuilder()
	str := ib.Name("myindex").Schema("kwil").Table("test").Columns("id").Using(spec.BTREE).Build()
	if str != `CREATE INDEX myindex ON "kwil"."test" USING btree (id);` {
		t.Fatal("unexpected index string: ", str)
	}
}
