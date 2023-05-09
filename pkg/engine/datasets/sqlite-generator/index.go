package sqlitegenerator

import (
	"fmt"
	"github.com/kwilteam/kwil-db/pkg/engine/models"
	"github.com/kwilteam/kwil-db/pkg/engine/types"
	"strings"
)

func indexTypeToSQLiteString(indexType types.IndexType) (string, error) {
	switch indexType {
	case types.BTREE:
		return "", nil
	case types.UNIQUE_BTREE:
		return " UNIQUE", nil
	default:
		return "", fmt.Errorf("unknown index type: %d", indexType)
	}
}

func GenerateCreateIndexStatements(tableName string, indexes []*models.Index) ([]string, error) {
	var statements []string

	for _, index := range indexes {
		indexType, err := indexTypeToSQLiteString(index.Type)
		if err != nil {
			return nil, err
		}
		columns := strings.Join(index.Columns, ", ")

		statement := fmt.Sprintf("CREATE%s INDEX %s ON %s (%s);", indexType, index.Name, tableName, columns)
		statements = append(statements, strings.TrimSpace(statement))
	}

	return statements, nil
}
