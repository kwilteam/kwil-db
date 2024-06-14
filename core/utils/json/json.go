// package json includes JSON utilities commonly used in Kwil.
package json

import (
	"encoding/json"
	"reflect"
	"strings"
)

// UnmarshalMapWithoutFloat unmarshals a JSON byte slice into a slice of maps.
// It will try to convert all return values into ints, but will keep them as strings if it fails.
// It ensures they aren't returned as floats, which is important for maintaining consistency
// with Kwil's decimal types. All returned types will be string, int64, or a []any.
func UnmarshalMapWithoutFloat(b []byte) ([]map[string]any, error) {
	d := json.NewDecoder(strings.NewReader(string(b)))
	d.UseNumber()

	// unmashal result
	var result []map[string]any
	err := d.Decode(&result)
	if err != nil {
		return nil, err
	}

	// convert numbers to int64
	result = convertJsonNumbers(result).([]map[string]any)

	return result, nil
}

// convertJsonNumbers recursively converts json.Number to int64.
// It traverses through the map and array and converts all json.Number to int64.
func convertJsonNumbers(val any) any {
	if val == nil {
		return nil
	}
	switch val := val.(type) {
	case map[string]any:
		for k, v := range val {
			val[k] = convertJsonNumbers(v)
		}
		return val
	case []map[string]any:
		for i, v := range val {
			for j, n := range v {
				v[j] = convertJsonNumbers(n)
			}
			val[i] = v
		}
		return val
	case []any:
		for i, v := range val {
			val[i] = convertJsonNumbers(v)
		}
		return val
	case json.Number:
		i, err := val.Int64()
		if err != nil {
			return val.String()
		}
		return i
	case string:
		return val
	case int64:
		return val
	default:
		// in case we are unmarshalling something crazy like a double nested slice,
		// we reflect on the value and recursively call convertJsonNumbers if it's a slice.
		typeOf := reflect.TypeOf(val)
		if typeOf.Kind() == reflect.Slice {
			s := reflect.ValueOf(val)
			for i := 0; i < s.Len(); i++ {
				s.Index(i).Set(reflect.ValueOf(convertJsonNumbers(s.Index(i).Interface())))
			}
			return s.Interface()
		}
		return val
	}
}
