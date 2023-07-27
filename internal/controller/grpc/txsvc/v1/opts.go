package txsvc

import (
	"github.com/kwilteam/kwil-db/internal/usecases/datasets"
	"github.com/kwilteam/kwil-db/pkg/log"
	kTx "github.com/kwilteam/kwil-db/pkg/tx"
)

type TxSvcOpt func(*Service)

func WithLogger(logger log.Logger) TxSvcOpt {
	return func(s *Service) {
		s.log = logger
	}
}

func WithAccountStore(store datasets.AccountStore) TxSvcOpt {
	return func(s *Service) {
		s.accountStore = store
	}
}

func WithSqliteFilePath(path string) TxSvcOpt {
	return func(s *Service) {
		s.sqliteFilePath = path
	}
}

func WithExtensions(extConfigs ...string) TxSvcOpt {
	return func(s *Service) {
		s.extensionUrls = extConfigs
	}
}

func WithTxHook(hook func(*kTx.Transaction) error) TxSvcOpt {
	return func(s *Service) {
		s.txHook = hook
	}
}
