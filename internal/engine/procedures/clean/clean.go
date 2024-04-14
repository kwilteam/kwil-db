// package clean cleans inputs to ensure they are valid.
// it will enforce:
// - all variables are given a name that will not collide with columns
// - all types referencing the local schema will be fully qualified
// - RETURN queries must return either a SELECT or have a RETURNING clause
// - SQL Loop Terms must be SELECTs (no RETURNING clauses allowed)
// - aliases for type "Schema" fields will be fully qualified
// - all SQL queries must be deterministic

package clean

import (
	"fmt"
	"strings"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/internal/engine"
	"github.com/kwilteam/kwil-db/internal/engine/procedures/traverse"
	"github.com/kwilteam/kwil-db/internal/engine/sqlanalyzer"
	"github.com/kwilteam/kwil-db/internal/engine/sqlanalyzer/parameters"
	parser "github.com/kwilteam/kwil-db/internal/parse/procedure"
	"github.com/kwilteam/kwil-db/internal/parse/sql/tree"
)

// CleanProcedure cleans a procedure to ensure it is valid.
// It will alter the underlying statements to ensure they are valid.
// It will return a list of types and their names that should be declared
// as part of the PLPGSQL function's signature. It also returns a map of
// session variables. These are things like @caller, but renamed to work within
// Postgres.
// It takes the parsed statements, the target procedure and its schema, the pg schema name,
// a prefix which it will used to prefix postgres session variables
// and a set of known postgres session variables and their types.
func CleanProcedure(stmts []parser.Statement, proc *types.Procedure, currentSchema *types.Schema, pgSchemaName, contextPrefix string, knownVars map[string]*types.DataType) (params []*types.NamedType, sessionVars map[string]*types.DataType, err error) {
	defer func() {
		if e := recover(); e != nil {
			var ok bool
			err, ok = e.(error)
			if !ok {
				err = fmt.Errorf("%v", e)
			}

		}
	}()

	c := cleaner{
		currentProc:        proc,
		sessionPrefix:      contextPrefix,
		currentSchema:      currentSchema,
		knownSessionVars:   knownVars,
		pgSchemaName:       pgSchemaName,
		cleanedSessionVars: map[string]*types.DataType{},
		sqlCanMutate:       !proc.IsView(),
	}

	// cleanedParams holds the cleaned procedure parameters.
	cleanedParams := make([]*types.NamedType, len(proc.Parameters))
	for i, param := range proc.Parameters {
		// copy param to avoid modifying the original

		named := &types.NamedType{
			Name: param.Name,
			Type: param.Type.Copy(),
		}

		c.cleanVar(&named.Name)

		cleanedParams[i] = named
	}

	t := traverse.Traverser{
		// statements
		StatementVariableDeclaration: func(svd *parser.StatementVariableDeclaration) {
			c.cleanVar(&svd.Name)
			c.cleanType(svd.Type)
		},
		StatementVariableAssignmentWithDeclaration: func(sva *parser.StatementVariableAssignmentWithDeclaration) {
			c.cleanVar(&sva.Name)
			c.cleanType(sva.Type)
		},
		StatementVariableAssignment: func(sva *parser.StatementVariableAssignment) {
			c.cleanVar(&sva.Name)
		},
		StatementForLoop: func(sfl *parser.StatementForLoop) {
			c.cleanVar(&sfl.Variable)
		},
		StatementSQL: func(ss *parser.StatementSQL) {
			c.cleanSQL(ss.Statement)
		},
		StatementReturn: func(sr *parser.StatementReturn) {
			if sr.SQL != nil {
				c.cleanSQL(sr.SQL)
				if !returnable(sr.SQL) {
					panic("RETURN statement must return a SELECT or have a RETURNING clause")
				}
			}
		},
		StatementProcedureCall: func(spc *parser.StatementProcedureCall) {
			cleanedVars := make([]*string, len(spc.Variables))
			for i, arg := range spc.Variables {
				if arg != nil {
					c.cleanVar(arg)
				}
				cleanedVars[i] = arg
			}
			spc.Variables = cleanedVars
		},
		// loop targets
		LoopTargetSQL: func(lts *parser.LoopTargetSQL) {
			c.cleanSQL(lts.Statement)

			if !returnable(lts.Statement) {
				panic("LOOP target must be a SELECT or have a RETURNING clause")
			}
		},
		// expressions
		ExpressionMakeArray: func(ema *parser.ExpressionMakeArray) {
			if len(ema.Values) == 0 {
				panic("ARRAY must have at least one element")
			}
		},
		ExpressionCall: func(ec *parser.ExpressionCall) {
			// we need to check if this is a built in function,
			// or a user-defined procedure. If it is a procedure,
			// we need to check if it is a view or not.

			ec.Name = strings.ToLower(ec.Name)
			_, ok := engine.Functions[ec.Name]
			if !ok {
				proc2, err := c.findProcedure(ec.Name)
				if err != nil {
					panic(err)
				}

				if proc.IsView() && !proc2.IsView() {
					panic(fmt.Errorf(`%w: "%s"`, engine.ErrReadOnlyProcedureCallsMutative, proc2.Name))
				}
			}
		},
		ExpressionVariable: func(ev *parser.ExpressionVariable) {
			c.cleanVar(&ev.Name)
		},
		ExpressionComparison: func(ec *parser.ExpressionComparison) {
			err := ec.Operator.Validate()
			if err != nil {
				panic(err)
			}
		},
		ExpressionArithmetic: func(ea *parser.ExpressionArithmetic) {
			err := ea.Operator.Validate()
			if err != nil {
				panic(err)
			}
		},
	}

	for _, stmt := range stmts {
		stmt.Accept(&t)
	}

	return cleanedParams, c.cleanedSessionVars, nil
}

type cleaner struct {
	currentSchema      *types.Schema
	currentProc        *types.Procedure
	sessionPrefix      string
	knownSessionVars   map[string]*types.DataType
	pgSchemaName       string
	cleanedSessionVars map[string]*types.DataType
	sqlCanMutate       bool
}

// cleanType ensures a type is fully qualified.
// If the type is not a default type, it will be fully qualified.
// If there is no schema specified, it will be assumed to be the current schema.
// If there is a schema specified, it will use the alias.
func (c *cleaner) cleanType(t *types.DataType) {
	err := t.Clean()
	if err != nil {
		panic(err)
	}
}

// cleanVar ensures a variable name is valid.
func (c *cleaner) cleanVar(n *string) {
	r := strings.ToLower(*n)

	if len(r) == 0 || len(r) > 32 {
		panic("variable names must be between 1 and 32 characters")
	}

	// variables in Kwil are defined with either $ or @, but these
	// are not allowed in plpgsql. Furthermore, plpgsql is very picky with
	// collisions, and will collide variable and column names. Therefore,
	// we remove the prefixes and given them a unique name. Since Kwil
	// enforces all columns to start with a letter, we can use underscores

	switch r[0] {
	case '$':
		// user-defined parameter
		*n = "_param_" + r[1:]
		return
	case '@':
		// for contextual parameters, we use postgres's current_setting()
		// feature for setting session variables. For example, @caller
		// is accessed via current_setting('ctx.caller')

		sesVar, ok := c.knownSessionVars[r[1:]]
		if !ok {
			panic(fmt.Errorf("%w: %s", engine.ErrUnknownContextualVariable, r[1:]))
		}

		// contextual parameter
		*n = fmt.Sprintf("current_setting('%s.%s')", c.sessionPrefix, r[1:])
		c.cleanedSessionVars[*n] = sesVar
		return
	default:
		// this should never happen
		panic("variable names must start with $ or @")
	}
}

// cleanSQL ensures the SQL AST is valid.
func (c *cleaner) cleanSQL(ast tree.AstNode) {
	err := sqlanalyzer.CleanAST(ast, c.currentSchema, c.pgSchemaName)
	if err != nil {
		panic(err)
	}

	isMutative, err := sqlanalyzer.IsMutative(ast)
	if err != nil {
		panic(err)
	}

	if !c.sqlCanMutate && isMutative {
		panic(engine.ErrReadOnlyProcedureContainsDML)
	}

	err = parameters.RenameVariables(ast, func(s string) string {
		c.cleanVar(&s)
		return s
	})
	if err != nil {
		panic(err)
	}
}

// findProcedure finds a procedure based on its name.
// it returns an error if the schema or procedure is not found.
func (c *cleaner) findProcedure(name string) (*types.Procedure, error) {
	for _, proc := range c.currentSchema.Procedures {
		if strings.EqualFold(proc.Name, name) {
			return proc, nil
		}
	}

	return nil, fmt.Errorf(`%w: "%s"`, engine.ErrUnknownFunctionOrProcedure, name)
}

// returnable returns true if the statement
// either has a top-level select, or a RETURNING clause.
func returnable(ast tree.AstNode) bool {
	_, ok := ast.(*tree.SelectStmt)
	if ok {
		return true
	}

	hasReturning := false

	err := ast.Walk(&tree.ImplementedListener{
		FuncEnterReturningClause: func(p0 *tree.ReturningClause) error {
			hasReturning = true
			return nil
		},
	})
	if err != nil {
		panic(err)
	}

	return hasReturning
}
