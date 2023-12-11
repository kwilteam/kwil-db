package sqlite

import (
	"context"
	"fmt"
)

// this file implements a kv in sqlite
const (
	kvTableName  = "_kv"
	createKvStmt = `
		CREATE TABLE IF NOT EXISTS ` + kvTableName + ` (
			key BLOB PRIMARY KEY,
			value BLOB NOT NULL
		) WITHOUT ROWID, STRICT;
	`

	insertKvStmt = `
		INSERT INTO ` + kvTableName + ` (key, value)
		VALUES ($key, $value)
		ON CONFLICT(key) DO UPDATE SET value = $value;
	`

	selectKvStmt = `
		SELECT value
		FROM ` + kvTableName + `
		WHERE key = $key;
	`

	deleteKvStmt = `
		DELETE FROM ` + kvTableName + `
		WHERE key = $key;
	`
)

// initializeKv initializes the kv table, if necessary
func initializeKv(ctx context.Context, conn *Connection) error {
	if conn.isReadonly() {
		return nil
	}

	exists, err := conn.TableExists(ctx, kvTableName)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	return execute(conn.conn, createKvStmt)
}

// Set sets a value for a key.
func (c *Connection) Set(ctx context.Context, key, value []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	res, err := c.execute(ctx, insertKvStmt, map[string]any{
		"$key":   key,
		"$value": value,
	})
	if err != nil {
		return nil
	}

	return res.Finish()
}

// Delete deletes a key.
func (c *Connection) Delete(ctx context.Context, key []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	res, err := c.execute(ctx, deleteKvStmt, map[string]any{
		"$key": key,
	})
	if err != nil {
		return err
	}

	return res.Finish()
}

// Get gets a value for a key.
func (c *Connection) Get(ctx context.Context, key []byte) ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	res, err := c.execute(ctx, selectKvStmt, map[string]any{
		"$key": key,
	})
	if err != nil {
		return nil, err
	}
	defer res.Finish()

	rowReturned := res.Next()
	if !rowReturned {
		return nil, nil
	}

	if res.Err() != nil {
		return nil, res.Err()
	}

	values, err := res.Values()
	if err != nil {
		return nil, err
	}

	if len(values) != 1 {
		return nil, fmt.Errorf("expected 1 value when retrieving kv store, got %d", len(values))
	}

	value, ok := values[0].([]byte)
	if !ok {
		return nil, fmt.Errorf("expected value to be []byte, got %T", values[0])
	}

	return value, nil
}
