package postgres

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kwilteam/kwil-db/internal/schemadef/schema"
	"github.com/kwilteam/kwil-db/internal/sqlx"
	"github.com/stretchr/testify/require"
)

func TestDriver_LockAcquired(t *testing.T) {
	db, m, err := sqlmock.New()
	require.NoError(t, err)
	name, hash := "name", 797654004
	m.ExpectQuery(sqlx.Escape("SELECT pg_try_advisory_lock($1)")).
		WithArgs(hash).
		WillReturnRows(sqlmock.NewRows([]string{"pg_advisory_lock"}).AddRow(1)).
		RowsWillBeClosed()
	m.ExpectQuery(sqlx.Escape("SELECT pg_advisory_unlock($1)")).
		WithArgs(hash).
		WillReturnRows(sqlmock.NewRows([]string{"pg_advisory_unlock"}).AddRow(1)).
		RowsWillBeClosed()

	d := &Driver{}
	d.ExecQuerier = db
	unlock, err := d.Lock(context.Background(), name, 0)
	require.NoError(t, err)
	require.NoError(t, unlock())
	require.NoError(t, m.ExpectationsWereMet())
}

func TestDriver_LockError(t *testing.T) {
	db, m, err := sqlmock.New()
	require.NoError(t, err)
	d := &Driver{}
	d.ExecQuerier = db
	name, hash := "migrate", 979249972

	t.Run("Timeout", func(t *testing.T) {
		m.ExpectQuery(sqlx.Escape("SELECT pg_advisory_lock($1)")).
			WithArgs(hash).
			WillReturnError(context.DeadlineExceeded).
			RowsWillBeClosed()
		unlock, err := d.Lock(context.Background(), name, time.Minute)
		require.Equal(t, sqlx.ErrLocked, err)
		require.Nil(t, unlock)
	})

	t.Run("Internal", func(t *testing.T) {
		m.ExpectQuery(sqlx.Escape("SELECT pg_advisory_lock($1)")).
			WithArgs(hash).
			WillReturnError(io.EOF).
			RowsWillBeClosed()
		unlock, err := d.Lock(context.Background(), name, time.Minute)
		require.Equal(t, io.EOF, err)
		require.Nil(t, unlock)
	})
}

func TestDriver_UnlockError(t *testing.T) {
	db, m, err := sqlmock.New()
	require.NoError(t, err)
	d := &Driver{}
	d.ExecQuerier = db
	name, hash := "up", 1551306158
	acquired := func() {
		m.ExpectQuery(sqlx.Escape("SELECT pg_try_advisory_lock($1)")).
			WithArgs(hash).
			WillReturnRows(sqlmock.NewRows([]string{"pg_try_advisory_lock"}).AddRow(1)).
			RowsWillBeClosed()
	}

	t.Run("NotHeld", func(t *testing.T) {
		acquired()
		unlock, err := d.Lock(context.Background(), name, 0)
		require.NoError(t, err)
		m.ExpectQuery(sqlx.Escape("SELECT pg_advisory_unlock($1)")).
			WithArgs(hash).
			WillReturnRows(sqlmock.NewRows([]string{"pg_advisory_unlock"}).AddRow(0)).
			RowsWillBeClosed()
		require.Error(t, unlock())
	})

	t.Run("Internal", func(t *testing.T) {
		acquired()
		unlock, err := d.Lock(context.Background(), name, 0)
		require.NoError(t, err)
		m.ExpectQuery(sqlx.Escape("SELECT pg_advisory_unlock($1)")).
			WithArgs(hash).
			WillReturnRows(sqlmock.NewRows([]string{"pg_advisory_unlock"}).AddRow(nil)).
			RowsWillBeClosed()
		require.Error(t, unlock())
	})
}

type mockInspector struct {
	schema.Inspector
	realm  *schema.Database
	schema *schema.Schema
}

func (m *mockInspector) InspectSchema(context.Context, string, *schema.InspectOptions) (*schema.Schema, error) {
	if m.schema == nil {
		return nil, &sqlx.NotExistError{}
	}
	return m.schema, nil
}

func (m *mockInspector) InspectRealm(context.Context, *schema.InspectRealmOption) (*schema.Database, error) {
	return m.realm, nil
}
