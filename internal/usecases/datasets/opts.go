package datasets

import (
	"github.com/kwilteam/kwil-db/pkg/engine"
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

func WithExtensions(extUrls ...string) DatasetUseCaseOpt {
	return func(u *DatasetUseCase) {
		u.extensionUrls = extUrls
	}
}
