package deposits

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	chainsync "kwil/internal/pkg/deposits/chain-sync"
	"kwil/internal/repository"
	chainClient "kwil/pkg/chain/client"
	"kwil/pkg/crypto"
	"kwil/pkg/log"
	"kwil/pkg/sql/sqlclient"
)

type Depositer interface {
	Start(ctx context.Context) error
}

// in the future we can make things like expirationPeriod and chunkSize configurable, but these values are good enough for now
type depositer struct {
	dao              repository.Queries
	db               *sqlclient.DB
	chain            chainsync.Chain
	log              log.Logger
	expirationPeriod int64
}

func NewDepositer(poolAddress string, db *sqlclient.DB, queries repository.Queries, chainClient chainClient.ChainClient, privateKey *ecdsa.PrivateKey, logger log.Logger) (Depositer, error) {
	// create the escrow contract
	escrowContract, err := chainClient.Contracts().Escrow(poolAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to create escrow contract: %w", err)
	}

	providerAddress := crypto.AddressFromPrivateKey(privateKey)

	// create the chain
	chainSynchronizer, err := chainsync.New(chainClient, escrowContract, queries, db, logger, providerAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to create chain synchronizer: %w", err)
	}

	return &depositer{
		dao:              queries,
		db:               db,
		chain:            chainSynchronizer,
		expirationPeriod: 100,
		log:              logger.Named("deposits-service"),
	}, nil
}
