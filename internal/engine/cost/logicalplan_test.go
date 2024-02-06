package cost_test

import (
	"fmt"
	"github.com/kwilteam/kwil-db/internal/engine/cost"
	"testing"
)

func TestLogicalPlan_String(t *testing.T) {
	ds := cost.NewMemDataSource(nil, nil)
	plan := cost.Scan("users", ds)
	plan = cost.Projection(plan, cost.Column("", "username"), cost.Column("", "age"))
	fmt.Println(cost.Format(plan, 0))
}

func TestLogicalPlan_DataFrame(t *testing.T) {
	ds := cost.NewMemDataSource(nil, nil)
	aop := cost.NewAlgebraOpBuilder(cost.Scan("users", ds))
	plan := aop.Filter(cost.Eq(cost.Column("", "age"), cost.LiteralInt(20))).
		Aggregate([]cost.LogicalExpr{cost.Column("", "state")},
			[]cost.AggregateExpr{cost.Count(cost.Column("", "username"))}).
		// the alias for aggregate result is bit weird
		Project(cost.Column("", "state"), cost.Alias(cost.Count(cost.Column("", "username")), "num")).
		LogicalPlan()

	fmt.Println(cost.Format(plan, 0))
}
