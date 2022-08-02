package tree

type CollationType string

const (
	CollationTypeBinary CollationType = "BINARY"
	CollationTypeNoCase CollationType = "NOCASE"
	CollationTypeRTrim  CollationType = "RTRIM"
)

func (c CollationType) String() string {
	c.check()
	return string(c)
}

func (c CollationType) check() {
	if !c.Valid() {
		panic("invalid collation type")
	}
}

func (c CollationType) Valid() bool {
	return c == CollationTypeBinary || c == CollationTypeNoCase || c == CollationTypeRTrim
}
