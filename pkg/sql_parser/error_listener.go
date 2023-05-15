package sql_parser

import (
	"github.com/antlr/antlr4/runtime/Go/antlr/v4"
	"github.com/kwilteam/kwil-db/pkg/kuneiform/scanner"
	"github.com/kwilteam/kwil-db/pkg/kuneiform/token"
	"github.com/pkg/errors"
)

type ErrorHandler struct {
	CurLine int
	Errors  scanner.ErrorList
}

func NewErrorHandler(currentLine int) *ErrorHandler {
	return &ErrorHandler{
		CurLine: currentLine,
	}
}

func (eh *ErrorHandler) Add(column int, err error) {
	eh.Errors.Add(token.Position{
		Line:   token.Pos(eh.CurLine),
		Column: token.Pos(column),
	}, err.Error())
}

type sqliteErrorListener struct {
	*antlr.DefaultErrorListener
	*ErrorHandler

	symbol string
}

func newSqliteErrorListener(eh *ErrorHandler) *sqliteErrorListener {
	return &sqliteErrorListener{
		ErrorHandler: eh,
		symbol:       "",
	}
}

func (s *sqliteErrorListener) SyntaxError(recognizer antlr.Recognizer, offendingSymbol interface{}, line, column int, msg string, e antlr.RecognitionException) {
	symbol := offendingSymbol.(antlr.Token)
	if s.symbol == "" {
		s.symbol = symbol.GetText()
	}
	// calculate relative line number
	relativeLine := line - 1
	defer func() {
		s.ErrorHandler.CurLine -= relativeLine
	}()
	s.ErrorHandler.CurLine += relativeLine
	s.ErrorHandler.Add(column, errors.Wrap(ErrSyntax, msg))
}

func (s *sqliteErrorListener) ReportAmbiguity(recognizer antlr.Parser, dfa *antlr.DFA, startIndex, stopIndex int, exact bool, ambigAlts *antlr.BitSet, configs antlr.ATNConfigSet) {
	//s.ErrorHandler.Add(startIndex, errors.Wrap(ErrAmbiguity, "ambiguity"))
}

func (s *sqliteErrorListener) ReportAttemptingFullContext(recognizer antlr.Parser, dfa *antlr.DFA, startIndex, stopIndex int, conflictingAlts *antlr.BitSet, configs antlr.ATNConfigSet) {
	//s.ErrorHandler.Add(startIndex, errors.Wrap(ErrAttemptingFullContext, "attempting full context"))
}

func (s *sqliteErrorListener) ReportContextSensitivity(recognizer antlr.Parser, dfa *antlr.DFA, startIndex, stopIndex, prediction int, configs antlr.ATNConfigSet) {
	//s.ErrorHandler.Add(startIndex, errors.Wrap(ErrContextSensitivity, "context sensitivity"))
}

