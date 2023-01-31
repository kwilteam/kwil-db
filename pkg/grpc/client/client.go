package client

import (
	"kwil/kwil/svc/accountsclient"
	"kwil/kwil/svc/pricingclient"
	"kwil/kwil/svc/txclient"
	"kwil/x/fund"
	"kwil/x/fund/ethereum"

	"google.golang.org/grpc"
)

type Client struct {
	// GRPC clients
	Accounts accountsclient.AccountsClient
	Txs      txclient.TxClient
	Pricing  pricingclient.PricingClient

	Chain fund.IFund
}

func NewClient(cc *grpc.ClientConn, chainCfg *fund.Config) (*Client, error) {
	var err error
	if chainCfg == nil {
		chainCfg, err = fund.NewConfig()
		if err != nil {
			return nil, err
		}
	}

	chainClient, err := ethereum.NewClient(chainCfg)
	if err != nil {
		return nil, err
	}

	return &Client{
		Accounts: accountsclient.New(cc),
		Txs:      txclient.New(cc),
		Pricing:  pricingclient.New(cc),
		Chain:    chainClient,
	}, nil
}
