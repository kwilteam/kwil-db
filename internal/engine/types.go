package engine

import (
	"strings"

	"github.com/kwilteam/kwil-db/core/types"
)

// Table is a table in the schema.
type Table struct {
	// Name is the name of the table.
	Name string
	// Columns is a list of columns in the table.
	Columns []*Column
	// Indexes is a list of indexes on the table.
	Indexes []*Index
	// Constraints are constraints on the table.
	Constraints map[string]*Constraint
}

func (t *Table) PrimaryKeyCols() []*Column {
	var pkCols []*Column
	for _, col := range t.Columns {
		if col.IsPrimaryKey {
			pkCols = append(pkCols, col)
		}
	}

	return pkCols
}

// HasPrimaryKey returns true if the column is part of the primary key.
func (t *Table) HasPrimaryKey(col string) bool {
	col = strings.ToLower(col)
	for _, c := range t.Columns {
		if c.Name == col && c.IsPrimaryKey {
			return true
		}
	}
	return false
}

// Column returns a column by name.
// If the column is not found, the second return value is false.
func (t *Table) Column(name string) (*Column, bool) {
	for _, col := range t.Columns {
		if col.Name == name {
			return col, true
		}
	}
	return nil, false
}

// SearchConstraint returns a list of constraints that match the given column and type.
func (t *Table) SearchConstraint(column string, constraint ConstraintType) []*Constraint {
	var constraints []*Constraint
	for _, c := range t.Constraints {
		if c.Type == constraint {
			for _, col := range c.Columns {
				if col == column {
					constraints = append(constraints, c)
				}
			}
		}
	}
	return constraints
}

// Column is a column in a table.
type Column struct {
	// Name is the name of the column.
	Name string
	// DataType is the data type of the column.
	DataType *types.DataType
	// Nullable is true if the column can be null.
	Nullable bool
	// IsPrimaryKey is true if the column is part of the primary key.
	IsPrimaryKey bool
}

// TODO: constraints should be tied to the table
// Constraint is a constraint in the schema.
type Constraint struct {
	// Name is the name of the constraint.
	// It must be unique within the schema.
	Name string
	// Type is the type of the constraint.
	Type ConstraintType
	// Columns is a list of column names that the constraint is on.
	Columns []string
}

type ConstraintType string

const (
	ConstraintUnique ConstraintType = "unique"
	ConstraintCheck  ConstraintType = "check"
	ConstraintFK     ConstraintType = "foreign_key"
)

// IndexType is a type of index (e.g. BTREE, UNIQUE_BTREE, PRIMARY)
type IndexType string

// Index is an index on a table.
type Index struct {
	Name    string    `json:"name"`
	Columns []string  `json:"columns"`
	Type    IndexType `json:"type"`
}

// index types
const (
	// BTREE is the default index type.
	BTREE IndexType = "BTREE"
	// UNIQUE_BTREE is a unique BTREE index.
	UNIQUE_BTREE IndexType = "UNIQUE_BTREE"
	// PRIMARY is a primary index.
	// Only one primary index is allowed per table.
	// A primary index cannot exist on a table that also has a primary key.
	PRIMARY IndexType = "PRIMARY"
)
