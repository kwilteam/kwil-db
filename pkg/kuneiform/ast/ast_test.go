package ast_test

import (
	"bytes"
	"flag"
	"kwil/pkg/kuneiform/parser"
	"os"
	"path/filepath"
	"testing"
)

var update = flag.Bool("update", false, "update .golden files")

func getGoldenFile(t *testing.T, actual []byte, goldenFile string) []byte {
	golden := filepath.Join("testdata", goldenFile)

	if *update {
		t.Logf("updating golden file %s", goldenFile)
		if err := os.WriteFile(golden, actual, 0644); err != nil {
			t.Fatalf("failed to update golden file: %v", err)
		}
		return actual
	}

	expected, err := os.ReadFile(golden)
	if err != nil {
		t.Fatalf("failed to read golden file: %v", err)
	}
	return expected
}

func TestAst_Generate(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "empty tables",
			input: `database test; table user{} table order{}`,
		},
		{
			name:  "table without attributes",
			input: `database test; table user{username text, age int, email text}`,
		},
		{
			name:  "table with attributes",
			input: `database test; table user{username text notnull, age int min(18) max(30), email text maxlen(50) minlen(10)}`,
		},
		{
			name:  "table with index",
			input: `database demo; table user{name text, age int, email text, #uname unique(name, email), #im index(email)}`,
		},
		{
			name: "table with action insert",
			input: `database demo;
                        table user{name text, age int, email text}
                        action create_user($name, $age) public {insert into user(name, age) values ($name, $age)}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, err := parser.Parse([]byte(tt.input), parser.WithTraceOff())

			if err != nil {
				t.Errorf("Parse() got error: %s", err)
			}

			got := a.Generate()
			want := getGoldenFile(t, got, t.Name()+".golden")
			if !bytes.Equal(got, want) {
				t.Errorf("Generate() = %v,\n            want       = %v", got, want)
			}
		})
	}
}
