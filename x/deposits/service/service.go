package service

import (
	"context"
	"fmt"
	ccDTO "kwil/x/chain/client/dto"
	"kwil/x/contracts/escrow"
	"kwil/x/deposits/dto"
	"kwil/x/deposits/repository"
	chainsync "kwil/x/deposits/service/chain-sync"
	"kwil/x/logx"
	"kwil/x/sqlx/sqlclient"
)

type DepositsService interface {
	Sync(ctx context.Context) error
	startWithdrawal(ctx context.Context, withdrawal dto.StartWithdrawal) error
}

// in the future we can make things like expirationPeriod and chunkSize configurable, but these values are good enough for now
type depositsService struct {
	dao              *repository.Queries
	db               *sqlclient.DB
	chain            chainsync.Chain
	log              logx.SugaredLogger
	expirationPeriod int64
}

func NewService(config dto.Config, db *sqlclient.DB, chainClient ccDTO.ChainClient, privateKey string) (DepositsService, error) {

	// create the escrow contract
	escrowContract, err := escrow.New(chainClient, privateKey, config.EscrowAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to create escrow contract: %w", err)
	}

	reposit := repository.New(db)

	// create the chain
	chainSynchronizer, err := chainsync.New(chainClient, escrowContract, reposit, db)
	if err != nil {
		return nil, fmt.Errorf("failed to create chain synchronizer: %w", err)
	}

	return &depositsService{
		dao:              reposit,
		db:               db,
		chain:            chainSynchronizer,
		expirationPeriod: 100,
		log:              logx.New().Named("deposits-service").Sugar(),
	}, nil
}
