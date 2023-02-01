package errors

import (
	"database/sql"
	"errors"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
)

func asPgErr(err error) (*pgconn.PgError, bool) {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return nil, false
	}
	return pgErr, true
}

// will detect if an error from postgres is a violation of a unique constraint
func IsUniqueViolation(err error) bool {
	pgErr, ok := asPgErr(err)
	if !ok {
		return false
	}

	if pgErr.Code == pgerrcode.UniqueViolation {
		return true
	}

	return false
}

func IsNoRowsInResult(err error) bool {
	return err == sql.ErrNoRows
}
