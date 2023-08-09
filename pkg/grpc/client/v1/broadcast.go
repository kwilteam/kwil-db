package client

import (
	"context"
	"fmt"
	"math/big"

	txpb "github.com/kwilteam/kwil-db/api/protobuf/tx/v1"
	kTx "github.com/kwilteam/kwil-db/pkg/tx"
)

func (c *Client) Broadcast(ctx context.Context, tx *kTx.Transaction) (*kTx.Receipt, error) {
	pbTx := ConvertTx(tx)
	res, err := c.txClient.Broadcast(ctx, &txpb.BroadcastRequest{Tx: pbTx})
	if err != nil {
		return nil, fmt.Errorf("TxServiceClient failed to Broadcast transaction: %w", err)
	}

	if res.Receipt == nil {
		return nil, fmt.Errorf("TxServiceClient failed to Broadcast transaction: receipt is nil")
	}

	txRes := ConvertReceipt(res.Receipt)

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
	pbTx := ConvertTx(tx)

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

func ConvertTx(incoming *kTx.Transaction) *txpb.Tx {
	return &txpb.Tx{
		Hash:        incoming.Hash,
		PayloadType: incoming.PayloadType.Int32(),
		Payload:     incoming.Payload,
		Fee:         incoming.Fee,
		Nonce:       incoming.Nonce,
		Signature:   convertActionSignature(incoming.Signature),
		Sender:      incoming.Sender,
	}
}

func ConvertReceipt(incoming *txpb.TxReceipt) *kTx.Receipt {
	return &kTx.Receipt{
		TxHash: incoming.TxHash,
		Fee:    incoming.Fee,
		Body:   incoming.Body,
	}
}
