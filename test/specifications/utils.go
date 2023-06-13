package specifications

import (
	"os"
	"testing"

	"github.com/kwilteam/kuneiform/kfparser"
	"github.com/kwilteam/kuneiform/schema"
	"github.com/kwilteam/kwil-db/pkg/engine/utils"
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

	db, err := kfparser.Parse(string(d))
	if err != nil {
		t.Fatal("cannot parse database schema", err)
	}

	l.Modifier(db)
	return db
}

func (l *FileDatabaseSchemaLoader) LoadWithoutValidation(t *testing.T, targetSchema *testSchema) *schema.Schema {
	t.Helper()

	d, err := os.ReadFile(targetSchema.GetFilePath())
	if err != nil {
		t.Fatal("cannot open database schema file", err)
	}

	db, _ := kfparser.Parse(string(d))
	if db == nil {
		t.Fatal("cannot parse database schema", err)
	}

	l.Modifier(db)

	return db
}

func GenerateSchemaId(owner, name string) string {
	return utils.GenerateDBID(name, owner)
}
