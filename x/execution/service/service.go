package service

import (
	"context"
	"kwil/x/execution/dto"
	"kwil/x/execution/executables"
	"kwil/x/execution/repository"
	"kwil/x/sqlx/sqlclient"
)

type ExecutionService interface {
	DeployDatabase(ctx context.Context, database *dto.Database) error
	DropDatabase(ctx context.Context, database *dto.DatabaseIdentifier) error
	ExecuteQuery(ctx context.Context, query *dto.ExecutionBody) error
}

type executionService struct {
	databases map[string]executables.ExecutablesInterface
	dao       *repository.Queries
	db        *sqlclient.DB
}

func NewExecutionService(db *sqlclient.DB) (ExecutionService, error) {
	ctx := context.Background() // sqlc only allows prepared queries with context
	pq, err := repository.Prepare(ctx, db)
	if err != nil {
		return nil, err
	}
	return &executionService{
		dao:       pq,
		db:        db,
		databases: make(map[string]executables.ExecutablesInterface),
	}, nil
}
