package datasets

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kwilteam/kwil-db/internal/entity"
	"github.com/kwilteam/kwil-db/pkg/engine2/dto"
)

func (u *DatasetUseCase) Query(ctx context.Context, query *entity.DBQuery) ([]byte, error) {
	db, err := u.engine.GetDataset(query.DBID)
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

func readQueryResult(res dto.Result) ([]byte, error) {
	records := res.Records()

	return json.Marshal(records)
}
