package token

import "kwil/pkg/log"

type TokenOpts func(*token)

func WithLogger(logger log.Logger) TokenOpts {
	return func(t *token) {
		t.log = logger
	}
}
