package ddl

import (
	"fmt"
	"strings"

	"github.com/kwilteam/kwil-db/internal/engine/types"
)

func indexTypeToSQLiteString(indexType types.IndexType) (string, error) {
	err := indexType.Clean()
	if err != nil {
		return "", err
	}

	switch indexType {
	case types.BTREE:
		return "", nil
	case types.UNIQUE_BTREE:
		return " UNIQUE", nil
	case types.PRIMARY:
		return " PRIMARY KEY", nil
	default:
		return "", fmt.Errorf("unknown index type: %s", indexType)
	}
}

func GenerateCreateIndexStatements(pgSchema, tableName string, indexes []*types.Index) ([]string, error) {
	var statements []string

	for _, index := range indexes {
		indexType, err := indexTypeToSQLiteString(index.Type)
		if err != nil {
			return nil, err
		}

		// Skip primary indexes, as they are created with the table
		if strings.EqualFold(index.Type.String(), types.PRIMARY.String()) {
			continue
		}

		cols := make([]string, len(index.Columns))
		for i, col := range index.Columns {
			cols[i] = wrapIdent(col)
		}
		columns := strings.Join(cols, ", ")

		statement := fmt.Sprintf("CREATE%s INDEX %s ON %s.%s (%s);", indexType, wrapIdent(index.Name),
			wrapIdent(pgSchema), wrapIdent(tableName), columns)
		statements = append(statements, strings.TrimSpace(statement))
	}

	return statements, nil
}
