package specifications

import (
	"kwil/pkg/engine/models"
	"kwil/test/acceptance/mocks"
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

	db := mocks.MOCK_DATASET1

	l.Modifier(&db)

	return &db
}

type CliDatabaseSchemaLoader struct {
	Flag     string
	Modifier func(db *models.Dataset)
}

func (l *CliDatabaseSchemaLoader) Load(t *testing.T) *models.Dataset {
	t.Helper()

	db := mocks.MOCK_DATASET1

	l.Modifier(&db)
	return &db
}
