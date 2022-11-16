package pdb

import (
	"fmt"
	"ksl/syntax/ast"
)

type Ast struct {
	Files map[string]*ast.File
	Tops  []ast.TopLevel
}

func NewAst(files ...*ast.File) *Ast {
	fileset := make(map[string]*ast.File)
	tops := make([]ast.TopLevel, 0)

	for _, file := range files {
		fileset[file.Name] = file
	}
	for _, f := range files {
		tops = append(tops, f.Entries...)
	}

	return &Ast{Files: fileset, Tops: tops}
}

func (s *Ast) Sources() map[string][]byte {
	sources := make(map[string][]byte)
	for _, file := range s.Files {
		sources[file.Name] = file.Contents
	}
	return sources
}

func (s *Ast) AddFile(file *ast.File) error {
	if _, ok := s.Files[file.Name]; ok {
		return fmt.Errorf("duplicate file %q", file.Name)
	}

	s.Files[file.Name] = file
	s.Tops = append(s.Tops, file.Entries...)
	return nil
}

func (s Ast) Entries() []ast.TopLevel       { return s.Tops[:] }
func (s Ast) GetEntry(top int) ast.TopLevel { return s.Tops[top] }

func (s Ast) GetNode(id NodeID) ast.Node {
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
		node := s.GetNode(id.node).(ast.Annotated)
		annots := node.GetAnnotations()
		return annots[id.annot]
	default:
		panic("unreachable")
	}
}

func (s Ast) GetDirective(id DirectiveID) *ast.Annotation { return s.Tops[id].(*ast.Annotation) }
func (s Ast) GetModel(id ModelID) *ast.Model              { return s.Tops[id].(*ast.Model) }
func (s Ast) GetBlock(id BlockID) *ast.Block              { return s.Tops[id].(*ast.Block) }
func (s Ast) GetEnum(id EnumID) *ast.Enum                 { return s.Tops[id].(*ast.Enum) }

func (s Ast) FindModelField(modelID ModelID, name string) (ModelFieldID, bool) {
	model := s.GetModel(modelID)
	for i, field := range model.Fields {
		if field.GetName() == name {
			return MakeModelFieldID(modelID, FieldID(i)), true
		}
	}
	return ModelFieldID{}, false
}

func (s Ast) GetModelField(id ModelFieldID) *ast.Field {
	return s.GetModel(id.model).Fields[id.field]
}

func (s Ast) GetEnumValue(id EnumValueID) *ast.EnumValue {
	return s.GetEnum(id.enum).Values[id.value]
}

func (s Ast) GetAnnotation(id AnnotID) *ast.Annotation {
	node := s.GetNode(id.node).(ast.Annotated)
	annots := node.GetAnnotations()
	return annots[id.annot]
}
