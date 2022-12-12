package types

import "fmt"

type IndexType int

const (
	INVALID_INDEX IndexType = iota
	BTREE
)

func (i *IndexType) String() string {
	switch *i {
	case BTREE:
		return "btree"
	}
	return "unknown"
}

// ConvertIndex converts a string to an IndexType
func (c *conversion) ConvertIndex(s string) (IndexType, error) {
	switch s {
	case "btree":
		return BTREE, nil
	}
	return INVALID_INDEX, fmt.Errorf("unknown index type")
}
