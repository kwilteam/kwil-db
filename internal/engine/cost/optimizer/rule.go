package optimizer

import (
	"github.com/kwilteam/kwil-db/internal/engine/cost/plantree"
)

type Rule interface {
	Transform(plan plantree.PlanNode) plantree.PlanNode
}

// RewriteRule is a rule that transforms a logical plan into an equivalent
// logical plan, with some additional optimizations.
type RewriteRule interface {
	Rule

	rewrite()
}

// ImplementRule is a rule that transforms a logical plan into a virtual plan.
type ImplementRule interface {
	Rule()

	impl()
}
