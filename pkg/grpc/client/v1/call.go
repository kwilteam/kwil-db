package client

import (
	"context"
	"encoding/json"
	"fmt"
	txpb "github.com/kwilteam/kwil-db/api/protobuf/tx/v1"
	"github.com/kwilteam/kwil-db/pkg/crypto"
	"github.com/kwilteam/kwil-db/pkg/tx"
)

func (c *Client) Call(ctx context.Context, req *tx.CallActionMessage) ([]map[string]any, error) {
	payload, err := req.Payload.Bytes()
	if err != nil {
		return nil, fmt.Errorf("failed to convert payload to bytes: %w", err)
	}

	grpcMsg := &txpb.CallRequest{
		Payload:   payload,
		Signature: convertActionSignature(req.Signature),
		Sender:    req.Sender,
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

func convertActionSignature(oldSig *crypto.Signature) *txpb.Signature {
	if oldSig == nil {
		return &txpb.Signature{}
	}

	newSig := &txpb.Signature{
		SignatureBytes: oldSig.Signature,
		SignatureType:  oldSig.Type.Int32(),
	}

	return newSig
}
