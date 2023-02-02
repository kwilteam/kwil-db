package kcli

import (
	"context"
	"kwil/pkg/grpc/client"
)

func (c *KwilClient) GetNodeInfo(ctx context.Context) (client.NodeInfo, error) {
	return c.Client.GetInfo(ctx)
}
