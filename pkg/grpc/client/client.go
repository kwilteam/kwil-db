package client

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	accountpb "kwil/api/protobuf/kwil/account/v0/gen/go"
	cfgpb "kwil/api/protobuf/kwil/configuration/v0/gen/go"
	pricingpb "kwil/api/protobuf/kwil/pricing/v0/gen/go"
	txpb "kwil/api/protobuf/kwil/tx/v0/gen/go"
	"kwil/internal/pkg/transport"
	"kwil/pkg/log"
)

type Config struct {
	Endpoint string `mapstructure:"endpoint"`
}

type Gr struct {
	infoClt    accountpb.AccountServiceClient
	txClt      txpb.TxServiceClient
	pricingClt pricingpb.PricingServiceClient
	cfgClt     cfgpb.ConfigServiceClient

	log  log.Logger
	conn *grpc.ClientConn
	cfg  *Config
}

// @yaiba TODO: manually declare dependencies
func NewClient(ctx context.Context, cfg *Config, log log.Logger, conn grpc.ClientConnInterface,
	infoClt accountpb.AccountServiceClient, txClt txpb.TxServiceClient,
	pricingClt pricingpb.PricingServiceClient, cfgClt cfgpb.ConfigServiceClient) GrpcClient {
	return &Gr{
		infoClt:    infoClt,
		txClt:      txClt,
		pricingClt: pricingClt,
		cfgClt:     cfgClt,
		conn:       conn.(*grpc.ClientConn),
		cfg:        cfg,
		log:        log,
	}
}

func New(ctx context.Context, cfg *Config, log log.Logger) (*Gr, error) {
	log.Debug("dail grpc server", zap.String("endpoint", cfg.Endpoint))
	conn, err := transport.Dial(ctx, cfg.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to dial server %s: %w", cfg.Endpoint, err)
	}
	return &Gr{
		infoClt:    accountpb.NewAccountServiceClient(conn),
		txClt:      txpb.NewTxServiceClient(conn),
		pricingClt: pricingpb.NewPricingServiceClient(conn),
		cfgClt:     cfgpb.NewConfigServiceClient(conn),
		cfg:        cfg,
		conn:       conn,
		log:        log,
	}, nil
}

func (c *Gr) Close() error {
	return c.conn.Close()
}
