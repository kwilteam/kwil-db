package txclient

import (
	"context"
	"fmt"
	txpb "kwil/api/protobuf/tx/v0/gen/go"
)

func (c *client) Ping(ctx context.Context) (string, error) {
	res, err := c.txs.Ping(ctx, &txpb.PingRequest{})
	if err != nil {
		return "", fmt.Errorf("failed to ping: %w", err)
	}

	return res.Message, nil
}
