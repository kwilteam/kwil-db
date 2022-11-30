package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"ksl/sqlclient"
	"ksl/sqldriver"
	"ksl/sqlx"
	"net/url"
	"time"
)

const DriverName = "postgres"

func init() {
	sqlclient.Register(
		DriverName,
		sqlclient.OpenerFunc(opener),
		sqlclient.RegisterDriverOpener(driverOpen),
		sqlclient.RegisterURLParser(urlParser{}),
	)
}

func driverOpen(db sqldriver.ExecQuerier) (sqldriver.Driver, error) {
	return NewClient(db), nil
}

func opener(u *url.URL) (*sqlclient.Client, error) {
	ur := urlParser{}.ParseURL(u)
	db, err := sql.Open("pgx", ur.DSN)
	if err != nil {
		return nil, err
	}

	drv, err := driverOpen(db)
	if err != nil {
		if cerr := db.Close(); cerr != nil {
			err = fmt.Errorf("%w: %v", err, cerr)
		}
		return nil, err
	}
	return &sqlclient.Client{
		Name:   DriverName,
		DB:     db,
		URL:    ur,
		Driver: drv,
	}, nil
}

type urlParser struct{}

func (urlParser) ParseURL(u *url.URL) *sqlclient.URL {
	q := u.Query()
	schema := q.Get("schema")
	q.Del("schema")
	if schema != "" {
		q.Set("search_path", schema)
	}
	u.RawQuery = q.Encode()
	return &sqlclient.URL{URL: u, DSN: u.String(), Schema: schema}
}

// ChangeSchema implements the sqlclient.SchemaChanger interface.
func (urlParser) ChangeSchema(u *url.URL, s string) *url.URL {
	nu := *u
	q := nu.Query()
	q.Set("search_path", s)
	nu.RawQuery = q.Encode()
	return &nu
}

func lockAcquire(ctx context.Context, conn sqldriver.ExecQuerier, id uint32, timeout time.Duration) error {
	switch {
	// With timeout (context-based).
	case timeout > 0:
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
		fallthrough
	// Infinite timeout.
	case timeout < 0:
		rows, err := conn.QueryContext(ctx, "SELECT pg_advisory_lock($1)", id)
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			err = sqldriver.ErrLocked
		}
		if err != nil {
			return err
		}
		return rows.Close()
	// No timeout.
	default:
		rows, err := conn.QueryContext(ctx, "SELECT pg_try_advisory_lock($1)", id)
		if err != nil {
			return err
		}
		acquired, err := sqlx.ScanNullBool(rows)
		if err != nil {
			return err
		}
		if !acquired.Bool {
			return sqldriver.ErrLocked
		}
		return nil
	}
}
