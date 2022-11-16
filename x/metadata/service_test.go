package metadata

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestService(t *testing.T) {
	service := NewTestService()
	ctx := context.Background()

	data := `
model Post {
	id        bigint     @id
	title     string?
	content   string?
	published bool
	author_id bigint?

	author    User?     @ref(fields: [author_id], references: [id])
}

model User {
	id         bigint    @id
	email      string   @unique
	name       string?
	dob        date?
	aliases    string[]
	created_at datetime
	updated_at datetime

	posts      Post[]
	vehicles   Vehicle[]
}

model Vehicle {
	id       bigint  @id
	name     string
	owner_id bigint

	owner    User    @ref(fields: [owner_id], references: [id])
}
`
	req := PlanRequest{Wallet: "postgres", Database: "public", SchemaData: []byte(data)}
	plan, err := service.Plan(ctx, req)
	require.NoError(t, err)
	require.NotEmpty(t, plan.Changes)

	err = service.Apply(ctx, plan.ID)
	require.NoError(t, err)

	db, err := service.GetMetadata(ctx, RequestMetadata{"postgres", "public"})
	require.NoError(t, err)
	require.NotNil(t, db)

	plan, err = service.Plan(ctx, req)
	require.NoError(t, err)
	require.Empty(t, plan.Changes)
}
