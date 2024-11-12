package order

import (
	"cmp"
	"sort"
)

// OrderMap orders a map lexicographically by its keys.
// It permits any map with keys that are generically orderable.
func OrderMap[O cmp.Ordered, T any](m map[O]T) []*KVPair[O, T] {
	keys := make([]O, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})

	result := make([]*KVPair[O, T], 0, len(m))

	for _, k := range keys {
		result = append(result, &KVPair[O, T]{
			Key:   k,
			Value: m[k],
		})
	}

	return result
}

// KVPair is a pair of key and value for a map.
// The key must be orderable.
type KVPair[O cmp.Ordered, T any] struct {
	Key   O
	Value T
}

// ToMap converts a slice of flattened pairs to a map.
func ToMap[O cmp.Ordered, T any](pairs []*KVPair[O, T]) map[O]T {
	result := make(map[O]T, len(pairs))

	for _, pair := range pairs {
		result[pair.Key] = pair.Value
	}

	return result
}
