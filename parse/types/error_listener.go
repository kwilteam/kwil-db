package types

import (
	"github.com/antlr4-go/antlr/v4"
)

type BaseErrorListener interface {
	// Err returns the error if there are any, otherwise it returns nil.
	Err() error
	// Errors returns the errors that have been collected.
	Errors() []*ParseError
	// Add adds errors from another error listener to this error listener.
	Add(errs ...*ParseError)
}

// AntlrErrorListener is an interface for error listeners required by functions that directly
// deal with Antlr parsers.
type AntlrErrorListener interface {
	BaseErrorListener
	antlr.ErrorListener
	// TokenErr adds an error where the error comes from an antlr generated token.
	TokenErr(t antlr.Token, errType ParseErrorType, msg string)
	// RuleErr adds an error where the position of an antlr generated rule is
	// known.
	RuleErr(ctx antlr.ParserRuleContext, errType ParseErrorType, msg string)
	// ChildFromToken creates a new error listener that is a child of the current error listener.
	// It will have the same starting position as the token.
	ChildFromToken(name string, t antlr.Token) *ErrorListener
}

// NativeErrorListener iis an interface for error listeners required by Kwil's native
// visitors.
type NativeErrorListener interface {
	BaseErrorListener
	// NodeErr adds an error where our native node type is identifiable.
	NodeErr(node *Node, errType ParseErrorType, msg string)
	// Child creates a new error listener. It will not have any of the errors from the parent,
	// and should simply be used for nested parsing.
	Child(name string, startLine, startCol int) *ErrorListener
}

// ErrorListener is an error listener that collects errors during parsing.
type ErrorListener struct {
	Errs      []*ParseError
	startLine int
	startCol  int
	name      string
}

var _ AntlrErrorListener = &ErrorListener{}
var _ NativeErrorListener = &ErrorListener{}

// ErrorListenerOptions allows for setting options on the ErrorListener.
type ErrorListenerOptions struct {
	// ParentNode is the parent position of the error listener.
	// For example, if the error listener is used in a sub-parser (e.g. the procedure parser),
	// the parent position should be the starting position of the procedure.
	ParentNode *Node
}

// NewErrorListener creates a new error listener with the given options.
func NewErrorListener() *ErrorListener {
	return &ErrorListener{
		Errs: make([]*ParseError, 0),
		name: "kuneiform",
	}
}

// Err returns the error if there are any, otherwise it returns nil.
func (e *ErrorListener) Err() error {
	return CombineParseErrors(e.Errs)
}

// Errors returns the errors that have been collected.
func (e *ErrorListener) Errors() []*ParseError {
	return e.Errs
}

// Add adds errors from another error listener to this error listener.
func (e *ErrorListener) Add(errs ...*ParseError) {
	e.Errs = append(e.Errs, errs...)
}

// NodeErr adds an error that comes from a node.
func (e *ErrorListener) NodeErr(node *Node, errType ParseErrorType, msg string) {
	e.Errs = append(e.Errs, &ParseError{
		ParserName: e.name,
		Type:       errType,
		Err:        msg,
		Node:       e.adjustNode(node),
	})
}

// adjustNode adjusts the node based on the starting position of the error listener.
// It returns a copy of the node with the adjusted position.
func (e *ErrorListener) adjustNode(node *Node) *Node {
	return &Node{
		IsSet:     true,
		StartLine: node.StartLine + e.startLine,
		StartCol:  node.StartCol + e.startCol,
		EndLine:   node.EndLine + e.startLine,
		EndCol:    node.EndCol + e.startCol,
	}
}

// TokenErr adds an error that comes from a token.
func (e *ErrorListener) TokenErr(t antlr.Token, errType ParseErrorType, msg string) {
	e.Errs = append(e.Errs, &ParseError{
		ParserName: e.name,
		Type:       errType,
		Err:        msg,
		Node:       e.adjustNode(unaryNode(t.GetLine()-1, t.GetColumn())),
	})
}

// RuleErr adds an error that comes from a parser rule.
func (e *ErrorListener) RuleErr(ctx antlr.ParserRuleContext, errType ParseErrorType, msg string) {
	node := &Node{}
	node.Set(ctx)
	e.Errs = append(e.Errs, &ParseError{
		ParserName: e.name,
		Type:       errType,
		Err:        msg,
		Node:       e.adjustNode(node),
	})
}

func (e *ErrorListener) SyntaxError(recognizer antlr.Recognizer, offendingSymbol interface{}, line, column int,
	msg string, ex antlr.RecognitionException) {
	e.Errs = append(e.Errs, &ParseError{
		ParserName: e.name,
		Type:       ParseErrorTypeSyntax,
		Err:        ErrSyntaxError.Error() + ": " + msg,
		Node:       e.adjustNode(unaryNode(line, column)),
	})
}

// Child creates a new error listener. It will not have any of the errors from the parent,
// and should simply be used for nested parsing.
func (l *ErrorListener) Child(name string, startLine, startCol int) *ErrorListener {
	return &ErrorListener{
		name:      name,
		Errs:      make([]*ParseError, 0),
		startLine: l.startLine + startLine,
		startCol:  l.startCol + startCol,
	}
}

// ChildFromToken creates a new error listener that is a child of the current error listener.
// It will have the same starting position as the token.
// It is defined here because we have to hack around Antlr's semantic of returning 1-indexed
// lines and 0-indexed columns, and we are encapsulating it in this package.
func (l *ErrorListener) ChildFromToken(name string, t antlr.Token) *ErrorListener {
	startline := t.GetLine() - 1
	startcol := t.GetColumn()
	return l.Child(name, startline, startcol)
}

// We do not need to do anything in the below methods because they are simply Antlr's way of reporting.
// We may want to add warnings in the future, but for now, we will ignore them.
// https://stackoverflow.com/questions/71056312/antlr-how-to-avoid-reportattemptingfullcontext-and-reportambiguity
func (l *ErrorListener) ReportAmbiguity(recognizer antlr.Parser, dfa *antlr.DFA, startIndex, stopIndex int,
	exact bool, ambigAlts *antlr.BitSet, configs *antlr.ATNConfigSet) {
}

func (l *ErrorListener) ReportAttemptingFullContext(recognizer antlr.Parser, dfa *antlr.DFA, startIndex,
	stopIndex int, conflictingAlts *antlr.BitSet, configs *antlr.ATNConfigSet) {
}

func (l *ErrorListener) ReportContextSensitivity(recognizer antlr.Parser, dfa *antlr.DFA, startIndex, stopIndex,
	prediction int, configs *antlr.ATNConfigSet) {
}
