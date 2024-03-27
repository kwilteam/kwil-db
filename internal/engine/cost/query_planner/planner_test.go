package query_planner

import (
	"flag"
	"fmt"
	"github.com/kwilteam/kwil-db/internal/engine/cost/internal/testkit"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kwilteam/kwil-db/internal/engine/cost/catalog"
	ds "github.com/kwilteam/kwil-db/internal/engine/cost/datasource"
	dt "github.com/kwilteam/kwil-db/internal/engine/cost/datatypes"
	"github.com/kwilteam/kwil-db/internal/engine/cost/logical_plan"
	sqlparser "github.com/kwilteam/kwil-db/parse/sql"
)

var (
	testData        = flag.String("test_data", "testdata/[^.]*", "test data glob")
	updateTestFiles = flag.Bool("update", false, "update test golden files")
)

type mockCatalog struct {
	tables map[string]*dt.Schema
}

func (m *mockCatalog) GetSchemaSource(tableRef *dt.TableRef) (ds.SchemaSource, error) {
	relName := tableRef.Table // for testing, we ignore the database name
	schema, ok := m.tables[relName]
	if !ok {
		return nil, fmt.Errorf("table %s not found", relName)
	}
	return ds.NewExampleSchemaSource(schema), nil
}

func initMockCatalog() *mockCatalog {
	stubUserData, _ := ds.NewCSVDataSource("../testdata/users.csv")
	stubPostData, _ := ds.NewCSVDataSource("../testdata/posts.csv")
	commentsData, _ := ds.NewCSVDataSource("../testdata/comments.csv")
	commentRelData, _ := ds.NewCSVDataSource("../testdata/comment_rel.csv")

	return &mockCatalog{
		tables: map[string]*dt.Schema{
			// for testing, we ignore the database name
			"users":       stubUserData.Schema(),
			"posts":       stubPostData.Schema(),
			"comments":    commentsData.Schema(),
			"comment_rel": commentRelData.Schema(),
		},
	}
}

func runToPlan(t *testing.T, sql string, cat catalog.Catalog) string {
	t.Helper()

	stmt, err := sqlparser.Parse(sql)
	assert.NoError(t, err)

	q := NewPlanner(cat)
	plan := q.ToPlan(stmt)
	return logical_plan.Format(plan, 0)
	//
}

func Test_queryPlanner_ToPlan_golden(t *testing.T) {
	// test with golden files, located in ./testdata
	cat := initMockCatalog()

	testFiles, err := filepath.Glob(*testData)
	assert.NoError(t, err)
	assert.NotEmpty(t, testFiles, "no test files found")

	for _, testFile := range testFiles {
		r, err := testkit.NewTestDataReader(testFile, *updateTestFiles)
		assert.NoError(t, err)

		for r.Next() {
			//fmt.Printf("Running test: %+v\n", r.Data)

			tc := r.Data
			sql := tc.Sql
			expected := tc.Expected

			t.Run(tc.CaseName, func(t *testing.T) {
				got := runToPlan(t, sql, cat)
				r.Record(got) // record the result for update purposes
				if !*updateTestFiles {
					assert.Equal(t, expected, got)
				}
			})

			r.Rewrite() // write the updated test file
		}
	}
}

func Test_queryPlanner_ToPlan(t *testing.T) {
	cat := initMockCatalog()

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
				"  NoRelation\n",
		},
		{
			name: "select string",
			sql:  "SELECT 'hello'",
			wt: "Projection: 'hello'\n" +
				"  NoRelation\n",
		},
		{
			name: "select value expression",
			sql:  "SELECT 1+2",
			wt: "Projection: 1 + 2\n" +
				"  NoRelation\n",
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
			wt: "Projection: users.id, users.username, users.age, users.state, users.wallet\n" +
				"  Scan: users\n",
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
			wt: "Projection: users.username, users.age\n" +
				"  Scan: users\n",
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

			q := NewPlanner(cat)
			plan := q.ToPlan(stmt)
			got := logical_plan.Format(plan, 0)
			assert.Equal(t, tt.wt, got)
		})
	}
}

//func runToPlanTest(sql string) {
//	cat := initMockCatalog()
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
