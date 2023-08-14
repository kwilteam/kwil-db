package specifications

import (
	"encoding/json"
	"github.com/kwilteam/kwil-db/pkg/transactions"
	"os"
	"testing"

	"github.com/kwilteam/kuneiform/kfparser"
	"github.com/kwilteam/kwil-db/pkg/engine/utils"
)

type DatabaseSchemaLoader interface {
	Load(t *testing.T, targetSchema *testSchema) *transactions.Schema
	LoadWithoutValidation(t *testing.T, targetSchema *testSchema) *transactions.Schema
}

type FileDatabaseSchemaLoader struct {
	Modifier func(db *transactions.Schema)
}

func (l *FileDatabaseSchemaLoader) Load(t *testing.T, targetSchema *testSchema) *transactions.Schema {
	t.Helper()

	d, err := os.ReadFile(targetSchema.GetFilePath())
	if err != nil {
		t.Fatal("cannot open database schema file", err)
	}

	astSchema, err := kfparser.Parse(string(d))
	if err != nil {
		t.Fatal("cannot parse database schema", err)
	}

	schemaJson, err := json.Marshal(astSchema)
	if err != nil {
		t.Fatal("failed to marshal schema: %w", err)
	}

	var db *transactions.Schema
	err = json.Unmarshal(schemaJson, &db)
	if err != nil {
		t.Fatal("failed to unmarshal schema json: %w", err)
	}

	l.Modifier(db)
	return db
}

func (l *FileDatabaseSchemaLoader) LoadWithoutValidation(t *testing.T, targetSchema *testSchema) *transactions.Schema {
	t.Helper()

	d, err := os.ReadFile(targetSchema.GetFilePath())
	if err != nil {
		t.Fatal("cannot open database schema file", err)
	}

	astSchema, err := kfparser.Parse(string(d))
	// ignore validation error
	if astSchema == nil {
		t.Fatal("cannot parse database schema", err)
	}

	schemaJson, err := json.Marshal(astSchema)
	if err != nil {
		t.Fatal("failed to marshal schema: %w", err)
	}

	var db *transactions.Schema
	err = json.Unmarshal(schemaJson, &db)
	if err != nil {
		t.Fatal("failed to unmarshal schema json: %w", err)
	}

	l.Modifier(db)

	return db
}

func GenerateSchemaId(owner, name string) string {
	return utils.GenerateDBID(name, owner)
}
