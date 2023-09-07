package client

import (
	"context"
	"fmt"
	"github.com/kwilteam/kwil-db/api/protobuf/conversion"
	txpb "github.com/kwilteam/kwil-db/api/protobuf/tx/v1"
	"github.com/kwilteam/kwil-db/pkg/client/types"
)

func (c *Client) TxQuery(ctx context.Context, txHash []byte) (*types.TxQueryResponse, error) {
	res, err := c.txClient.TxQuery(ctx, &txpb.TxQueryRequest{
		TxHash: txHash,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query: %w", err)
	}

	return conversion.ConvertToTxQueryResp(res)
}
