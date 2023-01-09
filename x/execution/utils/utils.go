package utils

import (
	"encoding/json"
	"kwil/x/types/databases"
)

func DBFromJson(bts []byte) (*databases.Database, error) {
	var db databases.Database
	err := json.Unmarshal(bts, &db)
	if err != nil {
		return nil, err
	}
	return &db, nil
}

func DBToJson(db *databases.Database) ([]byte, error) {
	return json.Marshal(db)
}
