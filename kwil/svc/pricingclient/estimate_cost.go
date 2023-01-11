package pricingclient

import (
	"context"
	"fmt"
	"kwil/x/proto/commonpb"
	"kwil/x/proto/pricingpb"
	"kwil/x/types/transactions"
	"kwil/x/utils/serialize"
)

func (c *client) EstimateCost(ctx context.Context, tx *transactions.Transaction) (string, error) {
	// convert transaction to proto
	pbTx, err := serialize.Convert[transactions.Transaction, commonpb.Tx](tx)
	if err != nil {
		return "", fmt.Errorf("failed to convert transaction: %w", err)
	}

	res, err := c.pricing.EstimateCost(ctx, &pricingpb.EstimateRequest{
		Tx: pbTx,
	})
	if err != nil {
		return "", fmt.Errorf("failed to estimate cost: %w", err)
	}

	return res.Cost, nil
}
