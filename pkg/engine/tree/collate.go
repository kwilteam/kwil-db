package tree

type CollationType string

const (
	CollationTypeNone   CollationType = ""
	CollationTypeBinary CollationType = "BINARY"
	CollationTypeNoCase CollationType = "NOCASE"
	CollationTypeRTrim  CollationType = "RTRIM"
)

func (c CollationType) String() string {
	return string(c)
}
