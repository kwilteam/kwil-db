package virtual_plan

import (
	"bytes"
	"fmt"

	"github.com/kwilteam/kwil-db/internal/engine/cost/datasource"
)

type VirtualPlan interface {
	fmt.Stringer

	// Schema returns the schema of the data that will be produced by this VirtualPlan.
	Schema() *datasource.Schema
	Inputs() []VirtualPlan

	Execute() *datasource.Result
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
