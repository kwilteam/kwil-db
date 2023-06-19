package serialize

import (
	"encoding/json"
	"reflect"
)

type Marshaler[T interface{}] struct {
	KnownImplementations []T
}

func (m Marshaler[T]) Marshal(v T) ([]byte, error) {
	return json.Marshal(container[T]{
		Value:                v,
		KnownImplementations: m.KnownImplementations,
	})
}

func (m Marshaler[T]) Unmarshal(bytes []byte) (T, error) {

	var v T
	cont := container[T]{
		Value:                v,
		KnownImplementations: m.KnownImplementations,
	}

	if err := json.Unmarshal(bytes, &cont); err != nil {
		return v, err
	}

	return cont.Value, nil
}

type container[T interface{}] struct {
	Value                T `json:"value"`
	KnownImplementations []T
}

// exporting this since it gets called recursively
func (c *container[T]) UnmarshalJSON(bytes []byte) error {
	var data struct {
		Type  string
		Value json.RawMessage
	}
	if err := json.Unmarshal(bytes, &data); err != nil {
		return err
	}

	for _, knownImplementation := range c.KnownImplementations {
		knownType := reflect.TypeOf(knownImplementation)
		if knownType.String() == data.Type {
			// Create a new pointer to a value of the concrete message type
			target := reflect.New(knownType)

			if err := c.UnmarshalJSON(data.Value); err != nil {
				return err
			}

			// Unmarshal the data to an interface to the concrete value (which will act as a pointer, don't ask why)
			if err := json.Unmarshal(data.Value, target.Interface()); err != nil {
				return err
			}
			// Now we get the element value of the target and convert it to the interface type (this is to get rid of a pointer type instead of a plain struct value)
			c.Value = target.Elem().Interface().(T)
			return nil
		}
	}

	return nil
}

func (c container[T]) MarshalJSON() ([]byte, error) {
	// Marshal to type and actual data to handle unmarshaling to specific interface type
	return json.Marshal(struct {
		Type  string
		Value any
	}{
		Type:  reflect.TypeOf(c.Value).String(),
		Value: c.Value,
	})
}
