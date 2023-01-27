package execution

type IndexType int

const (
	INVALID_INDEX_TYPE IndexType = iota + 100
	BTREE
	END_INDEX_TYPE
)

func (i *IndexType) String() string {
	switch *i {
	case BTREE:
		return "btree"
	}
	return "unknown"
}

func (i *IndexType) Int() int {
	return int(*i)
}

func (i *IndexType) IsValid() bool {
	return *i > INVALID_INDEX_TYPE && *i < END_INDEX_TYPE
}
