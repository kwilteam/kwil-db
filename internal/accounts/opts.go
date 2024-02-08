package accounts

import "github.com/kwilteam/kwil-db/core/log"

type AccountStoreOpts func(*AccountStore)

func WithLogger(logger log.Logger) AccountStoreOpts {
	return func(ar *AccountStore) {
		ar.log = logger
	}
}

func WithGasCosts(gas_enabled bool) AccountStoreOpts {
	return func(ar *AccountStore) {
		ar.gasEnabled = gas_enabled
	}
}
