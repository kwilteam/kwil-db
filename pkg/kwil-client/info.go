package kwil_client

import (
	"context"
	"kwil/pkg/grpc/client"
)

func (c *Client) GetNodeInfo(ctx context.Context) (client.NodeInfo, error) {
	return c.Kwil.GetInfo(ctx)
}
