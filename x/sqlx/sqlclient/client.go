package sqlclient

import (
	"context"
	"database/sql"
	"fmt"
	"kwil/x"
	"kwil/x/utils"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type DBTX interface {
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	PrepareContext(context.Context, string) (*sql.Stmt, error)
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}

type DB struct {
	*sql.DB
}

func Open(conn string, duration time.Duration) (*DB, error) {
	return open(conn, *x.NewDeadline(duration))
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

func (db *DB) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	stmt, err := db.DB.PrepareContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}
	return stmt, nil
}

func (db *DB) Close() error {
	return db.DB.Close()
}

func open(conn string, deadline x.Deadline) (*DB, error) {
	var outerErr error
	for {
		c, err := tryOpen(conn)
		if err == nil {
			return c, nil
		}

		outerErr = err
		if deadline.HasExpired() {
			break
		}

		timeout := time.Duration(utils.Min(500, deadline.RemainingMillis())) * time.Millisecond
		time.Sleep(timeout)
	}

	return nil, outerErr
}

func tryOpen(conn string) (*DB, error) {
	db, err := sql.Open("pgx", conn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	return &DB{db}, nil
}
