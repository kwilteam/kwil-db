package executor

import (
	"context"
	"errors"
	"fmt"
	"kwil/x/sqlx/schema"
	"kwil/x/sqlx/sqlclient"
)

var (
	ErrDBExists    = errors.New("database already exists")
	ErrDBNotExists = errors.New("database does not exist")
	ErrInvalidName = errors.New("invalid database name")
)

type Executor interface {
	ExecuteTx()
	Read()
	DeployDDL(ctx context.Context, ddl []byte) error
	DeleteDB(ctx context.Context, owner, name string) error
	DBExists(ctx context.Context, owner, name string) (bool, error)
}

type client struct {
	db *sqlclient.DB
}

func New(db *sqlclient.DB) *client {
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

	nm := yml.Owner + "_" + yml.Name

	// validate the name
	val, err := schema.CheckValidName(nm)
	if err != nil {
		return err
	}
	if !val {
		return ErrInvalidName
	}

	err = c.NewDB(ctx, nm)
	if err != nil {
		return err
	}

	stmts, err := yml.GenerateDDL()
	if err != nil {
		return fmt.Errorf("failed to generate ddl: %w", err)
	}

	for _, stmt := range stmts {
		_, err = c.db.ExecContext(ctx, stmt)
		if err != nil {
			break // break the loop, it will proceed to delete the schema
		}
	}
	if err != nil {
		err2 := c.DeleteDB(ctx, yml.Owner, yml.Name)
		if err2 != nil {
			return fmt.Errorf("failed to delete database: %w", err2)
		}
		return fmt.Errorf("failed to execute ddl: %w", err)
	}

	// add the queries
	for name, _ := range yml.Queries {
		err = c.AddQuery(ctx, nm, name, []byte("")) // TODO: add the query
		if err != nil {
			return fmt.Errorf("failed to add query: %w", err)
		}
	}

	// add the roles and permissions
	for name, rl := range yml.Roles {
		err = c.AddRole(ctx, nm, name)
		if err != nil {
			return fmt.Errorf("failed to add role: %w", err)
		}
		for _, q := range rl.Queries {
			err = c.AddQueryPermission(ctx, nm, name, q)
			if err != nil {
				return fmt.Errorf("failed to add query permission: %w", err)
			}
		}
	}

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
	fmt.Println("dropping schema", nm)
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
