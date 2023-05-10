package specifications

import (
	"github.com/kwilteam/kwil-db/pkg/kuneiform/parser"
	"github.com/kwilteam/kwil-db/pkg/kuneiform/schema"
	"os"
	"testing"
)

type DatabaseSchemaLoader interface {
	Load(t *testing.T) *schema.Schema
}

type FileDatabaseSchemaLoader struct {
	FilePath string
	Modifier func(db *schema.Schema)
}

func (l *FileDatabaseSchemaLoader) Load(t *testing.T) *schema.Schema {
	t.Helper()

	d, err := os.ReadFile(l.FilePath)
	if err != nil {
		t.Fatal("cannot open database schema file", err)
	}

	ast, err := parser.Parse(d)
	if err != nil {
		t.Fatal("cannot parse database schema", err)
	}

	db := ast.Schema()

	l.Modifier(db)

	return db
}
