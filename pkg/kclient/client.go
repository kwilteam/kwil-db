package kclient

import (
	"context"
	"fmt"
	"kwil/pkg/fund"
	"kwil/pkg/fund/ethereum"
	grpcClt "kwil/pkg/grpc/client"
	"kwil/pkg/log"
)

type Client struct {
	Config *Config
	// GRPC clients
	Kwil grpcClt.GrpcClient

	Fund fund.IFund
}

func New(ctx context.Context, cfg *Config) (*Client, error) {
	log := log.New(cfg.Log)
	chainClient, err := ethereum.NewClient(&cfg.Fund, log)
	if err != nil {
		return nil, err
	}

	kc, err := grpcClt.New(ctx, &cfg.Node, log)
	if err != nil {
		return nil, err
	}

	return &Client{
		Kwil:   kc,
		Fund:   chainClient,
		Config: cfg,
	}, nil
}

func (c *Client) Close() error {
	// err will overwrite the previous error
	var err error
	defer func() {
		if e := c.Kwil.Close(); e != nil {
			err = fmt.Errorf("%w", e)
		}
	}()

	defer func() {
		if e := c.Fund.Close(); e != nil {
			err = fmt.Errorf("%w", e)
		}
	}()

	return err
}
