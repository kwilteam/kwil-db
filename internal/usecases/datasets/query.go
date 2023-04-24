package datasets

import (
	"fmt"
	"kwil/internal/entity"
)

func (u *DatasetUseCase) Query(query *entity.DBQuery) ([]byte, error) {
	db, err := u.engine.GetDataset(query.DBID)
	if err != nil {
		return nil, fmt.Errorf("dataset not found")
	}

	res, err := db.Query(query.Query)
	if err != nil {
		return nil, fmt.Errorf("query error: %w", err)
	}

	bts, err := res.Bytes()
	if err != nil {
		return nil, fmt.Errorf("internal server error: error serializing query results: %w", err)
	}

	return bts, nil
}
