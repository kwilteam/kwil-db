package grpc

import (
	"context"
	"fmt"
	"os"

	txpb "github.com/kwilteam/kwil-db/core/rpc/protobuf/tx/v1"
	"github.com/kwilteam/kwil-db/core/rpc/transport"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	grpcInsecure "google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	txClient txpb.TxServiceClient
	conn     *grpc.ClientConn

	dailOpts []grpc.DialOption
}

type Option func(*Client) error

func WithTlsCert(certFile string) Option {
	return func(c *Client) error {
		tlsDailOption, err := CreateCertOption(certFile)
		if err != nil {
			return err
		}
		c.dailOpts = append(c.dailOpts, tlsDailOption)
		return nil
	}
}

func WithDialOptions(opts ...grpc.DialOption) Option {
	return func(c *Client) error {
		c.dailOpts = append(c.dailOpts, opts...)
		return nil
	}
}

func New(ctx context.Context, target string, opts ...Option) (*Client, error) {
	clt := &Client{
		dailOpts: []grpc.DialOption{
			grpc.WithDefaultCallOptions(
				grpc.MaxCallRecvMsgSize(64 * 1024 * 1024), // 64MiB limit on *responses*; sends are unlimited
			),
		},
	}

	for _, opt := range opts {
		if err := opt(clt); err != nil {
			return nil, err
		}
	}

	conn, err := grpc.DialContext(ctx, target, clt.dailOpts...)
	if err != nil {
		return nil, err
	}

	clt.txClient = txpb.NewTxServiceClient(conn)
	clt.conn = conn
	return clt, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) GetTarget() string {
	return c.conn.Target()
}

// CreateCertOption returns a grpc.DialOption that can be used to create a
// secure connection to the given certFile. If certFile is empty, the connection
// will be insecure.
func CreateCertOption(certFile string) (grpc.DialOption, error) {
	var transOpt credentials.TransportCredentials
	if certFile == "" {
		transOpt = grpcInsecure.NewCredentials()
	} else {
		pemCerts, err := os.ReadFile(certFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read cert file: %w", err)
		}

		tlsConfig, err := transport.NewClientTLSConfig(pemCerts)
		if err != nil {
			return nil, err
		}
		transOpt = credentials.NewTLS(tlsConfig)
	}

	return grpc.WithTransportCredentials(transOpt), nil
}
