package executor

import (
	"context"
	"fmt"
	"kwil/internal/pkg/graphql"
	"kwil/internal/repository"
	"kwil/pkg/databases"
	"kwil/pkg/databases/executables"
	"kwil/pkg/databases/spec"
	"kwil/pkg/log"
	"kwil/pkg/sql/sqlclient"
)

type Executor interface {
	DeployDatabase(ctx context.Context, database *databases.Database[*spec.KwilAny]) error
	DropDatabase(ctx context.Context, database *databases.DatabaseIdentifier) error
	ExecuteQuery(ctx context.Context, query *executables.ExecutionBody, caller string) error
	GetQueries(id string) ([]*executables.QuerySignature, error)
	GetDBIdentifier(id string) (*databases.DatabaseIdentifier, error)
}

type executor struct {
	hasura    graphql.Client
	databases map[string]*executables.DatabaseInterface
	dao       repository.Queries
	db        *sqlclient.DB
	log       log.Logger
}

func NewExecutor(ctx context.Context, db *sqlclient.DB, queries repository.Queries, mngr graphql.Client, logger log.Logger) (Executor, error) {
	exec := &executor{
		hasura:    mngr,
		dao:       queries,
		db:        db,
		databases: make(map[string]*executables.DatabaseInterface),
		log:       logger.Named(`executor`),
	}

	err := exec.loadExecutables(ctx)
	if err != nil {
		return nil, fmt.Errorf(`error loading executables: %w`, err)
	}

	return exec, nil
}
