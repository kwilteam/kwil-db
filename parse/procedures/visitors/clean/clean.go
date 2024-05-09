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
	"errors"
	"fmt"
	"strings"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/parse/metadata"
	parser "github.com/kwilteam/kwil-db/parse/procedures/parser"
	"github.com/kwilteam/kwil-db/parse/procedures/visitors/traverse"
	"github.com/kwilteam/kwil-db/parse/sql/sqlanalyzer"
	"github.com/kwilteam/kwil-db/parse/sql/sqlanalyzer/parameters"
	"github.com/kwilteam/kwil-db/parse/sql/tree"
	parseTypes "github.com/kwilteam/kwil-db/parse/types"
	"github.com/kwilteam/kwil-db/parse/util"
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
func CleanProcedure(stmts []parser.Statement, proc *types.Procedure, currentSchema *types.Schema, pgSchemaName string,
	errorListener parseTypes.NativeErrorListener) (params []*types.NamedType, sessionVars map[string]*types.DataType, err error) {
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
		currentSchema:      currentSchema,
		pgSchemaName:       pgSchemaName,
		cleanedSessionVars: map[string]*types.DataType{},
		sqlCanMutate:       !proc.IsView(),
		errs:               errorListener,
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
			c.cleanSQL(ss.Statement, ss.StatementLocation)
		},
		StatementReturn: func(sr *parser.StatementReturn) {
			if sr.SQL != nil {
				c.cleanSQL(sr.SQL, sr.SQLLocation)
				if !returnable(sr.SQL) {
					errorListener.NodeErr(&sr.Node, parseTypes.ParseErrorTypeSemantic,
						errors.New("RETURN statement must return a SELECT or have a RETURNING clause"))
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
			c.cleanSQL(lts.Statement, lts.StatementLocation)

			if !returnable(lts.Statement) {
				errorListener.NodeErr(&lts.Node, parseTypes.ParseErrorTypeSemantic,
					errors.New("LOOP target must be a SELECT or have a RETURNING clause"))
			}
		},
		// expressions
		ExpressionMakeArray: func(ema *parser.ExpressionMakeArray) {
			if len(ema.Values) == 0 {
				errorListener.NodeErr(&ema.Node, parseTypes.ParseErrorTypeSemantic,
					errors.New("ARRAY must have at least one element"))
			}
		},
		ExpressionCall: func(ec *parser.ExpressionCall) {
			// we need to check if this is a built in function,
			// or a user-defined procedure. If it is a procedure,
			// we need to check if it is a view or not.

			ec.Name = strings.ToLower(ec.Name)
			_, ok := metadata.Functions[ec.Name]
			if !ok {
				if c.isForeignProcedure(ec.Name) {
					return
				}

				proc2, err := c.findProcedure(ec.Name)
				if err != nil {
					panic(err)
				}

				if proc.IsView() && !proc2.IsView() {
					errorListener.NodeErr(&ec.Node, parseTypes.ParseErrorTypeSemantic,
						fmt.Errorf(`%w: "%s"`, parseTypes.ErrReadOnlyProcedureCallsMutative, proc2.Name))
				}
			}
		},
		ExpressionForeignCall: func(efc *parser.ExpressionForeignCall) {
			// this must be a locally defined procedure
			proc, err := c.findForeignProcedure(efc.Name)
			if err != nil {
				errorListener.NodeErr(&efc.Node, parseTypes.ParseErrorTypeSemantic, err)
			}

			if len(efc.ContextArgs) != 2 {
				errorListener.NodeErr(&efc.Node, parseTypes.ParseErrorTypeSemantic,
					errors.New("foreign procedure calls must have two context arguments"))
			}

			if len(efc.Arguments) != len(proc.Parameters) {
				errorListener.NodeErr(&efc.Node, parseTypes.ParseErrorTypeSemantic,
					errors.New("foreign procedure calls must have the same number of arguments as foreign procedure definition"))
			}

			efc.Name = strings.ToLower(efc.Name)
		},
		ExpressionVariable: func(ev *parser.ExpressionVariable) {
			c.cleanVar(&ev.Name)
		},
		ExpressionComparison: func(ec *parser.ExpressionComparison) {
			err := ec.Operator.Validate()
			if err != nil {
				// this should never happen
				errorListener.NodeErr(&ec.Node, parseTypes.ParseErrorTypeSyntax, err)
			}
		},
		ExpressionArithmetic: func(ea *parser.ExpressionArithmetic) {
			err := ea.Operator.Validate()
			if err != nil {
				// this should never happen
				errorListener.NodeErr(&ea.Node, parseTypes.ParseErrorTypeSyntax, err)
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
	pgSchemaName       string
	cleanedSessionVars map[string]*types.DataType
	sqlCanMutate       bool
	errs               parseTypes.NativeErrorListener
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
		*n = util.FormatParameterName(r)
		return
	case '@':
		// for contextual parameters, we use postgres's current_setting()
		// feature for setting session variables. For example, @caller
		// is accessed via current_setting('ctx.caller')

		sesVar, ok := metadata.GetSessionVariable(r)
		if !ok {
			panic(fmt.Errorf("%w: %s", parseTypes.ErrUnknownContextualVariable, r))
		}

		// contextual parameter
		*n = util.FormatContextualVariableName(r, sesVar)
		c.cleanedSessionVars[*n] = sesVar
		return
	default:
		// this should never happen
		panic("variable names must start with $ or @")
	}
}

// cleanSQL ensures the SQL AST is valid.
func (c *cleaner) cleanSQL(ast tree.AstNode, ctx *parseTypes.Node) {
	// if errors are encountered during cleaning, we should not continue
	// in order to avoid panics due to the SQL being invalid.
	errLis := c.errs.Child("sql-clean", ctx.StartLine, ctx.StartCol)
	err := sqlanalyzer.CleanAST(ast, c.currentSchema, c.pgSchemaName, errLis)
	if err != nil {
		panic(err)
	}
	if errLis.Err() != nil {
		c.errs.Add(errLis.Errors()...)
		return
	}

	isMutative, err := sqlanalyzer.IsMutative(ast)
	if err != nil {
		panic(err)
	}

	if !c.sqlCanMutate && isMutative {
		c.errs.NodeErr(ctx, parseTypes.ParseErrorTypeSemantic, parseTypes.ErrReadOnlyProcedureContainsDML)
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

	return nil, fmt.Errorf(`%w: "%s"`, parseTypes.ErrUnknownFunctionOrProcedure, name)
}

func (c *cleaner) isForeignProcedure(name string) bool {
	for _, proc := range c.currentSchema.ForeignProcedures {
		if strings.EqualFold(proc.Name, name) {
			return true
		}
	}

	return false
}

// findForeignProcedure finds a foreign procedure based on its name.
// it returns an error if the schema or procedure is not found.
func (c *cleaner) findForeignProcedure(name string) (*types.ForeignProcedure, error) {
	for _, proc := range c.currentSchema.ForeignProcedures {
		if strings.EqualFold(proc.Name, name) {
			return proc, nil
		}
	}

	return nil, fmt.Errorf(`%w: "%s"`, parseTypes.ErrUnknownForeignProcedure, name)
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
