package pdb

import (
	"ksl/syntax/nodes"

	"golang.org/x/exp/slices"
)

type Config struct {
	Backend *Backend
}

func (c *Config) BackendName() string {
	if c.Backend != nil {
		return c.Backend.Name
	}
	return ""
}

type Backend struct {
	Name   string
	Source DirectiveID
}

type IDAnnotation struct {
	Name        string
	MappedName  string
	Fields      []FieldRef
	SourceAnnot AnnotID
	SourceField *FieldID
}

type IndexType int

const (
	IndexTypeNormal IndexType = iota
	IndexTypeUnique
	IndexTypeFullText
)

type IndexAlgorithm string

const (
	BTree  IndexAlgorithm = "BTree"
	Hash   IndexAlgorithm = "Hash"
	Gist   IndexAlgorithm = "Gist"
	Gin    IndexAlgorithm = "Gin"
	SpGist IndexAlgorithm = "SpGist"
	Brin   IndexAlgorithm = "Brin"
)

func (a IndexAlgorithm) IsHash() bool { return a == Hash }

type IndexAnnotation struct {
	Type             IndexType
	Fields           []FieldRef
	SourceField      *FieldID
	Name             string
	Algorithm        IndexAlgorithm
	SourceAnnotation AnnotID
}

func (idx IndexAnnotation) IsUnique() bool { return idx.Type == IndexTypeUnique }
func (idx IndexAnnotation) HasFields(fields []FieldID) bool {
	if len(fields) != len(idx.Fields) {
		return false
	}

	slices.Sort(fields)
	idxFields := make([]FieldID, len(idx.Fields))
	for i, field := range idx.Fields {
		idxFields[i] = field.FieldID
	}
	slices.Sort(idxFields)
	return slices.Equal(fields, idxFields)
}

type ModelAnnotations struct {
	PrimaryKey *IDAnnotation
	Indexes    []*IndexAnnotation
	Ignored    bool
	Name       string
}

type EnumAnnotations struct {
	MappedName   string
	MappedValues map[EnumValueID]string
}

type FieldRef struct {
	ModelID ModelID
	FieldID FieldID
	Sort    string
}

type DefaultAnnotation struct {
	SourceAnnot AnnotID
	Value       nodes.Expression
}
