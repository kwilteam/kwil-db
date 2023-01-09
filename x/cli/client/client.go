package client

import (
	"context"
	"crypto/ecdsa"
	"kwil/x/transactions"
	accountTypes "kwil/x/types/accounts"

	"kwil/x/proto/accountspb"
	"kwil/x/proto/pricingpb"
	"kwil/x/proto/txpb"
	txUtils "kwil/x/transactions/utils"
	txTypes "kwil/x/types/transactions"

	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

type Client interface {
	UnconnectedClient
	GetAccount(ctx context.Context, address string) (*accountTypes.Account, error)
	EstimatePrice(ctx context.Context, tx *txTypes.Transaction) (string, error)
	Broadcast(ctx context.Context, tx *txTypes.Transaction) (*txTypes.Response, error)
	BuildTransaction(ctx context.Context, payloadType transactions.PayloadType, data interface{}, privateKey *ecdsa.PrivateKey) (*txTypes.Transaction, error)
}

type client struct {
	accounts accountspb.AccountServiceClient
	txs      txpb.TxServiceClient
	pricing  pricingpb.PricingServiceClient

	UnconnectedClient UnconnectedClient
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

func (c *client) GetAccount(ctx context.Context, address string) (*accountTypes.Account, error) {
	acc, err := c.accounts.GetAccount(ctx, &accountspb.GetAccountRequest{
		Address: address,
	})
	if err != nil {
		return nil, err
	}

	return &accountTypes.Account{
		Address: acc.Address,
		Balance: acc.Balance,
		Spent:   acc.Spent,
		Nonce:   acc.Nonce,
	}, nil
}

func (c *client) EstimatePrice(ctx context.Context, tx *txTypes.Transaction) (string, error) {
	// estimate cost
	fee, err := c.pricing.EstimateCost(ctx, &pricingpb.EstimateRequest{
		Tx: txUtils.TxToMsg(tx),
	})
	if err != nil {
		return "", err
	}

	return fee.Price, nil
}

func (c *client) Broadcast(ctx context.Context, tx *txTypes.Transaction) (*txTypes.Response, error) {
	// broadcast
	broadcast, err := c.txs.Broadcast(ctx, &txpb.BroadcastRequest{
		Tx: txUtils.TxToMsg(tx),
	})
	if err != nil {
		return nil, err
	}

	return &txTypes.Response{
		Hash: broadcast.Hash,
		Fee:  broadcast.Fee,
	}, nil
}
