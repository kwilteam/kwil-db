package client

import (
	"context"
	"fmt"
	commonpb "kwil/api/protobuf/kwil/common/v0/gen/go"
	pb "kwil/api/protobuf/kwil/pricing/v0/gen/go"
	"kwil/pkg/crypto/transactions"
	"kwil/pkg/utils/serialize"
)

func (c *Client) EstimateCost(ctx context.Context, tx *transactions.Transaction) (string, error) {
	// convert transaction to proto
	pbTx, err := serialize.Convert[transactions.Transaction, commonpb.Tx](tx)
	if err != nil {
		return "", fmt.Errorf("failed to convert transaction: %w", err)
	}

	res, err := c.pricingClt.EstimateCost(ctx, &pb.EstimateRequest{
		Tx: pbTx,
	})
	if err != nil {
		return "", fmt.Errorf("failed to estimate cost: %w", err)
	}

	return res.Cost, nil
}
