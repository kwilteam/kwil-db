package kclient

import (
	"context"
	gclient "kwil/pkg/grpc/client"
)

func (c *Client) GetServiceConfig(ctx context.Context) (gclient.SvcConfig, error) {
	return c.Kwil.GetServiceConfig(ctx)
}
