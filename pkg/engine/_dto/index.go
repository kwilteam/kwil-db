package dto

import (
	"fmt"
	"strings"
)

type IndexType string

type Index struct {
	Name    string    `json:"name" clean:"lower"`
	Columns []string  `json:"columns" clean:"lower"`
	Type    IndexType `json:"type" clean:"is_enum,index_type"`
}

func (i *Index) Clean() error {
	return runCleans(
		cleanIdent(&i.Name),
		cleanIdents(&i.Columns),
		i.Type.Clean(),
	)
}

const (
	BTREE        IndexType = "BTREE"
	UNIQUE_BTREE IndexType = "UNIQUE_BTREE"
	PRIMARY      IndexType = "PRIMARY"
)

func (i IndexType) String() string {
	return string(i)
}

func (i *IndexType) IsValid() bool {
	upper := strings.ToUpper(i.String())

	return upper == BTREE.String() ||
		upper == UNIQUE_BTREE.String() ||
		upper == PRIMARY.String()
}

func (i *IndexType) Clean() error {
	if !i.IsValid() {
		return fmt.Errorf("invalid index type: %s", i.String())
	}

	*i = IndexType(strings.ToUpper(i.String()))

	return nil
}
