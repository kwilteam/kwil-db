package sql

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestIsFatalDBError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
		{
			name: "ErrDBFailure",
			err:  ErrDBFailure,
			want: true,
		},
		{
			name: "wrapped ErrDBFailure",
			err:  errors.Join(errors.New("wrapped"), ErrDBFailure),
			want: true,
		},
		{
			name: "insufficient resources error",
			err:  &pgconn.PgError{Code: "53100"},
			want: true,
		},
		{
			name: "system error",
			err:  &pgconn.PgError{Code: "58000"},
			want: true,
		},
		{
			name: "internal error",
			err:  &pgconn.PgError{Code: "XX000"},
			want: true,
		},
		{
			name: "non-fatal pg error",
			err:  &pgconn.PgError{Code: "23505"},
			want: false,
		},
		{
			name: "generic error",
			err:  errors.New("some error"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsFatalDBError(tt.err); got != tt.want {
				t.Errorf("IsFatalDBError() = %v, want %v", got, tt.want)
			}
		})
	}
}
