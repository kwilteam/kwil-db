package dto

import (
	"context"
	"kwil/x/chain"
	"math/big"

	"github.com/ethereum/go-ethereum/ethclient"
)

type ChainProvider interface {
	HeaderByNumber(ctx context.Context, number *big.Int) (*Header, error)
	SubscribeNewHead(ctx context.Context, ch chan<- Header) (Subscription, error)
	ChainCode() chain.ChainCode
	AsEthClient() (*ethclient.Client, error)
}
