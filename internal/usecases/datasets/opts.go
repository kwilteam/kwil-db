package datasets

import (
	"github.com/kwilteam/kwil-db/pkg/balances"
	"github.com/kwilteam/kwil-db/pkg/engine"
	"github.com/kwilteam/kwil-db/pkg/extensions"
	"github.com/kwilteam/kwil-db/pkg/log"
)

type DatasetUseCaseOpt func(*DatasetUseCase)

func WithLogger(logger log.Logger) DatasetUseCaseOpt {
	return func(u *DatasetUseCase) {
		u.log = logger
	}
}

func WithAccountStore(store AccountStore) DatasetUseCaseOpt {
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

func WithSqliteFilePath(path string) DatasetUseCaseOpt {
	return func(u *DatasetUseCase) {
		u.sqliteFilePath = path
	}
}

func WithEngine(eng *engine.Engine) DatasetUseCaseOpt {
	return func(u *DatasetUseCase) {
		u.engine = eng
	}
}

func WithExtensions(extConfigs ...*extensions.ExtensionConfig) DatasetUseCaseOpt {
	return func(u *DatasetUseCase) {
		u.extensionConfigs = extConfigs
	}
}
