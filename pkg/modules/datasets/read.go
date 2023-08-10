package datasets

import (
	"context"
	"fmt"

	"github.com/kwilteam/kwil-db/pkg/engine"
	engineTypes "github.com/kwilteam/kwil-db/pkg/engine/types"
	transaction "github.com/kwilteam/kwil-db/pkg/tx"
)

// Call executes a call action on a database.  It is a read-only action.
// It returns the result of the call.
// If a message caller is specified, then it will check the signature of the message and use the caller as the caller of the action.
func (u *DatasetModule) Call(ctx context.Context, call *transaction.CallActionPayload, msg *transaction.SignedMessage[transaction.JsonPayload]) ([]map[string]any, error) {
	executionOpts := []engine.ExecutionOpt{
		engine.ReadOnly(true),
	}
	if msg.Sender != "" {
		err := msg.Verify()
		if err != nil {
			return nil, fmt.Errorf(`%w: failed to verify signed message: %s`, ErrAuthenticationFailed, err.Error())
		}

		executionOpts = append(executionOpts, engine.WithCaller(msg.Sender))
	}

	params := []map[string]any{}
	if call.Params != nil {
		params = append(params, call.Params)
	} else {
		params = append(params, map[string]any{})
	}

	return u.engine.Execute(ctx, call.DBID, call.Action, params, executionOpts...)
}

// Query executes an ad-hoc query on a database.  It is a read-only action.
// It returns the result of the query.
func (u *DatasetModule) Query(ctx context.Context, dbid string, query string) ([]map[string]any, error) {
	return u.engine.Query(ctx, dbid, query)
}

// GetSchema returns the schema of a database.
func (u *DatasetModule) GetSchema(ctx context.Context, dbid string) (*engineTypes.Schema, error) {
	return u.engine.GetSchema(ctx, dbid)
}

// ListOwnedDatabase returns a list of databases owned by an account.
func (u *DatasetModule) ListOwnedDatabases(ctx context.Context, owner string) ([]string, error) {
	return u.engine.ListDatasets(ctx, owner)
}
