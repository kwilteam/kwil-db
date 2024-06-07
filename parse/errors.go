package parse

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/antlr4-go/antlr/v4"
)

// WrapErrors wraps a collection of ParseErrors
func WrapErrors(errs ...*ParseError) ParseErrs {
	return &errorListener{errs: errs}
}

// ParseError is an error that occurred during parsing.
type ParseError struct {
	ParserName string    `json:"parser_name,omitempty"`
	Err        error     `json:"error"`
	Message    string    `json:"message,omitempty"`
	Position   *Position `json:"position,omitempty"`
}

// MarshalJSON marshals the error to JSON.
func (p *ParseError) MarshalJSON() ([]byte, error) {
	type Alias struct {
		ParserName string    `json:"parser_name,omitempty"`
		Message    string    `json:"message,omitempty"`
		Position   *Position `json:"position,omitempty"`
	}

	a := &Alias{
		ParserName: p.ParserName,
		Message:    p.Message,
		Position:   p.Position,
	}

	return json.Marshal(struct {
		Error *error `json:"error,omitempty"`
		*Alias
	}{
		Error: &p.Err,
		Alias: a,
	})
}

// Unwrap() allows errors.Is and errors.As to find wrapped errors.
func (p ParseError) Unwrap() error {
	return p.Err
}

// Error satisfies the standard library error interface.
func (p *ParseError) Error() string {
	// Add 1 to the column numbers to make them 1-indexed, since antlr-go is 0-indexed
	// for columns.
	return fmt.Sprintf("(%s) %s: %s\n start %d:%d end %d:%d", p.ParserName, p.Err.Error(), p.Message,
		p.Position.StartLine, p.Position.StartCol+1,
		p.Position.EndLine, p.Position.EndCol+1)
}

// ParseErrs is a collection of parse errors.
type ParseErrs interface {
	Err() error
	Errors() []*ParseError
	Add(...*ParseError)
	MarshalJSON() ([]byte, error)
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
	switch len(e.errs) {
	case 1:
		return e.errs[0]
	default:
		var errChain error
		for i, err := range e.errs {
			if i == 0 {
				errChain = err
				continue
			}
			errChain = fmt.Errorf("%w\n %w", errChain, err)
		}

		return fmt.Errorf("detected multiple parse errors:\n %w", errChain)
	}
}

// Add adds errors to the collection.
func (e *errorListener) Add(errs ...*ParseError) {
	e.errs = append(e.errs, errs...)
}

// Errors returns the errors that have been collected.
func (e *errorListener) Errors() []*ParseError {
	return e.errs
}

// MarshalJSON marshals the errors to JSON.
func (e *errorListener) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.errs)
}

// AddErr adds an error to the error listener.
func (e *errorListener) AddErr(node Node, err error, msg string, v ...any) {
	// TODO: we should change the ParseError struct. It should use the passed error as the "Type",
	// and replace the Err field with message.
	e.errs = append(e.errs, &ParseError{
		ParserName: e.name,
		Err:        err,
		Message:    fmt.Sprintf(msg, v...),
		Position:   node.GetPosition(),
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
	ErrUndeclaredVariable         = errors.New("undeclared variable")
	ErrVariableAlreadyDeclared    = errors.New("variable already declared")
	ErrType                       = errors.New("type error")
	ErrAssignment                 = errors.New("assignment error")
	ErrUnknownTable               = errors.New("unknown table reference")
	ErrTableDefinition            = errors.New("table definition error")
	ErrUnknownColumn              = errors.New("unknown column reference")
	ErrAmbiguousColumn            = errors.New("ambiguous column reference")
	ErrDuplicateResultColumnName  = errors.New("duplicate result column name")
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
	ErrBreak                     = errors.New("break error")
	ErrReturn                    = errors.New("return type error")
	ErrAggregate                 = errors.New("aggregate error")
	ErrUnknownContextualVariable = errors.New("unknown contextual variable")
	ErrIdentifier                = errors.New("identifier error")
	ErrActionNotFound            = errors.New("action not found")
	ErrViewMutatesState          = errors.New("view mutates state")
	ErrOrdering                  = errors.New("ordering error")
	ErrCrossScopeDeclaration     = errors.New("cross-scope declaration")
	ErrInvalidExcludedTable      = errors.New("invalid excluded table usage")
	ErrAmbiguousConflictTable    = errors.New("ambiguous conflict table")
	ErrCollation                 = errors.New("collation error")
)
