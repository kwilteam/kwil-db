package txclient

import (
	"context"
	"fmt"
	"kwil/x/proto/txpb"
)

func (c *client) ListDatabases(ctx context.Context, address string) ([]string, error) {
	res, err := c.txs.ListDatabases(ctx, &txpb.ListDatabasesRequest{
		Owner: address,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list databases: %w", err)
	}

	return res.Databases, nil
}
