package parse

import (
	"errors"
	"fmt"
	"strings"

	"github.com/antlr4-go/antlr/v4"
)

// ParseErrors is a collection of parse errors.
type ParseErrors []*ParseError

var _ interface{ Unwrap() []error } = (*ParseErrors)(nil)

// Unwrap allows errors.Is and error.As to identify wrapped errors.
func (p ParseErrors) Unwrap() []error {
	errs := make([]error, len(p))
	for i := range p {
		errs[i] = p[i]
	}
	return errs
}

// Err returns all the errors as a single error.
func (p ParseErrors) Err() error {
	if len(p) == 0 {
		return nil
	}
	return &p
}

// The zero value of a ParseErrors instance intentionally does not implement the
// error interface. The Err method will return nil if the length is zero.
var _ error = (*ParseErrors)(nil)

// Error implements the error interface.
func (p *ParseErrors) Error() string {
	errs := *p
	switch len(errs) {
	case 0: // use Err and this won't happen
		return "<nil>"
	case 1:
		return errs[0].Error()
	default:
		var str strings.Builder
		str.WriteString("detected multiple parse errors:")
		for i, err := range errs {
			str.WriteString(fmt.Sprintf("\n%d: %s", i, err.Error()))
		}
		return str.String()
	}
}

// Add adds errors to the collection.
func (p *ParseErrors) Add(errs ...*ParseError) {
	*p = append(*p, errs...)
}

// ParseError is an error that occurred during parsing.
type ParseError struct {
	ParserName string    `json:"parser_name,omitempty"`
	Err        error     `json:"error"`
	Message    string    `json:"message,omitempty"`
	Position   *Position `json:"position,omitempty"`
}

// Unwrap() allows errors.Is and errors.As to find wrapped errors.
func (p ParseError) Unwrap() error {
	return p.Err
}

// Error satisfies the standard library error interface.
func (p *ParseError) Error() string {
	// Add 1 to the column numbers to make them 1-indexed, since antlr-go is 0-indexed
	// for columns.
	return fmt.Sprintf("(%s) %s error: %s\nstart %d:%d end %d:%d", p.ParserName, p.Err.Error(), p.Message,
		p.Position.StartLine, p.Position.StartCol+1,
		p.Position.EndLine, p.Position.EndCol+1)
}

// ParseErrs is a collection of parse errors.
type ParseErrs interface {
	Err() error
	Errors() []*ParseError
	Add(...*ParseError)
}

// errorListener listens to errors emitted by Antlr, and also collects
// errors from Kwil's native validation logic.
type errorListener struct {
	errs []*ParseError
	name string
}

var _ antlr.ErrorListener = (*errorListener)(nil)
var _ ParseErrs = (*errorListener)(nil)

// newErrorListener creates a new error listener with the given options.
func newErrorListener(name string) *errorListener {
	return &errorListener{
		errs: make([]*ParseError, 0),
		name: name,
	}
}

// Err returns the error if there are any, otherwise it returns nil.
func (e *errorListener) Err() error {
	if len(e.errs) == 0 {
		return nil
	}
	pe := ParseErrors(e.errs)
	return &pe
}

// Add adds errors to the collection.
func (e *errorListener) Add(errs ...*ParseError) {
	e.errs = append(e.errs, errs...)
}

// Errors returns the errors that have been collected.
func (e *errorListener) Errors() []*ParseError {
	return e.errs
}

// AddErr adds an error to the error listener.
func (e *errorListener) AddErr(node Node, err error, msg string, v ...any) {
	// TODO: we should change the ParseError struct. It should use the passed error as the "Type",
	// and replace the Err field with message.
	e.errs = append(e.errs, &ParseError{
		ParserName: e.name,
		Err:        err,
		Message:    fmt.Sprintf(msg, v...),
		Position:   node.GetNode(),
	})
}

// TokenErr adds an error that comes from an Antlr token.
func (e *errorListener) TokenErr(t antlr.Token, err error, msg string, v ...any) {
	e.errs = append(e.errs, &ParseError{
		ParserName: e.name,
		Err:        err,
		Message:    fmt.Sprintf(msg, v...),
		Position:   unaryNode(t.GetLine(), t.GetColumn()),
	})
}

// RuleErr adds an error that comes from a Antlr parser rule.
func (e *errorListener) RuleErr(ctx antlr.ParserRuleContext, err error, msg string, v ...any) {
	node := &Position{}
	node.Set(ctx)
	e.errs = append(e.errs, &ParseError{
		ParserName: e.name,
		Err:        err,
		Message:    fmt.Sprintf(msg, v...),
		Position:   node,
	})
}

// SyntaxError implements the Antlr error listener interface.
func (e *errorListener) SyntaxError(recognizer antlr.Recognizer, offendingSymbol interface{}, line, column int,
	msg string, ex antlr.RecognitionException) {
	e.errs = append(e.errs, &ParseError{
		ParserName: e.name,
		Err:        ErrSyntax,
		Message:    msg,
		Position:   unaryNode(line, column),
	})
}

// We do not need to do anything in the below methods because they are simply Antlr's way of reporting.
// We may want to add warnings in the future, but for now, we will ignore them.
// https://stackoverflow.com/questions/71056312/antlr-how-to-avoid-reportattemptingfullcontext-and-reportambiguity

// ReportAmbiguity implements the Antlr error listener interface.
func (e *errorListener) ReportAmbiguity(recognizer antlr.Parser, dfa *antlr.DFA, startIndex, stopIndex int,
	exact bool, ambigAlts *antlr.BitSet, configs *antlr.ATNConfigSet) {
}

// ReportAttemptingFullContext implements the Antlr error listener interface.
func (e *errorListener) ReportAttemptingFullContext(recognizer antlr.Parser, dfa *antlr.DFA, startIndex,
	stopIndex int, conflictingAlts *antlr.BitSet, configs *antlr.ATNConfigSet) {
}

// ReportContextSensitivity implements the Antlr error listener interface.
func (e *errorListener) ReportContextSensitivity(recognizer antlr.Parser, dfa *antlr.DFA, startIndex, stopIndex,
	prediction int, configs *antlr.ATNConfigSet) {
}

var (
	ErrSyntax                     = errors.New("syntax error")
	ErrDuplicateBlock             = errors.New("duplicate block name")
	ErrInvalidIterable            = errors.New("invalid iterable")
	ErrUndeclaredVariable         = errors.New("undeclared variable")
	ErrVariableAlreadyDeclared    = errors.New("variable already declared")
	ErrType                       = errors.New("type error")
	ErrAssignment                 = errors.New("assignment error")
	ErrUnknownTable               = errors.New("unknown table reference")
	ErrTableDefinition            = errors.New("table definition error")
	ErrUnknownColumn              = errors.New("unknown column reference")
	ErrAmbiguousColumn            = errors.New("ambiguous column reference")
	ErrUnknownFunctionOrProcedure = errors.New("unknown function or procedure")
	// ErrFunctionSignature is returned when a function/procedure is called with the wrong number of arguments,
	// or returns an unexpected number of values / table.
	ErrFunctionSignature  = errors.New("function/procedure signature error")
	ErrTableAlreadyExists = errors.New("table already exists")
	// ErrResultShape is used if the result of a query is not in a shape we expect.
	ErrResultShape               = errors.New("result shape error")
	ErrUnnamedResultColumn       = errors.New("unnamed result column")
	ErrTableAlreadyJoined        = errors.New("table already joined")
	ErrUnnamedJoin               = errors.New("unnamed join")
	ErrJoin                      = errors.New("join error")
	ErrBreak                     = errors.New("break error")
	ErrReturn                    = errors.New("return type error")
	ErrAggregate                 = errors.New("aggregate error")
	ErrUnknownContextualVariable = errors.New("unknown contextual variable")
	ErrIdentifier                = errors.New("identifier error")
	ErrActionNotFound            = errors.New("action not found")
	ErrViewMutatesState          = errors.New("view mutates state")
	ErrInvalidActionExpression   = errors.New("invalid action expression")
)
