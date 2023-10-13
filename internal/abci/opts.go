package abci

import "github.com/kwilteam/kwil-db/core/log"

type AbciOpt func(*AbciApp)

func WithLogger(logger log.Logger) AbciOpt {
	return func(a *AbciApp) {
		a.log = logger
	}
}

func WithApplicationVersion(version uint64) AbciOpt {
	return func(a *AbciApp) {
		a.applicationVersion = version
	}
}
