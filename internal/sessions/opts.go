package sessions

import "github.com/kwilteam/kwil-db/core/log"

type CommitterOpt func(*MultiCommitter)

// WithLogger sets the logger to use for the committer.
func WithLogger(logger log.Logger) CommitterOpt {
	return func(c *MultiCommitter) {
		c.log = logger
	}
}
