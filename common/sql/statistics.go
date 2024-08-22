package sql

// NOTE: this file is TRANSITIONAL! These types are lifted from the
// unmerged internal/engine/costs/datatypes package.

import (
	"fmt"
	"strings"
)

// Statistics contains statistics about a table or a Plan. A Statistics can be
// derived directly from the underlying table, or derived from the statistics of
// its children.
type Statistics struct {
	RowCount int64

	ColumnStatistics []ColumnStatistics
	// NOTE: above may be better as []any to work with a generic ColStatsT[T]
}

func (s *Statistics) String() string {
	var st strings.Builder
	fmt.Fprintf(&st, "RowCount: %d", s.RowCount)
	if len(s.ColumnStatistics) > 0 {
		fmt.Fprintln(&st, "")
	}
	for i, cs := range s.ColumnStatistics {
		fmt.Fprintf(&st, " Column %d:\n", i)
		if _, ok := cs.Min.(string); ok {
			fmt.Fprintf(&st, " - Min/Max = %.64s / %.64s\n", cs.Min, cs.Max)
		} else {
			fmt.Fprintf(&st, " - Min/Max = %v / %v\n", cs.Min, cs.Max)
		}
		fmt.Fprintf(&st, " - NULL count = %v\n", cs.NullCount)
		fmt.Fprintf(&st, " - Num MCVs = %v\n", len(cs.MCFreqs))
		fmt.Fprintf(&st, " - Histogram = {%v}\n", cs.Histogram) // it's any, but also a fmt.Stringer
	}
	return st.String()
}

// ColumnStatistics contains statistics about a column.
type ColumnStatistics struct {
	NullCount int64

	Min      any
	MinCount int

	Max      any
	MaxCount int

	// MCVs are the most common values. It should be sorted by the value. It
	// should also be limited capacity, which means scan order has to be
	// deterministic since we have to throw out same-frequency observations.
	// (crap) Solution: multi-pass scan, merge lists, continue until no higher
	// freq values observed? OR when capacity reached, use a histogram? Do not
	// throw away MCVs, just start putting additional observations in to the
	// histogram instead.
	// MCVs []ValCount
	// MCVs map[cmp.Ordered]

	MCVals  []any // []T
	MCFreqs []int
	// ^ NOTE: MCVals was easier in many ways with just any.([]T), but other
	// ways much more inconvenient, so we have it as an []any.  May go back.

	// DistinctCount is harder. For example, unless we sub-sample
	// (deterministically), tracking distinct values could involve a data
	// structure with the same number of elements as rows in the table.
	// or sophisticated a algo e.g. https://github.com/axiomhq/hyperloglog
	// DistinctCount int64
	// alt, -1 means don't know

	// AvgSize can affect cost as it changes the number of "pages" in postgres
	// terminology, representing the size of data returned or processed by an
	// expression.
	AvgSize int64 // maybe: length of text, length of array, otherwise not used for scalar?

	Histogram any // histo[T]
}

/* Perhaps I should have started fresh with a fully generic column stats struct... under consideration.

type ColStatsT[T any] struct {
	NullCount int
	Min       T
	MinCount  int
	Max       T
	MaxCount  int
	MCVals    []T
	MCFreqs   []int
}
*/

func NewEmptyStatistics(numCols int) *Statistics {
	return &Statistics{
		RowCount:         0,
		ColumnStatistics: make([]ColumnStatistics, numCols),
	}
}

// ALL of the following types are from the initial query plan draft PR by Yaiba.
// Only TableRef gets much use in the current statistics work. An integration
// branch uses the other field and schema types a bit more, but it's easy to
// change any of this...

// TableRef is a PostgreSQL-schema-qualified table name.
type TableRef struct {
	Namespace string // e.g. schema in Postgres, derived from Kwil dataset schema DBID
	Table     string
}

// String returns the fully qualified table name as "namepace.table" if
// Namespace is set, otherwise it just returns the table name.
func (t *TableRef) String() string {
	if t.Namespace != "" {
		return fmt.Sprintf("%s.%s", t.Namespace, t.Table)
	}
	return t.Table
}

type ColumnDef struct {
	Relation *TableRef
	Name     string
}

func ColumnUnqualified(name string) *ColumnDef {
	return &ColumnDef{Name: name}
}

func Column(table *TableRef, name string) *ColumnDef {
	return &ColumnDef{Relation: table, Name: name}
}

// Field represents a field (column) in a schema.
type Field struct {
	Rel *TableRef

	Name     string
	Type     string
	Nullable bool
	HasIndex bool
}

func NewField(name string, dataType string, nullable bool) Field {
	return Field{Name: name, Type: dataType, Nullable: nullable}
}

func NewFieldWithRelation(name string, dataType string, nullable bool, relation *TableRef) Field {
	return Field{Name: name, Type: dataType, Nullable: nullable, Rel: relation}
}

func (f *Field) Relation() *TableRef {
	return f.Rel
}

func (f *Field) QualifiedColumn() *ColumnDef {
	return Column(f.Rel, f.Name)
}

// Schema represents a database as a slice of all columns in all relations. See
// also Field.
type Schema struct {
	Fields []Field
}

func NewSchema(fields ...Field) *Schema {
	return &Schema{Fields: fields}
}

func NewSchemaQualified(relation *TableRef, fields ...Field) *Schema {
	for i := range fields {
		fields[i].Rel = relation
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

type DataSource interface {
	// Schema returns the schema for the underlying data source
	Schema() *Schema
	// Statistics returns the statistics of the data source.
	Statistics() *Statistics
}
