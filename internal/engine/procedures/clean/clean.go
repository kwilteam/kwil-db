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
// as part of the PLPGSQL function's signature.
// It takes the parsed statements, the system schemas, the target procedure,
// the dbid of the current schema, a prefix which it will use to prefix
// postgres session variables, and a set of known postgres session variables and their types.
func CleanProcedure(stmts []parser.Statement, proc *types.Procedure, currentSchema *types.Schema, pgSchemaName, contextPrefix string, knownVars map[string]*types.DataType) (params []*types.NamedType, sessionVars map[string]*types.DataType, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = e.(error)
		}
	}()

	c := cleaner{
		currentProc:        proc,
		sessionPrefix:      contextPrefix,
		currentSchema:      currentSchema,
		knownSessionVars:   knownVars,
		pgSchemaName:       pgSchemaName,
		cleanedSessionVars: map[string]*types.DataType{},
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
		StatementReturnNext: func(srn *parser.StatementReturnNext) {
			c.cleanVar(&srn.Variable)
		},
		StatementProcedureCall: func(spc *parser.StatementProcedureCall) {
			cleanedVars := make([]string, len(spc.Variables))
			for i, arg := range spc.Variables {
				c.cleanVar(&arg)
				cleanedVars[i] = arg
			}
			spc.Variables = cleanedVars

			_, ok := engine.Functions[strings.ToLower(spc.Call.Name)]
			if !ok {
				_, err := c.findProcedure(spc.Call.Name)
				if err != nil {
					panic(err)
				}
			}

			spc.Call.Name = strings.ToLower(spc.Call.Name)
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
			_, ok := engine.Functions[strings.ToLower(ec.Name)]
			if !ok {
				_, err := c.findProcedure(ec.Name)
				if err != nil {
					panic(err)
				}
			}

			ec.Name = strings.ToLower(ec.Name)
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

	switch r[0] {
	case '$':
		// user-defined parameter
		*n = "_param_" + r[1:]
		return
	case '@':
		_, ok := c.knownSessionVars[r[1:]]
		if !ok {
			panic("unknown session variable: " + r[1:])
		}

		// contextual parameter
		*n = fmt.Sprintf("current_setting('%s.%s')", c.sessionPrefix, r[1:])
		c.cleanedSessionVars[*n] = c.knownSessionVars[r[1:]]
		return
	default:
		panic("variable names must start with $ or @")
	}
}

// cleanSQL ensures the SQL AST is valid.
func (c *cleaner) cleanSQL(ast tree.AstNode) {
	err := sqlanalyzer.CleanAST(ast, c.currentSchema.Tables, c.pgSchemaName)
	if err != nil {
		panic(err)
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

	return nil, fmt.Errorf("procedure not found: %s", name)
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
