package planner3

import (
	"fmt"

	"github.com/kwilteam/kwil-db/core/types"
)

// SchemaContext has the context for the database schema that the query planner
// is planning against, as well as holds important context such as common table
// expressions and other metadata.
type SchemaContext struct {
	// Schema is the underlying database schema that the query should
	// be evaluated against.
	Schema *types.Schema
	// CTEs are the common table expressions in the query.
	// This field should be updated as the query planner
	// processes the query.
	CTEs map[string]*Relation
	// Variables are the variables in the query.
	Variables map[string]*types.DataType
	// Objects are the objects in the query.
	// Kwil supports one-dimensional objects, so this would be
	// accessible via objname.fieldname.
	Objects map[string]map[string]*types.DataType
	// OuterRelation is the any relation in an outer query.
	// It is used to reference columns in the outer query
	// from a subquery (correlated subquery).
	OuterRelation *Relation
}

// Join returns a new shallow copied context that joins the given
// relation with the current relation.
func (s *SchemaContext) Join(relation *Relation) *SchemaContext {
	if s.OuterRelation == nil {
		s.OuterRelation = &Relation{}
	}

	newContext := *s     // shallow copy
	newRel := &Relation{ // new relation to not mutate the originals
		Columns: append(relation.Columns, s.OuterRelation.Columns...),
	}
	newContext.OuterRelation = newRel
	return &newContext
}

// Relation is the current relation in the query plan.
type Relation struct {
	Columns []*Column
}

func (s *Relation) ColumnsByParent(name string) []*Column {
	var columns []*Column
	for _, c := range s.Columns {
		if c.Parent == name {
			columns = append(columns, c)
		}
	}
	return columns
}

// Search searches for a column by parent and name.
// If the column is not found, an error is returned.
// If no parent is specified and many columns have the same name,
// an error is returned. The returned column will always be qualified.
func (s *Relation) Search(parent, name string) (*Column, error) {
	if parent == "" {
		var column *Column
		count := 0
		for _, c := range s.Columns {
			if c.Name == name {
				column = c
				count++
			}
		}
		if count == 0 {
			return nil, fmt.Errorf(`column "%s" not found`, name)
		}
		if count > 1 {
			return nil, fmt.Errorf(`column "%s" is ambiguous`, name)
		}

		// return a new instance since we are qualifying the column
		return &Column{
			Parent:    parent, // fully qualify the column
			Name:      column.Name,
			DataType:  column.DataType.Copy(),
			Nullable:  column.Nullable,
			HasIndex:  column.HasIndex,
			HasUnique: column.HasUnique,
		}, nil
	}

	for _, c := range s.Columns {
		if c.Parent == parent && c.Name == name {
			return c, nil
		}
	}

	return nil, fmt.Errorf(`column "%s" not found in table "%s"`, name, parent)
}

func relationFromTable(tbl *types.Table) *Relation {
	s := &Relation{}

	isNullable := func(col *types.Column) bool {
		isNullable := true
		for _, a := range col.Attributes {
			if a.Type == types.NOT_NULL || a.Type == types.PRIMARY_KEY {
				isNullable = false
			}
		}

		return isNullable
	}

	hasIndexAndUnique := func(col *types.Column) (hasIndex bool, isUnique bool) {
		for _, attr := range col.Attributes {
			if attr.Type == types.PRIMARY_KEY || attr.Type == types.UNIQUE {
				isUnique = true
				hasIndex = true
			}
		}

		for _, idx := range tbl.Indexes {
			// TODO: we need to account for composite indexes.
			// this is a deficiency in our current representation of columns,
			// so it cannot be accounted for here.

			if len(idx.Columns) == 1 && idx.Columns[0] == col.Name {
				hasIndex = true
				if idx.Type == types.PRIMARY || idx.Type == types.UNIQUE_BTREE {
					isUnique = true
				}
			}
		}

		return hasIndex, isUnique
	}

	for _, col := range tbl.Columns {
		newCol := &Column{
			Parent:   tbl.Name,
			Name:     col.Name,
			DataType: col.Type.Copy(),
		}

		newCol.Nullable = isNullable(col)

		newCol.HasIndex, newCol.HasUnique = hasIndexAndUnique(col)
	}

	return s
}

type Column struct {
	Parent   string          // Parent relation name
	Name     string          // Column name
	DataType *types.DataType // Column data type
	Nullable bool            // Column is nullable
	// TODO: we don't have a way to account for composite indexes.
	// This is ok for now, it will just make our cost estimates higher
	// for index seeks on composite indexes / primary keys.
	HasIndex  bool // Column has an index
	HasUnique bool // Column has a unique constraint or unique index
}

// ProjectedColumn represents a column in a projection.
type ProjectedColumn struct {
	Parent string // the parent relation name
	Name   string // the column name
	// If the column is referenced inside of an aggregate function,
	// (e.g. sum(col_name)), then Aggregated will be true.
	Aggregated bool
	DataType   *types.DataType
}
