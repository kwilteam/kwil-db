package datasets

import (
	"context"
	"fmt"

	"github.com/kwilteam/kwil-db/core/types/transactions"
	engineTypes "github.com/kwilteam/kwil-db/internal/engine/types"
	"github.com/kwilteam/kwil-db/internal/ident"
	"github.com/kwilteam/kwil-db/internal/sql"
)

// Call executes a call action on a database.  It is a read-only action.
// It returns the result of the call.
// If a message caller is specified, then it will check the signature of the message and use the caller as the caller of the action.
func (u *DatasetModule) Call(ctx context.Context, dbid string, action string, args []any, msg *transactions.CallMessage) ([]map[string]any, error) {
	var sender string
	var err error

	if len(msg.Sender) > 0 {
		if msg.AuthType != "" {
			sender, err = ident.Identifier(msg.AuthType, msg.Sender)
			if err != nil {
				return nil, fmt.Errorf("failed to get sender identifier: %w", err)
			}
		} else {
			// auth_type is the http payload field name
			return nil, fmt.Errorf("auth_type is required")
		}
	}

	results, err := u.engine.Execute(ctx, &engineTypes.ExecutionData{
		Dataset:          dbid,
		Procedure:        action,
		Mutative:         false,
		Args:             args,
		Caller:           msg.Sender,
		CallerIdentifier: sender,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute call: %w", err)
	}

	return getResultMap(results), nil
}

// Query executes an ad-hoc query on a database.  It is a read-only action.
// It returns the result of the query.
func (u *DatasetModule) Query(ctx context.Context, dbid string, query string) ([]map[string]any, error) {
	results, err := u.engine.Query(ctx, dbid, query)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	return getResultMap(results), nil
}

// GetSchema returns the schema of a database.
func (u *DatasetModule) GetSchema(ctx context.Context, dbid string) (*engineTypes.Schema, error) {
	return u.engine.GetSchema(ctx, dbid)
}

// ListOwnedDatabase returns a list of databases owned by a public key.
func (u *DatasetModule) ListOwnedDatabases(ctx context.Context, owner []byte) ([]string, error) {
	return u.engine.ListDatasets(ctx, owner)
}

func getResultMap(results *sql.ResultSet) []map[string]any {
	resMap := make([]map[string]any, 0)
	for _, result := range results.Rows {
		res := make(map[string]any)
		for i, column := range results.Columns {
			res[column] = result[i]
		}

		resMap = append(resMap, res)
	}

	return resMap
}
