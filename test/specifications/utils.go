package specifications

import (
	"encoding/json"
	"kwil/pkg/databases"
	"kwil/pkg/engine/models"
	"kwil/pkg/kl/parser"
	"os"
	"testing"

	"github.com/spf13/viper"
)

type DatabaseSchemaLoader interface {
	Load(t *testing.T) *models.Dataset
}

type FileDatabaseSchemaLoader struct {
	FilePath string
	Modifier func(db *models.Dataset)
}

func (l *FileDatabaseSchemaLoader) Load(t *testing.T) *models.Dataset {
	t.Helper()

	d, err := os.ReadFile(l.FilePath)
	//d, err := os.ReadFile("./data/database_schema.json")
	if err != nil {
		t.Fatal("cannot open database schema file", err)
	}

	// TODO: parse kl
	ast, err := parser.Parse(d)
	if err != nil {
		t.Fatal("cannot parse database schema", err)
	}

	dataset := ast.Dataset()

	l.Modifier(dataset)

	return dataset
}

type CliDatabaseSchemaLoader struct {
	Flag     string
	Modifier func(db *databases.Database[[]byte])
}

func (l *CliDatabaseSchemaLoader) Load(t *testing.T) *databases.Database[[]byte] {
	t.Helper()

	d, err := os.ReadFile(viper.GetString("path"))
	//d, err := os.ReadFile("./data/database_schema.json")
	if err != nil {
		t.Fatal("cannot open database schema file", err)
	}

	var db databases.Database[[]byte]
	err = json.Unmarshal(d, &db)
	if err != nil {
		t.Fatal("cannot parse database schema", err)
	}

	l.Modifier(&db)
	return &db
}
