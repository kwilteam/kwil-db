package executor

import (
	"context"
	"fmt"
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
	dao       repository.Queries
	db        *sqlclient.DB
}

func NewExecutor(ctx context.Context, db *sqlclient.DB, queries repository.Queries, mngr manager.Client) (Executor, error) {
	exec := &executor{
		hasura:    mngr,
		dao:       queries,
		db:        db,
		databases: make(map[string]executables.ExecutablesInterface),
	}

	err := exec.loadExecutables(ctx)
	if err != nil {
		return nil, fmt.Errorf(`error loading executables: %w`, err)
	}

	return exec, nil
}
