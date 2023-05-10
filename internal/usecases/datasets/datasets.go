package datasets

import (
	"github.com/kwilteam/kwil-db/pkg/balances"
	"github.com/kwilteam/kwil-db/pkg/engine"
	"github.com/kwilteam/kwil-db/pkg/engine/models"
	"github.com/kwilteam/kwil-db/pkg/log"
)

type DatasetUseCase struct {
	engine       engineInterface
	accountStore accountStore
	log          log.Logger

	sqliteFilePath string
}

func New(opts ...DatasetUseCaseOpt) (*DatasetUseCase, error) {
	u := &DatasetUseCase{
		log:            log.NewNoOp(),
		sqliteFilePath: "",
	}

	for _, opt := range opts {
		opt(u)
	}

	var err error
	if u.engine == nil {
		u.engine, err = engine.Open(
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

func (u *DatasetUseCase) engineOpts() []engine.MasterOpt {
	opts := make([]engine.MasterOpt, 0)
	if u.sqliteFilePath != "" {
		opts = append(opts, engine.WithPath(u.sqliteFilePath))
	}
	opts = append(opts, engine.WithLogger(u.log))

	return opts
}

func (u *DatasetUseCase) ListDatabases(owner string) ([]string, error) {
	return u.engine.ListDatabases(owner)
}

func (u *DatasetUseCase) GetSchema(dbid string) (*models.Dataset, error) {
	db, err := u.engine.GetDataset(dbid)
	if err != nil {
		return nil, err
	}

	return db.GetSchema(), nil
}

func (u *DatasetUseCase) Close() error {
	u.accountStore.Close()
	return u.engine.Close()
}
