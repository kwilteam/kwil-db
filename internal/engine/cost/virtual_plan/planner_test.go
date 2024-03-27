package virtual_plan

import (
	"fmt"
	ds "github.com/kwilteam/kwil-db/internal/engine/cost/datasource"
	dt "github.com/kwilteam/kwil-db/internal/engine/cost/datatypes"
	lp "github.com/kwilteam/kwil-db/internal/engine/cost/logical_plan"
	"github.com/kwilteam/kwil-db/internal/engine/cost/optimizer"
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

func Example_QueryPlanner_CreateVirtualPlan() {
	ctx := NewExecutionContext()
	df := ctx.csv("users", "../testdata/users.csv")
	plan := df.
		Filter(lp.Eq(lp.Column(stubTable, "age"),
			lp.LiteralNumeric(20))).
		Project(lp.Column(stubTable, "state"),
			lp.Column(stubTable, "username"),
		).
		LogicalPlan()

	fmt.Println(lp.Format(plan, 0))

	r := &optimizer.ProjectionRule{}
	got := r.Optimize(plan)

	fmt.Printf("---After optimization---\n\n")
	fmt.Println(lp.Format(got, 0))

	qp := NewPlanner()
	vp := qp.ToPlan(got)
	fmt.Printf("---Got virtual plan---\n\n")
	fmt.Println(Format(vp, 0))

	// Output:
	// Projection: users.state, users.username
	//   Filter: users.age = 20
	//     Scan: users
	//
	// ---After optimization---
	//
	// Projection: users.state, users.username
	//   Filter: users.age = 20
	//     Scan: users; projection=[age, state, username]
	//
	// ---Got virtual plan---
	//
	// VProjection: [state@1 username@2]
	//   VSelection: age@0 = 20
	//     VTableScan: schema=[id/int, username/string, age/int, state/string, wallet/string], projection=[age state username]
}
