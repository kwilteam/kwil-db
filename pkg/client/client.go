package client

import (
	"context"
	cc "kwil/pkg/chain/client"
	grpc "kwil/pkg/grpc/client"
)

type Client struct {
	endpoint    string
	grpc        grpc.Client
	chainClient cc.ChainClient
}

func New(ctx context.Context, rpcUrl string, opts ...ClientOption) (*Client, error) {
	/*
		c := &Client{
			endpoint: rpcUrl,
		}
		for _, opt := range opts {
			opt(c)
		}

		grpcClient, err := grpc.New(ctx, &grpc.Config{
			Addr: rpcUrl,
		})*/

	return &Client{endpoint: rpcUrl}, nil
}
