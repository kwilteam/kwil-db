package pg

// This file contains functions and variables for verification of the version
// and system settings of a postgres instance to be used by kwild.

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
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

func wantMinIntFn(wantMin int64) settingValidFn {
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

func wantStringFn(want string) settingValidFn {
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

func wantOnFn(on bool) settingValidFn {
	if on {
		return wantStringFn("on")
	}
	return wantStringFn("off")
}

var settingValidations = map[string]settingValidFn{
	"wal_level": wantStringFn("logical"),

	// There is one instance of the DB type that requires the a replication
	// slot to precommit: the one used by TxApp for processing blockchain
	// transactions. Require some extra for external debugging.

	"max_wal_senders":           wantMinIntFn(10),
	"max_replication_slots":     wantMinIntFn(10),
	"max_prepared_transactions": wantMinIntFn(2),

	"synchronous_commit": wantOnFn(true),
	"fsync":              wantOnFn(true),
	"max_connections":    wantMinIntFn(50),
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
