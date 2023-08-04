package abci

import "github.com/kwilteam/kwil-db/pkg/log"

type AbciOpt func(*AbciApp)

func WithLogger(logger log.Logger) AbciOpt {
	return func(a *AbciApp) {
		a.log = logger
	}
}
