package planner2

import (
	"fmt"

	"github.com/kwilteam/kwil-db/core/types"
)

type Relation interface {
	Schema(Catalog) *Schema
}

// Schema represents the schema of a relation.
type Schema struct {
	Fields []*Field
}

// FindField returns the field with the given name.
// If many fields have the same name and a relation is not specified, an error is returned.
func (s *Schema) FindField(relation, name string) (field *Field, found bool, err error) {
	if relation == "" {
		var field *Field
		count := 0
		for _, field := range s.Fields {
			if field.Name == name {
				field = field
				count++
			}
		}
		if count == 0 {
			return nil, false, fmt.Errorf(`column "%s" not found`, name)
		}
		if count > 1 {
			return nil, true, fmt.Errorf(`column "%s" is ambiguous`, name)
		}
		return field, true, nil
	}

	for _, field := range s.Fields {
		if field.ParentRelation == relation && field.Name == name {
			return field, true, nil
		}
	}

	return nil, false, fmt.Errorf(`column "%s" not found in table "%s"`, name, relation)
}

func (s *Schema) RowTypes() []*types.DataType {
	var types []*types.DataType
	for _, field := range s.Fields {
		types = append(types, field.Type)
	}
	return types
}

func mergeSchemas(schemas ...*Schema) *Schema {
	var fields []*Field
	for _, schema := range schemas {
		fields = append(fields, schema.Fields...)
	}
	return &Schema{Fields: fields}
}

// Field represents a field (column) in a schema.
type Field struct {
	// ParentRelation is the name of the relation this field belongs to.
	// It can be empty if the field is unqualified.
	ParentRelation string
	// Name is the name of the field.
	// It can be empty if the field is unqualified.
	Name     string
	Type     *types.DataType
	Nullable bool
	HasIndex bool
	Unique   bool
}
