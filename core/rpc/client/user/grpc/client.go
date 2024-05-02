// package grpc implements a gRPC client for the Kwil txsvc client.
package grpc

import (
	"context"
	"fmt"
	"os"

	"github.com/kwilteam/kwil-db/core/rpc/client/user"
	txpb "github.com/kwilteam/kwil-db/core/rpc/protobuf/tx/v1"
	"github.com/kwilteam/kwil-db/core/rpc/transport"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	grpcInsecure "google.golang.org/grpc/credentials/insecure"
)

// Client is the legacy gRPC user/tx service client.
// DEPRECATED: use the JSON-RPC service instead.
type Client struct {
	TxClient txpb.TxServiceClient
	conn     *grpc.ClientConn

	dialOpts []grpc.DialOption
}

var _ user.TxSvcClient = (*Client)(nil)

type Option func(*Client) error

func WithTlsCert(certFile string) Option {
	return func(c *Client) error {
		tlsDailOption, err := CreateCertOption(certFile)
		if err != nil {
			return err
		}
		c.dialOpts = append(c.dialOpts, tlsDailOption)
		return nil
	}
}

func WithDialOptions(opts ...grpc.DialOption) Option {
	return func(c *Client) error {
		c.dialOpts = append(c.dialOpts, opts...)
		return nil
	}
}

func New(ctx context.Context, target string, opts ...Option) (*Client, error) {
	clt := &Client{
		dialOpts: []grpc.DialOption{
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

	conn, err := grpc.DialContext(ctx, target, clt.dialOpts...)
	if err != nil {
		return nil, err
	}

	clt.TxClient = txpb.NewTxServiceClient(conn)
	clt.conn = conn
	return clt, nil
}

// WrapConn wraps an existing grpc.ClientConn with the TxServiceClient.
func WrapConn(conn *grpc.ClientConn) *Client {
	return &Client{
		TxClient: txpb.NewTxServiceClient(conn),
		conn:     conn,
	}
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
