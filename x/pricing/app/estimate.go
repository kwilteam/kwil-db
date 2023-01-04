package app

import (
	"context"
	pricing "kwil/x/proto/pricingpb"
)

// right now this uses the pricing entity but not the service
func (s *Service) EstimateCost(ctx context.Context, req *pricing.EstimateRequest) (*pricing.EstimateResponse, error) {
	p, err := s.pricing.EstimatePrice(req)
	if err != nil {
		return nil, err
	}
	return p, nil
}
