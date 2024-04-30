package typing

import (
	"errors"
	"fmt"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/parse/sql/tree"
)

// AnalyzeOptions is a set of options for type analysis.
type AnalyzeOptions struct {
	// BindParams are the bind parameters for the statement.
	BindParams map[string]*types.DataType
	// ArbitraryBinds will treat all bind parameters as unknown,
	// effectively disabling type checking for them.
	ArbitraryBinds bool
	// Qualify will qualify all column references in the statement.
	Qualify bool
	// VerifyProcedures will verify procedure calls in the statement.
	VerifyProcedures bool
	// Schema is the current database schema.
	Schema *types.Schema
}

// AnalyzeTypes will run type analysis on the given statement.
// It will return the relation that will be returned by the statement,
// if any.
func AnalyzeTypes(ast tree.AstNode, tables []*types.Table, options *AnalyzeOptions) (rel *Relation, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic during type analysis: %v", r)
		}
	}()

	if options == nil {
		options = &AnalyzeOptions{
			BindParams: make(map[string]*types.DataType),
			Schema:     &types.Schema{},
		}
	}

	tbls := make(map[string]*Relation)
	for _, t := range tables {
		r := tableToRelation(t)
		tbls[t.Name] = r.Relation
	}

	v := &typeVisitor{
		commonTables: tbls,
		ctes:         make(map[string]struct{}),
		options:      options,
	}

	res := ast.Accept(v)
	fn, ok := res.(returnFunc)
	if !ok {
		return nil, fmt.Errorf("unknown error: could not analyze types")
	}

	return fn(newEvaluationContext())
}

// evaluationContext is a context for evaluating expressions.
// It provides info on the available tables and bind parameters.
// Unlike the visitor, which holds all available tables, the
// context only holds tables that have been joined and are
// accessible in the current scope.
type evaluationContext struct {
	// joinedTables are tables that have been joined in the current scope.
	joinedTables map[string]*Relation
	// joinOrder is the order in which tables have been joined.
	// it matches the keys in joinedTables.
	joinOrder []string

	// outerTables are tables that are accessible in the outer scope.
	// they are not recorded in the joinOrder.
	outerTables map[string]*Relation
}

// findColumn finds a column in the joined tables.
// If the specified table is empty, it will loop through
// all tables to find the column.
// If it does not find the column, or finds several
// columns with the same name, it will return an error.
// It returns the relation the column is from, and the column itself.
func (e *evaluationContext) findColumn(table, column string) (fromRelation string, attribute *QualifiedAttribute, err error) {
	if table != "" {
		cols, ok := e.joinedTables[table]
		if !ok {
			// check outer tables
			cols, ok = e.outerTables[table]
			if !ok {
				return "", nil, fmt.Errorf("table %s not found", table)
			}
		}

		c, ok := cols.Attribute(column)
		if !ok {
			return "", nil, fmt.Errorf("column %s not found in table %s", column, table)
		}

		return table, &QualifiedAttribute{
			Name:      column,
			Attribute: c,
		}, nil
	}

	// if table is empty, loop through all tables
	var found bool
	var foundValue *QualifiedAttribute
	var foundTable string
	for tbl, cols := range e.joinedTables {
		t, ok := cols.Attribute(column)
		if !ok {
			continue
		}

		if found {
			return "", nil, fmt.Errorf(`%w: "%s"`, errAmbiguousColumn, column)
		}

		found = true
		foundValue = &QualifiedAttribute{
			Name:      column,
			Attribute: t,
		}
		foundTable = tbl
	}
	if !found {
		return "", nil, fmt.Errorf(`%w: "%s"`, errColumnNotFound, column)
	}

	return foundTable, foundValue, nil
}

// join joins new relations to the evaluation context.
// If there is a conflicting tabe name, it will return an error.
// If no table name is specified, it will be joined anonymously.
// Column conflicts in anonymous tables will return an error.
func (e *evaluationContext) join(relation *QualifiedRelation) error {
	if relation.Name == "" {
		// ensure an anonymous relation already exists
		if _, ok := e.joinedTables[""]; !ok {
			// if it does not exist, create it and add it to the join order
			e.joinedTables[""] = NewRelation()
			e.joinOrder = append(e.joinOrder, "")
		}

		return e.joinedTables[""].Merge(relation.Relation)
	}

	if _, ok := e.joinedTables[relation.Name]; ok {
		return fmt.Errorf("conflicting table name: %s", relation.Name)
	}
	if _, ok := e.outerTables[relation.Name]; ok {
		return fmt.Errorf("conflicting table name: %s", relation.Name)
	}

	e.joinedTables[relation.Name] = relation.Relation.Copy()
	e.joinOrder = append(e.joinOrder, relation.Name)

	return nil
}

// mergeAnonymousSafe merges an anonymous relation into the current scope.
// it is like joining, but if there is a naming conflict where the type
// of the column is the same, it will not return an error. It will not merge columns that
// make the relation ambiguous.
func (e *evaluationContext) mergeAnonymousSafe(relation *Relation) error {
	anonTbl, ok := e.joinedTables[""]
	if !ok {
		anonTbl = NewRelation()
		e.joinedTables[""] = anonTbl
	}

	// for each column in the new table, check if it is already in ANY
	// of the tables. If not, add it. If so, ensure the types
	// are the same.
	return relation.Loop(func(s string, a *Attribute) error {
		_, attr, err := e.findColumn("", s)
		// if no error, then the column exists, so check the type
		if err == nil {
			if !attr.Type.Equals(a.Type) {
				return fmt.Errorf("conflicting column type in ambiguous column: %s", s)
			}
			return nil
		}
		// if the column is not found, add it
		if errors.Is(err, errColumnNotFound) {
			return anonTbl.AddAttribute(&QualifiedAttribute{
				Name:      s,
				Attribute: a,
			})
		}
		// if it is ambiguous, then it already exists, so
		// we can ignore it
		if errors.Is(err, errAmbiguousColumn) {
			return nil
		}

		return err
	})
}

// loop loops through the joined tables in the evaluation context,
// in the order they were joined.
// Returning an error will stop the loop.
func (e *evaluationContext) Loop(f func(string, *Relation) error) error {
	for _, table := range e.joinOrder {
		t, ok := e.joinedTables[table]
		if !ok {
			panic("table not found during ordered loop")
		}

		err := f(table, t)
		if err != nil {
			return err
		}
	}

	return nil
}

// scope returns a new evaluation context, and moves all joined tables
// to the outerTables map. This is useful for subqueries, where the
// inner scope should not affect the outer scope.
func (e *evaluationContext) scope() *evaluationContext {
	newTables := make(map[string]*Relation)
	for k, v := range e.joinedTables {
		newTables[k] = v.Copy()
	}

	for k, v := range e.outerTables {
		newTables[k] = v.Copy()
	}

	return &evaluationContext{
		joinedTables: make(map[string]*Relation),
		joinOrder:    []string{},
		outerTables:  newTables,
	}
}

// copy returns a copy of the evaluation context.
func (e *evaluationContext) copy() *evaluationContext {
	newTables := make(map[string]*Relation)
	for k, v := range e.joinedTables {
		newTables[k] = v.Copy()
	}

	outerTables := make(map[string]*Relation)
	for k, v := range e.outerTables {
		outerTables[k] = v.Copy()
	}

	return &evaluationContext{
		joinedTables: newTables,
		joinOrder:    append([]string{}, e.joinOrder...),
		outerTables:  outerTables,
	}

}

func newEvaluationContext() *evaluationContext {
	return &evaluationContext{
		joinedTables: make(map[string]*Relation),
		joinOrder:    []string{},
		outerTables:  make(map[string]*Relation),
	}
}
