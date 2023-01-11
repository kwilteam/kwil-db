package pricingclient

import (
	"context"
	"kwil/x/proto/pricingpb"
	"kwil/x/types/transactions"

	"google.golang.org/grpc"
)

type PricingClient interface {
	EstimateCost(ctx context.Context, tx *transactions.Transaction) (string, error)
}

type client struct {
	pricing pricingpb.PricingServiceClient
}

func New(cc *grpc.ClientConn) PricingClient {
	return &client{
		pricing: pricingpb.NewPricingServiceClient(cc),
	}
}
