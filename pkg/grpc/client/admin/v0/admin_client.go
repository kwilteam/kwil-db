package admin

import (
	"context"
	"crypto/tls"

	admpb "github.com/kwilteam/kwil-db/api/protobuf/admin/v0"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// AdminClient manages a connection to an authenticated node administrative gRPC
// service.
type AdminClient struct {
	admClient admpb.AdminServiceClient
	conn      *grpc.ClientConn
}

// New constructs an AdminClient with the provided TLS configuration
func New(target string, tlsCfg *tls.Config, opts ...grpc.DialOption) (*AdminClient, error) {
	opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(tlsCfg)))
	conn, err := grpc.Dial(target, opts...)
	if err != nil {
		return nil, err
	}
	return &AdminClient{
		admClient: admpb.NewAdminServiceClient(conn),
		conn:      conn,
	}, nil
}

func (c *AdminClient) Close() error {
	return c.conn.Close()
}

func (c *AdminClient) Ping(ctx context.Context) (string, error) {
	resp, err := c.admClient.Ping(ctx, &admpb.PingRequest{})
	if err != nil {
		return "", err
	}
	return resp.Message, nil
}

func (c *AdminClient) Version(ctx context.Context) (string, error) {
	resp, err := c.admClient.Version(ctx, &admpb.VersionRequest{})
	if err != nil {
		return "", err
	}
	return resp.VersionString, nil
}
