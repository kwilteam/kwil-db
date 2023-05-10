package specifications

import (
	"context"
	"github.com/kwilteam/kwil-db/pkg/engine/models"
	"testing"

	"github.com/stretchr/testify/assert"
)

type SwitchWalletDsl interface {
	// ExecuteAction executes QUERY to a database
	// @yaiba TODO: owner is not needed?? because user can only execute queries using his private key
}

const (
	createUserQueryName = "create_user"
	listUsersQueryName  = "list_users"
)

func ExecutePermissionedActionSpecification(ctx context.Context, t *testing.T, execute ExecuteQueryDsl) {
	// create_user is public, list_users is private
	t.Logf("Executing permissioned action specification")

	db := SchemaLoader.Load(t)
	dbID := models.GenerateSchemaId(db.Owner, db.Name)

	createUserQueryInputs := []map[string]any{
		{
			"$id":       5729,
			"$username": "test_user",
			"$age":      102,
		},
	}

	_, _, err := execute.ExecuteAction(ctx, dbID, createUserQueryName, createUserQueryInputs)
	assert.NoError(t, err)

	// list_users is private, so it should fail
	_, _, err = execute.ExecuteAction(ctx, dbID, listUsersQueryName, nil)
	assert.Error(t, err)

	// adhoc query should fail
	_, err = execute.QueryDatabase(ctx, dbID, "SELECT * FROM users")
	assert.NoError(t, err)
}
