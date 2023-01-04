package service

import (
	"context"
	"fmt"
	"kwil/x/execution/dto"
)

func (s *executionService) Execute(ctx context.Context, body *dto.ExecutionBody) error {
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
