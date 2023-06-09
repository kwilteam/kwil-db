package client

import (
	"context"
	"fmt"

	"github.com/cometbft/cometbft/crypto/ed25519"
	txpb "github.com/kwilteam/kwil-db/api/protobuf/tx/v1"
)

func (c *Client) ApproveValidator(ctx context.Context, pubKey ed25519.PubKey) error {
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
