package txclient

import (
	"context"
	"fmt"
	"kwil/x/proto/txpb"
)

func (c *client) Ping(ctx context.Context) (string, error) {
	res, err := c.txs.Ping(ctx, &txpb.PingRequest{})
	if err != nil {
		return "", fmt.Errorf("failed to ping: %w", err)
	}

	return res.Message, nil
}
