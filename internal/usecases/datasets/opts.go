package datasets

import (
	"kwil/pkg/balances"
	"kwil/pkg/log"
)

type DatasetUseCaseOpt func(*DatasetUseCase)

func WithLogger(logger log.Logger) DatasetUseCaseOpt {
	return func(u *DatasetUseCase) {
		u.log = logger
	}
}

func WithAccountStore(store accountStore) DatasetUseCaseOpt {
	return func(u *DatasetUseCase) {
		u.accountStore = store
	}
}

// Warning: this will panic if the account store cannot be created
func WithTempAccountStore() DatasetUseCaseOpt {
	return func(u *DatasetUseCase) {
		var err error
		u.accountStore, err = balances.NewAccountStore(
			balances.WithPath("tmp/accounts/"),
			balances.Wipe(),
		)
		if err != nil {
			panic(err)
		}
	}
}

func WithEngine(engine engineInterface) DatasetUseCaseOpt {
	return func(u *DatasetUseCase) {
		u.engine = engine
	}
}

func WithSqliteFilePath(path string) DatasetUseCaseOpt {
	return func(u *DatasetUseCase) {
		u.sqliteFilePath = path
	}
}
