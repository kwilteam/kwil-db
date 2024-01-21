package mutative_test

import (
	"testing"

	"github.com/kwilteam/kwil-db/internal/engine/sqlanalyzer/mutative"
	sqlparser "github.com/kwilteam/kwil-db/parse/sql"
)

func Test_Mutativity(t *testing.T) {
	type testCase struct {
		name         string
		input        string
		wantMutative bool
	}

	testCases := []testCase{
		{
			name:         "simple select",
			input:        "SELECT * FROM users",
			wantMutative: false,
		},
		{
			name:         "simple insert",
			input:        "INSERT INTO users VALUES (1, 'test')",
			wantMutative: true,
		},
		{
			name:         "simple update",
			input:        "UPDATE users SET name = 'test' WHERE id = 1",
			wantMutative: true,
		},
		{
			name:         "simple delete",
			input:        "DELETE FROM users WHERE id = 1",
			wantMutative: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			stmt, err := sqlparser.Parse(tc.input)
			if err != nil {
				t.Fatalf("failed to parse input: %v", err)
			}

			mutativityWalker := mutative.NewMutativityWalker()

			err = stmt.Walk(mutativityWalker)
			if err != nil {
				t.Fatalf("failed to walk statement: %v", err)
			}

			if mutativityWalker.Mutative != tc.wantMutative {
				t.Fatalf("got mutativity %v, want %v", mutativityWalker.Mutative, tc.wantMutative)
			}
		})
	}
}
