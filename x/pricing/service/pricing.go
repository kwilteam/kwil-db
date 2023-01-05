package service

import (
	"context"
	"fmt"
	"kwil/x/pricing"
	"kwil/x/pricing/entity"
	"kwil/x/transactions"
	txDto "kwil/x/transactions/dto"
)

type PricingService interface {
	EstimatePrice(ctx context.Context, tx *txDto.Transaction) (string, error)
	GetPrice(tx *txDto.Transaction) (string, error)
}

type pricingService struct {
}

func NewService() *pricingService {
	return &pricingService{}
}

// for estimating a price before signing a tx
func (p *pricingService) EstimatePrice(ctx context.Context, tx *txDto.Transaction) (string, error) {
	// for now, we will just determine the request type and return a fixed price

	// just a passthrough for now until we implement the pricing service
	return p.GetPrice(tx)
}

// for getting a tx price at execution time
func (p *pricingService) GetPrice(tx *txDto.Transaction) (string, error) {
	var price string

	switch tx.PayloadType {
	case transactions.DEPLOY_DATABASE:
		price = entity.EstimatePrice(pricing.DEPLOY)
	case transactions.DROP_DATABASE:
		price = entity.EstimatePrice(pricing.DROP)
	case transactions.EXECUTE_QUERY:
		price = entity.EstimatePrice(pricing.QUERY)
	default:
		return "", fmt.Errorf("invalid payload type")
	}

	return price, nil
}
