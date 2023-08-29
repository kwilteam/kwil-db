package specifications

import (
	"context"
	"testing"

	"github.com/cstockton/go-conv"
	"github.com/kwilteam/kwil-db/pkg/client"
)

type ExecuteCallDsl interface {
	DatabaseIdentifier
	Call(ctx context.Context, dbid, action string, inputs []any, opts ...client.CallOpt) ([]map[string]any, error)
}

func ExecuteCallSpecification(ctx context.Context, t *testing.T, caller ExecuteCallDsl) {
	t.Logf("Executing ExecuteCallSpecification")

	db := SchemaLoader.Load(t, schemaTestDB)
	dbID := caller.DBID(db.Name)

	getPostInput := []any{
		[]any{1111},
	}

	results, err := caller.Call(ctx, dbID, "get_post_authenticated", getPostInput)
	if err != nil {
		t.Fatalf("error calling action: %s", err.Error())
	}
	checkGetPostResults(t, results)

	_, err = caller.Call(ctx, dbID, "get_post_authenticated", getPostInput, client.Authenticated(false))
	if err == nil {
		t.Errorf("expected error calling action without authentication")
	}

	results, err = caller.Call(ctx, dbID, "get_post_unauthenticated", getPostInput, client.Authenticated(false))
	if err != nil {
		t.Fatalf("error calling action: %s", err.Error())
	}
	checkGetPostResults(t, results)
}

func checkGetPostResults(t *testing.T, results []map[string]any) {
	if len(results) != 1 {
		t.Fatalf("expected 1 statement result, got %d", len(results))
	}

	returnedPost := results[0]

	postId, _ := conv.Int32(returnedPost["id"])

	if postId != 1111 {
		t.Errorf("expected post id to be 1111, got %d", postId)
	}
}
