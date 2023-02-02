package kcli

import (
	"context"
	fund2 "kwil/pkg/fund"
	"kwil/pkg/fund/ethereum"
	grpcClt "kwil/pkg/grpc/client"
)

type KwilClient struct {
	Config *Config
	// GRPC clients
	Client *grpcClt.Client

	Fund fund2.IFund
}

func New(ctx context.Context, cfg *Config) (*KwilClient, error) {
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

	kc, err := grpcClt.New(ctx, &cfg.Kwil)
	if err != nil {
		return nil, err
	}

	return &KwilClient{
		Client: kc,
		Fund:   chainClient,
	}, nil
}

//func NewConfig() (*GrpcConfig, error) {
//	return &GrpcConfig{
//		Endpoint: "localhost:50051",
//		Fund:     fund2.NewConfig(),
//	}, nil
//}
