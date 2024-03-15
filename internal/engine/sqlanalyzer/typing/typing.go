package typing

import (
	"fmt"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/internal/engine"
	"github.com/kwilteam/kwil-db/internal/parse/sql/tree"
)

// AnalyzeTypes will run type analysis on the given statement.
// It will return the relation that will be returned by the statement,
// if any.
func AnalyzeTypes(ast tree.AstNode, tables []*types.Table, bindParams map[string]*types.DataType) (rel *engine.Relation, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic during type analysis: %v", r)
		}
	}()

	tbls := make(map[string]*engine.Relation)
	for _, t := range tables {
		r := engine.TableToRelation(t)
		tbls[t.Name] = r.Relation
	}

	v := &typeVisitor{
		commonTables: tbls,
		ctes:         make(map[string]struct{}),
		bindParams:   bindParams,
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
	joinedTables map[string]*engine.Relation
	// joinOrder is the order in which tables have been joined.
	// it matches the keys in joinedTables.
	joinOrder []string

	// outerTables are tables that are accessible in the outer scope.
	// they are not recorded in the joinOrder.
	outerTables map[string]*engine.Relation
}

// findColumn finds a column in the joined tables.
// If the specified table is empty, it will loop through
// all tables to find the column.
// If it does not find the column, or finds several
// columns with the same name, it will return an error.
func (e *evaluationContext) findColumn(table, column string) (*engine.QualifiedAttribute, error) {
	if table != "" {
		cols, ok := e.joinedTables[table]
		if !ok {
			// check outer tables
			cols, ok = e.outerTables[table]
			if !ok {
				return nil, fmt.Errorf("table %s not found", table)
			}
		}

		c, ok := cols.Attribute(column)
		if !ok {
			return nil, fmt.Errorf("column %s not found in table %s", column, table)
		}

		return &engine.QualifiedAttribute{
			Name:      column,
			Attribute: c,
		}, nil
	}

	// if table is empty, loop through all tables
	var found bool
	var foundValue *engine.QualifiedAttribute
	for _, cols := range e.joinedTables {
		t, ok := cols.Attribute(column)
		if !ok {
			continue
		}

		if found {
			return nil, fmt.Errorf("ambiguous column name: %s", column)
		}

		found = true
		foundValue = &engine.QualifiedAttribute{
			Name:      column,
			Attribute: t,
		}
	}
	if !found {
		return nil, fmt.Errorf("column %s not found", column)
	}

	return foundValue, nil
}

// join joins new engine.relations to the evaluation context.
// If there is a conflicting tabe name, it will return an error.
// If no table name is specified, it will be joined anonymously.
// Column conflicts in anonymous tables will return an error.
func (e *evaluationContext) join(relation *engine.QualifiedRelation) error {
	if relation.Name == "" {
		// ensure an anonymous engine.relation already exists
		if _, ok := e.joinedTables[""]; !ok {
			e.joinedTables[""] = engine.NewRelation()
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

// loop loops through the joined tables in the evaluation context,
// in the order they were joined.
// Returning an error will stop the loop.
func (e *evaluationContext) Loop(f func(string, *engine.Relation) error) error {
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
	newTables := make(map[string]*engine.Relation)
	for k, v := range e.joinedTables {
		newTables[k] = v.Copy()
	}

	for k, v := range e.outerTables {
		newTables[k] = v.Copy()
	}

	return &evaluationContext{
		joinedTables: make(map[string]*engine.Relation),
		joinOrder:    []string{},
		outerTables:  newTables,
	}
}

// copy returns a copy of the evaluation context.
func (e *evaluationContext) copy() *evaluationContext {
	newTables := make(map[string]*engine.Relation)
	for k, v := range e.joinedTables {
		newTables[k] = v.Copy()
	}

	outerTables := make(map[string]*engine.Relation)
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
		joinedTables: make(map[string]*engine.Relation),
		joinOrder:    []string{},
		outerTables:  make(map[string]*engine.Relation),
	}
}
