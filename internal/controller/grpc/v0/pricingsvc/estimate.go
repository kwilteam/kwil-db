package pricingsvc

import (
	"context"
	"fmt"
	commonpb "kwil/api/protobuf/kwil/common/v0/gen/go"
	pb "kwil/api/protobuf/kwil/pricing/v0/gen/go"
	"kwil/pkg/crypto/transactions"
	"kwil/pkg/utils/serialize"
)

func (s *Service) EstimateCost(ctx context.Context, req *pb.EstimateRequest) (*pb.EstimateResponse, error) {
	tx, err := serialize.Convert[commonpb.Tx, transactions.Transaction](req.Tx)
	if err != nil {
		return nil, fmt.Errorf("failed to convert transaction: %w", err)
	}

	price, err := s.pricer.EstimatePrice(ctx, tx)
	if err != nil {
		return nil, fmt.Errorf("failed to estimate price: %w", err)
	}

	return &pb.EstimateResponse{
		Cost: price,
	}, nil
}
