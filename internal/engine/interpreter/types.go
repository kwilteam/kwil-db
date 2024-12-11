package interpreter

import (
	"fmt"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/parse"
)

type Action struct {
	// Name is the name of the action.
	// It should always be lower case.
	Name string `json:"name"`

	// Parameters are the input parameters of the action.
	Parameters []*NamedType `json:"parameters"`

	// Public is true if the action is public.
	Public bool `json:"public"`

	// Modifiers modify the access to the action.
	Modifiers []Modifier `json:"modifiers"`

	// Body is the logic of the action.
	// TODO: delete this and just pass around strings
	Body []parse.ActionStmt

	// RawStatement is the unparsed CREATE ACTION statement.
	RawStatement string `json:"raw_statement"`

	// Returns specifies the return types of the action.
	Returns *ActionReturn `json:"return_types"`
}

// FromAST sets the fields of the action from an AST node.
func (a *Action) FromAST(ast *parse.CreateActionStatement) error {
	a.Name = ast.Name
	a.Public = ast.Public
	a.RawStatement = ast.Raw
	a.Body = ast.Statements

	a.Parameters = convertNamedTypes(ast.Parameters)

	if ast.Returns != nil {
		a.Returns = &ActionReturn{
			IsTable: ast.Returns.IsTable,
			Fields:  convertNamedTypes(ast.Returns.Fields),
		}
	}

	modSet := make(map[Modifier]struct{})
	a.Modifiers = []Modifier{}
	for _, m := range ast.Modifiers {
		mod := Modifier(m)

		if !mod.Valid() {
			return fmt.Errorf("unknown modifier %s", mod)
		}

		if _, ok := modSet[mod]; !ok {
			modSet[mod] = struct{}{}
			a.Modifiers = append(a.Modifiers, mod)
		}
	}

	return nil
}

// convertNamedTypes converts a list of named types from the AST to the internal representation.
func convertNamedTypes(params []*parse.NamedType) []*NamedType {
	namedTypes := make([]*NamedType, len(params))
	for i, p := range params {
		namedTypes[i] = &NamedType{
			Name: p.Name,
			Type: p.Type,
		}
	}
	return namedTypes
}

// OwnerOnly returns true if the action is owner only.
func (a *Action) OwnerOnly() bool {
	for _, m := range a.Modifiers {
		if m == ModifierOwner {
			return true
		}
	}
	return false
}

// IsView returns true if the action is a view.
func (a *Action) IsView() bool {
	for _, m := range a.Modifiers {
		if m == ModifierView {
			return true
		}
	}
	return false
}

// NamedType is a parameter in a procedure.
type NamedType struct {
	// Name is the name of the parameter.
	// It should always be lower case.
	// If it is a procedure parameter, it should begin
	// with a $.
	Name string `json:"name"`
	// Type is the type of the parameter.
	Type *types.DataType `json:"type"`
}

// ActionReturn holds the return type of a procedure.
// EITHER the Type field is set, OR the Table field is set.
type ActionReturn struct {
	IsTable bool         `json:"is_table"`
	Fields  []*NamedType `json:"fields"`
}

// Modifier modifies the access to a procedure.
type Modifier string

const (
	// View means that an action does not modify the database.
	ModifierView Modifier = "VIEW"

	// Owner requires that the caller is the owner of the database.
	ModifierOwner Modifier = "OWNER"
)

func (m Modifier) Valid() bool {
	switch m {
	case ModifierView, ModifierOwner:
		return true
	}
	return false
}

// // Table is a table in the schema.
// type Table struct {
// 	// Name is the name of the table.
// 	Name string
// 	// Columns is a list of columns in the table.
// 	Columns []*Column
// 	// Indexes is a list of indexes on the table.
// 	Indexes []*Index
// 	// Constraints are constraints on the table.
// 	Constraints map[string]*Constraint
// }

// func (t *Table) PrimaryKeyCols() []*Column {
// 	var pkCols []*Column
// 	for _, col := range t.Columns {
// 		if col.IsPrimaryKey {
// 			pkCols = append(pkCols, col)
// 		}
// 	}

// 	return pkCols
// }

// // HasPrimaryKey returns true if the column is part of the primary key.
// func (t *Table) HasPrimaryKey(col string) bool {
// 	col = strings.ToLower(col)
// 	for _, c := range t.Columns {
// 		if c.Name == col && c.IsPrimaryKey {
// 			return true
// 		}
// 	}
// 	return false
// }

// // Column returns a column by name.
// // If the column is not found, the second return value is false.
// func (t *Table) Column(name string) (*Column, bool) {
// 	for _, col := range t.Columns {
// 		if col.Name == name {
// 			return col, true
// 		}
// 	}
// 	return nil, false
// }

// // SearchConstraint returns a list of constraints that match the given column and type.
// func (t *Table) SearchConstraint(column string, constraint ConstraintType) []*Constraint {
// 	var constraints []*Constraint
// 	for _, c := range t.Constraints {
// 		if c.Type == constraint {
// 			for _, col := range c.Columns {
// 				if col == column {
// 					constraints = append(constraints, c)
// 				}
// 			}
// 		}
// 	}
// 	return constraints
// }

// // Column is a column in a table.
// type Column struct {
// 	// Name is the name of the column.
// 	Name string
// 	// DataType is the data type of the column.
// 	DataType *types.DataType
// 	// DefaultValue is the default value of the column.
// 	DefaultValue any // can be nil
// 	// Nullable is true if the column can be null.
// 	Nullable bool
// 	// IsPrimaryKey is true if the column is part of the primary key.
// 	IsPrimaryKey bool
// }

// // TODO: constraints should be tied to the table
// // Constraint is a constraint in the schema.
// type Constraint struct {
// 	// Name is the name of the constraint.
// 	// It must be unique within the schema.
// 	Name string
// 	// Type is the type of the constraint.
// 	Type ConstraintType
// 	// Columns is a list of column names that the constraint is on.
// 	Columns []string
// }

// type ConstraintType string

// const (
// 	ConstraintUnique ConstraintType = "unique"
// 	ConstraintCheck  ConstraintType = "check"
// 	ConstraintFK     ConstraintType = "foreign_key"
// )

// // IndexType is a type of index (e.g. BTREE, UNIQUE_BTREE, PRIMARY)
// type IndexType string

// // Index is an index on a table.
// type Index struct {
// 	Name    string    `json:"name"`
// 	Columns []string  `json:"columns"`
// 	Type    IndexType `json:"type"`
// }

// // index types
// const (
// 	// BTREE is the default index type.
// 	BTREE IndexType = "BTREE"
// 	// UNIQUE_BTREE is a unique BTREE index.
// 	UNIQUE_BTREE IndexType = "UNIQUE_BTREE"
// 	// PRIMARY is a primary index.
// 	// Only one primary index is allowed per table.
// 	// A primary index cannot exist on a table that also has a primary key.
// 	PRIMARY IndexType = "PRIMARY"
// )
