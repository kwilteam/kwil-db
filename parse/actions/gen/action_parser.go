// Code generated from ActionParser.g4 by ANTLR 4.13.1. DO NOT EDIT.

package actgrammar // ActionParser
import (
	"fmt"
	"strconv"
	"sync"

	"github.com/antlr4-go/antlr/v4"
)

// Suppress unused import errors
var _ = fmt.Printf
var _ = strconv.Itoa
var _ = sync.Once{}

type ActionParser struct {
	*antlr.BaseParser
}

var ActionParserParserStaticData struct {
	once                   sync.Once
	serializedATN          []int32
	LiteralNames           []string
	SymbolicNames          []string
	RuleNames              []string
	PredictionContextCache *antlr.PredictionContextCache
	atn                    *antlr.ATN
	decisionToDFA          []*antlr.DFA
}

func actionparserParserInit() {
	staticData := &ActionParserParserStaticData
	staticData.LiteralNames = []string{
		"", "';'", "'('", "')'", "','", "'$'", "'@'", "'='", "'.'", "'+'", "'-'",
		"'*'", "'/'", "'%'", "'<'", "'<='", "'>'", "'>='", "'!='", "'<>'", "",
		"", "", "", "", "'not'", "'and'", "'or'",
	}
	staticData.SymbolicNames = []string{
		"", "SCOL", "L_PAREN", "R_PAREN", "COMMA", "DOLLAR", "AT", "ASSIGN",
		"PERIOD", "PLUS", "MINUS", "STAR", "DIV", "MOD", "LT", "LT_EQ", "GT",
		"GT_EQ", "SQL_NOT_EQ1", "SQL_NOT_EQ2", "SELECT_", "INSERT_", "UPDATE_",
		"DELETE_", "WITH_", "NOT_", "AND_", "OR_", "SQL_KEYWORDS", "SQL_STMT",
		"IDENTIFIER", "VARIABLE", "BLOCK_VARIABLE", "UNSIGNED_NUMBER_LITERAL",
		"STRING_LITERAL", "WS", "TERMINATOR", "BLOCK_COMMENT", "LINE_COMMENT",
	}
	staticData.RuleNames = []string{
		"statement", "literal_value", "action_name", "stmt", "sql_stmt", "call_stmt",
		"call_receivers", "call_body", "variable", "block_var", "extension_call_name",
		"fn_name", "sfn_name", "fn_arg_list", "fn_arg_expr",
	}
	staticData.PredictionContextCache = antlr.NewPredictionContextCache()
	staticData.serializedATN = []int32{
		4, 1, 38, 146, 2, 0, 7, 0, 2, 1, 7, 1, 2, 2, 7, 2, 2, 3, 7, 3, 2, 4, 7,
		4, 2, 5, 7, 5, 2, 6, 7, 6, 2, 7, 7, 7, 2, 8, 7, 8, 2, 9, 7, 9, 2, 10, 7,
		10, 2, 11, 7, 11, 2, 12, 7, 12, 2, 13, 7, 13, 2, 14, 7, 14, 1, 0, 4, 0,
		32, 8, 0, 11, 0, 12, 0, 33, 1, 0, 1, 0, 1, 1, 1, 1, 1, 2, 1, 2, 1, 3, 1,
		3, 3, 3, 44, 8, 3, 1, 4, 1, 4, 1, 4, 1, 5, 1, 5, 1, 5, 3, 5, 52, 8, 5,
		1, 5, 1, 5, 1, 5, 1, 6, 1, 6, 1, 6, 5, 6, 60, 8, 6, 10, 6, 12, 6, 63, 9,
		6, 1, 7, 1, 7, 1, 7, 1, 7, 1, 7, 1, 8, 1, 8, 1, 9, 1, 9, 1, 10, 1, 10,
		1, 10, 1, 10, 1, 11, 1, 11, 3, 11, 80, 8, 11, 1, 12, 1, 12, 1, 13, 3, 13,
		85, 8, 13, 1, 13, 1, 13, 5, 13, 89, 8, 13, 10, 13, 12, 13, 92, 9, 13, 1,
		14, 1, 14, 1, 14, 1, 14, 1, 14, 1, 14, 1, 14, 1, 14, 1, 14, 5, 14, 103,
		8, 14, 10, 14, 12, 14, 106, 9, 14, 1, 14, 3, 14, 109, 8, 14, 1, 14, 1,
		14, 1, 14, 1, 14, 1, 14, 1, 14, 1, 14, 1, 14, 1, 14, 1, 14, 3, 14, 121,
		8, 14, 1, 14, 1, 14, 1, 14, 1, 14, 1, 14, 1, 14, 1, 14, 1, 14, 1, 14, 1,
		14, 1, 14, 1, 14, 1, 14, 1, 14, 1, 14, 1, 14, 1, 14, 1, 14, 5, 14, 141,
		8, 14, 10, 14, 12, 14, 144, 9, 14, 1, 14, 0, 1, 28, 15, 0, 2, 4, 6, 8,
		10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 0, 5, 1, 0, 33, 34, 1, 0, 9, 10,
		1, 0, 11, 13, 1, 0, 14, 17, 2, 0, 7, 7, 18, 19, 152, 0, 31, 1, 0, 0, 0,
		2, 37, 1, 0, 0, 0, 4, 39, 1, 0, 0, 0, 6, 43, 1, 0, 0, 0, 8, 45, 1, 0, 0,
		0, 10, 51, 1, 0, 0, 0, 12, 56, 1, 0, 0, 0, 14, 64, 1, 0, 0, 0, 16, 69,
		1, 0, 0, 0, 18, 71, 1, 0, 0, 0, 20, 73, 1, 0, 0, 0, 22, 79, 1, 0, 0, 0,
		24, 81, 1, 0, 0, 0, 26, 84, 1, 0, 0, 0, 28, 120, 1, 0, 0, 0, 30, 32, 3,
		6, 3, 0, 31, 30, 1, 0, 0, 0, 32, 33, 1, 0, 0, 0, 33, 31, 1, 0, 0, 0, 33,
		34, 1, 0, 0, 0, 34, 35, 1, 0, 0, 0, 35, 36, 5, 0, 0, 1, 36, 1, 1, 0, 0,
		0, 37, 38, 7, 0, 0, 0, 38, 3, 1, 0, 0, 0, 39, 40, 5, 30, 0, 0, 40, 5, 1,
		0, 0, 0, 41, 44, 3, 8, 4, 0, 42, 44, 3, 10, 5, 0, 43, 41, 1, 0, 0, 0, 43,
		42, 1, 0, 0, 0, 44, 7, 1, 0, 0, 0, 45, 46, 5, 29, 0, 0, 46, 47, 5, 1, 0,
		0, 47, 9, 1, 0, 0, 0, 48, 49, 3, 12, 6, 0, 49, 50, 5, 7, 0, 0, 50, 52,
		1, 0, 0, 0, 51, 48, 1, 0, 0, 0, 51, 52, 1, 0, 0, 0, 52, 53, 1, 0, 0, 0,
		53, 54, 3, 14, 7, 0, 54, 55, 5, 1, 0, 0, 55, 11, 1, 0, 0, 0, 56, 61, 3,
		16, 8, 0, 57, 58, 5, 4, 0, 0, 58, 60, 3, 16, 8, 0, 59, 57, 1, 0, 0, 0,
		60, 63, 1, 0, 0, 0, 61, 59, 1, 0, 0, 0, 61, 62, 1, 0, 0, 0, 62, 13, 1,
		0, 0, 0, 63, 61, 1, 0, 0, 0, 64, 65, 3, 22, 11, 0, 65, 66, 5, 2, 0, 0,
		66, 67, 3, 26, 13, 0, 67, 68, 5, 3, 0, 0, 68, 15, 1, 0, 0, 0, 69, 70, 5,
		31, 0, 0, 70, 17, 1, 0, 0, 0, 71, 72, 5, 32, 0, 0, 72, 19, 1, 0, 0, 0,
		73, 74, 5, 30, 0, 0, 74, 75, 5, 8, 0, 0, 75, 76, 5, 30, 0, 0, 76, 21, 1,
		0, 0, 0, 77, 80, 3, 20, 10, 0, 78, 80, 3, 4, 2, 0, 79, 77, 1, 0, 0, 0,
		79, 78, 1, 0, 0, 0, 80, 23, 1, 0, 0, 0, 81, 82, 5, 30, 0, 0, 82, 25, 1,
		0, 0, 0, 83, 85, 3, 28, 14, 0, 84, 83, 1, 0, 0, 0, 84, 85, 1, 0, 0, 0,
		85, 90, 1, 0, 0, 0, 86, 87, 5, 4, 0, 0, 87, 89, 3, 28, 14, 0, 88, 86, 1,
		0, 0, 0, 89, 92, 1, 0, 0, 0, 90, 88, 1, 0, 0, 0, 90, 91, 1, 0, 0, 0, 91,
		27, 1, 0, 0, 0, 92, 90, 1, 0, 0, 0, 93, 94, 6, 14, -1, 0, 94, 121, 3, 2,
		1, 0, 95, 121, 3, 16, 8, 0, 96, 121, 3, 18, 9, 0, 97, 98, 3, 24, 12, 0,
		98, 108, 5, 2, 0, 0, 99, 104, 3, 28, 14, 0, 100, 101, 5, 4, 0, 0, 101,
		103, 3, 28, 14, 0, 102, 100, 1, 0, 0, 0, 103, 106, 1, 0, 0, 0, 104, 102,
		1, 0, 0, 0, 104, 105, 1, 0, 0, 0, 105, 109, 1, 0, 0, 0, 106, 104, 1, 0,
		0, 0, 107, 109, 5, 11, 0, 0, 108, 99, 1, 0, 0, 0, 108, 107, 1, 0, 0, 0,
		108, 109, 1, 0, 0, 0, 109, 110, 1, 0, 0, 0, 110, 111, 5, 3, 0, 0, 111,
		121, 1, 0, 0, 0, 112, 113, 5, 2, 0, 0, 113, 114, 3, 28, 14, 0, 114, 115,
		5, 3, 0, 0, 115, 121, 1, 0, 0, 0, 116, 117, 7, 1, 0, 0, 117, 121, 3, 28,
		14, 8, 118, 119, 5, 25, 0, 0, 119, 121, 3, 28, 14, 3, 120, 93, 1, 0, 0,
		0, 120, 95, 1, 0, 0, 0, 120, 96, 1, 0, 0, 0, 120, 97, 1, 0, 0, 0, 120,
		112, 1, 0, 0, 0, 120, 116, 1, 0, 0, 0, 120, 118, 1, 0, 0, 0, 121, 142,
		1, 0, 0, 0, 122, 123, 10, 7, 0, 0, 123, 124, 7, 2, 0, 0, 124, 141, 3, 28,
		14, 8, 125, 126, 10, 6, 0, 0, 126, 127, 7, 1, 0, 0, 127, 141, 3, 28, 14,
		7, 128, 129, 10, 5, 0, 0, 129, 130, 7, 3, 0, 0, 130, 141, 3, 28, 14, 6,
		131, 132, 10, 4, 0, 0, 132, 133, 7, 4, 0, 0, 133, 141, 3, 28, 14, 5, 134,
		135, 10, 2, 0, 0, 135, 136, 5, 26, 0, 0, 136, 141, 3, 28, 14, 3, 137, 138,
		10, 1, 0, 0, 138, 139, 5, 27, 0, 0, 139, 141, 3, 28, 14, 2, 140, 122, 1,
		0, 0, 0, 140, 125, 1, 0, 0, 0, 140, 128, 1, 0, 0, 0, 140, 131, 1, 0, 0,
		0, 140, 134, 1, 0, 0, 0, 140, 137, 1, 0, 0, 0, 141, 144, 1, 0, 0, 0, 142,
		140, 1, 0, 0, 0, 142, 143, 1, 0, 0, 0, 143, 29, 1, 0, 0, 0, 144, 142, 1,
		0, 0, 0, 12, 33, 43, 51, 61, 79, 84, 90, 104, 108, 120, 140, 142,
	}
	deserializer := antlr.NewATNDeserializer(nil)
	staticData.atn = deserializer.Deserialize(staticData.serializedATN)
	atn := staticData.atn
	staticData.decisionToDFA = make([]*antlr.DFA, len(atn.DecisionToState))
	decisionToDFA := staticData.decisionToDFA
	for index, state := range atn.DecisionToState {
		decisionToDFA[index] = antlr.NewDFA(state, index)
	}
}

// ActionParserInit initializes any static state used to implement ActionParser. By default the
// static state used to implement the parser is lazily initialized during the first call to
// NewActionParser(). You can call this function if you wish to initialize the static state ahead
// of time.
func ActionParserInit() {
	staticData := &ActionParserParserStaticData
	staticData.once.Do(actionparserParserInit)
}

// NewActionParser produces a new parser instance for the optional input antlr.TokenStream.
func NewActionParser(input antlr.TokenStream) *ActionParser {
	ActionParserInit()
	this := new(ActionParser)
	this.BaseParser = antlr.NewBaseParser(input)
	staticData := &ActionParserParserStaticData
	this.Interpreter = antlr.NewParserATNSimulator(this, staticData.atn, staticData.decisionToDFA, staticData.PredictionContextCache)
	this.RuleNames = staticData.RuleNames
	this.LiteralNames = staticData.LiteralNames
	this.SymbolicNames = staticData.SymbolicNames
	this.GrammarFileName = "ActionParser.g4"

	return this
}

// ActionParser tokens.
const (
	ActionParserEOF                     = antlr.TokenEOF
	ActionParserSCOL                    = 1
	ActionParserL_PAREN                 = 2
	ActionParserR_PAREN                 = 3
	ActionParserCOMMA                   = 4
	ActionParserDOLLAR                  = 5
	ActionParserAT                      = 6
	ActionParserASSIGN                  = 7
	ActionParserPERIOD                  = 8
	ActionParserPLUS                    = 9
	ActionParserMINUS                   = 10
	ActionParserSTAR                    = 11
	ActionParserDIV                     = 12
	ActionParserMOD                     = 13
	ActionParserLT                      = 14
	ActionParserLT_EQ                   = 15
	ActionParserGT                      = 16
	ActionParserGT_EQ                   = 17
	ActionParserSQL_NOT_EQ1             = 18
	ActionParserSQL_NOT_EQ2             = 19
	ActionParserSELECT_                 = 20
	ActionParserINSERT_                 = 21
	ActionParserUPDATE_                 = 22
	ActionParserDELETE_                 = 23
	ActionParserWITH_                   = 24
	ActionParserNOT_                    = 25
	ActionParserAND_                    = 26
	ActionParserOR_                     = 27
	ActionParserSQL_KEYWORDS            = 28
	ActionParserSQL_STMT                = 29
	ActionParserIDENTIFIER              = 30
	ActionParserVARIABLE                = 31
	ActionParserBLOCK_VARIABLE          = 32
	ActionParserUNSIGNED_NUMBER_LITERAL = 33
	ActionParserSTRING_LITERAL          = 34
	ActionParserWS                      = 35
	ActionParserTERMINATOR              = 36
	ActionParserBLOCK_COMMENT           = 37
	ActionParserLINE_COMMENT            = 38
)

// ActionParser rules.
const (
	ActionParserRULE_statement           = 0
	ActionParserRULE_literal_value       = 1
	ActionParserRULE_action_name         = 2
	ActionParserRULE_stmt                = 3
	ActionParserRULE_sql_stmt            = 4
	ActionParserRULE_call_stmt           = 5
	ActionParserRULE_call_receivers      = 6
	ActionParserRULE_call_body           = 7
	ActionParserRULE_variable            = 8
	ActionParserRULE_block_var           = 9
	ActionParserRULE_extension_call_name = 10
	ActionParserRULE_fn_name             = 11
	ActionParserRULE_sfn_name            = 12
	ActionParserRULE_fn_arg_list         = 13
	ActionParserRULE_fn_arg_expr         = 14
)

// IStatementContext is an interface to support dynamic dispatch.
type IStatementContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	EOF() antlr.TerminalNode
	AllStmt() []IStmtContext
	Stmt(i int) IStmtContext

	// IsStatementContext differentiates from other interfaces.
	IsStatementContext()
}

type StatementContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyStatementContext() *StatementContext {
	var p = new(StatementContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ActionParserRULE_statement
	return p
}

func InitEmptyStatementContext(p *StatementContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ActionParserRULE_statement
}

func (*StatementContext) IsStatementContext() {}

func NewStatementContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *StatementContext {
	var p = new(StatementContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = ActionParserRULE_statement

	return p
}

func (s *StatementContext) GetParser() antlr.Parser { return s.parser }

func (s *StatementContext) EOF() antlr.TerminalNode {
	return s.GetToken(ActionParserEOF, 0)
}

func (s *StatementContext) AllStmt() []IStmtContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IStmtContext); ok {
			len++
		}
	}

	tst := make([]IStmtContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IStmtContext); ok {
			tst[i] = t.(IStmtContext)
			i++
		}
	}

	return tst
}

func (s *StatementContext) Stmt(i int) IStmtContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IStmtContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IStmtContext)
}

func (s *StatementContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *StatementContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *StatementContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case ActionParserVisitor:
		return t.VisitStatement(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *ActionParser) Statement() (localctx IStatementContext) {
	localctx = NewStatementContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 0, ActionParserRULE_statement)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(31)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for ok := true; ok; ok = ((int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&3758096384) != 0) {
		{
			p.SetState(30)
			p.Stmt()
		}

		p.SetState(33)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(35)
		p.Match(ActionParserEOF)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ILiteral_valueContext is an interface to support dynamic dispatch.
type ILiteral_valueContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	STRING_LITERAL() antlr.TerminalNode
	UNSIGNED_NUMBER_LITERAL() antlr.TerminalNode

	// IsLiteral_valueContext differentiates from other interfaces.
	IsLiteral_valueContext()
}

type Literal_valueContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyLiteral_valueContext() *Literal_valueContext {
	var p = new(Literal_valueContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ActionParserRULE_literal_value
	return p
}

func InitEmptyLiteral_valueContext(p *Literal_valueContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ActionParserRULE_literal_value
}

func (*Literal_valueContext) IsLiteral_valueContext() {}

func NewLiteral_valueContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Literal_valueContext {
	var p = new(Literal_valueContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = ActionParserRULE_literal_value

	return p
}

func (s *Literal_valueContext) GetParser() antlr.Parser { return s.parser }

func (s *Literal_valueContext) STRING_LITERAL() antlr.TerminalNode {
	return s.GetToken(ActionParserSTRING_LITERAL, 0)
}

func (s *Literal_valueContext) UNSIGNED_NUMBER_LITERAL() antlr.TerminalNode {
	return s.GetToken(ActionParserUNSIGNED_NUMBER_LITERAL, 0)
}

func (s *Literal_valueContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Literal_valueContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Literal_valueContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case ActionParserVisitor:
		return t.VisitLiteral_value(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *ActionParser) Literal_value() (localctx ILiteral_valueContext) {
	localctx = NewLiteral_valueContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 2, ActionParserRULE_literal_value)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(37)
		_la = p.GetTokenStream().LA(1)

		if !(_la == ActionParserUNSIGNED_NUMBER_LITERAL || _la == ActionParserSTRING_LITERAL) {
			p.GetErrorHandler().RecoverInline(p)
		} else {
			p.GetErrorHandler().ReportMatch(p)
			p.Consume()
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IAction_nameContext is an interface to support dynamic dispatch.
type IAction_nameContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	IDENTIFIER() antlr.TerminalNode

	// IsAction_nameContext differentiates from other interfaces.
	IsAction_nameContext()
}

type Action_nameContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyAction_nameContext() *Action_nameContext {
	var p = new(Action_nameContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ActionParserRULE_action_name
	return p
}

func InitEmptyAction_nameContext(p *Action_nameContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ActionParserRULE_action_name
}

func (*Action_nameContext) IsAction_nameContext() {}

func NewAction_nameContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Action_nameContext {
	var p = new(Action_nameContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = ActionParserRULE_action_name

	return p
}

func (s *Action_nameContext) GetParser() antlr.Parser { return s.parser }

func (s *Action_nameContext) IDENTIFIER() antlr.TerminalNode {
	return s.GetToken(ActionParserIDENTIFIER, 0)
}

func (s *Action_nameContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Action_nameContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Action_nameContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case ActionParserVisitor:
		return t.VisitAction_name(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *ActionParser) Action_name() (localctx IAction_nameContext) {
	localctx = NewAction_nameContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 4, ActionParserRULE_action_name)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(39)
		p.Match(ActionParserIDENTIFIER)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IStmtContext is an interface to support dynamic dispatch.
type IStmtContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Sql_stmt() ISql_stmtContext
	Call_stmt() ICall_stmtContext

	// IsStmtContext differentiates from other interfaces.
	IsStmtContext()
}

type StmtContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyStmtContext() *StmtContext {
	var p = new(StmtContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ActionParserRULE_stmt
	return p
}

func InitEmptyStmtContext(p *StmtContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ActionParserRULE_stmt
}

func (*StmtContext) IsStmtContext() {}

func NewStmtContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *StmtContext {
	var p = new(StmtContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = ActionParserRULE_stmt

	return p
}

func (s *StmtContext) GetParser() antlr.Parser { return s.parser }

func (s *StmtContext) Sql_stmt() ISql_stmtContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ISql_stmtContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ISql_stmtContext)
}

func (s *StmtContext) Call_stmt() ICall_stmtContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ICall_stmtContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ICall_stmtContext)
}

func (s *StmtContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *StmtContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *StmtContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case ActionParserVisitor:
		return t.VisitStmt(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *ActionParser) Stmt() (localctx IStmtContext) {
	localctx = NewStmtContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 6, ActionParserRULE_stmt)
	p.SetState(43)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case ActionParserSQL_STMT:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(41)
			p.Sql_stmt()
		}

	case ActionParserIDENTIFIER, ActionParserVARIABLE:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(42)
			p.Call_stmt()
		}

	default:
		p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ISql_stmtContext is an interface to support dynamic dispatch.
type ISql_stmtContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	SQL_STMT() antlr.TerminalNode
	SCOL() antlr.TerminalNode

	// IsSql_stmtContext differentiates from other interfaces.
	IsSql_stmtContext()
}

type Sql_stmtContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptySql_stmtContext() *Sql_stmtContext {
	var p = new(Sql_stmtContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ActionParserRULE_sql_stmt
	return p
}

func InitEmptySql_stmtContext(p *Sql_stmtContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ActionParserRULE_sql_stmt
}

func (*Sql_stmtContext) IsSql_stmtContext() {}

func NewSql_stmtContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Sql_stmtContext {
	var p = new(Sql_stmtContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = ActionParserRULE_sql_stmt

	return p
}

func (s *Sql_stmtContext) GetParser() antlr.Parser { return s.parser }

func (s *Sql_stmtContext) SQL_STMT() antlr.TerminalNode {
	return s.GetToken(ActionParserSQL_STMT, 0)
}

func (s *Sql_stmtContext) SCOL() antlr.TerminalNode {
	return s.GetToken(ActionParserSCOL, 0)
}

func (s *Sql_stmtContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Sql_stmtContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Sql_stmtContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case ActionParserVisitor:
		return t.VisitSql_stmt(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *ActionParser) Sql_stmt() (localctx ISql_stmtContext) {
	localctx = NewSql_stmtContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 8, ActionParserRULE_sql_stmt)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(45)
		p.Match(ActionParserSQL_STMT)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(46)
		p.Match(ActionParserSCOL)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ICall_stmtContext is an interface to support dynamic dispatch.
type ICall_stmtContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Call_body() ICall_bodyContext
	SCOL() antlr.TerminalNode
	Call_receivers() ICall_receiversContext
	ASSIGN() antlr.TerminalNode

	// IsCall_stmtContext differentiates from other interfaces.
	IsCall_stmtContext()
}

type Call_stmtContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyCall_stmtContext() *Call_stmtContext {
	var p = new(Call_stmtContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ActionParserRULE_call_stmt
	return p
}

func InitEmptyCall_stmtContext(p *Call_stmtContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ActionParserRULE_call_stmt
}

func (*Call_stmtContext) IsCall_stmtContext() {}

func NewCall_stmtContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Call_stmtContext {
	var p = new(Call_stmtContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = ActionParserRULE_call_stmt

	return p
}

func (s *Call_stmtContext) GetParser() antlr.Parser { return s.parser }

func (s *Call_stmtContext) Call_body() ICall_bodyContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ICall_bodyContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ICall_bodyContext)
}

func (s *Call_stmtContext) SCOL() antlr.TerminalNode {
	return s.GetToken(ActionParserSCOL, 0)
}

func (s *Call_stmtContext) Call_receivers() ICall_receiversContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ICall_receiversContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ICall_receiversContext)
}

func (s *Call_stmtContext) ASSIGN() antlr.TerminalNode {
	return s.GetToken(ActionParserASSIGN, 0)
}

func (s *Call_stmtContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Call_stmtContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Call_stmtContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case ActionParserVisitor:
		return t.VisitCall_stmt(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *ActionParser) Call_stmt() (localctx ICall_stmtContext) {
	localctx = NewCall_stmtContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 10, ActionParserRULE_call_stmt)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(51)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == ActionParserVARIABLE {
		{
			p.SetState(48)
			p.Call_receivers()
		}
		{
			p.SetState(49)
			p.Match(ActionParserASSIGN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	}
	{
		p.SetState(53)
		p.Call_body()
	}
	{
		p.SetState(54)
		p.Match(ActionParserSCOL)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ICall_receiversContext is an interface to support dynamic dispatch.
type ICall_receiversContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllVariable() []IVariableContext
	Variable(i int) IVariableContext
	AllCOMMA() []antlr.TerminalNode
	COMMA(i int) antlr.TerminalNode

	// IsCall_receiversContext differentiates from other interfaces.
	IsCall_receiversContext()
}

type Call_receiversContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyCall_receiversContext() *Call_receiversContext {
	var p = new(Call_receiversContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ActionParserRULE_call_receivers
	return p
}

func InitEmptyCall_receiversContext(p *Call_receiversContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ActionParserRULE_call_receivers
}

func (*Call_receiversContext) IsCall_receiversContext() {}

func NewCall_receiversContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Call_receiversContext {
	var p = new(Call_receiversContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = ActionParserRULE_call_receivers

	return p
}

func (s *Call_receiversContext) GetParser() antlr.Parser { return s.parser }

func (s *Call_receiversContext) AllVariable() []IVariableContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IVariableContext); ok {
			len++
		}
	}

	tst := make([]IVariableContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IVariableContext); ok {
			tst[i] = t.(IVariableContext)
			i++
		}
	}

	return tst
}

func (s *Call_receiversContext) Variable(i int) IVariableContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IVariableContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IVariableContext)
}

func (s *Call_receiversContext) AllCOMMA() []antlr.TerminalNode {
	return s.GetTokens(ActionParserCOMMA)
}

func (s *Call_receiversContext) COMMA(i int) antlr.TerminalNode {
	return s.GetToken(ActionParserCOMMA, i)
}

func (s *Call_receiversContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Call_receiversContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Call_receiversContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case ActionParserVisitor:
		return t.VisitCall_receivers(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *ActionParser) Call_receivers() (localctx ICall_receiversContext) {
	localctx = NewCall_receiversContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 12, ActionParserRULE_call_receivers)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(56)
		p.Variable()
	}
	p.SetState(61)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == ActionParserCOMMA {
		{
			p.SetState(57)
			p.Match(ActionParserCOMMA)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(58)
			p.Variable()
		}

		p.SetState(63)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ICall_bodyContext is an interface to support dynamic dispatch.
type ICall_bodyContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Fn_name() IFn_nameContext
	L_PAREN() antlr.TerminalNode
	Fn_arg_list() IFn_arg_listContext
	R_PAREN() antlr.TerminalNode

	// IsCall_bodyContext differentiates from other interfaces.
	IsCall_bodyContext()
}

type Call_bodyContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyCall_bodyContext() *Call_bodyContext {
	var p = new(Call_bodyContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ActionParserRULE_call_body
	return p
}

func InitEmptyCall_bodyContext(p *Call_bodyContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ActionParserRULE_call_body
}

func (*Call_bodyContext) IsCall_bodyContext() {}

func NewCall_bodyContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Call_bodyContext {
	var p = new(Call_bodyContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = ActionParserRULE_call_body

	return p
}

func (s *Call_bodyContext) GetParser() antlr.Parser { return s.parser }

func (s *Call_bodyContext) Fn_name() IFn_nameContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IFn_nameContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IFn_nameContext)
}

func (s *Call_bodyContext) L_PAREN() antlr.TerminalNode {
	return s.GetToken(ActionParserL_PAREN, 0)
}

func (s *Call_bodyContext) Fn_arg_list() IFn_arg_listContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IFn_arg_listContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IFn_arg_listContext)
}

func (s *Call_bodyContext) R_PAREN() antlr.TerminalNode {
	return s.GetToken(ActionParserR_PAREN, 0)
}

func (s *Call_bodyContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Call_bodyContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Call_bodyContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case ActionParserVisitor:
		return t.VisitCall_body(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *ActionParser) Call_body() (localctx ICall_bodyContext) {
	localctx = NewCall_bodyContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 14, ActionParserRULE_call_body)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(64)
		p.Fn_name()
	}
	{
		p.SetState(65)
		p.Match(ActionParserL_PAREN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(66)
		p.Fn_arg_list()
	}
	{
		p.SetState(67)
		p.Match(ActionParserR_PAREN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IVariableContext is an interface to support dynamic dispatch.
type IVariableContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	VARIABLE() antlr.TerminalNode

	// IsVariableContext differentiates from other interfaces.
	IsVariableContext()
}

type VariableContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyVariableContext() *VariableContext {
	var p = new(VariableContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ActionParserRULE_variable
	return p
}

func InitEmptyVariableContext(p *VariableContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ActionParserRULE_variable
}

func (*VariableContext) IsVariableContext() {}

func NewVariableContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *VariableContext {
	var p = new(VariableContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = ActionParserRULE_variable

	return p
}

func (s *VariableContext) GetParser() antlr.Parser { return s.parser }

func (s *VariableContext) VARIABLE() antlr.TerminalNode {
	return s.GetToken(ActionParserVARIABLE, 0)
}

func (s *VariableContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *VariableContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *VariableContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case ActionParserVisitor:
		return t.VisitVariable(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *ActionParser) Variable() (localctx IVariableContext) {
	localctx = NewVariableContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 16, ActionParserRULE_variable)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(69)
		p.Match(ActionParserVARIABLE)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IBlock_varContext is an interface to support dynamic dispatch.
type IBlock_varContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	BLOCK_VARIABLE() antlr.TerminalNode

	// IsBlock_varContext differentiates from other interfaces.
	IsBlock_varContext()
}

type Block_varContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyBlock_varContext() *Block_varContext {
	var p = new(Block_varContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ActionParserRULE_block_var
	return p
}

func InitEmptyBlock_varContext(p *Block_varContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ActionParserRULE_block_var
}

func (*Block_varContext) IsBlock_varContext() {}

func NewBlock_varContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Block_varContext {
	var p = new(Block_varContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = ActionParserRULE_block_var

	return p
}

func (s *Block_varContext) GetParser() antlr.Parser { return s.parser }

func (s *Block_varContext) BLOCK_VARIABLE() antlr.TerminalNode {
	return s.GetToken(ActionParserBLOCK_VARIABLE, 0)
}

func (s *Block_varContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Block_varContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Block_varContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case ActionParserVisitor:
		return t.VisitBlock_var(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *ActionParser) Block_var() (localctx IBlock_varContext) {
	localctx = NewBlock_varContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 18, ActionParserRULE_block_var)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(71)
		p.Match(ActionParserBLOCK_VARIABLE)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IExtension_call_nameContext is an interface to support dynamic dispatch.
type IExtension_call_nameContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllIDENTIFIER() []antlr.TerminalNode
	IDENTIFIER(i int) antlr.TerminalNode
	PERIOD() antlr.TerminalNode

	// IsExtension_call_nameContext differentiates from other interfaces.
	IsExtension_call_nameContext()
}

type Extension_call_nameContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyExtension_call_nameContext() *Extension_call_nameContext {
	var p = new(Extension_call_nameContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ActionParserRULE_extension_call_name
	return p
}

func InitEmptyExtension_call_nameContext(p *Extension_call_nameContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ActionParserRULE_extension_call_name
}

func (*Extension_call_nameContext) IsExtension_call_nameContext() {}

func NewExtension_call_nameContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Extension_call_nameContext {
	var p = new(Extension_call_nameContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = ActionParserRULE_extension_call_name

	return p
}

func (s *Extension_call_nameContext) GetParser() antlr.Parser { return s.parser }

func (s *Extension_call_nameContext) AllIDENTIFIER() []antlr.TerminalNode {
	return s.GetTokens(ActionParserIDENTIFIER)
}

func (s *Extension_call_nameContext) IDENTIFIER(i int) antlr.TerminalNode {
	return s.GetToken(ActionParserIDENTIFIER, i)
}

func (s *Extension_call_nameContext) PERIOD() antlr.TerminalNode {
	return s.GetToken(ActionParserPERIOD, 0)
}

func (s *Extension_call_nameContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Extension_call_nameContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Extension_call_nameContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case ActionParserVisitor:
		return t.VisitExtension_call_name(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *ActionParser) Extension_call_name() (localctx IExtension_call_nameContext) {
	localctx = NewExtension_call_nameContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 20, ActionParserRULE_extension_call_name)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(73)
		p.Match(ActionParserIDENTIFIER)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(74)
		p.Match(ActionParserPERIOD)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(75)
		p.Match(ActionParserIDENTIFIER)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IFn_nameContext is an interface to support dynamic dispatch.
type IFn_nameContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Extension_call_name() IExtension_call_nameContext
	Action_name() IAction_nameContext

	// IsFn_nameContext differentiates from other interfaces.
	IsFn_nameContext()
}

type Fn_nameContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyFn_nameContext() *Fn_nameContext {
	var p = new(Fn_nameContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ActionParserRULE_fn_name
	return p
}

func InitEmptyFn_nameContext(p *Fn_nameContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ActionParserRULE_fn_name
}

func (*Fn_nameContext) IsFn_nameContext() {}

func NewFn_nameContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Fn_nameContext {
	var p = new(Fn_nameContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = ActionParserRULE_fn_name

	return p
}

func (s *Fn_nameContext) GetParser() antlr.Parser { return s.parser }

func (s *Fn_nameContext) Extension_call_name() IExtension_call_nameContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExtension_call_nameContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExtension_call_nameContext)
}

func (s *Fn_nameContext) Action_name() IAction_nameContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IAction_nameContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IAction_nameContext)
}

func (s *Fn_nameContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Fn_nameContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Fn_nameContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case ActionParserVisitor:
		return t.VisitFn_name(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *ActionParser) Fn_name() (localctx IFn_nameContext) {
	localctx = NewFn_nameContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 22, ActionParserRULE_fn_name)
	p.SetState(79)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 4, p.GetParserRuleContext()) {
	case 1:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(77)
			p.Extension_call_name()
		}

	case 2:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(78)
			p.Action_name()
		}

	case antlr.ATNInvalidAltNumber:
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ISfn_nameContext is an interface to support dynamic dispatch.
type ISfn_nameContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	IDENTIFIER() antlr.TerminalNode

	// IsSfn_nameContext differentiates from other interfaces.
	IsSfn_nameContext()
}

type Sfn_nameContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptySfn_nameContext() *Sfn_nameContext {
	var p = new(Sfn_nameContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ActionParserRULE_sfn_name
	return p
}

func InitEmptySfn_nameContext(p *Sfn_nameContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ActionParserRULE_sfn_name
}

func (*Sfn_nameContext) IsSfn_nameContext() {}

func NewSfn_nameContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Sfn_nameContext {
	var p = new(Sfn_nameContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = ActionParserRULE_sfn_name

	return p
}

func (s *Sfn_nameContext) GetParser() antlr.Parser { return s.parser }

func (s *Sfn_nameContext) IDENTIFIER() antlr.TerminalNode {
	return s.GetToken(ActionParserIDENTIFIER, 0)
}

func (s *Sfn_nameContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Sfn_nameContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Sfn_nameContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case ActionParserVisitor:
		return t.VisitSfn_name(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *ActionParser) Sfn_name() (localctx ISfn_nameContext) {
	localctx = NewSfn_nameContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 24, ActionParserRULE_sfn_name)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(81)
		p.Match(ActionParserIDENTIFIER)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IFn_arg_listContext is an interface to support dynamic dispatch.
type IFn_arg_listContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllFn_arg_expr() []IFn_arg_exprContext
	Fn_arg_expr(i int) IFn_arg_exprContext
	AllCOMMA() []antlr.TerminalNode
	COMMA(i int) antlr.TerminalNode

	// IsFn_arg_listContext differentiates from other interfaces.
	IsFn_arg_listContext()
}

type Fn_arg_listContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyFn_arg_listContext() *Fn_arg_listContext {
	var p = new(Fn_arg_listContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ActionParserRULE_fn_arg_list
	return p
}

func InitEmptyFn_arg_listContext(p *Fn_arg_listContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ActionParserRULE_fn_arg_list
}

func (*Fn_arg_listContext) IsFn_arg_listContext() {}

func NewFn_arg_listContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Fn_arg_listContext {
	var p = new(Fn_arg_listContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = ActionParserRULE_fn_arg_list

	return p
}

func (s *Fn_arg_listContext) GetParser() antlr.Parser { return s.parser }

func (s *Fn_arg_listContext) AllFn_arg_expr() []IFn_arg_exprContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IFn_arg_exprContext); ok {
			len++
		}
	}

	tst := make([]IFn_arg_exprContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IFn_arg_exprContext); ok {
			tst[i] = t.(IFn_arg_exprContext)
			i++
		}
	}

	return tst
}

func (s *Fn_arg_listContext) Fn_arg_expr(i int) IFn_arg_exprContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IFn_arg_exprContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IFn_arg_exprContext)
}

func (s *Fn_arg_listContext) AllCOMMA() []antlr.TerminalNode {
	return s.GetTokens(ActionParserCOMMA)
}

func (s *Fn_arg_listContext) COMMA(i int) antlr.TerminalNode {
	return s.GetToken(ActionParserCOMMA, i)
}

func (s *Fn_arg_listContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Fn_arg_listContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Fn_arg_listContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case ActionParserVisitor:
		return t.VisitFn_arg_list(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *ActionParser) Fn_arg_list() (localctx IFn_arg_listContext) {
	localctx = NewFn_arg_listContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 26, ActionParserRULE_fn_arg_list)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(84)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if (int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&33319552516) != 0 {
		{
			p.SetState(83)
			p.fn_arg_expr(0)
		}

	}
	p.SetState(90)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == ActionParserCOMMA {
		{
			p.SetState(86)
			p.Match(ActionParserCOMMA)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(87)
			p.fn_arg_expr(0)
		}

		p.SetState(92)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IFn_arg_exprContext is an interface to support dynamic dispatch.
type IFn_arg_exprContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// GetElevate_expr returns the elevate_expr rule contexts.
	GetElevate_expr() IFn_arg_exprContext

	// GetUnary_expr returns the unary_expr rule contexts.
	GetUnary_expr() IFn_arg_exprContext

	// SetElevate_expr sets the elevate_expr rule contexts.
	SetElevate_expr(IFn_arg_exprContext)

	// SetUnary_expr sets the unary_expr rule contexts.
	SetUnary_expr(IFn_arg_exprContext)

	// Getter signatures
	Literal_value() ILiteral_valueContext
	Variable() IVariableContext
	Block_var() IBlock_varContext
	Sfn_name() ISfn_nameContext
	L_PAREN() antlr.TerminalNode
	R_PAREN() antlr.TerminalNode
	STAR() antlr.TerminalNode
	AllFn_arg_expr() []IFn_arg_exprContext
	Fn_arg_expr(i int) IFn_arg_exprContext
	AllCOMMA() []antlr.TerminalNode
	COMMA(i int) antlr.TerminalNode
	MINUS() antlr.TerminalNode
	PLUS() antlr.TerminalNode
	NOT_() antlr.TerminalNode
	DIV() antlr.TerminalNode
	MOD() antlr.TerminalNode
	LT() antlr.TerminalNode
	LT_EQ() antlr.TerminalNode
	GT() antlr.TerminalNode
	GT_EQ() antlr.TerminalNode
	ASSIGN() antlr.TerminalNode
	SQL_NOT_EQ1() antlr.TerminalNode
	SQL_NOT_EQ2() antlr.TerminalNode
	AND_() antlr.TerminalNode
	OR_() antlr.TerminalNode

	// IsFn_arg_exprContext differentiates from other interfaces.
	IsFn_arg_exprContext()
}

type Fn_arg_exprContext struct {
	antlr.BaseParserRuleContext
	parser       antlr.Parser
	elevate_expr IFn_arg_exprContext
	unary_expr   IFn_arg_exprContext
}

func NewEmptyFn_arg_exprContext() *Fn_arg_exprContext {
	var p = new(Fn_arg_exprContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ActionParserRULE_fn_arg_expr
	return p
}

func InitEmptyFn_arg_exprContext(p *Fn_arg_exprContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ActionParserRULE_fn_arg_expr
}

func (*Fn_arg_exprContext) IsFn_arg_exprContext() {}

func NewFn_arg_exprContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Fn_arg_exprContext {
	var p = new(Fn_arg_exprContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = ActionParserRULE_fn_arg_expr

	return p
}

func (s *Fn_arg_exprContext) GetParser() antlr.Parser { return s.parser }

func (s *Fn_arg_exprContext) GetElevate_expr() IFn_arg_exprContext { return s.elevate_expr }

func (s *Fn_arg_exprContext) GetUnary_expr() IFn_arg_exprContext { return s.unary_expr }

func (s *Fn_arg_exprContext) SetElevate_expr(v IFn_arg_exprContext) { s.elevate_expr = v }

func (s *Fn_arg_exprContext) SetUnary_expr(v IFn_arg_exprContext) { s.unary_expr = v }

func (s *Fn_arg_exprContext) Literal_value() ILiteral_valueContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ILiteral_valueContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ILiteral_valueContext)
}

func (s *Fn_arg_exprContext) Variable() IVariableContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IVariableContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IVariableContext)
}

func (s *Fn_arg_exprContext) Block_var() IBlock_varContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IBlock_varContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IBlock_varContext)
}

func (s *Fn_arg_exprContext) Sfn_name() ISfn_nameContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ISfn_nameContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ISfn_nameContext)
}

func (s *Fn_arg_exprContext) L_PAREN() antlr.TerminalNode {
	return s.GetToken(ActionParserL_PAREN, 0)
}

func (s *Fn_arg_exprContext) R_PAREN() antlr.TerminalNode {
	return s.GetToken(ActionParserR_PAREN, 0)
}

func (s *Fn_arg_exprContext) STAR() antlr.TerminalNode {
	return s.GetToken(ActionParserSTAR, 0)
}

func (s *Fn_arg_exprContext) AllFn_arg_expr() []IFn_arg_exprContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IFn_arg_exprContext); ok {
			len++
		}
	}

	tst := make([]IFn_arg_exprContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IFn_arg_exprContext); ok {
			tst[i] = t.(IFn_arg_exprContext)
			i++
		}
	}

	return tst
}

func (s *Fn_arg_exprContext) Fn_arg_expr(i int) IFn_arg_exprContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IFn_arg_exprContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IFn_arg_exprContext)
}

func (s *Fn_arg_exprContext) AllCOMMA() []antlr.TerminalNode {
	return s.GetTokens(ActionParserCOMMA)
}

func (s *Fn_arg_exprContext) COMMA(i int) antlr.TerminalNode {
	return s.GetToken(ActionParserCOMMA, i)
}

func (s *Fn_arg_exprContext) MINUS() antlr.TerminalNode {
	return s.GetToken(ActionParserMINUS, 0)
}

func (s *Fn_arg_exprContext) PLUS() antlr.TerminalNode {
	return s.GetToken(ActionParserPLUS, 0)
}

func (s *Fn_arg_exprContext) NOT_() antlr.TerminalNode {
	return s.GetToken(ActionParserNOT_, 0)
}

func (s *Fn_arg_exprContext) DIV() antlr.TerminalNode {
	return s.GetToken(ActionParserDIV, 0)
}

func (s *Fn_arg_exprContext) MOD() antlr.TerminalNode {
	return s.GetToken(ActionParserMOD, 0)
}

func (s *Fn_arg_exprContext) LT() antlr.TerminalNode {
	return s.GetToken(ActionParserLT, 0)
}

func (s *Fn_arg_exprContext) LT_EQ() antlr.TerminalNode {
	return s.GetToken(ActionParserLT_EQ, 0)
}

func (s *Fn_arg_exprContext) GT() antlr.TerminalNode {
	return s.GetToken(ActionParserGT, 0)
}

func (s *Fn_arg_exprContext) GT_EQ() antlr.TerminalNode {
	return s.GetToken(ActionParserGT_EQ, 0)
}

func (s *Fn_arg_exprContext) ASSIGN() antlr.TerminalNode {
	return s.GetToken(ActionParserASSIGN, 0)
}

func (s *Fn_arg_exprContext) SQL_NOT_EQ1() antlr.TerminalNode {
	return s.GetToken(ActionParserSQL_NOT_EQ1, 0)
}

func (s *Fn_arg_exprContext) SQL_NOT_EQ2() antlr.TerminalNode {
	return s.GetToken(ActionParserSQL_NOT_EQ2, 0)
}

func (s *Fn_arg_exprContext) AND_() antlr.TerminalNode {
	return s.GetToken(ActionParserAND_, 0)
}

func (s *Fn_arg_exprContext) OR_() antlr.TerminalNode {
	return s.GetToken(ActionParserOR_, 0)
}

func (s *Fn_arg_exprContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Fn_arg_exprContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Fn_arg_exprContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case ActionParserVisitor:
		return t.VisitFn_arg_expr(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *ActionParser) Fn_arg_expr() (localctx IFn_arg_exprContext) {
	return p.fn_arg_expr(0)
}

func (p *ActionParser) fn_arg_expr(_p int) (localctx IFn_arg_exprContext) {
	var _parentctx antlr.ParserRuleContext = p.GetParserRuleContext()

	_parentState := p.GetState()
	localctx = NewFn_arg_exprContext(p, p.GetParserRuleContext(), _parentState)
	var _prevctx IFn_arg_exprContext = localctx
	var _ antlr.ParserRuleContext = _prevctx // TODO: To prevent unused variable warning.
	_startState := 28
	p.EnterRecursionRule(localctx, 28, ActionParserRULE_fn_arg_expr, _p)
	var _la int

	var _alt int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(120)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case ActionParserUNSIGNED_NUMBER_LITERAL, ActionParserSTRING_LITERAL:
		{
			p.SetState(94)
			p.Literal_value()
		}

	case ActionParserVARIABLE:
		{
			p.SetState(95)
			p.Variable()
		}

	case ActionParserBLOCK_VARIABLE:
		{
			p.SetState(96)
			p.Block_var()
		}

	case ActionParserIDENTIFIER:
		{
			p.SetState(97)
			p.Sfn_name()
		}
		{
			p.SetState(98)
			p.Match(ActionParserL_PAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(108)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		switch p.GetTokenStream().LA(1) {
		case ActionParserL_PAREN, ActionParserPLUS, ActionParserMINUS, ActionParserNOT_, ActionParserIDENTIFIER, ActionParserVARIABLE, ActionParserBLOCK_VARIABLE, ActionParserUNSIGNED_NUMBER_LITERAL, ActionParserSTRING_LITERAL:
			{
				p.SetState(99)
				p.fn_arg_expr(0)
			}
			p.SetState(104)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)

			for _la == ActionParserCOMMA {
				{
					p.SetState(100)
					p.Match(ActionParserCOMMA)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				{
					p.SetState(101)
					p.fn_arg_expr(0)
				}

				p.SetState(106)
				p.GetErrorHandler().Sync(p)
				if p.HasError() {
					goto errorExit
				}
				_la = p.GetTokenStream().LA(1)
			}

		case ActionParserSTAR:
			{
				p.SetState(107)
				p.Match(ActionParserSTAR)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

		case ActionParserR_PAREN:

		default:
		}
		{
			p.SetState(110)
			p.Match(ActionParserR_PAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case ActionParserL_PAREN:
		{
			p.SetState(112)
			p.Match(ActionParserL_PAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(113)

			var _x = p.fn_arg_expr(0)

			localctx.(*Fn_arg_exprContext).elevate_expr = _x
		}
		{
			p.SetState(114)
			p.Match(ActionParserR_PAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case ActionParserPLUS, ActionParserMINUS:
		{
			p.SetState(116)
			_la = p.GetTokenStream().LA(1)

			if !(_la == ActionParserPLUS || _la == ActionParserMINUS) {
				p.GetErrorHandler().RecoverInline(p)
			} else {
				p.GetErrorHandler().ReportMatch(p)
				p.Consume()
			}
		}
		{
			p.SetState(117)

			var _x = p.fn_arg_expr(8)

			localctx.(*Fn_arg_exprContext).unary_expr = _x
		}

	case ActionParserNOT_:
		{
			p.SetState(118)
			p.Match(ActionParserNOT_)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(119)

			var _x = p.fn_arg_expr(3)

			localctx.(*Fn_arg_exprContext).unary_expr = _x
		}

	default:
		p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
		goto errorExit
	}
	p.GetParserRuleContext().SetStop(p.GetTokenStream().LT(-1))
	p.SetState(142)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 11, p.GetParserRuleContext())
	if p.HasError() {
		goto errorExit
	}
	for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
		if _alt == 1 {
			if p.GetParseListeners() != nil {
				p.TriggerExitRuleEvent()
			}
			_prevctx = localctx
			p.SetState(140)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}

			switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 10, p.GetParserRuleContext()) {
			case 1:
				localctx = NewFn_arg_exprContext(p, _parentctx, _parentState)
				p.PushNewRecursionContext(localctx, _startState, ActionParserRULE_fn_arg_expr)
				p.SetState(122)

				if !(p.Precpred(p.GetParserRuleContext(), 7)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 7)", ""))
					goto errorExit
				}
				{
					p.SetState(123)
					_la = p.GetTokenStream().LA(1)

					if !((int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&14336) != 0) {
						p.GetErrorHandler().RecoverInline(p)
					} else {
						p.GetErrorHandler().ReportMatch(p)
						p.Consume()
					}
				}
				{
					p.SetState(124)
					p.fn_arg_expr(8)
				}

			case 2:
				localctx = NewFn_arg_exprContext(p, _parentctx, _parentState)
				p.PushNewRecursionContext(localctx, _startState, ActionParserRULE_fn_arg_expr)
				p.SetState(125)

				if !(p.Precpred(p.GetParserRuleContext(), 6)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 6)", ""))
					goto errorExit
				}
				{
					p.SetState(126)
					_la = p.GetTokenStream().LA(1)

					if !(_la == ActionParserPLUS || _la == ActionParserMINUS) {
						p.GetErrorHandler().RecoverInline(p)
					} else {
						p.GetErrorHandler().ReportMatch(p)
						p.Consume()
					}
				}
				{
					p.SetState(127)
					p.fn_arg_expr(7)
				}

			case 3:
				localctx = NewFn_arg_exprContext(p, _parentctx, _parentState)
				p.PushNewRecursionContext(localctx, _startState, ActionParserRULE_fn_arg_expr)
				p.SetState(128)

				if !(p.Precpred(p.GetParserRuleContext(), 5)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 5)", ""))
					goto errorExit
				}
				{
					p.SetState(129)
					_la = p.GetTokenStream().LA(1)

					if !((int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&245760) != 0) {
						p.GetErrorHandler().RecoverInline(p)
					} else {
						p.GetErrorHandler().ReportMatch(p)
						p.Consume()
					}
				}
				{
					p.SetState(130)
					p.fn_arg_expr(6)
				}

			case 4:
				localctx = NewFn_arg_exprContext(p, _parentctx, _parentState)
				p.PushNewRecursionContext(localctx, _startState, ActionParserRULE_fn_arg_expr)
				p.SetState(131)

				if !(p.Precpred(p.GetParserRuleContext(), 4)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 4)", ""))
					goto errorExit
				}
				{
					p.SetState(132)
					_la = p.GetTokenStream().LA(1)

					if !((int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&786560) != 0) {
						p.GetErrorHandler().RecoverInline(p)
					} else {
						p.GetErrorHandler().ReportMatch(p)
						p.Consume()
					}
				}
				{
					p.SetState(133)
					p.fn_arg_expr(5)
				}

			case 5:
				localctx = NewFn_arg_exprContext(p, _parentctx, _parentState)
				p.PushNewRecursionContext(localctx, _startState, ActionParserRULE_fn_arg_expr)
				p.SetState(134)

				if !(p.Precpred(p.GetParserRuleContext(), 2)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 2)", ""))
					goto errorExit
				}
				{
					p.SetState(135)
					p.Match(ActionParserAND_)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				{
					p.SetState(136)
					p.fn_arg_expr(3)
				}

			case 6:
				localctx = NewFn_arg_exprContext(p, _parentctx, _parentState)
				p.PushNewRecursionContext(localctx, _startState, ActionParserRULE_fn_arg_expr)
				p.SetState(137)

				if !(p.Precpred(p.GetParserRuleContext(), 1)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 1)", ""))
					goto errorExit
				}
				{
					p.SetState(138)
					p.Match(ActionParserOR_)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				{
					p.SetState(139)
					p.fn_arg_expr(2)
				}

			case antlr.ATNInvalidAltNumber:
				goto errorExit
			}

		}
		p.SetState(144)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 11, p.GetParserRuleContext())
		if p.HasError() {
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.UnrollRecursionContexts(_parentctx)
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

func (p *ActionParser) Sempred(localctx antlr.RuleContext, ruleIndex, predIndex int) bool {
	switch ruleIndex {
	case 14:
		var t *Fn_arg_exprContext = nil
		if localctx != nil {
			t = localctx.(*Fn_arg_exprContext)
		}
		return p.Fn_arg_expr_Sempred(t, predIndex)

	default:
		panic("No predicate with index: " + fmt.Sprint(ruleIndex))
	}
}

func (p *ActionParser) Fn_arg_expr_Sempred(localctx antlr.RuleContext, predIndex int) bool {
	switch predIndex {
	case 0:
		return p.Precpred(p.GetParserRuleContext(), 7)

	case 1:
		return p.Precpred(p.GetParserRuleContext(), 6)

	case 2:
		return p.Precpred(p.GetParserRuleContext(), 5)

	case 3:
		return p.Precpred(p.GetParserRuleContext(), 4)

	case 4:
		return p.Precpred(p.GetParserRuleContext(), 2)

	case 5:
		return p.Precpred(p.GetParserRuleContext(), 1)

	default:
		panic("No predicate with index: " + fmt.Sprint(predIndex))
	}
}
