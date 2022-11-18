package ksl_test

import (
	"context"
	"fmt"
	_ "ksl/postgres"
	"ksl/schema"
	"ksl/sqlclient"
	"ksl/sqlschema"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestKSL(t *testing.T) {
	data := `
// model Vehicle {
// 	id     int     @id
// 	owners User[]
// }

model User {
	id int @id
	name string?
	email string @unique
	site_id int

	site Site @ref(fields: [site_id], references: [id])
	posts Post[]
	// vehicles Vehicle[]
}

model Site {
	id int @id
	name string

	users User[]
	posts Post[]
}

model Post {
	id int @id
	title string
	author_id int
	site_id int

	site Site @ref(fields: [site_id], references: [id])
	author User @ref(fields: [author_id], references: [id])
}`

	ksch := schema.Parse([]byte(data), "test.ksl")
	if ksch.HasErrors() {
		ksch.WriteDiagnostics(os.Stdout, true)
		t.FailNow()
	}

	target := sqlschema.CalculateSqlSchema(ksch, "public")

	client, err := sqlclient.Open("postgres://localhost:5432/postgres?sslmode=disable")
	require.NoError(t, err)

	ctx := context.Background()

	source, err := client.DescribeContext(ctx, "public")
	require.NoError(t, err)

	steps, err := client.Diff(source, target)
	require.NoError(t, err)

	migration := sqlschema.Migration{Before: source, After: target, Changes: steps}
	plan, err := client.PlanContext(ctx, migration)
	fmt.Fprintf(os.Stdout, "%s", plan)
	require.NoError(t, err)

	err = client.ApplyMigration(ctx, plan)
	require.NoError(t, err)
}
