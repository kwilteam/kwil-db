package service

import (
	"context"
	"kwil/x/chain-client/dto"
	"kwil/x/logx"
	"time"
)

type ChainClient interface {
	Listen(ctx context.Context, confirmed bool) (<-chan int64, error)
	GetLatestBlock(ctx context.Context, confirmed bool) (int64, error)
	GetDeposits(ctx context.Context, start, end int64) ([]*dto.DepositEvent, error)
	GetWithdrawals(ctx context.Context, start, end int64) ([]*dto.WithdrawalConfirmationEvent, error)
	ReturnFunds(ctx context.Context, params *dto.ReturnFundsParams) (*dto.ReturnFundsResponse, error)
}

// chainClient implements the ChainClient interface
type chainClient struct {
	client                client
	log                   logx.SugaredLogger
	timeout               time.Duration
	requiredConfirmations int64
}

// client is an interface that allows us to mock the specific chain clients (e.g. EVMClient)
type client interface {
	EscrowContract
	Subscribe(ctx context.Context, confirmed bool) (subscription, error)
	GetLatestBlock(ctx context.Context) (int64, error)
}

func NewChainClient() ChainClient {
	return &chainClient{}
}
