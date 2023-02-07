package execution

import "fmt"

type indexes struct{}

var Indexes = &indexes{}

// ConvertIndex converts a string to an IndexType
func (c *indexes) ConvertIndex(s string) (IndexType, error) {
	switch s {
	case "btree":
		return BTREE, nil
	}
	return INVALID_INDEX_TYPE, fmt.Errorf("unknown index type")
}
