package manager

import (
	"context"
	"fmt"
	"kwil/x/sqlx/schema"
	"kwil/x/sqlx/sqlclient"
)

type ExecutionManager interface {
}

type Tx struct {
	query string
	args  []struct {
		name  string
		value string
	}
	caller string
	db     string
}

type executionManager struct {
	cache  *SchemaCache
	client *sqlclient.DB
}

func NewExecutionManager(cache *SchemaCache, client *sqlclient.DB) *executionManager {
	return &executionManager{
		cache:  cache,
		client: client,
	}
}

func (m *executionManager) Execute(ctx context.Context, tx *Tx) error {
	// Checking if the wallet has permission to execute the query
	ok, err := m.cache.WalletHasPermission(ctx, tx.caller, tx.db, tx.query)
	if err != nil {
		return fmt.Errorf("failed to check permission: %w", err)
	}
	if !ok {
		return fmt.Errorf("wallet %s does not have permission to execute query %s", tx.caller, tx.query)
	}

	executable, ok := m.cache.GetExecutable(tx.db, tx.query)
	if !ok {
		return fmt.Errorf("query %s not found", tx.query)
	}

	inputs := make(schema.UserInputs)
	// Executing the query
	for _, arg := range tx.args {
		inputs[arg.name] = arg.value
	}

	executableInputs, err := executable.PrepareInputs(&inputs)
	if err != nil {
		return fmt.Errorf("failed to prepare inputs: %w", err)
	}

	_, err = m.client.ExecContext(ctx, executable.Statement, executableInputs)

	return err
}
