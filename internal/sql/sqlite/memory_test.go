package sqlite_test

import (
	"testing"

	"github.com/kwilteam/kwil-db/internal/sql/sqlite"
	"github.com/stretchr/testify/assert"
)

// testing a readonly in-memory database
func Test_InMemory(t *testing.T) {
	type testCase struct {
		name    string
		stmt    string
		args    map[string]any
		want    []map[string]any
		wantErr bool
	}

	cases := []testCase{
		{
			name: "literal - succeed",
			stmt: "SELECT 'hello' AS result",
			want: []map[string]any{
				{"result": "hello"},
			},
		},
		{
			name: "integer literal - succeed",
			stmt: "SELECT 1 AS result",
			want: []map[string]any{
				{"result": int64(1)},
			},
		},
		{
			name: "bind parameter with math - succeed",
			stmt: "SELECT $id + 1 AS result",
			args: map[string]any{
				"$id": 1,
			},
			want: []map[string]any{
				{"result": int64(2)},
			},
		},
		{
			name: "function with bind parameter with math - succeed",
			stmt: "SELECT ABS($id - 100) AS result",
			args: map[string]any{
				"$id": 1,
			},
			want: []map[string]any{
				{"result": int64(99)},
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			conn, err := sqlite.OpenReadOnlyMemory()
			if err != nil {
				t.Fatalf("failed to open readonly memory connection: %v", err)
			}
			defer conn.Close()

			got, err := conn.Query(tt.stmt, tt.args)
			if (err != nil) != tt.wantErr {
				t.Fatalf("got error: %v, want error: %v", err, tt.wantErr)
			}

			assert.ElementsMatch(t, tt.want, got)
		})
	}
}
