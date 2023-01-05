package serialize

import "encoding/json"

func Serialize(v any) ([]byte, error) {
	return json.Marshal(v)
}

func Deserialize[T any](bts []byte) (T, error) {
	var v T
	err := json.Unmarshal(bts, &v)
	if err != nil {
		return v, err
	}
	return v, nil
}
