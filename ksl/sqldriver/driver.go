package sqldriver

import (
	"context"
	"fmt"
	"io"
	"ksl/sqlschema"

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

type Caller interface {
	Call(ctx context.Context, name string, args ...any) (any, error)
}

type Driver interface {
	ServerVersion(ctx context.Context) (string, error)

	Locker
	ExecQuerier
	sqlschema.Planner
	sqlschema.Differ
	sqlschema.Describer
	sqlschema.Migrator
	io.Closer
}
