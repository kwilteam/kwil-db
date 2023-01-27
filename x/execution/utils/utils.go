package utils

import (
	"encoding/json"
	"kwil/x/types/databases"
)

// TODO: delete

func DBFromJson(bts []byte) (*databases.Database[[]byte], error) {
	var db databases.Database[[]byte]
	err := json.Unmarshal(bts, &db)
	if err != nil {
		return nil, err
	}
	return &db, nil
}

func DBToJson(db *databases.Database[[]byte]) ([]byte, error) {
	return json.Marshal(db)
}
