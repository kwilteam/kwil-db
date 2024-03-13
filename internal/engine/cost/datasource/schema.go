package datasource

import (
	"fmt"
	"strings"
)

// Field represents a field in a schema.
type Field struct {
	Name string
	Type string
}

type Schema struct {
	Fields []Field
}

func (s *Schema) String() string {
	var fields []string
	for _, f := range s.Fields {
		fields = append(fields, fmt.Sprintf("%s/%s", f.Name, f.Type))
	}
	return fmt.Sprintf("[%s]", strings.Join(fields, ", "))
}

func (s *Schema) Select(projection ...string) *Schema {
	fieldIndex := s.mapProjection(projection)

	newFields := make([]Field, len(projection))
	for i, idx := range fieldIndex {
		newFields[i] = s.Fields[idx]
	}

	return NewSchema(newFields...)
}

// mapProjection maps the projection to the index of the fields in the schema.
func (s *Schema) mapProjection(projection []string) []int {
	fieldIndexMap := make(map[string]int)
	for i, field := range s.Fields {
		fieldIndexMap[field.Name] = i
	}

	newFieldsIndex := make([]int, len(projection))
	for i, name := range projection {
		newFieldsIndex[i] = fieldIndexMap[name]
	}

	return newFieldsIndex
}

func (s *Schema) Join(other *Schema) *Schema {
	fields := make([]Field, len(s.Fields)+len(other.Fields))
	copy(fields, s.Fields)
	copy(fields[len(s.Fields):], other.Fields)
	return NewSchema(fields...)
}

func NewSchema(fields ...Field) *Schema {
	return &Schema{Fields: fields}
}
