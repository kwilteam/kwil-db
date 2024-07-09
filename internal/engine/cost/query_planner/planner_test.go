package query_planner

import (
	"flag"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/internal/engine/cost/internal/testkit"
	"github.com/kwilteam/kwil-db/internal/engine/cost/logical_plan"
	"github.com/kwilteam/kwil-db/parse"
)

var (
	testData        = flag.String("test_data", "testdata/[^.]*", "test data glob")
	updateTestFiles = flag.Bool("update", false, "update test golden files")
)

//func Test_queryPlanner_ToPlan_golden(t *testing.T) {
//	// test with golden files, located in ./testdata
//	cat := testkit.InitMockCatalog()
//
//	testFiles, err := filepath.Glob(*testData)
//	assert.NoError(t, err)
//	assert.NotEmpty(t, testFiles, "no test files found")
//
//	for _, testFile := range testFiles {
//		r, err := testkit.NewTestDataReader(testFile, *updateTestFiles)
//		assert.NoError(t, err)
//
//		for r.Next() {
//			//fmt.Printf("Running test: %+v\n", r.Data)
//
//			tc := r.Data
//			sql := tc.Sql
//			expected := tc.Expected
//
//			t.Run(tc.CaseName, func(t *testing.T) {
//				got := runToPlan(t, sql, cat)
//				r.Record(got) // record the result for update purposes
//				if !*updateTestFiles {
//					assert.Equal(t, expected, got)
//				}
//			})
//
//			r.Rewrite() // write the updated test file
//		}
//	}
//}

func Test_queryPlanner_ToPlan(t *testing.T) {
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
			wt: "Projection: 1\n" +
				"  NoRelationOp\n",
		},
		{
			name: "select int union",
			sql:  "SELECT 1 UNION SELECT 2",
			wt: `Sort:  ASC NULLS LAST
  Distinct
    UNION: Projection: 1, Projection: 2
      Projection: 1
        NoRelationOp
      Projection: 2
        NoRelationOp
`, // is this right?
		},
		{
			name: "select string",
			sql:  "SELECT 'hello'",
			wt: "Projection: 'hello'\n" +
				"  NoRelationOp\n",
		},
		{
			name: "select value expression",
			sql:  "SELECT 1+2",
			wt: "Projection: 1 + 2\n" +
				"  NoRelationOp\n",
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
			wt: "Sort: id ASC NULLS LAST\n" +
				"  Projection: users.id, users.username, users.age, users.state, users.wallet\n" +
				"    Scan: users\n",
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
			wt: "Sort: id ASC NULLS LAST\n" +
				"  Projection: users.username, users.age\n" +
				"    Scan: users\n",
		},
		{
			name: "select column with alias",
			sql:  "select username as name from users",
			wt: "Sort: id ASC NULLS LAST\n" +
				"  Projection: users.username AS name\n" +
				"    Scan: users\n",
		},
		{
			name: "select column expression",
			sql:  "select username, age+10 from users",
			wt: "Sort: id ASC NULLS LAST\n" +
				"  Projection: users.username, users.age + 10\n" +
				"    Scan: users\n",
		},
		{
			name: "select with where",
			sql:  "select username, age from users where age > 20",
			wt: "Sort: id ASC NULLS LAST\n" +
				"  Projection: users.username, users.age\n" +
				"    Filter: users.age > 20\n" +
				"      Scan: users\n",
		},
		{
			name: "select with multiple where",
			sql:  "select username, age from users where age > 20 and state = 'CA'",
			wt: "Sort: id ASC NULLS LAST\n" +
				"  Projection: users.username, users.age\n" +
				"    Filter: users.age > 20 AND users.state = 'CA'\n" +
				"      Scan: users\n",
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
				"  Sort: id ASC NULLS LAST\n" +
				"    Projection: users.username, users.age\n" +
				"      Scan: users\n",
		},
		{
			name: "select with limit and offset",
			sql:  "select username, age from users limit 10 offset 5",
			wt: "Limit: skip=5, fetch=10\n" +
				"  Sort: id ASC NULLS LAST\n" +
				"    Projection: users.username, users.age\n" +
				"      Scan: users\n",
		},
		{
			name: "select with order by default",
			sql:  "select username, age from users order by age",
			wt: "Sort: age ASC NULLS LAST, id ASC NULLS LAST\n" +
				"  Projection: users.username, users.age\n" +
				"    Scan: users\n",
		},
		{
			name: "select with order by desc",
			sql:  "select username, age from users order by age desc",
			wt: "Sort: age DESC NULLS LAST, id ASC NULLS LAST\n" +
				"  Projection: users.username, users.age\n" +
				"    Scan: users\n",
		},
		/////////////////////// subquery
		{
			name: "select with subquery",
			sql:  "select username, age from (select * from users) as u",
			wt: "Sort: id ASC NULLS LAST, username ASC NULLS LAST, age ASC NULLS LAST, state ASC NULLS LAST, wallet ASC NULLS LAST\n" +
				"  Projection: users.username, users.age\n" +
				"    Sort: id ASC NULLS LAST\n" +
				"      Projection: users.id, users.username, users.age, users.state, users.wallet\n" +
				"        Scan: users\n",
		},
		/////////////////////// two relations

	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// NOTE: These tests' expected plans are assuming the ORDER BY added
			// by ParseSQL. With ParseSQLWithoutValidation, the "Sort:" would
			// not be there unless it was in the origin al SQL statement.
			pr, err := parse.ParseSQL(tt.sql, &types.Schema{
				Name:   "",
				Tables: []*types.Table{testkit.MockUsersSchemaTable},
			})

			assert.NoError(t, err)
			assert.NoError(t, pr.ParseErrs.Err())

			q := NewPlanner(cat)
			plan := q.ToPlan(pr.AST)
			got := logical_plan.Format(plan, 0)
			assert.Equal(t, tt.wt, got)
		})
	}
}

//func runToPlanTest(sql string) {
//  cat := testkit.InitMockCatalog()
//	q := NewPlanner(cat)
//	stmt, err := sqlparser.Parse(sql)
//	if err != nil {
//		log.Fatal(fmt.Sprintf("failed to parse sql: %s", err))
//	}
//
//	plan := q.ToPlan(stmt)
//	fmt.Println(logical_plan.Format(plan, 0))
//}
//
//func Example_queryPlanner_ToPlan_select_wildcard() {
//	sql := "SELECT * FROM users"
//	runToPlanTest(sql)
//	// Output:
//	// Projection: id, username, age, state, wallet
//	//   Scan: users; projection=[]
//}
