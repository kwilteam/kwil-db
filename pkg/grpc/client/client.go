package client

import (
	"context"
	"fmt"
	accountpb "kwil/api/protobuf/kwil/account/v0/gen/go"
	pricingpb "kwil/api/protobuf/kwil/pricing/v0/gen/go"
	txpb "kwil/api/protobuf/kwil/tx/v0/gen/go"
	"kwil/internal/pkg/transport"
	"kwil/pkg/crypto/transactions"
	"kwil/pkg/databases"
	"kwil/pkg/fund/accounts"
	"kwil/pkg/log"
	"kwil/pkg/types/execution"

	"go.uber.org/zap"
	"google.golang.org/grpc"
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

type Client struct {
	infoClt    accountpb.AccountServiceClient
	txClt      txpb.TxServiceClient
	pricingClt pricingpb.PricingServiceClient

	log  log.Logger
	conn *grpc.ClientConn
	cfg  *Config
}

func NewClient(ctx context.Context, cfg *Config, log log.Logger, conn grpc.ClientConnInterface, infoClt accountpb.AccountServiceClient, txClt txpb.TxServiceClient, pricingClt pricingpb.PricingServiceClient) GrpcClient {
	return &Client{
		infoClt:    infoClt,
		txClt:      txClt,
		pricingClt: pricingClt,
		conn:       conn.(*grpc.ClientConn),
		cfg:        cfg,
		log:        log,
	}
}

func New(ctx context.Context, cfg *Config, log log.Logger) (*Client, error) {
	log.Debug("dail grpc server", zap.String("endpoint", cfg.Endpoint))
	conn, err := transport.Dial(ctx, cfg.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to dial server %s: %w", cfg.Endpoint, err)
	}
	return &Client{
		infoClt:    accountpb.NewAccountServiceClient(conn),
		txClt:      txpb.NewTxServiceClient(conn),
		pricingClt: pricingpb.NewPricingServiceClient(conn),
		cfg:        cfg,
		conn:       conn,
		log:        log,
	}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}
