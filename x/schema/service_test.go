package schema

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestService(t *testing.T) {
	service := NewTestService()
	ctx := context.Background()

	data := `
	table posts {
		id:        int64     @id
		title:     string?   @size(256)
		content:   string?
		published: bool      @default(true)
		author_id: int64     @foreign_key(users.id, on_delete="SET NULL")
	}
	table users {
		id:         int64    @id
		email:      string   @unique
		name:       string?
		dob:        date?
		aliases:    string[]
		created_at: datetime
		updated_at: datetime
	}
	table vehicles {
		id:       int64  @id
		name:     string
		owner_id: int64  @foreign_key(users.id, on_delete="SET NULL")
	}
`
	req := PlanRequest{Wallet: "postgres", Database: "public", SchemaData: []byte(data)}
	plan, err := service.Plan(ctx, req)
	require.NoError(t, err)
	require.NotEmpty(t, plan.Changes)

	err = service.Apply(ctx, plan.ID)
	require.NoError(t, err)

	db, err := service.GetDatabase(ctx, "postgres", "public")
	require.NoError(t, err)
	require.NotNil(t, db)

	plan, err = service.Plan(ctx, req)
	require.NoError(t, err)
	require.Empty(t, plan.Changes)
}
