package client

import (
	"context"
	"fmt"
	accountspb "github.com/kwilteam/kwil-db/api/protobuf/accounts/v0"
	cfgpb "github.com/kwilteam/kwil-db/api/protobuf/config/v0"
	pricingpb "github.com/kwilteam/kwil-db/api/protobuf/pricing/v0"
	txpb "github.com/kwilteam/kwil-db/api/protobuf/tx/v0"
	"github.com/kwilteam/kwil-db/internal/pkg/transport"
	"github.com/kwilteam/kwil-db/pkg/log"

	"google.golang.org/grpc"
)

type Config struct {
	Addr string `mapstructure:"addr"`
}

type Client struct {
	accountClt accountspb.AccountServiceClient
	txClt      txpb.TxServiceClient
	pricingClt pricingpb.PricingServiceClient
	cfgClt     cfgpb.ConfigServiceClient

	conn *grpc.ClientConn
}

// @yaiba TODO: manually declare dependencies
func NewClient(ctx context.Context, log log.Logger, conn grpc.ClientConnInterface,
	accountClt accountspb.AccountServiceClient, txClt txpb.TxServiceClient,
	pricingClt pricingpb.PricingServiceClient, cfgClt cfgpb.ConfigServiceClient) GrpcClient {
	return &Client{
		accountClt: accountClt,
		txClt:      txClt,
		pricingClt: pricingClt,
		cfgClt:     cfgClt,
		conn:       conn.(*grpc.ClientConn),
	}
}

func New(ctx context.Context, target string, opts ...grpc.DialOption) (*Client, error) {
	conn, err := transport.Dial(ctx, target, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to dial server %s: %w", target, err)
	}
	return &Client{
		accountClt: accountspb.NewAccountServiceClient(conn),
		txClt:      txpb.NewTxServiceClient(conn),
		pricingClt: pricingpb.NewPricingServiceClient(conn),
		cfgClt:     cfgpb.NewConfigServiceClient(conn),
		conn:       conn,
	}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) GetTarget() string {
	return c.conn.Target()
}
