package sqlx

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ValidString reports if the given string is not null and valid.
func ValidString(s sql.NullString) bool {
	return s.Valid && s.String != "" && strings.ToLower(s.String) != "null"
}

// ScanOne scans one record and closes the rows at the end.
func ScanOne(rows *sql.Rows, dest ...any) error {
	defer rows.Close()
	if !rows.Next() {
		return sql.ErrNoRows
	}
	if err := rows.Scan(dest...); err != nil {
		return err
	}
	return rows.Close()
}

// ScanNullBool scans one sql.NullBool record and closes the rows at the end.
func ScanNullBool(rows *sql.Rows) (sql.NullBool, error) {
	var b sql.NullBool
	return b, ScanOne(rows, &b)
}

// ScanStrings scans sql.Rows into a slice of strings and closes it at the end.
func ScanStrings(rows *sql.Rows) ([]string, error) {
	defer rows.Close()
	var vs []string
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		vs = append(vs, v)
	}
	return vs, nil
}

// ValuesEqual checks if the 2 string slices are equal (including their order).
func ValuesEqual(v1, v2 []string) bool {
	if len(v1) != len(v2) {
		return false
	}
	for i := range v1 {
		if v1[i] != v2[i] {
			return false
		}
	}
	return true
}

// IsQuoted reports if the given string is quoted with one of the given quotes (e.g. ', ", `).
func IsQuoted(s string, q ...byte) bool {
	for i := range q {
		if l, r := strings.IndexByte(s, q[i]), strings.LastIndexByte(s, q[i]); l < r && l == 0 && r == len(s)-1 {
			return true
		}
	}
	return false
}

// IsLiteralBool reports if the given string is a valid literal bool.
func IsLiteralBool(s string) bool {
	_, err := strconv.ParseBool(s)
	return err == nil
}

// IsLiteralNumber reports if the given string is a literal number.
func IsLiteralNumber(s string) bool {
	// Hex digits.
	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
		// Some databases allow odd length hex string.
		_, err := strconv.ParseUint(s[2:], 16, 64)
		return err == nil
	}
	// Digits with optional exponent.
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

// MayWrap ensures the given string is wrapped with parentheses.
// Used by the different drivers to turn strings valid expressions.
func MayWrap(s string) string {
	n := len(s) - 1
	if len(s) < 2 || s[0] != '(' || s[n] != ')' || !balanced(s[1:n]) {
		return "(" + s + ")"
	}
	return s
}

func balanced(expr string) bool {
	return ExprLastIndex(expr) == len(expr)-1
}

// ExprLastIndex scans the first expression in the given string until
// its end and returns its last index.
func ExprLastIndex(expr string) int {
	var l, r int
	for i := 0; i < len(expr); i++ {
	Top:
		switch expr[i] {
		case '(':
			l++
		case ')':
			r++
		// String or identifier.
		case '\'', '"', '`':
			for j := i + 1; j < len(expr); j++ {
				switch expr[j] {
				case '\\':
					j++
				case expr[i]:
					i = j
					break Top
				}
			}
			// Unexpected EOS.
			return -1
		}
		// Balanced parens and we reached EOS or a terminator.
		if l == r && (i == len(expr)-1 || expr[i+1] == ',') {
			return i
		} else if r > l {
			return -1
		}
	}
	return -1
}

// P returns a pointer to v.
func P[T any](v T) *T {
	return &v
}

// V returns the value p is pointing to.
// If p is nil, the zero value is returned.
func V[T any](p *T) (v T) {
	if p != nil {
		v = *p
	}
	return
}

// Unquote single or double quotes.
func Unquote(s string) (string, error) {
	switch {
	case IsQuoted(s, '"'):
		return strconv.Unquote(s)
	case IsQuoted(s, '\''):
		return strings.ReplaceAll(s[1:len(s)-1], "''", "'"), nil
	default:
		return s, nil
	}
}

// SingleQuote quotes the given string with single quote.
func SingleQuote(s string) (string, error) {
	switch {
	case IsQuoted(s, '\''):
		return s, nil
	case IsQuoted(s, '"'):
		v, err := strconv.Unquote(s)
		if err != nil {
			return "", err
		}
		s = v
		fallthrough
	default:
		return "'" + strings.ReplaceAll(s, "'", "''") + "'", nil
	}
}

// Escape escapes all regular expression metacharacters in the given query.
func Escape(query string) string {
	rows := strings.Split(query, "\n")
	for i := range rows {
		rows[i] = strings.TrimPrefix(rows[i], " ")
	}
	query = strings.Join(rows, " ")
	return strings.TrimSpace(regexp.QuoteMeta(query)) + "$"
}

// A NotExistError wraps another error to retain its original text
// but makes it possible to the migrator to catch it.
type NotExistError struct {
	Err error
}

func (e NotExistError) Error() string { return e.Err.Error() }

// IsNotExistError reports if an error is a NotExistError.
func IsNotExistError(err error) bool {
	if err == nil {
		return false
	}
	var e *NotExistError
	return errors.As(err, &e)
}

// ExecQuerier wraps the two standard sql.DB methods.
type ExecQuerier interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

type (
	// ExecQueryCloser is the interface that groups
	// Close with the schema.ExecQuerier methods.
	ExecQueryCloser interface {
		ExecQuerier
		io.Closer
	}
	nopCloser struct {
		ExecQuerier
	}
)

// Close implements the io.Closer interface.
func (nopCloser) Close() error { return nil }

// SingleConn returns a closable single connection from the given ExecQuerier.
// If the ExecQuerier is already bound to a single connection (e.g. Tx, Conn),
// the connection will return as-is with a NopCloser.
func SingleConn(ctx context.Context, conn ExecQuerier) (ExecQueryCloser, error) {
	// A standard sql.DB or a wrapper of it.
	if opener, ok := conn.(interface {
		Conn(context.Context) (*sql.Conn, error)
	}); ok {
		return opener.Conn(ctx)
	}
	// Tx and Conn are bounded to a single connection.
	// We use sql/driver.Tx to cover also custom Tx structs.
	_, ok1 := conn.(driver.Tx)
	_, ok2 := conn.(*sql.Conn)
	if ok1 || ok2 {
		return nopCloser{ExecQuerier: conn}, nil
	}
	return nil, fmt.Errorf("cannot obtain a single connection from %T", conn)
}

type (
	// UnlockFunc is returned by the Locker to explicitly
	// release the named "advisory lock".
	UnlockFunc func() error

	// Locker is an interface that is optionally implemented by the different drivers
	// for obtaining an "advisory lock" with the given name.
	Locker interface {
		// Lock acquires a named "advisory lock", using the given timeout. Negative value means no timeout,
		// and the zero value means a "try lock" mode. i.e. return immediately if the lock is already taken.
		// The returned unlock function is used to release the advisory lock acquired by the session.
		//
		// An ErrLocked is returned if the operation failed to obtain the lock in all different timeout modes.
		Lock(ctx context.Context, name string, timeout time.Duration) (UnlockFunc, error)
	}
)
