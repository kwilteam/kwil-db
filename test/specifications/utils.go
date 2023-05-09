package specifications

import (
	"github.com/kwilteam/kwil-db/pkg/engine/models"
	"github.com/kwilteam/kwil-db/pkg/kuneiform/parser"
	"os"
	"testing"
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

	ast, err := parser.Parse(d)
	if err != nil {
		t.Fatal("cannot parse database schema", err)
	}

	dataset := ast.Dataset()

	l.Modifier(dataset)

	return dataset
}
