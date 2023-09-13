package specifications

import (
	"context"
	"fmt"
	"testing"

	"github.com/cstockton/go-conv"
	"github.com/stretchr/testify/assert"
)

func ExecuteDBDeleteSpecification(ctx context.Context, t *testing.T, execute ExecuteQueryDsl) {
	t.Logf("Executing delete action specification")
	// Given a valid database schema
	db := SchemaLoader.Load(t, SchemaTestDB)
	dbID := execute.DBID(db.Name)

	actionName := "delete_user_by_id"

	// get user id
	res, err := execute.QueryDatabase(ctx, dbID, "SELECT * FROM users")
	assert.NoError(t, err)

	records := res.Export()
	assert.NoError(t, err)

	if len(records) == 0 {
		t.Errorf("must have at least 1 user to test delete specification")
	}

	user1Id, err := conv.Int32(records[0]["id"])
	assert.NoError(t, err)

	txHash, err := execute.ExecuteAction(ctx, dbID, actionName, []any{user1Id})
	assert.NoError(t, err)

	expectTxSuccess(t, execute, ctx, txHash, defaultTxQueryTimeout)()

	// check that user is deleted
	res, err = execute.QueryDatabase(ctx, dbID, fmt.Sprintf("SELECT * FROM users WHERE id = %d", user1Id))
	assert.NoError(t, err)

	records = res.Export()
	assert.NoError(t, err)

	if len(records) != 0 {
		t.Errorf("expected 0 user statement result, got %d", len(records))
	}
}
