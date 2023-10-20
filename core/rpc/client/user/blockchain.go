package client

import (
	"context"
	"fmt"

	"github.com/kwilteam/kwil-db/core/rpc/conversion"
	txpb "github.com/kwilteam/kwil-db/core/rpc/protobuf/tx/v1"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/types/transactions"
)

func (c *Client) TxQuery(ctx context.Context, txHash []byte) (*transactions.TcTxQueryResponse, error) {
	res, err := c.txClient.TxQuery(ctx, &txpb.TxQueryRequest{
		TxHash: txHash,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query: %w", err)
	}

	return conversion.ConvertToTxQueryResp(res)
}

// ChainInfo gets information on the blockchain of the remote host.
func (c *Client) ChainInfo(ctx context.Context) (*types.ChainInfo, error) {
	res, err := c.txClient.ChainInfo(ctx, &txpb.ChainInfoRequest{})
	if err != nil {
		return nil, err
	}
	return &types.ChainInfo{
		ChainID:     res.ChainId,
		BlockHeight: res.Height,
		BlockHash:   res.Hash,
	}, nil
}
