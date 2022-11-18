package dml

type ReferentialAction int

const (
	NoAction ReferentialAction = iota
	Restrict
	Cascade
	SetNull
	SetDefault
)

type FieldArity int

const (
	Required FieldArity = iota
	Optional
	List
)

type IndexType int

const (
	Normal IndexType = iota
	Unique
)

type IndexAlgorithm int

const (
	BTree IndexAlgorithm = iota
	Hash
	Gist
	Gin
	SpGist
	Brin
)

type SortOrder int

const (
	Ascending SortOrder = iota
	Descending
)
