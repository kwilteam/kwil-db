package dto

import (
	"context"
	"math/big"
)

type ChainProvider interface {
	HeaderByNumber(ctx context.Context, number *big.Int) (*Header, error)
	SubscribeNewHead(ctx context.Context, ch chan<- Header) (Subscription, error) // TODO: make this
}
