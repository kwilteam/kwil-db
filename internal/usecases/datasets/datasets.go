package datasets

import (
	"fmt"
	"kwil/pkg/balances"
	"kwil/pkg/engine"
	"kwil/pkg/engine/models"
	"kwil/pkg/log"
)

type DatasetUseCase struct {
	engine       *engine.Engine
	accountStore *balances.AccountStore
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
	u.engine, err = engine.Open(
		engine.WithLogger(u.log),
	)
	if err != nil {
		return nil, err
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
	db, ok := u.engine.Datasets[dbid]
	if !ok {
		return nil, fmt.Errorf("dataset not found")
	}

	return db.GetSchema(), nil
}
