package txsvc

import (
	"context"

	txpb "github.com/kwilteam/kwil-db/api/protobuf/tx/v1"
	"github.com/kwilteam/kwil-db/internal/app/kwild/config"
	"github.com/kwilteam/kwil-db/internal/usecases/datasets"
	"github.com/kwilteam/kwil-db/pkg/crypto"
	"github.com/kwilteam/kwil-db/pkg/log"
)

type Service struct {
	txpb.UnimplementedTxServiceServer

	log log.Logger

	executor       datasets.DatasetUseCaseInterface
	cfg            *config.KwildConfig
	accountStore   datasets.AccountStore
	sqliteFilePath string
	extensionUrls  []string

	providerAddress string
}

func NewService(ctx context.Context, config *config.KwildConfig, opts ...TxSvcOpt) (*Service, error) {
	s := &Service{
		log:             log.NewNoOp(),
		cfg:             config,
		providerAddress: crypto.AddressFromPrivateKey(config.PrivateKey),
		extensionUrls:   []string{},
	}

	for _, opt := range opts {
		opt(s)
	}

	dataSetOpts := getDatasetUseCaseOpts(s)

	var err error
	s.executor, err = datasets.New(ctx,
		dataSetOpts...,
	)
	if err != nil {
		return nil, err
	}

	return s, nil
}

func getDatasetUseCaseOpts(s *Service) []datasets.DatasetUseCaseOpt {
	opts := make([]datasets.DatasetUseCaseOpt, 0)
	if s.accountStore != nil {
		// if an account store is provided, use it
		// otherwise, the dataset use case will create its own
		opts = append(opts, datasets.WithAccountStore(s.accountStore))
	}
	if s.sqliteFilePath != "" {
		// if a sqlite file path is provided, use it
		// otherwise, the dataset use case will create its own
		opts = append(opts, datasets.WithSqliteFilePath(s.sqliteFilePath))
	}

	if len(s.extensionUrls) > 0 {
		opts = append(opts, datasets.WithExtensions(s.extensionUrls...))
	}

	opts = append(opts, datasets.WithLogger(s.log))
	return opts
}
