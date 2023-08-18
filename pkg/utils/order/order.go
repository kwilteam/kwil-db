package order

import (
	"sort"

	"golang.org/x/exp/constraints"
)

// OrderMapLexicographically orders a map lexicographically by its keys.
// It permits any map with keys that are generically orderable.
// TODO: once upgraded to go 1.21, I believe 'Ordered' is in the standard library
func OrderMapLexicographically[S constraints.Ordered, T any](m map[S]T) []*struct {
	Id    S
	Value T
} {
	keys := make([]S, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})

	result := make([]*struct {
		Id    S
		Value T
	}, 0, len(m))

	for _, k := range keys {
		result = append(result, &struct {
			Id    S
			Value T
		}{
			Id:    k,
			Value: m[k],
		})
	}

	return result
}
