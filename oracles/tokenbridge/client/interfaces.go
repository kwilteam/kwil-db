package client

import (
	"context"
	"crypto/ecdsa"
	"math/big"

	"github.com/kwilteam/kwil-db/core/types/chain"
	"github.com/kwilteam/kwil-db/oracles/tokenbridge/types"
)

type TokenBridgeClient interface {
	ChainClient
	TokenContract
	EscrowContract
}

type ChainClient interface {
	Close() error
	GetAccountNonce(ctx context.Context, addr string) (uint64, error)
	SuggestGasPrice(ctx context.Context) (*big.Int, error)
	HeaderByNumber(ctx context.Context, number *big.Int) (*chain.Header, error)
	SubscribeNewHead(ctx context.Context, ch chan<- chain.Header) (chain.Subscription, error)
}

type TokenContract interface {
	Allowance(ctx context.Context, owner, spender string) (*big.Int, error)
	Approve(ctx context.Context, spender string, amount *big.Int, privateKey *ecdsa.PrivateKey) (string, error)
	BalanceOf(ctx context.Context, address string) (*big.Int, error)
}

type EscrowContract interface {
	Deposit(ctx context.Context, amount *big.Int, privateKey *ecdsa.PrivateKey) (string, error)
	TokenAddress() string
	DepositBalance(ctx context.Context, address string) (*big.Int, error)
	GetDeposits(ctx context.Context, from uint64, to *uint64) ([]*types.AccountCredit, error)
}
