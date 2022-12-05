package sqlclient

import (
	"context"
	"database/sql"
	"fmt"
)

type DB struct {
	*sql.DB
}

func Open(conn string) (*DB, error) {
	db, err := sql.Open("postgres", conn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	return &DB{db}, nil
}

func (db *DB) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	rows, err := db.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query database: %w", err)
	}
	return rows, nil
}

func (db *DB) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return db.DB.QueryRowContext(ctx, query, args...)
}

func (db *DB) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	res, err := db.DB.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	return res, nil
}

func (db *DB) BeginTx(ctx context.Context) (*sql.Tx, error) {
	tx, err := db.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	return tx, nil
}

// ExecTx executes the string as a transaction.
func (db *DB) ExecTx(query string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec(query)
	if err != nil {
		if err := tx.Rollback(); err != nil {
			return err
		}
		return err
	}

	return tx.Commit()
}

func (db *DB) Close() error {
	return db.DB.Close()
}
