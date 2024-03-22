package datatypes

import (
	"fmt"
	"slices"
	"strings"
)

type TableRef struct {
	// NOTE: Does the comment clearly explain the purpose of the struct?
	//
	// In Kwil, the database name user specifies in Kuneiform is mapped to
	// underlying database/schema(dbid) in Postgres.
	// So the schema is transparent to the user, to avoid confusion (the term
	// schema is used mostly for the schema of the table, in the context of
	// cost model.), we use the term db.
	DB    string // e.g. schema in Postgres. All schemas are in same database.
	Table string
}

func (t *TableRef) String() string {
	if t.DB != "" {
		return fmt.Sprintf("%s.%s", t.DB, t.Table)
	}
	return t.Table
}

// Resolve resolves the table reference to a fully qualified table name.
func (t *TableRef) Resolve(defaultDB string) string {
	db := t.DB
	if db == "" {
		db = defaultDB
	}
	return fmt.Sprintf("%s.%s", db, t.Table)
}

func TableRefUnqualified(table string) *TableRef {
	return &TableRef{Table: table}
}

func TableRefQualified(db, table string) *TableRef {
	return &TableRef{DB: db, Table: table}
}

// Match checks if the given table reference matches the current table reference.
// Not set fields are ignored, meaning it's optimistic to assume equal.
func (t *TableRef) Match(other *TableRef) bool {
	if t.DB != "" {
		return t.DB == other.DB && t.Table == other.Table
	} else {
		return t.Table == other.Table
	}
}

// OfRelation is an interface that represents an object that is part of a relation.
type OfRelation interface {
	Relation() *TableRef
}

//type ofRelationBase struct {
//	Relation *TableRef
//}
//
//func (o *ofRelationBase) Relation() *TableRef {
//	return o.Relation
//}

// Field represents a field in a schema.
type Field struct {
	// ofRelationBase is used to implement the OfRelation interface.
	//ofRelationBase
	relation *TableRef

	Name string
	Type string
}

func NewField(name, typ string) Field {
	return Field{Name: name, Type: typ}
}

func NewFieldWithRelation(name, typ string, relation *TableRef) Field {
	return Field{Name: name, Type: typ, relation: relation}
}

func (f *Field) Relation() *TableRef {
	return f.relation
}

func (f *Field) QualifiedColumn() *ColumnDef {
	return Column(f.relation, f.Name)
}

type Schema struct {
	Fields []Field
}

func NewSchema(fields ...Field) *Schema {
	return &Schema{Fields: fields}
}

func NewSchemaQualified(relation *TableRef, fields ...Field) *Schema {
	for i := range fields {
		fields[i].relation = relation
	}
	return &Schema{Fields: fields}
}

func (s *Schema) String() string {
	var fields []string
	for _, f := range s.Fields {
		fields = append(fields, fmt.Sprintf("%s/%s", f.Name, f.Type))
	}
	return fmt.Sprintf("[%s]", strings.Join(fields, ", "))
}

func (s *Schema) Select(projection ...string) *Schema {
	if len(projection) == 0 {
		return NewSchema(s.Fields...)
	}

	fieldIndex := s.MapProjection(projection)

	newFields := make([]Field, len(projection))
	for i, idx := range fieldIndex {
		newFields[i] = s.Fields[idx]
	}

	return NewSchema(newFields...)
}

func (s *Schema) Field(i int) Field {
	return s.Fields[i]
}

// MapProjection maps the projection to the index of the fields in the schema.
// NOTE: originally it's not exported, should come back to this later.
func (s *Schema) MapProjection(projection []string) []int {
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

// Join creates a new schema by joining the fields of the current schema with
// the fields of another schema.
// NOTE: should do this on clone of the schema.
func (s *Schema) Join(other *Schema) *Schema {
	fields := make([]Field, len(s.Fields)+len(other.Fields))
	copy(fields, s.Fields)
	copy(fields[len(s.Fields):], other.Fields)
	return NewSchema(fields...)
}

func (s *Schema) indexOfField(relation *TableRef, name string) int {
	for i, f := range s.Fields {
		if relation != nil { // the field to look for is qualified
			if f.Relation() != nil { // current field is qualified
				if f.Relation().Match(relation) && f.Name == name {
					return i
				}
			}
			//else { // current field is unqualified
			//
			//}
		} else { // the field to look for is unqualified
			if f.Name == name {
				return i
			}
		}
	}
	return -1
}

func (s *Schema) fieldByQualifiedName(relation *TableRef, name string) Field {
	idx := s.indexOfField(relation, name)
	if idx == -1 {
		panic(fmt.Sprintf("field %s.%s not found", relation.Table, name))
		//return nil
	}
	return s.Fields[idx]
}

func (s *Schema) fieldByUnqualifiedName(name string) Field {
	var found []Field
	for _, f := range s.Fields {
		if f.Name == name {
			found = append(found, f)
		}
	}

	switch len(found) {
	case 0:
		panic(fmt.Sprintf("field %s not found", name))
	case 1:
		return found[0]
	default:
		// the field without relation is the one we want
		for _, f := range found {
			if f.Relation() == nil {
				return f
			}
		}
		panic(fmt.Sprintf("ambiguous field %s", name))
	}
}

func (s *Schema) FieldFromColumn(column *ColumnDef) Field {
	if column.Relation == nil {
		return s.fieldByUnqualifiedName(column.Name)
	}
	return s.fieldByQualifiedName(column.Relation, column.Name)
}

// Merge modifies the current schema by merging it with another schema, any
// duplicate fields will be ignored.
// NOTE: should do this on clone of the schema.
func (s *Schema) Merge(other *Schema) *Schema {
	for _, f := range other.Fields {
		//duplicated := false
		//if f.Relation() != nil {
		//	duplicated = s.ContainsQualifiedColumn(f.Relation(), f.Name)
		//} else {
		//	duplicated = s.ContainsUnqualifiedColumn(f.Name)
		//}

		duplicated := s.ContainsColumn(f.Relation(), f.Name)
		if !duplicated {
			s.Fields = append(s.Fields, f)
		}
	}

	return s
}

func (s *Schema) ContainsUnqualifiedColumn(name string) bool {
	return slices.ContainsFunc(s.Fields, func(field Field) bool {
		return field.Name == name
	})
}

func (s *Schema) ContainsQualifiedColumn(relation *TableRef, name string) bool {
	return slices.ContainsFunc(s.Fields, func(field Field) bool {
		return field.Relation() == relation && field.Name == name
	})
}

// ContainsColumn checks if the schema contains the given column.
// It dispatches to ContainsQualifiedColumn or ContainsUnqualifiedColumn based
// on if the relation of the column is set.
func (s *Schema) ContainsColumn(relation *TableRef, name string) bool {
	if relation == nil {
		return s.ContainsUnqualifiedColumn(name)
	}
	return s.ContainsQualifiedColumn(relation, name)
}

//
//func (s *Schema) FieldFromColumn(column *ColumnDef) *Field {
//	if column.Relation == nil {
//		return s.FieldByName(column.Name)
//	}
//	return s.FieldByRelationAndName(column.Relation, column.Name)
//}

func (s *Schema) Clone() *Schema {
	return NewSchema(slices.Clone(s.Fields)...) //shallow clone
}
