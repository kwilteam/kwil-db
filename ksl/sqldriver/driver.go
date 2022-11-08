package sqldriver

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"hash/fnv"
	"net/url"
	"strconv"
	"time"

	"ksl/sqlclient"
	"ksl/sqlutil"

	"ksl/sqlspec"
)

type (
	Driver struct {
		sqlspec.ExecQuerier
		sqlspec.Differ
		sqlspec.Inspector
		sqlspec.Planner
		schema string
	}
)

var _ sqlspec.PlanApplier = (*Driver)(nil)

// DriverName holds the name used for registration.
const DriverName = "postgres"

func init() {
	sqlclient.Register(
		DriverName,
		sqlclient.OpenerFunc(opener),
		sqlclient.RegisterDriverOpener(Open),
		sqlclient.RegisterFlavours("postgresql"),
		sqlclient.RegisterURLParser(urlParser{}),
	)
}

func opener(_ context.Context, u *url.URL) (*sqlclient.Client, error) {
	ur := urlParser{}.ParseURL(u)
	db, err := sql.Open(DriverName, ur.DSN)
	if err != nil {
		return nil, err
	}
	drv, err := Open(db)
	if err != nil {
		if cerr := db.Close(); cerr != nil {
			err = fmt.Errorf("%w: %v", err, cerr)
		}
		return nil, err
	}
	drv.(*Driver).schema = ur.Schema
	return &sqlclient.Client{
		Name:   DriverName,
		DB:     db,
		URL:    ur,
		Driver: drv,
	}, nil
}

// Open opens a new PostgreSQL driver.
func Open(db sqlspec.ExecQuerier) (sqlspec.Driver, error) {
	rows, err := db.QueryContext(context.Background(), paramsQuery)
	if err != nil {
		return nil, fmt.Errorf("postgres: scanning system variables: %w", err)
	}
	params, err := sqlutil.ScanStrings(rows)
	if err != nil {
		return nil, fmt.Errorf("postgres: failed scanning rows: %w", err)
	}
	if len(params) != 3 && len(params) != 4 {
		return nil, fmt.Errorf("postgres: unexpected number of rows: %d", len(params))
	}
	ctype, collate := params[1], params[2]
	var version int
	if version, err = strconv.Atoi(params[0]); err != nil {
		return nil, fmt.Errorf("postgres: malformed version: %s: %w", params[0], err)
	}
	if version < 10_00_00 {
		return nil, fmt.Errorf("postgres: unsupported postgres version: %d", version)
	}

	return &Driver{
		ExecQuerier: db,
		Differ:      sqlspec.NewDiffer(),
		Inspector:   sqlspec.NewInspector(db, version, ctype, collate),
		Planner:     sqlspec.NewPlanner(),
	}, nil
}

// ApplyChanges applies the changes on the database. An error is returned
// if the driver is unable to produce a plan to do so, or one of the statements
// is failed or unsupported.
func (d *Driver) ApplyChanges(ctx context.Context, changes []sqlspec.SchemaChange, opts ...sqlspec.PlanOption) error {
	return sqlspec.ApplyChanges(ctx, changes, d, opts...)
}

// Lock implements the sqlspec.Locker interface.
func (d *Driver) Lock(ctx context.Context, name string, timeout time.Duration) (sqlutil.UnlockFunc, error) {
	conn, err := sqlutil.SingleConn(ctx, d.ExecQuerier)
	if err != nil {
		return nil, err
	}
	h := fnv.New32()
	h.Write([]byte(name))
	id := h.Sum32()
	if err := acquire(ctx, conn, id, timeout); err != nil {
		conn.Close()
		return nil, err
	}
	return func() error {
		defer conn.Close()
		rows, err := conn.QueryContext(ctx, "SELECT pg_advisory_unlock($1)", id)
		if err != nil {
			return err
		}
		switch released, err := sqlutil.ScanNullBool(rows); {
		case err != nil:
			return err
		case !released.Valid || !released.Bool:
			return fmt.Errorf("sql/postgres: failed releasing lock %d", id)
		}
		return nil
	}, nil
}

func acquire(ctx context.Context, conn sqlutil.ExecQuerier, id uint32, timeout time.Duration) error {
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
			err = sqlutil.ErrLocked
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
		acquired, err := sqlutil.ScanNullBool(rows)
		if err != nil {
			return err
		}
		if !acquired.Bool {
			return sqlutil.ErrLocked
		}
		return nil
	}
}

type urlParser struct{}

// ParseURL implements the sqlclient.URLParser interface.
func (urlParser) ParseURL(u *url.URL) *sqlclient.URL {
	return &sqlclient.URL{URL: u, DSN: u.String(), Schema: u.Query().Get("search_path")}
}

// ChangeSchema implements the sqlclient.SchemaChanger interface.
func (urlParser) ChangeSchema(u *url.URL, s string) *url.URL {
	nu := *u
	q := nu.Query()
	q.Set("search_path", s)
	nu.RawQuery = q.Encode()
	return &nu
}

const paramsQuery = `SELECT setting FROM pg_settings WHERE name IN ('lc_collate', 'lc_ctype', 'server_version_num', 'crdb_version') ORDER BY name DESC`
