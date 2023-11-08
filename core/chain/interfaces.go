package chain

import (
	"context"
	"math/big"

	"github.com/kwilteam/kwil-db/core/types/chain"
)

type ChainClient interface {
	ChainCode() chain.ChainCode
	Endpoint() string
	Close() error
	GetAccountNonce(ctx context.Context, addr string) (uint64, error)
	SuggestGasPrice(ctx context.Context) (*big.Int, error)
	HeaderByNumber(ctx context.Context, number *big.Int) (*chain.Header, error)
	SubscribeNewHead(ctx context.Context, ch chan<- chain.Header) (chain.Subscription, error)
}
