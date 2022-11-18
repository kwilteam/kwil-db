package sqldriver

import (
	"context"
	"fmt"
	"io"
	"ksl/sqlmigrate"

	"database/sql"
	"database/sql/driver"
	"time"
)

var (
	ErrLocked    = fmt.Errorf("can't acquire lock")
	ErrNotLocked = fmt.Errorf("can't unlock, as not currently locked")
)

type ExecQuerier interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

type ExecQueryCloser interface {
	ExecQuerier
	io.Closer
}
type nopCloser struct {
	ExecQuerier
}

func (nopCloser) Close() error { return nil }

func SingleConn(ctx context.Context, conn ExecQuerier) (ExecQueryCloser, error) {
	if opener, ok := conn.(interface {
		Conn(context.Context) (*sql.Conn, error)
	}); ok {
		return opener.Conn(ctx)
	}
	// Tx and Conn are bounded to a single connection.
	_, ok1 := conn.(driver.Tx)
	_, ok2 := conn.(*sql.Conn)
	if ok1 || ok2 {
		return nopCloser{ExecQuerier: conn}, nil
	}
	return nil, fmt.Errorf("cannot obtain a single connection from %T", conn)
}

type UnlockFunc func() error

type Locker interface {
	Lock(ctx context.Context, name string, timeout time.Duration) (UnlockFunc, error)
}

type Executor interface {
	ExecuteInsert(ctx context.Context, stmt InsertStatement) error
	ExecuteUpdate(ctx context.Context, stmt UpdateStatement) error
	ExecuteDelete(ctx context.Context, stmt DeleteStatement) error
	ExecuteSelect(ctx context.Context, stmt SelectStatement) ([]map[string]any, error)
}

type SelectStatement struct {
	Database string
	Table    string
	Where    map[string]any
}

type InsertStatement struct {
	Database string
	Table    string
	Input    map[string]any
}

type UpdateStatement struct {
	Database string
	Table    string
	Input    map[string]any
	Where    map[string]any
}

type DeleteStatement struct {
	Database string
	Table    string
	Where    map[string]any
}

type Driver interface {
	ServerVersion(ctx context.Context) (string, error)

	Locker
	ExecQuerier
	Executor
	sqlmigrate.Planner
	sqlmigrate.Differ
	sqlmigrate.Describer
	sqlmigrate.Migrator
	io.Closer
}
