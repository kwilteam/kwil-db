package executor

import (
	"context"
	"errors"
	"fmt"
	"kwil/x/sqlx/schema"
	"kwil/x/sqlx/schema_manager"
	"kwil/x/sqlx/sqlclient"
)

type Executor interface {
	ExecuteTx()
	Read()
	DeployDDL(ctx context.Context, ddl []byte) error
	DeleteDB(ctx context.Context, owner, name string) error
	DBExists(ctx context.Context, owner, name string) (bool, error)
}

type client struct {
	manager schema_manager.Manager
}

func New(db *sqlclient.DB) *client {
	return &client{
		manager: schema_manager.New(db),
	}
}

func (c *client) ExecuteTx() {
	panic("implement me")
}

func (c *client) Read() {
	panic("implement me")
}

func (c *client) DeployDDL(ctx context.Context, ddl []byte) error {
	db, err := schema.MarshalDatabase(ddl)
	if err != nil {
		return fmt.Errorf("failed to read yaml: %w", err)
	}
	err = db.Validate()
	if err != nil {
		return fmt.Errorf("failed to validate ddl: %w", err)
	}

	return db.Store(ctx, c.manager)
}

// DeleteDB will drop a schema in the database named owner_name
func (c *client) DeleteDB(ctx context.Context, owner, name string) error {

	nm := owner + "_" + name
	val, err := schema.CheckValidName(nm)
	if err != nil {
		return err
	}
	if !val {
		return fmt.Errorf("invalid name: %s", nm)
	}

	// check if the schema already exists
	exists, err := c.manager.SchemaExists(ctx, nm)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("schema %s does not exist", nm)
	}

	// drop the schema
	return c.manager.DeleteDB(ctx, nm)
}

// DBExists will check if a schema exists in the database named owner_name
func (c *client) DBExists(ctx context.Context, owner, name string) (bool, error) {
	nm := owner + "_" + name
	val, err := schema.CheckValidName(nm)
	if err != nil {
		return false, err
	}
	if !val {
		return false, errors.New("invalid name")
	}

	return c.manager.SchemaExists(ctx, nm)
}
