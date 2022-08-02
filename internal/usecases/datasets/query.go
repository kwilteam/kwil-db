package datasets

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kwilteam/kwil-db/internal/entity"
)

func (u *DatasetUseCase) Query(ctx context.Context, query *entity.DBQuery) ([]byte, error) {
	db, err := u.engine.GetDataset(ctx, query.DBID)
	if err != nil {
		return nil, fmt.Errorf("dataset not found")
	}

	res, err := db.Query(ctx, query.Query, nil)
	if err != nil {
		return nil, fmt.Errorf("query error: %w", err)
	}

	bts, err := readQueryResult(res)
	if err != nil {
		return nil, fmt.Errorf("internal server error: error serializing query results: %w", err)
	}

	return bts, nil
}

func readQueryResult(res []map[string]any) ([]byte, error) {
	return json.Marshal(res)
}
