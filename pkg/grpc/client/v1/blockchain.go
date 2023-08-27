package client

import (
	"context"
	"fmt"

	txpb "github.com/kwilteam/kwil-db/api/protobuf/tx/v1"
)

func (c *Client) TxQuery(ctx context.Context, txHash []byte) (*txpb.TxQueryResponse, error) {
	res, err := c.txClient.TxQuery(ctx, &txpb.TxQueryRequest{
		TxHash: txHash,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query: %w", err)
	}

	return res, nil
}
