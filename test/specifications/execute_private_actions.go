package specifications

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func ExecutePrivateActionSpecification(ctx context.Context, t *testing.T, execute ExecuteActionsDsl) {
	t.Logf("Executing private action specification")

	db := SchemaLoader.Load(t, SchemaTestDB)
	dbID := execute.DBID(db.Name)

	id := int32(2833)
	postTitle := "test_post_private"
	postContent := "content for test_post_private"

	createPostActionInput := []any{id, postTitle, postContent}

	txHash, err := execute.Execute(ctx, dbID, "create_post_private", createPostActionInput)
	require.NoError(t, err, "error executing private action")

	expectTxFail(t, execute, ctx, txHash, defaultTxQueryTimeout)

	if hasUser(ctx, t, execute, dbID, id) {
		t.Fatalf("user %d should not exist", id)
	}

	// calling nested should work
	txHash, err = execute.Execute(ctx, dbID, "create_post_nested", createPostActionInput)
	require.NoError(t, err, "error executing nested action")

	expectTxSuccess(t, execute, ctx, txHash, defaultTxQueryTimeout)

	if !hasUser(ctx, t, execute, dbID, id) {
		t.Fatalf("user %d should exist", id)
	}
}

func hasUser(ctx context.Context, t *testing.T, execute ExecuteActionsDsl, dbid string, id int32) bool {
	records, err := execute.QueryDatabase(ctx, dbid, fmt.Sprintf("SELECT * FROM posts WHERE id = %d", id))
	require.NoError(t, err)
	require.NotNil(t, records)

	mapRecords := records.Export()
	return len(mapRecords) != 0
}
