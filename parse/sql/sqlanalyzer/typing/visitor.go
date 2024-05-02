package typing

import (
	"fmt"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/parse/metadata"
	"github.com/kwilteam/kwil-db/parse/sql/tree"
	parseTypes "github.com/kwilteam/kwil-db/parse/types"
	"github.com/kwilteam/kwil-db/parse/util"
)

type typeVisitor struct {
	*tree.BaseAstVisitor
	// commonTables are tables that are globally available.
	// these are either tables that have been defined in the schema,
	// or common table expressions.
	// they are defined at the beginning of the query, and do not
	// change.
	commonTables map[string]*Relation
	// ctes is a set of common table expressions that have been defined.
	// all of the keys can be found in commonTables.
	ctes    map[string]struct{}
	options *AnalyzeOptions
}

var _ tree.AstVisitor = &typeVisitor{}

// evalFunc is a function that allows modifying an evaluation context.
type evalFunc func(e *evaluationContext)

// BEGIN evalFunc

func (t *typeVisitor) VisitCTE(p0 *tree.CTE) any {
	return evalFunc(func(e *evaluationContext) {
		relation := p0.Select.Accept(t).(returnFunc)(e)

		_, ok := t.commonTables[p0.Table]
		if ok {
			t.options.ErrorListener.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeSemantic, fmt.Sprintf("common table expression conflicts with existing table %s", p0.Table))
			return
		}

		t.commonTables[p0.Table] = relation
		t.ctes[p0.Table] = struct{}{}
	})
}

func (t *typeVisitor) VisitRelationJoin(p0 *tree.RelationJoin) any {
	return evalFunc(func(e *evaluationContext) {
		p0.Relation.Accept(t).(evalFunc)(e)
		for _, join := range p0.Joins {
			join.Accept(t).(evalFunc)(e)
		}
	})
}

func (t *typeVisitor) VisitRelationSubquery(p0 *tree.RelationSubquery) any {
	return evalFunc(func(e *evaluationContext) {
		r := p0.Select.Accept(t).(returnFunc)(e)

		err := e.join(&QualifiedRelation{
			Name:     p0.Alias, // this can be ""
			Relation: r,
		})
		if err != nil {
			t.options.ErrorListener.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeSemantic, err.Error())
		}
	})
}

func (t *typeVisitor) VisitRelationFunction(p0 *tree.RelationFunction) any {
	// check the function is a procedure that returns a table, and has the same
	// number of inputs as the function has parameters
	return evalFunc(func(e *evaluationContext) {
		// we can ignore the attribute here, we simply want to make sure we visit
		p0.Function.Accept(t).(attributeFn)(e)

		parameters, returns, err := util.FindProcOrForeign(t.options.Schema, p0.Function.Function)
		if err != nil {
			t.options.ErrorListener.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeSemantic, err.Error())
			return
		}

		if len(p0.Function.Inputs) != len(parameters) {
			t.options.ErrorListener.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeSemantic, fmt.Sprintf("procedure %s expected %d inputs, received %d", p0.Function.Function, len(parameters), len(p0.Function.Inputs)))
			return
		}

		for i, in := range p0.Function.Inputs {
			attr := in.Accept(t).(attributeFn)(e)

			if !attr.Type.Equals(parameters[i]) {
				t.options.ErrorListener.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeType, fmt.Sprintf("procedure %s expected input %d to be %s, received %s", p0.Function.Function, i, parameters[i].String(), attr.Type.String()))
				// we don't have to return here, we can continue to check the return type
			}
		}

		if returns == nil {
			t.options.ErrorListener.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeType, fmt.Sprintf("procedure %s does not return a table", p0.Function.Function))
			return
		}
		if !returns.IsTable {
			t.options.ErrorListener.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeType, fmt.Sprintf("procedure %s does not return a table", p0.Function.Function))
			return
		}

		rel := newRelation()
		for _, retCol := range returns.Fields {
			err := rel.AddAttribute(&QualifiedAttribute{
				Name: retCol.Name,
				Attribute: &Attribute{
					Type: retCol.Type,
				},
			})
			if err != nil {
				t.options.ErrorListener.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeSemantic, err.Error())
				return
			}
		}

		// add the relation to the context
		err = e.join(&QualifiedRelation{
			Name:     p0.Alias,
			Relation: rel,
		})
		if err != nil {
			t.options.ErrorListener.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeSemantic, err.Error())
		}
	})
}

func (t *typeVisitor) VisitRelationTable(p0 *tree.RelationTable) any {
	return evalFunc(func(e *evaluationContext) {
		tbl, ok := t.commonTables[p0.Name]
		if !ok {
			t.options.ErrorListener.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeSemantic, fmt.Sprintf("table %s not found", p0.Name))
			return
		}

		name := p0.Name
		if p0.Alias != "" {
			name = p0.Alias
		}

		err := e.join(&QualifiedRelation{
			Name:     name,
			Relation: tbl,
		})
		if err != nil {
			t.options.ErrorListener.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeSemantic, err.Error())
		}
	})
}

// the rest of the evalFunc visitors do not actually modify the evaluation context

func (t *typeVisitor) VisitUpsert(p0 *tree.Upsert) any {
	return evalFunc(func(e *evaluationContext) {
		if p0.ConflictTarget != nil {
			p0.ConflictTarget.Accept(t).(evalFunc)(e)
		}

		for _, set := range p0.Updates {
			set.Accept(t).(evalFunc)(e)
		}

		if p0.Where != nil {
			attr := p0.Where.Accept(t).(attributeFn)(e)

			if !attr.Type.Equals(types.BoolType) {
				t.options.ErrorListener.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeType, fmt.Sprintf("where clause must evaluate to boolean. Received: %s", attr.Type.String()))
				return
			}
		}
	})
}

func (t *typeVisitor) VisitUpdateSetClause(p0 *tree.UpdateSetClause) any {
	return evalFunc(func(e *evaluationContext) {
		// check that the columns exist
		// we can only update columns in the first table
		if len(e.joinOrder) == 0 {
			t.options.ErrorListener.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeSemantic, "no table to update")
			return
		}
		for _, col := range p0.Columns {
			_, _, err := e.findColumn(e.joinOrder[0], col)
			if err != nil {
				t.options.ErrorListener.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeSemantic, err.Error())
				return
			}
		}

		if p0.Expression != nil {
			// we can disregard the attribute here, we just want to visit
			p0.Expression.Accept(t).(attributeFn)(e)
		}
	})
}

func (t *typeVisitor) VisitConflictTarget(p0 *tree.ConflictTarget) any {
	return evalFunc(func(e *evaluationContext) {
		// check that the columns exist
		if len(e.joinOrder) == 0 {
			t.options.ErrorListener.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeSemantic, "no table to update")
			return
		}
		for _, col := range p0.IndexedColumns {
			_, _, err := e.findColumn(e.joinOrder[0], col)
			if err != nil {
				t.options.ErrorListener.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeSemantic, err.Error())
				return
			}
		}

		if p0.Where != nil {
			attr := p0.Where.Accept(t).(attributeFn)(e)

			if !attr.Type.Equals(types.BoolType) {
				t.options.ErrorListener.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeType, fmt.Sprintf("where clause must evaluate to boolean. Received: %s", attr.Type.String()))
			}
		}
	})
}

func (t *typeVisitor) VisitLimit(p0 *tree.Limit) any {
	return evalFunc(func(e *evaluationContext) {
		if p0.Expression != nil {
			limit := p0.Expression.Accept(t).(attributeFn)(e)

			if !limit.Type.Equals(types.IntType) {
				t.options.ErrorListener.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeType, fmt.Sprintf("limit must be an integer. Received: %s", limit.Type.String()))
				// we can continue here, since this will not affect future evaluation
			}
		}

		if p0.Offset != nil {
			offset := p0.Offset.Accept(t).(attributeFn)(e)

			if !offset.Type.Equals(types.IntType) {
				t.options.ErrorListener.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeType, fmt.Sprintf("offset must be an integer. Received: %s", offset.Type.String()))
			}
		}
	})
}

func (t *typeVisitor) VisitOrderBy(p0 *tree.OrderBy) any {
	return evalFunc(func(e *evaluationContext) {
		for _, term := range p0.OrderingTerms {
			term.Accept(t).(evalFunc)(e)
		}
	})
}

func (t *typeVisitor) VisitOrderingTerm(p0 *tree.OrderingTerm) any {
	return evalFunc(func(e *evaluationContext) {
		if p0.Expression == nil {
			return // not sure if this is possible, don't believe it is
		}
		p0.Expression.Accept(t).(attributeFn)(e)
	})
}

func (t *typeVisitor) VisitGroupBy(p0 *tree.GroupBy) any {
	return evalFunc(func(e *evaluationContext) {
		for _, col := range p0.Expressions {
			col.Accept(t).(attributeFn)(e)
		}

		if p0.Having != nil {
			attr := p0.Having.Accept(t).(attributeFn)(e)

			if !attr.Type.Equals(types.BoolType) {
				t.options.ErrorListener.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeType, fmt.Sprintf("having clause must be boolean. Received: %s", attr.Type.String()))
				return
			}
		}
	})
}

func (t *typeVisitor) VisitJoinPredicate(p0 *tree.JoinPredicate) any {
	return evalFunc(func(e *evaluationContext) {
		p0.Table.Accept(t).(evalFunc)(e)

		if p0.Constraint != nil {
			r := p0.Constraint.Accept(t).(attributeFn)(e)

			if !r.Type.Equals(types.BoolType) {
				t.options.ErrorListener.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeType, fmt.Sprintf("join constraint must be boolean. Received: %s", r.Type.String()))
			}
		}
	})
}

// END evalFunc

// attributeFn is returned from all visitor expressions.
// It allows us to evaluate return attributes once we
// have more context.
// The attribute name can be blank, and will only be set
// if the expression is a column.
type attributeFn func(ev *evaluationContext) *QualifiedAttribute

// BEGIN attributeFn

func (t *typeVisitor) VisitExpressionArithmetic(p0 *tree.ExpressionArithmetic) any {
	return attributeFn(func(ev *evaluationContext) *QualifiedAttribute {
		left := p0.Left.Accept(t).(attributeFn)
		right := p0.Right.Accept(t).(attributeFn)

		at := left(ev)
		if !at.Type.Equals(types.IntType) {
			t.options.ErrorListener.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeType, fmt.Sprintf("arithmetic expression expected int. Received: %s", at.Type.String()))
			return unknownAttr()
		}

		bt := right(ev)
		if !bt.Type.Equals(types.IntType) {
			t.options.ErrorListener.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeType, fmt.Sprintf("arithmetic expression expected int. Received: %s", bt.Type.String()))
			return unknownAttr()
		}

		if p0.TypeCast != nil {
			return anonAttr(p0.TypeCast)
		}

		return anonAttr(types.IntType)
	})
}

func (t *typeVisitor) VisitExpressionBetween(p0 *tree.ExpressionBetween) any {
	return attributeFn(func(ev *evaluationContext) *QualifiedAttribute {
		expr := p0.Expression.Accept(t).(attributeFn)
		left := p0.Left.Accept(t).(attributeFn)
		right := p0.Right.Accept(t).(attributeFn)

		et := expr(ev)

		lt := left(ev)

		rt := right(ev)

		if !et.Type.Equals(lt.Type) {
			t.options.ErrorListener.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeType, fmt.Sprintf("between expression expected %s. Received: %s", et.Type.Name, lt.Type.String()))
			return unknownAttr()
		}

		if !et.Type.Equals(rt.Type) {
			t.options.ErrorListener.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeType, fmt.Sprintf("between expression expected %s. Received: %s", et.Type.Name, rt.Type.String()))
			return unknownAttr()
		}

		if p0.TypeCast != nil {
			return anonAttr(p0.TypeCast)
		}

		return anonAttr(types.BoolType)
	})
}

func (t *typeVisitor) VisitExpressionBinaryComparison(p0 *tree.ExpressionBinaryComparison) any {
	return attributeFn(func(ev *evaluationContext) *QualifiedAttribute {
		left := p0.Left.Accept(t).(attributeFn)
		right := p0.Right.Accept(t).(attributeFn)

		at := left(ev)
		bt := right(ev)

		if !at.Type.Equals(bt.Type) {
			t.options.ErrorListener.NodeErr(parseTypes.MergeNodes(p0.Left.GetNode(), p0.Right.GetNode()), parseTypes.ParseErrorTypeType, fmt.Sprintf("cannot compare types: left: %s right: %s", at.Type.String(), bt.Type.String()))
			return unknownAttr()
		}

		if p0.TypeCast != nil {
			return anonAttr(p0.TypeCast)
		}

		return anonAttr(types.BoolType)
	})
}

func (t *typeVisitor) VisitExpressionBindParameter(p0 *tree.ExpressionBindParameter) any {
	return attributeFn(func(ev *evaluationContext) *QualifiedAttribute {
		c, ok := t.options.BindParams[p0.Parameter]
		if !ok {
			if t.options.ArbitraryBinds {
				c = types.UnknownType
			} else {
				t.options.ErrorListener.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeSemantic, fmt.Sprintf("bind parameter %s not found", util.UnformatParameterName(p0.Parameter)))
				return unknownAttr()
			}
		}

		if p0.TypeCast != nil {
			return anonAttr(p0.TypeCast)
		}

		return anonAttr(c)
	})
}

func (t *typeVisitor) VisitExpressionCase(p0 *tree.ExpressionCase) any {
	return attributeFn(func(ev *evaluationContext) *QualifiedAttribute {
		// whenTypes must always be bool, unless there is a case expression
		// If a case expression is present, then when clause must be the same type as the case expression
		expectedWhenType := types.BoolType
		if p0.CaseExpression != nil {
			c := p0.CaseExpression.Accept(t).(attributeFn)
			ct := c(ev)

			expectedWhenType = ct.Type
		}

		var neededType *types.DataType

		for _, w := range p0.WhenThenPairs {
			when := w[0].Accept(t).(attributeFn)
			whenType := when(ev)

			if !whenType.Type.Equals(expectedWhenType) {
				t.options.ErrorListener.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeType, fmt.Sprintf("when clause expected %s. Received: %s", expectedWhenType.String(), whenType.Type.String()))
				return unknownAttr()
			}

			then := w[1].Accept(t).(attributeFn)
			thenType := then(ev)

			if neededType == nil {
				neededType = thenType.Type
			}

			if !neededType.Equals(thenType.Type) {
				t.options.ErrorListener.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeType, fmt.Sprintf("all THEN types must be the same. Received: %s and %s", neededType.String(), thenType.Type.String()))
				return unknownAttr()
			}
		}

		if p0.ElseExpression != nil {
			e := p0.ElseExpression.Accept(t).(attributeFn)
			eType := e(ev)

			if !neededType.Equals(eType.Type) {
				t.options.ErrorListener.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeType, fmt.Sprintf("ELSE type must match THEN type. Received: %s and %s", neededType.String(), eType.Type.String()))
				return unknownAttr()
			}
		}

		if p0.TypeCast != nil {
			return anonAttr(p0.TypeCast)
		}

		return anonAttr(neededType)
	})
}

func (t *typeVisitor) VisitExpressionCollate(p0 *tree.ExpressionCollate) any {
	return attributeFn(func(ev *evaluationContext) *QualifiedAttribute {
		rel := p0.Expression.Accept(t).(attributeFn)(ev)

		if p0.TypeCast != nil {
			return anonAttr(p0.TypeCast)
		}

		return rel
	})
}

func (t *typeVisitor) VisitExpressionColumn(p0 *tree.ExpressionColumn) any {
	return attributeFn(func(ev *evaluationContext) *QualifiedAttribute {
		// if table is not qualified, we will attempt to qualify, and return an error on ambiguity
		tbl, col, err := ev.findColumn(p0.Table, p0.Column)
		if err != nil {
			t.options.ErrorListener.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeSemantic, err.Error())
			return unknownAttr()
		}

		if p0.Table == "" && t.options.Qualify {
			p0.Table = tbl // this will modify the statement
		}

		if p0.TypeCast != nil {
			return &QualifiedAttribute{
				Name: p0.Column,
				Attribute: &Attribute{
					Type: p0.TypeCast,
				},
			}
		}

		return col
	})
}

func (t *typeVisitor) VisitExpressionFunction(p0 *tree.ExpressionFunction) any {
	return attributeFn(func(ev *evaluationContext) *QualifiedAttribute {
		funcDef, ok := metadata.Functions[p0.Function]
		if !ok {
			// can be a procedure/foreign procedure
			params, returns, err := util.FindProcOrForeign(t.options.Schema, p0.Function)
			if err != nil {
				t.options.ErrorListener.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeSemantic, err.Error())
				return unknownAttr()
			}

			if len(p0.Inputs) != len(params) {
				t.options.ErrorListener.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeSemantic, fmt.Sprintf("procedure %s expected %d inputs, received %d", p0.Function, len(params), len(p0.Inputs)))
				return unknownAttr()
			}

			for i, in := range p0.Inputs {
				attr := in.Accept(t).(attributeFn)(ev)

				if !attr.Type.Equals(params[i]) {
					t.options.ErrorListener.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeType, fmt.Sprintf("procedure %s expected input %d to be %s, received %s", p0.Function, i, params[i].String(), attr.Type.String()))
					return unknownAttr()
				}
			}

			if returns == nil {
				return anonAttr(types.NullType)
			}

			if returns.IsTable {
				return anonAttr(types.NullType)
			}

			if len(returns.Fields) != 1 {
				t.options.ErrorListener.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeSemantic, fmt.Sprintf("procedure %s must return exactly one column", p0.Function))
				return unknownAttr()
			}

			return anonAttr(returns.Fields[0].Type)
		}

		var argTypes []*types.DataType
		for _, arg := range p0.Inputs {
			attr := arg.Accept(t).(attributeFn)(ev)
			argTypes = append(argTypes, attr.Type)
		}

		returnType, err := funcDef.ValidateArgs(argTypes)
		if err != nil {
			t.options.ErrorListener.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeType, err.Error())
			return unknownAttr()
		}

		if p0.TypeCast != nil {
			return anonAttr(p0.TypeCast)
		}

		return anonAttr(returnType)
	})
}

func (t *typeVisitor) VisitExpressionIs(p0 *tree.ExpressionIs) any {
	return attributeFn(func(ev *evaluationContext) *QualifiedAttribute {
		l := p0.Left.Accept(t).(attributeFn)
		r := p0.Right.Accept(t).(attributeFn)

		lt := l(ev)

		rt := r(ev)

		if !lt.Type.Equals(rt.Type) && !lt.Type.Equals(types.NullType) && !rt.Type.Equals(types.NullType) {
			t.options.ErrorListener.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeType, fmt.Sprintf("is expression expected %s. Received: %s", lt.Type.String(), rt.Type.String()))
			return unknownAttr()
		}

		if p0.TypeCast != nil {
			return anonAttr(p0.TypeCast)
		}

		return anonAttr(types.BoolType)
	})
}

func (t *typeVisitor) VisitExpressionList(p0 *tree.ExpressionList) any {
	return attributeFn(func(ev *evaluationContext) *QualifiedAttribute {
		var lastType *types.DataType
		for _, e := range p0.Expressions {
			et := e.Accept(t).(attributeFn)
			etType := et(ev)

			if lastType == nil {
				lastType = etType.Type
				continue
			}

			if !lastType.Equals(etType.Type) {
				t.options.ErrorListener.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeType, fmt.Sprintf("expression list expected %s. Received: %s", lastType.String(), etType.Type.String()))
				return unknownAttr()
			}
		}

		if p0.TypeCast != nil {
			return anonAttr(p0.TypeCast)
		}

		return anonAttr(lastType)
	})
}

func (t *typeVisitor) VisitExpressionTextLiteral(p0 *tree.ExpressionTextLiteral) any {
	return attributeFn(func(ev *evaluationContext) *QualifiedAttribute {
		if p0.TypeCast != nil {
			return anonAttr(p0.TypeCast)
		}

		return anonAttr(types.TextType)
	})
}

func (t *typeVisitor) VisitExpressionNumericLiteral(p0 *tree.ExpressionNumericLiteral) any {
	return attributeFn(func(ev *evaluationContext) *QualifiedAttribute {
		if p0.TypeCast != nil {
			return anonAttr(p0.TypeCast)
		}

		return anonAttr(types.IntType)
	})
}

func (t *typeVisitor) VisitExpressionBooleanLiteral(p0 *tree.ExpressionBooleanLiteral) any {
	return attributeFn(func(ev *evaluationContext) *QualifiedAttribute {
		if p0.TypeCast != nil {
			return anonAttr(p0.TypeCast)
		}

		return anonAttr(types.BoolType)
	})
}

func (t *typeVisitor) VisitExpressionNullLiteral(p0 *tree.ExpressionNullLiteral) any {
	return attributeFn(func(ev *evaluationContext) *QualifiedAttribute {
		if p0.TypeCast != nil {
			return anonAttr(p0.TypeCast)
		}

		return anonAttr(types.NullType)
	})
}

func (t *typeVisitor) VisitExpressionBlobLiteral(p0 *tree.ExpressionBlobLiteral) any {
	return attributeFn(func(ev *evaluationContext) *QualifiedAttribute {
		if p0.TypeCast != nil {
			return anonAttr(p0.TypeCast)
		}

		return anonAttr(types.BlobType)
	})
}

func (t *typeVisitor) VisitExpressionSelect(p0 *tree.ExpressionSelect) any {
	return attributeFn(func(ev *evaluationContext) *QualifiedAttribute {
		r := p0.Select.Accept(t).(returnFunc)(ev)

		shape := r.Shape()
		if len(shape) != 1 && !p0.IsExists {
			t.options.ErrorListener.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeSemantic, "subquery must return exactly one column")
			return unknownAttr()
		}

		if p0.TypeCast != nil {
			return anonAttr(p0.TypeCast)
		}

		if p0.IsExists {
			return anonAttr(types.BoolType)
		}

		return anonAttr(shape[0])
	})
}

func (t *typeVisitor) VisitExpressionStringCompare(p0 *tree.ExpressionStringCompare) any {
	return attributeFn(func(ev *evaluationContext) *QualifiedAttribute {
		a := p0.Left.Accept(t).(attributeFn)
		b := p0.Right.Accept(t).(attributeFn)

		// do these both need to be text? I believe so
		at := a(ev)
		bt := b(ev)

		if !at.Type.Equals(bt.Type) {
			t.options.ErrorListener.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeType, fmt.Sprintf("string comparison expression expected %s. Received: %s", at.Type.String(), bt.Type.String()))
			return unknownAttr()
		}

		if p0.Escape != nil {
			esc := p0.Escape.Accept(t).(attributeFn)
			et := esc(ev)

			if !et.Type.Equals(types.TextType) {
				t.options.ErrorListener.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeType, fmt.Sprintf("string comparison escape expected text. Received: %s", et.Type.String()))
				return unknownAttr()
			}
		}

		if p0.TypeCast != nil {
			return anonAttr(p0.TypeCast)
		}

		return anonAttr(types.BoolType)
	})
}

func (t *typeVisitor) VisitExpressionUnary(p0 *tree.ExpressionUnary) any {
	return attributeFn(func(ev *evaluationContext) *QualifiedAttribute {
		o := p0.Operand.Accept(t).(attributeFn)
		ot := o(ev)

		if !ot.Type.Equals(types.IntType) {
			t.options.ErrorListener.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeType, fmt.Sprintf("unary expression expected int. Received: %s", ot.Type.String()))
			return unknownAttr()
		}

		if p0.TypeCast != nil {
			return anonAttr(p0.TypeCast)
		}

		return anonAttr(types.IntType)
	})
}

// anonAttr is a helper function that creates an anonymous attribute
func anonAttr(t *types.DataType) *QualifiedAttribute {
	return &QualifiedAttribute{
		Attribute: &Attribute{
			Type: t,
		},
	}
}

// unknownAttr returns an unknown attribute.
// It should be used to avoid nil pointer dereferences.
func unknownAttr() *QualifiedAttribute {
	return anonAttr(types.UnknownType)
}

// END attributeFn

// returnFunc if a function that returns a relation.
// it is returned from INSERT, UPDATE, DELETE, and SELECT cores
// and stmts, as well as SimpleSelects.
type returnFunc func(e *evaluationContext) *Relation

// BEGIN returnFunc

func (t *typeVisitor) VisitInsertStmt(p0 *tree.InsertStmt) any {
	return returnFunc(func(e *evaluationContext) *Relation {
		for _, cte := range p0.CTE {
			cte.Accept(t).(evalFunc)(e)
		}

		return p0.Core.Accept(t).(returnFunc)(e)
	})
}

func (t *typeVisitor) VisitInsertCore(p0 *tree.InsertCore) any {
	return returnFunc(func(e *evaluationContext) *Relation {
		// we only search the visitor for the table,
		// since contextual table (such as CTEs) cannot be
		// inserted into.
		tbl, ok := t.commonTables[p0.Table]
		if !ok {
			t.options.ErrorListener.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeSemantic, fmt.Sprintf("table %s not found", p0.Table))
			return newRelation()
		}

		_, ok = t.ctes[p0.Table]
		if ok {
			t.options.ErrorListener.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeSemantic, fmt.Sprintf("cannot insert into common table expression %s", p0.Table))
			return newRelation()
		}

		// check that the columns exist
		for _, col := range p0.Columns {
			_, ok := tbl.Attribute(col)
			if !ok {
				t.options.ErrorListener.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeSemantic, fmt.Sprintf("column %s not found", col))
				return newRelation()
			}
		}

		// Postgres has a weird quirk with inserts:
		// tables can be aliased (e.g. insert into foo as bar),
		// but bar cannot be used in a subquery in the insert statement,
		// while foo can. The alias is only useable in the returning clause.
		// Therefore, we will not add the alias to the context.
		for _, row := range p0.Values {
			if len(row) != len(p0.Columns) {
				t.options.ErrorListener.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeSemantic, "mismatched column/value count")
				return newRelation()
			}

			for i, val := range row {
				attr := val.Accept(t).(attributeFn)(e)

				expectedAttr, ok := tbl.Attribute(p0.Columns[i])
				if !ok {
					t.options.ErrorListener.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeSemantic, fmt.Sprintf("unknown column %s", p0.Columns[i]))
					return newRelation()
				}

				if !expectedAttr.Type.Equals(attr.Type) {
					t.options.ErrorListener.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeSemantic, fmt.Sprintf("%s: type mismatch for column %s", ErrInvalidType.Error(), p0.Columns[i]))
					return newRelation()
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

		err := e2.join(&QualifiedRelation{
			Name:     name,
			Relation: tbl,
		})
		if err != nil {
			t.options.ErrorListener.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeSemantic, err.Error())
			return newRelation()
		}

		// similar to values, aliased insert tables cannot be used in the
		// conflict target or set clause. We will not add the alias to the context.
		if p0.Upsert != nil {
			p0.Upsert.Accept(t).(evalFunc)(e2)
		}

		// handle returning:

		if p0.ReturningClause == nil {
			return newRelation()
		}

		result := newRelation()

		p0.ReturningClause.Accept(t).(resultFunc)(e2, result)

		return result
	})
}

func (t *typeVisitor) VisitUpdateStmt(p0 *tree.UpdateStmt) any {
	return returnFunc(func(e *evaluationContext) *Relation {
		for _, cte := range p0.CTE {
			cte.Accept(t).(evalFunc)(e)
		}

		return p0.Core.Accept(t).(returnFunc)(e)
	})
}

func (t *typeVisitor) VisitUpdateCore(p0 *tree.UpdateCore) any {
	return returnFunc(func(e *evaluationContext) *Relation {
		tbl, ok := t.commonTables[p0.QualifiedTableName.TableName]
		if !ok {
			t.options.ErrorListener.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeSemantic, fmt.Sprintf("unknown table %s", p0.QualifiedTableName.TableName))
			return newRelation()
		}

		_, ok = t.ctes[p0.QualifiedTableName.TableName]
		if ok {
			t.options.ErrorListener.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeSemantic, fmt.Sprintf("cannot update common table expression %s", p0.QualifiedTableName.TableName))
			return newRelation()
		}

		name := p0.QualifiedTableName.TableName
		if p0.QualifiedTableName.TableAlias != "" {
			name = p0.QualifiedTableName.TableAlias
		}

		// we now want to update our context with joined relations since they can
		// be accessed in both the set clause and the where clause
		e2 := e.scope()

		err := e2.join(&QualifiedRelation{
			Name:     name,
			Relation: tbl,
		})
		if err != nil {
			t.options.ErrorListener.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeSemantic, err.Error())
			return newRelation()
		}

		if p0.From != nil {
			p0.From.Accept(t).(evalFunc)(e2)

			for _, set := range p0.UpdateSetClause {
				set.Accept(t).(evalFunc)(e2)
			}
		}

		if p0.Where != nil {
			whereType := p0.Where.Accept(t).(attributeFn)(e2)

			if !whereType.Type.Equals(types.BoolType) {
				t.options.ErrorListener.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeSemantic, fmt.Sprintf("%s: where clause must be boolean. Got %s", ErrInvalidType.Error(), whereType.Type.String()))
				return newRelation()
			}
		}

		if p0.Returning == nil {
			return newRelation()
		}

		result := newRelation()

		p0.Returning.Accept(t).(resultFunc)(e2, result)

		return result
	})
}

func (t *typeVisitor) VisitDeleteStmt(p0 *tree.DeleteStmt) any {
	return returnFunc(func(e *evaluationContext) *Relation {
		for _, cte := range p0.CTE {
			cte.Accept(t).(evalFunc)(e)
		}

		return p0.Core.Accept(t).(returnFunc)(e)
	})
}

func (t *typeVisitor) VisitDeleteCore(p0 *tree.DeleteCore) any {
	return returnFunc(func(e *evaluationContext) *Relation {
		tbl, ok := t.commonTables[p0.QualifiedTableName.TableName]
		if !ok {
			t.options.ErrorListener.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeSemantic, fmt.Sprintf("unknown table %s", p0.QualifiedTableName.TableName))
			return newRelation()
		}

		_, ok = t.ctes[p0.QualifiedTableName.TableName]
		if ok {
			t.options.ErrorListener.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeSemantic, fmt.Sprintf("cannot delete from common table expression %s", p0.QualifiedTableName.TableName))
			return newRelation()
		}

		name := p0.QualifiedTableName.TableName
		if p0.QualifiedTableName.TableAlias != "" {
			name = p0.QualifiedTableName.TableAlias
		}

		// we want to use the new context within the where clause
		// DELETE can use the alias in both the where clause and the returning clause
		e2 := e.scope()
		err := e2.join(&QualifiedRelation{
			Name:     name,
			Relation: tbl,
		})
		if err != nil {
			t.options.ErrorListener.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeSemantic, err.Error())
			return newRelation()
		}

		if p0.Where != nil {
			whereType := p0.Where.Accept(t).(attributeFn)(e2)

			if !whereType.Type.Equals(types.BoolType) {
				t.options.ErrorListener.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeSemantic, fmt.Sprintf("%s, where clause must be boolean. Got %s", ErrInvalidType.Error(), whereType.Type.String()))
				return newRelation()
			}
		}

		if p0.Returning == nil {
			return newRelation()
		}

		result := newRelation()

		p0.Returning.Accept(t).(resultFunc)(e2, result)
		return result
	})
}

func (t *typeVisitor) VisitSelectStmt(p0 *tree.SelectStmt) any {
	return returnFunc(func(e *evaluationContext) *Relation {
		for _, cte := range p0.CTE {
			cte.Accept(t).(evalFunc)(e)
		}

		return p0.Stmt.Accept(t).(returnFunc)(e)
	})
}

func (t *typeVisitor) VisitSelectCore(p0 *tree.SelectCore) any {
	return returnFunc(func(e *evaluationContext) *Relation {
		// we make a new scope so that we can join the tables
		// without affecting the outer scope
		e2 := e.scope()
		// we need to ensure that the relations all have the same shape
		res := p0.SimpleSelects[0].Accept(t).(returnFunc)(e2)

		expectedShape := res.Shape()

		for _, sel := range p0.SimpleSelects[1:] {
			// we create a separate scope for each select.
			selectCtx := e.scope()
			r := sel.Accept(t).(returnFunc)(selectCtx)

			shape := r.Shape()

			if len(shape) != len(expectedShape) {
				t.options.ErrorListener.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeSemantic, fmt.Sprintf("%s: compound selects must return the same number of columns. Expected %d. Received: %d", ErrCompoundShape.Error(), len(expectedShape), len(shape)))
				return newRelation()
			}

			for i, col := range shape {
				if !col.Equals(expectedShape[i]) {
					t.options.ErrorListener.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeSemantic, fmt.Sprintf("%s: compound selects must return the same types: Expected %s Received: %s", ErrCompoundShape.Error(), expectedShape[i].Name, col.Name))
					return newRelation()
				}
			}
		}

		// if this is a compound select, the joined tables from the selects are not in scope, and
		// we must instead join an anonymous relation that is the compound select.
		// if there is only one select, we can reference the joined tables.
		/* example query:
		 	SELECT * FROM (
				SELECT id FROM foo
				UNION
				SELECT id FROM bar
			)
			ORDER BY id
		)
		*/
		var e3 *evaluationContext
		if len(p0.SimpleSelects) > 1 {
			// copy in case we are in a correlated subquery
			e3 = e.copy()

			// we need to add the first returned relation anonymously to the context
			// so that we can use it in the ordering and limit
			err := e3.join(&QualifiedRelation{
				Name:     "",
				Relation: res,
			})
			if err != nil {
				t.options.ErrorListener.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeSemantic, err.Error())
				return newRelation()
			}
		} else {
			e3 = e2
		}

		if p0.OrderBy != nil {
			p0.OrderBy.Accept(t).(evalFunc)(e3)
		}

		if p0.Limit != nil {
			p0.Limit.Accept(t).(evalFunc)(e3)
		}

		return res
	})
}

func (t *typeVisitor) VisitSimpleSelect(p0 *tree.SimpleSelect) any {
	return returnFunc(func(e *evaluationContext) *Relation {
		if p0.From != nil {
			// we need to build the evaluation context based on the relation
			p0.From.Accept(t).(evalFunc)(e)
		}

		if p0.Where != nil {
			a := p0.Where.Accept(t).(attributeFn)(e)
			if !a.Attribute.Type.Equals(types.BoolType) {
				t.options.ErrorListener.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeSemantic, fmt.Sprintf("%s: where clause must be boolean", ErrInvalidType.Error()))
				return newRelation()
			}
		}

		// make an empty relation for the result
		result := newRelation()

		// apply the result columns
		for _, col := range p0.Columns {
			col.Accept(t).(resultFunc)(e, result)
		}

		if p0.GroupBy != nil {
			// group by context is a very weird case in postgres.
			// It can reference all joined tables, and can use both aliases
			// and unaliased columns. We therefore need to create a new context
			// that contains all of the old tables, with the aliases added
			// anonymously.

			e2 := e.copy()
			err := e2.mergeAnonymousSafe(result)
			if err != nil {
				t.options.ErrorListener.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeSemantic, err.Error())
				return newRelation()
			}

			p0.GroupBy.Accept(t).(evalFunc)(e2)
		}

		return result
	})
}

// END returnFunc

// resultFunc is a function that allows modifying a relation
// that will be returned by a relationFunc.
// It returned from ResultColumns and Returning Clauses.
// ResultColumns define the return relation from a SELECT,
// and Returning Clauses define the return relation from
// INSERT, UPDATE, and DELETE (if there is a RETURNING clause).
type resultFunc func(e *evaluationContext, r *Relation)

// BEGIN resultFunc

func (t *typeVisitor) VisitResultColumnExpression(p0 *tree.ResultColumnExpression) any {
	return resultFunc(func(e *evaluationContext, r *Relation) {
		c := p0.Expression.Accept(t).(attributeFn)
		val := c(e)

		if p0.Alias != "" {
			val.Name = p0.Alias
		}

		err := r.AddAttribute(val)
		if err != nil {
			t.options.ErrorListener.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeSemantic, err.Error())
		}
	})
}

func (t *typeVisitor) VisitResultColumnStar(p0 *tree.ResultColumnStar) any {
	return resultFunc(func(e *evaluationContext, r *Relation) {
		err := e.Loop(func(_ string, r2 *Relation) error {
			return r.Merge(r2)
		})
		if err != nil {
			t.options.ErrorListener.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeSemantic, err.Error())
		}
	})
}

func (t *typeVisitor) VisitResultColumnTable(p0 *tree.ResultColumnTable) any {
	return resultFunc(func(e *evaluationContext, r *Relation) {
		tbl, ok := e.joinedTables[p0.TableName]
		if !ok {
			t.options.ErrorListener.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeSemantic, fmt.Sprintf("table %s not found", p0.TableName))
			return
		}

		err := r.Merge(tbl)
		if err != nil {
			t.options.ErrorListener.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeSemantic, err.Error())
		}
	})
}

func (t *typeVisitor) VisitReturningClause(p0 *tree.ReturningClause) any {
	return resultFunc(func(e *evaluationContext, r *Relation) {
		for _, col := range p0.Returned {
			col.Accept(t).(resultFunc)(e, r)
		}
	})
}

func (t *typeVisitor) VisitReturningClauseColumn(p0 *tree.ReturningClauseColumn) any {
	return resultFunc(func(e *evaluationContext, r *Relation) {
		// this can either be return * or return expr

		// case 1: return *, preserving order
		if p0.All {
			err := e.Loop(func(_ string, r2 *Relation) error {
				return r.Merge(r2)
			})
			if err != nil {
				t.options.ErrorListener.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeSemantic, err.Error())
			}

			return
		}

		if p0.Expression == nil {
			t.options.ErrorListener.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeSemantic, "invalid returning clause")
			return
		}

		// case 2: return expr
		attribute := p0.Expression.Accept(t).(attributeFn)(e)

		// attempt to set the alias
		// if the attribute is not from a column,
		// and there is no alias, this will fail,
		// as the attribute will be anonymous
		// and therefore not accessible in the
		// returned relation.
		if p0.Alias != "" {
			attribute.Name = p0.Alias
		}

		err := r.AddAttribute(attribute)
		if err != nil {
			t.options.ErrorListener.NodeErr(p0.GetNode(), parseTypes.ParseErrorTypeSemantic, err.Error())
		}
	})
}

// END resultFunc
