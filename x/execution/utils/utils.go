package utils

import (
	"encoding/json"
	"kwil/x/execution/dto"
)

func DBFromJson(bts []byte) (*dto.Database, error) {
	var db dto.Database
	err := json.Unmarshal(bts, &db)
	if err != nil {
		return nil, err
	}
	return &db, nil
}

func DBToJson(db *dto.Database) ([]byte, error) {
	return json.Marshal(db)
}
