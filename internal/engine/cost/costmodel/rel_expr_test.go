package costmodel

import (
	"github.com/kwilteam/kwil-db/internal/engine/cost/internal/testkit"
	"github.com/kwilteam/kwil-db/internal/engine/cost/query_planner"
	sqlparser "github.com/kwilteam/kwil-db/parse/sql"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_RelExpr_String(t *testing.T) {
	tests := []struct {
		name string
		r    *RelExpr
		want string
	}{
		{
			name: "test",
			r:    &RelExpr{},
			want: "test\n\n  stat: &{0 []}\n  cost: 0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.r.String(); got != tt.want {
				t.Errorf("RelExpr.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_NewRelExpr(t *testing.T) {
	cat := testkit.InitMockCatalog()

	tests := []struct {
		name string
		sql  string
		wt   string // want
	}{
		/////////////////////// no relation
		{
			name: "select int",
			sql:  "SELECT 1",
			wt: "Projection: 1, Stat: (RowCount: 0), Cost: 0\n" +
				"  NoRelationOp, Stat: (RowCount: 0), Cost: 0\n",
		},
		{
			name: "select string",
			sql:  "SELECT 'hello'",
			wt: "Projection: 'hello', Stat: (RowCount: 0), Cost: 0\n" +
				"  NoRelationOp, Stat: (RowCount: 0), Cost: 0\n",
		},
		{
			name: "select value expression",
			sql:  "SELECT 1+2",
			wt: "Projection: 1 + 2, Stat: (RowCount: 0), Cost: 0\n" +
				"  NoRelationOp, Stat: (RowCount: 0), Cost: 0\n",
		},
		// TODO: add function metadata to catalog
		// TODO: add support for functions in logical expr
		//{
		//	name: "select function abs",
		//	sql:  "SELECT ABS(-1)",
		//	wt:   "",
		//},
		/////////////////////// one relation
		{
			name: "select wildcard",
			sql:  "SELECT * FROM users",
			wt: "Projection: users.id, users.username, users.age, users.state, users.wallet, Stat: (RowCount: 0), Cost: 0\n" +
				"  Scan: users, Stat: (RowCount: 5), Cost: 0\n",
		},
		//{ // TODO?
		//	name: "select wildcard, deduplication",
		//	sql:  "SELECT *, age FROM users",
		//	wt: "Projection: users.id, users.username, users.age, users.state, users.wallet\n" +
		//		"  Scan: users; projection=[]\n",
		//},
		{
			name: "select columns",
			sql:  "select username, age from users",
			wt: "Projection: users.username, users.age, Stat: (RowCount: 0), Cost: 0\n" +
				"  Scan: users, Stat: (RowCount: 5), Cost: 0\n",
		},
		{
			name: "select column with alias",
			sql:  "select username as name from users",
			wt: "Projection: users.username AS name\n" +
				"  Scan: users\n",
		},
		{
			name: "select column expression",
			sql:  "select username, age+10 from users",
			wt: "Projection: users.username, users.age + 10\n" +
				"  Scan: users\n",
		},
		{
			name: "select with where",
			sql:  "select username, age from users where age > 20",
			wt: "Projection: users.username, users.age\n" +
				"  Filter: users.age > 20\n" +
				"    Scan: users\n",
		},
		{
			name: "select with multiple where",
			sql:  "select username, age from users where age > 20 and state = 'CA'",
			wt: "Projection: users.username, users.age\n" +
				"  Filter: users.age > 20 AND users.state = 'CA'\n" +
				"    Scan: users\n",
		},
		//{
		//	name: "select with group by",
		//	sql:  "select username, count(*) from users group by username",
		//	wt:   "GroupBy: users.username\n",
		//},
		{
			name: "select with limit, without offset",
			sql:  "select username, age from users limit 10",
			wt: "Limit: skip=0, fetch=10\n" +
				"  Projection: users.username, users.age\n" +
				"    Scan: users\n",
		},
		{
			name: "select with limit and offset",
			sql:  "select username, age from users limit 10 offset 5",
			wt: "Limit: skip=5, fetch=10\n" +
				"  Projection: users.username, users.age\n" +
				"    Scan: users\n",
		},
		{
			name: "select with order by default",
			sql:  "select username, age from users order by age",
			wt: "Sort: age ASC NULLS LAST\n" +
				"  Projection: users.username, users.age\n" +
				"    Scan: users\n",
		},
		{
			name: "select with order by desc",
			sql:  "select username, age from users order by age desc",
			wt: "Sort: age DESC NULLS FIRST\n" +
				"  Projection: users.username, users.age\n" +
				"    Scan: users\n",
		},
		/////////////////////// subquery
		{
			name: "select with subquery",
			sql:  "select username, age from (select * from users) as u",
			wt: "Projection: users.username, users.age\n" +
				"  Projection: users.id, users.username, users.age, users.state, users.wallet\n" +
				"    Scan: users\n",
		},
		/////////////////////// two relations

	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stmt, err := sqlparser.Parse(tt.sql)
			assert.NoError(t, err)

			q := query_planner.NewPlanner(cat)
			plan := q.ToPlan(stmt)
			rel := BuildRelExpr(plan)
			assert.Equal(t, tt.wt, Format(rel, 0))
		})
	}
}
