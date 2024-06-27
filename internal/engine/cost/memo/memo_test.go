package memo

import (
	"fmt"
	ds "github.com/kwilteam/kwil-db/internal/engine/cost/datasource"
	dt "github.com/kwilteam/kwil-db/internal/engine/cost/datatypes"
	"github.com/kwilteam/kwil-db/internal/engine/cost/internal/testkit"
	lp "github.com/kwilteam/kwil-db/internal/engine/cost/logical_plan"
	"testing"
)

var stubDS, _ = ds.NewCSVDataSource("../testdata/users.csv")
var stubTable = &dt.TableRef{Table: "users"}

func TestMemo(t *testing.T) {
	catalog := testkit.InitMockCatalog()
	dataSrc, err := catalog.GetDataSource(stubTable)
	if err != nil {
		panic(err)
	}

	df := lp.NewDataFrame(
		lp.Scan(stubTable, dataSrc, nil))

	plan := df.
		Filter(lp.Eq(lp.Column(stubTable, "age"),
			lp.LiteralNumeric(20))).
		Project(lp.Column(stubTable, "state"),
			lp.Column(stubTable, "username"),
		).
		LogicalPlan()

	memo := NewMemo()

	g := memo.Init(plan)

	fmt.Println(Format(g.logical[0], 0))

	//for _, g := range memo.groups {
	//	fmt.Println(g)
	//}
}
