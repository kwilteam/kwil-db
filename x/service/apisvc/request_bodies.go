package apisvc

import "encoding/json"

type RequestBody interface {
	DropDatabaseBody | ExecuteTxBody
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

type DropDatabaseBody struct {
	Database string `json:"database"`
}

type ExecuteTxBody struct {
}
