package sqlparser

import (
	"errors"
	"fmt"

	"github.com/antlr4-go/antlr/v4"
)

var ErrInvalidSyntax = errors.New("syntax error")

type ErrorList []error

func (e *ErrorList) Add(msg string) {
	*e = append(*e, errors.New(msg))
}

func (e *ErrorList) AddError(err error) {
	*e = append(*e, err)
}

// Unwrap is used by the standard library's errors.Is function.
func (e ErrorList) Unwrap() []error {
	return e
}

var _ error = ErrorList{}
var _ error = (*ErrorList)(nil)

// Error satisfies the standard library error interface.
func (e ErrorList) Error() string {
	switch len(e) {
	case 0:
		return "no errors"
	case 1:
		return e[0].Error()
	default:
		return fmt.Sprintf("%s (with %d+ errors)", e[0], len(e)-1)
	}
}

func (e ErrorList) Err() error {
	if len(e) == 0 {
		return nil
	}
	return e
}

type ErrorListener struct {
	ErrorList
}

var _ antlr.ErrorListener = &ErrorListener{}

func NewErrorListener() *ErrorListener {
	return &ErrorListener{
		ErrorList: ErrorList{},
	}
}

func (l *ErrorListener) SyntaxError(recognizer antlr.Recognizer, offendingSymbol interface{}, line, column int,
	msg string, e antlr.RecognitionException) {
	//symbol := offendingSymbol.(antlr.Token)
	l.AddError(fmt.Errorf(`%w: line %d:%d "%s"`, ErrInvalidSyntax, line, column, msg))
}

func (l *ErrorListener) ReportAmbiguity(recognizer antlr.Parser, dfa *antlr.DFA, startIndex, stopIndex int,
	exact bool, ambigAlts *antlr.BitSet, configs *antlr.ATNConfigSet) {
	//l.ErrorHandler.Add(startIndex, errors.Wrap(ErrAmbiguity, "ambiguity"))
}

func (l *ErrorListener) ReportAttemptingFullContext(recognizer antlr.Parser, dfa *antlr.DFA, startIndex,
	stopIndex int, conflictingAlts *antlr.BitSet, configs *antlr.ATNConfigSet) {
	//l.ErrorHandler.Add(startIndex, errors.Wrap(ErrAttemptingFullContext, "attempting full context"))
}

func (l *ErrorListener) ReportContextSensitivity(recognizer antlr.Parser, dfa *antlr.DFA, startIndex, stopIndex,
	prediction int, configs *antlr.ATNConfigSet) {
	//l.ErrorHandler.Add(startIndex, errors.Wrap(ErrContextSensitivity, "context sensitivity"))
}
