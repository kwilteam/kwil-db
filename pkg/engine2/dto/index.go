package dto

type IndexType string

type Index struct {
	Name    string    `json:"name" clean:"lower"`
	Columns []string  `json:"columns" clean:"lower"`
	Type    IndexType `json:"type" clean:"is_enum,index_type"`
}

const (
	BTREE        IndexType = "BTREE"
	UNIQUE_BTREE IndexType = "UNIQUE_BTREE"
)

func (i *IndexType) String() string {
	return string(*i)
}

func (i *IndexType) IsValid() bool {
	return *i == BTREE || *i == UNIQUE_BTREE
}
