package client

import (
	txpb "github.com/kwilteam/kwil-db/api/protobuf/tx/v1"

	"google.golang.org/grpc"
)

type Client struct {
	txClient txpb.TxServiceClient
	conn     *grpc.ClientConn
}

func New(target string, opts ...grpc.DialOption) (*Client, error) {
	conn, err := grpc.Dial(target, opts...)
	if err != nil {
		return nil, err
	}
	return &Client{
		txClient: txpb.NewTxServiceClient(conn),
		conn:     conn,
	}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) GetTarget() string {
	return c.conn.Target()
}
