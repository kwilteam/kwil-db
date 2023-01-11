package client

import (
	"fmt"
	"kwil/kwil/svc/accountsclient"
	"kwil/kwil/svc/pricingclient"
	"kwil/kwil/svc/txclient"
	chainClient "kwil/x/chain/client"
	chainClientDto "kwil/x/chain/client/dto"
	chainClientService "kwil/x/chain/client/service"
	"kwil/x/contracts/escrow"
	"kwil/x/contracts/token"

	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

type Client struct {

	// GRPC clients
	Accounts accountsclient.AccountsClient
	Txs      txclient.TxClient
	Pricing  pricingclient.PricingClient

	// Blockchain clients / contracts / interfaces
	Escrow      escrow.EscrowContract
	Token       token.TokenContract
	ChainClient chainClient.ChainClient

	EscrowedTokenAddress string

	Config *ClientConfig
}

func NewClient(cc *grpc.ClientConn, v *viper.Viper) (*Client, error) {
	conf, err := NewClientConfig(v)
	if err != nil {
		return nil, err
	}

	fundingPool := v.GetString("funding-pool")
	ethProvider := v.GetString("eth-provider")

	chnClient, err := chainClientService.NewChainClientExplicit(&chainClientDto.Config{
		ChainCode:             int64(conf.ChainCode),
		Endpoint:              "wss://" + ethProvider,
		ReconnectionInterval:  30,
		RequiredConfirmations: 12,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create chain client: %v", err)
	}

	// escrow
	escrowCtr, err := escrow.New(chnClient, conf.PrivateKey, fundingPool)
	if err != nil {
		return nil, fmt.Errorf("failed to create escrow contract: %v", err)
	}

	// erc20 address
	tokenAddress := escrowCtr.TokenAddress()

	// erc20
	erc20Ctr, err := token.New(chnClient, conf.PrivateKey, tokenAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to create erc20 contract: %v", err)
	}

	return &Client{
		Accounts: accountsclient.New(cc),
		Txs:      txclient.New(cc),
		Pricing:  pricingclient.New(cc),

		Escrow:               escrowCtr,
		Token:                erc20Ctr,
		ChainClient:          chnClient,
		EscrowedTokenAddress: tokenAddress,

		Config: conf,
	}, nil
}
