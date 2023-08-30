package client

import (
	"context"
	"fmt"

	txpb "github.com/kwilteam/kwil-db/api/protobuf/tx/v1"
)

func (c *Client) ListDatabases(ctx context.Context, ownerPubKey []byte) ([]string, error) {
	res, err := c.txClient.ListDatabases(ctx, &txpb.ListDatabasesRequest{
		Owner: ownerPubKey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list databases: %w", err)
	}

	return res.Databases, nil
}
