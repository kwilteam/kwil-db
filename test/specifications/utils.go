package specifications

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/kwilteam/kwil-db/pkg/engine/utils"
	"github.com/kwilteam/kwil-db/pkg/kuneiform/parser"
	"github.com/kwilteam/kwil-db/pkg/kuneiform/schema"
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

	jsonBytes, err := json.Marshal(db)
	if err != nil {
		t.Fatal("cannot marshal database schema", err)
	}

	os.WriteFile("./db.json", jsonBytes, 0644)

	l.Modifier(db)

	return db
}

func GenerateSchemaId(owner, name string) string {
	return utils.GenerateDBID(name, owner)
}
