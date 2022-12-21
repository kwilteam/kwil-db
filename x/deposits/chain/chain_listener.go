package chain

import (
	"context"
	ccDTO "kwil/x/chain-client/dto"
	"kwil/x/deposits/repository"
	"kwil/x/logx"
	"kwil/x/sqlx/sqlclient"
	"sync"
)

type ChainClient interface {
	Listen(ctx context.Context, confirmed bool) (<-chan int64, error)
	GetLatestBlock(ctx context.Context, confirmed bool) (int64, error)
	GetDeposits(ctx context.Context, start, end int64) ([]*ccDTO.DepositEvent, error)
	GetWithdrawals(ctx context.Context, start, end int64) ([]*ccDTO.WithdrawalConfirmationEvent, error)
	ReturnFunds(ctx context.Context, params *ccDTO.ReturnFundsParams) (*ccDTO.ReturnFundsResponse, error)
}

type chain struct {
	db          *sqlclient.DB
	dao         *repository.Queries
	chainClient ChainClient
	log         logx.SugaredLogger
	chunkSize   int64
	mu          *sync.Mutex
	height      int64
}

func New(client ChainClient, db *sqlclient.DB) *chain {
	return &chain{
		db:          db,
		dao:         repository.New(db),
		chainClient: client,
		log:         logx.New().Named("deposit-chain-client").Sugar(),
		chunkSize:   100000,
		mu:          &sync.Mutex{},
		height:      0,
	}
}
