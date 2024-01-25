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

func (s *ScalarFunction) Walk(w AstWalker) error {
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
	FunctionABSGetter       = NewScalarFunctionWithGetter("abs", 1, 1, false)
	FunctionCOALESCEGetter  = NewScalarFunctionWithGetter("coalesce", 2, 0, false)
	FunctionERRORGetter     = NewScalarFunctionWithGetter("error", 1, 1, false)
	FunctionFORMATGetter    = NewScalarFunctionWithGetter("format", 1, 0, false)
	FunctionGLOBGetter      = NewScalarFunctionWithGetter("glob", 2, 2, false)
	FunctionHEXGetter       = NewScalarFunctionWithGetter("hex", 1, 1, false)
	FunctionIFNULLGetter    = NewScalarFunctionWithGetter("ifnull", 2, 2, false)
	FunctionIIFGetter       = NewScalarFunctionWithGetter("iif", 3, 3, false)
	FunctionINSTRGetter     = NewScalarFunctionWithGetter("instr", 2, 2, false)
	FunctionLENGTHGetter    = NewScalarFunctionWithGetter("length", 1, 1, false)
	FunctionLIKEGetter      = NewScalarFunctionWithGetter("like", 2, 3, false)
	FunctionLOWERGetter     = NewScalarFunctionWithGetter("lower", 1, 1, false)
	FunctionLTRIMGetter     = NewScalarFunctionWithGetter("ltrim", 1, 2, false)
	FunctionNULLIFGetter    = NewScalarFunctionWithGetter("nullif", 2, 2, false)
	FunctionQUOTEGetter     = NewScalarFunctionWithGetter("quote", 1, 1, false)
	FunctionREPLACEGetter   = NewScalarFunctionWithGetter("replace", 3, 3, false)
	FunctionRTRIMGetter     = NewScalarFunctionWithGetter("rtrim", 1, 2, false)
	FunctionSIGNGetter      = NewScalarFunctionWithGetter("sign", 1, 1, false)
	FunctionSUBSTRGetter    = NewScalarFunctionWithGetter("substr", 2, 3, false)
	FunctionTRIMGetter      = NewScalarFunctionWithGetter("trim", 1, 2, false)
	FunctionTYPEOFGetter    = NewScalarFunctionWithGetter("typeof", 1, 1, false)
	FunctionUNHEXGetter     = NewScalarFunctionWithGetter("unhex", 1, 2, false)
	FunctionUNICODEGetter   = NewScalarFunctionWithGetter("unicode", 1, 1, false)
	FunctionUPPERGetter     = NewScalarFunctionWithGetter("upper", 1, 1, false)
	FunctionAddressGetter   = NewScalarFunctionWithGetter("address", 1, 1, false)
	FunctionPublicKeyGetter = NewScalarFunctionWithGetter("public_key", 1, 2, false)
)

// SQLFunctionGetterMap is a map of function names to SQLFunctionGetters
var SQLFunctionGetterMap = map[string]SQLFunctionGetter{
	// Scalar functions
	"abs":      FunctionABSGetter,
	"coalesce": FunctionCOALESCEGetter,
	"error":    FunctionERRORGetter,
	"format":   FunctionFORMATGetter,
	"glob":     FunctionGLOBGetter,
	"hex":      FunctionHEXGetter,
	"ifnull":   FunctionIFNULLGetter,
	"iif":      FunctionIIFGetter,
	"instr":    FunctionINSTRGetter,
	"length":   FunctionLENGTHGetter,
	"like":     FunctionLIKEGetter,
	"lower":    FunctionLOWERGetter,
	"ltrim":    FunctionLTRIMGetter,
	"nullif":   FunctionNULLIFGetter,
	"quote":    FunctionQUOTEGetter,
	"replace":  FunctionREPLACEGetter,
	"rtrim":    FunctionRTRIMGetter,
	"sign":     FunctionSIGNGetter,
	"substr":   FunctionSUBSTRGetter,
	"trim":     FunctionTRIMGetter,
	"typeof":   FunctionTYPEOFGetter,
	"unhex":    FunctionUNHEXGetter,
	"unicode":  FunctionUNICODEGetter,
	"upper":    FunctionUPPERGetter,

	// @caller functions
	"address":    FunctionAddressGetter,
	"public_key": FunctionPublicKeyGetter,

	// DateTime functions
	"date":      FunctionDATEGetter,
	"time":      FunctionTIMEGetter,
	"datetime":  FunctionDATETIMEGetter,
	"unixepoch": FunctionUNIXEPOCHGetter,
	"strftime":  FunctionSTRFTIMEGetter,

	// Aggregate functions
	"count":        FunctionCOUNTGetter,
	"min":          FunctionMINGetter,
	"max":          FunctionMAXGetter,
	"group_concat": FunctionGROUPCONCATGetter,
}

var (
	FunctionABS = ScalarFunction{
		AnySQLFunction: AnySQLFunction{
			FunctionName: "abs",
			Min:          1,
			Max:          1,
		}}
	FunctionCOALESCE = ScalarFunction{
		AnySQLFunction: AnySQLFunction{
			FunctionName: "coalesce",
			Min:          2,
		}}
	FunctionERROR = ScalarFunction{
		AnySQLFunction: AnySQLFunction{
			FunctionName: "error",
			Min:          1,
			Max:          1,
		}}
	FunctionFORMAT = ScalarFunction{
		AnySQLFunction: AnySQLFunction{
			FunctionName: "format",
			Min:          1,
		}}
	FunctionGLOB = ScalarFunction{
		AnySQLFunction: AnySQLFunction{
			FunctionName: "glob",
			Min:          2,
			Max:          2,
		}}
	FunctionHEX = ScalarFunction{
		AnySQLFunction: AnySQLFunction{
			FunctionName: "hex",
			Min:          1,
			Max:          1,
		}}
	FunctionIFNULL = ScalarFunction{
		AnySQLFunction: AnySQLFunction{
			FunctionName: "ifnull",
			Min:          2,
			Max:          2,
		}}
	FunctionIIF = ScalarFunction{
		AnySQLFunction: AnySQLFunction{
			FunctionName: "iif",
			Min:          3,
			Max:          3,
		}}
	FunctionINSTR = ScalarFunction{
		AnySQLFunction: AnySQLFunction{
			FunctionName: "instr",
			Min:          2,
			Max:          2,
		}}
	FunctionLENGTH = ScalarFunction{
		AnySQLFunction: AnySQLFunction{
			FunctionName: "length",
			Min:          1,
			Max:          1,
		}}
	FunctionLIKE = ScalarFunction{
		AnySQLFunction: AnySQLFunction{
			FunctionName: "like",
			Min:          2,
			Max:          3,
		}}
	FunctionLOWER = ScalarFunction{
		AnySQLFunction: AnySQLFunction{
			FunctionName: "lower",
			Min:          1,
			Max:          1,
		}}
	FunctionLTRIM = ScalarFunction{
		AnySQLFunction: AnySQLFunction{
			FunctionName: "ltrim",
			Min:          1,
			Max:          2,
		}}
	FunctionNULLIF = ScalarFunction{
		AnySQLFunction: AnySQLFunction{
			FunctionName: "nullif",
			Min:          2,
			Max:          2,
		}}
	FunctionQUOTE = ScalarFunction{
		AnySQLFunction: AnySQLFunction{
			FunctionName: "quote",
			Min:          1,
			Max:          1,
		}}
	FunctionREPLACE = ScalarFunction{
		AnySQLFunction: AnySQLFunction{
			FunctionName: "replace",
			Min:          3,
			Max:          3,
		}}
	FunctionRTRIM = ScalarFunction{
		AnySQLFunction: AnySQLFunction{
			FunctionName: "rtrim",
			Min:          1,
			Max:          2,
		}}
	FunctionSIGN = ScalarFunction{
		AnySQLFunction: AnySQLFunction{
			FunctionName: "sign",
			Min:          1,
			Max:          1,
		}}
	FunctionSUBSTR = ScalarFunction{
		AnySQLFunction: AnySQLFunction{
			FunctionName: "substr",
			Min:          2,
			Max:          3,
		}}
	FunctionTRIM = ScalarFunction{
		AnySQLFunction: AnySQLFunction{
			FunctionName: "trim",
			Min:          1,
			Max:          2,
		}}
	FunctionTYPEOF = ScalarFunction{
		AnySQLFunction: AnySQLFunction{
			FunctionName: "typeof",
			Min:          1,
			Max:          1,
		}}
	FunctionUNHEX = ScalarFunction{
		AnySQLFunction: AnySQLFunction{
			FunctionName: "unhex",
			Min:          1,
			Max:          2,
		}}
	FunctionUNICODE = ScalarFunction{
		AnySQLFunction: AnySQLFunction{
			FunctionName: "unicode",
			Min:          1,
			Max:          1,
		}}
	FunctionUPPER = ScalarFunction{
		AnySQLFunction: AnySQLFunction{
			FunctionName: "upper",
			Min:          1,
			Max:          1,
		}}
	FunctionAddress = ScalarFunction{
		AnySQLFunction: AnySQLFunction{
			FunctionName: "address",
			Min:          1,
			Max:          1,
		}}
	FunctionPublicKey = ScalarFunction{
		AnySQLFunction: AnySQLFunction{
			FunctionName: "public_key",
			Min:          1,
			Max:          2,
		}}
)

// SQLFunctions is a map of all functions of all types
var SQLFunctions = map[string]SQLFunction{
	// Scalar functions
	"abs":      &FunctionABS,
	"coalesce": &FunctionCOALESCE,
	"error":    &FunctionERROR,
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

	// @caller functions
	"address":    &FunctionAddress,
	"public_key": &FunctionPublicKey,
	// DateTime functions
	"date":      &FunctionDATE,
	"time":      &FunctionTIME,
	"datetime":  &FunctionDATETIME,
	"unixepoch": &FunctionUNIXEPOCH,
	"strftime":  &FunctionSTRFTIME,
	// Aggregate functions
	"count":        &FunctionCOUNT,
	"min":          &FunctionMIN,
	"max":          &FunctionMAX,
	"group_concat": &FunctionGROUPCONCAT,
}
