package executor

import (
	"context"
	"fmt"
	"kwil/internal/pkg/graphql/manager"
	"kwil/kwil/repository"
	"kwil/pkg/logger"
	"kwil/pkg/sql/sqlclient"
	"kwil/x/execution/executables"
	anytype "kwil/x/types/data_types/any_type"
	"kwil/x/types/databases"
	execTypes "kwil/x/types/execution"
)

type Executor interface {
	DeployDatabase(ctx context.Context, database *databases.Database[anytype.KwilAny]) error
	DropDatabase(ctx context.Context, database *databases.DatabaseIdentifier) error
	ExecuteQuery(ctx context.Context, query *execTypes.ExecutionBody[anytype.KwilAny], caller string) error
	GetExecutables(id string) ([]*execTypes.Executable, error)
	GetDBIdentifier(id string) (*databases.DatabaseIdentifier, error)
}

type executor struct {
	hasura    manager.Client
	databases map[string]executables.ExecutablesInterface
	dao       repository.Queries
	db        *sqlclient.DB
	log       logger.Logger
}

func NewExecutor(ctx context.Context, db *sqlclient.DB, queries repository.Queries, mngr manager.Client) (Executor, error) {
	exec := &executor{
		hasura:    mngr,
		dao:       queries,
		db:        db,
		databases: make(map[string]executables.ExecutablesInterface),
		log:       logger.New().Named(`executor`),
	}

	err := exec.loadExecutables(ctx)
	if err != nil {
		return nil, fmt.Errorf(`error loading executables: %w`, err)
	}

	return exec, nil
}
