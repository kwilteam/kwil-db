package pricer

import (
	"context"
	"github.com/kwilteam/kwil-db/internal/usecases/executor"
	"github.com/kwilteam/kwil-db/pkg/accounts"
	"github.com/kwilteam/kwil-db/pkg/pricing"
)

// this is by no means a complete implementation of the pricing service

const (
	CREATE_PRICE = "1000000000000000000"
	DROP_PRICE   = "10000000000000"
	QUERY_PRICE  = "2000000000000000"
)

func GetPrice(p pricing.PricingRequestType) string {
	switch p {
	case pricing.DEPLOY:
		return CREATE_PRICE
	case pricing.DROP:
		return DROP_PRICE
	case pricing.QUERY:
		return QUERY_PRICE
	}
	return "0"
}

func EstimateQueryPrice(ctx context.Context, tx *accounts.Transaction, exec executor.Executor) (string, error) {
	e, err := NewQueryPriceEstimator(ctx, tx, exec)
	if err != nil {
		return QUERY_PRICE, err
	}

	return e.Estimate()
}
