package apisvc

import (
	"encoding/json"
)

type RequestBody interface {
	DropDatabaseBody | ExecuteTxBody | CreateDatabaseBody
}

func Marshal[B RequestBody](v B) ([]byte, error) {
	return json.Marshal(v)
}

func Unmarshal[B RequestBody](data []byte) (*B, error) {
	out := new(B)
	if err := json.Unmarshal(data, out); err != nil {
		return nil, err
	}
	return out, nil
}

type CreateDatabaseBody struct {
	Database []byte `json:"database" yaml:"database"`
}

type DropDatabaseBody struct {
	Database string `json:"database" yaml:"database"`
}

type ExecuteTxBody struct {
}
