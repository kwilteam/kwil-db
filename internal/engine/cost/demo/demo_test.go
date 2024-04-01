package demo

import (
	"context"
	"fmt"

	ds "github.com/kwilteam/kwil-db/internal/engine/cost/datasource"
	dt "github.com/kwilteam/kwil-db/internal/engine/cost/datatypes"
	"github.com/kwilteam/kwil-db/internal/engine/cost/query_planner"
	sqlparser "github.com/kwilteam/kwil-db/parse/sql"
)

var (
	stubUserData, _   = ds.NewCSVDataSource("../testdata/users.csv")
	stubPostData, _   = ds.NewCSVDataSource("../testdata/posts.csv")
	commentsData, _   = ds.NewCSVDataSource("../testdata/comments.csv")
	commentRelData, _ = ds.NewCSVDataSource("../testdata/comment_rel.csv")
)

type mockCatalog struct {
	tables   map[string]*dt.Schema
	dataSrcs map[string]ds.DataSource
}

func (m *mockCatalog) GetDataSource(tableRef *dt.TableRef) (ds.DataSource, error) {
	relName := tableRef.Table // for testing, we ignore the database name
	if _, ok := m.tables[relName]; !ok {
		return nil, fmt.Errorf("table %s not found", relName)
	}
	return m.dataSrcs[relName], nil
}

func initMockCatalog() *mockCatalog {
	return &mockCatalog{
		dataSrcs: map[string]ds.DataSource{
			"users":       stubUserData,
			"posts":       stubPostData,
			"comments":    commentsData,
			"comment_rel": commentRelData,
		},
		tables: map[string]*dt.Schema{
			// for testing, we ignore the database name
			"users":       stubUserData.Schema(),
			"posts":       stubPostData.Schema(),
			"comments":    commentsData.Schema(),
			"comment_rel": commentRelData.Schema(),
		},
	}
}

func ExampleDemo() {
	// enter engine
	rawSql := "SELECT state, username FROM users WHERE age = 20"
	stmt, err := sqlparser.Parse(rawSql)
	if err != nil {
		panic(err)
	}

	// load into engine
	catalog := initMockCatalog()

	planner := query_planner.NewPlanner(catalog)
	plan := planner.ToPlan(stmt)

	//opt := optimizer.NewOptimizer()
	//plan := opt.Optimize(plan)

	ctx := NewExecutionContext()
	res := ctx.Execute(context.TODO(), plan)
	fmt.Println(res.ToCsv())

	// Output:
	// state,username
	// CA,Adam
}
