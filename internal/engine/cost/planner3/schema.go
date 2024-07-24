package planner3

import (
	"fmt"

	"github.com/kwilteam/kwil-db/core/types"
)

type Schema struct {
	Columns []*Column
}

func (s *Schema) ColumnsByParent(name string) []*Column {
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
func (s *Schema) Search(parent, name string) (*ProjectedColumn, error) {
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
		return &ProjectedColumn{
			Parent:     parent, // fully qualify the column
			Name:       column.Name,
			DataType:   column.DataType.Copy(),
			Aggregated: false,
		}, nil
	}

	for _, c := range s.Columns {
		if c.Parent == parent && c.Name == name {
			return &ProjectedColumn{
				Parent:     parent,
				Name:       c.Name,
				DataType:   c.DataType.Copy(),
				Aggregated: false,
			}, nil
		}
	}

	return nil, fmt.Errorf(`column "%s" not found in table "%s"`, name, parent)
}

func schemaFromTable(tbl *types.Table) *Schema {
	s := &Schema{}

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
	Parent     string // the parent relation name
	Name       string // the column name
	Aggregated bool   // whether the column is aggregated
	DataType   *types.DataType
}
