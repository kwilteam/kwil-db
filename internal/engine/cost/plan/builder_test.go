package plan

import (
	"fmt"
	"testing"

	sqlparser "github.com/kwilteam/kwil-db/parse/sql"
	"github.com/kwilteam/kwil-db/parse/sql/tree"
	"github.com/stretchr/testify/assert"
)

func TestPlanBuilder_build(t1 *testing.T) {

	tests := []struct {
		name    string
		args    string
		want    *OperationBuilder
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "simple select",
			args: "select * from t",
		},
		{
			name: "select limit",
			args: "select * from t limit 1",
		},
		{
			name: "select limit offset",
			args: "select * from t limit 1 offset 2",
		},
		{
			name: "select limit offset 2",
			args: "select * from t limit 2, 1",
		},
		{
			name: "select order by",
			args: "select * from t order by t.c1",
		},
		{
			name: "select order by limit",
			args: "select * from t order by t.c1 limit 1",
		},
		{
			name: "filter where",
			args: "select * from t where t.c1 = 1",
		},
		{
			name: "filter where and",
			args: "select * from t where t.c1 = 1 and t.c2 = 2",
		},
		{
			name: "group by",
			args: "select t.c1 as a1, count(t.c2) from t group by t.c1",
		},
		{
			name: "group by having",
			args: "select t.c1 as a1, count(t.c2) from t group by t.c1 having t.c1 = 1",
		},
		{
			name: "join select from star",
			args: "select * from t as t1 join t as t2 on t1.c1 = t2.c2",
		},
		{
			name: "join select from table",
			args: "select t1.name, t2.age from t as t1 join t as t2 on t1.c1 = t2.c2",
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			b := NewBuilder()

			ast, err := sqlparser.Parse(tt.args)
			assert.NoError(t1, err)

			got := b.build(ast.(*tree.Select))
			//if (err != nil) != tt.wantErr {
			//	t1.Errorf("Transform() error = %v, wantErr %v", err, tt.wantErr)
			//	return
			//}
			//if !reflect.DeepEqual(got, tt.want) {
			//	t1.Errorf("Transform() got = %v, want %v", got, tt.want)
			//}

			fmt.Println(got.Build().Explain())
		})
	}
}
