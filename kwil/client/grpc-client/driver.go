package grpc_client

import (
	"context"
	"fmt"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"kwil/x/types/databases"
	"kwil/x/types/transactions"

	"sync"
)

// Driver is a driver for the grpc client for integration tests
type Driver struct {
	Addr string

	connOnce sync.Once
	conn     *grpc.ClientConn
	client   *Client
}

func (d *Driver) DeployDatabase(ctx context.Context, db *databases.Database) (*transactions.Response, error) {
	client, err := d.getClient()

	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	return client.DeployDatabase(ctx, db)
}

func (d *Driver) Close() error {
	if d.conn != nil {
		return d.conn.Close()
	}
	return nil
}

func (d *Driver) getClient() (*Client, error) {
	var err error
	d.connOnce.Do(func() {
		d.conn, err = grpc.Dial(d.Addr, grpc.WithInsecure())
		//d.conn, err = grpc.Dial(d.Addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return
		}
		d.client, err = NewClient(d.conn, viper.GetViper())
	})
	return d.client, err
}
