package client

import (
	"context"
	"fmt"
	txpb "kwil/api/protobuf/tx/v1"
	kTx "kwil/pkg/tx"
	"kwil/pkg/utils/serialize"
	"math/big"
)

func (c *Client) Broadcast(ctx context.Context, tx *kTx.Transaction) (*kTx.Receipt, error) {
	pbTx, err := serialize.Convert[kTx.Transaction, txpb.Tx](tx)
	if err != nil {
		return nil, fmt.Errorf("failed to convert transaction: %w", err)
	}

	res, err := c.txClient.Broadcast(ctx, &txpb.BroadcastRequest{Tx: pbTx})
	if err != nil {
		return nil, fmt.Errorf("TxServiceClient failed to Broadcast transaction: %w", err)
	}

	txRes, err := serialize.Convert[txpb.BroadcastResponse, kTx.Receipt](res)
	if err != nil {
		return nil, fmt.Errorf("failed to convert response: %w", err)
	}

	return txRes, nil
}

func (c *Client) Ping(ctx context.Context) (string, error) {
	res, err := c.txClient.Ping(ctx, &txpb.PingRequest{})
	if err != nil {
		return "", fmt.Errorf("failed to ping: %w", err)
	}

	return res.Message, nil
}

func (c *Client) EstimateCost(ctx context.Context, tx *kTx.Transaction) (*big.Int, error) {
	// convert transaction to proto
	pbTx, err := serialize.Convert[kTx.Transaction, txpb.Tx](tx)
	if err != nil {
		return nil, fmt.Errorf("failed to convert transaction: %w", err)
	}

	res, err := c.txClient.EstimatePrice(ctx, &txpb.EstimatePriceRequest{
		Tx: pbTx,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to estimate cost: %w", err)
	}

	bigCost, ok := new(big.Int).SetString(res.Price, 10)
	if !ok {
		return nil, fmt.Errorf("failed to convert price to big.Int")
	}

	return bigCost, nil
}
