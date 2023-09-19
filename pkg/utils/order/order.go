package order

import (
	"cmp"
	"sort"
)

// OrderMap orders a map lexicographically by its keys.
// It permits any map with keys that are generically orderable.
func OrderMap[O cmp.Ordered, T any](m map[O]T) []*FlattenedPair[O, T] {
	keys := make([]O, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})

	result := make([]*FlattenedPair[O, T], 0, len(m))

	for _, k := range keys {
		result = append(result, &FlattenedPair[O, T]{
			Key:   k,
			Value: m[k],
		})
	}

	return result
}

// FlattenPair is a pair of key and value.
type FlattenedPair[O cmp.Ordered, T any] struct {
	Key   O
	Value T
}

// ToMap converts a slice of flattened pairs to a map.
func ToMap[O cmp.Ordered, T any](pairs []*FlattenedPair[O, T]) map[O]T {
	result := make(map[O]T, len(pairs))

	for _, pair := range pairs {
		result[pair.Key] = pair.Value
	}

	return result
}
