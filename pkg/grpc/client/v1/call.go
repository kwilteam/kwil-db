package client

import (
	"context"
	"encoding/json"
	"fmt"

	txpb "github.com/kwilteam/kwil-db/api/protobuf/tx/v1"
	"github.com/kwilteam/kwil-db/pkg/transactions"
)

func (c *Client) Call(ctx context.Context, req *transactions.SignedMessage) ([]map[string]any, error) {

	grpcMsg := &txpb.CallRequest{
		Payload:   req.Message,
		Signature: convertActionSignature(req.Signature),
		Sender:    req.Sender.Bytes(),
	}

	res, err := c.txClient.Call(ctx, grpcMsg)

	if err != nil {
		return nil, fmt.Errorf("failed to call: %w", err)
	}

	var result []map[string]any
	err = json.Unmarshal(res.Result, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal result: %w", err)
	}

	return result, nil
}
