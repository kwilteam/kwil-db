package executor

import (
	"context"
	"fmt"
	"kwil/internal/pkg/graphql/manager"
	"kwil/internal/repository"
	"kwil/pkg/databases"
	"kwil/pkg/databases/executables"
	"kwil/pkg/log"
	"kwil/pkg/sql/sqlclient"
	"kwil/pkg/types/data_types/any_type"
	"kwil/pkg/types/execution"
)

type Executor interface {
	DeployDatabase(ctx context.Context, database *databases.Database[anytype.KwilAny]) error
	DropDatabase(ctx context.Context, database *databases.DatabaseIdentifier) error
	ExecuteQuery(ctx context.Context, query *execution.ExecutionBody[anytype.KwilAny], caller string) error
	GetExecutables(id string) ([]*execution.Executable, error)
	GetDBIdentifier(id string) (*databases.DatabaseIdentifier, error)
}

type executor struct {
	hasura    manager.Client
	databases map[string]executables.ExecutablesInterface
	dao       repository.Queries
	db        *sqlclient.DB
	log       log.Logger
}

func NewExecutor(ctx context.Context, db *sqlclient.DB, queries repository.Queries, mngr manager.Client, logger log.Logger) (Executor, error) {
	exec := &executor{
		hasura:    mngr,
		dao:       queries,
		db:        db,
		databases: make(map[string]executables.ExecutablesInterface),
		log:       logger.Named(`executor`),
	}

	err := exec.loadExecutables(ctx)
	if err != nil {
		return nil, fmt.Errorf(`error loading executables: %w`, err)
	}

	return exec, nil
}
