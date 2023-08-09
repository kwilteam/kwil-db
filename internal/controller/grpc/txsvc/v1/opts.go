package txsvc

import (
	"github.com/kwilteam/kwil-db/pkg/log"
)

type TxSvcOpt func(*Service)

func WithLogger(logger log.Logger) TxSvcOpt {
	return func(s *Service) {
		s.log = logger
	}
}
