package pricingsvc

import (
	"context"
	"fmt"
	commonpb "kwil/api/protobuf/common/v0"
	pricingpb "kwil/api/protobuf/pricing/v0"
	"kwil/pkg/accounts"
	"kwil/pkg/utils/serialize"
)

func (s *Service) EstimateCost(ctx context.Context, req *pricingpb.EstimateRequest) (*pricingpb.EstimateResponse, error) {
	tx, err := serialize.Convert[commonpb.Tx, accounts.Transaction](req.Tx)
	if err != nil {
		return nil, fmt.Errorf("failed to convert transaction: %w", err)
	}
	price, err := s.pricer.EstimatePrice(ctx, tx, s.executor)
	if err != nil {
		return nil, fmt.Errorf("failed to estimate price: %w", err)
	}

	return &pricingpb.EstimateResponse{
		Cost: price,
	}, nil
}
