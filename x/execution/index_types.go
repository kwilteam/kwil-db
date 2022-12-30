package execution

type IndexType int

const (
	INVALID_INDEX IndexType = iota
	BTREE
	END_INDEX
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
	return *i > INVALID_INDEX && *i < END_INDEX
}
