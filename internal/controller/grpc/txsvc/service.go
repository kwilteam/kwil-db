package txsvc

import (
	txpb "kwil/api/protobuf/tx/v1"
	"kwil/internal/app/kwild/config"
	"kwil/internal/usecases/datasets"
	"kwil/pkg/log"
)

type Service struct {
	txpb.UnimplementedTxServiceServer

	log log.Logger

	executor datasets.DatasetUseCaseInterface
	cfg      *config.AppConfig
}

func NewService(config *config.AppConfig, opts ...TxSvcOpt) (*Service, error) {
	s := &Service{
		log: log.NewNoOp(),
		cfg: config,
	}

	for _, opt := range opts {
		opt(s)
	}

	var err error
	s.executor, err = datasets.New(
		datasets.WithLogger(s.log),
	)
	if err != nil {
		return nil, err
	}

	return s, nil
}
