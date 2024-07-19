package virtual_plan

import (
	"bytes"
	"context"
	"fmt"

	"github.com/kwilteam/kwil-db/internal/engine/cost/datasource"
	"github.com/kwilteam/kwil-db/internal/engine/cost/datatypes"
)

type VirtualPlan interface {
	fmt.Stringer

	// Schema returns the schema of the data that will be produced by this VirtualPlan.
	Schema() *datatypes.Schema
	Inputs() []VirtualPlan

	// Execute executes the plan and returns the result
	// This for testing purposes
	Execute(context.Context) *datasource.Result
	Statistics() *datatypes.Statistics
	Cost() int64
}

func Format(plan VirtualPlan, indent int) string {
	var msg bytes.Buffer
	for i := 0; i < indent; i++ {
		msg.WriteString(" ")
	}
	msg.WriteString(plan.String())
	msg.WriteString("\n")
	for _, child := range plan.Inputs() {
		msg.WriteString(Format(child, indent+2))
	}
	return msg.String()
}
