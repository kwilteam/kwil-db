package datasets

import (
	"context"

	"github.com/kwilteam/kwil-db/internal/entity"
	"github.com/kwilteam/kwil-db/pkg/balances"
	"github.com/kwilteam/kwil-db/pkg/engine2"
	"github.com/kwilteam/kwil-db/pkg/log"
)

type DatasetUseCase struct {
	engine       engine2.Engine
	accountStore AccountStore
	log          log.Logger

	sqliteFilePath string
}

func New(ctx context.Context, opts ...DatasetUseCaseOpt) (DatasetUseCaseInterface, error) {
	u := &DatasetUseCase{
		log:            log.NewNoOp(),
		sqliteFilePath: "",
	}

	for _, opt := range opts {
		opt(u)
	}

	var err error
	if u.engine == nil {
		u.engine, err = engine2.Open(ctx,
			u.engineOpts()...,
		)
		if err != nil {
			return nil, err
		}
	}

	if u.accountStore == nil {
		u.accountStore, err = balances.NewAccountStore(
			balances.WithLogger(u.log),
		)
		if err != nil {
			return nil, err
		}
	}

	return u, nil
}

func (u *DatasetUseCase) engineOpts() []engine2.EngineOpt {
	opts := []engine2.EngineOpt{
		engine2.WithLogger(u.log),
	}
	if u.sqliteFilePath != "" {
		opts = append(opts, engine2.WithPath(u.sqliteFilePath))
	}

	return opts
}

func (u *DatasetUseCase) ListDatabases(ctx context.Context, owner string) ([]string, error) {
	return u.engine.ListDatasets(ctx, owner)
}

func (u *DatasetUseCase) GetSchema(dbid string) (*entity.Schema, error) {
	db, err := u.engine.GetDataset(dbid)
	if err != nil {
		return nil, err
	}

	actions := db.ListActions()
	tables := db.ListTables()

	return &entity.Schema{
		Owner:   db.Owner(),
		Name:    db.Name(),
		Actions: convertActions(actions),
		Tables:  convertTables(tables),
	}, nil
}

func (u *DatasetUseCase) Close() error {
	u.accountStore.Close()
	return u.engine.Close(true)
}
