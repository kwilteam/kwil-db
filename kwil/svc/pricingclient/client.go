package pricingclient

import (
	"context"
	pb "kwil/api/protobuf/pricing/v0/gen/go"
	"kwil/x/types/transactions"

	"google.golang.org/grpc"
)

type PricingClient interface {
	EstimateCost(ctx context.Context, tx *transactions.Transaction) (string, error)
}

type client struct {
	pricing pb.PricingServiceClient
}

func New(cc *grpc.ClientConn) PricingClient {
	return &client{
		pricing: pb.NewPricingServiceClient(cc),
	}
}
