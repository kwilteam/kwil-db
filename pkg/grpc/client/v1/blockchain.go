package client

import (
	"context"
	"fmt"

	txpb "github.com/kwilteam/kwil-db/api/protobuf/tx/v1"
	"github.com/kwilteam/kwil-db/pkg/client/types"
	"github.com/kwilteam/kwil-db/pkg/grpc/client/v1/conversion"
)

func (c *Client) TxQuery(ctx context.Context, txHash []byte) (*types.TcTxQueryResponse, error) {
	res, err := c.txClient.TxQuery(ctx, &txpb.TxQueryRequest{
		TxHash: txHash,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query: %w", err)
	}

	return conversion.ConvertToTxQueryResp(res)
}
