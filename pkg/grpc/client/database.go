package client

import (
	"context"
	"fmt"
	txpb "kwil/api/protobuf/kwil/tx/v0/gen/go"
)

func (c *Gr) ListDatabases(ctx context.Context, address string) ([]string, error) {
	res, err := c.txClt.ListDatabases(ctx, &txpb.ListDatabasesRequest{
		Owner: address,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list databases: %w", err)
	}

	return res.Databases, nil
}
