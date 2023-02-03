package kwil_client

import (
	"context"
	"fmt"
	"kwil/pkg/fund"
	"kwil/pkg/fund/ethereum"
	grpcClt "kwil/pkg/grpc/client"
)

type Client struct {
	Config *Config
	// GRPC clients
	Kwil *grpcClt.Client

	Fund fund.IFund
}

func New(ctx context.Context, cfg *Config) (*Client, error) {
	//var err error
	//if Config.Fund == nil {
	//	Config.Fund, err = fund2.NewConfig()
	//	if err != nil {
	//		return nil, err
	//	}
	//}

	chainClient, err := ethereum.NewClient(&cfg.Fund)
	if err != nil {
		return nil, err
	}

	kc, err := grpcClt.New(ctx, &cfg.Node)
	if err != nil {
		return nil, err
	}

	return &Client{
		Kwil: kc,
		Fund: chainClient,
	}, nil
}

//func NewConfig() (*GrpcConfig, error) {
//	return &GrpcConfig{
//		Endpoint: "localhost:50051",
//		Fund:     fund2.NewConfig(),
//	}, nil
//}

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
