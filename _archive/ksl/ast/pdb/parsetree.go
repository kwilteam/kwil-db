package pdb

import (
	"fmt"
	"ksl/syntax/nodes"
)

type ParseTree struct {
	Files map[string]*nodes.File
	Tops  []nodes.TopLevel
}

func NewParseTree(files ...*nodes.File) *ParseTree {
	fileset := make(map[string]*nodes.File)
	tops := make([]nodes.TopLevel, 0)

	for _, file := range files {
		fileset[file.Name] = file
	}
	for _, f := range files {
		tops = append(tops, f.Entries...)
	}

	return &ParseTree{Files: fileset, Tops: tops}
}

func (s *ParseTree) Sources() map[string][]byte {
	sources := make(map[string][]byte)
	for _, file := range s.Files {
		sources[file.Name] = file.Contents
	}
	return sources
}

func (s *ParseTree) AddFile(file *nodes.File) error {
	if _, ok := s.Files[file.Name]; ok {
		return fmt.Errorf("duplicate file %q", file.Name)
	}

	s.Files[file.Name] = file
	s.Tops = append(s.Tops, file.Entries...)
	return nil
}

func (s ParseTree) Entries() []nodes.TopLevel       { return s.Tops[:] }
func (s ParseTree) GetEntry(top int) nodes.TopLevel { return s.Tops[top] }

func (s ParseTree) GetNode(id NodeID) nodes.Node {
	switch id := id.(type) {
	case ModelID:
		return s.GetModel(id)
	case BlockID:
		return s.GetBlock(id)
	case EnumID:
		return s.GetEnum(id)
	case ModelFieldID:
		return s.GetModelField(id)
	case EnumValueID:
		return s.GetEnumValue(id)
	case AnnotID:
		node := s.GetNode(id.node).(nodes.Annotated)
		annots := node.GetAnnotations()
		return annots[id.annot]
	default:
		panic("unreachable")
	}
}

func (s ParseTree) GetDirective(id DirectiveID) *nodes.Annotation {
	return s.Tops[id].(*nodes.Annotation)
}
func (s ParseTree) GetModel(id ModelID) *nodes.Model { return s.Tops[id].(*nodes.Model) }
func (s ParseTree) GetBlock(id BlockID) *nodes.Block { return s.Tops[id].(*nodes.Block) }
func (s ParseTree) GetEnum(id EnumID) *nodes.Enum    { return s.Tops[id].(*nodes.Enum) }

func (s ParseTree) FindModelField(modelID ModelID, name string) (ModelFieldID, bool) {
	model := s.GetModel(modelID)
	for i, field := range model.Fields {
		if field.GetName() == name {
			return MakeModelFieldID(modelID, FieldID(i)), true
		}
	}
	return ModelFieldID{}, false
}

func (s ParseTree) GetModelField(id ModelFieldID) *nodes.Field {
	return s.GetModel(id.model).Fields[id.field]
}

func (s ParseTree) GetEnumValue(id EnumValueID) *nodes.EnumValue {
	return s.GetEnum(id.enum).Values[id.value]
}

func (s ParseTree) GetAnnotation(id AnnotID) *nodes.Annotation {
	node := s.GetNode(id.node).(nodes.Annotated)
	annots := node.GetAnnotations()
	return annots[id.annot]
}
