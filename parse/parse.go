// package parse contains logic for parsing Kuneiform schemas, procedures, actions,
// and SQL.
package parse

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"

	"github.com/antlr4-go/antlr/v4"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/parse/gen"
)

// Parse parses a Kuneiform schema. It will perform syntax, semantic, and type
// analysis, and return any errors.
func Parse(kf []byte) (*types.Schema, error) {
	res, err := ParseAndValidate(kf)
	if err != nil {
		return nil, err
	}

	if res.Err() != nil {
		return nil, res.Err()
	}

	return res.Schema, nil
}

// SchemaParseResult is the result of parsing a schema.
// It returns the resulting schema, as well as any expected errors that occurred during parsing.
// Unexpected errors will not be returned here, but instead in the function returning this type.
type SchemaParseResult struct {
	// Schema is the parsed schema.
	// The schema can be nil if there are errors.
	Schema *types.Schema `json:"schema"`
	// ParseErrs is the error listener that contains all the errors that occurred during parsing.
	ParseErrs ParseErrs `json:"parse_errs,omitempty"`
	// SchemaInfo is the information about the schema.
	SchemaInfo *SchemaInfo `json:"schema_info,omitempty"`
	// ParsedActions is the ASTs of the parsed actions.
	ParsedActions map[string][]ActionStmt `json:"parsed_actions,omitempty"`
	// ParsedProcedures is the ASTs of the parsed procedures.
	ParsedProcedures map[string][]ProcedureStmt `json:"parsed_procedures,omitempty"`
}

func (r *SchemaParseResult) Err() error {
	return r.ParseErrs.Err()
}

// ParseAndValidate parses and validates an entire schema.
// It returns the parsed schema, as well as any errors that occurred during parsing and validation.
// It is meant to be used by parsing tools and the CLI. Most external users should use Parse instead.
func ParseAndValidate(kf []byte) (*SchemaParseResult, error) {
	res, err := ParseSchemaWithoutValidation(kf)
	if err != nil {
		return nil, err
	}

	// if there is a syntax error, we shouldn't continue with validation.
	// We should still return a nil error, as the caller should read the error
	// from the ParseErrs field.
	if res.ParseErrs.Err() != nil {
		return res, nil
	}

	// we clean the schema only after checking for parser errors, since parser errors
	// might be the reason the schema is invalid in the first place.
	err = res.Schema.Clean()
	if err != nil {
		// all clean validations should get caught before this point, however if they don't
		// this will throw an error during parsing, instead of during transaction execution.
		return nil, err
	}

	for _, proc := range res.Schema.Procedures {
		ast := res.ParsedProcedures[proc.Name]
		block := res.SchemaInfo.Blocks[proc.Name]

		procRes, err := analyzeProcedureAST(proc, res.Schema, ast, &block.Position)
		if err != nil {
			return nil, err
		}

		res.ParseErrs.Add(procRes.ParseErrs.Errors()...)
	}

	for _, act := range res.Schema.Actions {
		ast := res.ParsedActions[act.Name]
		actRes, err := analyzeActionAST(act, res.Schema, ast)
		if err != nil {
			return nil, err
		}

		res.ParseErrs.Add(actRes.ParseErrs.Errors()...)
	}

	return res, nil
}

// ParseSchemaWithoutValidation parses a Kuneiform schema.
// It will not perform validations on the actions and procedures.
// Most users should use ParseAndValidate instead.
func ParseSchemaWithoutValidation(kf []byte) (res *SchemaParseResult, err error) {
	errLis, stream, parser, deferFn := setupParser(string(kf), "schema")
	res = &SchemaParseResult{
		ParseErrs:        errLis,
		ParsedActions:    make(map[string][]ActionStmt),
		ParsedProcedures: make(map[string][]ProcedureStmt),
	}

	visitor := newSchemaVisitor(stream, errLis)
	visitor.actions = res.ParsedActions
	visitor.procedures = res.ParsedProcedures

	defer func() {
		err2 := deferFn(recover())
		if err2 != nil {
			err = err2
		}
	}()

	schema, ok := parser.Schema_entry().Accept(visitor).(*types.Schema)
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
	// CompoundVariables are variables that are created in the procedure.
	CompoundVariables map[string]struct{}
	// AnonymousReceivers are the anonymous receivers that are used in the procedure,
	// in the order they appear
	AnonymousReceivers []*types.DataType
}

// ParseProcedure parses a procedure.
// It takes the procedure definition, as well as the schema.
// It performs type and semantic checks on the procedure.
func ParseProcedure(proc *types.Procedure, schema *types.Schema) (res *ProcedureParseResult, err error) {
	return analyzeProcedureAST(proc, schema, nil, &Position{}) // zero position is fine here
}

// analyzeProcedureAST analyzes the AST of a procedure.
// If AST is nil, it will parse it from the provided body. This is useful because ASTs
// with custom error positions can be passed in.
func analyzeProcedureAST(proc *types.Procedure, schema *types.Schema, ast []ProcedureStmt, procPos *Position) (res *ProcedureParseResult, err error) {
	errLis, stream, parser, deferFn := setupParser(proc.Body, "procedure")
	defer func() {
		err2 := deferFn(recover())
		if err2 != nil {
			err = err2
		}
	}()

	res = &ProcedureParseResult{
		ParseErrs:         errLis,
		Variables:         make(map[string]*types.DataType),
		CompoundVariables: make(map[string]struct{}),
	}

	if ast == nil {
		schemaVisitor := newSchemaVisitor(stream, errLis)
		// first parse the body, then visit it.
		res.AST = parser.Procedure_entry().Accept(schemaVisitor).([]ProcedureStmt)
	} else {
		res.AST = ast
	}

	// if there are expected errors, return the parse errors.
	if errLis.Err() != nil {
		return res, nil
	}

	// set the parameters as the initial vars
	vars := makeSessionVars()
	for _, v := range proc.Parameters {
		vars[v.Name] = v.Type
	}

	visitor := &procedureAnalyzer{
		sqlAnalyzer: sqlAnalyzer{
			blockContext: blockContext{
				schema:             schema,
				variables:          vars,
				anonymousVariables: make(map[string]map[string]*types.DataType),
				errs:               errLis,
			},
			sqlCtx: newSQLContext(),
		},
		procCtx: newProcedureContext(proc),
		procResult: struct {
			allLoopReceivers   []*loopTargetTracker
			anonymousReceivers []*types.DataType
			allVariables       map[string]*types.DataType
		}{
			allVariables: make(map[string]*types.DataType),
		},
	}

	// visit the AST
	returns := false
	for _, stmt := range res.AST {
		res := stmt.Accept(visitor).(*procedureStmtResult)
		if res.willReturn {
			returns = true
		}
		// visitor.sqlAnalyzer.reset() // only want to reset mutative
	}

	// if the procedure is expecting a return that is not a table, and it does not guarantee
	// returning a value, we should add an error.
	if proc.Returns != nil && !returns && !proc.Returns.IsTable {
		if len(res.AST) == 0 {
			errLis.AddErr(procPos, ErrReturn, "procedure does not return a value")
		} else {
			errLis.AddErr(res.AST[len(res.AST)-1], ErrReturn, "procedure does not return a value")
		}
	}

	for k, v := range visitor.procResult.allVariables {
		res.Variables[k] = v
	}

	for _, v := range visitor.procResult.allLoopReceivers {
		// if type is nil, it is a compound variable, and we add it to the loop variables
		// if not nil, it is a value, and we add it to the other variables
		if v.dataType == nil {
			res.CompoundVariables[v.name.String()] = struct{}{}
		} else {
			res.Variables[v.name.String()] = v.dataType
		}
	}

	// we also need to add all input variables to the variables list
	for _, v := range proc.Parameters {
		res.Variables[v.Name] = v.Type
	}

	res.AnonymousReceivers = visitor.procResult.anonymousReceivers

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
	parser, errLis, sqlVis, parseVis, deferFn, err := setupSQLParser(sql, schema)

	res = &SQLParseResult{
		ParseErrs: errLis,
	}

	defer func() {
		err2 := deferFn(recover())
		if err2 != nil {
			err = err2
		}
	}()

	res.AST = parser.Sql_entry().Accept(parseVis).(*SQLStatement)

	if errLis.Err() != nil {
		return res, nil
	}

	res.AST.Accept(sqlVis)
	res.Mutative = sqlVis.sqlResult.Mutative

	return res, err
}

// ParseSQLWithoutValidation parses a SQL AST, but does not perform any validation
// or analysis. ASTs returned from this should not be used in production, as they
// might contain errors, and are not deterministically ordered.
func ParseSQLWithoutValidation(sql string, schema *types.Schema) (res *SQLStatement, err error) {
	defer func() {
		err2 := recover()
		if err2 != nil {
			err = fmt.Errorf("panic: %v", err2)
		}
	}()

	parser, errLis, _, parseVis, deferFn, err := setupSQLParser(sql, schema)
	if err != nil {
		return nil, err
	}

	defer func() {
		err2 := deferFn(recover())
		if err2 != nil {
			err = err2
		}
	}()

	res = parser.Sql_entry().Accept(parseVis).(*SQLStatement)

	if errLis.Err() != nil {
		return nil, errLis.Err()
	}

	return res, nil
}

// setupSQLParser sets up the SQL parser.
func setupSQLParser(sql string, schema *types.Schema) (parser *gen.KuneiformParser, errLis *errorListener, sqlVisitor *sqlAnalyzer, parserVisitor *schemaVisitor, deferFn func(any) error, err error) {
	if sql == "" {
		return nil, nil, nil, nil, nil, fmt.Errorf("empty SQL statement")
	}
	// add semicolon to the end of the statement, if it is not there
	if !strings.HasSuffix(sql, ";") {
		sql += ";"
	}

	errLis, stream, parser, deferFn := setupParser(sql, "sql")

	sqlVisitor = &sqlAnalyzer{
		blockContext: blockContext{
			schema:             schema,
			variables:          make(map[string]*types.DataType), // no variables exist for pure SQL calls
			anonymousVariables: make(map[string]map[string]*types.DataType),
			errs:               errLis,
		},
		sqlCtx: newSQLContext(),
	}
	sqlVisitor.sqlCtx.inLoneSQL = true

	parserVisitor = newSchemaVisitor(stream, errLis)

	return parser, errLis, sqlVisitor, parserVisitor, deferFn, err
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
	return analyzeActionAST(action, schema, nil)
}

// analyzeActionAST analyzes the AST of an action.
// If AST is nil, it will parse it from the provided body. This is useful because ASTs
// with custom error positions can be passed in.
func analyzeActionAST(action *types.Action, schema *types.Schema, ast []ActionStmt) (res *ActionParseResult, err error) {
	errLis, stream, parser, deferFn := setupParser(action.Body, "action")

	res = &ActionParseResult{
		ParseErrs: errLis,
	}

	defer func() {
		err2 := deferFn(recover())
		if err2 != nil {
			err = err2
		}
	}()

	if ast == nil {
		schemaVisitor := newSchemaVisitor(stream, errLis)
		res.AST = parser.Action_entry().Accept(schemaVisitor).([]ActionStmt)
	} else {
		res.AST = ast
	}

	if errLis.Err() != nil {
		return res, nil
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

	for _, stmt := range res.AST {
		stmt.Accept(visitor)
		if sqlStmt, ok := stmt.(*ActionStmtSQL); ok {
			sqlStmt.Mutative = visitor.sqlResult.Mutative
		}
		visitor.sqlAnalyzer.reset()
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
	tokens := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	parser = gen.NewKuneiformParser(tokens)
	errLis.toks = tokens

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

// RecursivelyVisitPositions traverses a structure recursively, visiting all position struct types.
// It is used in both parsing tools, as well as in tests.
// WARNING: This function should NEVER be used in consensus, since it is non-deterministic.
func RecursivelyVisitPositions(v any, fn func(GetPositioner)) {
	visitRecursive(reflect.ValueOf(v), reflect.TypeOf((*GetPositioner)(nil)).Elem(), func(v reflect.Value) {
		if v.CanInterface() {
			a := v.Interface().(GetPositioner)
			fn(a)
		}
	})
}

// visitRecursive is a recursive function that visits all types that implement the target interface.
func visitRecursive(v reflect.Value, target reflect.Type, fn func(reflect.Value)) {
	if v.Type().Implements(target) {
		// check if the value is nil
		if !v.IsNil() {
			fn(v)
		}
	}

	switch v.Kind() {
	case reflect.Ptr, reflect.Interface:
		if v.IsNil() {
			return
		}

		visitRecursive(v.Elem(), target, fn)
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			visitRecursive(v.Field(i), target, fn)
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			visitRecursive(v.Index(i), target, fn)
		}
	case reflect.Map:
		for _, key := range v.MapKeys() {
			visitRecursive(v.MapIndex(key), target, fn)
		}
	}
}
