package datasets

import (
	"kwil/pkg/balances"
	"kwil/pkg/engine"
	"kwil/pkg/engine/models"
	"kwil/pkg/log"
)

type DatasetUseCase struct {
	engine       engineInterface
	accountStore accountStore
	log          log.Logger
}

func New(opts ...DatasetUseCaseOpt) (*DatasetUseCase, error) {
	u := &DatasetUseCase{
		log: log.NewNoOp(),
	}

	for _, opt := range opts {
		opt(u)
	}

	var err error
	if u.engine == nil {
		u.engine, err = engine.Open(
			engine.WithLogger(u.log),
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
