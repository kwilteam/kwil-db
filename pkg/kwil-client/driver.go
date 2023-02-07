package kwil_client

import (
	"context"
	"database/sql"
	"fmt"
	"kwil/pkg/sql/sqlclient"
	"kwil/pkg/types/data_types/any_type"
	"kwil/pkg/types/databases"
	"strings"
	"sync"
	"time"

	"google.golang.org/grpc"
)

// Driver is a driver for the grpc client for integration tests
type Driver struct {
	cfg *Config

	connOnce sync.Once
	conn     *grpc.ClientConn
	client   *Client

	// TODO remove this, use graphql
	dbUrl string
}

func NewDriver(cfg *Config, dbUrl string) *Driver {
	return &Driver{
		cfg:   cfg,
		dbUrl: dbUrl,
	}
}

func (d *Driver) DeployDatabase(ctx context.Context, db *databases.Database[[]byte]) error {
	client, err := d.getClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	_, err = client.DeployDatabase(ctx, db)
	return err
}

func (d *Driver) DatabaseShouldExists(ctx context.Context, owner string, dbName string) error {
	client, err := d.getClient(ctx)
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

func (d *Driver) ExecuteQuery(ctx context.Context, dbName string, queryName string, queryInputs []any) error {
	client, err := d.getClient(ctx)
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

	_, err = client.ExecuteDatabase(ctx, dbName, queryName, ins)
	return err
}

func (d *Driver) DropDatabase(ctx context.Context, dbName string) error {
	client, err := d.getClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	_, err = client.DropDatabase(ctx, dbName)
	return err
}

func (d *Driver) QueryDatabase(ctx context.Context, _sql string, args ...interface{}) (*sql.Row, error) {
	client, err := sqlclient.Open(d.dbUrl, 3*time.Second)
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

func (d *Driver) getClient(ctx context.Context) (*Client, error) {
	var err error
	d.connOnce.Do(func() {
		d.client, err = New(ctx, d.cfg)
		if err != nil {
			return
		}
	})
	return d.client, err
}
