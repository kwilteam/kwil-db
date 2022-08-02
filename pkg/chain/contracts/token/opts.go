package token

import "github.com/kwilteam/kwil-db/pkg/log"

type TokenOpts func(*token)

func WithLogger(logger log.Logger) TokenOpts {
	return func(t *token) {
		t.log = logger
	}
}
