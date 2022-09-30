package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"hash/fnv"
	"net/url"
	"strconv"
	"time"

	"kwil/x/schemadef/schema"
	"kwil/x/sql/sqlclient"
	sqlx "kwil/x/sql/x"
)

type (
	// Driver represents a PostgreSQL driver for introspecting database schemas,
	// generating diff between schema elements and apply migrations changes.
	Driver struct {
		conn
		schema.Differ
		schema.Inspector
		schema.Planner
		schema string // the schema given in the `search_path` parameter (if given)
	}

	// database connection and its information.
	conn struct {
		sqlx.ExecQuerier
		// System variables that are set on `Open`.
		collate string
		ctype   string
		version int
	}
)

var _ schema.PlanApplier = (*Driver)(nil)

// DriverName holds the name used for registration.
const DriverName = "postgres"

func init() {
	sqlclient.Register(
		DriverName,
		sqlclient.OpenerFunc(opener),
		sqlclient.RegisterDriverOpener(Open),
		sqlclient.RegisterFlavours("postgresql"),
		sqlclient.RegisterCodec(MarshalHCL, EvalHCL),
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
func Open(db sqlx.ExecQuerier) (schema.Driver, error) {
	c := conn{ExecQuerier: db}
	rows, err := db.QueryContext(context.Background(), paramsQuery)
	if err != nil {
		return nil, fmt.Errorf("postgres: scanning system variables: %w", err)
	}
	params, err := sqlx.ScanStrings(rows)
	if err != nil {
		return nil, fmt.Errorf("postgres: failed scanning rows: %w", err)
	}
	if len(params) != 3 && len(params) != 4 {
		return nil, fmt.Errorf("postgres: unexpected number of rows: %d", len(params))
	}
	c.ctype, c.collate = params[1], params[2]
	if c.version, err = strconv.Atoi(params[0]); err != nil {
		return nil, fmt.Errorf("postgres: malformed version: %s: %w", params[0], err)
	}
	if c.version < 10_00_00 {
		return nil, fmt.Errorf("postgres: unsupported postgres version: %d", c.version)
	}
	return &Driver{
		conn:      c,
		Differ:    &schema.Diff{DiffDriver: &diff{}},
		Inspector: &inspect{c},
		Planner:   &planner{},
	}, nil
}

// ApplyChanges applies the changes on the database. An error is returned
// if the driver is unable to produce a plan to do so, or one of the statements
// is failed or unsupported.
func (d *Driver) ApplyChanges(ctx context.Context, changes []schema.SchemaChange, opts ...schema.PlanOption) error {
	return schema.ApplyChanges(ctx, changes, d, opts...)
}

// Lock implements the schema.Locker interface.
func (d *Driver) Lock(ctx context.Context, name string, timeout time.Duration) (sqlx.UnlockFunc, error) {
	conn, err := sqlx.SingleConn(ctx, d.ExecQuerier)
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
		switch released, err := sqlx.ScanNullBool(rows); {
		case err != nil:
			return err
		case !released.Valid || !released.Bool:
			return fmt.Errorf("sql/postgres: failed releasing lock %d", id)
		}
		return nil
	}, nil
}

func acquire(ctx context.Context, conn sqlx.ExecQuerier, id uint32, timeout time.Duration) error {
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
			err = sqlx.ErrLocked
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
			return sqlx.ErrLocked
		}
		return nil
	}
}

// supportsIndexInclude reports if the server supports the INCLUDE clause.
func (c *conn) supportsIndexInclude() bool {
	return c.version >= 11_00_00
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

// Standard column types (and their aliases) as defined in
// PostgreSQL codebase/website.
const (
	TypeBit     = "bit"
	TypeBitVar  = "bit varying"
	TypeBoolean = "boolean"
	TypeBool    = "bool" // boolean.
	TypeBytea   = "bytea"

	TypeCharacter = "character"
	TypeChar      = "char" // character
	TypeCharVar   = "character varying"
	TypeVarChar   = "varchar" // character varying
	TypeText      = "text"

	TypeSmallInt = "smallint"
	TypeInteger  = "integer"
	TypeBigInt   = "bigint"
	TypeInt      = "int"  // integer.
	TypeInt2     = "int2" // smallint.
	TypeInt4     = "int4" // integer.
	TypeInt8     = "int8" // bigint.

	TypeCIDR     = "cidr"
	TypeInet     = "inet"
	TypeMACAddr  = "macaddr"
	TypeMACAddr8 = "macaddr8"

	TypeCircle  = "circle"
	TypeLine    = "line"
	TypeLseg    = "lseg"
	TypeBox     = "box"
	TypePath    = "path"
	TypePolygon = "polygon"
	TypePoint   = "point"

	TypeDate          = "date"
	TypeTime          = "time"   // time without time zone
	TypeTimeTZ        = "timetz" // time with time zone
	TypeTimeWTZ       = "time with time zone"
	TypeTimeWOTZ      = "time without time zone"
	TypeTimestamp     = "timestamp" // timestamp without time zone
	TypeTimestampTZ   = "timestamptz"
	TypeTimestampWTZ  = "timestamp with time zone"
	TypeTimestampWOTZ = "timestamp without time zone"

	TypeDouble = "double precision"
	TypeReal   = "real"
	TypeFloat8 = "float8" // double precision
	TypeFloat4 = "float4" // real
	TypeFloat  = "float"  // float(p).

	TypeNumeric = "numeric"
	TypeDecimal = "decimal" // numeric

	TypeSmallSerial = "smallserial" // smallint with auto_increment.
	TypeSerial      = "serial"      // integer with auto_increment.
	TypeBigSerial   = "bigserial"   // bigint with auto_increment.
	TypeSerial2     = "serial2"     // smallserial
	TypeSerial4     = "serial4"     // serial
	TypeSerial8     = "serial8"     // bigserial

	TypeArray       = "array"
	TypeXML         = "xml"
	TypeJSON        = "json"
	TypeJSONB       = "jsonb"
	TypeUUID        = "uuid"
	TypeMoney       = "money"
	TypeInterval    = "interval"
	TypeUserDefined = "user-defined"
)

// List of supported index types.
const (
	IndexTypeBTree      = "BTREE"
	IndexTypeHash       = "HASH"
	IndexTypeGIN        = "GIN"
	IndexTypeGiST       = "GIST"
	IndexTypeBRIN       = "BRIN"
	defaultPagePerRange = 128
)

// List of "GENERATED" types.
const (
	GeneratedTypeAlways    = "ALWAYS"
	GeneratedTypeByDefault = "BY_DEFAULT" // BY DEFAULT.
)

// List of PARTITION KEY types.
const (
	PartitionTypeRange = "RANGE"
	PartitionTypeList  = "LIST"
	PartitionTypeHash  = "HASH"
)
