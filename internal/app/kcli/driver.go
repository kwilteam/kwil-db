package kcli

import (
	"context"
	"database/sql"
	"fmt"
	"kwil/pkg/fund"
	"kwil/pkg/sql/sqlclient"
	anytype "kwil/x/types/data_types/any_type"
	"kwil/x/types/databases"

	"strings"
	"sync"
	"time"

	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

// Driver is a driver for the grpc client for integration tests
type Driver struct {
	Addr string

	connOnce   sync.Once
	conn       *grpc.ClientConn
	client     *KwilClient
	fundConfig *fund.Config
}

func (d *Driver) DeployDatabase(ctx context.Context, db *databases.Database[[]byte]) error {
	client, err := d.getClient()
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	_, err = client.DeployDatabase(ctx, db)
	return err
}

func (d *Driver) DatabaseShouldExists(ctx context.Context, owner string, dbName string) error {
	client, err := d.getClient()
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	schema, err := client.GetDatabaseSchema(ctx, owner, dbName)
	if err != nil {
		return fmt.Errorf("failed to get database schema: %w", err)
	}

	if strings.ToLower(schema.Owner) == strings.ToLower(owner) && schema.Name == dbName {
		return nil
	} else {
		return fmt.Errorf("database does not exist")
	}
}

func (d *Driver) ExecuteQuery(ctx context.Context, owner string, dbName string, queryName string, queryInputs []any) error {
	client, err := d.getClient()
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	ins := make([]anytype.KwilAny, len(queryInputs))
	for i := 0; i < len(queryInputs); i++ {
		ins[i], err = anytype.New(queryInputs[i])
		if err != nil {
			return fmt.Errorf("failed to create query input: %w", err)
		}
	}

	_, err = client.ExecuteDatabase(ctx, owner, dbName, queryName, ins)
	return err
}

func (d *Driver) DropDatabase(ctx context.Context, owner string, dbName string) error {
	client, err := d.getClient()
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	_, err = client.DropDatabase(ctx, owner, dbName)
	return err
}

func (d *Driver) QueryDatabase(ctx context.Context, _sql string, args ...interface{}) (*sql.Row, error) {
	client, err := sqlclient.Open(viper.GetString("PG_DATABASE_URL"), 3*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to open sql client: %w", err)
	}
	defer client.Close()

	return client.QueryRow(ctx, _sql, args...), nil
}

func (d *Driver) Close() error {
	if d.conn != nil {
		return d.conn.Close()
	}
	return nil
}

func (d *Driver) GetFundConfig() *fund.Config {
	return d.fundConfig
}

func (d *Driver) SetFundConfig(cfg *fund.Config) {
	d.fundConfig = cfg
}

func (d *Driver) getClient() (*KwilClient, error) {
	var err error
	d.connOnce.Do(func() {
		d.conn, err = grpc.Dial(d.Addr, grpc.WithInsecure())
		//d.conn, err = grpc.Dial(d.Addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return
		}
		//d.client, err = New(d.conn, d.fundConfig)
		d.client, err = New(context.Background(), nil)
	})
	return d.client, err
}
