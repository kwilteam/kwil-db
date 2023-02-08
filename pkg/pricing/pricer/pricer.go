package pricer

import (
	"context"
	"fmt"
	transactions2 "kwil/pkg/crypto/transactions"
	"kwil/pkg/pricing"
)

type Pricer interface {
	EstimatePrice(ctx context.Context, tx *transactions2.Transaction) (string, error)
	GetPrice(tx *transactions2.Transaction) (string, error)
}

type pricer struct{}

func NewPricer() Pricer {
	return &pricer{}
}

// for estimating a price before signing a tx
func (p *pricer) EstimatePrice(ctx context.Context, tx *transactions2.Transaction) (string, error) {
	// for now, we will just determine the request type and return a fixed price

	// just a passthrough for now until we implement the pricing service
	return p.GetPrice(tx)
}

// for getting a tx price at databases time
func (p *pricer) GetPrice(tx *transactions2.Transaction) (string, error) {
	var price string

	switch tx.PayloadType {
	case transactions2.DEPLOY_DATABASE:
		price = GetPrice(pricing.DEPLOY)
	case transactions2.DROP_DATABASE:
		price = GetPrice(pricing.DROP)
	case transactions2.EXECUTE_QUERY:
		price = GetPrice(pricing.QUERY)
	default:
		return "", fmt.Errorf("invalid payload type.  received: %d", tx.PayloadType)
	}

	return price, nil
}
