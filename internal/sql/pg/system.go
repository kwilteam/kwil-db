package pg

// This file contains functions and variables for verification of the version
// and system settings of a postgres instance to be used by kwild.

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/kwilteam/kwil-db/common/sql"
)

const (
	// sqlPGVersion returns a long version string that includes compilation information.
	// e.g. "PostgreSQL 16.1 (Ubuntu 16.1-1.pgdg22.04+1) on x86_64-pc-linux-gnu, compiled by gcc (Ubuntu 11.4.0-1ubuntu1~22.04) 11.4.0, 64-bit"
	// This is helpful for debugging with this logged.
	sqlPGVersion = `SELECT version();`

	// sqlPGVersionNum returns the integer version that we can parse to check for support.
	sqlPGVersionNum = `SELECT current_setting('server_version_num')::int4;` // e.g. 160001
)

// The supported version of PostgreSQL. We must allow only one major version
// since any changes in behavior that are not expected and coordinated with an
// upgrade can cause consensus failures.
const (
	verMajorRequired = 16
	verMinorRequired = 1
)

func validateVersion(pgVerNum uint32, reqMajor, reqMinor uint32) (major, minor uint32, ok bool) {
	major, minor = pgVerNum/10_000, pgVerNum%10_000
	if major != reqMajor || minor < reqMinor {
		return major, minor, false
	}
	return major, minor, true
}

// setTimezoneUTC sets the postgres connection's time zone to UTC. This is done
// to ensure that when and if we support date and time with TIMESTAMP or
// TIMESTAMPTZ the results are consistent. This only applies to this
// connection's setting, not the entire postgres instance.
func setTimezoneUTC(ctx context.Context, conn *pgx.Conn) error {
	_, err := conn.Exec(ctx, `SET TIME ZONE UTC;`)
	return err
}

// pgVersion retrieves the version of the connected PostgreSQL server. The
// version string is a long version description that starts with "PostgreSQL"
// and includes information about the system and compiler that built it. The
// uint32 number is the mod 10000 encoding of the major.minor version. Use
// validateVersion to decode and validate this numeric version.
func pgVersion(ctx context.Context, conn *pgx.Conn) (ver string, verNum uint32, err error) {
	err = conn.QueryRow(ctx, sqlPGVersion).Scan(&ver)
	if err != nil {
		return
	}
	var verInt4 pgtype.Int4 // scan and convert from TEXT
	err = conn.QueryRow(ctx, sqlPGVersionNum).Scan(&verInt4)
	verNum = uint32(verInt4.Int32)
	return
}

type settingValidFn func(val string) error

func wantIntFn(want int64) settingValidFn { //nolint:unused
	return func(val string) error {
		num, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return err
		}
		if num != want {
			return fmt.Errorf("require %d, but setting is %d", want, num)
		}
		return nil
	}
}

func wantMinIntFn(wantMin int64) settingValidFn { //nolint:unused
	return func(val string) error {
		num, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return err
		}
		if num < wantMin {
			return fmt.Errorf("require at least %d, but setting is %d", wantMin, num)
		}
		return nil
	}
}

func wantMaxIntFn(wantMax int64) settingValidFn { //nolint:unused
	return func(val string) error {
		num, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return err
		}
		if num > wantMax {
			return fmt.Errorf("require at most %d, but setting is %d", wantMax, num)
		}
		return nil
	}
}

func wantStringFn(want string) settingValidFn { //nolint:unused
	want = strings.TrimSpace(want)
	if want == "" {
		panic("empty want string is invalid")
	}
	return func(val string) error {
		if !strings.EqualFold(strings.TrimSpace(val), want) {
			return fmt.Errorf("require %q, but setting is %q", want, val)
		}
		return nil
	}
}

func wantOnFn(on bool) settingValidFn { //nolint:unused
	if on {
		return wantStringFn("on")
	}
	return wantStringFn("off")
}

// orValidFn creates a settings validation function that passes if *any* of the
// conditions pass. This is useful for instance if there are two acceptable
// string values, or if an integer value of either exactly 0 or >50 are
// acceptable as no single min/max condition captures that criteria.
func orValidFn(fns ...settingValidFn) settingValidFn { //nolint:unused
	return func(val string) error {
		var err error
		for i, fn := range fns {
			erri := fn(val)
			if erri == nil {
				return nil
			}
			err = errors.Join(err, fmt.Errorf("condition % 2d: %w", i, erri))
		}
		return errors.Join(errors.New("no condition is satisfied"), err)
	}
}

// andValidFn creates a settings validation function that passes only if *all*
// of the conditions pass. This can be used to define a range, or enumerate a
// list of unacceptable values.
func andValidFn(fns ...settingValidFn) settingValidFn { //nolint:unused
	return func(val string) error {
		var err error
		for _, fn := range fns {
			erri := fn(val)
			if erri != nil {
				return errors.Join(err, erri)
			}
		}
		return nil
	}
}

var settingValidations = map[string]settingValidFn{
	"synchronous_commit": wantOnFn(true),
	"fsync":              wantOnFn(true),
	"max_connections":    wantMinIntFn(50),
	"wal_level":          wantStringFn("logical"),

	// There is one instance of the DB type that requires the a replication
	// slot to precommit: the one used by TxApp for processing blockchain
	// transactions. Require some extra for external debugging.

	"max_wal_senders":           wantMinIntFn(10),
	"max_replication_slots":     wantMinIntFn(10),
	"max_prepared_transactions": wantMinIntFn(2),
	"wal_sender_timeout":        orValidFn(wantIntFn(0), wantMinIntFn(3_600_000)), // ms units, 0 for no limit or 1 hr min

	// We shouldn't have idle abandoned transactions, but we need to investigate
	// the effect of allowing this kind of clean up.
	"idle_in_transaction_timeout": wantIntFn(0), // disable disable idle transaction timeout for now

	// Behavior related settings that must be set properly for determinism
	// https://www.postgresql.org/docs/16/runtime-config-compatible.html
	// The following are the documented defaults, which we enforce.
	"array_nulls":                 wantOnFn(true),                // recognize NULL in array parser, false is for pre-8.2 compat
	"standard_conforming_strings": wantOnFn(true),                // backslashes (\) in string literals are treated literally -- only escape syntax (E'...') is still usable if needed
	"transform_null_equals":       wantOnFn(false),               // do not treat "expr=NULL" as "expr IS NULL"
	"backslash_quote":             wantStringFn("safe_encoding"), // reject escaped single quotes like \', require standard ''
	"lo_compat_privileges":        wantOnFn(false),               // access-controlled large object storage

	// server_encoding is a read-only setting that allows postgres to report the
	// character encoding of the connected database. The default for new
	// databases is set by `initdb` when creating the cluster:
	//   The database cluster will be initialized with locale "en_US.utf8".
	//   The default database encoding has accordingly been set to "UTF8".
	// Or it can be set for a new data base like
	//   CREATE DATABASE ... WITH ENCODING 'UTF8'
	"server_encoding": wantStringFn("UTF8"),
}

type QueryScanner interface {
	QueryScanFn(ctx context.Context, sql string,
		scans []any, fn func() error, args ...any) error
}

func queryRowFunc(ctx context.Context, conn *pgx.Conn, sql string,
	scans []any, fn func() error, args ...any) error {
	rows, _ := conn.Query(ctx, sql, args...)
	_, err := pgx.ForEachRow(rows, scans, fn)
	return err
}

func QueryRowFunc(ctx context.Context, tx sql.Executor, sql string,
	scans []any, fn func() error, args ...any) error {
	conner, ok := tx.(conner)
	if !ok {
		return errors.New("no conn access")
	}
	conn := conner.Conn()
	return queryRowFunc(ctx, conn, sql, scans, fn, args...)
}

func (tx *nestedTx) QueryScanFn(ctx context.Context, sql string,
	scans []any, fn func() error, args ...any) error {

	conn := tx.Conn()
	return queryRowFunc(ctx, conn, sql, scans, fn, args...)
}

type FieldDesc struct {
	Name                 string
	TableOID             uint32
	TableAttributeNumber uint16
	DataTypeOID          uint32
	DataTypeSize         int16
	TypeModifier         int32
	Format               int16
}

func queryRowFuncAny(ctx context.Context, conn *pgx.Conn, sql string,
	fn func(fields []FieldDesc, vals []any) error, args ...any) error {
	rows, _ := conn.Query(ctx, sql, args...)
	fields := rows.FieldDescriptions()
	pgFields := make([]FieldDesc, len(fields))
	for i, f := range fields {
		pgFields[i] = FieldDesc(f)
	}
	defer rows.Close()

	for rows.Next() {
		vals, err := rows.Values()
		if err != nil {
			return err
		}

		err = fn(pgFields, vals)
		if err != nil {
			return err
		}
	}

	return rows.Err()
}

func QueryRowFuncAny(ctx context.Context, tx sql.Executor, sql string,
	fn func([]FieldDesc, []any) error, args ...any) error {
	conner, ok := tx.(conner)
	if !ok {
		return errors.New("no conn access")
	}
	conn := conner.Conn()
	return queryRowFuncAny(ctx, conn, sql, fn, args...)
}

func TextValue(val any) (string, bool) {
	switch str := val.(type) {
	case string:
		return str, true
	case pgtype.Text:
		return str.String, str.Valid
	case *pgtype.Text:
		return str.String, str.Valid
	}
	return "", false
}

type ColInfo struct {
	Pos      int
	Name     string
	DataType string
	Nullable bool
	Default  any
}

func (ci *ColInfo) ScanVal() any {
	if ci.IsInt() {
		var v pgtype.Int8 // int64
		return &v
	}
	if ci.IsText() {
		var v pgtype.Text // string
		return &v
	}
	if ci.IsByteA() {
		var v []byte
		return &v
	}
	if ci.IsNumeric() {
		var v pgtype.Numeric
		return &v
	}
	if ci.IsFloat() {
		var v pgtype.Float8
		return &v
	}
	if ci.IsUUID() {
		var v pgtype.UUID // wrongish default. user could manually go with our types.UUID instead
		return &v
	}
	var v any
	return &v
}

type ColType string

const (
	ColTypeInt     ColType = "int"
	ColTypeText    ColType = "text"
	ColTypeByteA   ColType = "bytea"
	ColTypeUUID    ColType = "uuid"
	ColTypeNumeric ColType = "numeric"
	ColTypeFloat   ColType = "float"
	ColTypeUnknown ColType = "unknown"
)

func (ci *ColInfo) Type() ColType {
	if ci.IsInt() {
		return ColTypeInt
	}
	if ci.IsText() {
		return ColTypeText
	}
	if ci.IsByteA() {
		return ColTypeByteA
	}
	if ci.IsNumeric() {
		return ColTypeNumeric
	}
	if ci.IsFloat() {
		return ColTypeFloat
	}
	if ci.IsUUID() {
		return ColTypeUUID
	}
	return ColTypeUnknown
}

func (ci *ColInfo) IsInt() bool {
	switch strings.ToLower(ci.DataType) {
	case "bigint", "integer", "int", "int2", "int4", "int8":
		return true
	}
	return false
}

func (ci *ColInfo) IsText() bool {
	switch strings.ToLower(ci.DataType) {
	case "text", "varchar":
		return true
	}
	return false
}

func (ci *ColInfo) IsFloat() bool {
	switch strings.ToLower(ci.DataType) {
	case "float32", "float64":
		return true
	}
	return false
}

func (ci *ColInfo) IsUUID() bool {
	switch strings.ToLower(ci.DataType) {
	case "uuid":
		return true
	}
	return false

}

func (ci *ColInfo) IsByteA() bool {
	switch strings.ToLower(ci.DataType) {
	case "bytea":
		return true
	}
	return false
}

func (ci *ColInfo) IsNumeric() bool {
	dt := strings.ToLower(ci.DataType)
	if strings.HasPrefix(dt, "numeric") {
		return true
	}
	return dt == "uint256"
}

func columnInfo(ctx context.Context, conn *pgx.Conn, tbl string) ([]ColInfo, error) {
	var colInfo []ColInfo

	// get column data types
	sql := `SELECT ordinal_position, column_name, data_type, is_nullable, column_default
        FROM information_schema.columns
        WHERE table_name = '` + tbl + `'`

	var pos int
	var colName, dataType string
	var isNullable string
	var colDefault any
	scans := []any{&pos, &colName, &dataType, &isNullable, &colDefault}
	err := queryRowFunc(ctx, conn, sql, scans, func() error {
		colInfo = append(colInfo, ColInfo{pos, colName, dataType,
			strings.EqualFold(isNullable, "yes"), colDefault})
		return nil
	})
	if err != nil {
		return nil, err
	}

	slices.SortFunc(colInfo, func(a, b ColInfo) int {
		return cmp.Compare(a.Pos, b.Pos)
	})

	return colInfo, nil
}

func ColumnInfo(ctx context.Context, tx sql.Executor, tbl string) ([]ColInfo, error) {
	conner, ok := tx.(conner)
	if !ok {
		return nil, errors.New("no conn access")
	}
	conn := conner.Conn()
	return columnInfo(ctx, conn, tbl)
}

func verifySettings(ctx context.Context, conn *pgx.Conn) error {
	checkSettings := make([]string, 0, len(settingValidations))
	for name := range settingValidations {
		checkSettings = append(checkSettings, name)
	}
	// For each setting, get its value and ensure that it passes it's validation function
	rows, _ := conn.Query(ctx, `SELECT name, setting, unit, short_desc, source FROM pg_settings WHERE name = ANY($1);`, checkSettings)
	var name, setting, unit, shortDesc, source pgtype.Text
	scans := []any{&name, &setting, &unit, &shortDesc, &source}
	_, err := pgx.ForEachRow(rows, scans, func() error {
		fn, have := settingValidations[name.String]
		if !have {
			return fmt.Errorf("unexpected setting %q", name.String)
		}
		if !setting.Valid {
			return errors.New("not set")
		}
		err := fn(setting.String)
		if err != nil {
			return fmt.Errorf("failed validation for setting %q (source = %q): %w",
				name.String, source.String, err)
		}
		return nil
	})
	return err
}
