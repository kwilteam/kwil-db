package executor

import (
	"context"
	"errors"
)

var (
	ErrDBExists    = errors.New("database already exists")
	ErrDBNotExists = errors.New("database does not exist")
)

// createDB will create a schema in the database named owner_name
func (c *client) createDB(ctx context.Context, owner, name string) error {
	db, err := c.getDB("public")
	if err != nil {
		return err
	}

	nm := owner + "_" + name

	// check if the schema already exists
	var exists bool
	err = db.DB.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM pg_namespace WHERE nspname = $1)", nm).Scan(&exists)
	if err != nil {
		return err
	}
	if exists {
		return ErrDBExists
	}

	// create the schema
	_, err = db.DB.ExecContext(ctx, "CREATE SCHEMA "+nm)
	if err != nil {
		return err
	}

	return nil
}

// dropDB will drop a schema in the database named owner_name
func (c *client) dropDB(ctx context.Context, owner, name string) error {
	db, err := c.getDB("public")
	if err != nil {
		return err
	}

	nm := owner + "_" + name

	// check if the schema already exists
	var exists bool
	err = db.DB.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM pg_namespace WHERE nspname = $1)", nm).Scan(&exists)
	if err != nil {
		return err
	}
	if !exists {
		return ErrDBNotExists
	}

	// drop the schema
	_, err = db.DB.ExecContext(ctx, "DROP SCHEMA "+nm+" CASCADE")
	if err != nil {
		return err
	}

	return nil
}
