package sqlspec_test

import (
	"strings"
	"testing"
	"unicode"

	"github.com/stretchr/testify/require"
	"ksl/sqlspec"
)

func TestMarshal(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "simple",
			input: "table foo { id: int64 }",
		},
		{
			name:  "simple with default",
			input: "table foo { id: int64 @default(1) }",
		},
		{
			name:  "simple with default and unique",
			input: "table foo { id: int64 @default(1) @unique }",
		},
		{
			name:  "simple with id",
			input: "table foo { id: int64 @id }",
		},
		{
			name: "simple with id and default",
			input: `table foo {
				id: int64 @id
				name: string @default("foo") @size(256)
			}`,
		},
		{
			name: "foo and bar with foreign key",
			input: `table foo {
				id: int64 @id
				name: string @default("foo") @size(256) @unique
			}
			table bar {
				id: int64 @id
				foo_id: int64 @foreign_key(foo.id, on_delete="CASCADE", on_update="SET NULL")
			}
			`,
		},
		{
			name: "foo and bar with foreign key and index",
			input: `table foo {
				id: int64 @id
				name: string @default("foo") @size(256) @unique
			}
			table bar {
				id: int64
				name: string @size(1024)
				@@foreign_key(columns=[id, name], references=[foo.id, foo.name], on_delete="CASCADE", on_update="SET NULL")
				@@id([id, name])
			}
			`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			realm, diags := sqlspec.Unmarshal([]byte(test.input), "test.kwil")
			require.Empty(t, diags)

			data, err := sqlspec.MarshalSpec(realm)
			require.NoError(t, err)

			expected := removeWs(test.input)
			actual := removeWs(string(data))
			require.Equal(t, expected, actual)
		})
	}
}

func removeWs(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, s)
}
