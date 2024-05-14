package txsvc

import (
	"time"

	"github.com/kwilteam/kwil-db/core/log"
)

type TxSvcOpt func(*Service)

func WithLogger(logger log.Logger) TxSvcOpt {
	return func(s *Service) {
		s.log = logger
	}
}

// WithReadTxTimeout sets a timeout for read-only DB transactions, as used by
// the Query and Call methods of Service.
func WithReadTxTimeout(timeout time.Duration) TxSvcOpt {
	return func(s *Service) {
		s.readTxTimeout = timeout
	}
}
