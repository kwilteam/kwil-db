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

func WithAccountStore(store *balances.AccountStore) DatasetUseCaseOpt {
	return func(u *DatasetUseCase) {
		u.accountStore = store
	}
}
