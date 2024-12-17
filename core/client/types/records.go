package client

import (
	"fmt"
)

// Records provides an export helper for a set of records.
type Records []map[string]any

// ToStrings converts the values in each map to strings.
func (r Records) ToStrings() []map[string]string {
	if r == nil {
		return nil
	}

	records := make([]map[string]string, len(r))
	for i, record := range r {
		records[i] = stringifyMap(record)
	}
	return records
}

// stringifyMap converts a Records instance into a map with the values converted to strings.
func stringifyMap(r map[string]any) map[string]string {
	rec := make(map[string]string, len(r))
	for k, v := range r {
		rec[k] = fmt.Sprintf("%v", v)
	}
	return rec
}
