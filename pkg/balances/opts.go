package balances

import "github.com/kwilteam/kwil-db/pkg/log"

type balancesOpts func(*AccountStore)

func WithPath(path string) balancesOpts {
	return func(ar *AccountStore) {
		ar.path = path
	}
}

// Wipe will delete the database file and recreate it
func Wipe() balancesOpts {
	return func(ar *AccountStore) {
		ar.wipe = true
	}
}

func WithLogger(logger log.Logger) balancesOpts {
	return func(ar *AccountStore) {
		ar.log = logger
	}
}

func WithGasCosts(gas_enabled bool) balancesOpts {
	return func(ar *AccountStore) {
		ar.gas_enabled = gas_enabled
	}
}

func WithNonces(nonces_enabled bool) balancesOpts {
	return func(ar *AccountStore) {
		ar.nonces_enabled = nonces_enabled
	}
}
