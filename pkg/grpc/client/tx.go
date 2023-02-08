package client

import (
	"context"
	"fmt"
	commonpb "kwil/api/protobuf/kwil/common/v0/gen/go"
	"kwil/api/protobuf/kwil/tx/v0/gen/go"
	"kwil/pkg/crypto/transactions"
	"kwil/pkg/utils/serialize"
)

func (c *Client) Broadcast(ctx context.Context, tx *transactions.Transaction) (*transactions.Response, error) {
	// convert transaction to proto
	pbTx, err := serialize.Convert[transactions.Transaction, commonpb.Tx](tx)
	if err != nil {
		return nil, fmt.Errorf("failed to convert transaction: %w", err)
	}

	res, err := c.txClt.Broadcast(ctx, &_go.BroadcastRequest{Tx: pbTx})
	if err != nil {
		return nil, fmt.Errorf("TxServiceClient failed to Broadcast transaction: %w", err)
	}

	// convert response to transaction
	txRes, err := serialize.Convert[_go.BroadcastResponse, transactions.Response](res)
	if err != nil {
		return nil, fmt.Errorf("failed to convert response: %w", err)
	}

	return txRes, nil
}

func (c *Client) Ping(ctx context.Context) (string, error) {
	res, err := c.txClt.Ping(ctx, &_go.PingRequest{})
	if err != nil {
		return "", fmt.Errorf("failed to ping: %w", err)
	}

	return res.Message, nil
}
