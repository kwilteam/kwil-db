package serialize

// the serialize package is used to serializing and deserializing database metadata, with the goal of being able to
// support multiple versions of different metadata structs (i.e., table changes) in the future

type Serializable struct {
	Name    string
	Type    TypeIdentifier
	Version int64
	Data    []byte
}

const (
	tableVersion     = 1
	actionVersion    = 1
	extensionVersion = 1
)

type serializer interface {
	Serialize() ([]byte, error)
}

type TypeIdentifier string

const (
	IdentifierTable     TypeIdentifier = "table"
	IdentifierAction    TypeIdentifier = "action"
	IdentifierExtension TypeIdentifier = "extension"
)

func (t TypeIdentifier) String() string {
	return string(t)
}
