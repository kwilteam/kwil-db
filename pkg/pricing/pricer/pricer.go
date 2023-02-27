package pricer

import (
	"context"
	"fmt"
	"kwil/internal/usecases/executor"
	"kwil/pkg/accounts"
	"kwil/pkg/pricing"
)

type Pricer interface {
	EstimatePrice(ctx context.Context, tx *accounts.Transaction, exec executor.Executor) (string, error)
	GetPrice(ctx context.Context, tx *accounts.Transaction, exec executor.Executor) (string, error)
}

type pricer struct{}

func NewPricer() Pricer {
	return &pricer{}
}

// for estimating a price before signing a tx
func (p *pricer) EstimatePrice(ctx context.Context, tx *accounts.Transaction, exec executor.Executor) (string, error) {
	// for now, we will just determine the request type and return a fixed price
	var price string
	var err error

	switch tx.PayloadType {
	case accounts.DEPLOY_DATABASE:
		price = GetPrice(pricing.DEPLOY)
	case accounts.DROP_DATABASE:
		price = GetPrice(pricing.DROP)
	case accounts.EXECUTE_QUERY:
		price = GetPrice(pricing.QUERY)
		//price, err = EstimateQueryPrice(ctx, tx, exec)
	default:
		return "", fmt.Errorf("invalid payload type.  received: %d", tx.PayloadType)
	}
	return price, err
}

// for getting a tx price at execution time
func (p *pricer) GetPrice(ctx context.Context, tx *accounts.Transaction, exec executor.Executor) (string, error) {
	var price string
	var err error

	switch tx.PayloadType {
	case accounts.DEPLOY_DATABASE:
		price = GetPrice(pricing.DEPLOY)
	case accounts.DROP_DATABASE:
		price = GetPrice(pricing.DROP)
	case accounts.EXECUTE_QUERY:
		price = GetPrice(pricing.QUERY)
		//price, err = EstimateQueryPrice(ctx, tx, exec)
	default:
		return "", fmt.Errorf("invalid payload type.  received: %d", tx.PayloadType)
	}
	return price, err
}
