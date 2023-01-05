package app

import (
	"context"
	pricing "kwil/x/proto/pricingpb"
	txUtils "kwil/x/transactions/utils"
)

// right now this uses the pricing entity but not the service
func (s *Service) EstimateCost(ctx context.Context, req *pricing.EstimateRequest) (*pricing.EstimateResponse, error) {
	p, err := s.pricing.EstimatePrice(ctx, txUtils.TxFromMsg(req.Tx))
	if err != nil {
		return nil, err
	}
	return &pricing.EstimateResponse{
		Price: p,
	}, nil
}
