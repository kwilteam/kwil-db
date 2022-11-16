package sqlschema_test

import (
	"fmt"
	"ksl/postgres"
	"ksl/schema"
	"ksl/sqlschema"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDiffer(t *testing.T) {
	fromData := ``

	toData := `
model Person {
	id     int     @id
	cars  Car[]
}

model Car {
	id     int     @id
	owners Person[]
}`

	from := schema.Parse([]byte(fromData), "from.ksl")
	to := schema.Parse([]byte(toData), "to.ksl")

	source := sqlschema.CalculateSqlSchema(from, "test")
	target := sqlschema.CalculateSqlSchema(to, "test")

	differ := sqlschema.NewDiffer(postgres.Backend{})
	steps, err := differ.Diff(source, target)

	require.NoError(t, err)

	planner := postgres.Planner{}
	plan, err := planner.Plan(sqlschema.Migration{Before: source, After: target, Changes: steps})
	require.NoError(t, err)

	for _, step := range plan.Statements {
		fmt.Fprintf(os.Stdout, "%s", step.String())
	}

	_ = steps
}
