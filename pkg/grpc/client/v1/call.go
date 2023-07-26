package client

import (
	"context"
	"encoding/json"
	"fmt"

	txpb "github.com/kwilteam/kwil-db/api/protobuf/tx/v1"
)

func (c *Client) Call(ctx context.Context, dbid string, action string, params []map[string]any) ([]map[string]any, error) {
	req := &txpb.CallRequest{
		Dbid:   dbid,
		Action: action,
		Inputs: make([]*txpb.ActionInput, len(params)),
	}

	for i, param := range params {
		input := make(map[string]string)
		for k, v := range param {
			input[k] = fmt.Sprintf("%v", v)
		}
		req.Inputs[i] = &txpb.ActionInput{Input: input}
	}

	res, err := c.txClient.Call(ctx, req)

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
