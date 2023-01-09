package executor

import (
	"context"
	"kwil/kwil/repository"
	"kwil/x/execution/executables"
	"kwil/x/graphql/manager"
	"kwil/x/sqlx/sqlclient"
	"kwil/x/types/databases"
	execTypes "kwil/x/types/execution"
)

type Executor interface {
	DeployDatabase(ctx context.Context, database *databases.Database) error
	DropDatabase(ctx context.Context, database *databases.DatabaseIdentifier) error
	ExecuteQuery(ctx context.Context, query *execTypes.ExecutionBody) error
}

type executor struct {
	hasura    manager.Client
	databases map[string]executables.ExecutablesInterface
	dao       *repository.Queries
	db        *sqlclient.DB
}

func NewExecutor(db *sqlclient.DB, queries *repository.Queries, mngr manager.Client) (Executor, error) {
	return &executor{
		hasura:    mngr,
		dao:       queries,
		db:        db,
		databases: make(map[string]executables.ExecutablesInterface),
	}, nil
}
