package client

import (
	"context"
	"fmt"
	txpb "kwil/api/protobuf/tx/v1"
)

func (c *Client) ListDatabases(ctx context.Context, owner string) ([]string, error) {
	res, err := c.txClient.ListDatabases(ctx, &txpb.ListDatabasesRequest{
		Owner: owner,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list databases: %w", err)
	}

	return res.Databases, nil
}
