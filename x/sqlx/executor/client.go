package executor

import (
	"context"
	"fmt"
	"ksl/sqlclient"
	"kwil/x/metadata"
	"kwil/x/sqlx/schema"
)

type Client interface {
	ExecuteTx()
	Read()
	DeployDDL(ctx context.Context, ddl []byte) error
	DeleteDB()
}

type client struct {
	mp *metadata.ConnectionProvider
}

func NewClient(mp *metadata.ConnectionProvider) *client {
	return &client{
		mp: mp,
	}
}

func (c *client) ExecuteTx() {
}

func (c *client) Read(ctx context.Context, q string) error {
	db, err := c.getDB("kwil")
	if err != nil {
		return err
	}
	defer db.Close()

	res, err := db.QueryContext(ctx, q)
	if err != nil {
		return err
	}
	defer res.Close()

	for res.Next() {
		var id int
		var name string
		var b string
		var c string
		err = res.Scan(&id, &name, &b, &c)
		if err != nil {
			return err
		}
		fmt.Println(id, name)
	}

	return nil
}

func (c *client) DeployDDL(ctx context.Context, ddl []byte) error {
	yml, err := schema.ReadYaml(ddl)
	if err != nil {
		return fmt.Errorf("failed to read yaml: %w", err)
	}
	err = yml.Validate()
	if err != nil {
		return fmt.Errorf("failed to validate ddl: %w", err)
	}

	// create database
	err = c.createDB(ctx, yml.Owner, yml.Name)
	if err != nil {
		if err == ErrDBExists {
			return ErrDBExists
		}
		return fmt.Errorf("failed to create database: %w", err)
	}

	stmts, err := yml.GenerateDDL()
	if err != nil {
		return fmt.Errorf("failed to generate ddl: %w", err)
	}

	db, err := c.getDB(yml.Owner)
	if err != nil {
		return fmt.Errorf("failed to get db: %w", err)
	}
	defer db.Close()

	_, err = db.ExecContext(ctx, stmts)
	if err != nil {
		return fmt.Errorf("failed to execute ddl: %w", err)
	}

	return nil
}

func (c *client) DeleteDB() {
}

func (c *client) getDB(wallet string) (*sqlclient.Client, error) {
	conn, err := c.mp.GetConnectionInfo(wallet)
	if err != nil {
		return nil, err
	}
	db, err := sqlclient.Open(conn)
	if err != nil {
		return nil, err
	}
	return db, nil
}
