package specifications

import (
	"os"
	"testing"

	"github.com/kwilteam/kwil-db/pkg/engine/utils"
	"github.com/kwilteam/kwil-db/pkg/kuneiform/parser"
	"github.com/kwilteam/kwil-db/pkg/kuneiform/schema"
)

type DatabaseSchemaLoader interface {
	Load(t *testing.T, targetSchema *testSchema) *schema.Schema
	LoadWithoutValidation(t *testing.T, targetSchema *testSchema) *schema.Schema
}

type FileDatabaseSchemaLoader struct {
	Modifier func(db *schema.Schema)
}

func (l *FileDatabaseSchemaLoader) Load(t *testing.T, targetSchema *testSchema) *schema.Schema {
	t.Helper()

	d, err := os.ReadFile(targetSchema.GetFilePath())
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

func (l *FileDatabaseSchemaLoader) LoadWithoutValidation(t *testing.T, targetSchema *testSchema) *schema.Schema {
	t.Helper()

	d, err := os.ReadFile(targetSchema.GetFilePath())
	if err != nil {
		t.Fatal("cannot open database schema file", err)
	}

	ast, _ := parser.Parse(d)
	if ast == nil {
		t.Fatal("cannot parse database schema", err)
	}

	db := ast.Schema()

	l.Modifier(db)

	return db
}

func GenerateSchemaId(owner, name string) string {
	return utils.GenerateDBID(name, owner)
}
