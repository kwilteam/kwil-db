// package json includes JSON utilities commonly used in Kwil.
package json

import (
	"encoding/json"
	"strings"
)

// UnmarshalMapWithoutFloat unmarshals a JSON byte slice into a slice of maps.
// It will try to convert all return values into ints, but will keep them as strings if it fails.
// It ensures they aren't returned as floats, which is important for maintaining consistency
// with Kwil's decimal types. All returned types will be string or int64.
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
	for _, record := range result {
		for k, v := range record {
			if num, ok := v.(json.Number); ok {
				i, err := num.Int64()
				if err != nil {
					record[k] = num.String()
				} else {
					record[k] = i
				}
			} else if num, ok := v.([]any); ok {
				for j, n := range num {
					if n, ok := n.(json.Number); ok {
						i, err := n.Int64()
						if err != nil {
							num[j] = n.String()
						} else {
							num[j] = i
						}
					}
				}
			}
		}
	}

	return result, nil
}
