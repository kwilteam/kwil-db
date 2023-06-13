package client

import (
	"context"
	"fmt"

	txpb "github.com/kwilteam/kwil-db/api/protobuf/tx/v1"
	kTx "github.com/kwilteam/kwil-db/pkg/tx"
)

func (c *Client) ApproveValidator(ctx context.Context, pubKey []byte) error {

	resp, err := c.txClient.ApproveValidator(ctx, &txpb.ValidatorApprovalRequest{PubKey: pubKey})
	if err != nil {
		return fmt.Errorf("failed to approve Validator: %w", err)
	}

	if resp.Status != txpb.RequestStatus_OK {
		return fmt.Errorf("failed to approve Validator with error: %s", resp.Log)
	}

	fmt.Printf("Validator %s has been approved, Log: %s\n", pubKey, resp.Log)
	return nil
}

func (c *Client) ValidatorJoin(ctx context.Context, tx *kTx.Transaction) (*kTx.Receipt, error) {
	pbTx := ConvertTx(tx)
	fmt.Println("Broadcasting ValidatorJoin transaction")
	resp, err := c.txClient.ValidatorJoin(ctx, &txpb.ValidatorJoinRequest{Tx: pbTx})
	if err != nil {
		fmt.Println("TxServiceClient failed to join Validator", err)
		return nil, fmt.Errorf("TxServiceClient failed to join Validator: %w", err)
	}

	if resp.Receipt == nil {
		fmt.Println("TxServiceClient failed to join Validator: receipt is nil")
		return nil, fmt.Errorf("TxServiceClient failed to join Validator: receipt is nil")
	}

	txRes := ConvertReceipt(resp.Receipt)
	return txRes, nil
}
