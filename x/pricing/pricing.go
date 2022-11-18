package pricing

import (
	"context"
	"math/big"
)

/*
	This is mostly just a stand-in for the pricing service.
	Once we have better pricing, we can replace this with a real service, I just wanted to make this so that there
	is an expectation that Kwil is not free, and give us a way to add more pricing options as we get to a workable model.
*/

type Service interface {
	GetPrice(ctx context.Context) (*big.Int, error)
	GetPriceForDDL(ctx context.Context) (*big.Int, error)
}

type pricingService struct {
	prices prices
}

func NewService() *pricingService {
	return &pricingService{
		prices: prices{
			ddl:    big.NewInt(1000000000000000000),
			insert: big.NewInt(2000000000000000),
		},
	}
}

func (s *pricingService) GetPrice(ctx context.Context) (*big.Int, error) {
	return s.prices.insert, nil
}

func (s *pricingService) GetPriceForDDL(ctx context.Context) (*big.Int, error) {
	return s.prices.ddl, nil
}

// this is a temporary placeholder until we get better pricing
type prices struct {
	ddl    *big.Int `default:"100000"`
	insert *big.Int `default:"2000"`
}
