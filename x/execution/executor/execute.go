package executor

import (
	"context"
	"fmt"
	"kwil/x/types/databases"
	execTypes "kwil/x/types/execution"
)

func (s *executor) ExecuteQuery(ctx context.Context, body *execTypes.ExecutionBody) error {
	db, ok := s.databases[body.Database]
	if !ok {
		return fmt.Errorf("database %s not found", body.Database)
	}

	// check if user can execute
	if !db.CanExecute(body.Caller, body.Query) {
		return fmt.Errorf("user %s cannot execute %s", body.Caller, body.Query)
	}

	// prepare query
	args, err := db.Prepare(body.Query, body.Caller, body.Inputs)
	if err != nil {
		return fmt.Errorf("failed to prepare query: %w", err)
	}

	// execute query
	_, err = s.db.ExecContext(ctx, body.Query, args...)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}

	return nil
}

func (s *executor) GetExecutables(id string) ([]*execTypes.Executable, error) {
	execInterface, ok := s.databases[id]
	if !ok {
		return nil, fmt.Errorf("database id %s not found", id)
	}
	return execInterface.ListExecutables(), nil
}

func (s *executor) GetDBIdentifier(id string) (*databases.DatabaseIdentifier, error) {
	db, ok := s.databases[id]
	if !ok {
		return nil, fmt.Errorf("database id %s not found", id)
	}
	return db.GetIdentifier(), nil
}
