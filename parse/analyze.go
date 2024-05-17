package parse

import (
	"fmt"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/utils/order"
)

/*
	this file performs analysis of SQL and procedures. It performs several main types of validation:
	1. Type checking: it ensures that all all statements and expressions return correct types.
	This is critical because plpgsql only throws type errors at runtime, which is really bad
	for a smart contract language.
	2. Deterministic ordering: it ensures that all queries have deterministic ordering, even if
	not specified by the query writer. It adds necessary ordering clauses to achieve this.
	3. Aggregate checks: it ensures that aggregate functions are used correctly, and that they
	can not be used to create non-determinism, and also that they return errors that would otherwise
	be thrown by Postgres at runtime.
	4. Mutative checks: it analyzes whether or not a procedure / sql statement is attempting to
	modify state. It does not return an error if it does, but will return a boolean indicating
	whether or not it is mutative. This can be used by callers to ensure that VIEW procedures
	are not mutative, which would otherwise only be checked at runtime when executing them with
	a read-only transaction.
	5. Contextual statement checks: Procedure statements that can only be used in certain contexts
	(e.g. loop breaks and RETURN NEXT) are checked to ensure that they are only used in loops.
	6. PLPGSQL Variable Declarations: It analyzes what variables should be declared at the top
	of a PLPGSQL statement, and what types they should be.
	7. Cartesian Join Checks: All joins must be joined using =, and one side of the join condition
	must be a unique column with no other math applied to it. Primary keys are also counted as unique,
	unless it is a compound primary key.

	DETERMINISTIC ORDERING RULES:
	If a SELECT statement is a simple select (e.g. does not use compound operators):

		1. All joined tables that are physical (and not subqueries or procedure calls) are ordered by their primary keys,
		in the order they are joined.

		2. If a SELECT has a DISTINCT clause, it will order by all columns being returned. The reason
		for this can be seen here: https://stackoverflow.com/questions/3492093/does-postgresql-support-distinct-on-multiple-columns.
		All previous rules do not apply.

		3. If a SELECT has a GROUP BY clause, all columns specified in the GROUP BY clause will be ordered.
		All previous rules do not apply.

	If a SELECT statement is a compound select (e.g. uses UNION, UNION ALL INTERSECT, EXCEPT):

		1. All returned columns are ordered by their position in the SELECT statement.

		2. If any compound SELECT statement has a GROUP BY, then it will return an error.
		This is a remnant of SQLite's rudimentary indexing, but these queries are fairly uncommon,
		and therefore not allowed for the time being


	AGGREGATE FUNCTION RULES:

		1. Aggregate functions can only be used in the SELECT clause, and not in the WHERE clause.

		2. All columns referenced in HAVING or return columns must be in the GROUP BY clause, unless they are
		in an aggregate function.

		3. All columns used within aggregate functions cannot be specified in the GROUP BY clause.

		4. If there is an aggregate in the return columns and no GROUP BY clause, then there can only
		be one column in the return columns (the column with the aggregate function).
*/

// blockContext is the context for the current block. This is can be an action, procedure,
// or sql block.
type blockContext struct {
	// schema is the current schema
	schema *types.Schema
	// variables holds information about all variable declarations in the block
	// It holds both user variables like $arg1, $arg2, and contextual variables,
	// like @caller and @txid. It will be populated while the analyzer is running,
	// but is prepopulated with the procedure's arguments and contextual variables.
	variables map[string]*types.DataType
	// anonymousVariables holds information about all anonymous variable declarations in the block.
	// Anonymous variables are objects with fields, such as the receiver of loops.
	// The map maps the name to the fields to their data types.
	// The map will be populated while the analyzer is running.
	anonymousVariables map[string]map[string]*types.DataType
	// errs is used for passing errors back to the caller.
	errs *errorListener
}

// variableExists checks if a variable exists in the current block.
// It will check both user variables and anonymous variables.
func (b *blockContext) variableExists(name string) bool {
	_, ok := b.variables[name]
	if ok {
		return true
	}

	_, ok = b.anonymousVariables[name]
	return ok
}

// sqlContext is the context of the current SQL statement
type sqlContext struct {
	// joinedRelations tracks all relations joined on the current SQL statement.
	joinedRelations []*Relation
	// outerRelations are relations that are not joined on the scope, but are available.
	// These are typically outer queries in a subquery. Calling these will be a correlated subquery.
	outerRelations []*Relation
	// joinedTables maps all used table names/aliases to their table definitions.
	// The tables named here are also included in joinedRelations, but not
	// all joinedRelations are in this map. This map ONLY includes actual SQL
	// tables joined in this context, not joined subqueries or procedure calls.
	// These are used for default ordering.
	joinedTables map[string]*types.Table
	// ctes are the common table expressions in the current scope.
	ctes []*Relation
	// outerScope is the scope of the outer query.
	outerScope *sqlContext

	// temp are values that are temporary and not even saved within the same scope.
	// they are used in highly specific contexts, and shouldn't be relied on unless
	// specifically documented. All temp values are denoted with a _.

	// inAggregate is true if we are within an aggregate functions parameters.
	// it should only be used in ExpressionColumn, and set in ExpressionFunctionCall.
	_inAggregate bool
	// containsAggregate is true if the current expression contains an aggregate function.
	// it is set in ExpressionFunctionCall, and accessed/reset in SelectCore.
	_containsAggregate bool
	// columnInAggregate is the column found within an aggregate function,
	// comprised of the relation and attribute.
	// It is set in ExpressionColumn, and accessed/reset in
	// SelectCore. It is nil if none are found.
	_columnInAggregate *[2]string
	// columnsOutsideAggregate are columns found outside of an aggregate function.
	// It is set in ExpressionColumn, and accessed/reset in
	// SelectCore
	_columnsOutsideAggregate [][2]string
}

func newSQLContext() *sqlContext {
	return &sqlContext{
		joinedTables: make(map[string]*types.Table),
	}
}

// setTempValuesToZero resets all temp values to their zero values.
func (s *sqlContext) setTempValuesToZero() {
	s._inAggregate = false
	s._containsAggregate = false
	s._columnInAggregate = nil
	s._columnsOutsideAggregate = nil
}

// copy copies the sqlContext.
// it does not copy the outer scope.
func (c *sqlContext) copy() *sqlContext {
	joinedRelations := make([]*Relation, len(c.joinedRelations))
	for i, r := range c.joinedRelations {
		joinedRelations[i] = r.Copy()
	}

	outerRelations := make([]*Relation, len(c.outerRelations))
	for i, r := range c.outerRelations {
		outerRelations[i] = r.Copy()
	}

	// ctes don't need to be copied right now since they are not modified.
	colsOutsideAgg := make([][2]string, len(c._columnsOutsideAggregate))
	copy(colsOutsideAgg, c._columnsOutsideAggregate)

	return &sqlContext{
		joinedRelations: joinedRelations,
		outerRelations:  outerRelations,
		ctes:            c.ctes,
		joinedTables:    c.joinedTables,
	}
}

// joinRelation adds a relation to the context.
func (c *sqlContext) joinRelation(r *Relation) error {
	// check if the relation is already joined
	_, ok := c.getJoinedRelation(r.Name)
	if ok {
		return ErrTableAlreadyJoined
	}

	c.joinedRelations = append(c.joinedRelations, r)
	return nil
}

// join joins a table. It will return an error if the table is already joined.
func (c *sqlContext) join(name string, t *types.Table) error {
	_, ok := c.joinedTables[name]
	if ok {
		return ErrTableAlreadyJoined
	}

	c.joinedTables[name] = t
	return nil
}

// getJoinedRelation returns the relation with the given name.
func (c *sqlContext) getJoinedRelation(name string) (*Relation, bool) {
	for _, r := range c.joinedRelations {
		if r.Name == name {
			return r, true
		}
	}

	return nil, false
}

// getOuterRelation returns the relation with the given name.
func (c *sqlContext) getOuterRelation(name string) (*Relation, bool) {
	for _, r := range c.outerRelations {
		if r.Name == name {
			return r, true
		}
	}

	return nil, false
}

// findAttribute searches for a attribute in the specified relation.
// if the relation is empty, it will search all joined and outer relations.
// If the relation is empty and many columns are found, it will return an error.
// It returns both an error and an error message in case of an error.
// This is because it is meant to pass errors back to the error listener.
func (c *sqlContext) findAttribute(relation string, column string) (relName string, attr *Attribute, msg string, err error) {
	if relation == "" {
		foundAttrs := make([]*Attribute, 0)

		for _, r := range c.joinedRelations {
			for _, a := range r.Attributes {
				if a.Name == column {
					relName = r.Name
					foundAttrs = append(foundAttrs, a)
				}
			}
		}

		for _, r := range c.outerRelations {
			for _, a := range r.Attributes {
				if a.Name == column {
					relName = r.Name
					foundAttrs = append(foundAttrs, a)
				}
			}
		}

		switch len(foundAttrs) {
		case 0:
			return "", nil, column, ErrUnknownColumn
		case 1:
			return relName, foundAttrs[0], "", nil
		default:
			return "", nil, column, ErrAmbiguousColumn
		}
	}

	r, ok := c.getJoinedRelation(relation)
	if !ok {
		r, ok = c.getOuterRelation(relation)
		if !ok {
			return "", nil, relation, ErrUnknownTable
		}
	}

	for _, a := range r.Attributes {
		if a.Name == column {
			return r.Name, a, "", nil
		}
	}

	return "", nil, relation + "." + column, ErrUnknownColumn
}

// findColumn searches for a column and table in the tables of joinedTables.
// It works similar to findAttribute, where if the table is empty, it will search all tables.
// If the table is empty and many columns are found, it will return an error.
// It returns both an error and an error message in case of an error.
func (c *sqlContext) findColumn(table string, column string) (*types.Table, *types.Column, string, error) {
	if table == "" {
		found := make([]struct {
			table *types.Table
			col   *types.Column
		}, 0)

		for _, t := range c.joinedTables {
			col, ok := t.FindColumn(column)
			if ok {
				found = append(found, struct {
					table *types.Table
					col   *types.Column
				}{table: t, col: col})
			}
		}

		switch len(found) {
		case 0:
			return nil, nil, column, ErrUnknownColumn
		case 1:
			return found[0].table, found[0].col, "", nil
		default:
			return nil, nil, column, ErrAmbiguousColumn
		}
	}

	t, ok := c.joinedTables[table]
	if !ok {
		return nil, nil, table, ErrUnknownTable
	}

	col, found := t.FindColumn(column)
	if !found {
		return nil, nil, column, ErrUnknownColumn
	}

	return t, col, "", nil
}

// colIsUnique checks if the given column is unique. It also requires the column's
// table to be passed, because it will return true if the column is the sole primary key.
// It the table is passed as an empty string, it will search all joined tables.
// it will return an error and a message for the error if one is encountered.
func (s *sqlContext) colIsUnique(tblStr string, colStr string) (bool, string, error) {
	tbl, col, msg, err := s.findColumn(tblStr, colStr)
	if err != nil {
		return false, msg, err
	}

	if col.HasAttribute(types.UNIQUE) {
		return true, "", nil
	}

	pks, err := tbl.GetPrimaryKey()
	if err != nil {
		// error shouldn't ever happen because we should have validated
		// the schema already, but just in case
		return false, err.Error(), ErrTableDefinition
	}

	if len(pks) != 1 {
		return false, "", nil
	}

	return pks[0] == col.Name, "", nil
}

// scope moves the current scope to outer scope,
// and sets the current scope to a new scope.
func (c *sqlContext) scope() {
	// copy the outer tables and joined tables to avoid modifying the outer scope.
	outerTbls := make([]*Relation, len(c.joinedRelations)+len(c.outerRelations))
	copy(outerTbls, c.joinedRelations)
	copy(outerTbls[len(c.joinedRelations):], c.outerRelations)

	// move to the outer scope
	c.outerScope = c

	c.outerRelations = outerTbls
	c.joinedRelations = nil
	c.setTempValuesToZero()

	// ctes don't need to be copied since they are not modified,
	// and are available across all scopes.
}

// popScope moves the current scope to the outer scope.
func (c *sqlContext) popScope() {
	*c = *c.outerScope
}

/*
	this visitor breaks down nodes into 4 different types:
	- Expressions: expressions simply return *Attribute. The name on all of these will be empty UNLESS it is a column reference.
	- CommonTableExpressions: the only node that can directly add tables to outerRelations slice.

*/

// sqlAnalyzer visits SQL nodes and analyzes them.
type sqlAnalyzer struct {
	UnimplementedSqlVisitor
	blockContext
	sqlCtx    *sqlContext
	sqlResult sqlAnalyzeResult
}

type sqlAnalyzeResult struct {
	Mutative bool
}

// startSQLAnalyze initializes all fields of the sqlAnalyzer.
func (s *sqlAnalyzer) startSQLAnalyze() {
	s.sqlCtx = &sqlContext{
		joinedTables: make(map[string]*types.Table),
	}
}

// endSQLAnalyze is called at the end of the analysis.
func (s *sqlAnalyzer) endSQLAnalyze() *sqlAnalyzeResult {
	res := s.sqlResult
	s.sqlCtx = nil
	return &res
}

var _ Visitor = (*sqlAnalyzer)(nil)

// typeErr should be used when a type error is encountered.
// It returns an unknown attribute and adds an error to the error listener.
func (s *sqlAnalyzer) typeErr(node Node, t1, t2 *types.DataType) *types.DataType {
	s.errs.AddErr(node, ErrType, fmt.Sprintf("%s != %s", t1.String(), t2.String()))
	return cast(node, types.UnknownType)
}

// expressionTypeErr should be used if we expect an expression to return a *types.DataType,
// but it returns something else. It will attempt to read the actual type and create an error
// message that is helpful for the end user.
func (s *sqlAnalyzer) expressionTypeErr(e Expression) *types.DataType {
	// if expression is a receiver from a loop, it will be a map
	_, ok := e.Accept(s).(map[string]*types.DataType)
	if ok {
		s.errs.AddErr(e, ErrType, "invalid usage of compound type, expected scalar value")
		return cast(e, types.UnknownType)
	}

	// if expression is a procedure call that returns a table, it will be a slice of attributes
	_, ok = e.Accept(s).([]*Attribute)
	if ok {
		s.errs.AddErr(e, ErrType, "procedure returns table, not a scalar value")
		return cast(e, types.UnknownType)
	}

	// if it iis a procedure call that returns many values, it will be a slice of data types
	vals, ok := e.Accept(s).([]*types.DataType)
	if ok {
		s.errs.AddErr(e, ErrType, "expected procedure to return a single value, returns %d", len(vals))
		return cast(e, types.UnknownType)

	}

	s.errs.AddErr(e, ErrType, "could not infer expected type")
	return cast(e, types.UnknownType)
}

// cast will return the type case if one exists. If not, it will simply
// return the passed type.
func cast(castable any, fallback *types.DataType) *types.DataType {
	if castable == nil {
		return fallback
	}

	c, ok := castable.(interface{ GetTypeCast() *types.DataType })
	if !ok {
		return fallback
	}

	if c.GetTypeCast() == nil {
		return fallback
	}

	return c.GetTypeCast()
}

func (s *sqlAnalyzer) VisitExpressionLiteral(p0 *ExpressionLiteral) any {
	return cast(p0, p0.Type)
}

func (s *sqlAnalyzer) VisitExpressionFunctionCall(p0 *ExpressionFunctionCall) any {
	// function call should either be against a known function, or a procedure.
	fn, ok := Functions[p0.Name]
	if !ok {
		// if not found, it might be a schema procedure.
		proc, found := s.schema.FindProcedure(p0.Name)
		if !found {
			s.errs.AddErr(p0, ErrUnknownFunctionOrProcedure, p0.Name)
			return cast(p0, types.UnknownType)
		}

		// if it is a procedure, it cannot use distinct or *
		if p0.Distinct {
			s.errs.AddErr(p0, ErrFunctionSignature, "cannot use DISTINCT when calling procedure %s", p0.Name)
			return cast(p0, types.UnknownType)
		}
		if p0.Star {
			s.errs.AddErr(p0, ErrFunctionSignature, "cannot use * when calling procedure %s", p0.Name)
			return cast(p0, types.UnknownType)
		}

		// verify the inputs
		if len(p0.Args) != len(proc.Parameters) {
			s.errs.AddErr(p0, ErrFunctionSignature, "expected %d arguments, received %d", len(proc.Parameters), len(p0.Args))
			return cast(p0, types.UnknownType)
		}

		for i, arg := range p0.Args {
			dt, ok := arg.Accept(s).(*types.DataType)
			if !ok {
				return s.expressionTypeErr(arg)
			}

			if !dt.Equals(proc.Parameters[i].Type) {
				return s.typeErr(arg, dt, proc.Parameters[i].Type)
			}
		}

		return s.returnProcedureReturnExpr(p0, p0.Name, proc.Returns)
	}

	// the function is a built in function. If using DISTINCT, it needs to be an aggregate
	// if using *, it needs to support it.
	if p0.Distinct && !fn.IsAggregate {
		s.errs.AddErr(p0, ErrFunctionSignature, "DISTINCT can only be used with aggregate functions")
		return cast(p0, types.UnknownType)
	}

	if fn.IsAggregate {
		s.sqlCtx._inAggregate = true
		s.sqlCtx._containsAggregate = true
		defer func() { s.sqlCtx._inAggregate = false }()
	}

	// if the function is called with *, we need to ensure it supports it.
	// If not, then we validate all args and return the type.
	var returnType *types.DataType
	if p0.Star {
		if fn.StarArgReturn == nil {
			s.errs.AddErr(p0, ErrFunctionSignature, "function does not support *")
			return cast(p0, types.UnknownType)
		}

		// if calling with *, it must have no args
		if len(p0.Args) != 0 {
			s.errs.AddErr(p0, ErrFunctionSignature, "function does not accept arguments when using *")
			return cast(p0, types.UnknownType)
		}

		returnType = fn.StarArgReturn
	} else {
		argTyps := make([]*types.DataType, len(p0.Args))
		for i, arg := range p0.Args {
			dt, ok := arg.Accept(s).(*types.DataType)
			if !ok {
				return s.expressionTypeErr(arg)
			}

			argTyps[i] = dt
		}

		var err error
		returnType, err = fn.ValidateArgs(argTyps)
		if err != nil {
			s.errs.AddErr(p0, ErrFunctionSignature, err.Error())
			return cast(p0, types.UnknownType)
		}
	}

	return cast(p0, returnType)
}

func (s *sqlAnalyzer) VisitExpressionForeignCall(p0 *ExpressionForeignCall) any {
	// foreign call must be defined as a foreign procedure
	proc, found := s.schema.FindForeignProcedure(p0.Name)
	if !found {
		s.errs.AddErr(p0, ErrUnknownFunctionOrProcedure, p0.Name)
		return cast(p0, types.UnknownType)
	}

	if len(p0.ContextualArgs) != 2 {
		s.errs.AddErr(p0, ErrFunctionSignature, "expected 2 contextual arguments, received %d", len(p0.ContextualArgs))
		return cast(p0, types.UnknownType)
	}

	// contextual args have to be strings
	for _, ctxArgs := range p0.ContextualArgs {
		dt, ok := ctxArgs.Accept(s).(*types.DataType)
		if !ok {
			return s.expressionTypeErr(ctxArgs)
		}

		if !dt.Equals(types.TextType) {
			s.errs.AddErr(ctxArgs, ErrFunctionSignature, "expected text type, received %s", dt.String())
			return cast(p0, types.UnknownType)
		}
	}

	// verify the inputs
	if len(p0.Args) != len(proc.Parameters) {
		s.errs.AddErr(p0, ErrFunctionSignature, "expected %d arguments, received %d", len(proc.Parameters), len(p0.Args))
		return cast(p0, types.UnknownType)
	}

	for i, arg := range p0.Args {
		dt, ok := arg.Accept(s).(*types.DataType)
		if !ok {
			return s.expressionTypeErr(arg)
		}

		if !dt.Equals(proc.Parameters[i]) {
			return s.typeErr(arg, dt, proc.Parameters[i])
		}
	}

	return s.returnProcedureReturnExpr(p0, p0.Name, proc.Returns)
}

// returnProcedureReturnExpr handles a procedure return used as an expression return. It mandates
// that the procedure returns a single value, or a table.
func (s *sqlAnalyzer) returnProcedureReturnExpr(p0 Node, procedureName string, ret *types.ProcedureReturn) any {
	// if an expression calls a function, it should return exactly one value or a table.
	if ret == nil {
		s.errs.AddErr(p0, ErrFunctionSignature, "procedure %s does not return a value", procedureName)
		return cast(p0, types.UnknownType)
	}

	// if it returns a table, we need to return it as a set of attributes.
	if ret.IsTable {
		attrs := make([]*Attribute, len(ret.Fields))
		for i, f := range ret.Fields {
			attrs[i] = &Attribute{
				Name: f.Name,
				Type: f.Type,
			}
		}

		return attrs
	}

	switch len(ret.Fields) {
	case 0:
		s.errs.AddErr(p0, ErrFunctionSignature, "procedure %s does not return a value", procedureName)
		return cast(p0, types.UnknownType)
	case 1:
		return cast(p0, ret.Fields[0].Type)
	default:
		castable, ok := p0.(interface{ GetTypeCast() *types.DataType })
		if ok {
			if castable.GetTypeCast() != nil {
				s.errs.AddErr(p0, ErrType, "cannot type cast multiple return values")
			}
		}

		retVals := make([]*types.DataType, len(ret.Fields))
		for i, f := range ret.Fields {
			retVals[i] = f.Type.Copy()
		}

		return retVals
	}
}

func (s *sqlAnalyzer) VisitExpressionVariable(p0 *ExpressionVariable) any {
	dt, ok := s.blockContext.variables[p0.String()]
	if !ok {
		// if not found, it could be an anonymous variable.
		anonVar, ok := s.blockContext.anonymousVariables[p0.String()]
		if ok {
			// if it is anonymous, we cannot type cast, since it is a compound type.
			if p0.GetTypeCast() != nil {
				s.errs.AddErr(p0, ErrType, "cannot type cast compound variable")
			}

			return anonVar
		}

		// if not found, then var does not exist.
		s.errs.AddErr(p0, ErrUndeclaredVariable, p0.String())
		return cast(p0, types.UnknownType)
	}

	return cast(p0, dt)
}

func (s *sqlAnalyzer) VisitExpressionArrayAccess(p0 *ExpressionArrayAccess) any {
	idxAttr, ok := p0.Index.Accept(s).(*types.DataType)
	if !ok {
		return s.expressionTypeErr(p0.Index)
	}
	if !idxAttr.Equals(types.IntType) {
		return s.typeErr(p0.Index, idxAttr, types.IntType)
	}

	arrAttr, ok := p0.Array.Accept(s).(*types.DataType)
	if !ok {
		return s.expressionTypeErr(p0.Array)
	}

	if !arrAttr.IsArray {
		s.errs.AddErr(p0.Array, ErrType, "expected array")
		return cast(p0, types.UnknownType)
	}

	return cast(p0, &types.DataType{
		Name:     arrAttr.Name,
		Metadata: arrAttr.Metadata,
		// leave IsArray as false since we are accessing an element.
	})
}

func (s *sqlAnalyzer) VisitExpressionMakeArray(p0 *ExpressionMakeArray) any {
	if len(p0.Values) == 0 {
		s.errs.AddErr(p0, ErrAssignment, "array instantiation must have at least one element")
		return cast(p0, types.UnknownType)
	}

	first, ok := p0.Values[0].Accept(s).(*types.DataType)
	if !ok {
		return s.expressionTypeErr(p0.Values[0])
	}

	for _, v := range p0.Values {
		typ, ok := v.Accept(s).(*types.DataType)
		if !ok {
			return s.expressionTypeErr(v)
		}

		if !typ.Equals(first) {
			return s.typeErr(v, typ, first)
		}
	}

	return cast(p0, &types.DataType{
		Name:     first.Name,
		Metadata: first.Metadata,
		IsArray:  true,
	})
}

func (s *sqlAnalyzer) VisitExpressionFieldAccess(p0 *ExpressionFieldAccess) any {
	// field access needs to be accessing a compound type.
	// currently, compound types can only be anonymous variables declared
	// as loop receivers.
	anonType, ok := p0.Record.Accept(s).(map[string]*types.DataType)
	if !ok {
		s.errs.AddErr(p0.Record, ErrType, "cannot access field on non-compound type")
		return cast(p0, types.UnknownType)
	}

	dt, ok := anonType[p0.Field]
	if !ok {
		s.errs.AddErr(p0, ErrType, fmt.Sprintf("unknown field %s", p0.Field))
		return cast(p0, types.UnknownType)
	}

	return cast(p0, dt)
}

func (s *sqlAnalyzer) VisitExpressionParenthesized(p0 *ExpressionParenthesized) any {
	dt, ok := p0.Inner.Accept(s).(*types.DataType)
	if !ok {
		return s.expressionTypeErr(p0.Inner)
	}

	return cast(p0, dt)
}

func (s *sqlAnalyzer) VisitExpressionComparison(p0 *ExpressionComparison) any {
	left, ok := p0.Left.Accept(s).(*types.DataType)
	if !ok {
		return s.expressionTypeErr(p0.Left)
	}

	right, ok := p0.Right.Accept(s).(*types.DataType)
	if !ok {
		return s.expressionTypeErr(p0.Right)
	}

	if !left.Equals(right) {
		return s.typeErr(p0.Right, right, left)
	}

	return cast(p0, types.BoolType)
}

func (s *sqlAnalyzer) VisitExpressionLogical(p0 *ExpressionLogical) any {
	left, ok := p0.Left.Accept(s).(*types.DataType)
	if !ok {
		return s.expressionTypeErr(p0.Left)
	}

	right, ok := p0.Right.Accept(s).(*types.DataType)
	if !ok {
		return s.expressionTypeErr(p0.Right)
	}

	if !left.Equals(types.BoolType) {
		return s.typeErr(p0.Left, left, types.BoolType)
	}

	if !right.Equals(types.BoolType) {
		return s.typeErr(p0.Right, right, types.BoolType)
	}

	return cast(p0, types.BoolType)
}

func (s *sqlAnalyzer) VisitExpressionArithmetic(p0 *ExpressionArithmetic) any {
	left, ok := p0.Left.Accept(s).(*types.DataType)
	if !ok {
		return s.expressionTypeErr(p0.Left)
	}

	right, ok := p0.Right.Accept(s).(*types.DataType)
	if !ok {
		return s.expressionTypeErr(p0.Right)
	}

	if !left.Equals(right) {
		return s.typeErr(p0.Right, right, left)
	}

	// both must be numeric UNLESS it is a concat
	if p0.Operator == ArithmeticOperatorConcat {
		if !left.Equals(types.TextType) {
			s.errs.AddErr(p0.Left, ErrType, "expected text type, received %s", left.String())
			return cast(p0, types.UnknownType)
		}
	} else {
		if !left.IsNumeric() {
			s.errs.AddErr(p0.Left, ErrType, "expected numeric type, received %s", left.String())
			return cast(p0, types.UnknownType)
		}
	}

	return cast(p0, left)
}

func (s *sqlAnalyzer) VisitExpressionUnary(p0 *ExpressionUnary) any {
	e, ok := p0.Expression.Accept(s).(*types.DataType)
	if !ok {
		return s.expressionTypeErr(p0.Expression)
	}

	switch p0.Operator {
	default:
		panic("unknown unary operator")
	case UnaryOperatorPos:
		if !e.IsNumeric() {
			s.errs.AddErr(p0.Expression, ErrType, "expected numeric type, received %s", e.String())
			return cast(p0, types.UnknownType)
		}
	case UnaryOperatorNeg:
		if !e.IsNumeric() {
			s.errs.AddErr(p0.Expression, ErrType, "expected numeric type, received %s", e.String())
			return cast(p0, types.UnknownType)
		}

		if e.Equals(types.Uint256Type) {
			s.errs.AddErr(p0.Expression, ErrType, "cannot negate uint256")
			return cast(p0, types.UnknownType)
		}
	case UnaryOperatorNot:
		if !e.Equals(types.BoolType) {
			s.errs.AddErr(p0.Expression, ErrType, "expected boolean type, received %s", e.String())
			return cast(p0, types.UnknownType)
		}
	}

	return cast(p0, e)
}

func (s *sqlAnalyzer) VisitExpressionColumn(p0 *ExpressionColumn) any {
	// findColumn accounts for empty tables in search, so we do not have to
	// worry about it being qualified or not.
	relName, col, msg, err := s.sqlCtx.findAttribute(p0.Table, p0.Column)
	if err != nil {
		s.errs.AddErr(p0, err, msg)
		return cast(p0, types.UnknownType)
	}

	if s.sqlCtx._inAggregate {
		if s.sqlCtx._columnInAggregate != nil {
			s.errs.AddErr(p0, ErrAggregate, "cannot use multiple columns in aggregate function args")
			return cast(p0, types.UnknownType)
		}

		s.sqlCtx._columnInAggregate = &[2]string{relName, col.Name}
	} else {
		s.sqlCtx._columnsOutsideAggregate = append(s.sqlCtx._columnsOutsideAggregate, [2]string{relName, col.Name})
	}

	return cast(p0, col.Type)
}

func (s *sqlAnalyzer) VisitExpressionCollate(p0 *ExpressionCollate) any {
	e, ok := p0.Expression.Accept(s).(*types.DataType)
	if !ok {
		return s.expressionTypeErr(p0.Expression)
	}

	if !e.Equals(types.TextType) {
		s.errs.AddErr(p0.Expression, ErrType, "expected text type, received %s", e.String())
		return cast(p0, types.UnknownType)
	}

	return cast(p0, e)
}

func (s *sqlAnalyzer) VisitExpressionStringComparison(p0 *ExpressionStringComparison) any {
	left, ok := p0.Left.Accept(s).(*types.DataType)
	if !ok {
		return s.expressionTypeErr(p0.Left)
	}

	right, ok := p0.Right.Accept(s).(*types.DataType)
	if !ok {
		return s.expressionTypeErr(p0.Right)
	}

	if !left.Equals(types.TextType) {
		return s.typeErr(p0.Left, left, types.TextType)
	}

	if !right.Equals(types.TextType) {
		return s.typeErr(p0.Right, right, types.TextType)
	}

	return cast(p0, types.BoolType)
}

func (s *sqlAnalyzer) VisitExpressionIs(p0 *ExpressionIs) any {
	left, ok := p0.Left.Accept(s).(*types.DataType)
	if !ok {
		return s.expressionTypeErr(p0.Left)
	}

	right, ok := p0.Right.Accept(s).(*types.DataType)
	if !ok {
		return s.expressionTypeErr(p0.Right)
	}

	// right has to be null, unless distinct is true. If distinct is true,
	// then left and right must be the same type
	if p0.Distinct {
		if !left.Equals(right) {
			return s.typeErr(p0.Right, right, left)
		}
	} else {
		if !right.Equals(types.NullType) {
			return s.typeErr(p0.Right, right, types.NullType)
		}
	}

	return cast(p0, types.BoolType)
}

func (s *sqlAnalyzer) VisitExpressionIn(p0 *ExpressionIn) any {
	exprType, ok := p0.Expression.Accept(s).(*types.DataType)
	if !ok {
		return s.expressionTypeErr(p0.Expression)
	}

	switch {
	case len(p0.List) > 0:
		for _, e := range p0.List {
			dt, ok := e.Accept(s).(*types.DataType)
			if !ok {
				return s.expressionTypeErr(e)
			}

			if !dt.Equals(exprType) {
				return s.typeErr(e, dt, exprType)
			}
		}
	case p0.Subquery != nil:
		rel, ok := p0.Subquery.Accept(s).([]*Attribute)
		if !ok {
			panic("expected query to return attributes")
		}

		if len(rel) != 1 {
			s.errs.AddErr(p0.Subquery, ErrResultShape, "subquery expressions must return exactly 1 column, received %d", len(rel))
			return cast(p0, types.UnknownType)
		}

		if !rel[0].Type.Equals(exprType) {
			return s.typeErr(p0.Subquery, rel[0].Type, exprType)
		}
	default:
		panic("list or subquery must be set for in expression")
	}

	return cast(p0, types.BoolType)
}

func (s *sqlAnalyzer) VisitExpressionBetween(p0 *ExpressionBetween) any {
	between, ok := p0.Expression.Accept(s).(*types.DataType)
	if !ok {
		return s.expressionTypeErr(p0.Expression)
	}

	lower, ok := p0.Lower.Accept(s).(*types.DataType)
	if !ok {
		return s.expressionTypeErr(p0.Lower)
	}

	upper, ok := p0.Upper.Accept(s).(*types.DataType)
	if !ok {
		return s.expressionTypeErr(p0.Upper)
	}

	if !between.Equals(lower) {
		return s.typeErr(p0.Lower, lower, between)
	}

	if !between.Equals(upper) {
		return s.typeErr(p0.Upper, upper, between)
	}

	if !between.IsNumeric() {
		s.errs.AddErr(p0.Expression, ErrType, "expected numeric type, received %s", between.String())
		return cast(p0, types.UnknownType)
	}

	return cast(p0, types.BoolType)
}

func (s *sqlAnalyzer) VisitExpressionSubquery(p0 *ExpressionSubquery) any {
	// subquery should return a table
	rel, ok := p0.Subquery.Accept(s).([]*Attribute)
	if !ok {
		panic("expected query to return attributes")
	}

	if len(rel) != 1 {
		s.errs.AddErr(p0, ErrResultShape, "subquery expressions must return exactly 1 column, received %d", len(rel))
		return cast(p0, types.UnknownType)
	}

	if p0.Not || p0.Exists {
		if p0.GetTypeCast() != nil {
			s.errs.AddErr(p0, ErrType, "cannot type cast subquery with EXISTS")
		}
		return types.BoolType
	}

	return cast(p0, rel[0].Type)
}

func (s *sqlAnalyzer) VisitExpressionCase(p0 *ExpressionCase) any {
	// all whens in a case statement must be bool, unless there is an expression
	// that occurs after CASE. In that case, whens all must match the case expression type.
	expectedWhenType := types.BoolType
	if p0.Case != nil {
		caseType, ok := p0.Case.Accept(s).(*types.DataType)
		if !ok {
			return s.expressionTypeErr(p0.Case)
		}

		expectedWhenType = caseType
	}

	// all thens and else must return the same type.
	var returnType *types.DataType
	for _, w := range p0.WhenThen {
		when, ok := w[0].Accept(s).(*types.DataType)
		if !ok {
			return s.expressionTypeErr(w[0])
		}

		if !when.Equals(expectedWhenType) {
			return s.typeErr(w[0], when, expectedWhenType)
		}

		then, ok := w[1].Accept(s).(*types.DataType)
		if !ok {
			return s.expressionTypeErr(w[1])
		}

		if returnType == nil {
			returnType = then
		}

		if !then.Equals(returnType) {
			return s.typeErr(w[1], then, returnType)
		}
	}

	if p0.Else != nil {
		elseType, ok := p0.Else.Accept(s).(*types.DataType)
		if !ok {
			return s.expressionTypeErr(p0.Else)
		}

		if returnType != nil && !elseType.Equals(returnType) {
			return s.typeErr(p0.Else, elseType, returnType)
		}
	}

	return cast(p0, returnType)
}

// The below methods are responsible for manipulating the sql context and identifying
// the resulting relations.

func (s *sqlAnalyzer) VisitCommonTableExpression(p0 *CommonTableExpression) any {
	// check that the table does not already exist
	_, ok := s.sqlCtx.getOuterRelation(p0.Name)
	if ok {
		s.errs.AddErr(p0, ErrTableAlreadyExists, p0.Name)
		return nil
	}

	_, ok = s.schema.FindTable(p0.Name)
	if ok {
		s.errs.AddErr(p0, ErrTableAlreadyExists, p0.Name)
		return nil
	}

	rel, ok := p0.Query.Accept(s).([]*Attribute)
	if !ok {
		// panic because it is an internal error.
		// I guess we could just let it panic without ok,
		// but this is more descriptive.
		panic("expected query to return attributes")
	}

	if len(p0.Columns) != len(rel) {
		s.errs.AddErr(p0, ErrResultShape, "expected %d columns, received %d", len(p0.Columns), len(rel))
		return nil
	}

	// rename the columns and add the relation to the outer scope
	for i, col := range p0.Columns {
		rel[i].Name = col
	}

	s.sqlCtx.outerRelations = append(s.sqlCtx.outerRelations, &Relation{
		Name:       p0.Name,
		Attributes: rel,
	})

	return nil
}

func (s *sqlAnalyzer) VisitSQLStatement(p0 *SQLStatement) any {
	for _, cte := range p0.CTEs {
		cte.Accept(s)
	}

	rel, ok := p0.SQL.Accept(s).([]*Attribute)
	if !ok {
		panic("expected query to return attributes")
	}

	return rel
}

func (s *sqlAnalyzer) VisitSelectStatement(p0 *SelectStatement) any {
	// for each subquery, we need to create a new scope.

	// all select cores will need their own scope. They all also need to have the
	// same shape as each other
	s.sqlCtx.scope()
	rel1, ok := p0.SelectCores[0].Accept(s).([]*Attribute)
	if !ok {
		panic("expected query to return attributes")
	}
	// keep the rel1 scope as we may need to reference the joined
	// tables later.
	rel1Scope := s.sqlCtx.copy()
	s.sqlCtx.popScope()

	isCompound := false
	compoundHasGroupBy := false
	// we visit the rest of the select cores to check the shape
	for _, core := range p0.SelectCores[1:] {
		isCompound = true
		if core.GroupBy != nil {
			compoundHasGroupBy = true
		}

		s.sqlCtx.scope()
		rel2, ok := core.Accept(s).([]*Attribute)
		if !ok {
			panic("expected query to return attributes")
		}
		s.sqlCtx.popScope()

		if !ShapesMatch(rel1, rel2) {
			s.errs.AddErr(core, ErrResultShape, "expected shape to match previous select core")
			return rel1
		}
	}

	// If it is not a compound select, we should use the scope from the first select core,
	// so that we can analyze joined tables in the order and limit clauses. It if is a compound
	// select, then we should flatten all joined tables into a single anonymous table. This can
	// then be referenced in order bys and limits. If there are column conflicts in the flattened column,
	// we should return an error, since there will be no way for us to inform postgres of our default ordering.
	if isCompound {
		// if compound, flatten
		flattened, conflictCol, err := Flatten(rel1Scope.joinedRelations...)
		if err != nil {
			s.errs.AddErr(p0, err, conflictCol)
			return rel1
		}

		// we can simply assign this to the rel1Scope, since we we will not
		// need it past this point. We can add it as an unnamed relation.
		rel1Scope.joinedRelations = []*Relation{{Attributes: flattened}}

		// if a compound select, then we have the following default ordering rules:
		// 1. All columns returned will be ordered in the order they are returned.
		// 2. If the statement includes a group by in one of the select cores, then
		// we throw an error. This is a relic of SQLite's rudimentary referencing, however
		// since it is such an uncommon query anyways, we have decided to not support it
		// until we have time for more testing.
		if compoundHasGroupBy || p0.SelectCores[0].GroupBy != nil {
			s.errs.AddErr(p0, ErrAggregate, "cannot use group by in compound select")
			return rel1
		}

		// order all flattened returns
		for _, attr := range flattened {
			p0.Ordering = append(p0.Ordering, &OrderingTerm{
				Expression: &ExpressionColumn{
					// leave column blank, since we are referencing a column that no
					// longer knows what table it is from due to the compound.
					Column: attr.Name,
				},
			})
		}
	} else {
		// if it is not a compound, then we apply the following default ordering rules (after the user defined):
		// 1. Each primary key for each schema table joined is ordered in ascending order.
		// The tables and columns for all joined tables will be sorted alphabetically.
		// If table aliases are used, they will be used instead of the name.
		// 2. If the select core contains DISTINCT, then the above does not apply, and
		// we order by all columns returned, in the order they are returned.
		// 3. If there is a group by clause, none of the above apply, and instead we order by
		// all columns specified in the group by.
		if p0.SelectCores[0].GroupBy != nil {
			// reset and visit the group by to get the columns
			var colsToOrder [][2]string
			for _, g := range p0.SelectCores[0].GroupBy {
				s.sqlCtx.setTempValuesToZero()
				g.Accept(s)

				if len(s.sqlCtx._columnsOutsideAggregate) > 1 {
					s.errs.AddErr(g, ErrAggregate, "cannot use multiple columns in group by")
					return rel1
				}

				colsToOrder = append(colsToOrder, s.sqlCtx._columnsOutsideAggregate...)
			}

			// order the columns
			for _, col := range colsToOrder {
				p0.Ordering = append(p0.Ordering, &OrderingTerm{
					Expression: &ExpressionColumn{
						Table:  col[0],
						Column: col[1],
					},
				})
			}
		} else if !p0.SelectCores[0].Distinct {
			// if distinct, order by all columns returned
			for _, attr := range rel1 {
				p0.Ordering = append(p0.Ordering, &OrderingTerm{
					Expression: &ExpressionColumn{
						Table:  "",
						Column: attr.Name,
					},
				})
			}
		} else {
			// if not distinct, order by primary keys in all joined tables
			for _, tbl := range order.OrderMap(s.sqlCtx.joinedTables) {
				pks, err := tbl.Value.GetPrimaryKey()
				if err != nil {
					s.errs.AddErr(p0, err, "could not get primary key for table %s", tbl.Key)
				}

				for _, pk := range pks {
					p0.Ordering = append(p0.Ordering, &OrderingTerm{
						Expression: &ExpressionColumn{
							Table:  tbl.Key,
							Column: pk,
						},
					})
				}
			}
		}
	}

	oldScope := *s.sqlCtx
	s.sqlCtx = rel1Scope
	defer func() { s.sqlCtx = &oldScope }()

	// analyze the ordering, limit, and offset
	for _, o := range p0.Ordering {
		o.Accept(s)
	}

	if p0.Limit != nil {
		dt, ok := p0.Limit.Accept(s).(*types.DataType)
		if !ok {
			s.expressionTypeErr(p0.Limit)
			return rel1
		}

		if !dt.IsNumeric() {
			s.errs.AddErr(p0.Limit, ErrType, "expected numeric type, received %s", dt.String())
		}
	}

	if p0.Offset != nil {
		dt, ok := p0.Offset.Accept(s).(*types.DataType)
		if !ok {
			s.expressionTypeErr(p0.Offset)
			return rel1
		}

		if !dt.IsNumeric() {
			s.errs.AddErr(p0.Offset, ErrType, "expected numeric type, received %s", dt.String())
		}
	}

	return rel1
}

// There are some rules for select cores that are necessary for non-determinism:
// 1. If a SELECT is DISTINCT and contains a GROUP BY, we return an error since we cannot
// order it.
// 2. If a result column uses an aggregate function AND there is no GROUP BY, then all
// result columns must be aggregate functions if they reference a column in a table.
// 3. If there is a GROUP BY, then all result columns must be aggregate functions UNLESS
// the column is specified in the GROUP BY
func (s *sqlAnalyzer) VisitSelectCore(p0 *SelectCore) any {
	// we first need to visit the from and join in order to join
	// all tables to the context.
	// we will visit columns last since it will determine our return type.
	p0.From.Accept(s)
	for _, j := range p0.Joins {
		j.Accept(s)
	}

	if p0.Where != nil {
		s.sqlCtx.setTempValuesToZero()
		whereType, ok := p0.Where.Accept(s).(*types.DataType)
		if !ok {
			return s.expressionTypeErr(p0.Where)
		}

		// if it contains an aggregate, throw an error
		if s.sqlCtx._containsAggregate {
			s.errs.AddErr(p0.Where, ErrAggregate, "cannot use aggregate function in WHERE")
			return []*Attribute{}
		}

		if !whereType.Equals(types.BoolType) {
			s.errs.AddErr(p0.Where, ErrType, "expected boolean type, received %s", whereType.String())
		}
	}

	hasGroupBy := false
	// colsInGroupBy tracks the table and column names that are in the group by.
	colsInGroupBy := make(map[[2]string]struct{})
	for _, g := range p0.GroupBy {
		hasGroupBy = true

		// we need to get all columns used in the group by.
		// If more than one column is used per group by, or if an aggregate is
		// used, we return an error.
		s.sqlCtx.setTempValuesToZero()

		// group by return type is not important
		g.Accept(s)

		if s.sqlCtx._containsAggregate {
			s.errs.AddErr(g, ErrAggregate, "cannot use aggregate function in group by")
			return []*Attribute{}
		}
		if len(s.sqlCtx._columnsOutsideAggregate) != 1 {
			s.errs.AddErr(g, ErrAggregate, "group by must reference exactly one column")
			return []*Attribute{}
		}

		_, ok := colsInGroupBy[s.sqlCtx._columnsOutsideAggregate[0]]
		if ok {
			s.errs.AddErr(g, ErrAggregate, "cannot use column in group by more than once")
			return []*Attribute{}
		}
		colsInGroupBy[s.sqlCtx._columnsOutsideAggregate[0]] = struct{}{}

		if p0.Having != nil {
			s.sqlCtx.setTempValuesToZero()
			havingType, ok := p0.Having.Accept(s).(*types.DataType)
			if !ok {
				return s.expressionTypeErr(p0.Having)
			}

			// columns in having must be in the group by if not in aggregate
			for _, col := range s.sqlCtx._columnsOutsideAggregate {
				if _, ok := colsInGroupBy[col]; !ok {
					s.errs.AddErr(p0.Having, ErrAggregate, "column used in having must be in group by")
				}
			}

			if s.sqlCtx._columnInAggregate != nil {
				if _, ok := colsInGroupBy[*s.sqlCtx._columnInAggregate]; !ok {
					s.errs.AddErr(p0.Having, ErrAggregate, "cannot use column in aggregate if not in group by")
				}
			}

			if !havingType.Equals(types.BoolType) {
				s.errs.AddErr(p0.Having, ErrType, "expected boolean type, received %s", havingType.String())
			}
		}
	}

	if hasGroupBy && p0.Distinct {
		s.errs.AddErr(p0, ErrAggregate, "cannot use DISTINCT with GROUP BY")
		return []*Attribute{}
	}

	var res []*Attribute
	for _, c := range p0.Columns {
		// for each result column, we need to check that:
		// IF THERE IS A GROUP BY:
		// 1. if it is an aggregate, then its column is not in the group by
		// 2. for any column that occurs outside of an aggregate, it is also in the group by
		// IF THERE IS NOT A GROUP BY:
		// 3. if there is an aggregate, then it can be the only return column

		// reset to be sure
		s.sqlCtx.setTempValuesToZero()

		attrs, ok := c.Accept(s).([]*Attribute)
		if !ok {
			panic("expected query to return attributes")
		}

		if !hasGroupBy && s.sqlCtx._containsAggregate {
			if len(p0.Columns) != 1 {
				s.errs.AddErr(c, ErrAggregate, "cannot return multiple values in SELECT that uses aggregate function and no group by")
			}
		} else {
			// if column used in aggregate, ensure it is not in group by
			if s.sqlCtx._columnInAggregate != nil {
				if _, ok := colsInGroupBy[*s.sqlCtx._columnInAggregate]; ok {
					s.errs.AddErr(c, ErrAggregate, "cannot use column in aggregate function and in group by")
				}
			}

			// ensure all columns used outside aggregate are in group by
			for _, col := range s.sqlCtx._columnsOutsideAggregate {
				if _, ok := colsInGroupBy[col]; !ok {
					s.errs.AddErr(c, ErrAggregate, "column used outside aggregate must be included in group by")
				}
			}
		}

		var amiguousCol string
		var err error
		res, amiguousCol, err = Coalesce(append(res, attrs...)...)
		if err != nil {
			s.errs.AddErr(c, err, amiguousCol)
			return res
		}
	}

	return res
}

func (s *sqlAnalyzer) VisitRelationTable(p0 *RelationTable) any {
	// table must either be a common table expression, or a table in the schema.
	var rel *Relation
	tbl, ok := s.schema.FindTable(p0.Table)
	if !ok {
		cte, ok := s.sqlCtx.getOuterRelation(p0.Table)
		if !ok {
			s.errs.AddErr(p0, ErrUnknownTable, p0.Table)
			return []*Attribute{}
		}

		rel = cte.Copy()
	} else {
		var err error
		rel, err = tableToRelation(tbl)
		if err != nil {
			s.errs.AddErr(p0, err, "table: %s", p0.Table)
			return []*Attribute{}
		}

		// since we have joined a new table, we need to add it to the joined tables.
		name := p0.Table
		if p0.Alias != "" {
			name = p0.Alias
		}

		err = s.sqlCtx.join(name, tbl)
		if err != nil {
			s.errs.AddErr(p0, err, name)
			return []*Attribute{}
		}
	}

	// if there is an alias, we rename the relation
	if p0.Alias != "" {
		rel.Name = p0.Alias
	}

	err := s.sqlCtx.joinRelation(rel)
	if err != nil {
		s.errs.AddErr(p0, err, p0.Table)
		return []*Attribute{}
	}

	return nil
}

func (s *sqlAnalyzer) VisitRelationSubquery(p0 *RelationSubquery) any {
	relation, ok := p0.Subquery.Accept(s).([]*Attribute)
	if !ok {
		panic("expected query to return attributes")
	}

	// alias is required for subquery joins
	if p0.Alias == "" {
		s.errs.AddErr(p0, ErrUnnamedJoin, "subquery must have an alias")
		return []*Attribute{}
	}

	err := s.sqlCtx.joinRelation(&Relation{
		Name:       p0.Alias,
		Attributes: relation,
	})
	if err != nil {
		s.errs.AddErr(p0, err, p0.Alias)
		return []*Attribute{}
	}

	return nil
}

func (s *sqlAnalyzer) VisitRelationFunctionCall(p0 *RelationFunctionCall) any {
	// the function call here must return []*Attribute
	// this logic is handled in returnProcedureReturnExpr.
	ret, ok := p0.FunctionCall.Accept(s).([]*Attribute)
	if !ok {
		s.errs.AddErr(p0, ErrType, "cannot join procedure that does not return type table")
	}

	// alias is required for function call joins
	if p0.Alias == "" {
		s.errs.AddErr(p0, ErrUnnamedJoin, "function call must have an alias")
		return []*Attribute{}
	}

	err := s.sqlCtx.joinRelation(&Relation{
		Name:       p0.Alias,
		Attributes: ret,
	})
	if err != nil {
		s.errs.AddErr(p0, err, p0.Alias)
		return []*Attribute{}
	}

	return nil
}

func (s *sqlAnalyzer) VisitJoin(p0 *Join) any {
	// to protect against cartesian joins, we:
	// - check that the condition is a comparison expression
	// - check the comparison expression is an equality
	// - check that one side of the expression is a unique column

	compare, ok := p0.On.(*ExpressionComparison)
	if !ok {
		s.errs.AddErr(p0.On, ErrJoin, "join conditions must be comparison expressions")
		return []*Attribute{}
	}

	if compare.Operator != ComparisonOperatorEqual {
		s.errs.AddErr(p0.On, ErrJoin, "join conditions must be use = operator")
		return []*Attribute{}
	}

	// get the cols to check if they are unique
	var cols []*ExpressionColumn
	left, ok := compare.Left.(*ExpressionColumn)
	if ok {
		cols = append(cols, left)
	}
	right, ok := compare.Right.(*ExpressionColumn)
	if ok {
		cols = append(cols, right)
	}

	var hasUnique bool
	var err error
	var msg string
	switch len(cols) {
	case 0:
		s.errs.AddErr(p0.On, ErrJoin, "join condition must have at least one column")
		return []*Attribute{}
	case 1:
		// if there is only one column, we need to check if it is unique
		hasUnique, msg, err = s.sqlCtx.colIsUnique(cols[0].Table, cols[0].Column)
	case 2:
		// if there are two columns, we need to check if one is unique
		hasUnique, msg, err = s.sqlCtx.colIsUnique(cols[0].Table, cols[0].Column)
		if err != nil {
			s.errs.AddErr(p0.On, err, msg)
			return []*Attribute{}
		}

		// if it is unique, we do not have to check the second column
		if hasUnique {
			break
		}

		hasUnique, msg, err = s.sqlCtx.colIsUnique(cols[1].Table, cols[1].Column)
	default:
		panic("expected 1 or 2 columns")
	}
	if err != nil {
		s.errs.AddErr(p0.On, err, msg)
		return []*Attribute{}
	}

	if !hasUnique {
		s.errs.AddErr(p0.On, ErrJoin, "join condition must have at least one unique column")
		return []*Attribute{}
	}

	// call visit on the comparison to perform regular type checking
	p0.Relation.Accept(s)
	dt, ok := p0.On.Accept(s).(*types.DataType)
	if !ok {
		return s.expressionTypeErr(p0.On)
	}

	if !dt.Equals(types.BoolType) {
		s.errs.AddErr(p0.On, ErrType, "expected boolean type for comparison, received %s", dt.String())
	}

	return []*Attribute{}
}

func (s *sqlAnalyzer) VisitUpdateStatement(p0 *UpdateStatement) any {
	s.sqlResult.Mutative = true

	tbl, msg, err := s.joinTableFromSchema(p0.Table, p0.Alias)
	if err != nil {
		s.errs.AddErr(p0, err, msg)
		return nil
	}

	// we visit from and joins first to fill out the context, since those tables can be
	// referenced in the set expression.
	p0.From.Accept(s)
	for _, j := range p0.Joins {
		j.Accept(s)
	}

	for _, set := range p0.SetClause {
		// this calls VisitUpdateSetClause, defined directly below.
		attr := set.Accept(s).(*Attribute)

		// we will see if the table being updated has this column, and if it
		// is of the correct type.
		col, ok := tbl.FindColumn(attr.Name)
		if !ok {
			s.errs.AddErr(set, ErrUnknownColumn, attr.Name)
			continue
		}

		if !col.Type.Equals(attr.Type) {
			s.errs.AddErr(set, ErrType, "expected %s, received %s", col.Type.String(), attr.Type.String())
		}
	}

	whereType, ok := p0.Where.Accept(s).(*types.DataType)
	if !ok {
		s.expressionTypeErr(p0.Where)
		return nil

	}

	if !whereType.Equals(types.BoolType) {
		s.errs.AddErr(p0.Where, ErrType, "expected boolean type, received %s", whereType.String())
		return nil
	}

	return nil
}

// UpdateSetClause will map the updated column to the type it is being
// set to. Since it does not have context as to the table being acted on,
// it is the responsibility of the caller to validate the types. It will simply
// return the column and the type it is being set to, as an attribute.
func (s *sqlAnalyzer) VisitUpdateSetClause(p0 *UpdateSetClause) any {
	dt, ok := p0.Value.Accept(s).(*types.DataType)
	if !ok {
		return s.expressionTypeErr(p0.Value)
	}

	return &Attribute{
		Name: p0.Column,
		Type: dt,
	}
}

// result columns return []*Attribute
func (s *sqlAnalyzer) VisitResultColumnExpression(p0 *ResultColumnExpression) any {
	e, ok := p0.Expression.Accept(s).(*types.DataType)
	if !ok {
		return s.expressionTypeErr(p0.Expression)
	}

	attr := &Attribute{
		Name: p0.Alias,
		Type: e,
	}

	// ResultColumnExpressions always need to have aliases, unless the expression
	// is a column.
	if attr.Name == "" {
		col, ok := p0.Expression.(*ExpressionColumn)
		if !ok {
			s.errs.AddErr(p0, ErrUnnamedResultColumn, "results must either be column references or have an alias")
		}

		attr.Name = col.Column
	}

	return []*Attribute{attr}
}

func (s *sqlAnalyzer) VisitResultColumnWildcard(p0 *ResultColumnWildcard) any {
	// if the table is specified, we need to return all columns from that table.
	if p0.Table != "" {
		tbl, ok := s.sqlCtx.getJoinedRelation(p0.Table)
		if !ok {
			s.errs.AddErr(p0, ErrUnknownTable, p0.Table)
			return []*Attribute{}
		}

		return tbl.Attributes
	}

	// if table is empty, we flatten all joined relations.
	flattened, conflictCol, err := Flatten(s.sqlCtx.joinedRelations...)
	if err != nil {
		s.errs.AddErr(p0, err, conflictCol)
		return []*Attribute{}
	}

	return flattened
}

func (s *sqlAnalyzer) VisitDeleteStatement(p0 *DeleteStatement) any {
	s.sqlResult.Mutative = true

	_, msg, err := s.joinTableFromSchema(p0.Table, p0.Alias)
	if err != nil {
		s.errs.AddErr(p0, err, msg)
		return nil

	}

	p0.From.Accept(s)
	for _, j := range p0.Joins {
		j.Accept(s)
	}

	whereType, ok := p0.Where.Accept(s).(*types.DataType)
	if !ok {
		s.expressionTypeErr(p0.Where)
		return nil

	}

	if !whereType.Equals(types.BoolType) {
		s.errs.AddErr(p0.Where, ErrType, "expected boolean type, received %s", whereType.String())
		return nil

	}

	return nil

}

func (s *sqlAnalyzer) VisitInsertStatement(p0 *InsertStatement) any {
	s.sqlResult.Mutative = true

	tbl, msg, err := s.joinTableFromSchema(p0.Table, p0.Alias)
	if err != nil {
		s.errs.AddErr(p0, err, msg)
		return nil
	}

	// all columns specified need to exist within the table
	// we will keep track of the types of columns in the order
	// they are specified, to match against the values. If columns
	// are not specified, we simply get call the table's columns.
	var colTypes []*types.DataType
	if len(p0.Columns) == 0 {
		for _, col := range tbl.Columns {
			colTypes = append(colTypes, col.Type)
		}
	} else {
		for _, col := range p0.Columns {
			c, ok := tbl.FindColumn(col)
			if !ok {
				s.errs.AddErr(p0, ErrUnknownColumn, col)
				return nil
			}
			colTypes = append(colTypes, c.Type)
		}
	}

	for _, valList := range p0.Values {
		if len(valList) != len(colTypes) {
			s.errs.AddErr(p0, ErrResultShape, "expected %d values, received %d", len(colTypes), len(valList))
			return nil
		}

		for i, val := range valList {
			dt, ok := val.Accept(s).(*types.DataType)
			if !ok {
				s.expressionTypeErr(val)
				return nil
			}

			if !dt.Equals(colTypes[i]) {
				s.errs.AddErr(val, ErrType, "expected %s, received %s", colTypes[i].String(), dt.String())
			}
		}
	}

	if p0.Upsert != nil {
		p0.Upsert.Accept(s)
	}

	return nil

}

// joinTableFromSchema joins a table from the schema to the sql context.
// It will return an error if the table is already joined, or if the table
// is not in the schema. Optionally, an alias can be passed, which will join
// the table with the alias name. If there is an error, it returns the error
// and a message.
func (s *sqlAnalyzer) joinTableFromSchema(table string, alias string) (*types.Table, string, error) {
	tbl, ok := s.schema.FindTable(table)
	if !ok {
		return nil, table, ErrUnknownTable
	}

	name := tbl.Name
	if alias != "" {
		name = alias
	}

	err := s.sqlCtx.join(name, tbl)
	if err != nil {
		return nil, name, err
	}

	return tbl, "", nil
}

func (s *sqlAnalyzer) VisitUpsertClause(p0 *UpsertClause) any {
	// upsert clause can only be called in an insert. Inserts will
	// always have exactly 1 table joined to the context. We will
	// need to retrieve the one table, verify all conflict columns
	// are valid columns, and validate that any DoUpdate clause
	// references a real column and is assigning it to the correct type.
	if len(s.sqlCtx.joinedRelations) != 1 {
		// panicking because this is an internal bug in context scoping
		panic("expected exactly 1 table to be joined in upsert clause")
	}

	rel := s.sqlCtx.joinedRelations[0]
	for _, col := range p0.ConflictColumns {
		_, ok := rel.FindAttribute(col)
		if !ok {
			s.errs.AddErr(p0, ErrUnknownColumn, "conflict column %s", col)
			return nil
		}
	}

	for _, set := range p0.DoUpdate {
		attr := set.Accept(s).(*Attribute)

		foundAttr, ok := rel.FindAttribute(attr.Name)
		if !ok {
			s.errs.AddErr(p0, ErrUnknownColumn, "update column %s", attr.Name)
			continue
		}

		if !foundAttr.Type.Equals(attr.Type) {
			s.errs.AddErr(p0, ErrType, "expected %s, received %s", foundAttr.Type.String(), attr.Type.String())
			return nil
		}
	}

	if p0.ConflictWhere != nil {
		dt, ok := p0.ConflictWhere.Accept(s).(*types.DataType)
		if !ok {
			s.expressionTypeErr(p0.ConflictWhere)
			return nil
		}

		if !dt.Equals(types.BoolType) {
			s.errs.AddErr(p0.ConflictWhere, ErrType, "expected boolean type, received %s", dt.String())
			return nil
		}
	}

	if p0.UpdateWhere != nil {
		dt, ok := p0.UpdateWhere.Accept(s).(*types.DataType)
		if !ok {
			s.expressionTypeErr(p0.UpdateWhere)
			return nil
		}

		if !dt.Equals(types.BoolType) {
			s.errs.AddErr(p0.UpdateWhere, ErrType, "expected boolean type, received %s", dt.String())
			return nil
		}
	}

	return nil
}

func (s *sqlAnalyzer) VisitOrderingTerm(p0 *OrderingTerm) any {
	// visit the expression. We do not have to worry about what
	// it returns though
	p0.Expression.Accept(s)
	return nil
}

// tableToRelation converts a table to a relation.
func tableToRelation(t *types.Table) (*Relation, error) {
	attrs := make([]*Attribute, len(t.Columns))
	for i, col := range t.Columns {
		attrs[i] = &Attribute{
			Name: col.Name,
			Type: col.Type.Copy(),
		}
	}

	return &Relation{
		Name:       t.Name,
		Attributes: attrs,
	}, nil
}

// procedureContext holds context for the procedure analyzer.
type procedureContext struct {
	// procedureDefinition is the definition for the procedure that we are
	// currently analyzing.
	procedureDefinition *types.Procedure
	// activeLoopReceivers track the variable name for the current loop.
	// The innermost nested loop will be at the 0-index. If we are
	// not in a loop, the slice will be empty.
	activeLoopReceivers []string
	// allLoopReceivers tracks all loop receivers that have occurred over the lifetime
	// of the procedure. This is used to generate variables to hold the loop target
	// in plpgsql.
	allLoopReceivers []*loopTargetTracker
	// anonymousReceivers track the data types of procedure return values
	// that the user throws away. In the procedure call
	// `$var1, _, $var2 := proc_that_returns_3_values()`, the underscore is
	// the anonymous receiver. This slice tracks the types for each of the
	// receivers as it encounters them, so that it can generate a throw-away
	// variable in plpgsql
	anonymousReceivers []*types.DataType
}

func newProcedureContext(proc *types.Procedure) *procedureContext {
	return &procedureContext{
		procedureDefinition: proc,
	}
}

// loopTargetTracker is used to track the target of a loop.
type loopTargetTracker struct {
	// name is the variable name of the loop target.
	name *ExpressionVariable
	// dataType is the data type of the loop target.
	// If the loop target is an anonymous variable, then it will be nil.
	dataType *types.DataType
}

// procedureAnalyzer analyes the procedural language. Since the procedural
// language can execute sql statements, it uses the sqlAnalyzer.
type procedureAnalyzer struct {
	sqlAnalyzer
	procCtx *procedureContext
}

// startProcedureAnalyze starts the analysis of a procedure.
func (p *procedureAnalyzer) startSQLAnalyze() {
	p.sqlAnalyzer.startSQLAnalyze()
}

// endProcedureAnalyze ends the analysis of a procedure.
func (p *procedureAnalyzer) endSQLAnalyze(node Node) {
	sqlRes := p.sqlAnalyzer.endSQLAnalyze()
	if sqlRes.Mutative {
		if p.procCtx.procedureDefinition.IsView() {
			p.errs.AddErr(node, ErrViewMutatesState, "SQL statement mutates state in view procedure")
		}
	}
}

var _ Visitor = (*procedureAnalyzer)(nil)

func (p *procedureAnalyzer) VisitProcedureStmtDeclaration(p0 *ProcedureStmtDeclaration) any {
	// we will check if the variable has already been declared, and if so, error.

	if p.variableExists(p0.Variable.String()) {
		p.errs.AddErr(p0, ErrVariableAlreadyDeclared, p0.Variable.String())
		return nil
	}

	p.variables[p0.Variable.String()] = p0.Type

	return nil
}

func (p *procedureAnalyzer) VisitProcedureStmtAssignment(p0 *ProcedureStmtAssignment) any {
	// ensure the variable already exists, and we are assigning the correct type.
	dt, ok := p0.Variable.Accept(p).(*types.DataType)
	if !ok {
		p.expressionTypeErr(p0.Variable)
		return nil
	}

	// since this is only assignment and not declaration, it needs to already have been declared.
	// We do not need to check anonymous variables here because they cannot be assigned to.
	v, ok := p.variables[p0.Variable.String()]
	if !ok {
		p.errs.AddErr(p0, ErrUndeclaredVariable, p0.Variable.String())
		return nil
	}

	if !v.Equals(dt) {
		p.errs.AddErr(p0, ErrType, "expected %s, received %s", v.String(), dt.String())
	}

	return nil
}

func (p *procedureAnalyzer) VisitProcedureStmtDeclareAndAssign(p0 *ProcedureStmtDeclareAndAssign) any {
	// we will check if the variable has already been declared, and if so, error.
	if p.variableExists(p0.Variable.String()) {
		p.errs.AddErr(p0, ErrVariableAlreadyDeclared, p0.Variable.String())
		return nil
	}

	dt, ok := p0.Value.Accept(p).(*types.DataType)
	if !ok {
		p.expressionTypeErr(p0.Value)
		return nil
	}

	// the type can be inferred from the value.
	// If the user explicitly declared a type, the inferred
	// type should match
	if p0.Type != nil {
		if !p0.Type.Equals(dt) {
			p.errs.AddErr(p0, ErrType, "declared type: %s, inferred type: %s", p0.Type.String(), dt.String())
			return nil
		}
	}

	p.variables[p0.Variable.String()] = dt

	return nil
}

func (p *procedureAnalyzer) VisitProcedureStmtCall(p0 *ProcedureStmtCall) any {
	var callReturns []*types.DataType
	// it might return a single value
	returns1, ok := p0.Call.Accept(p).(*types.DataType)
	if ok {
		callReturns = append(callReturns, returns1)
	} else {
		// or it might return multiple values
		returns2, ok := p0.Call.Accept(p).([]*types.DataType)
		if !ok {
			p.errs.AddErr(p0.Call, ErrType, "expected function/procedure to return one or more variables")
			return nil
		}

		callReturns = returns2
	}

	if len(p0.Receivers) != len(callReturns) {
		p.errs.AddErr(p0, ErrResultShape, "function/procedure returns %d value(s), statement has %d receiver(s)", len(callReturns), len(p0.Receivers))
		return nil
	}

	for i, r := range p0.Receivers {
		// if the receiver is nil, we will not assign it to a variable, as it is an
		// anonymous receiver.
		if r == nil {
			p.procCtx.anonymousReceivers = append(p.procCtx.anonymousReceivers, callReturns[i])
			continue
		}

		// ensure the receiver is not already an anonymous variable
		if _, ok := p.anonymousVariables[r.String()]; ok {
			p.errs.AddErr(r, ErrVariableAlreadyDeclared, r.String())
			continue
		}

		// if the variable has been declared, the type must match. otherwise, declare it.
		declaredType, ok := p.variables[r.String()]
		if ok {
			if !declaredType.Equals(callReturns[i]) {
				p.errs.AddErr(r, ErrType, "expected %s, received %s", declaredType.String(), callReturns[i].String())
				continue
			}
		} else {
			p.variables[r.String()] = callReturns[i]
		}
	}

	return nil
}

// This function is a bit convoluted, but it handles a lot of logic. It checks that the loop
// target variable can actually be declared by plpgsql, and then has to allow it to be accessed
// in the current block context. Once we exit the for loop, it has to make it no longer accessible
// in the context, BUT needs to still keep track of it. It needs to keep track of its data type,
// and whether it is a compound type, so that plpgsql knows whether to declare it as a RECORD
// or as a scalar type.
func (p *procedureAnalyzer) VisitProcedureStmtForLoop(p0 *ProcedureStmtForLoop) any {
	// check to make sure the receiver has not already been declared
	if p.variableExists(p0.Receiver.String()) {
		p.errs.AddErr(p0.Receiver, ErrVariableAlreadyDeclared, p0.Receiver.String())
		return nil
	}

	tracker := &loopTargetTracker{
		name: p0.Receiver,
	}

	// get the type from the loop term.
	// can be a scalar if the term is a range or array,
	// and an object if it is a sql statement.
	scalarVal, ok := p0.LoopTerm.Accept(p).(*types.DataType)
	if ok {
		p.variables[p0.Receiver.String()] = scalarVal
		tracker.dataType = scalarVal
	} else {
		compound, ok := p0.LoopTerm.Accept(p).(map[string]*types.DataType)
		if !ok {
			panic("expected loop term to return scalar or compound type")
		}
		p.anonymousVariables[p0.Receiver.String()] = compound
		// we do not set the tracker type here, since it is an anonymous variable.
	}

	// we now need to add the loop target.
	// if it already has been used, we will error.
	for _, t := range p.procCtx.allLoopReceivers {
		if t.name.String() == p0.Receiver.String() {
			p.errs.AddErr(p0.Receiver, ErrVariableAlreadyDeclared, p0.Receiver.String())
			return nil
		}
	}

	p.procCtx.activeLoopReceivers = append([]string{tracker.name.String()}, p.procCtx.activeLoopReceivers...)
	p.procCtx.allLoopReceivers = append(p.procCtx.allLoopReceivers, tracker)

	// we will now visit the statements in the loop.
	for _, stmt := range p0.Body {
		stmt.Accept(p)
	}

	// pop the loop target
	if len(p.procCtx.activeLoopReceivers) == 1 {
		p.procCtx.activeLoopReceivers = nil
	} else {
		p.procCtx.activeLoopReceivers = p.procCtx.activeLoopReceivers[1:]
	}

	if tracker.dataType == nil {
		delete(p.anonymousVariables, p0.Receiver.String())
	} else {
		delete(p.variables, p0.Receiver.String())
	}

	return nil
}

func (p *procedureAnalyzer) VisitLoopTermRange(p0 *LoopTermRange) any {
	// range loops are always integers
	start, ok := p0.Start.Accept(p).(*types.DataType)
	if !ok {
		return p.expressionTypeErr(p0.Start)
	}

	end, ok := p0.End.Accept(p).(*types.DataType)
	if !ok {
		return p.expressionTypeErr(p0.End)
	}

	// the types have to be ints

	if !start.Equals(types.IntType) {
		p.errs.AddErr(p0.Start, ErrType, "expected int, received %s", start.String())
	}

	if !end.Equals(types.IntType) {
		p.errs.AddErr(p0.End, ErrType, "expected int, received %s", end.String())
	}

	return types.IntType
}

func (p *procedureAnalyzer) VisitLoopTermSQL(p0 *LoopTermSQL) any {
	p.startSQLAnalyze()
	rels, ok := p0.Statement.Accept(p).([]*Attribute)
	if !ok {
		panic("expected query to return attributes")
	}
	p.endSQLAnalyze(p0.Statement)

	// we need to convert the attributes into an object
	obj := make(map[string]*types.DataType)
	for _, rel := range rels {
		obj[rel.Name] = rel.Type
	}

	return obj
}

func (p *procedureAnalyzer) VisitLoopTermVariable(p0 *LoopTermVariable) any {
	// we need to ensure the variable exists
	dt, ok := p0.Variable.Accept(p).(*types.DataType)
	if !ok {
		return p.expressionTypeErr(p0.Variable)
	}

	return dt
}

func (p *procedureAnalyzer) VisitProcedureStmtIf(p0 *ProcedureStmtIf) any {
	for _, c := range p0.IfThens {
		c.Accept(p)
	}

	if p0.Else != nil {
		for _, stmt := range p0.Else {
			stmt.Accept(p)
		}
	}

	return nil
}

func (p *procedureAnalyzer) VisitIfThen(p0 *IfThen) any {
	dt, ok := p0.If.Accept(p).(*types.DataType)
	if !ok {
		p.expressionTypeErr(p0.If)
		return nil
	}

	if !dt.Equals(types.BoolType) {
		p.errs.AddErr(p0.If, ErrType, "expected boolean type, received %s", dt.String())
		return nil
	}

	for _, stmt := range p0.Then {
		stmt.Accept(p)
	}

	return nil
}

func (p *procedureAnalyzer) VisitProcedureStmtSQL(p0 *ProcedureStmtSQL) any {
	p.startSQLAnalyze()
	defer p.endSQLAnalyze(p0.SQL)

	_, ok := p0.SQL.Accept(p).([]*Attribute)
	if !ok {
		panic("expected query to return attributes")
	}

	return nil
}

func (p *procedureAnalyzer) VisitProcedureStmtBreak(p0 *ProcedureStmtBreak) any {
	if len(p.procCtx.activeLoopReceivers) == 0 {
		p.errs.AddErr(p0, ErrBreak, "break statement outside of loop")
	}

	return nil
}

func (p *procedureAnalyzer) VisitProcedureStmtReturn(p0 *ProcedureStmtReturn) any {
	if p.procCtx.procedureDefinition.Returns == nil {
		p.errs.AddErr(p0, ErrFunctionSignature, "procedure does not return any values")
		return nil
	}
	returns := p.procCtx.procedureDefinition.Returns

	if p0.SQL != nil {
		if !returns.IsTable {
			p.errs.AddErr(p0, ErrReturn, "procedure expects scalar returns, cannot return SQL statement")
			return nil
		}

		p.startSQLAnalyze()
		defer p.endSQLAnalyze(p0.SQL)

		res, ok := p0.SQL.Accept(p).([]*Attribute)
		if !ok {
			panic("expected query to return attributes")
		}

		if len(res) != len(returns.Fields) {
			p.errs.AddErr(p0, ErrReturn, "expected %d return table columns, received %d", len(returns.Fields), len(res))
			return nil
		}

		// we will compare the return types to the procedure definition
		for i, r := range res {
			retField := returns.Fields[i]
			if !r.Type.Equals(retField.Type) {
				p.errs.AddErr(p0, ErrReturn, "expected column type %s, received column type %s", retField.Type.String(), r.Type.String())
			}

			if r.Name != retField.Name {
				p.errs.AddErr(p0, ErrReturn, "expected column name %s, received column name %s", retField.Name, r.Name)
			}
		}

		return nil
	}
	if returns.IsTable {
		p.errs.AddErr(p0, ErrReturn, "procedure expects table returns, cannot return scalar values")
		return nil
	}

	if len(p0.Values) != len(returns.Fields) {
		p.errs.AddErr(p0, ErrReturn, "expected %d return values, received %d", len(returns.Fields), len(p0.Values))
		return nil
	}

	for i, v := range p0.Values {
		dt, ok := v.Accept(p).(*types.DataType)
		if !ok {
			return p.expressionTypeErr(v)
		}

		if !dt.Equals(returns.Fields[i].Type) {
			p.errs.AddErr(p0, ErrReturn, "expected %s, received %s", returns.Fields[i].Type.String(), dt.String())
		}
	}

	return nil
}

func (p *procedureAnalyzer) VisitProcedureStmtReturnNext(p0 *ProcedureStmtReturnNext) any {
	if p.procCtx.procedureDefinition.Returns == nil {
		p.errs.AddErr(p0, ErrFunctionSignature, "procedure does not return any values")
		return nil
	}

	if !p.procCtx.procedureDefinition.Returns.IsTable {
		p.errs.AddErr(p0, ErrReturn, "procedure expects scalar returns, cannot return next")
		return nil
	}

	if len(p0.Values) != len(p.procCtx.procedureDefinition.Returns.Fields) {
		p.errs.AddErr(p0, ErrReturn, "expected %d return values, received %d", len(p.procCtx.procedureDefinition.Returns.Fields), len(p0.Values))
		return nil
	}

	for i, v := range p0.Values {
		dt, ok := v.Accept(p).(*types.DataType)
		if !ok {
			return p.expressionTypeErr(v)
		}

		if !dt.Equals(p.procCtx.procedureDefinition.Returns.Fields[i].Type) {
			p.errs.AddErr(p0, ErrReturn, "expected %s, received %s", p.procCtx.procedureDefinition.Returns.Fields[i].Type.String(), dt.String())
		}
	}

	return nil
}
