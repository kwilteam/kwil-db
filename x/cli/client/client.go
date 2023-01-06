package client

import (
	"context"
	accountsDto "kwil/x/accounts/dto"

	"kwil/x/proto/accountspb"
	"kwil/x/proto/pricingpb"
	"kwil/x/proto/txpb"
	txDto "kwil/x/transactions/dto"
	txUtils "kwil/x/transactions/utils"

	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

type Client interface {
	GetAccount(ctx context.Context, address string) (*accountsDto.Account, error)
	EstimatePrice(ctx context.Context, tx *txDto.Transaction) (string, error)
	Broadcast(ctx context.Context, tx *txDto.Transaction) (*txDto.Response, error)
}

type client struct {
	accounts accountspb.AccountServiceClient
	txs      txpb.TxServiceClient
	pricing  pricingpb.PricingServiceClient

	UnconnectedClient *UnconnectedClient
}

func NewClient(cc *grpc.ClientConn, v *viper.Viper) (Client, error) {
	unconndClient, err := NewUnconnectedClient(v)
	if err != nil {
		return nil, err
	}

	return &client{
		accounts: accountspb.NewAccountServiceClient(cc),
		txs:      txpb.NewTxServiceClient(cc),
		pricing:  pricingpb.NewPricingServiceClient(cc),

		UnconnectedClient: unconndClient,
	}, nil
}

func (c *client) GetAccount(ctx context.Context, address string) (*accountsDto.Account, error) {
	acc, err := c.accounts.GetAccount(ctx, &accountspb.GetAccountRequest{
		Address: address,
	})
	if err != nil {
		return nil, err
	}

	return &accountsDto.Account{
		Address: acc.Address,
		Balance: acc.Balance,
		Spent:   acc.Spent,
		Nonce:   acc.Nonce,
	}, nil
}

func (c *client) EstimatePrice(ctx context.Context, tx *txDto.Transaction) (string, error) {
	// estimate cost
	fee, err := c.pricing.EstimateCost(ctx, &pricingpb.EstimateRequest{
		Tx: txUtils.TxToMsg(tx),
	})
	if err != nil {
		return "", err
	}

	return fee.Price, nil
}

func (c *client) Broadcast(ctx context.Context, tx *txDto.Transaction) (*txDto.Response, error) {
	// broadcast
	broadcast, err := c.txs.Broadcast(ctx, &txpb.BroadcastRequest{
		Tx: txUtils.TxToMsg(tx),
	})
	if err != nil {
		return nil, err
	}

	return &txDto.Response{
		Hash: broadcast.Hash,
		Fee:  broadcast.Fee,
	}, nil
}
