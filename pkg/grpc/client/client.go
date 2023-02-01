package client

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	accountpb "kwil/api/protobuf/account/v0/gen/go"
	pricingpb "kwil/api/protobuf/pricing/v0/gen/go"
	txpb "kwil/api/protobuf/tx/v0/gen/go"
	"kwil/internal/pkg/transport"
	"kwil/x/types/accounts"
	"kwil/x/types/databases"
	"kwil/x/types/execution"
	"kwil/x/types/transactions"
)

type GrpcClient interface {
	ListDatabases(ctx context.Context, address string) ([]string, error)
	GetExecutablesById(ctx context.Context, id string) ([]*execution.Executable, error)
	GetSchema(ctx context.Context, owner string, dbName string) (*databases.Database[[]byte], error)
	GetSchemaById(ctx context.Context, id string) (*databases.Database[[]byte], error)
	EstimateCost(ctx context.Context, tx *transactions.Transaction) (string, error)
	Broadcast(ctx context.Context, tx *transactions.Transaction) (*transactions.Response, error)
	Ping(ctx context.Context) (string, error)
	GetAccount(ctx context.Context, address string) (accounts.Account, error)
	Close() error
}

type Config struct {
	Endpoint string `mapstructure:"endpoint"`
}

func (c *Config) toConfig() (*clientConfig, error) {
	return &clientConfig{
		Endpoint: c.Endpoint,
	}, nil
}

type clientConfig struct {
	Endpoint string
	// TODO: use logger
	Log *zap.Logger
}

type Client struct {
	accountClt accountpb.AccountServiceClient
	txClt      txpb.TxServiceClient
	pricingClt pricingpb.PricingServiceClient

	conn   *grpc.ClientConn
	Config *clientConfig
}

func New(ctx context.Context, cfg *Config) (*Client, error) {
	clientCfg, err := cfg.toConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to create client config: %w", err)
	}

	conn, err := transport.Dial(ctx, cfg.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to dial server %s: %w", cfg.Endpoint, err)
	}
	return &Client{
		accountClt: accountpb.NewAccountServiceClient(conn),
		txClt:      txpb.NewTxServiceClient(conn),
		pricingClt: pricingpb.NewPricingServiceClient(conn),
		Config:     clientCfg,
		conn:       conn,
	}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}
