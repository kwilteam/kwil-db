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

// convert wil serialize and deserialize values into different types
// they should only be passed concrete types, not pointers
func Convert[T1 any, T2 any](v *T1) (*T2, error) {
	bts, err := Serialize(v)
	if err != nil {
		return nil, err
	}

	v2, err := Deserialize[T2](bts)
	if err != nil {
		return nil, err
	}

	return &v2, nil
}
