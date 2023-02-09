package client

import (
	"context"
	"fmt"
	commonpb "kwil/api/protobuf/common/v0"
	pricingpb "kwil/api/protobuf/pricing/v0"
	"kwil/pkg/accounts"
	"kwil/pkg/utils/serialize"
)

func (c *Client) EstimateCost(ctx context.Context, tx *accounts.Transaction) (string, error) {
	// convert transaction to proto
	pbTx, err := serialize.Convert[accounts.Transaction, commonpb.Tx](tx)
	if err != nil {
		return "", fmt.Errorf("failed to convert transaction: %w", err)
	}

	res, err := c.pricingClt.EstimateCost(ctx, &pricingpb.EstimateRequest{
		Tx: pbTx,
	})
	if err != nil {
		return "", fmt.Errorf("failed to estimate cost: %w", err)
	}

	return res.Cost, nil
}
