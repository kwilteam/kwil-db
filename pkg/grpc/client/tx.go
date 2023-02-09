package client

import (
	"context"
	"fmt"
	commonpb "kwil/api/protobuf/common/v0"
	txpb "kwil/api/protobuf/tx/v0"
	"kwil/pkg/crypto/transactions"
	"kwil/pkg/utils/serialize"
)

func (c *Client) Broadcast(ctx context.Context, tx *transactions.Transaction) (*transactions.Response, error) {
	// convert transaction to proto
	pbTx, err := serialize.Convert[transactions.Transaction, commonpb.Tx](tx)
	if err != nil {
		return nil, fmt.Errorf("failed to convert transaction: %w", err)
	}

	res, err := c.txClt.Broadcast(ctx, &txpb.BroadcastRequest{Tx: pbTx})
	if err != nil {
		return nil, fmt.Errorf("TxServiceClient failed to Broadcast transaction: %w", err)
	}

	// convert response to transaction
	txRes, err := serialize.Convert[txpb.BroadcastResponse, transactions.Response](res)
	if err != nil {
		return nil, fmt.Errorf("failed to convert response: %w", err)
	}

	return txRes, nil
}

func (c *Client) Ping(ctx context.Context) (string, error) {
	res, err := c.txClt.Ping(ctx, &txpb.PingRequest{})
	if err != nil {
		return "", fmt.Errorf("failed to ping: %w", err)
	}

	return res.Message, nil
}
