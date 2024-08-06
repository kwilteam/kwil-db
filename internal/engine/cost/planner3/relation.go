package planner3

import (
	"fmt"

	"github.com/kwilteam/kwil-db/core/types"
)

/*
	! writing this as of 8/1/2024
	I am super fucking stuck on how we can track column references across different scopes.
	In particular, queries that are both correlated and aggregated are hard to track, because
	we need to maintain aggregation rules across many scopes depending on where the subquery takes
	place. Handling this information in the SchemaContext makes sense, because it makes it easy to
	decide what a scope should care about, and thus I think the core issue is in the representation
	of columns themselves. There are two reasons I think this is the case.

	1. We are using the Column struct to represent column references in the
	SchemaContext, but this doesnt include any information as to whether the Column is aggregated.
	This is currently not possible because correlation is tracked at the subquery level, and not the
	column level.

	2. Enforcing rules on column references for aggregate queries is super fucking hard because we dont
	have concrete information on the columns being referenced when we need to apply the aggregation rules
	(which we are currenly trying to do at either the planning level(*1) or at the subquery relation analysis).
	Instead, it seems like the best place to handle this reference itself is within the LogicalExpression
	that references the column. Not only is this cleaner because it removes the need to re-traverse the
	expressions with a visitor, but it also gives us accurate error locations if we need to raise one.

	Therefore, my working theory is that our representation of columns themselves is inadequate. It is currently only
	based on the physical column within the schema, and _not_ based on what the query does with the column.

	(*1) I think tonight I have sufficiently ruled out that this logic should never be enforced within the planner.
	The planner itself should be dumb, and purely responsible for generating a logical plan based off of the statement
	and NOT the underlying schema.

	Moving forward, I will attempt to re-define column references within the schema context to be based off of the
	current analysis. I will still need to include information about the physical column, since this will be necessary
	for handling composite indexes.

	Things I am probably doing right:
	 - The order of operations in my plan construction is probably right. Reference rules that previously seemed arbitrary
	 in Postgres now totally make sense (and are even very predictable), so I think I am on the right track there.
	 - Separating the returned relation structure from the referenceable structure (the return value of Relation() vs the
	 contents of SchemaContext) is probably the right move. It's pretty clear how I can handle _most_ context passing logic
	 between scopes (aside from Column references). Furthermore, I think I can imagine a way of handling projection pushdown
	 with this model with a query optimizer, so it seems like a good move.
	 - Passing context on the current relation to subquery expressions while scoping subquery relations is probably right.
	 This is totally necessary for allowing subquery correlation in filters while not allowing correlation with joined
	 subqueries (which wouldn't even make sense, as I have no fucking clue what that would even do in terms of the relational
	 model).

	Things I am probably doing wrong:
	- Intermixing physical column info with referenced column info. This is a major source of my problems with handling
	subquery correlation detection (and enforcing subsequent aggregation rules).
	- Related to the above, but handling aggregation logic purely at the scope level is clearly wrong. I'm still unsure
	if moving this totally into some prospective new ColumnRef struct will be adequate (tbd), but I can definitely determine
	that only enforcing this logic at the scope level is inadequate.
*/

// PlanContext has the context for the database schema that the query planner
// is planning against, as well as holds important context such as common table
// expressions and other metadata.
type PlanContext struct {
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
	// CurrentRelation is the current relation in the query plan.
	// It is used to reference columns in the current relation.
	// These columns can be used in both expressions and returns.
	CurrentRelation *Relation

	// OuterRelation is the any relation in an outer query.
	// It is used to reference columns in the outer query
	// from a subquery (correlated subquery). These columns
	// can be used in both expressions, but not in returns.
	OuterRelation *Relation
}

// NewScope returns a shallow copied context where the current relation
// has become part of the outer relation. It also creates new aggregate
// metadata for the new context.
// It does not modify the original context.
func (s *PlanContext) NewScope() *PlanContext {
	newContext := *s // shallow copy

	if s.OuterRelation == nil {
		s.OuterRelation = &Relation{}
	}
	if s.CurrentRelation == nil {
		s.CurrentRelation = &Relation{}
	}

	newContext.OuterRelation = &Relation{
		Columns: append(s.CurrentRelation.Columns, s.OuterRelation.Columns...),
	}

	newContext.CurrentRelation = &Relation{}

	return &newContext
}

// Copy copies the schema context.
// It shallow copies everything except for the current relation, which
// it will deep copy everything (including the columns themselves).
// It should be used when entering Subquery expressions, but not subquery
// relations.
func (s *PlanContext) Copy() *PlanContext {
	newContext := *s // shallow copy
	rel := &Relation{}

	if s.CurrentRelation != nil {
		for _, c := range s.CurrentRelation.Columns {
			rel.Columns = append(rel.Columns, c)
		}
	}

	newContext.CurrentRelation = rel
	return &newContext
}

// planRelation takes a LogicalPlan and updates the context based on the contents
// of the plan. It returns the relation that the plan represents.
// It will perform type validations.
func (s *PlanContext) planRelation(rel LogicalPlan) (*Relation, error) {
	switch n := rel.(type) {
	default:
		panic(fmt.Sprintf("unexpected node type %T", n))
	case *TableScanSource:
		// TODO: idk if the below comment will be true. REVISIT
		// we will add the table to the context in scan plan
		tbl, ok := s.Schema.FindTable(n.TableName)
		if !ok {
			return nil, fmt.Errorf(`table "%s" not found`, n.TableName)
		}

		return relationFromTable(tbl), nil
	case *ProcedureScanSource:
		// TODO: idk if the below comment will be true. REVISIT
		// we will add the procedure relation to the context in scan plan

		// should either be a foreign procedure or a local procedure
		var expectedArgs []*types.DataType
		var returns *types.ProcedureReturn
		if n.IsForeign {
			proc, ok := s.Schema.FindForeignProcedure(n.ProcedureName)
			if !ok {
				return nil, fmt.Errorf(`foreign procedure "%s" not found`, n.ProcedureName)
			}
			returns = proc.Returns
			expectedArgs = proc.Parameters

			if len(n.ContextualArgs) != 2 {
				return nil, fmt.Errorf("foreign procedure requires 2 arguments")
			}

			// both arguments should be strings
			if err := s.evaluatesTo(n.ContextualArgs, []*types.DataType{types.TextType, types.TextType}, &Relation{}); err != nil {
				return nil, err
			}
		} else {
			proc, ok := s.Schema.FindProcedure(n.ProcedureName)
			if !ok {
				return nil, fmt.Errorf(`procedure "%s" not found`, n.ProcedureName)
			}

			returns = proc.Returns
			for _, arg := range proc.Parameters {
				expectedArgs = append(expectedArgs, arg.Type)
			}
		}
		if returns == nil {
			return nil, fmt.Errorf(`procedure "%s" does not return anything`, n.ProcedureName)
		}
		if !returns.IsTable {
			return nil, fmt.Errorf(`procedure "%s" does not return a table`, n.ProcedureName)
		}

		// there is no current relation that exprs can be evaluated against
		// because we are in a scan
		if err := s.evaluatesTo(n.Args, expectedArgs, &Relation{}); err != nil {
			return nil, err
		}

		var cols []*ReferenceableColumn
		for _, field := range returns.Fields {
			cols = append(cols, &ReferenceableColumn{
				// the Parent will get set by the ScanAlias
				Name:     field.Name,
				DataType: field.Type.Copy(),
			})
		}

		return &Relation{Columns: cols}, nil
	case *SubqueryScanSource:
		return s.planRelation(n.Subquery)
	case *Scan:
		rel, err := s.planRelation(n.Child)
		if err != nil {
			return nil, err
		}

		for _, col := range rel.Columns {
			col.Parent = n.RelationName
		}

		return rel, nil
	case *Project:
		rel, err := s.planRelation(n.Child)
		if err != nil {
			return nil, err
		}

		var fields []*Field
	}
}

// areOfType is a helper method that checks if a slice of LogicalExprs will evaluate
// to the slice of data types.
func (s *PlanContext) evaluatesTo(exprs []LogicalExpr, types []*types.DataType, currentRel *Relation) error {
	if len(exprs) != len(types) {
		return fmt.Errorf("expected %d expressions, got %d", len(types), len(exprs))
	}

	for i, expr := range exprs {
		dt, err := s.planExpression(expr, currentRel)
		if err != nil {
			return err
		}

		scalar, err := dt.Scalar()
		if err != nil {
			return err
		}

		if !scalar.Equals(types[i]) {
			return fmt.Errorf("expected expression %d to be of type %s, got %s", i+1, types[i], scalar)
		}
	}

	return nil
}

// planExpression takes a LogicalExpr and updates the context based on the contents
// of the expression. It returns the ReturnableType of the expression.
// the currentRel is the relation that the expression is being evaluated in.
func (s *PlanContext) planExpression(expr LogicalExpr, currentRel *Relation) (*Field, error) {

}

// Join returns a new shallow copied context that joins the given
// relation with the current relation.
func (s *PlanContext) Join(relation *Relation) *PlanContext {
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

// join joins the current relation with the given relation.
func (s *PlanContext) join(relation *Relation) {
	if s.CurrentRelation == nil {
		s.CurrentRelation = &Relation{}
	}

	s.CurrentRelation.Columns = append(s.CurrentRelation.Columns, relation.Columns...)
}

// Relation is the current relation in the query plan.
type Relation struct {
	Columns []*ReferenceableColumn
}

func (s *Relation) ColumnsByParent(name string) []*ReferenceableColumn {
	var columns []*ReferenceableColumn
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
func (s *Relation) Search(parent, name string) (*ReferenceableColumn, error) {
	if parent == "" {
		var column *ReferenceableColumn
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
		return &ReferenceableColumn{
			Parent:   parent, // fully qualify the column
			Name:     column.Name,
			DataType: column.DataType.Copy(),
		}, nil
	}

	for _, c := range s.Columns {
		if c.Parent == parent && c.Name == name {
			return c, nil
		}
	}

	return nil, fmt.Errorf(`column "%s" not found in table "%s"`, name, parent)
}

// Relation2 is a relation that is returned from a query.
// TODO: delete Relation in favor of Relation2
type Relation2 struct {
	Fields []*Field
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

// ReferenceableColumn is a column that can be referenced in a query.
// They are used to represent columns that can be used in expressions.
type ReferenceableColumn struct {
	Parent   string          // the parent relation name
	Name     string          // the column name
	DataType *types.DataType // the column data type
}

// Field is a field in a relation.
// Parent and Name can be empty, if the expression
// is a constant.
type Field struct {
	Parent string // the parent relation name
	Name   string // the field name
	// val is the value of the field.
	// it can be either a single value or a map of values,
	// depending on the field type.
	// This value should be accessed using the Scalar() or Object()
	val any
}

func (f *Field) Scalar() (*types.DataType, error) {
	dt, ok := f.val.(*types.DataType)
	if !ok {
		// can be triggered by a user if they try to directly use an object
		_, ok = f.val.(map[string]*types.DataType)
		if ok {
			return nil, fmt.Errorf("referenced field is an object, expected scalar or array. specify a field to access using the . operator")
		}

		// not user error
		panic(fmt.Sprintf("unexpected return type %T", f.val))
	}
	return dt, nil
}

func (f *Field) Object() (map[string]*types.DataType, error) {
	obj, ok := f.val.(map[string]*types.DataType)
	if !ok {
		// this can be triggered by a user if they try to use dot notation
		// on a scalar
		v, ok := f.val.(*types.DataType)
		if ok {
			if v.IsArray {
				return nil, fmt.Errorf("referenced expression is an array, expected object")
			}
			return nil, fmt.Errorf("referenced expression is a scalar, expected object")
		}

		// this is an internal bug
		panic(fmt.Sprintf("unexpected return type %T", f.val))
	}
	return obj, nil
}

// ! (8/1/2024)
// ProjectedColumn is _maybe_ on the right track, but should probably be rethought.
// It was created when I was trying to enforce aggregation rules in the planner.

// ProjectedColumn represents a column in a projection.
type ProjectedColumn struct {
	Parent string // the parent relation name
	Name   string // the column name
	// If the column is referenced inside of an aggregate function,
	// (e.g. sum(col_name)), then Aggregated will be true.
	Aggregated bool
	DataType   *types.DataType
}

// ReturnedType is a struct that is returned from the Scalar() method
// of LogicalExpr implementations. It can be used to coerce the return type,
// and to handle error returns. Callers should never access the fields directly.
type ReturnedType struct {
	// val is the data type that is returned by the expression.
	// It is either a single data type or a map of data types.
	val any
	// err is the error that was returned during the evaluation of the expression.
	// It is added here as a convenience so that DataType itself does not have to
	// return an error, requiring the callers to check for errors twice.
	err error
}

// Scalar attempts to coerce the return type to a single data type.
func (r *ReturnedType) Scalar() (*types.DataType, error) {
	if r.err != nil {
		return nil, r.err
	}

	dt, ok := r.val.(*types.DataType)
	if !ok {
		// this can be triggered by a user if they try to directly use an object
		// in an expression
		_, ok = r.val.(map[string]*types.DataType)
		if ok {
			return nil, fmt.Errorf("referenced expression is an object, expected scalar or array. specify a field to access using the . operator")
		}

		// this is an internal bug
		panic(fmt.Sprintf("unexpected return type %T", r.val))
	}
	return dt, nil
}

// Object attempts to coerce the return type to a map of data types.
func (r *ReturnedType) Object() (map[string]*types.DataType, error) {
	if r.err != nil {
		return nil, r.err
	}

	obj, ok := r.val.(map[string]*types.DataType)
	if !ok {
		// this can be triggered by a user if they try to use dot notation
		// on a scalar
		v, ok := r.val.(*types.DataType)
		if ok {
			if v.IsArray {
				return nil, fmt.Errorf("referenced expression is an array, expected object")
			}
			return nil, fmt.Errorf("referenced expression is a scalar, expected object")
		}

		// this is an internal bug
		panic(fmt.Sprintf("unexpected return type %T", r.val))
	}
	return obj, nil
}
