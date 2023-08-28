package sqlsession_test

import (
	"context"
	"testing"

	sqlsession "github.com/kwilteam/kwil-db/pkg/sessions/sql-session"
	sqlTesting "github.com/kwilteam/kwil-db/pkg/sql/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// this tests the issue where, despite being "finished" with resources, new resources
// cannot be seen
func Test_SQLSession(t *testing.T) {
	ctx := context.Background()
	db, td, err := sqlTesting.OpenTestDB("testdb")
	require.NoError(t, err)
	defer td()

	committable := sqlsession.NewSqlCommitable(db)
	err = committable.BeginCommit(ctx)
	require.NoError(t, err)

	// WRITE DATA HERE

	err = db.Execute(ctx, "CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT);", nil)
	require.NoError(t, err)

	err = db.Execute(ctx, "INSERT INTO test (id, name) VALUES (1, 'test');", nil)
	require.NoError(t, err)

	// END WRITE DATA

	wal := newMockAppender()
	err = committable.EndCommit(ctx, wal.Append)
	require.NoError(t, err)

	// phase 2

	err = committable.BeginApply(ctx)
	require.NoError(t, err)

	for _, data := range wal.data {
		err = committable.Apply(ctx, data)
		require.NoError(t, err)
	}

	err = committable.EndApply(ctx)
	require.NoError(t, err)

	// READ DATA HERE
	result, err := db.Query(ctx, "SELECT * FROM test;", nil)
	require.NoError(t, err)

	assert.Equal(t, 1, len(result))
}

func newMockAppender() *mockAppender {
	return &mockAppender{
		data: make([][]byte, 0),
	}
}

type mockAppender struct {
	data [][]byte
}

func (m *mockAppender) Append(data []byte) error {
	m.data = append(m.data, data)
	return nil
}
