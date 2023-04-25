package escrow

import "kwil/pkg/log"

type EscrowOpts func(*escrow)

func WithLogger(logger log.Logger) EscrowOpts {
	return func(e *escrow) {
		e.log = logger
	}
}
