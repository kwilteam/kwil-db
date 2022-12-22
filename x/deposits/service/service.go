package service

import (
	"context"
	"fmt"
	"kwil/x/cfgx"
	chainClient "kwil/x/chain/client/service"
	chainProviderDTO "kwil/x/chain/provider/dto"
	"kwil/x/contracts/escrow"
	"kwil/x/deposits/dto"
	"kwil/x/deposits/repository"
	chainsync "kwil/x/deposits/service/chain-sync"
	"kwil/x/logx"
	"kwil/x/sqlx/sqlclient"
	"sync"
)

type DepositsService interface {
	Spend(ctx context.Context, spend dto.Spend) error
	GetBalancesAndSpent(ctx context.Context, wallet string) (*dto.Balance, error)
	Deposit(ctx context.Context, deposit dto.Deposit) error
	startWithdrawal(ctx context.Context, withdrawal dto.StartWithdrawal) error
}

// in the future we can make things like expirationPeriod and chunkSize configurable, but these values are good enough for now
type depositsService struct {
	dao              *repository.Queries
	db               *sqlclient.DB
	chain            chainsync.Chain
	log              logx.SugaredLogger
	expirationPeriod int64
	mu               *sync.Mutex
}

func NewService(cfg cfgx.Config, db *sqlclient.DB, provider chainProviderDTO.ChainProvider, privateKey string) (DepositsService, error) {
	address := cfg.GetString("contracts.escrow.address", "")
	if address == "" {
		return nil, fmt.Errorf("escrow contract address not set")
	}

	// create the escrow contract
	escrowContract, err := escrow.New(provider, privateKey, address)
	if err != nil {
		return nil, fmt.Errorf("failed to create escrow contract: %w", err)
	}

	chainClient := chainClient.NewChainClient(cfg, provider)

	reposit := repository.New(db)

	// create the chain
	chainSynchronizer := chainsync.New(chainClient, escrowContract, reposit, db)

	return &depositsService{
		dao:              reposit,
		db:               db,
		chain:            chainSynchronizer,
		expirationPeriod: 100,
		log:              logx.New().Named("deposits-service").Sugar(),
	}, nil
}
