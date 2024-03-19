package query_planner

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kwilteam/kwil-db/internal/engine/cost/logical_plan"
	sqlparser "github.com/kwilteam/kwil-db/parse/sql"
)

func Test_queryPlanner_ToPlan(t *testing.T) {
	type args struct {
		stmt string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
		{
			name: "simple select",
			args: args{
				stmt: "SELECT * FROM users",
			},
			want: "Scan: users; projection=[]",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast, err := sqlparser.Parse(tt.args.stmt)
			assert.NoError(t, err)

			q := &queryPlanner{}
			got := q.ToPlan(ast)
			explain := fmt.Sprintf(logical_plan.Format(got, 0))
			assert.Equal(t, tt.want, explain)
		})
	}
}
