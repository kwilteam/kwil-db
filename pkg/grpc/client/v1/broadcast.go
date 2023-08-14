package client

import (
	"context"
	"fmt"
	"math/big"

	txpb "github.com/kwilteam/kwil-db/api/protobuf/tx/v1"
	"github.com/kwilteam/kwil-db/pkg/transactions"
)

func (c *Client) Broadcast(ctx context.Context, tx *transactions.Transaction) (*transactions.TransactionStatus, error) {
	pbTx := convertTx(tx)
	res, err := c.txClient.Broadcast(ctx, &txpb.BroadcastRequest{Tx: pbTx})
	if err != nil {
		return nil, fmt.Errorf("TxServiceClient failed to Broadcast transaction: %w", err)
	}

	if res.Status == nil {
		return nil, fmt.Errorf("TxServiceClient failed to Broadcast transaction: receipt is nil")
	}

	txRes, err := convertTransactionStatus(res.Status)
	if err != nil {
		return nil, fmt.Errorf("TxServiceClient failed to convert transaction status: %w", err)
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

func (c *Client) EstimateCost(ctx context.Context, tx *transactions.Transaction) (*big.Int, error) {
	// convert transaction to proto
	pbTx := convertTx(tx)

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

	fmt.Println("Estimated cost:", bigCost)
	return bigCost, nil
}
