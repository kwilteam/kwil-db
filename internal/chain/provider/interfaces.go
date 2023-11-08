package provider

import (
	"context"
	"crypto/ecdsa"
	"math/big"

	"github.com/kwilteam/kwil-db/internal/chain/types"
)

type ChainProvider interface {
	ProviderClient
	TokenContract
	EscrowContract
}

// Provider is the interface that wraps the basic methods for interacting with a blockchain
type ProviderClient interface {
	ChainCode() types.ChainCode
	Endpoint() string
	Close() error

	GetAccountNonce(ctx context.Context, addr string) (uint64, error)
	SuggestGasPrice(ctx context.Context) (*big.Int, error)

	// HeaderByNumber(number uint64) (*types.Header, error)
	// SubscribeNewHead(ch chan<- *types.Header) (types.Subscription, error)
}

type TokenContract interface {
	Allowance(ctx context.Context, owner, spender string) (*big.Int, error)
	Approve(ctx context.Context, spender string, amount *big.Int, privateKey *ecdsa.PrivateKey) (*types.ApproveResponse, error)
}

type EscrowContract interface {
	Deposit(ctx context.Context, params *types.DepositParams, privateKey *ecdsa.PrivateKey) (*types.DepositResponse, error)
}
