package pricer

import (
	"context"
	"fmt"
	"kwil/pkg/accounts"
	"kwil/pkg/pricing"
)

type Pricer interface {
	EstimatePrice(ctx context.Context, tx *accounts.Transaction) (string, error)
	GetPrice(tx *accounts.Transaction) (string, error)
}

type pricer struct{}

func NewPricer() Pricer {
	return &pricer{}
}

// for estimating a price before signing a tx
func (p *pricer) EstimatePrice(ctx context.Context, tx *accounts.Transaction) (string, error) {
	// for now, we will just determine the request type and return a fixed price

	// just a passthrough for now until we implement the pricing service
	return p.GetPrice(tx)
}

// for getting a tx price at databases time
func (p *pricer) GetPrice(tx *accounts.Transaction) (string, error) {
	var price string

	switch tx.PayloadType {
	case accounts.DEPLOY_DATABASE:
		price = GetPrice(pricing.DEPLOY)
	case accounts.DROP_DATABASE:
		price = GetPrice(pricing.DROP)
	case accounts.EXECUTE_QUERY:
		price = GetPrice(pricing.QUERY)
	default:
		return "", fmt.Errorf("invalid payload type.  received: %d", tx.PayloadType)
	}

	return price, nil
}
