package parse

import (
	"fmt"
	"maps"

	"github.com/kwilteam/kwil-db/core/types"
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

// unknownPosition is used to signal that a position is unknown in the AST because
// it was not present in the parse statement. This is used for when we make modifications
// to the ast, e.g. for enforcing ordering.
func unknownPosition() Position {
	return Position{
		IsSet:     true,
		StartLine: -1,
		StartCol:  -1,
		EndLine:   -1,
		EndCol:    -1,
	}
}

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

// copyVariables copies both the user variables and anonymous variables.
func (b *blockContext) copyVariables() (map[string]*types.DataType, map[string]map[string]*types.DataType) {
	// we do not need to deep copy anonymousVariables because anonymousVariables maps an object name
	// to an objects fields and their data types. The only way to declare an object in Kuneiform
	// is for $row in SELECT ..., the $row will have fields. Since these variables can only be declared once
	// per procedure, we do not need to worry about the object having different fields throughout the
	// procedure.
	return maps.Clone(b.variables), maps.Clone(b.anonymousVariables)
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
	// isInlineAction is true if the visitor is analyzing a SQL expression within an in-line
	// statement in an action
	isInlineAction bool
	// inConflict is true if we are in an ON CONFLICT clause
	inConflict bool
	// targetTable is the name (or alias) of the table being inserted, updated, or deleted to/from.
	// It is not set if we are not in an insert, update, or delete statement.
	targetTable string
	// hasAnonymousTable is true if an unnamed table has been joined. If this is true,
	// it can be the only table joined in a select statement.
	hasAnonymousTable bool
	// inSelect is true if we are in a select statement.
	inSelect bool

	// temp are values that are temporary and not even saved within the same scope.
	// they are used in highly specific contexts, and shouldn't be relied on unless
	// specifically documented. All temp values are denoted with a _.

	// inAggregate is true if we are within an aggregate functions parameters.
	// it should only be used in ExpressionColumn, and set in ExpressionFunctionCall.
	_inAggregate bool
	// containsAggregate is true if the current expression contains an aggregate function.
	// it is set in ExpressionFunctionCall, and accessed/reset in SelectCore.
	_containsAggregate bool
	// containsAggregateWithoutGroupBy is true if the current expression contains an aggregate function,
	// but there is no GROUP BY clause. This is set in SelectCore, and accessed in SelectStatement.
	_containsAggregateWithoutGroupBy bool
	// columnInAggregate is the column found within an aggregate function,
	// comprised of the relation and attribute.
	// It is set in ExpressionColumn, and accessed/reset in
	// SelectCore. It is nil if none are found.
	_columnInAggregate *[2]string
	// columnsOutsideAggregate are columns found outside of an aggregate function.
	// It is set in ExpressionColumn, and accessed/reset in
	// SelectCore
	_columnsOutsideAggregate [][2]string
	// inOrdering is true if we are in an ordering clause
	_inOrdering bool
	// result is the result of a query. It is only set when analyzing the ordering clause
	_result []*Attribute
}

func newSQLContext() sqlContext {
	return sqlContext{
		joinedTables: make(map[string]*types.Table),
	}
}

// setTempValuesToZero resets all temp values to their zero values.
func (s *sqlContext) setTempValuesToZero() {
	s._inAggregate = false
	s._containsAggregate = false
	s._columnInAggregate = nil
	s._columnsOutsideAggregate = nil
	s._inOrdering = false
	s._result = nil
}

// copy copies the sqlContext.
// it does not copy the outer scope.
func (c *sqlContext) copy() sqlContext {
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

	return sqlContext{
		joinedRelations:                  joinedRelations,
		outerRelations:                   outerRelations,
		ctes:                             c.ctes,
		joinedTables:                     c.joinedTables,
		_containsAggregateWithoutGroupBy: c._containsAggregateWithoutGroupBy, // we want to carry this over
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

// the following special table names track table names that mean something in the context of the SQL statement.
const (
	tableExcluded = "excluded"
)

// findAttribute searches for a attribute in the specified relation.
// if the relation is empty, it will search all joined relations.
// It does NOT search the outer relations unless specifically specified;
// this matches Postgres' behavior.
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

		switch len(foundAttrs) {
		case 0:
			return "", nil, column, ErrUnknownColumn
		case 1:
			return relName, foundAttrs[0], "", nil
		default:
			return "", nil, column, ErrAmbiguousColumn
		}
	}

	// if referencing excluded, we should instead look at the target table,
	// since the excluded data will always match the failed insert.
	if relation == tableExcluded {
		// excluded can only be used in an ON CONFLICT clause
		if !c.inConflict {
			return "", nil, relation, fmt.Errorf("%w: excluded table can only be used in an ON CONFLICT clause", ErrInvalidExcludedTable)
		}
		relation = c.targetTable
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

// scope moves the current scope to outer scope,
// and sets the current scope to a new scope.
func (c *sqlContext) scope() {
	c2 := &sqlContext{
		joinedRelations: make([]*Relation, len(c.joinedRelations)),
		outerRelations:  make([]*Relation, len(c.outerRelations)),
		joinedTables:    make(map[string]*types.Table),
		// we do not need to copy ctes since they are not ever modified.
		targetTable:       c.targetTable,
		isInlineAction:    c.isInlineAction,
		inConflict:        c.inConflict,
		inSelect:          c.inSelect,
		hasAnonymousTable: c.hasAnonymousTable,
	}
	// copy all non-temp values
	for i, r := range c.outerRelations {
		c2.outerRelations[i] = r.Copy()
	}

	for i, r := range c.joinedRelations {
		c2.joinedRelations[i] = r.Copy()
	}

	for k, t := range c.joinedTables {
		c2.joinedTables[k] = t.Copy()
	}

	// move joined relations to the outside
	c.outerRelations = append(c.outerRelations, c.joinedRelations...)

	// zero everything else
	c.joinedRelations = nil
	c.joinedTables = make(map[string]*types.Table)
	c.setTempValuesToZero()

	// we do NOT change the inAction, inConflict, or targetTable values,
	// since these apply in all nested scopes.

	// we do not alter inSelect, but we do alter hasAnonymousTable.
	c2.hasAnonymousTable = false

	c2.outerScope = c.outerScope
	c.outerScope = c2
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
	sqlCtx    sqlContext
	sqlResult sqlAnalyzeResult
}

// reset resets the sqlAnalyzer.
func (s *sqlAnalyzer) reset() {
	// we don't need to touch the block context, since it does not change here.
	s.sqlCtx = newSQLContext()
	s.sqlResult = sqlAnalyzeResult{}
}

type sqlAnalyzeResult struct {
	Mutative bool
}

// startSQLAnalyze initializes all fields of the sqlAnalyzer.
func (s *sqlAnalyzer) startSQLAnalyze() {
	s.sqlCtx = sqlContext{
		joinedTables: make(map[string]*types.Table),
	}
}

// endSQLAnalyze is called at the end of the analysis.
func (s *sqlAnalyzer) endSQLAnalyze() *sqlAnalyzeResult {
	res := s.sqlResult
	s.sqlCtx = sqlContext{}
	return &res
}

var _ Visitor = (*sqlAnalyzer)(nil)

// typeErr should be used when a type error is encountered.
// It returns an unknown attribute and adds an error to the error listener.
func (s *sqlAnalyzer) typeErr(node Node, t1, t2 *types.DataType) *types.DataType {
	s.errs.AddErr(node, ErrType, fmt.Sprintf("%s != %s", t1.String(), t2.String()))
	return cast(node, types.UnknownType)
}

// expect is a helper function that expects a certain type, and adds an error if it is not found.
func (s *sqlAnalyzer) expect(node Node, t *types.DataType, expected *types.DataType) {
	if !t.Equals(expected) {
		s.errs.AddErr(node, ErrType, fmt.Sprintf("expected %s, received %s", expected.String(), t.String()))
	}
}

// expectedNumeric is a helper function that expects a numeric type, and adds an error if it is not found.
func (s *sqlAnalyzer) expectedNumeric(node Node, t *types.DataType) {
	if !t.IsNumeric() {
		s.errs.AddErr(node, ErrType, fmt.Sprintf("expected numeric type, received %s", t.String()))
	}
}

// expressionTypeErr should be used if we expect an expression to return a *types.DataType,
// but it returns something else. It will attempt to read the actual type and create an error
// message that is helpful for the end user.
func (s *sqlAnalyzer) expressionTypeErr(e Expression) *types.DataType {

	// prefixMsg is a function used to attempt to infer more information about
	// the error. expressionTypeErr is typically triggered when someone uses a function/procedure
	// with an incompatible return type. prefixMsg will attempt to get the name of the function/procedure
	prefixMsg := func() string {
		msg := "expression"
		if call, ok := e.(ExpressionCall); ok {
			msg = fmt.Sprintf(`function/procedure "%s"`, call.FunctionName())
		}
		return msg
	}

	switch v := e.Accept(s).(type) {
	case *types.DataType:
		// if it is a basic expression returning a scalar (e.g. "'hello'" or "abs(-1)"),
		// or a procedure that returns exactly one scalar value.
		// This should never happen, since expressionTypeErr is called when the expression
		// does not return a *types.DataType.
		panic("api misuse: expressionTypeErr should only be called when the expression does not return a *types.DataType")
	case map[string]*types.DataType:
		// if it is a loop receiver on a select statement (e.g. "for $row in select * from table")
		s.errs.AddErr(e, ErrType, "invalid usage of compound type. you must reference a field using $compound.field notation")
	case []*types.DataType:
		// if it is a procedure than returns several scalar values
		s.errs.AddErr(e, ErrType, "expected %s to return a single value, returns %d values", prefixMsg(), len(v))
	case *returnsTable:
		// if it is a procedure that returns a table
		s.errs.AddErr(e, ErrType, "%s returns table, not scalar values", prefixMsg())
	case nil:
		// if it is a procedure that returns nothing
		s.errs.AddErr(e, ErrType, "%s does not return any value", prefixMsg())
	default:
		// unknown
		s.errs.AddErr(e, ErrType, "internal bug: could not infer expected type")
	}

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

		if !proc.IsView() {
			s.sqlResult.Mutative = true
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

	if s.sqlCtx._inOrdering && s.sqlCtx._inAggregate {
		s.errs.AddErr(p0, ErrOrdering, "cannot use aggregate functions in ORDER BY clause")
		return cast(p0, types.UnknownType)
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

	// callers of this visitor know that a nil return means a function does not
	// return anything. We explicitly return nil instead of a nil *types.DataType
	if returnType == nil {
		return nil
	}

	return cast(p0, returnType)
}

func (s *sqlAnalyzer) VisitExpressionForeignCall(p0 *ExpressionForeignCall) any {
	if s.sqlCtx.isInlineAction {
		s.errs.AddErr(p0, ErrFunctionSignature, "foreign calls are not supported in in-line action statements")
	}

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

		s.expect(ctxArgs, dt, types.TextType)
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
func (s *sqlAnalyzer) returnProcedureReturnExpr(p0 ExpressionCall, procedureName string, ret *types.ProcedureReturn) any {
	// if an expression calls a function, it should return exactly one value or a table.
	if ret == nil {
		if p0.GetTypeCast() != nil {
			s.errs.AddErr(p0, ErrType, "cannot typecast procedure %s because does not return a value", procedureName)
		}
		return nil
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

		return &returnsTable{
			attrs: attrs,
		}
	}

	switch len(ret.Fields) {
	case 0:
		s.errs.AddErr(p0, ErrFunctionSignature, "procedure %s does not return a value", procedureName)
		return cast(p0, types.UnknownType)
	case 1:
		return cast(p0, ret.Fields[0].Type)
	default:
		if p0.GetTypeCast() != nil {
			s.errs.AddErr(p0, ErrType, "cannot type cast multiple return values")
		}

		retVals := make([]*types.DataType, len(ret.Fields))
		for i, f := range ret.Fields {
			retVals[i] = f.Type.Copy()
		}

		return retVals
	}
}

// returnsTable is a special struct returned by returnProcedureReturnExpr when a procedure returns a table.
// It is used internally to detect when a procedure returns a table, so that we can properly throw type errors
// with helpful messages when a procedure returning a table is used in a position where a scalar value is expected.
type returnsTable struct {
	attrs []*Attribute
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
	if s.sqlCtx.isInlineAction {
		s.errs.AddErr(p0, ErrAssignment, "array access is not supported in in-line action statements")
	}

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
	if s.sqlCtx.isInlineAction {
		s.errs.AddErr(p0, ErrAssignment, "array instantiation is not supported in in-line action statements")
	}

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
	if s.sqlCtx.isInlineAction {
		s.errs.AddErr(p0, ErrAssignment, "field access is not supported in in-line action statements")
	}

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
	if s.sqlCtx.isInlineAction {
		s.errs.AddErr(p0, ErrAssignment, "logical expressions are not supported in in-line action statements")
	}

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

	// both must be numeric UNLESS it is a concat
	if p0.Operator == ArithmeticOperatorConcat {
		if !left.Equals(types.TextType) || !right.Equals(types.TextType) {
			// Postgres supports concatenation on non-text types, but we do not,
			// so we give a more descriptive error here.
			// see the note at the top of: https://www.postgresql.org/docs/16.1/functions-string.html
			s.errs.AddErr(p0.Left, ErrType, "concatenation only allowed on text types. received %s and %s", left.String(), right.String())
			return cast(p0, types.UnknownType)
		}
	} else {
		s.expectedNumeric(p0.Left, left)
	}

	// we check this after to return a more helpful error message if
	// the user is not concatenating strings.
	if !left.Equals(right) {
		return s.typeErr(p0.Right, right, left)
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
		s.expectedNumeric(p0.Expression, e)
	case UnaryOperatorNeg:
		s.expectedNumeric(p0.Expression, e)

		if e.Equals(types.Uint256Type) {
			s.errs.AddErr(p0.Expression, ErrType, "cannot negate uint256")
			return cast(p0, types.UnknownType)
		}
	case UnaryOperatorNot:
		s.expect(p0.Expression, e, types.BoolType)
	}

	return cast(p0, e)
}

func (s *sqlAnalyzer) VisitExpressionColumn(p0 *ExpressionColumn) any {
	if s.sqlCtx.isInlineAction {
		s.errs.AddErr(p0, ErrAssignment, "column references are not supported in in-line action statements")
	}

	// there is a special case, where if we are within an ORDER BY clause,
	// we can access columns in the result set. We should search that first
	// before searching all joined tables, as result set columns with conflicting
	// names are given precedence over joined tables.
	if s.sqlCtx._inOrdering && p0.Table == "" {
		attr := findAttribute(s.sqlCtx._result, p0.Column)
		// short-circuit if we find the column, otherwise proceed to normal search
		if attr != nil {
			return cast(p0, attr.Type)
		}
	}

	// if we are in an upsert and the column references a column name in the target table
	// AND the table is not specified, we need to throw an ambiguity error. For conflict tables,
	// the user HAS to specify whether the upsert value is from the existing table or excluded table.
	if s.sqlCtx.inConflict && p0.Table == "" {
		mainTbl, ok := s.sqlCtx.joinedTables[s.sqlCtx.targetTable]
		// if not ok, then we are in a subquery or something else, and we can ignore this check.
		if ok {
			if _, ok = mainTbl.FindColumn(p0.Column); ok {
				s.errs.AddErr(p0, ErrAmbiguousConflictTable, `upsert value is ambigous. specify whether the column is from "%s" or "%s"`, s.sqlCtx.targetTable, tableExcluded)
				return cast(p0, types.UnknownType)

			}
		}
	}

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

var supportedCollations = map[string]struct{}{
	"nocase": {},
}

func (s *sqlAnalyzer) VisitExpressionCollate(p0 *ExpressionCollate) any {
	if s.sqlCtx.isInlineAction {
		s.errs.AddErr(p0, ErrAssignment, "collate is not supported in in-line action statements")
	}

	e, ok := p0.Expression.Accept(s).(*types.DataType)
	if !ok {
		return s.expressionTypeErr(p0.Expression)
	}

	if !e.Equals(types.TextType) {
		return s.typeErr(p0.Expression, e, types.TextType)
	}

	_, ok = supportedCollations[p0.Collation]
	if !ok {
		s.errs.AddErr(p0, ErrCollation, `unsupported collation "%s"`, p0.Collation)
	}

	return cast(p0, e)
}

func (s *sqlAnalyzer) VisitExpressionStringComparison(p0 *ExpressionStringComparison) any {
	if s.sqlCtx.isInlineAction {
		s.errs.AddErr(p0, ErrAssignment, "string comparison is not supported in in-line action statements")
	}

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
	if s.sqlCtx.isInlineAction {
		s.errs.AddErr(p0, ErrAssignment, "IS expression is not supported in in-line action statements")
	}

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
	if s.sqlCtx.isInlineAction {
		s.errs.AddErr(p0, ErrAssignment, "IN expression is not supported in in-line action statements")
	}

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
	if s.sqlCtx.isInlineAction {
		s.errs.AddErr(p0, ErrAssignment, "BETWEEN expression is not supported in in-line action statements")
	}

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

	s.expectedNumeric(p0.Expression, between)

	return cast(p0, types.BoolType)
}

func (s *sqlAnalyzer) VisitExpressionSubquery(p0 *ExpressionSubquery) any {
	if s.sqlCtx.isInlineAction {
		s.errs.AddErr(p0, ErrAssignment, "subquery is not supported in in-line action statements")
	}

	// subquery should return a table
	rel, ok := p0.Subquery.Accept(s).([]*Attribute)
	if !ok {
		panic("expected query to return attributes")
	}

	if len(rel) != 1 {
		s.errs.AddErr(p0, ErrResultShape, "subquery expressions must return exactly 1 column, received %d", len(rel))
		return cast(p0, types.UnknownType)
	}

	if p0.Exists {
		if p0.GetTypeCast() != nil {
			s.errs.AddErr(p0, ErrType, "cannot type cast subquery with EXISTS")
		}
		return types.BoolType
	}

	return cast(p0, rel[0].Type)
}

func (s *sqlAnalyzer) VisitExpressionCase(p0 *ExpressionCase) any {
	if s.sqlCtx.isInlineAction {
		s.errs.AddErr(p0, ErrAssignment, "CASE expression is not supported in in-line action statements")
	}

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

		// if return type is not set, set it to the first then
		if returnType == nil {
			returnType = then
		}
		// if the return type is of type null, we should keep trying
		// to reset until we get a non-null type
		if returnType.EqualsStrict(types.NullType) {
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
	s.sqlCtx.inSelect = true

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

	// we want to re-set the rel1 scope, since it is used in ordering,
	// as well as grouping re-checks if the statement is not a compound select.
	// e.g. "select a, b from t1 union select c, d from t2 order by a"
	oldScope := s.sqlCtx
	s.sqlCtx = rel1Scope
	defer func() { s.sqlCtx = oldScope }()

	// If it is not a compound select, we should use the scope from the first select core,
	// so that we can analyze joined tables in the order and limit clauses. It if is a compound
	// select, then we should flatten all joined tables into a single anonymous table. This can
	// then be referenced in order bys and limits. If there are column conflicts in the flattened column,
	// we should return an error, since there will be no way for us to inform postgres of our default ordering.
	if isCompound {
		// we can simply assign this to the rel1Scope, since we we will not
		// need it past this point. We can add it as an unnamed relation.
		rel1Scope.joinedRelations = []*Relation{{Attributes: rel1}}

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
		for _, attr := range rel1 {
			p0.Ordering = append(p0.Ordering, &OrderingTerm{
				Position: unknownPosition(),
				Expression: &ExpressionColumn{
					Position: unknownPosition(),
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
		// If table aliases are used, they will be used instead of the name. This must include
		// subqueries and function joins; even though those are ordered, they still need to
		// be ordered in the outermost select.
		// see: https://www.reddit.com/r/PostgreSQL/comments/u6icv9/is_original_sort_order_preserve_after_joining/
		// TODO: we can likely make some significant optimizations here by only applying ordering
		// on the outermost query UNLESS aggregates are used in the subquery, but that is a future
		// optimization.
		// 2. If the select core contains DISTINCT, then the above does not apply, and
		// we order by all columns returned, in the order they are returned.
		// 3. If there is a group by clause, none of the above apply, and instead we order by
		// all columns specified in the group by.
		// 4. If there is an aggregate clause with no group by, then no ordering is applied.

		// addressing point 4: if there is an aggregate clause with no group by, then no ordering is applied.
		if s.sqlCtx._containsAggregateWithoutGroupBy {
			// do nothing.
		} else if p0.SelectCores[0].GroupBy != nil {
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
					Position: unknownPosition(),
					Expression: &ExpressionColumn{
						Position: unknownPosition(),
						Table:    col[0],
						Column:   col[1],
					},
				})
			}
		} else if p0.SelectCores[0].Distinct {
			// if distinct, order by all columns returned
			for _, attr := range rel1 {
				p0.Ordering = append(p0.Ordering, &OrderingTerm{
					Position: unknownPosition(),
					Expression: &ExpressionColumn{
						Position: unknownPosition(),
						Table:    "",
						Column:   attr.Name,
					},
				})
			}
		} else {
			// if not distinct, order by primary keys in all joined tables
			for _, rel := range rel1Scope.joinedRelations {
				// if it is a table, we only order by primary key.
				// otherwise, order by all columns.
				tbl, ok := rel1Scope.joinedTables[rel.Name]
				if ok {
					pks, err := tbl.GetPrimaryKey()
					if err != nil {
						s.errs.AddErr(p0, err, "could not get primary key for table %s", rel.Name)
					}

					for _, pk := range pks {
						p0.Ordering = append(p0.Ordering, &OrderingTerm{
							Position: unknownPosition(),
							Expression: &ExpressionColumn{
								Position: unknownPosition(),
								Table:    rel.Name,
								Column:   pk,
							},
						})
					}

					continue
				}

				// if not a table, order by all columns
				for _, attr := range rel.Attributes {
					p0.Ordering = append(p0.Ordering, &OrderingTerm{
						Position: unknownPosition(),
						Expression: &ExpressionColumn{
							Position: unknownPosition(),
							Table:    rel.Name,
							Column:   attr.Name,
						},
					})
				}
			}
		}
	}

	// we need to inform the analyzer that we are in ordering
	s.sqlCtx._inOrdering = true
	s.sqlCtx._result = rel1

	// if the user is trying to order and there is an aggregate without group by, we should throw an error.
	if s.sqlCtx._containsAggregateWithoutGroupBy && len(p0.Ordering) > 0 {
		s.errs.AddErr(p0, ErrAggregate, "cannot use order by with aggregate function without group by")
		return rel1
	}
	// analyze the ordering, limit, and offset
	for _, o := range p0.Ordering {
		o.Accept(s)
	}

	// unset the ordering context
	s.sqlCtx._inOrdering = false
	s.sqlCtx._result = nil

	if p0.Limit != nil {
		dt, ok := p0.Limit.Accept(s).(*types.DataType)
		if !ok {
			s.expressionTypeErr(p0.Limit)
			return rel1
		}

		s.expectedNumeric(p0.Limit, dt)
	}

	if p0.Offset != nil {
		dt, ok := p0.Offset.Accept(s).(*types.DataType)
		if !ok {
			s.expressionTypeErr(p0.Offset)
			return rel1
		}

		s.expectedNumeric(p0.Offset, dt)

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
	if p0.From != nil {
		p0.From.Accept(s)
		for _, j := range p0.Joins {
			j.Accept(s)
		}
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

		s.expect(p0.Where, whereType, types.BoolType)
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
					s.errs.AddErr(p0.Having, ErrAggregate, "column used in having must be in group by, or must be in aggregate function")
				}
			}

			// COMMENTING THIS OUT: if a column is in an aggregate in the having, then it is ok if it is not in the group by
			// if s.sqlCtx._columnInAggregate != nil {
			// 	if _, ok := colsInGroupBy[*s.sqlCtx._columnInAggregate]; !ok {
			// 		s.errs.AddErr(p0.Having, ErrAggregate, "cannot use column in having if not in group by or in aggregate function")
			// 	}
			// }

			s.expect(p0.Having, havingType, types.BoolType)
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
			s.sqlCtx._containsAggregateWithoutGroupBy = true
		} else if hasGroupBy {
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
	if s.sqlCtx.hasAnonymousTable {
		s.errs.AddErr(p0, ErrUnnamedJoin, "statement uses an unnamed subquery or procedure join. to join another table, alias the subquery or procedure")
		return []*Attribute{}
	}

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
	if s.sqlCtx.hasAnonymousTable {
		s.errs.AddErr(p0, ErrUnnamedJoin, "statement uses an unnamed subquery or procedure join. to join another table, alias the subquery or procedure")
		return []*Attribute{}
	}

	relation, ok := p0.Subquery.Accept(s).([]*Attribute)
	if !ok {
		panic("expected query to return attributes")
	}

	// alias is usually required for subquery joins
	if p0.Alias == "" {
		// if alias is not given, then this must be a select and there must be exactly one table joined
		if !s.sqlCtx.inSelect {
			s.errs.AddErr(p0, ErrUnnamedJoin, "joins against subqueries must be aliased")
			return []*Attribute{}
		}

		// must be no relations, since this needs to be the first and only relation
		if len(s.sqlCtx.joinedRelations) != 0 {
			s.errs.AddErr(p0, ErrUnnamedJoin, "joins against subqueries must be aliased")
			return []*Attribute{}
		}

		s.sqlCtx.hasAnonymousTable = true
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
	if s.sqlCtx.hasAnonymousTable {
		s.errs.AddErr(p0, ErrUnnamedJoin, "statement uses an unnamed subquery or procedure join. to join another table, alias the subquery or procedure")
		return []*Attribute{}
	}

	// the function call here must return []*Attribute
	// this logic is handled in returnProcedureReturnExpr.
	ret, ok := p0.FunctionCall.Accept(s).(*returnsTable)
	if !ok {
		s.errs.AddErr(p0, ErrType, "cannot join procedure that does not return type table")
	}

	// alias is usually required for subquery joins
	if p0.Alias == "" {
		// if alias is not given, then this must be a select and there must be exactly one table joined
		if !s.sqlCtx.inSelect {
			s.errs.AddErr(p0, ErrUnnamedJoin, "joins against procedures must be aliased")
			return []*Attribute{}
		}

		// must be no relations, since this needs to be the first and only relation
		if len(s.sqlCtx.joinedRelations) != 0 {
			s.errs.AddErr(p0, ErrUnnamedJoin, "joins against procedures must be aliased")
			return []*Attribute{}
		}

		s.sqlCtx.hasAnonymousTable = true
	}

	err := s.sqlCtx.joinRelation(&Relation{
		Name:       p0.Alias,
		Attributes: ret.attrs,
	})
	if err != nil {
		s.errs.AddErr(p0, err, p0.Alias)
		return []*Attribute{}
	}

	return nil
}

func (s *sqlAnalyzer) VisitJoin(p0 *Join) any {
	// call visit on the comparison to perform regular type checking
	p0.Relation.Accept(s)
	dt, ok := p0.On.Accept(s).(*types.DataType)
	if !ok {
		s.expressionTypeErr(p0.On)
		return nil
	}

	s.expect(p0.On, dt, types.BoolType)

	return nil
}

func (s *sqlAnalyzer) VisitUpdateStatement(p0 *UpdateStatement) any {
	s.sqlResult.Mutative = true

	tbl, msg, err := s.setTargetTable(p0.Table, p0.Alias)
	if err != nil {
		s.errs.AddErr(p0, err, msg)
		return []*Attribute{}
	}

	if p0.From != nil {
		// we visit from and joins first to fill out the context, since those tables can be
		// referenced in the set expression.
		p0.From.Accept(s)
		for _, j := range p0.Joins {
			j.Accept(s)
		}
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
			s.typeErr(set, attr.Type, col.Type)
		}
	}

	if p0.Where != nil {
		whereType, ok := p0.Where.Accept(s).(*types.DataType)
		if !ok {
			s.expressionTypeErr(p0.Where)
			return []*Attribute{}
		}

		s.expect(p0.Where, whereType, types.BoolType)
	}

	return []*Attribute{}
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
		// if returning a column and not aliased, we give it the column name.
		// otherwise, we simply leave the name blank. It will not be referenceable
		if ok {
			attr.Name = col.Column
		}
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

	_, msg, err := s.setTargetTable(p0.Table, p0.Alias)
	if err != nil {
		s.errs.AddErr(p0, err, msg)
		return []*Attribute{}

	}

	if p0.From != nil {
		p0.From.Accept(s)
		for _, j := range p0.Joins {
			j.Accept(s)
		}
	}

	if p0.Where != nil {
		whereType, ok := p0.Where.Accept(s).(*types.DataType)
		if !ok {
			s.expressionTypeErr(p0.Where)
			return []*Attribute{}
		}

		s.expect(p0.Where, whereType, types.BoolType)
	}

	return []*Attribute{}

}

func (s *sqlAnalyzer) VisitInsertStatement(p0 *InsertStatement) any {
	s.sqlResult.Mutative = true

	tbl, msg, err := s.setTargetTable(p0.Table, p0.Alias)
	if err != nil {
		s.errs.AddErr(p0, err, msg)
		return []*Attribute{}
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
				return []*Attribute{}
			}
			colTypes = append(colTypes, c.Type)
		}
	}

	for _, valList := range p0.Values {
		if len(valList) != len(colTypes) {
			s.errs.AddErr(p0, ErrResultShape, "expected %d values, received %d", len(colTypes), len(valList))
			return []*Attribute{}
		}

		for i, val := range valList {
			dt, ok := val.Accept(s).(*types.DataType)
			if !ok {
				s.expressionTypeErr(val)
				return []*Attribute{}
			}

			if !dt.Equals(colTypes[i]) {
				s.typeErr(val, dt, colTypes[i])
			}
		}
	}

	if p0.Upsert != nil {
		s.sqlCtx.inConflict = true
		p0.Upsert.Accept(s)
		s.sqlCtx.inConflict = false
	}

	return []*Attribute{}

}

// setTargetTable joins a table from the schema to the sql context, for
// usage in an insert, delete, or update statement.
// It will return an error if the table is already joined, or if the table
// is not in the schema. Optionally, an alias can be passed, which will join
// the table with the alias name. If there is an error, it returns the error
// and a message. It should only be used in INSERT, DELETE, and UPDATE statements.
func (s *sqlAnalyzer) setTargetTable(table string, alias string) (*types.Table, string, error) {
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

	rel, err := tableToRelation(tbl)
	if err != nil {
		return nil, name, err
	}

	rel.Name = name

	err = s.sqlCtx.joinRelation(rel)
	if err != nil {
		return nil, name, err
	}

	s.sqlCtx.targetTable = name

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
			s.typeErr(set, attr.Type, foundAttr.Type)
			return nil
		}
	}

	if p0.ConflictWhere != nil {
		dt, ok := p0.ConflictWhere.Accept(s).(*types.DataType)
		if !ok {
			s.expressionTypeErr(p0.ConflictWhere)
			return nil
		}

		s.expect(p0.ConflictWhere, dt, types.BoolType)
	}

	if p0.UpdateWhere != nil {
		dt, ok := p0.UpdateWhere.Accept(s).(*types.DataType)
		if !ok {
			s.expressionTypeErr(p0.UpdateWhere)
			return nil
		}

		s.expect(p0.UpdateWhere, dt, types.BoolType)
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

// procedureAnalyzer analyzes the procedural language. Since the procedural
// language can execute sql statements, it uses the sqlAnalyzer.
type procedureAnalyzer struct {
	sqlAnalyzer
	procCtx *procedureContext
	// procResult stores data that the analyzer will return with the parsed procedure.
	// The information is used by the code generator to generate the plpgsql code.
	procResult struct {
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
		// allVariables is a map of all variables declared in the procedure.
		// The key is the variable name, and the value is the data type.
		// This does not include any variable declared by a FOR LOOP.
		allVariables map[string]*types.DataType
	}
}

// markDeclared checks if the variable has been declared in the same procedure body,
// but in a different scope. PLPGSQL cannot handle redeclaration, so we need to check for this.
// It will throw the error in the method, and let the caller continue since this is not a critical
// parsing bug. It will mark the variable as declared if it has not been declared yet.
func (p *procedureAnalyzer) markDeclared(p0 Node, name string, typ *types.DataType) {
	dt, ok := p.procResult.allVariables[name]
	if !ok {
		p.procResult.allVariables[name] = typ
		return
	}

	if !dt.Equals(typ) {
		p.errs.AddErr(p0, ErrCrossScopeDeclaration, `variable %s is declared in a different scope in this procedure as a different type.
		This is not supported.`, name)
	}
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
		return zeroProcedureReturn()
	}

	// TODO: we need to figure out how to undeclare a variable if it is declared in a loop/if block

	p.variables[p0.Variable.String()] = p0.Type
	p.markDeclared(p0.Variable, p0.Variable.String(), p0.Type)

	// now that it is declared, we can visit it
	p0.Variable.Accept(p)

	return zeroProcedureReturn()
}

func (p *procedureAnalyzer) VisitProcedureStmtAssignment(p0 *ProcedureStmtAssign) any {
	// visit the value first to get the data type
	dt, ok := p0.Value.Accept(p).(*types.DataType)
	if !ok {
		p.expressionTypeErr(p0.Value)
		return zeroProcedureReturn()
	}

	// the variable can be either an ExpressionVariable or an ExpressionArrayAccess
	// If it is an ExpressionVariable, we need to declare it

	exprVar, ok := p0.Variable.(*ExpressionVariable)
	if ok {
		_, ok = p.variables[exprVar.String()]
		if !ok {
			// if it does not exist, we can declare it here.
			p.variables[exprVar.String()] = dt
			p.markDeclared(p0.Variable, exprVar.String(), dt)
			return zeroProcedureReturn()
		}
	}

	// the type can be inferred from the value.
	// If the user explicitly declared a type, the inferred
	// type should match
	if p0.Type != nil {
		if !p0.Type.Equals(dt) {
			p.errs.AddErr(p0, ErrType, "declared type: %s, inferred type: %s", p0.Type.String(), dt.String())
			return zeroProcedureReturn()
		}
	}

	// ensure the variable already exists, and we are assigning the correct type.
	dt2, ok := p0.Variable.Accept(p).(*types.DataType)
	if !ok {
		p.expressionTypeErr(p0.Variable)
		return zeroProcedureReturn()
	}

	if !dt2.Equals(dt) {
		p.typeErr(p0, dt2, dt)
	}

	return zeroProcedureReturn()
}

func (p *procedureAnalyzer) VisitProcedureStmtCall(p0 *ProcedureStmtCall) any {
	// we track if sqlResult has already been set to alreadyMutative to avoid throwing
	// an incorrect error below.
	alreadyMutative := p.sqlResult.Mutative

	var callReturns []*types.DataType

	// procedure calls can return many different types of values.
	switch v := p0.Call.Accept(p).(type) {
	case *types.DataType:
		callReturns = []*types.DataType{v}
	case []*types.DataType:
		callReturns = v
	case *returnsTable:
		// if a procedure that returns a table is being called in a
		// procedure, we need to ensure there are no receivers, since
		// it is impossible to assign a table to a variable.
		// we will also not add these to the callReturns, since they are
		// table columns, and not assignable variables
		if len(p0.Receivers) != 0 {
			p.errs.AddErr(p0, ErrResultShape, "procedure returns table, cannot assign to variable(s)")
			return zeroProcedureReturn()
		}
	case nil:
		// do nothing
	default:
		p.expressionTypeErr(p0.Call)
		return zeroProcedureReturn()
	}

	// if calling the `error` function, then this branch will return
	exits := false
	if p0.Call.FunctionName() == "error" {
		exits = true
	}

	// if calling a non-view procedure, the above will set the sqlResult to be mutative
	// if this procedure is a view, we should throw an error.
	if !alreadyMutative && p.sqlResult.Mutative && p.procCtx.procedureDefinition.IsView() {
		p.errs.AddErr(p0, ErrViewMutatesState, `view procedure calls non-view procedure "%s"`, p0.Call.FunctionName())
	}

	// users can discard returns by simply not having receivers.
	// if there are no receivers, we can return early.
	if len(p0.Receivers) == 0 {
		return &procedureStmtResult{
			willReturn: exits,
		}
	}

	// we do not have to capture all return values, but we need to ensure
	// we do not have more receivers than return values.
	if len(p0.Receivers) != len(callReturns) {
		p.errs.AddErr(p0, ErrResultShape, `function/procedure "%s" returns %d value(s), statement expects %d value(s)`, p0.Call.FunctionName(), len(callReturns), len(p0.Receivers))
		return zeroProcedureReturn()
	}

	for i, r := range p0.Receivers {
		// if the receiver is nil, we will not assign it to a variable, as it is an
		// anonymous receiver.
		if r == nil {
			p.procResult.anonymousReceivers = append(p.procResult.anonymousReceivers, callReturns[i])
			continue
		}

		// ensure the receiver is not already an anonymous variable
		if _, ok := p.anonymousVariables[r.String()]; ok {
			p.errs.AddErr(r, ErrVariableAlreadyDeclared, r.String())
			continue
		}

		// if the variable has been declared, the type must match. otherwise, declare it and infer the type.
		declaredType, ok := p.variables[r.String()]
		if ok {
			if !declaredType.Equals(callReturns[i]) {
				p.typeErr(r, callReturns[i], declaredType)
				continue
			}
		} else {
			p.variables[r.String()] = callReturns[i]
			p.markDeclared(r, r.String(), callReturns[i])
		}
	}

	return &procedureStmtResult{
		willReturn: exits,
	}
}

// VisitProcedureStmtForLoop visits a for loop statement.
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
		return zeroProcedureReturn()
	}

	tracker := &loopTargetTracker{
		name: p0.Receiver,
	}

	// get the type from the loop term.
	// can be a scalar if the term is a range or array,
	// and an object if it is a sql statement.
	res := p0.LoopTerm.Accept(p)
	scalarVal, ok := res.(*types.DataType)

	// we copy the variables to ensure that the loop target is only accessible in the loop.
	vars, anonVars := p.copyVariables()
	defer func() {
		p.variables = vars
		p.anonymousVariables = anonVars
	}()

	// we do not mark declared here since these are loop receivers,
	// and they get tracked in a separate slice than other variables.
	if ok {
		// if here, we are looping over an array or range.
		// we need to use the returned type, but remove the IsArray
		rec := scalarVal.Copy()
		rec.IsArray = false
		p.variables[p0.Receiver.String()] = rec
		tracker.dataType = rec
	} else {
		// if we are here, we are looping over a select.
		compound, ok := res.(map[string]*types.DataType)
		if !ok {
			p.expressionTypeErr(p0.LoopTerm)
			return zeroProcedureReturn()
		}
		p.anonymousVariables[p0.Receiver.String()] = compound
		// we do not set the tracker type here, since it is an anonymous variable.
	}

	// we now need to add the loop target.
	// if it already has been used, we will error.
	for _, t := range p.procResult.allLoopReceivers {
		if t.name.String() == p0.Receiver.String() {
			p.errs.AddErr(p0.Receiver, ErrVariableAlreadyDeclared, p0.Receiver.String())
			return zeroProcedureReturn()
		}
	}

	p.procCtx.activeLoopReceivers = append([]string{tracker.name.String()}, p.procCtx.activeLoopReceivers...)
	p.procResult.allLoopReceivers = append(p.procResult.allLoopReceivers, tracker)

	// returns tracks whether this loop is guaranteed to exit.
	returns := false
	canBreakPrematurely := false
	// we will now visit the statements in the loop.
	for _, stmt := range p0.Body {
		res := stmt.Accept(p).(*procedureStmtResult)
		if res.canBreak {
			canBreakPrematurely = true
		}
		if res.willReturn {
			returns = true
		}
	}
	// if it is possible for a for loop to break prematurely, then it is possible
	// that it does not include a return, and so we need to inform the caller
	// that it does not guarantee a return.
	if canBreakPrematurely {
		returns = false
	}

	// pop the loop target
	if len(p.procCtx.activeLoopReceivers) == 1 {
		p.procCtx.activeLoopReceivers = nil
	} else {
		p.procCtx.activeLoopReceivers = p.procCtx.activeLoopReceivers[1:]
	}

	return &procedureStmtResult{
		willReturn: returns,
	}
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

	p.expect(p0.Start, start, types.IntType)
	p.expect(p0.End, end, types.IntType)

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
	canBreak := false

	allThensReturn := true
	for _, c := range p0.IfThens {
		res := c.Accept(p).(*procedureStmtResult)
		if !res.willReturn {
			allThensReturn = false
		}
		if res.canBreak {
			canBreak = true
		}
	}

	// initialize to true, so that if else does not exist, we know we still exit.
	// It gets set to false if we encounter an else block.
	elseReturns := true
	if p0.Else != nil {
		vars, anonVars := p.copyVariables()
		defer func() {
			p.variables = vars
			p.anonymousVariables = anonVars
		}()

		elseReturns = false
		for _, stmt := range p0.Else {
			res := stmt.Accept(p).(*procedureStmtResult)
			if res.willReturn {
				elseReturns = true
			}
			if res.canBreak {
				canBreak = true
			}
		}
	}

	return &procedureStmtResult{
		willReturn: allThensReturn && elseReturns,
		canBreak:   canBreak,
	}
}

func (p *procedureAnalyzer) VisitIfThen(p0 *IfThen) any {
	dt, ok := p0.If.Accept(p).(*types.DataType)
	if !ok {
		p.expressionTypeErr(p0.If)
		return zeroProcedureReturn()
	}

	p.expect(p0.If, dt, types.BoolType)

	canBreak := false
	returns := false

	vars, anonVars := p.copyVariables()
	defer func() {
		p.variables = vars
		p.anonymousVariables = anonVars
	}()

	for _, stmt := range p0.Then {
		res := stmt.Accept(p).(*procedureStmtResult)
		if res.willReturn {
			returns = true
		}
		if res.canBreak {
			canBreak = true
		}
	}

	return &procedureStmtResult{
		willReturn: returns,
		canBreak:   canBreak,
	}
}

func (p *procedureAnalyzer) VisitProcedureStmtSQL(p0 *ProcedureStmtSQL) any {
	p.startSQLAnalyze()
	defer p.endSQLAnalyze(p0.SQL)

	_, ok := p0.SQL.Accept(p).([]*Attribute)
	if !ok {
		panic("expected query to return attributes")
	}

	return zeroProcedureReturn()
}

func (p *procedureAnalyzer) VisitProcedureStmtBreak(p0 *ProcedureStmtBreak) any {
	if len(p.procCtx.activeLoopReceivers) == 0 {
		p.errs.AddErr(p0, ErrBreak, "break statement outside of loop")
	}

	return &procedureStmtResult{
		canBreak: true,
	}
}

func (p *procedureAnalyzer) VisitProcedureStmtReturn(p0 *ProcedureStmtReturn) any {
	if p.procCtx.procedureDefinition.Returns == nil {
		if len(p0.Values) != 0 {
			p.errs.AddErr(p0, ErrFunctionSignature, "procedure does not return any values")
		}
		if p0.SQL != nil {
			p.errs.AddErr(p0, ErrFunctionSignature, "cannot return SQL statement from procedure that does not return any values")
		}
		return &procedureStmtResult{
			willReturn: true,
		}
	}
	returns := p.procCtx.procedureDefinition.Returns

	if p0.SQL != nil {
		if !returns.IsTable {
			p.errs.AddErr(p0, ErrReturn, "procedure expects scalar returns, cannot return SQL statement")
			return &procedureStmtResult{
				willReturn: true,
			}
		}

		p.startSQLAnalyze()
		defer p.endSQLAnalyze(p0.SQL)

		res, ok := p0.SQL.Accept(p).([]*Attribute)
		if !ok {
			panic("expected query to return attributes")
		}

		if len(res) != len(returns.Fields) {
			p.errs.AddErr(p0, ErrReturn, "expected %d return table columns, received %d", len(returns.Fields), len(res))
			return &procedureStmtResult{
				willReturn: true,
			}
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

		return &procedureStmtResult{
			willReturn: true,
		}
	}
	if returns.IsTable {
		p.errs.AddErr(p0, ErrReturn, "procedure expects table returns, cannot return scalar values")
		return &procedureStmtResult{
			willReturn: true,
		}
	}

	if len(p0.Values) != len(returns.Fields) {
		p.errs.AddErr(p0, ErrReturn, "expected %d return values, received %d", len(returns.Fields), len(p0.Values))
		return &procedureStmtResult{
			willReturn: true,
		}
	}

	for i, v := range p0.Values {
		dt, ok := v.Accept(p).(*types.DataType)
		if !ok {
			p.expressionTypeErr(v)
			return &procedureStmtResult{
				willReturn: true,
			}
		}

		if !dt.Equals(returns.Fields[i].Type) {
			p.typeErr(v, dt, returns.Fields[i].Type)
		}
	}

	return &procedureStmtResult{
		willReturn: true,
	}
}

func (p *procedureAnalyzer) VisitProcedureStmtReturnNext(p0 *ProcedureStmtReturnNext) any {
	if p.procCtx.procedureDefinition.Returns == nil {
		p.errs.AddErr(p0, ErrFunctionSignature, "procedure does not return any values")
		return &procedureStmtResult{
			willReturn: true,
		}
	}

	if !p.procCtx.procedureDefinition.Returns.IsTable {
		p.errs.AddErr(p0, ErrReturn, "procedure expects scalar returns, cannot return next")
		return &procedureStmtResult{
			willReturn: true,
		}
	}

	if len(p0.Values) != len(p.procCtx.procedureDefinition.Returns.Fields) {
		p.errs.AddErr(p0, ErrReturn, "expected %d return values, received %d", len(p.procCtx.procedureDefinition.Returns.Fields), len(p0.Values))
		return &procedureStmtResult{
			willReturn: true,
		}
	}

	for i, v := range p0.Values {
		dt, ok := v.Accept(p).(*types.DataType)
		if !ok {
			p.expressionTypeErr(v)
			return &procedureStmtResult{
				willReturn: true,
			}
		}

		if !dt.Equals(p.procCtx.procedureDefinition.Returns.Fields[i].Type) {
			p.typeErr(v, dt, p.procCtx.procedureDefinition.Returns.Fields[i].Type)
		}
	}

	return &procedureStmtResult{
		willReturn: true,
	}
}

// zeroProcedureReturn creates a new procedure return with all 0 values.
func zeroProcedureReturn() *procedureStmtResult {
	return &procedureStmtResult{}
}

// procedureStmtResult is returned from each procedure statement visit.
type procedureStmtResult struct {
	// willReturn is true if the statement contains a return statement that it will
	// always hit. This is used to determine if a path will exit a procedure.
	// it is used to tell whether or not a statement can potentially exit a procedure,
	// since all procedures that have an expected return must always return that value.
	// It only tells us whether or not a return is guaranteed to be hit from a statement.
	// The return types are checked at the point of the return statement.
	willReturn bool
	// canBreak is true if the statement that can break a for loop it is in.
	// For example, an IF statement that breaks a for loop will set canBreak to true.
	canBreak bool
}
