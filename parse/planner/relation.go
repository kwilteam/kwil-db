package planner

import (
	"errors"
	"fmt"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/parse"
)

// EvaluateContext has the context for the database schema that the query planner
// is planning against, as well as holds important context such as common table
// expressions and other metadata.
type EvaluateContext struct {
	plan *planContext

	// OuterRelation is the any relation in an outer query.
	// It is used to reference columns in the outer query
	// from a subquery (correlated subquery). These columns
	// can be used in both expressions, but not in returns.
	OuterRelation *Relation
	// Correlations are the columns that are correlated in the query.
	Correlations []*ColumnRef
}

func newEvalCtx(plan *planContext) *EvaluateContext {
	return &EvaluateContext{
		plan:          plan,
		OuterRelation: &Relation{},
	}
}

// evalRelation takes a LogicalPlan and updates the context based on the contents
// of the plan. It returns the relation that the plan represents.
// It will perform type validations.
func (s *EvaluateContext) evalRelation(rel LogicalPlan) (*Relation, error) {
	switch n := rel.(type) {
	case *EmptyScan:
		return &Relation{}, nil
	case *Scan:
		rel, err := s.evalScanSource(n.Source)
		if err != nil {
			return nil, err
		}

		for _, col := range rel.Fields {
			col.Parent = n.RelationName
		}

		return rel, nil
	case *Return:
		rel, err := s.evalRelation(n.Child)
		if err != nil {
			return nil, err
		}

		if len(rel.Fields) != len(n.Fields) {
			return nil, fmt.Errorf("expected %d columns, got %d", len(n.Fields), len(rel.Fields))
		}

		newFields := make([]*Field, len(n.Fields))
		for i, field := range rel.Fields {
			scalar, err := field.Scalar()
			if err != nil {
				return nil, err
			}

			newFields[i] = &Field{
				Name: n.Fields[i],
				val:  scalar,
				// we discard the parent because it is not needed,
				// as postgres does not return parent names in the result
			}
		}

		return &Relation{Fields: newFields}, nil
	case *Project:
		rel, err := s.evalRelation(n.Child)
		if err != nil {
			return nil, err
		}

		for _, expand := range n.expandFuncs {
			n.Expressions = append(n.Expressions, expand(rel)...)
		}
		n.expandFuncs = nil // never want to expand more than once

		fields, err := s.planManyExpressions(n.Expressions, rel)
		if err != nil {
			return nil, err
		}

		return &Relation{Fields: fields}, nil
	case *Filter:
		rel, err := s.evalRelation(n.Child)
		if err != nil {
			return nil, err
		}

		if err := s.evalsTo(n.Condition, types.BoolType, rel); err != nil {
			return nil, err
		}

		return rel, nil
	case *Join:
		left, err := s.evalRelation(n.Left)
		if err != nil {
			return nil, err
		}
		right, err := s.evalRelation(n.Right)
		if err != nil {
			return nil, err
		}

		rel := &Relation{
			Fields: append(left.Fields, right.Fields...),
		}

		if err := s.evalsTo(n.Condition, types.BoolType, rel); err != nil {
			return nil, err
		}

		return rel, nil
	case *Sort:
		rel, err := s.evalRelation(n.Child)
		if err != nil {
			return nil, err
		}

		for _, expr := range n.SortExpressions {
			field, err := s.evalExpression(expr.Expr, rel)
			if err != nil {
				return nil, err
			}

			// as long as the expression is a scalar, we are good
			_, err = field.Scalar()
			if err != nil {
				return nil, err
			}
		}

		return rel, nil
	case *Limit:
		rel, err := s.evalRelation(n.Child)
		if err != nil {
			return nil, err
		}

		if n.Offset != nil {
			if err := s.evalsTo(n.Offset, types.IntType, rel); err != nil {
				return nil, err
			}
		}

		field, err := s.evalExpression(n.Limit, rel)
		if err != nil {
			return nil, err
		}

		_, err = field.Scalar()
		if err != nil {
			return nil, err
		}

		return rel, nil
	case *Aggregate:
		rel, err := s.evalRelation(n.Child)
		if err != nil {
			return nil, err
		}

		// TODO: we need to use aggregate.go to enforce aggregation rules

		if _, err := s.manyAreScalar(n.GroupingExpressions, rel); err != nil {
			return nil, err
		}

		// need to do this manually because it is a concrete type
		for _, agg := range n.AggregateExpressions {
			field, err := s.evalExpression(agg, rel)
			if err != nil {
				return nil, err
			}

			_, err = field.Scalar()
			if err != nil {
				return nil, err
			}
		}

		// TODO: this is incorrect, we actually should return the grouping expressions
		// and the aggregate expressions

		return rel, nil
	case *Distinct:
		return s.evalRelation(n.Child)
	case *SetOperation:
		left, err := s.evalRelation(n.Left)
		if err != nil {
			return nil, err
		}

		right, err := s.evalRelation(n.Right)
		if err != nil {
			return nil, err
		}

		if len(left.Fields) != len(right.Fields) {
			return nil, fmt.Errorf("set operations must have the same number of columns")
		}

		for i := range left.Fields {
			leftScal, err := left.Fields[i].Scalar()
			if err != nil {
				return nil, err
			}

			rightScal, err := right.Fields[i].Scalar()
			if err != nil {
				return nil, err
			}

			if !leftScal.Equals(rightScal) {
				return nil, fmt.Errorf("compound operations must have the same data types")
			}
		}

		// parent tables cannot be referenced after a set operation.
		// e.g. "SELECT * FROM users UNION SELECT * FROM posts SORT BY id;" is valid,
		// but "SELECT * FROM users UNION SELECT * FROM posts SORT BY users.id;" is not

		for _, col := range left.Fields {
			col.Parent = ""
		}

		return left, nil
	case *Subplan:
		rel, err := s.evalRelation(n.Plan)
		if err != nil {
			return nil, err
		}

		return rel, nil
	case *CartesianProduct:
		left, err := s.evalRelation(n.Left)
		if err != nil {
			return nil, err
		}

		right, err := s.evalRelation(n.Right)
		if err != nil {
			return nil, err
		}

		return &Relation{
			Fields: append(left.Fields, right.Fields...),
		}, nil
	case *Insert:
		// for modifications, we simply evaluate and typecheck the expressions.
		// They dont return relations since they are not queries and we don't
		// support RETURNING clauses

		var expectedTypes []*types.DataType
		tbl, ok := s.plan.Schema.FindTable(n.Table)
		if !ok {
			return nil, fmt.Errorf(`table "%s" not found`, n.Table)
		}

		// colSet for quick lookup later
		colSet := make(map[string]*types.DataType)
		for _, col := range tbl.Columns {
			colSet[col.Name] = col.Type
			expectedTypes = append(expectedTypes, col.Type.Copy())
		}

		for _, newTuple := range n.Values {
			for i, expr := range newTuple {
				if err := s.evalsTo(expr, expectedTypes[i], &Relation{}); err != nil {
					return nil, err
				}
			}
		}

		tableRel := relationFromTable(tbl)

		if n.ConflictResolution != nil {
			res, ok := n.ConflictResolution.(*ConflictUpdate)
			if ok {
				for _, expr := range res.Assignments {
					// TODO: we need a way to get an EXCLUDED relation: https://www.jooq.org/doc/latest/manual/sql-building/sql-statements/insert-statement/insert-on-conflict-excluded/
					// TODO: we also need to include the alias here!
					scalar, err := s.isScalar(expr.Value, tableRel)
					if err != nil {
						return nil, err
					}

					expected, ok := colSet[expr.Column]
					if !ok {
						return nil, fmt.Errorf(`column "%s" not found in table "%s"`, expr.Column, n.Table)
					}

					if !expected.Equals(scalar) {
						return nil, fmt.Errorf(`conflict resolution requires the same type as the column "%s"`, expr.Column)
					}
				}

				if res.ConflictFilter != nil {
					scalar, err := s.isScalar(res.ConflictFilter, tableRel)
					if err != nil {
						return nil, err
					}

					if !scalar.Equals(types.BoolType) {
						return nil, fmt.Errorf("conflict filter requires a boolean type, got %s", scalar)
					}
				}
			}
			// nothing to do if it is a DO NOTHING resolution
		}

		return &Relation{}, nil
	case *Update:
		tbl, ok := s.plan.Schema.FindTable(n.Table)
		if !ok {
			return nil, fmt.Errorf(`table "%s" not found`, n.Table)
		}

		colSet := make(map[string]*types.DataType)
		for _, col := range tbl.Columns {
			colSet[col.Name] = col.Type
		}

		child, err := s.evalRelation(n.Child)
		if err != nil {
			return nil, err
		}

		for _, expr := range n.Assignments {
			scalar, err := s.isScalar(expr.Value, child)
			if err != nil {
				return nil, err
			}

			expected, ok := colSet[expr.Column]
			if !ok {
				return nil, fmt.Errorf(`column "%s" not found in table "%s"`, expr.Column, n.Table)
			}

			if !expected.Equals(scalar) {
				return nil, fmt.Errorf(`assignment requires the same type as the column "%s"`, expr.Column)
			}
		}

		return &Relation{}, nil
	case *Delete:
		_, ok := s.plan.Schema.FindTable(n.Table)
		if !ok {
			return nil, fmt.Errorf(`table "%s" not found`, n.Table)
		}

		_, err := s.evalRelation(n.Child)
		if err != nil {
			return nil, err
		}

		// we simply need to check these exist and are scalar
		return &Relation{}, nil
	}

	return nil, fmt.Errorf("unexpected node type %T", rel)
}

// planManyExpressions plans many expressions and returns the Field of the
// expressions. It will return an error if any of the expressions are invalid.
func (s *EvaluateContext) planManyExpressions(exprs []LogicalExpr, currentRel *Relation) ([]*Field, error) {
	var fields []*Field
	for _, expr := range exprs {
		field, err := s.evalExpression(expr, currentRel)
		if err != nil {
			return nil, err
		}

		fields = append(fields, field)
	}

	return fields, nil
}

// evalScanSource evaluates the source of a scan and returns the relation that
// the scan represents. It will perform type validations.
func (s *EvaluateContext) evalScanSource(source ScanSource) (*Relation, error) {
	switch n := source.(type) {
	case *TableScanSource:
		switch n.Type {
		case TableSourcePhysical:
			tbl, ok := s.plan.Schema.FindTable(n.TableName)
			if !ok {
				return nil, fmt.Errorf(`table "%s" not found`, n.TableName)
			}

			n.rel = relationFromTable(tbl)
			return n.rel.Copy(), nil
		case TableSourceCTE:
			cte, ok := s.plan.CTEs[n.TableName]
			if !ok {
				return nil, fmt.Errorf(`cte "%s" not found`, n.TableName)
			}

			n.rel = cte.Copy()
			return cte.Copy(), nil
		default:
			panic(fmt.Sprintf("unexpected table source type %d", n.Type))
		}
	case *ProcedureScanSource:
		// should either be a foreign procedure or a local procedure
		var expectedArgs []*types.DataType
		var returns *types.ProcedureReturn
		if n.IsForeign {
			proc, ok := s.plan.Schema.FindForeignProcedure(n.ProcedureName)
			if !ok {
				return nil, fmt.Errorf(`foreign procedure "%s" not found`, n.ProcedureName)
			}
			returns = proc.Returns
			expectedArgs = proc.Parameters

			if len(n.ContextualArgs) != 2 {
				return nil, fmt.Errorf("foreign procedure requires 2 arguments")
			}

			// both arguments should be strings
			if err := s.manyEvalTo(n.ContextualArgs, []*types.DataType{types.TextType, types.TextType}, &Relation{}); err != nil {
				return nil, err
			}
		} else {
			proc, ok := s.plan.Schema.FindProcedure(n.ProcedureName)
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
		if err := s.manyEvalTo(n.Args, expectedArgs, &Relation{}); err != nil {
			return nil, err
		}

		var cols []*Field
		for _, field := range returns.Fields {
			cols = append(cols, &Field{
				// the Parent will get set by the ScanAlias
				Name: field.Name,
				val:  field.Type.Copy(),
			})
		}

		n.rel = &Relation{Fields: cols}

		return n.rel.Copy(), nil
	case *Subquery:
		if !n.ReturnsRelation {
			panic("internal bug: planner planned a join against a scalar subquery")
		}

		// we pass an empty relation because the subquery can't
		// refer to the current relation, but they can be correlated against some
		// outer relation.
		// for example, "select * from users u inner join (select * from posts where posts.id = u.id) as p on u.id=p.id;"
		// is invalid, but
		// "select * from users where id = (select posts.id from posts inner join (select * from posts where id = users.id) as s on s.id=posts.id);"
		// is valid
		rel, err := s.evalSubquery(n, &Relation{})
		if err != nil {
			return nil, err
		}

		return rel, nil
	}

	return nil, fmt.Errorf("unexpected node type %T", source)
}

// evalExpression takes a LogicalExpr and updates the context based on the contents
// of the expression. It returns the Field of the expression.
// the currentRel is the relation that the expression is being evaluated in.
func (s *EvaluateContext) evalExpression(expr LogicalExpr, currentRel *Relation) (*Field, error) {
	switch n := expr.(type) {
	case *Literal:
		return anonField(n.Type), nil
	case *Variable:
		dt, ok := s.plan.Variables[n.VarName]
		if !ok {
			// might be an object
			obj, ok := s.plan.Objects[n.VarName]
			if !ok {
				return nil, fmt.Errorf(`variable "%s" not found`, n.VarName)
			}

			copyMap := make(map[string]*types.DataType)
			for k, v := range obj {
				copyMap[k] = v.Copy()
			}
			n.dataType = copyMap

			return &Field{
				Name: n.VarName,
				val:  obj,
			}, nil
		}

		n.dataType = dt.Copy()

		return anonField(dt), nil
	case *ColumnRef:
		field, err := currentRel.Search(n.Parent, n.ColumnName)
		if errors.Is(err, errColumnNotFound) {
			// check outer relation
			field, err = s.OuterRelation.Search(n.Parent, n.ColumnName)
			if err != nil {
				return nil, err
			}

			// add a copy of the column to the correlations
			s.Correlations = append(s.Correlations, &ColumnRef{
				Parent:     n.Parent,
				ColumnName: n.ColumnName,
			})

			n.Parent = field.Parent

			sc, err := field.Scalar()
			if err != nil {
				return nil, err
			}
			n.dataType = sc.Copy()

			return field, nil
		}
		n.Parent = field.Parent
		scalar, err := field.Scalar()
		if err != nil {
			return nil, err
		}
		n.dataType = scalar.Copy()

		return field, err
	case *AggregateFunctionCall:
		fn, ok := parse.Functions[n.FunctionName]
		if !ok {
			// should get caught during parsing and/or planning phase,
			// but just in case
			return nil, fmt.Errorf(`function "%s" not found`, n.FunctionName)
		}

		args, err := s.manyAreScalar(n.Args, currentRel)
		if err != nil {
			return nil, err
		}

		returnType, err := fn.ValidateArgs(args)
		if err != nil {
			return nil, err
		}

		n.returnType = returnType.Copy()

		return &Field{
			Name: n.FunctionName,
			val:  returnType,
		}, nil
	case *ScalarFunctionCall:
		fn, ok := parse.Functions[n.FunctionName]
		if !ok {
			// should get caught during parsing and/or planning phase,
			return nil, fmt.Errorf(`function "%s" not found`, n.FunctionName)
		}

		args, err := s.manyAreScalar(n.Args, currentRel)
		if err != nil {
			return nil, err
		}

		returnType, err := fn.ValidateArgs(args)
		if err != nil {
			return nil, err
		}

		n.returnType = returnType.Copy()

		return &Field{
			Name: n.FunctionName,
			val:  returnType,
		}, nil
	case *ProcedureCall:
		var neededArgs []*types.DataType
		var returns *types.ProcedureReturn
		if n.Foreign {
			foreignProc, ok := s.plan.Schema.FindForeignProcedure(n.ProcedureName)
			if !ok {
				return nil, fmt.Errorf(`foreign procedure "%s" not found`, n.ProcedureName)
			}
			neededArgs = foreignProc.Parameters
			returns = foreignProc.Returns

			// if it is foreign, there must be two contextual arguments, both evaluating to strings
			if err := s.manyEvalTo(n.ContextArgs, []*types.DataType{types.TextType, types.TextType}, currentRel); err != nil {
				return nil, err
			}
		} else {
			proc, ok := s.plan.Schema.FindProcedure(n.ProcedureName)
			if !ok {
				return nil, fmt.Errorf(`procedure "%s" not found`, n.ProcedureName)
			}
			for _, arg := range proc.Parameters {
				neededArgs = append(neededArgs, arg.Type.Copy())
			}
			returns = proc.Returns
		}

		if returns == nil {
			return nil, fmt.Errorf(`procedure "%s" does not return anything`, n.ProcedureName)
		}
		if returns.IsTable {
			return nil, fmt.Errorf(`procedure "%s" returns a table, use a procedure scan instead`, n.ProcedureName)
		}
		if len(returns.Fields) != 1 {
			return nil, fmt.Errorf(`procedure "%s" must return exactly one column`, n.ProcedureName)
		}

		if err := s.manyEvalTo(n.Args, neededArgs, currentRel); err != nil {
			return nil, err
		}

		n.returnType = returns.Fields[0].Type.Copy()

		return &Field{
			Name: n.ProcedureName,
			val:  returns.Fields[0].Type.Copy(),
		}, nil
	case *ArithmeticOp:
		left, err := s.isScalar(n.Left, currentRel)
		if err != nil {
			return nil, err
		}

		right, err := s.isScalar(n.Right, currentRel)
		if err != nil {
			return nil, err
		}

		if !left.IsNumeric() {
			return nil, fmt.Errorf("arithmetic operation requires numeric types, got %s", left)
		}

		if !left.Equals(right) {
			return nil, fmt.Errorf("arithmetic operation requires the same data types, got %s and %s", left, right)
		}

		return anonField(left), nil
	case *ComparisonOp:
		left, err := s.isScalar(n.Left, currentRel)
		if err != nil {
			return nil, err
		}

		right, err := s.isScalar(n.Right, currentRel)
		if err != nil {
			return nil, err
		}

		if !left.Equals(right) {
			return nil, fmt.Errorf("comparison operation requires the same data types, got %s and %s", left, right)
		}

		return anonField(types.BoolType), nil
	case *LogicalOp:
		if err := s.manyEvalTo([]LogicalExpr{n.Left, n.Right}, []*types.DataType{types.BoolType, types.BoolType}, currentRel); err != nil {
			return nil, err
		}

		return anonField(types.BoolType), nil
	case *UnaryOp:
		dt, err := s.isScalar(n.Expr, currentRel)
		if err != nil {
			return nil, err
		}

		switch n.Op {
		case Negate:
			if !dt.IsNumeric() {
				return nil, fmt.Errorf("negation requires a numeric type, got %s", dt)
			}
			if dt.Equals(types.Uint256Type) {
				return nil, fmt.Errorf("negation is not supported for type %s", dt)
			}
		case Not:
			if !dt.Equals(types.BoolType) {
				return nil, fmt.Errorf("logical negation requires a boolean type, got %s", dt)
			}
		case Positive:
			if !dt.IsNumeric() {
				return nil, fmt.Errorf("positive sign requires a numeric type, got %s", dt)
			}
		}

		return anonField(dt), nil
	case *TypeCast:
		_, err := s.isScalar(n.Expr, currentRel)
		if err != nil {
			return nil, err
		}

		// we can provide further validation with logic on what types
		// can be casted to what, but for now we assume any cast is valid

		return anonField(n.Type), nil
	case *AliasExpr:
		field, err := s.evalExpression(n.Expr, currentRel)
		if err != nil {
			return nil, err
		}

		field.Parent = ""
		field.Name = n.Alias
		return field, nil
	case *ArrayAccess:
		dt, err := s.isScalar(n.Array, currentRel)
		if err != nil {
			return nil, err
		}

		if !dt.IsArray {
			return nil, fmt.Errorf("cannot access array elements of non-array type %s", dt.String())
		}

		if err := s.evalsTo(n.Index, types.IntType, currentRel); err != nil {
			return nil, err
		}

		dt2 := dt.Copy()
		dt2.IsArray = false
		return anonField(dt2), nil
	case *ArrayConstructor:
		if len(n.Elements) == 0 {
			return nil, fmt.Errorf("array constructor must have at least one element")
		}

		var dt *types.DataType
		for _, expr := range n.Elements {
			field, err := s.isScalar(expr, currentRel)
			if err != nil {
				return nil, err
			}

			if dt == nil {
				dt = field
			} else {
				if !dt.Equals(field) {
					return nil, fmt.Errorf("all elements in array constructor must be of the same type")
				}
			}
		}

		return anonField(types.ArrayType(dt)), nil
	case *FieldAccess:
		field, err := s.evalExpression(n.Object, currentRel)
		if err != nil {
			return nil, err
		}

		obj, err := field.Object()
		if err != nil {
			return nil, err
		}

		objField, ok := obj[n.Key]
		if !ok {
			return nil, fmt.Errorf(`field "%s" not found in object`, n.Key)
		}

		return anonField(objField), nil
	case *SubqueryExpr:
		if n.Query.ReturnsRelation {
			panic("internal bug: planner planned a table subquery in an expression")
		}

		rel, err := s.evalSubquery(n.Query, currentRel)
		if err != nil {
			return nil, err
		}

		// subquery must return exactly one column

		if len(rel.Fields) != 1 {
			return nil, fmt.Errorf("subquery must return exactly one column, got %d", len(rel.Fields))
		}

		_, err = rel.Fields[0].Scalar()
		if err != nil {
			return nil, err
		}

		if n.Exists {
			return anonField(types.BoolType), nil
		}

		return rel.Fields[0], nil
	case *Collate:
		field, err := s.evalExpression(n.Expr, currentRel)
		if err != nil {
			return nil, err
		}

		scalar, err := field.Scalar()
		if err != nil {
			return nil, err
		}

		switch n.Collation {
		default:
			panic(fmt.Sprintf("unexpected collation %s", n.Collation))
		case NoCaseCollation:
			if !scalar.Equals(types.TextType) {
				return nil, fmt.Errorf("collation requires a text type, got %s", field)
			}
		}

		// return the field b/c if you have
		// "SELECT name COLLATE NOCASE FROM users",
		// Postgres will return column "name"
		return field, nil
	case *IsIn:
		left, err := s.isScalar(n.Left, currentRel)
		if err != nil {
			return nil, err
		}

		if n.Subquery != nil {
			right, err := s.isScalar(n.Subquery, currentRel)
			if err != nil {
				return nil, err
			}

			if !left.Equals(right) {
				return nil, fmt.Errorf("is in requires the same data types, got %s and %s", left, right)
			}
		} else {
			for _, expr := range n.Expressions {
				right, err := s.isScalar(expr, currentRel)
				if err != nil {
					return nil, err
				}

				if !left.Equals(right) {
					return nil, fmt.Errorf("is in requires the same data types, got %s and %s", left, right)
				}
			}
		}

		return anonField(types.BoolType), nil
	case *Case:
		// expectedWhenType is the type that we want every
		// WHEN to evaluate to. This is set by SELECT [expr] CASE ...
		// if there is no expr, then it expects a boolean type
		expectedWhenType := types.BoolType
		if n.Value != nil {
			var err error
			expectedWhenType, err = s.isScalar(n.Value, currentRel)
			if err != nil {
				return nil, err
			}
		}

		// the type that will be returned by the CASE statement
		var returnType *types.DataType

		for _, when := range n.WhenClauses {
			if err := s.evalsTo(when[0], expectedWhenType, currentRel); err != nil {
				return nil, err
			}

			then, err := s.isScalar(when[1], currentRel)
			if err != nil {
				return nil, err
			}

			if returnType == nil {
				returnType = then
			} else {
				if !returnType.Equals(then) {
					return nil, fmt.Errorf("all THEN clauses in CASE must be of the same type")
				}
			}
		}

		if n.Else != nil {
			elseType, err := s.isScalar(n.Else, currentRel)
			if err != nil {
				return nil, err
			}

			// I dont think returnType can be nil here because
			// "SELECT CASE ELSE 1 END" is invalid,
			// but just in case
			if returnType != nil {
				returnType = elseType
			} else {
				if !returnType.Equals(elseType) {
					return nil, fmt.Errorf("ELSE clause in CASE must be of the same type as the THEN clauses")
				}
			}
		}

		// also don't think this can be nil, but just in case
		if returnType == nil {
			return nil, fmt.Errorf("CASE must have at least one THEN clause")
		}

		return anonField(returnType), nil
	}

	return nil, fmt.Errorf("unexpected node type %T", expr)
}

// evalSubquery evaluates a subquery and returns the relation that the subquery
// represents. It will perform type validations. It takes the relation of the calling
// query to allow for correlated subqueries.
func (s *EvaluateContext) evalSubquery(sub *Subquery, currentRel *Relation) (*Relation, error) {
	// for a subquery, we will add the current relation to the outer relation,
	// to allow for correlated subqueries
	oldOuter := s.OuterRelation
	oldCorrelations := s.Correlations

	s.OuterRelation = &Relation{
		Fields: append(s.OuterRelation.Fields, currentRel.Fields...),
	}
	// we don't need access to the old correlations since we will simply
	// recognize them as correlated again if they are used in the subquery
	s.Correlations = []*ColumnRef{}

	defer func() {
		s.OuterRelation = oldOuter
		s.Correlations = oldCorrelations
	}()

	rel, err := s.evalRelation(sub.Plan)
	if err != nil {
		return nil, err
	}

	// for all new correlations, we need to check if they are present on
	// the oldOuter relation. If so, then we simply add them as correlated
	// to the subplan. If not, then we also need to pass them back to the
	// oldCorrelations so that they can be used in the outer query (in the case
	// of a multi-level correlated subquery)
	oldMap := make(map[[2]string]struct{})
	for _, cor := range oldCorrelations {
		oldMap[[2]string{cor.Parent, cor.ColumnName}] = struct{}{}
	}

	for _, cor := range s.Correlations {
		_, err = currentRel.Search(cor.Parent, cor.ColumnName)
		// if the column is not found in the current relation, then we need to
		// pass it back to the oldCorrelations
		if errors.Is(err, errColumnNotFound) {
			// if not known to the outer correlation, then add it
			_, ok := oldMap[[2]string{cor.Parent, cor.ColumnName}]
			if !ok {
				oldCorrelations = append(oldCorrelations, cor)
				continue
			}
		} else if err != nil {
			// some other error occurred
			return nil, err
		}
		// if no error, it is correlated to this query, do nothing
	}
	sub.Correlated = s.Correlations

	return rel, nil
}

/*
	Helpers for removing duplicate code
*/

// manyEvalTo is a helper method that checks if a slice of LogicalExprs will evaluate
// to the slice of data types.
func (s *EvaluateContext) manyEvalTo(exprs []LogicalExpr, types []*types.DataType, currentRel *Relation) error {
	if len(exprs) != len(types) {
		return fmt.Errorf("expected %d expressions, got %d", len(types), len(exprs))
	}

	for i, expr := range exprs {
		if err := s.evalsTo(expr, types[i], currentRel); err != nil {
			return err
		}
	}

	return nil
}

// evalsTo is a helper method that checks if a LogicalExpr will evaluate to the
// given data type.
func (s *EvaluateContext) evalsTo(expr LogicalExpr, dt *types.DataType, currentRel *Relation) error {
	field, err := s.evalExpression(expr, currentRel)
	if err != nil {
		return err
	}

	scalar, err := field.Scalar()
	if err != nil {
		return err
	}

	if !scalar.Equals(dt) {
		return fmt.Errorf("expected expression to be of type %s, got %s", dt, scalar)
	}

	return nil
}

// manyAreScalar checks that many LogicalExprs are scalar.
func (s *EvaluateContext) manyAreScalar(exprs []LogicalExpr, currentRel *Relation) ([]*types.DataType, error) {
	var vals []*types.DataType
	for _, expr := range exprs {
		scalar, err := s.isScalar(expr, currentRel)
		if err != nil {
			return nil, err
		}

		vals = append(vals, scalar)
	}

	return vals, nil
}

// isScalar checks that a LogicalExpr is scalar.
func (s *EvaluateContext) isScalar(expr LogicalExpr, currentRel *Relation) (*types.DataType, error) {
	field, err := s.evalExpression(expr, currentRel)
	if err != nil {
		return nil, err
	}

	return field.Scalar()
}

// Relation is the current relation in the query plan.
type Relation struct {
	Fields []*Field
}

func (r *Relation) Copy() *Relation {
	var fields []*Field
	for _, f := range r.Fields {
		var val any
		switch v := f.val.(type) {
		case *types.DataType:
			val = v.Copy()
		case map[string]*types.DataType:
			val = make(map[string]*types.DataType)
			for k, v := range v {
				val.(map[string]*types.DataType)[k] = v.Copy()
			}
		}

		fields = append(fields, &Field{
			Parent: f.Parent,
			Name:   f.Name,
			val:    val,
		})
	}
	return &Relation{
		Fields: fields,
	}
}

func (s *Relation) ColumnsByParent(name string) []*Field {
	var columns []*Field
	for _, c := range s.Fields {
		if c.Parent == name {
			columns = append(columns, c)
		}
	}
	return columns
}

var errColumnNotFound = fmt.Errorf("column not found")

// Search searches for a column by parent and name.
// If the column is not found, an error is returned.
// If no parent is specified and many columns have the same name,
// an error is returned.
func (s *Relation) Search(parent, name string) (*Field, error) {
	if parent == "" {
		var column *Field
		count := 0
		for _, c := range s.Fields {
			if c.Name == name {
				column = c
				count++
			}
		}
		if count == 0 {
			return nil, fmt.Errorf(`%w: "%s"`, errColumnNotFound, name)
		}
		if count > 1 {
			return nil, fmt.Errorf(`column "%s" is ambiguous`, name)
		}

		// return a new instance since we are qualifying the column
		return &Field{
			Parent: column.Parent, // fully qualify the column
			Name:   column.Name,
			val:    column.val,
		}, nil
	}

	for _, c := range s.Fields {
		if c.Parent == parent && c.Name == name {
			// shallow copy
			return &Field{
				Parent: parent,
				Name:   name,
				val:    c.val,
			}, nil
		}
	}

	return nil, fmt.Errorf(`%w: "%s.%s"`, errColumnNotFound, parent, name)
}

func relationFromTable(tbl *types.Table) *Relation {
	s := &Relation{}
	for _, col := range tbl.Columns {
		s.Fields = append(s.Fields, &Field{
			Parent: tbl.Name,
			Name:   col.Name,
			val:    col.Type.Copy(),
		})
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
// is a constant. If this is the last expression in a relation,
// the "Name" field will be the name of the column in the result.
type Field struct {
	// TODO: idk if parent is needed
	Parent string // the parent relation name
	Name   string // the field name
	// val is the value of the field.
	// it can be either a single value or a map of values,
	// depending on the field type.
	// This value should be accessed using the Scalar() or Object()
	val any
}

func anonField(dt *types.DataType) *Field {
	return &Field{
		val: dt,
	}
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
