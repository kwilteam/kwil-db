// package parse contains logic for parsing Kuneiform schemas, procedures, actions,
// and SQL.
package parse

import (
	"fmt"
	"runtime"

	"github.com/antlr4-go/antlr/v4"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/parse/gen"
)

// SchemaParseResult is the result of parsing a schema.
// It returns the resulting schema, as well as any expected errors that occurred during parsing.
// Unexpected errors will not be returned here, but instead in the function returning this type.
type SchemaParseResult struct {
	// Schema is the parsed schema.
	// The schema can be nil if there are errors.
	Schema *types.Schema
	// ParseErrs is the error listener that contains all the errors that occurred during parsing.
	ParseErrs ParseErrs
	// SchemaInfo is the information about the schema.
	SchemaInfo *SchemaInfo
}

func (r *SchemaParseResult) Err() error {
	return r.ParseErrs.Err()
}

// ParseAndValidate parses and validates an entire schema.
func ParseAndValidate(kf []byte) (*SchemaParseResult, error) {
	res, err := ParseSchema(kf)
	if err != nil {
		return nil, err
	}

	if res.ParseErrs.Err() != nil {
		return res, nil
	}

	for _, proc := range res.Schema.Procedures {
		procRes, err := ParseProcedure(proc, res.Schema)
		if err != nil {
			return nil, err
		}

		res.ParseErrs.Add(procRes.ParseErrs.Errors()...)
	}

	for _, act := range res.Schema.Actions {
		actRes, err := ParseAction(act, res.Schema)
		if err != nil {
			return nil, err
		}

		res.ParseErrs.Add(actRes.ParseErrs.Errors()...)
	}

	return res, nil
}

// ParseSchema parses a Kuneiform schema.
// TODO: we should delete ParseKuneiform, in favor of ParseSchema.
func ParseSchema(kf []byte) (res *SchemaParseResult, err error) {
	errLis, stream, parser, deferFn := setupParser(string(kf), "schema")

	res = &SchemaParseResult{
		ParseErrs: errLis,
	}

	visitor := newSchemaVisitor(stream, errLis)

	defer func() {
		err = deferFn(recover())
	}()

	schema, ok := parser.Schema().Accept(visitor).(*types.Schema)
	if !ok {
		err = fmt.Errorf("error parsing schema: could not detect return schema. this is likely a bug in the parser")
	}

	res.Schema = schema
	res.SchemaInfo = visitor.schemaInfo

	if errLis.Err() != nil {
		return res, nil
	}

	return res, err
}

// ProcedureParseResult is the result of parsing a procedure.
// It returns the procedure body AST, as well as any errors that occurred during parsing.
// Unexpected errors will not be returned here, but instead in the function returning this type.
type ProcedureParseResult struct {
	// AST is the abstract syntax tree of the procedure.
	AST []ProcedureStmt
	// Errs are the errors that occurred during parsing and analysis.
	// These include syntax errors, type errors, etc.
	ParseErrs ParseErrs
	// Variables are all variables that are used in the procedure.
	Variables map[string]*types.DataType
	// AnonymousVariables are variables that are created in the procedure.
	AnonymousVariables map[string]map[string]*types.DataType
}

// ParseProcedure parses a procedure.
// It takes the procedure definition, as well as the schema.
// It performs type and semantic checks on the procedure.
func ParseProcedure(proc *types.Procedure, schema *types.Schema) (res *ProcedureParseResult, err error) {
	errLis, stream, parser, deferFn := setupParser(proc.Body, "procedure")
	res = &ProcedureParseResult{
		ParseErrs:          errLis,
		Variables:          makeSessionVars(),
		AnonymousVariables: make(map[string]map[string]*types.DataType),
	}

	// set the parameters as the initial vars
	for _, v := range proc.Parameters {
		res.Variables[v.Name] = v.Type
	}

	visitor := &procedureAnalyzer{
		sqlAnalyzer: sqlAnalyzer{
			blockContext: blockContext{
				schema:             schema,
				variables:          res.Variables,
				anonymousVariables: res.AnonymousVariables,
				errs:               errLis,
			},
			sqlCtx: newSQLContext(),
		},
		procCtx: newProcedureContext(proc),
	}

	defer func() {
		err = deferFn(recover())
	}()

	schemaVisitor := newSchemaVisitor(stream, errLis)
	// first parse the body, then visit it.
	res.AST = parser.Procedure_block().Accept(schemaVisitor).([]ProcedureStmt)

	// if there are expected errors, return the parse errors.
	if errLis.Err() != nil {
		return res, nil
	}

	// visit the AST
	for _, stmt := range res.AST {
		stmt.Accept(visitor)
	}

	return res, err
}

// SQLParseResult is the result of parsing an SQL statement.
// It returns the SQL AST, as well as any errors that occurred during parsing.
// Unexpected errors will not be returned here, but instead in the function returning this type.
type SQLParseResult struct {
	// AST is the abstract syntax tree of the SQL statement.
	AST *SQLStatement
	// Errs are the errors that occurred during parsing and analysis.
	// These include syntax errors, type errors, etc.
	ParseErrs ParseErrs

	// Mutative is true if the statement mutates state.
	Mutative bool
}

// ParseSQL parses an SQL statement.
// It requires a schema to be passed in, since SQL statements may reference
// schema objects.
func ParseSQL(sql string, schema *types.Schema) (res *SQLParseResult, err error) {
	errLis, stream, parser, deferFn := setupParser(sql, "sql")

	res = &SQLParseResult{
		ParseErrs: errLis,
	}

	visitor := &sqlAnalyzer{
		blockContext: blockContext{
			schema:             schema,
			variables:          make(map[string]*types.DataType), // no variables exist for pure SQL calls
			anonymousVariables: make(map[string]map[string]*types.DataType),
			errs:               errLis,
		},
		sqlCtx: newSQLContext(),
	}

	defer func() {
		err = deferFn(recover())
	}()

	schemaVisitor := newSchemaVisitor(stream, errLis)

	res.AST = parser.Sql().Accept(schemaVisitor).(*SQLStatement)

	if errLis.Err() != nil {
		return res, nil
	}

	res.AST.Accept(visitor)
	res.Mutative = visitor.sqlResult.Mutative

	return res, err
}

// ActionParseResult is the result of parsing an action.
// It returns the action body AST, as well as any errors that occurred during parsing.
// Unexpected errors will not be returned here, but instead in the function returning this type.
type ActionParseResult struct {
	AST []ActionStmt
	// Errs are the errors that occurred during parsing and analysis.
	// These include syntax errors, type errors, etc.
	ParseErrs ParseErrs
}

// ParseAction parses a Kuneiform action.
// It requires a schema to be passed in, since actions may reference
// schema objects.
func ParseAction(action *types.Action, schema *types.Schema) (res *ActionParseResult, err error) {
	errLis, stream, parser, deferFn := setupParser(action.Body, "action")

	res = &ActionParseResult{
		ParseErrs: errLis,
	}

	vars := makeSessionVars()
	for _, v := range action.Parameters {
		vars[v] = types.UnknownType
	}

	visitor := &actionAnalyzer{
		sqlAnalyzer: sqlAnalyzer{
			blockContext: blockContext{
				schema:             schema,
				variables:          vars,
				anonymousVariables: make(map[string]map[string]*types.DataType),
				errs:               errLis,
			},
			sqlCtx: newSQLContext(),
		},
		schema: schema,
	}

	defer func() {
		err = deferFn(recover())
	}()

	schemaVisitor := newSchemaVisitor(stream, errLis)

	res.AST = parser.Action_block().Accept(schemaVisitor).([]ActionStmt)

	if errLis.Err() != nil {
		return res, nil
	}

	for _, stmt := range res.AST {
		stmt.Accept(visitor)
	}

	return res, err
}

// setupParser sets up the necessary antlr objects for parsing.
// It returns an error listener, an input stream, a parser, and a function that
// handles returned errors. The function should be called within deferred panic catch.
// The deferFn will decide whether errors should be swallowed based on the error listener.
func setupParser(inputStream string, errLisName string) (errLis *errorListener,
	stream *antlr.InputStream, parser *gen.KuneiformParser, deferFn func(any) error) {
	errLis = newErrorListener(errLisName)
	stream = antlr.NewInputStream(inputStream)

	lexer := gen.NewKuneiformLexer(stream)
	parser = gen.NewKuneiformParser(antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel))

	// remove defaults
	lexer.RemoveErrorListeners()
	parser.RemoveErrorListeners()
	lexer.AddErrorListener(errLis)
	parser.AddErrorListener(errLis)

	parser.BuildParseTrees = true

	deferFn = func(e any) (err error) {
		if e != nil {
			var ok bool
			err, ok = e.(error)
			if !ok {
				err = fmt.Errorf("panic: %v", e)
			}
		}

		// if there is a panic, it may be due to a syntax error.
		// therefore, we should check for syntax errors first and if
		// any occur, swallow the panic and return the syntax errors.
		// If the issue persists past syntax errors, the else block
		// will return the error.
		if errLis.Err() != nil {
			return nil
		} else if err != nil {
			// stack trace since this
			buf := make([]byte, 1<<16)

			stackSize := runtime.Stack(buf, false)
			err = fmt.Errorf("%w\n\n%s", err, buf[:stackSize])

			return err
		}

		return nil
	}

	return errLis, stream, parser, deferFn
}
