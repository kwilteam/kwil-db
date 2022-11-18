package ksl_test

import (
	"context"
	"fmt"
	"ksl/ast"
	_ "ksl/postgres"
	"ksl/sqlclient"
	"ksl/sqldriver"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestKSL(t *testing.T) {
	data := `
model User {
	id int @id
	name string
	email string @unique
	aliases string[]

	posts Post[]
	comments Comment[]
}

model Post {
	id int @id
	title string
	content string
	published bool
	author_id int

	comments Comment[]
	author User @ref(fields: [author_id], references: [id])
}

model Comment {
	id int @id
	content string
	author_id int
	post_id int

	author User @ref(fields: [author_id], references: [id])
	post Post @ref(fields: [post_id], references: [id])
}`

	schemaAst := ast.Parse([]byte(data), "test.ksl")
	if schemaAst.HasErrors() {
		schemaAst.WriteDiagnostics(os.Stdout, true)
		t.FailNow()
	}

	client, err := sqlclient.Open("postgres://localhost:5432/postgres?sslmode=disable")
	require.NoError(t, err)

	ctx := context.Background()

	err = client.ExecuteInsert(ctx, sqldriver.InsertStatement{Database: "public", Table: "User", Input: map[string]any{"id": 1, "name": "Bryan", "email": "bryan@kwil.com"}})
	require.NoError(t, err)
	err = client.ExecuteInsert(ctx, sqldriver.InsertStatement{Database: "public", Table: "Post", Input: map[string]any{"id": 1, "title": "My New Blog", "content": "This is my new blog", "published": true, "author_id": 1}})
	require.NoError(t, err)
	err = client.ExecuteInsert(ctx, sqldriver.InsertStatement{Database: "public", Table: "Comment", Input: map[string]any{"id": 1, "content": "This is a comment", "author_id": 1, "post_id": 1}})
	require.NoError(t, err)

	err = client.ExecuteUpdate(ctx, sqldriver.UpdateStatement{Database: "public", Table: "User", Input: map[string]any{"name": "Matteson"}, Where: map[string]any{"id": 1}})
	require.NoError(t, err)

	results, err := client.ExecuteSelect(ctx, sqldriver.SelectStatement{Database: "public", Table: "User", Where: map[string]any{}})
	require.NoError(t, err)

	err = client.ExecuteDelete(ctx, sqldriver.DeleteStatement{Database: "public", Table: "Comment", Where: map[string]any{"id": 1}})
	require.NoError(t, err)
	err = client.ExecuteDelete(ctx, sqldriver.DeleteStatement{Database: "public", Table: "Post", Where: map[string]any{"id": 1}})
	require.NoError(t, err)
	err = client.ExecuteDelete(ctx, sqldriver.DeleteStatement{Database: "public", Table: "User", Where: map[string]any{"id": 1}})
	require.NoError(t, err)

	fmt.Printf("%v", results)

	// source, err := client.DescribeContext(ctx, "public")
	// require.NoError(t, err)

	// steps, err := client.Diff(source, target)
	// require.NoError(t, err)

	// migration := sqlmigrate.Migration{Before: source, After: target, Changes: steps}
	// plan, err := client.PlanContext(ctx, migration)
	// fmt.Fprintf(os.Stdout, "%s", plan)
	// require.NoError(t, err)

	// err = client.ApplyMigration(ctx, plan)
	// require.NoError(t, err)
}
