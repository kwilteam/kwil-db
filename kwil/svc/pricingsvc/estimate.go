package pricingsvc

import (
	"context"
	"fmt"
	"kwil/x/proto/pricingpb"
	txUtils "kwil/x/transactions/utils"
)

func (s *Service) EstimateCost(ctx context.Context, req *pricingpb.EstimateRequest) (*pricingpb.EstimateResponse, error) {
	price, err := s.pricer.EstimatePrice(ctx, txUtils.TxFromMsg(req.Tx))
	if err != nil {
		return nil, fmt.Errorf("failed to estimate price: %w", err)
	}

	return &pricingpb.EstimateResponse{
		Price: price,
	}, nil
}
