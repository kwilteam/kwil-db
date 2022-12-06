package schema

import (
	"gopkg.in/yaml.v2"
)

func MarshalDatabase(bts []byte) (*Database, error) {
	var db Database
	err := yaml.Unmarshal(bts, &db)
	if err != nil {
		return nil, err
	}
	return &db, nil
}
