package client

import (
	"context"
	"encoding/json"
	"fmt"

	txpb "github.com/kwilteam/kwil-db/api/protobuf/tx/v1"
)

func (c *Client) Query(ctx context.Context, dbid string, query string) ([]map[string]any, error) {
	res, err := c.txClient.Query(ctx, &txpb.QueryRequest{
		Dbid:  dbid,
		Query: query,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query: %w", err)
	}

	var result []map[string]any
	err = json.Unmarshal(res.Result, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal result: %w", err)
	}

	return result, nil
}
