package txsvc

import (
	"kwil/pkg/balances"
	"kwil/pkg/log"
)

type TxSvcOpt func(*Service)

func WithLogger(logger log.Logger) TxSvcOpt {
	return func(s *Service) {
		s.log = logger
	}
}

func WithAccountStore(store *balances.AccountStore) TxSvcOpt {
	return func(s *Service) {
		s.accountStore = store
	}
}
