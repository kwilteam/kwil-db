package tree

// AnySQLFunction is a function that can be used in a SQL statement
// String is a function that takes a slice of Expressions and returns a string of the function invocation
// ex: func(args []Expression) string { return "ABS(" + args[0].ToSQL() + ")" }
// There is one generic String method for AnySQLFunction, and each type (i.e. scalar, aggregate, window, etc) will have its own StringAll method
type AnySQLFunction struct {
	FunctionName string
	Min          uint8 // Optional min length of arguments
	Max          uint8 // Optional max length of arguments
}

// types of functions (like scalar, aggregate, window, etc) are extenstions of sqlFunction; this is an interface to accept any of them
type SQLFunction interface {
	String([]Expression) string

	// StringAll calls the function with a "*" argument. This is used for COUNT(*), for example
	StringAll() string
}

// buildFunctionString is a helper function to build a function string
// it will write the string as FUNC( fn )
func (s *AnySQLFunction) buildFunctionString(fn func(*sqlBuilder)) string {
	stmt := newSQLBuilder()
	stmt.Write(SPACE)
	stmt.WriteString(s.FunctionName)
	stmt.WriteString("(")
	fn(stmt)
	stmt.WriteString(")")
	stmt.Write(SPACE)
	return stmt.String()
}

// String is a generic function that takes a slice of Expressions and returns a string of the function invocation
func (s *AnySQLFunction) String(exprs []Expression) string {
	if s.Min > 0 && len(exprs) < int(s.Min) {
		panic("too few arguments for function " + s.FunctionName)
	}
	if s.Max > 0 && len(exprs) > int(s.Max) {
		panic("too many arguments for function " + s.FunctionName)
	}

	return s.buildFunctionString(func(stmt *sqlBuilder) {
		for i, expr := range exprs {
			if i > 0 && i < len(exprs) {
				stmt.Write(COMMA, SPACE)
			}
			stmt.WriteString(expr.ToSQL())
		}
	})
}

// StringAll calls the function with a "*" argument. This is used for COUNT(*), for example
func (s *AnySQLFunction) StringAll() string {
	return s.buildFunctionString(func(stmt *sqlBuilder) {
		stmt.Write(ASTERISK)
	})
}

type ScalarFunction struct {
	AnySQLFunction
}

// This will panic if called, since "*" is not a valid argument for a scalar function
func (s *ScalarFunction) StringAll() string {
	panic("cannot use '*' as an input for scalar function " + s.FunctionName)
}

var (
	FunctionABS = ScalarFunction{AnySQLFunction{
		FunctionName: "abs",
		Min:          1,
		Max:          1,
	}}
	FunctionCOALESCE = ScalarFunction{AnySQLFunction{
		FunctionName: "coalesce",
		Min:          2,
	}}
	FunctionFORMAT = ScalarFunction{AnySQLFunction{
		FunctionName: "format",
		Min:          1,
	}}
	FunctionGLOB = ScalarFunction{AnySQLFunction{
		FunctionName: "glob",
		Min:          2,
		Max:          2,
	}}
	FunctionHEX = ScalarFunction{AnySQLFunction{
		FunctionName: "hex",
		Min:          1,
		Max:          1,
	}}
	FunctionIFNULL = ScalarFunction{AnySQLFunction{
		FunctionName: "ifnull",
		Min:          2,
		Max:          2,
	}}
	FunctionIIF = ScalarFunction{AnySQLFunction{
		FunctionName: "iif",
		Min:          3,
		Max:          3,
	}}
	FunctionINSTR = ScalarFunction{AnySQLFunction{
		FunctionName: "instr",
		Min:          2,
		Max:          3,
	}}
	FunctionLENGTH = ScalarFunction{AnySQLFunction{
		FunctionName: "length",
		Min:          1,
		Max:          1,
	}}
	FunctionLIKE = ScalarFunction{AnySQLFunction{
		FunctionName: "like",
		Min:          2,
		Max:          3,
	}}
	FunctionLOWER = ScalarFunction{AnySQLFunction{
		FunctionName: "lower",
		Min:          1,
		Max:          1,
	}}
	FunctionLTRIM = ScalarFunction{AnySQLFunction{
		FunctionName: "ltrim",
		Min:          1,
		Max:          2,
	}}
	FunctionMAX = ScalarFunction{AnySQLFunction{
		FunctionName: "max",
		Min:          1,
	}}
	FunctionMIN = ScalarFunction{AnySQLFunction{
		FunctionName: "min",
		Min:          1,
	}}
	FunctionNULLIF = ScalarFunction{AnySQLFunction{
		FunctionName: "nullif",
		Min:          2,
		Max:          2,
	}}
	FunctionQUOTE = ScalarFunction{AnySQLFunction{
		FunctionName: "quote",
		Min:          1,
		Max:          1,
	}}
	FunctionREPLACE = ScalarFunction{AnySQLFunction{
		FunctionName: "replace",
		Min:          3,
		Max:          3,
	}}
	FunctionRTRIM = ScalarFunction{AnySQLFunction{
		FunctionName: "rtrim",
		Min:          1,
		Max:          2,
	}}
	FunctionSIGN = ScalarFunction{AnySQLFunction{
		FunctionName: "sign",
		Min:          1,
		Max:          1,
	}}
	FunctionSUBSTR = ScalarFunction{AnySQLFunction{
		FunctionName: "substr",
		Min:          2,
		Max:          3,
	}}
	FunctionTRIM = ScalarFunction{AnySQLFunction{
		FunctionName: "trim",
		Min:          1,
		Max:          3,
	}}
	FunctionTYPEOF = ScalarFunction{AnySQLFunction{
		FunctionName: "typeof",
		Min:          1,
		Max:          1,
	}}
	FunctionUNHEX = ScalarFunction{AnySQLFunction{
		FunctionName: "unhex",
		Min:          1,
		Max:          1,
	}}
	FunctionUNICODE = ScalarFunction{AnySQLFunction{
		FunctionName: "unicode",
		Min:          1,
		Max:          1,
	}}
	FunctionUPPER = ScalarFunction{AnySQLFunction{
		FunctionName: "upper",
		Min:          1,
		Max:          1,
	}}
)

// SQLFunctions is a map of all functions of all types
var SQLFunctions = map[string]SQLFunction{
	// Scalar functions
	"abs":      &FunctionABS,
	"coalesce": &FunctionCOALESCE,
	"format":   &FunctionFORMAT,
	"glob":     &FunctionGLOB,
	"hex":      &FunctionHEX,
	"ifnull":   &FunctionIFNULL,
	"iif":      &FunctionIIF,
	"instr":    &FunctionINSTR,
	"length":   &FunctionLENGTH,
	"like":     &FunctionLIKE,
	"lower":    &FunctionLOWER,
	"ltrim":    &FunctionLTRIM,
	"max":      &FunctionMAX,
	"min":      &FunctionMIN,
	"nullif":   &FunctionNULLIF,
	"quote":    &FunctionQUOTE,
	"replace":  &FunctionREPLACE,
	"rtrim":    &FunctionRTRIM,
	"sign":     &FunctionSIGN,
	"substr":   &FunctionSUBSTR,
	"trim":     &FunctionTRIM,
	"typeof":   &FunctionTYPEOF,
	"unhex":    &FunctionUNHEX,
	"unicode":  &FunctionUNICODE,
	"upper":    &FunctionUPPER,
	// DateTime functions
	"date":      &FunctionDATE,
	"time":      &FunctionTIME,
	"datetime":  &FunctionDATETIME,
	"unixepoch": &FunctionUNIXEPOCH,
	"strftime":  &FunctionSTRFTIME,
}
