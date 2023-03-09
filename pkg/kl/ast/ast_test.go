package ast_test

import (
	"bytes"
	"flag"
	"kwil/pkg/kl/parser"
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
			name:  "one table with three columns",
			input: `database test{table user{user_id int notnull,username string null,gender bool}}`,
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
				t.Errorf("Generate() = %v,\nwant = %v", got, want)
			}
		})
	}
}
