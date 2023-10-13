package sessions

import "github.com/kwilteam/kwil-db/core/log"

type CommiterOpt func(*AtomicCommitter)

// WithLogger sets the logger for the session.
func WithLogger(logger log.Logger) CommiterOpt {
	return func(a *AtomicCommitter) {
		a.log = logger
	}
}
