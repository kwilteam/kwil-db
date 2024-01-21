package tree

import (
	sqlwriter "github.com/kwilteam/kwil-db/parse/sql/tree/sql-writer"
)

// AnySQLFunction is a function that can be used in a SQL statement
// String is a function that takes a slice of Expressions and returns a string of the function invocation
// ex: func(args []Expression) string { return "ABS(" + args[0].ToSQL() + ")" }
// There is one generic String method for AnySQLFunction, and each type (i.e. scalar, aggregate, window, etc) will have its own StringAll method
type AnySQLFunction struct {
	distinct     bool
	FunctionName string
	Min          uint8 // Optional min length of arguments
	Max          uint8 // Optional max length of arguments
}

// types of functions (like scalar, aggregate, window, etc) are extenstions of sqlFunction; this is an interface to accept any of them
type SQLFunction interface {
	AstWalker

	Name() string
	ToString(...Expression) string
	SetDistinct(bool)
}

// SetDistinct sets the distinct flag on the function
func (s *AnySQLFunction) SetDistinct(distinct bool) {
	s.distinct = distinct
}

// buildFunctionString is a helper function to build a function string
// it will write the string as FUNC( fn )
func (s *AnySQLFunction) buildFunctionString(fn func(*sqlwriter.SqlWriter)) string {
	stmt := sqlwriter.NewWriter()
	stmt.WriteString(s.FunctionName)
	stmt.Token.Lparen()
	fn(stmt)
	stmt.Token.Rparen()
	return stmt.String()
}

// Name returns the name of the function
func (s *AnySQLFunction) Name() string {
	return s.FunctionName
}

// String is a generic function that takes a slice of Expressions and returns a string of the function invocation
func (s *AnySQLFunction) string(exprs ...Expression) string {
	if s.Min > 0 && len(exprs) < int(s.Min) {
		panic("too few arguments for function " + s.FunctionName)
	}
	if s.Max > 0 && len(exprs) > int(s.Max) {
		panic("too many arguments for function " + s.FunctionName)
	}

	if len(exprs) == 0 {
		return s.stringAll()
	}

	return s.buildFunctionString(func(stmt *sqlwriter.SqlWriter) {
		stmt.WriteList(len(exprs), func(i int) {
			stmt.WriteString(exprs[i].ToSQL())
		})
	})
}

func (s *AnySQLFunction) ToString(exprs ...Expression) string {
	return s.string(exprs...)
}

// StringAll calls the function with a "*" argument. This is used for COUNT(*), for example
func (s *AnySQLFunction) stringAll() string {
	return s.buildFunctionString(func(stmt *sqlwriter.SqlWriter) {
		stmt.Token.Asterisk()
	})
}

type ScalarFunction struct {
	node

	AnySQLFunction
}

func (s *ScalarFunction) Accept(v AstVisitor) any {
	return v.VisitScalarFunc(s)
}

func (s *ScalarFunction) Walk(w AstListener) error {
	return run(
		w.EnterScalarFunc(s),
		w.ExitScalarFunc(s),
	)
}

func NewScalarFunctionWithGetter(name string, min uint8, max uint8, distinct bool) SQLFunctionGetter {
	return func(pos *Position) SQLFunction {
		return &ScalarFunction{
			AnySQLFunction: AnySQLFunction{
				FunctionName: name,
				Min:          min,
				Max:          max,
				distinct:     distinct,
			},
		}
	}
}

// SQLFunctionGetter is a function that returns a SQLFunction given a position
type SQLFunctionGetter func(pos *Position) SQLFunction

var (
	FunctionABS = ScalarFunction{
		AnySQLFunction: AnySQLFunction{
			FunctionName: "abs",
			Min:          1,
			Max:          1,
		}}
	FunctionERROR = ScalarFunction{
		AnySQLFunction: AnySQLFunction{
			FunctionName: "error",
			Min:          1,
			Max:          1,
		}}
	FunctionLENGTH = ScalarFunction{
		AnySQLFunction: AnySQLFunction{
			FunctionName: "length",
			Min:          1,
			Max:          1,
		}}
	FunctionLOWER = ScalarFunction{
		AnySQLFunction: AnySQLFunction{
			FunctionName: "lower",
			Min:          1,
			Max:          1,
		}}
	FunctionUPPER = ScalarFunction{
		AnySQLFunction: AnySQLFunction{
			FunctionName: "upper",
			Min:          1,
			Max:          1,
		}}
	FunctionFORMAT = ScalarFunction{
		AnySQLFunction: AnySQLFunction{
			FunctionName: "format",
			Min:          1,
		}}
	FunctionABSGetter    = NewScalarFunctionWithGetter("abs", 1, 1, false)
	FunctionFORMATGetter = NewScalarFunctionWithGetter("format", 1, 0, false)
	FunctionErrorGetter  = NewScalarFunctionWithGetter("error", 1, 1, false)
	FunctionLengthGetter = NewScalarFunctionWithGetter("length", 1, 1, false)
	FunctionLOWERGetter  = NewScalarFunctionWithGetter("lower", 1, 1, false)
	FunctionUPPERGetter  = NewScalarFunctionWithGetter("upper", 1, 1, false)
)

// SQLFunctions is a map of all functions of all types
var SQLFunctions = map[string]SQLFunction{
	// Built-In Scalar functions
	"abs":    &FunctionABS,
	"length": &FunctionLENGTH,
	"lower":  &FunctionLOWER,
	"upper":  &FunctionUPPER,
	"format": &FunctionFORMAT,

	// Aggregate functions
	"count": &FunctionCOUNT,
	"sum":   &FunctionSUM,

	// custom
	"error": &FunctionERROR,
}
