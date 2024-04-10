package typing

import (
	"fmt"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/internal/engine"
	"github.com/kwilteam/kwil-db/internal/parse/sql/tree"
)

type typeVisitor struct {
	*tree.BaseAstVisitor
	// commonTables are tables that are globally available.
	// these are either tables that have been defined in the schema,
	// or common table expressions.
	// they are defined at the beginning of the query, and do not
	// change.
	commonTables map[string]*engine.Relation
	// ctes is a set of common table expressions that have been defined.
	// all of the keys can be found in commonTables.
	ctes map[string]struct{}
	// bindParams are the parameters that are available in the query
	// we know the types from the kuneiform schema
	bindParams map[string]*types.DataType
	// arbitraryBinds is a flag that allows us to ignore bind parameters
	// when type checking
	arbitraryBinds bool
}

var _ tree.AstVisitor = &typeVisitor{}

// evalFunc is a function that allows modifying an evaluation context.
type evalFunc func(e *evaluationContext) error

// BEGIN evalFunc

func (t *typeVisitor) VisitCTE(p0 *tree.CTE) any {
	return evalFunc(func(e *evaluationContext) error {
		relation, err := p0.Select.Accept(t).(returnFunc)(e)
		if err != nil {
			return err
		}

		_, ok := t.commonTables[p0.Table]
		if ok {
			return fmt.Errorf("common table expression conflicts with existing table %s", p0.Table)
		}

		t.commonTables[p0.Table] = relation
		t.ctes[p0.Table] = struct{}{}

		return nil
	})
}

func (t *typeVisitor) VisitRelationJoin(p0 *tree.RelationJoin) any {
	return evalFunc(func(e *evaluationContext) error {
		err := p0.Relation.Accept(t).(evalFunc)(e)
		if err != nil {
			return err
		}

		for _, join := range p0.Joins {
			err = join.Accept(t).(evalFunc)(e)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

func (t *typeVisitor) VisitRelationSubquery(p0 *tree.RelationSubquery) any {
	return evalFunc(func(e *evaluationContext) error {
		r, err := p0.Select.Accept(t).(returnFunc)(e)
		if err != nil {
			return err
		}

		return e.join(&engine.QualifiedRelation{
			Name:     p0.Alias, // this can be ""
			Relation: r,
		})
	})
}

func (t *typeVisitor) VisitRelationTable(p0 *tree.RelationTable) any {
	return evalFunc(func(e *evaluationContext) error {
		tbl, ok := t.commonTables[p0.Name]
		if !ok {
			return fmt.Errorf("table %s not found", p0.Name)
		}

		name := p0.Name
		if p0.Alias != "" {
			name = p0.Alias
		}

		return e.join(&engine.QualifiedRelation{
			Name:     name,
			Relation: tbl,
		})
	})
}

// the rest of the evalFunc visitors do not actually modify the evaluation context

func (t *typeVisitor) VisitUpsert(p0 *tree.Upsert) any {
	return evalFunc(func(e *evaluationContext) error {
		if p0.ConflictTarget != nil {
			err := p0.ConflictTarget.Accept(t).(evalFunc)(e)
			if err != nil {
				return err
			}
		}

		for _, set := range p0.Updates {
			err := set.Accept(t).(evalFunc)(e)
			if err != nil {
				return err
			}
		}

		if p0.Where != nil {
			attr, err := p0.Where.Accept(t).(attributeFn)(e)
			if err != nil {
				return err
			}

			if !attr.Type.Equals(types.BoolType) {
				return fmt.Errorf("%w: where clause must be boolean. Received: %s", ErrInvalidType, attr.Type.String())
			}
		}

		return nil
	})
}

func (t *typeVisitor) VisitUpdateSetClause(p0 *tree.UpdateSetClause) any {
	return evalFunc(func(e *evaluationContext) error {
		// check that the columns exist
		// we can only update columns in the first table
		if len(e.joinOrder) == 0 {
			return fmt.Errorf("no table to update")
		}
		for _, col := range p0.Columns {
			_, err := e.findColumn(e.joinOrder[0], col)
			if err != nil {
				return err
			}
		}

		if p0.Expression != nil {
			_, err := p0.Expression.Accept(t).(attributeFn)(e)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

func (t *typeVisitor) VisitConflictTarget(p0 *tree.ConflictTarget) any {
	return evalFunc(func(e *evaluationContext) error {
		// check that the columns exist
		if len(e.joinOrder) == 0 {
			return fmt.Errorf("no table to update")
		}
		for _, col := range p0.IndexedColumns {
			_, err := e.findColumn(e.joinOrder[0], col)
			if err != nil {
				return err
			}
		}

		if p0.Where != nil {
			attr, err := p0.Where.Accept(t).(attributeFn)(e)
			if err != nil {
				return err
			}

			if !attr.Type.Equals(types.BoolType) {
				return fmt.Errorf("%w: where clause must be boolean. Received: %s", ErrInvalidType, attr.Type.String())
			}
		}

		return nil
	})
}

func (t *typeVisitor) VisitLimit(p0 *tree.Limit) any {
	return evalFunc(func(e *evaluationContext) error {
		if p0.Expression != nil {
			limit, err := p0.Expression.Accept(t).(attributeFn)(e)
			if err != nil {
				return err
			}

			if !limit.Type.Equals(types.IntType) {
				return fmt.Errorf("%w: limit must be an integer. Received: %s", ErrInvalidType, limit.Type.String())
			}
		}

		if p0.Offset != nil {
			offset, err := p0.Offset.Accept(t).(attributeFn)(e)
			if err != nil {
				return err
			}

			if !offset.Type.Equals(types.IntType) {
				return fmt.Errorf("%w: offset must be an integer. Received: %s", ErrInvalidType, offset.Type.String())
			}
		}

		return nil
	})
}

func (t *typeVisitor) VisitOrderBy(p0 *tree.OrderBy) any {
	return evalFunc(func(e *evaluationContext) error {
		for _, term := range p0.OrderingTerms {
			err := term.Accept(t).(evalFunc)(e)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

func (t *typeVisitor) VisitOrderingTerm(p0 *tree.OrderingTerm) any {
	return evalFunc(func(e *evaluationContext) error {
		if p0.Expression == nil {
			return nil // not sure if this is possible, don't believe it is
		}
		_, err := p0.Expression.Accept(t).(attributeFn)(e)
		return err
	})
}

func (t *typeVisitor) VisitGroupBy(p0 *tree.GroupBy) any {
	return evalFunc(func(e *evaluationContext) error {
		for _, col := range p0.Expressions {
			_, err := col.Accept(t).(attributeFn)(e)
			if err != nil {
				return err
			}
		}

		if p0.Having != nil {
			attr, err := p0.Having.Accept(t).(attributeFn)(e)
			if err != nil {
				return err
			}

			if !attr.Type.Equals(types.BoolType) {
				return fmt.Errorf("%w: having clause must be boolean. Received: %s", ErrInvalidType, attr.Type.String())
			}
		}

		return nil
	})
}

func (t *typeVisitor) VisitJoinPredicate(p0 *tree.JoinPredicate) any {
	return evalFunc(func(e *evaluationContext) error {
		err := p0.Table.Accept(t).(evalFunc)(e)
		if err != nil {
			return err
		}

		if p0.Constraint != nil {
			r, err := p0.Constraint.Accept(t).(attributeFn)(e)
			if err != nil {
				return err
			}

			if !r.Type.Equals(types.BoolType) {
				return fmt.Errorf("%w: join constraint must be boolean. Received: %s", ErrInvalidType, r.Type.String())
			}
		}

		return nil
	})
}

// END evalFunc

// attributeFn is returned from all visitor expressions.
// It allows us to evaluate return attributes once we
// have more context.
// The attribute name can be blank, and will only be set
// if the expression is a column.
type attributeFn func(ev *evaluationContext) (*engine.QualifiedAttribute, error)

// BEGIN attributeFn

func (t *typeVisitor) VisitExpressionArithmetic(p0 *tree.ExpressionArithmetic) any {
	return attributeFn(func(ev *evaluationContext) (*engine.QualifiedAttribute, error) {
		a := p0.Left.Accept(t).(attributeFn)
		b := p0.Right.Accept(t).(attributeFn)

		at, err := a(ev)
		if err != nil {
			return nil, err
		}
		if !at.Type.Equals(types.IntType) {
			return nil, fmt.Errorf("%w: arithmetic expression expected int. Received: %s", ErrInvalidType, at.Type.String())
		}

		bt, err := b(ev)
		if err != nil {
			return nil, err
		}
		if !bt.Type.Equals(types.IntType) {
			return nil, fmt.Errorf("%w: arithmetic expression expected int. Received: %s", ErrInvalidType, bt.Type.String())
		}

		if p0.TypeCast != nil {
			return anonAttr(p0.TypeCast), nil
		}

		return anonAttr(types.IntType), nil
	})
}

func (t *typeVisitor) VisitExpressionBetween(p0 *tree.ExpressionBetween) any {
	return attributeFn(func(ev *evaluationContext) (*engine.QualifiedAttribute, error) {
		e := p0.Expression.Accept(t).(attributeFn)
		l := p0.Left.Accept(t).(attributeFn)
		r := p0.Right.Accept(t).(attributeFn)

		et, err := e(ev)
		if err != nil {
			return nil, err
		}

		lt, err := l(ev)
		if err != nil {
			return nil, err
		}

		rt, err := r(ev)
		if err != nil {
			return nil, err
		}

		if !et.Type.Equals(lt.Type) {
			return nil, fmt.Errorf("%w: between expression expected %s. Received: %s", ErrInvalidType, et.Type.Name, lt.Type.String())
		}

		if !et.Type.Equals(rt.Type) {
			return nil, fmt.Errorf("%w: between expression expected %s. Received: %s", ErrInvalidType, et.Type.Name, rt.Type.String())
		}

		if p0.TypeCast != nil {
			return anonAttr(p0.TypeCast), nil
		}

		return anonAttr(types.BoolType), nil
	})
}

func (t *typeVisitor) VisitExpressionBinaryComparison(p0 *tree.ExpressionBinaryComparison) any {
	return attributeFn(func(ev *evaluationContext) (*engine.QualifiedAttribute, error) {
		a := p0.Left.Accept(t).(attributeFn)
		b := p0.Right.Accept(t).(attributeFn)

		at, err := a(ev)
		if err != nil {
			return nil, err
		}
		bt, err := b(ev)
		if err != nil {
			return nil, err
		}

		if !at.Type.Equals(bt.Type) {
			return nil, fmt.Errorf("%w: comparison expression expected %s. Received: %s", ErrInvalidType, at.Type.String(), bt.Type.String())
		}

		if p0.TypeCast != nil {
			return anonAttr(p0.TypeCast), nil
		}

		return anonAttr(types.BoolType), nil
	})
}

func (t *typeVisitor) VisitExpressionBindParameter(p0 *tree.ExpressionBindParameter) any {
	return attributeFn(func(ev *evaluationContext) (*engine.QualifiedAttribute, error) {
		c, ok := t.bindParams[p0.Parameter]
		if !ok {
			if t.arbitraryBinds {
				c = types.UnknownType
			} else {
				return nil, fmt.Errorf("bind parameter %s not found", p0.Parameter)
			}
		}

		if p0.TypeCast != nil {
			return anonAttr(p0.TypeCast), nil
		}

		return anonAttr(c), nil
	})
}

func (t *typeVisitor) VisitExpressionCase(p0 *tree.ExpressionCase) any {
	return attributeFn(func(ev *evaluationContext) (*engine.QualifiedAttribute, error) {
		if p0.CaseExpression != nil {
			c := p0.CaseExpression.Accept(t).(attributeFn)
			ct, err := c(ev)
			if err != nil {
				return nil, err
			}
			if !ct.Type.Equals(types.BoolType) {
				return nil, fmt.Errorf("%w: expected bool. Received: %s", ErrInvalidType, ct.Type.String())
			}
		}

		var neededType *types.DataType

		for _, w := range p0.WhenThenPairs {
			when := w[0].Accept(t).(attributeFn)
			whenType, err := when(ev)
			if err != nil {
				return nil, err
			}
			if !whenType.Type.Equals(types.BoolType) {
				return nil, fmt.Errorf("%w: expected bool. Received %s", ErrInvalidType, whenType.Type.String())
			}

			then := w[1].Accept(t).(attributeFn)
			thenType, err := then(ev)
			if err != nil {
				return nil, err
			}

			if neededType == nil {
				neededType = thenType.Type
			}

			if !neededType.Equals(thenType.Type) {
				return nil, fmt.Errorf("%w: all THEN types must be the same. Received: %s and %s", ErrInvalidType, neededType.String(), thenType.Type.String())
			}
		}

		if p0.ElseExpression != nil {
			e := p0.ElseExpression.Accept(t).(attributeFn)
			eType, err := e(ev)
			if err != nil {
				return nil, err
			}

			if !neededType.Equals(eType.Type) {
				return nil, fmt.Errorf("%w: ELSE type must match THEN type. Received: %s and %s", ErrInvalidType, neededType.String(), eType.Type.String())
			}
		}

		if p0.TypeCast != nil {
			return anonAttr(p0.TypeCast), nil
		}

		return anonAttr(neededType), nil
	})
}

func (t *typeVisitor) VisitExpressionCollate(p0 *tree.ExpressionCollate) any {
	return attributeFn(func(ev *evaluationContext) (*engine.QualifiedAttribute, error) {
		rel, err := p0.Expression.Accept(t).(attributeFn)(ev)
		if err != nil {
			return nil, err
		}

		if p0.TypeCast != nil {
			return anonAttr(p0.TypeCast), nil
		}

		return rel, nil
	})
}

func (t *typeVisitor) VisitExpressionColumn(p0 *tree.ExpressionColumn) any {
	return attributeFn(func(ev *evaluationContext) (*engine.QualifiedAttribute, error) {
		col, err := ev.findColumn(p0.Table, p0.Column)
		if err != nil {
			return nil, err
		}

		if p0.TypeCast != nil {
			return &engine.QualifiedAttribute{
				Name: p0.Column,
				Attribute: &engine.Attribute{
					Type: p0.TypeCast,
				},
			}, nil
		}

		return col, nil
	})
}

func (t *typeVisitor) VisitExpressionFunction(p0 *tree.ExpressionFunction) any {
	return attributeFn(func(ev *evaluationContext) (*engine.QualifiedAttribute, error) {
		funcDef, ok := engine.Functions[p0.Function]
		if !ok {
			return nil, fmt.Errorf("function %s not found", p0.Function)
		}

		var argTypes []*types.DataType
		for _, arg := range p0.Inputs {
			attr, err := arg.Accept(t).(attributeFn)(ev)
			if err != nil {
				return nil, err
			}

			argTypes = append(argTypes, attr.Type)
		}

		returnType, err := funcDef.Args(argTypes)
		if err != nil {
			return nil, err
		}

		if p0.TypeCast != nil {
			return anonAttr(p0.TypeCast), nil
		}

		return anonAttr(returnType), nil
	})
}

func (t *typeVisitor) VisitExpressionIs(p0 *tree.ExpressionIs) any {
	return attributeFn(func(ev *evaluationContext) (*engine.QualifiedAttribute, error) {
		l := p0.Left.Accept(t).(attributeFn)
		r := p0.Right.Accept(t).(attributeFn)

		lt, err := l(ev)
		if err != nil {
			return nil, err
		}

		rt, err := r(ev)
		if err != nil {
			return nil, err
		}

		if !lt.Type.Equals(rt.Type) && !lt.Type.Equals(types.NullType) && !rt.Type.Equals(types.NullType) {
			return nil, fmt.Errorf("%w: comparing different types: %s and %s", ErrInvalidType, lt.Type.String(), rt.Type.String())
		}

		if p0.TypeCast != nil {
			return anonAttr(p0.TypeCast), nil
		}

		return anonAttr(types.BoolType), nil
	})
}

func (t *typeVisitor) VisitExpressionList(p0 *tree.ExpressionList) any {
	return attributeFn(func(ev *evaluationContext) (*engine.QualifiedAttribute, error) {
		var lastType *types.DataType
		for _, e := range p0.Expressions {
			et := e.Accept(t).(attributeFn)
			etType, err := et(ev)
			if err != nil {
				return nil, err
			}

			if lastType == nil {
				lastType = etType.Type
				continue
			}

			if !lastType.Equals(etType.Type) {
				return nil, fmt.Errorf("%w: cannot assign type %s to expression list of type %s", ErrInvalidType, etType.Type.String(), lastType.String())
			}
		}

		if p0.TypeCast != nil {
			return anonAttr(p0.TypeCast), nil
		}

		return anonAttr(lastType), nil
	})
}

func (t *typeVisitor) VisitExpressionTextLiteral(p0 *tree.ExpressionTextLiteral) any {
	return attributeFn(func(ev *evaluationContext) (*engine.QualifiedAttribute, error) {
		if p0.TypeCast != nil {
			return anonAttr(p0.TypeCast), nil
		}

		return anonAttr(types.TextType), nil
	})
}

func (t *typeVisitor) VisitExpressionNumericLiteral(p0 *tree.ExpressionNumericLiteral) any {
	return attributeFn(func(ev *evaluationContext) (*engine.QualifiedAttribute, error) {
		if p0.TypeCast != nil {
			return anonAttr(p0.TypeCast), nil
		}

		return anonAttr(types.IntType), nil
	})
}

func (t *typeVisitor) VisitExpressionBooleanLiteral(p0 *tree.ExpressionBooleanLiteral) any {
	return attributeFn(func(ev *evaluationContext) (*engine.QualifiedAttribute, error) {
		if p0.TypeCast != nil {
			return anonAttr(p0.TypeCast), nil
		}

		return anonAttr(types.BoolType), nil
	})
}

func (t *typeVisitor) VisitExpressionNullLiteral(p0 *tree.ExpressionNullLiteral) any {
	return attributeFn(func(ev *evaluationContext) (*engine.QualifiedAttribute, error) {
		if p0.TypeCast != nil {
			return anonAttr(p0.TypeCast), nil
		}

		return anonAttr(types.NullType), nil
	})
}

func (t *typeVisitor) VisitExpressionBlobLiteral(p0 *tree.ExpressionBlobLiteral) any {
	return attributeFn(func(ev *evaluationContext) (*engine.QualifiedAttribute, error) {
		if p0.TypeCast != nil {
			return anonAttr(p0.TypeCast), nil
		}

		return anonAttr(types.BlobType), nil
	})
}

func (t *typeVisitor) VisitExpressionSelect(p0 *tree.ExpressionSelect) any {
	return attributeFn(func(ev *evaluationContext) (*engine.QualifiedAttribute, error) {
		r, err := p0.Select.Accept(t).(returnFunc)(ev)
		if err != nil {
			return nil, err
		}

		shape := r.Shape()
		if len(shape) != 1 && !p0.IsExists {
			return nil, fmt.Errorf("subquery must return exactly one column")
		}

		if p0.TypeCast != nil {
			return anonAttr(p0.TypeCast), nil
		}

		if p0.IsExists {
			return anonAttr(types.BoolType), nil
		}

		return anonAttr(shape[0]), nil
	})
}

func (t *typeVisitor) VisitExpressionStringCompare(p0 *tree.ExpressionStringCompare) any {
	return attributeFn(func(ev *evaluationContext) (*engine.QualifiedAttribute, error) {
		a := p0.Left.Accept(t).(attributeFn)
		b := p0.Right.Accept(t).(attributeFn)

		// do these both need to be text? I believe so
		at, err := a(ev)
		if err != nil {
			return nil, err
		}
		bt, err := b(ev)
		if err != nil {
			return nil, err
		}
		if !at.Type.Equals(bt.Type) {
			return nil, fmt.Errorf("%w: string comparison expression expected %s. Received: %s", ErrInvalidType, at.Type.String(), bt.Type.String())
		}

		if p0.Escape != nil {
			esc := p0.Escape.Accept(t).(attributeFn)
			et, err := esc(ev)
			if err != nil {
				return nil, err
			}

			if !et.Type.Equals(types.TextType) {
				return nil, fmt.Errorf("%w: string comparison expected text. Received: %s", ErrInvalidType, et.Type.String())
			}
		}

		if p0.TypeCast != nil {
			return anonAttr(p0.TypeCast), nil
		}

		return anonAttr(types.BoolType), nil
	})
}

func (t *typeVisitor) VisitExpressionUnary(p0 *tree.ExpressionUnary) any {
	return attributeFn(func(ev *evaluationContext) (*engine.QualifiedAttribute, error) {
		o := p0.Operand.Accept(t).(attributeFn)
		ot, err := o(ev)
		if err != nil {
			return nil, err
		}

		if !ot.Type.Equals(types.IntType) {
			return nil, fmt.Errorf("%w: expected int. Received: %s", ErrInvalidType, ot.Type.String())
		}

		if p0.TypeCast != nil {
			return anonAttr(p0.TypeCast), nil
		}

		return anonAttr(types.IntType), nil
	})
}

// anonAttr is a helper function that creates an anonymous attribute
func anonAttr(t *types.DataType) *engine.QualifiedAttribute {
	return &engine.QualifiedAttribute{
		Attribute: &engine.Attribute{
			Type: t,
		},
	}
}

// END attributeFn

// returnFunc if a function that returns a relation.
// it is returned from INSERT, UPDATE, DELETE, and SELECT cores
// and stmts, as well as SimpleSelects.
type returnFunc func(e *evaluationContext) (*engine.Relation, error)

// BEGIN returnFunc

func (t *typeVisitor) VisitInsertStmt(p0 *tree.InsertStmt) any {
	return returnFunc(func(e *evaluationContext) (*engine.Relation, error) {
		for _, cte := range p0.CTE {
			err := cte.Accept(t).(evalFunc)(e)
			if err != nil {
				return nil, err
			}
		}

		return p0.Core.Accept(t).(returnFunc)(e)
	})
}

func (t *typeVisitor) VisitInsertCore(p0 *tree.InsertCore) any {
	return returnFunc(func(e *evaluationContext) (*engine.Relation, error) {
		// we only search the visitor for the table,
		// since contextual table (such as CTEs) cannot be
		// inserted into.
		tbl, ok := t.commonTables[p0.Table]
		if !ok {
			return nil, fmt.Errorf("table %s not found", p0.Table)
		}

		_, ok = t.ctes[p0.Table]
		if ok {
			return nil, fmt.Errorf("cannot insert into common table expression %s", p0.Table)
		}

		// check that the columns exist
		for _, col := range p0.Columns {
			_, ok := tbl.Attribute(col)
			if !ok {
				return nil, fmt.Errorf("column %s not found", col)
			}
		}

		// Postgres has a weird quirk with inserts:
		// tables can be aliased (e.g. insert into foo as bar),
		// but bar cannot be used in a subquery in the insert statement,
		// while foo can. The alias is only useable in the returning clause.
		// Therefore, we will not add the alias to the context.
		for _, row := range p0.Values {
			if len(row) != len(p0.Columns) {
				return nil, fmt.Errorf("mismatched column/value count")
			}

			for i, val := range row {
				attr, err := val.Accept(t).(attributeFn)(e)
				if err != nil {
					return nil, err
				}

				expectedAttr, ok := tbl.Attribute(p0.Columns[i])
				if !ok {
					return nil, fmt.Errorf("unknown column %s", p0.Columns[i])
				}

				if !expectedAttr.Type.Equals(attr.Type) {
					return nil, fmt.Errorf("%w: type mismatch for column %s", ErrInvalidType, p0.Columns[i])
				}
			}
		}

		// common table expressions cannot be returned
		// we want a new context that only has this table

		name := p0.Table
		if p0.TableAlias != "" {
			name = p0.TableAlias
		}

		e2 := e.scope()

		err := e2.join(&engine.QualifiedRelation{
			Name:     name,
			Relation: tbl,
		})
		if err != nil {
			return nil, err
		}

		// similar to values, aliased insert tables cannot be used in the
		// conflict target or set clause. We will not add the alias to the context.
		if p0.Upsert != nil {
			err := p0.Upsert.Accept(t).(evalFunc)(e2)
			if err != nil {
				return nil, err
			}
		}

		// handle returning:

		if p0.ReturningClause == nil {
			return engine.NewRelation(), nil
		}

		result := engine.NewRelation()

		err = p0.ReturningClause.Accept(t).(resultFunc)(e2, result)
		if err != nil {
			return nil, err
		}

		return result, nil
	})
}

func (t *typeVisitor) VisitUpdateStmt(p0 *tree.UpdateStmt) any {
	return returnFunc(func(e *evaluationContext) (*engine.Relation, error) {
		for _, cte := range p0.CTE {
			err := cte.Accept(t).(evalFunc)(e)
			if err != nil {
				return nil, err
			}
		}

		return p0.Core.Accept(t).(returnFunc)(e)
	})
}

func (t *typeVisitor) VisitUpdateCore(p0 *tree.UpdateCore) any {
	return returnFunc(func(e *evaluationContext) (*engine.Relation, error) {
		tbl, ok := t.commonTables[p0.QualifiedTableName.TableName]
		if !ok {
			return nil, fmt.Errorf("unknown table %s", p0.QualifiedTableName.TableName)
		}

		_, ok = t.ctes[p0.QualifiedTableName.TableName]
		if ok {
			return nil, fmt.Errorf("cannot update common table expression %s", p0.QualifiedTableName.TableName)
		}

		name := p0.QualifiedTableName.TableName
		if p0.QualifiedTableName.TableAlias != "" {
			name = p0.QualifiedTableName.TableAlias
		}

		// we now want to update our context with joined relations since they can
		// be accessed in both the set clause and the where clause
		e2 := e.scope()

		err := e2.join(&engine.QualifiedRelation{
			Name:     name,
			Relation: tbl,
		})
		if err != nil {
			return nil, err
		}

		if p0.From != nil {
			err = p0.From.Accept(t).(evalFunc)(e2)
			if err != nil {
				return nil, err
			}

			for _, set := range p0.UpdateSetClause {
				err := set.Accept(t).(evalFunc)(e2)
				if err != nil {
					return nil, err
				}
			}
		}

		if p0.Where != nil {
			whereType, err := p0.Where.Accept(t).(attributeFn)(e2)
			if err != nil {
				return nil, err
			}

			if !whereType.Type.Equals(types.BoolType) {
				return nil, fmt.Errorf("%w: where clause must be boolean. Got %s", ErrInvalidType, whereType.Type.String())
			}
		}

		if p0.Returning == nil {
			return engine.NewRelation(), nil
		}

		result := engine.NewRelation()

		err = p0.Returning.Accept(t).(resultFunc)(e2, result)
		if err != nil {
			return nil, err
		}

		return result, nil
	})
}

func (t *typeVisitor) VisitDeleteStmt(p0 *tree.DeleteStmt) any {
	return returnFunc(func(e *evaluationContext) (*engine.Relation, error) {
		for _, cte := range p0.CTE {
			err := cte.Accept(t).(evalFunc)(e)
			if err != nil {
				return nil, err
			}
		}

		return p0.Core.Accept(t).(returnFunc)(e)
	})
}

func (t *typeVisitor) VisitDeleteCore(p0 *tree.DeleteCore) any {
	return returnFunc(func(e *evaluationContext) (*engine.Relation, error) {
		tbl, ok := t.commonTables[p0.QualifiedTableName.TableName]
		if !ok {
			return nil, fmt.Errorf("unknown table %s", p0.QualifiedTableName.TableName)
		}

		_, ok = t.ctes[p0.QualifiedTableName.TableName]
		if ok {
			return nil, fmt.Errorf("cannot delete from common table expression %s", p0.QualifiedTableName.TableName)
		}

		name := p0.QualifiedTableName.TableName
		if p0.QualifiedTableName.TableAlias != "" {
			name = p0.QualifiedTableName.TableAlias
		}

		// we want to use the new context within the where clause
		// DELETE can use the alias in both the where clause and the returning clause
		e2 := e.scope()
		err := e2.join(&engine.QualifiedRelation{
			Name:     name,
			Relation: tbl,
		})
		if err != nil {
			return nil, err
		}

		if p0.Where != nil {
			whereType, err := p0.Where.Accept(t).(attributeFn)(e2)
			if err != nil {
				return nil, err
			}

			if !whereType.Type.Equals(types.BoolType) {
				return nil, fmt.Errorf("%w, where clause must be boolean. Got %s", ErrInvalidType, whereType.Type.String())
			}
		}

		if p0.Returning == nil {
			return engine.NewRelation(), nil
		}

		result := engine.NewRelation()

		err = p0.Returning.Accept(t).(resultFunc)(e2, result)
		if err != nil {
			return nil, err
		}

		return result, nil
	})
}

func (t *typeVisitor) VisitSelectStmt(p0 *tree.SelectStmt) any {
	return returnFunc(func(e *evaluationContext) (*engine.Relation, error) {
		for _, cte := range p0.CTE {
			err := cte.Accept(t).(evalFunc)(e)
			if err != nil {
				return nil, err
			}
		}

		return p0.Stmt.Accept(t).(returnFunc)(e)
	})
}

func (t *typeVisitor) VisitSelectCore(p0 *tree.SelectCore) any {
	return returnFunc(func(e *evaluationContext) (*engine.Relation, error) {
		// we make a new scope so that we can join the tables
		// without affecting the outer scope
		e2 := e.scope()
		// we need to ensure that the relations all have the same shape
		res, err := p0.SimpleSelects[0].Accept(t).(returnFunc)(e2)
		if err != nil {
			return nil, err
		}

		expectedShape := res.Shape()

		for _, sel := range p0.SimpleSelects[1:] {
			// we create a separate scope for each select.
			selectCtx := e.scope()
			r, err := sel.Accept(t).(returnFunc)(selectCtx)
			if err != nil {
				return nil, err
			}

			shape := r.Shape()

			if len(shape) != len(expectedShape) {
				return nil, fmt.Errorf("%w: compound selects must return the same number of columns. Expected %d. Received: %d", ErrCompoundShape, len(expectedShape), len(shape))
			}

			for i, col := range shape {
				if !col.Equals(expectedShape[i]) {
					return nil, fmt.Errorf("%w: compound selects must return the same types: Expected %s Received: %s", ErrCompoundShape, expectedShape[i].Name, col.Name)
				}
			}
		}

		// if this is a compound select, the joined tables from the selects are not in scope, and
		// we must instead join an anonymous relation that is the compound select.
		// if there is only one select, we can reference the joined tables.
		var e3 *evaluationContext
		if len(p0.SimpleSelects) > 1 {
			// copy in case we are in a correlated subquery
			e3 = e.copy()

			// we need to add the first returned relation anonymously to the context
			// so that we can use it in the ordering and limit
			err = e3.join(&engine.QualifiedRelation{
				Name:     "",
				Relation: res,
			})
			if err != nil {
				return nil, err
			}
		} else {
			e3 = e2
		}

		if p0.OrderBy != nil {
			err := p0.OrderBy.Accept(t).(evalFunc)(e3)
			if err != nil {
				return nil, err
			}
		}

		if p0.Limit != nil {
			err := p0.Limit.Accept(t).(evalFunc)(e3)
			if err != nil {
				return nil, err
			}
		}

		return res, nil
	})
}

func (t *typeVisitor) VisitSimpleSelect(p0 *tree.SimpleSelect) any {
	return returnFunc(func(e *evaluationContext) (*engine.Relation, error) {
		if p0.From != nil {
			// we need to build the evaluation context based on the relation
			err := p0.From.Accept(t).(evalFunc)(e)
			if err != nil {
				return nil, err
			}
		}

		if p0.Where != nil {
			a, err := p0.Where.Accept(t).(attributeFn)(e)
			if err != nil {
				return nil, err
			}
			if !a.Attribute.Type.Equals(types.BoolType) {
				return nil, fmt.Errorf("%w: where clause must be boolean", ErrInvalidType)
			}
		}

		if p0.GroupBy != nil {
			err := p0.GroupBy.Accept(t).(evalFunc)(e)
			if err != nil {
				return nil, err
			}
		}

		// make an empty relation for the result
		result := engine.NewRelation()

		// apply the result columns
		for _, col := range p0.Columns {
			err := col.Accept(t).(resultFunc)(e, result)
			if err != nil {
				return nil, err
			}
		}

		return result, nil
	})
}

// END returnFunc

// resultFunc is a function that allows modifying a relation
// that will be returned by a relationFunc.
// It returned from ResultColumns and Returning Clauses.
// ResultColumns define the return relation from a SELECT,
// and Returning Clauses define the return relation from
// INSERT, UPDATE, and DELETE (if there is a RETURNING clause).
type resultFunc func(e *evaluationContext, r *engine.Relation) error

// BEGIN resultFunc

func (t *typeVisitor) VisitResultColumnExpression(p0 *tree.ResultColumnExpression) any {
	return resultFunc(func(e *evaluationContext, r *engine.Relation) error {
		c := p0.Expression.Accept(t).(attributeFn)
		val, err := c(e)
		if err != nil {
			return err
		}

		if p0.Alias != "" {
			val.Name = p0.Alias
		}

		return r.AddAttribute(val)
	})
}

func (t *typeVisitor) VisitResultColumnStar(p0 *tree.ResultColumnStar) any {
	return resultFunc(func(e *evaluationContext, r *engine.Relation) error {
		return e.Loop(func(_ string, r2 *engine.Relation) error {
			return r.Merge(r2)
		})
	})
}

func (t *typeVisitor) VisitResultColumnTable(p0 *tree.ResultColumnTable) any {
	return resultFunc(func(e *evaluationContext, r *engine.Relation) error {
		tbl, ok := e.joinedTables[p0.TableName]
		if !ok {
			return fmt.Errorf("table %s not found", p0.TableName)
		}

		return r.Merge(tbl)
	})
}

func (t *typeVisitor) VisitReturningClause(p0 *tree.ReturningClause) any {
	return resultFunc(func(e *evaluationContext, r *engine.Relation) error {
		for _, col := range p0.Returned {
			err := col.Accept(t).(resultFunc)(e, r)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (t *typeVisitor) VisitReturningClauseColumn(p0 *tree.ReturningClauseColumn) any {
	return resultFunc(func(e *evaluationContext, r *engine.Relation) error {
		// this can either be return * or return expr

		// case 1: return *, preserving order
		if p0.All {
			return e.Loop(func(_ string, r2 *engine.Relation) error {
				return r.Merge(r2)
			})
		}

		if p0.Expression == nil {
			return fmt.Errorf("invalid returning clause")
		}

		// case 2: return expr
		attribute, err := p0.Expression.Accept(t).(attributeFn)(e)
		if err != nil {
			return err
		}

		// attempt to set the alias
		// if the attribute is not from a column,
		// and there is no alias, this will fail,
		// as the attribute will be anonymous
		// and therefore not accessible in the
		// returned relation.
		if p0.Alias != "" {
			attribute.Name = p0.Alias
		}

		return r.AddAttribute(attribute)
	})
}

// END resultFunc
