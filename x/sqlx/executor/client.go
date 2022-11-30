package executor

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"kwil/x/sqlx/schema"
)

var (
	ErrDBExists    = errors.New("database already exists")
	ErrDBNotExists = errors.New("database does not exist")
	ErrInvalidName = errors.New("invalid database name")
)

type Client interface {
	ExecuteTx()
	Read()
	DeployDDL(ctx context.Context, ddl []byte) error
	DeleteDB(ctx context.Context, owner, name string) error
	DBExists(ctx context.Context, owner, name string) (bool, error)
}

type client struct {
	db *sql.DB
}

func NewClient(db *sql.DB) *client {
	return &client{
		db: db,
	}
}

func (c *client) ExecuteTx() {
	panic("implement me")
}

func (c *client) Read() {
	panic("implement me")
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

	_, err = c.db.ExecContext(ctx, stmts)
	if err != nil {
		return fmt.Errorf("failed to execute ddl: %w", err)
	}

	// TODO: read in the roles and queries and store them

	return nil
}

// DeleteDB will drop a schema in the database named owner_name
func (c *client) DeleteDB(ctx context.Context, owner, name string) error {

	nm := owner + "_" + name
	val, err := schema.CheckValidName(nm)
	if err != nil {
		return err
	}
	if !val {
		return ErrInvalidName
	}

	// check if the schema already exists
	exists, err := c.schemaExists(ctx, nm)
	if err != nil {
		return err
	}
	if !exists {
		return ErrDBNotExists
	}

	// Delete the schema
	_, err = c.db.ExecContext(ctx, "DROP SCHEMA "+nm+" CASCADE")
	if err != nil {
		return err
	}

	return nil
}

// DBExists will check if a schema exists in the database named owner_name
func (c *client) DBExists(ctx context.Context, owner, name string) (bool, error) {
	nm := owner + "_" + name
	val, err := schema.CheckValidName(nm)
	if err != nil {
		return false, err
	}
	if !val {
		return false, ErrInvalidName
	}

	return c.schemaExists(ctx, nm)
}

func (c *client) schemaExists(ctx context.Context, schema string) (bool, error) {
	var exists bool
	err := c.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM pg_namespace WHERE nspname = $1)", schema).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

// createDB will create a schema in the database named owner_name
func (c *client) createDB(ctx context.Context, owner, name string) error {

	nm := owner + "_" + name

	// check if the schema already exists
	var exists bool
	err := c.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM pg_namespace WHERE nspname = $1)", nm).Scan(&exists)
	if err != nil {
		return err
	}
	if exists {
		return ErrDBExists
	}

	// create the schema
	_, err = c.db.ExecContext(ctx, "CREATE SCHEMA "+nm)
	if err != nil {
		return err
	}

	return nil
}
