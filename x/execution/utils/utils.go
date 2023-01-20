package utils

import (
	"encoding/json"
	anytype "kwil/x/types/data_types/any_type"
	"kwil/x/types/databases"
)

// TODO: delete

func DBFromJson(bts []byte) (*databases.Database[anytype.KwilAny], error) {
	var db databases.Database[anytype.KwilAny]
	err := json.Unmarshal(bts, &db)
	if err != nil {
		return nil, err
	}
	return &db, nil
}

func DBToJson(db *databases.Database[anytype.KwilAny]) ([]byte, error) {
	return json.Marshal(db)
}
