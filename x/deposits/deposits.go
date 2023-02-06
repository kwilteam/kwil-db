package deposits

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"kwil/kwil/repository"
	chainClient "kwil/pkg/chain/client"
	"kwil/pkg/fund"
	"kwil/pkg/log"
	"kwil/pkg/sql/sqlclient"
	"kwil/x/contracts/escrow"
	chainsync "kwil/x/deposits/chain-sync"
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

func NewDepositer(config *fund.Config, db *sqlclient.DB, queries repository.Queries, chainClient chainClient.ChainClient, privateKey *ecdsa.PrivateKey, logger log.Logger) (Depositer, error) {

	// create the escrow contract
	escrowContract, err := escrow.New(chainClient, privateKey, config.PoolAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to create escrow contract: %w", err)
	}

	// create the chain
	chainSynchronizer, err := chainsync.New(chainClient, escrowContract, queries, db, logger)
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
