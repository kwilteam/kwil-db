package client

import (
	"context"
	"encoding/json"
	"fmt"

	txpb "github.com/kwilteam/kwil-db/api/protobuf/tx/v1"
	"github.com/kwilteam/kwil-db/pkg/transactions"
)

func (c *Client) Call(ctx context.Context, req *transactions.CallMessage) ([]map[string]any, error) {
	var sender []byte
	if req.Sender != nil {
		sender = req.Sender
	}

	callReq := &txpb.CallRequest{
		Body: &txpb.CallRequest_Body{
			Description: req.Body.Description,
			Payload:     req.Body.Payload,
		},
		Signature:     convertActionSignature(req.Signature),
		Sender:        sender,
		Serialization: req.Serialization.String(),
	}

	res, err := c.txClient.Call(ctx, callReq)

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
