// Code generated from ProcedureParser.g4 by ANTLR 4.13.1. DO NOT EDIT.

package gen // ProcedureParser
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

type ProcedureParser struct {
	*antlr.BaseParser
}

var ProcedureParserParserStaticData struct {
	once                   sync.Once
	serializedATN          []int32
	LiteralNames           []string
	SymbolicNames          []string
	RuleNames              []string
	PredictionContextCache *antlr.PredictionContextCache
	atn                    *antlr.ATN
	decisionToDFA          []*antlr.DFA
}

func procedureparserParserInit() {
	staticData := &ProcedureParserParserStaticData
	staticData.LiteralNames = []string{
		"", "';'", "'('", "')'", "'{'", "'}'", "','", "'::'", "':'", "'$'",
		"'@'", "':='", "'.'", "'['", "']'", "'''", "'_'", "'+'", "'-'", "'*'",
		"'/'", "'%'", "'<'", "'<='", "'>'", "'>='", "'!='", "'=='", "", "'for'",
		"'in'", "'if'", "'elseif'", "'else'", "'to'", "'return'", "'break'",
		"'next'", "", "", "", "", "'null'",
	}
	staticData.SymbolicNames = []string{
		"", "SEMICOLON", "LPAREN", "RPAREN", "LBRACE", "RBRACE", "COMMA", "TYPE_CAST",
		"COLON", "DOLLAR", "AT", "ASSIGN", "PERIOD", "LBRACKET", "RBRACKET",
		"SINGLE_QUOTE", "UNDERSCORE", "PLUS", "MINUS", "MUL", "DIV", "MOD",
		"LT", "LT_EQ", "GT", "GT_EQ", "NEQ", "EQ", "ANY_SQL", "FOR", "IN", "IF",
		"ELSEIF", "ELSE", "TO", "RETURN", "BREAK", "NEXT", "BOOLEAN_LITERAL",
		"INT_LITERAL", "BLOB_LITERAL", "TEXT_LITERAL", "NULL_LITERAL", "IDENTIFIER",
		"VARIABLE", "WS", "TERMINATOR", "BLOCK_COMMENT", "LINE_COMMENT",
	}
	staticData.RuleNames = []string{
		"program", "statement", "variable_or_underscore", "type", "expression",
		"expression_list", "expression_make_array", "call_expression", "range",
		"if_then_block",
	}
	staticData.PredictionContextCache = antlr.NewPredictionContextCache()
	staticData.serializedATN = []int32{
		4, 1, 48, 190, 2, 0, 7, 0, 2, 1, 7, 1, 2, 2, 7, 2, 2, 3, 7, 3, 2, 4, 7,
		4, 2, 5, 7, 5, 2, 6, 7, 6, 2, 7, 7, 7, 2, 8, 7, 8, 2, 9, 7, 9, 1, 0, 5,
		0, 22, 8, 0, 10, 0, 12, 0, 25, 9, 0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
		1, 1, 5, 1, 34, 8, 1, 10, 1, 12, 1, 37, 9, 1, 1, 1, 1, 1, 3, 1, 41, 8,
		1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
		1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 3, 1, 64, 8, 1,
		1, 1, 1, 1, 5, 1, 68, 8, 1, 10, 1, 12, 1, 71, 9, 1, 1, 1, 1, 1, 1, 1, 1,
		1, 1, 1, 5, 1, 78, 8, 1, 10, 1, 12, 1, 81, 9, 1, 1, 1, 1, 1, 1, 1, 5, 1,
		86, 8, 1, 10, 1, 12, 1, 89, 9, 1, 1, 1, 3, 1, 92, 8, 1, 1, 1, 1, 1, 1,
		1, 1, 1, 1, 1, 1, 1, 1, 1, 3, 1, 101, 8, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
		1, 1, 1, 3, 1, 109, 8, 1, 1, 2, 1, 2, 1, 3, 1, 3, 1, 3, 3, 3, 116, 8, 3,
		1, 4, 1, 4, 1, 4, 1, 4, 1, 4, 1, 4, 1, 4, 1, 4, 1, 4, 1, 4, 1, 4, 1, 4,
		1, 4, 3, 4, 131, 8, 4, 1, 4, 1, 4, 1, 4, 1, 4, 1, 4, 1, 4, 1, 4, 1, 4,
		1, 4, 1, 4, 1, 4, 1, 4, 1, 4, 1, 4, 1, 4, 1, 4, 1, 4, 5, 4, 150, 8, 4,
		10, 4, 12, 4, 153, 9, 4, 1, 5, 1, 5, 1, 5, 5, 5, 158, 8, 5, 10, 5, 12,
		5, 161, 9, 5, 1, 6, 1, 6, 3, 6, 165, 8, 6, 1, 6, 1, 6, 1, 7, 1, 7, 1, 7,
		3, 7, 172, 8, 7, 1, 7, 1, 7, 1, 8, 1, 8, 1, 8, 1, 8, 1, 9, 1, 9, 1, 9,
		5, 9, 183, 8, 9, 10, 9, 12, 9, 186, 9, 9, 1, 9, 1, 9, 1, 9, 0, 1, 8, 10,
		0, 2, 4, 6, 8, 10, 12, 14, 16, 18, 0, 4, 2, 0, 16, 16, 44, 44, 1, 0, 22,
		27, 1, 0, 19, 21, 1, 0, 17, 18, 217, 0, 23, 1, 0, 0, 0, 2, 108, 1, 0, 0,
		0, 4, 110, 1, 0, 0, 0, 6, 112, 1, 0, 0, 0, 8, 130, 1, 0, 0, 0, 10, 154,
		1, 0, 0, 0, 12, 162, 1, 0, 0, 0, 14, 168, 1, 0, 0, 0, 16, 175, 1, 0, 0,
		0, 18, 179, 1, 0, 0, 0, 20, 22, 3, 2, 1, 0, 21, 20, 1, 0, 0, 0, 22, 25,
		1, 0, 0, 0, 23, 21, 1, 0, 0, 0, 23, 24, 1, 0, 0, 0, 24, 1, 1, 0, 0, 0,
		25, 23, 1, 0, 0, 0, 26, 27, 5, 44, 0, 0, 27, 28, 3, 6, 3, 0, 28, 29, 5,
		1, 0, 0, 29, 109, 1, 0, 0, 0, 30, 35, 3, 4, 2, 0, 31, 32, 5, 6, 0, 0, 32,
		34, 3, 4, 2, 0, 33, 31, 1, 0, 0, 0, 34, 37, 1, 0, 0, 0, 35, 33, 1, 0, 0,
		0, 35, 36, 1, 0, 0, 0, 36, 38, 1, 0, 0, 0, 37, 35, 1, 0, 0, 0, 38, 39,
		5, 11, 0, 0, 39, 41, 1, 0, 0, 0, 40, 30, 1, 0, 0, 0, 40, 41, 1, 0, 0, 0,
		41, 42, 1, 0, 0, 0, 42, 43, 3, 14, 7, 0, 43, 44, 5, 1, 0, 0, 44, 109, 1,
		0, 0, 0, 45, 46, 5, 44, 0, 0, 46, 47, 5, 11, 0, 0, 47, 48, 3, 8, 4, 0,
		48, 49, 5, 1, 0, 0, 49, 109, 1, 0, 0, 0, 50, 51, 5, 44, 0, 0, 51, 52, 3,
		6, 3, 0, 52, 53, 5, 11, 0, 0, 53, 54, 3, 8, 4, 0, 54, 55, 5, 1, 0, 0, 55,
		109, 1, 0, 0, 0, 56, 57, 5, 29, 0, 0, 57, 58, 5, 44, 0, 0, 58, 63, 5, 30,
		0, 0, 59, 64, 3, 16, 8, 0, 60, 64, 3, 14, 7, 0, 61, 64, 5, 44, 0, 0, 62,
		64, 5, 28, 0, 0, 63, 59, 1, 0, 0, 0, 63, 60, 1, 0, 0, 0, 63, 61, 1, 0,
		0, 0, 63, 62, 1, 0, 0, 0, 64, 65, 1, 0, 0, 0, 65, 69, 5, 4, 0, 0, 66, 68,
		3, 2, 1, 0, 67, 66, 1, 0, 0, 0, 68, 71, 1, 0, 0, 0, 69, 67, 1, 0, 0, 0,
		69, 70, 1, 0, 0, 0, 70, 72, 1, 0, 0, 0, 71, 69, 1, 0, 0, 0, 72, 109, 5,
		5, 0, 0, 73, 74, 5, 31, 0, 0, 74, 79, 3, 18, 9, 0, 75, 76, 5, 32, 0, 0,
		76, 78, 3, 18, 9, 0, 77, 75, 1, 0, 0, 0, 78, 81, 1, 0, 0, 0, 79, 77, 1,
		0, 0, 0, 79, 80, 1, 0, 0, 0, 80, 91, 1, 0, 0, 0, 81, 79, 1, 0, 0, 0, 82,
		83, 5, 33, 0, 0, 83, 87, 5, 4, 0, 0, 84, 86, 3, 2, 1, 0, 85, 84, 1, 0,
		0, 0, 86, 89, 1, 0, 0, 0, 87, 85, 1, 0, 0, 0, 87, 88, 1, 0, 0, 0, 88, 90,
		1, 0, 0, 0, 89, 87, 1, 0, 0, 0, 90, 92, 5, 5, 0, 0, 91, 82, 1, 0, 0, 0,
		91, 92, 1, 0, 0, 0, 92, 109, 1, 0, 0, 0, 93, 94, 5, 28, 0, 0, 94, 109,
		5, 1, 0, 0, 95, 96, 5, 36, 0, 0, 96, 109, 5, 1, 0, 0, 97, 100, 5, 35, 0,
		0, 98, 101, 3, 10, 5, 0, 99, 101, 5, 28, 0, 0, 100, 98, 1, 0, 0, 0, 100,
		99, 1, 0, 0, 0, 101, 102, 1, 0, 0, 0, 102, 109, 5, 1, 0, 0, 103, 104, 5,
		35, 0, 0, 104, 105, 5, 37, 0, 0, 105, 106, 3, 10, 5, 0, 106, 107, 5, 1,
		0, 0, 107, 109, 1, 0, 0, 0, 108, 26, 1, 0, 0, 0, 108, 40, 1, 0, 0, 0, 108,
		45, 1, 0, 0, 0, 108, 50, 1, 0, 0, 0, 108, 56, 1, 0, 0, 0, 108, 73, 1, 0,
		0, 0, 108, 93, 1, 0, 0, 0, 108, 95, 1, 0, 0, 0, 108, 97, 1, 0, 0, 0, 108,
		103, 1, 0, 0, 0, 109, 3, 1, 0, 0, 0, 110, 111, 7, 0, 0, 0, 111, 5, 1, 0,
		0, 0, 112, 115, 5, 43, 0, 0, 113, 114, 5, 13, 0, 0, 114, 116, 5, 14, 0,
		0, 115, 113, 1, 0, 0, 0, 115, 116, 1, 0, 0, 0, 116, 7, 1, 0, 0, 0, 117,
		118, 6, 4, -1, 0, 118, 131, 5, 41, 0, 0, 119, 131, 5, 38, 0, 0, 120, 131,
		5, 39, 0, 0, 121, 131, 5, 42, 0, 0, 122, 131, 5, 40, 0, 0, 123, 131, 3,
		12, 6, 0, 124, 131, 3, 14, 7, 0, 125, 131, 5, 44, 0, 0, 126, 127, 5, 2,
		0, 0, 127, 128, 3, 8, 4, 0, 128, 129, 5, 3, 0, 0, 129, 131, 1, 0, 0, 0,
		130, 117, 1, 0, 0, 0, 130, 119, 1, 0, 0, 0, 130, 120, 1, 0, 0, 0, 130,
		121, 1, 0, 0, 0, 130, 122, 1, 0, 0, 0, 130, 123, 1, 0, 0, 0, 130, 124,
		1, 0, 0, 0, 130, 125, 1, 0, 0, 0, 130, 126, 1, 0, 0, 0, 131, 151, 1, 0,
		0, 0, 132, 133, 10, 3, 0, 0, 133, 134, 7, 1, 0, 0, 134, 150, 3, 8, 4, 4,
		135, 136, 10, 2, 0, 0, 136, 137, 7, 2, 0, 0, 137, 150, 3, 8, 4, 3, 138,
		139, 10, 1, 0, 0, 139, 140, 7, 3, 0, 0, 140, 150, 3, 8, 4, 2, 141, 142,
		10, 6, 0, 0, 142, 143, 5, 13, 0, 0, 143, 144, 3, 8, 4, 0, 144, 145, 5,
		14, 0, 0, 145, 150, 1, 0, 0, 0, 146, 147, 10, 5, 0, 0, 147, 148, 5, 12,
		0, 0, 148, 150, 5, 43, 0, 0, 149, 132, 1, 0, 0, 0, 149, 135, 1, 0, 0, 0,
		149, 138, 1, 0, 0, 0, 149, 141, 1, 0, 0, 0, 149, 146, 1, 0, 0, 0, 150,
		153, 1, 0, 0, 0, 151, 149, 1, 0, 0, 0, 151, 152, 1, 0, 0, 0, 152, 9, 1,
		0, 0, 0, 153, 151, 1, 0, 0, 0, 154, 159, 3, 8, 4, 0, 155, 156, 5, 6, 0,
		0, 156, 158, 3, 8, 4, 0, 157, 155, 1, 0, 0, 0, 158, 161, 1, 0, 0, 0, 159,
		157, 1, 0, 0, 0, 159, 160, 1, 0, 0, 0, 160, 11, 1, 0, 0, 0, 161, 159, 1,
		0, 0, 0, 162, 164, 5, 13, 0, 0, 163, 165, 3, 10, 5, 0, 164, 163, 1, 0,
		0, 0, 164, 165, 1, 0, 0, 0, 165, 166, 1, 0, 0, 0, 166, 167, 5, 14, 0, 0,
		167, 13, 1, 0, 0, 0, 168, 169, 5, 43, 0, 0, 169, 171, 5, 2, 0, 0, 170,
		172, 3, 10, 5, 0, 171, 170, 1, 0, 0, 0, 171, 172, 1, 0, 0, 0, 172, 173,
		1, 0, 0, 0, 173, 174, 5, 3, 0, 0, 174, 15, 1, 0, 0, 0, 175, 176, 3, 8,
		4, 0, 176, 177, 5, 8, 0, 0, 177, 178, 3, 8, 4, 0, 178, 17, 1, 0, 0, 0,
		179, 180, 3, 8, 4, 0, 180, 184, 5, 4, 0, 0, 181, 183, 3, 2, 1, 0, 182,
		181, 1, 0, 0, 0, 183, 186, 1, 0, 0, 0, 184, 182, 1, 0, 0, 0, 184, 185,
		1, 0, 0, 0, 185, 187, 1, 0, 0, 0, 186, 184, 1, 0, 0, 0, 187, 188, 5, 5,
		0, 0, 188, 19, 1, 0, 0, 0, 18, 23, 35, 40, 63, 69, 79, 87, 91, 100, 108,
		115, 130, 149, 151, 159, 164, 171, 184,
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

// ProcedureParserInit initializes any static state used to implement ProcedureParser. By default the
// static state used to implement the parser is lazily initialized during the first call to
// NewProcedureParser(). You can call this function if you wish to initialize the static state ahead
// of time.
func ProcedureParserInit() {
	staticData := &ProcedureParserParserStaticData
	staticData.once.Do(procedureparserParserInit)
}

// NewProcedureParser produces a new parser instance for the optional input antlr.TokenStream.
func NewProcedureParser(input antlr.TokenStream) *ProcedureParser {
	ProcedureParserInit()
	this := new(ProcedureParser)
	this.BaseParser = antlr.NewBaseParser(input)
	staticData := &ProcedureParserParserStaticData
	this.Interpreter = antlr.NewParserATNSimulator(this, staticData.atn, staticData.decisionToDFA, staticData.PredictionContextCache)
	this.RuleNames = staticData.RuleNames
	this.LiteralNames = staticData.LiteralNames
	this.SymbolicNames = staticData.SymbolicNames
	this.GrammarFileName = "ProcedureParser.g4"

	return this
}

// ProcedureParser tokens.
const (
	ProcedureParserEOF             = antlr.TokenEOF
	ProcedureParserSEMICOLON       = 1
	ProcedureParserLPAREN          = 2
	ProcedureParserRPAREN          = 3
	ProcedureParserLBRACE          = 4
	ProcedureParserRBRACE          = 5
	ProcedureParserCOMMA           = 6
	ProcedureParserTYPE_CAST       = 7
	ProcedureParserCOLON           = 8
	ProcedureParserDOLLAR          = 9
	ProcedureParserAT              = 10
	ProcedureParserASSIGN          = 11
	ProcedureParserPERIOD          = 12
	ProcedureParserLBRACKET        = 13
	ProcedureParserRBRACKET        = 14
	ProcedureParserSINGLE_QUOTE    = 15
	ProcedureParserUNDERSCORE      = 16
	ProcedureParserPLUS            = 17
	ProcedureParserMINUS           = 18
	ProcedureParserMUL             = 19
	ProcedureParserDIV             = 20
	ProcedureParserMOD             = 21
	ProcedureParserLT              = 22
	ProcedureParserLT_EQ           = 23
	ProcedureParserGT              = 24
	ProcedureParserGT_EQ           = 25
	ProcedureParserNEQ             = 26
	ProcedureParserEQ              = 27
	ProcedureParserANY_SQL         = 28
	ProcedureParserFOR             = 29
	ProcedureParserIN              = 30
	ProcedureParserIF              = 31
	ProcedureParserELSEIF          = 32
	ProcedureParserELSE            = 33
	ProcedureParserTO              = 34
	ProcedureParserRETURN          = 35
	ProcedureParserBREAK           = 36
	ProcedureParserNEXT            = 37
	ProcedureParserBOOLEAN_LITERAL = 38
	ProcedureParserINT_LITERAL     = 39
	ProcedureParserBLOB_LITERAL    = 40
	ProcedureParserTEXT_LITERAL    = 41
	ProcedureParserNULL_LITERAL    = 42
	ProcedureParserIDENTIFIER      = 43
	ProcedureParserVARIABLE        = 44
	ProcedureParserWS              = 45
	ProcedureParserTERMINATOR      = 46
	ProcedureParserBLOCK_COMMENT   = 47
	ProcedureParserLINE_COMMENT    = 48
)

// ProcedureParser rules.
const (
	ProcedureParserRULE_program                = 0
	ProcedureParserRULE_statement              = 1
	ProcedureParserRULE_variable_or_underscore = 2
	ProcedureParserRULE_type                   = 3
	ProcedureParserRULE_expression             = 4
	ProcedureParserRULE_expression_list        = 5
	ProcedureParserRULE_expression_make_array  = 6
	ProcedureParserRULE_call_expression        = 7
	ProcedureParserRULE_range                  = 8
	ProcedureParserRULE_if_then_block          = 9
)

// IProgramContext is an interface to support dynamic dispatch.
type IProgramContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllStatement() []IStatementContext
	Statement(i int) IStatementContext

	// IsProgramContext differentiates from other interfaces.
	IsProgramContext()
}

type ProgramContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyProgramContext() *ProgramContext {
	var p = new(ProgramContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ProcedureParserRULE_program
	return p
}

func InitEmptyProgramContext(p *ProgramContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ProcedureParserRULE_program
}

func (*ProgramContext) IsProgramContext() {}

func NewProgramContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ProgramContext {
	var p = new(ProgramContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = ProcedureParserRULE_program

	return p
}

func (s *ProgramContext) GetParser() antlr.Parser { return s.parser }

func (s *ProgramContext) AllStatement() []IStatementContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IStatementContext); ok {
			len++
		}
	}

	tst := make([]IStatementContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IStatementContext); ok {
			tst[i] = t.(IStatementContext)
			i++
		}
	}

	return tst
}

func (s *ProgramContext) Statement(i int) IStatementContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IStatementContext); ok {
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

	return t.(IStatementContext)
}

func (s *ProgramContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ProgramContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ProgramContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case ProcedureParserVisitor:
		return t.VisitProgram(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *ProcedureParser) Program() (localctx IProgramContext) {
	localctx = NewProgramContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 0, ProcedureParserRULE_program)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(23)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for (int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&26494311137280) != 0 {
		{
			p.SetState(20)
			p.Statement()
		}

		p.SetState(25)
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

// IStatementContext is an interface to support dynamic dispatch.
type IStatementContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser
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
	p.RuleIndex = ProcedureParserRULE_statement
	return p
}

func InitEmptyStatementContext(p *StatementContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ProcedureParserRULE_statement
}

func (*StatementContext) IsStatementContext() {}

func NewStatementContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *StatementContext {
	var p = new(StatementContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = ProcedureParserRULE_statement

	return p
}

func (s *StatementContext) GetParser() antlr.Parser { return s.parser }

func (s *StatementContext) CopyAll(ctx *StatementContext) {
	s.CopyFrom(&ctx.BaseParserRuleContext)
}

func (s *StatementContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *StatementContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

type Stmt_ifContext struct {
	StatementContext
}

func NewStmt_ifContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *Stmt_ifContext {
	var p = new(Stmt_ifContext)

	InitEmptyStatementContext(&p.StatementContext)
	p.parser = parser
	p.CopyAll(ctx.(*StatementContext))

	return p
}

func (s *Stmt_ifContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Stmt_ifContext) IF() antlr.TerminalNode {
	return s.GetToken(ProcedureParserIF, 0)
}

func (s *Stmt_ifContext) AllIf_then_block() []IIf_then_blockContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IIf_then_blockContext); ok {
			len++
		}
	}

	tst := make([]IIf_then_blockContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IIf_then_blockContext); ok {
			tst[i] = t.(IIf_then_blockContext)
			i++
		}
	}

	return tst
}

func (s *Stmt_ifContext) If_then_block(i int) IIf_then_blockContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IIf_then_blockContext); ok {
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

	return t.(IIf_then_blockContext)
}

func (s *Stmt_ifContext) AllELSEIF() []antlr.TerminalNode {
	return s.GetTokens(ProcedureParserELSEIF)
}

func (s *Stmt_ifContext) ELSEIF(i int) antlr.TerminalNode {
	return s.GetToken(ProcedureParserELSEIF, i)
}

func (s *Stmt_ifContext) ELSE() antlr.TerminalNode {
	return s.GetToken(ProcedureParserELSE, 0)
}

func (s *Stmt_ifContext) LBRACE() antlr.TerminalNode {
	return s.GetToken(ProcedureParserLBRACE, 0)
}

func (s *Stmt_ifContext) RBRACE() antlr.TerminalNode {
	return s.GetToken(ProcedureParserRBRACE, 0)
}

func (s *Stmt_ifContext) AllStatement() []IStatementContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IStatementContext); ok {
			len++
		}
	}

	tst := make([]IStatementContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IStatementContext); ok {
			tst[i] = t.(IStatementContext)
			i++
		}
	}

	return tst
}

func (s *Stmt_ifContext) Statement(i int) IStatementContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IStatementContext); ok {
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

	return t.(IStatementContext)
}

func (s *Stmt_ifContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case ProcedureParserVisitor:
		return t.VisitStmt_if(s)

	default:
		return t.VisitChildren(s)
	}
}

type Stmt_breakContext struct {
	StatementContext
}

func NewStmt_breakContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *Stmt_breakContext {
	var p = new(Stmt_breakContext)

	InitEmptyStatementContext(&p.StatementContext)
	p.parser = parser
	p.CopyAll(ctx.(*StatementContext))

	return p
}

func (s *Stmt_breakContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Stmt_breakContext) BREAK() antlr.TerminalNode {
	return s.GetToken(ProcedureParserBREAK, 0)
}

func (s *Stmt_breakContext) SEMICOLON() antlr.TerminalNode {
	return s.GetToken(ProcedureParserSEMICOLON, 0)
}

func (s *Stmt_breakContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case ProcedureParserVisitor:
		return t.VisitStmt_break(s)

	default:
		return t.VisitChildren(s)
	}
}

type Stmt_variable_assignment_with_declarationContext struct {
	StatementContext
}

func NewStmt_variable_assignment_with_declarationContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *Stmt_variable_assignment_with_declarationContext {
	var p = new(Stmt_variable_assignment_with_declarationContext)

	InitEmptyStatementContext(&p.StatementContext)
	p.parser = parser
	p.CopyAll(ctx.(*StatementContext))

	return p
}

func (s *Stmt_variable_assignment_with_declarationContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Stmt_variable_assignment_with_declarationContext) VARIABLE() antlr.TerminalNode {
	return s.GetToken(ProcedureParserVARIABLE, 0)
}

func (s *Stmt_variable_assignment_with_declarationContext) Type_() ITypeContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITypeContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ITypeContext)
}

func (s *Stmt_variable_assignment_with_declarationContext) ASSIGN() antlr.TerminalNode {
	return s.GetToken(ProcedureParserASSIGN, 0)
}

func (s *Stmt_variable_assignment_with_declarationContext) Expression() IExpressionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *Stmt_variable_assignment_with_declarationContext) SEMICOLON() antlr.TerminalNode {
	return s.GetToken(ProcedureParserSEMICOLON, 0)
}

func (s *Stmt_variable_assignment_with_declarationContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case ProcedureParserVisitor:
		return t.VisitStmt_variable_assignment_with_declaration(s)

	default:
		return t.VisitChildren(s)
	}
}

type Stmt_variable_declarationContext struct {
	StatementContext
}

func NewStmt_variable_declarationContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *Stmt_variable_declarationContext {
	var p = new(Stmt_variable_declarationContext)

	InitEmptyStatementContext(&p.StatementContext)
	p.parser = parser
	p.CopyAll(ctx.(*StatementContext))

	return p
}

func (s *Stmt_variable_declarationContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Stmt_variable_declarationContext) VARIABLE() antlr.TerminalNode {
	return s.GetToken(ProcedureParserVARIABLE, 0)
}

func (s *Stmt_variable_declarationContext) Type_() ITypeContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITypeContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ITypeContext)
}

func (s *Stmt_variable_declarationContext) SEMICOLON() antlr.TerminalNode {
	return s.GetToken(ProcedureParserSEMICOLON, 0)
}

func (s *Stmt_variable_declarationContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case ProcedureParserVisitor:
		return t.VisitStmt_variable_declaration(s)

	default:
		return t.VisitChildren(s)
	}
}

type Stmt_return_nextContext struct {
	StatementContext
}

func NewStmt_return_nextContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *Stmt_return_nextContext {
	var p = new(Stmt_return_nextContext)

	InitEmptyStatementContext(&p.StatementContext)
	p.parser = parser
	p.CopyAll(ctx.(*StatementContext))

	return p
}

func (s *Stmt_return_nextContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Stmt_return_nextContext) RETURN() antlr.TerminalNode {
	return s.GetToken(ProcedureParserRETURN, 0)
}

func (s *Stmt_return_nextContext) NEXT() antlr.TerminalNode {
	return s.GetToken(ProcedureParserNEXT, 0)
}

func (s *Stmt_return_nextContext) Expression_list() IExpression_listContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpression_listContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpression_listContext)
}

func (s *Stmt_return_nextContext) SEMICOLON() antlr.TerminalNode {
	return s.GetToken(ProcedureParserSEMICOLON, 0)
}

func (s *Stmt_return_nextContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case ProcedureParserVisitor:
		return t.VisitStmt_return_next(s)

	default:
		return t.VisitChildren(s)
	}
}

type Stmt_for_loopContext struct {
	StatementContext
}

func NewStmt_for_loopContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *Stmt_for_loopContext {
	var p = new(Stmt_for_loopContext)

	InitEmptyStatementContext(&p.StatementContext)
	p.parser = parser
	p.CopyAll(ctx.(*StatementContext))

	return p
}

func (s *Stmt_for_loopContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Stmt_for_loopContext) FOR() antlr.TerminalNode {
	return s.GetToken(ProcedureParserFOR, 0)
}

func (s *Stmt_for_loopContext) AllVARIABLE() []antlr.TerminalNode {
	return s.GetTokens(ProcedureParserVARIABLE)
}

func (s *Stmt_for_loopContext) VARIABLE(i int) antlr.TerminalNode {
	return s.GetToken(ProcedureParserVARIABLE, i)
}

func (s *Stmt_for_loopContext) IN() antlr.TerminalNode {
	return s.GetToken(ProcedureParserIN, 0)
}

func (s *Stmt_for_loopContext) LBRACE() antlr.TerminalNode {
	return s.GetToken(ProcedureParserLBRACE, 0)
}

func (s *Stmt_for_loopContext) RBRACE() antlr.TerminalNode {
	return s.GetToken(ProcedureParserRBRACE, 0)
}

func (s *Stmt_for_loopContext) Range_() IRangeContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IRangeContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IRangeContext)
}

func (s *Stmt_for_loopContext) Call_expression() ICall_expressionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ICall_expressionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ICall_expressionContext)
}

func (s *Stmt_for_loopContext) ANY_SQL() antlr.TerminalNode {
	return s.GetToken(ProcedureParserANY_SQL, 0)
}

func (s *Stmt_for_loopContext) AllStatement() []IStatementContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IStatementContext); ok {
			len++
		}
	}

	tst := make([]IStatementContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IStatementContext); ok {
			tst[i] = t.(IStatementContext)
			i++
		}
	}

	return tst
}

func (s *Stmt_for_loopContext) Statement(i int) IStatementContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IStatementContext); ok {
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

	return t.(IStatementContext)
}

func (s *Stmt_for_loopContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case ProcedureParserVisitor:
		return t.VisitStmt_for_loop(s)

	default:
		return t.VisitChildren(s)
	}
}

type Stmt_returnContext struct {
	StatementContext
}

func NewStmt_returnContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *Stmt_returnContext {
	var p = new(Stmt_returnContext)

	InitEmptyStatementContext(&p.StatementContext)
	p.parser = parser
	p.CopyAll(ctx.(*StatementContext))

	return p
}

func (s *Stmt_returnContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Stmt_returnContext) RETURN() antlr.TerminalNode {
	return s.GetToken(ProcedureParserRETURN, 0)
}

func (s *Stmt_returnContext) SEMICOLON() antlr.TerminalNode {
	return s.GetToken(ProcedureParserSEMICOLON, 0)
}

func (s *Stmt_returnContext) Expression_list() IExpression_listContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpression_listContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpression_listContext)
}

func (s *Stmt_returnContext) ANY_SQL() antlr.TerminalNode {
	return s.GetToken(ProcedureParserANY_SQL, 0)
}

func (s *Stmt_returnContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case ProcedureParserVisitor:
		return t.VisitStmt_return(s)

	default:
		return t.VisitChildren(s)
	}
}

type Stmt_procedure_callContext struct {
	StatementContext
}

func NewStmt_procedure_callContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *Stmt_procedure_callContext {
	var p = new(Stmt_procedure_callContext)

	InitEmptyStatementContext(&p.StatementContext)
	p.parser = parser
	p.CopyAll(ctx.(*StatementContext))

	return p
}

func (s *Stmt_procedure_callContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Stmt_procedure_callContext) Call_expression() ICall_expressionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ICall_expressionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ICall_expressionContext)
}

func (s *Stmt_procedure_callContext) SEMICOLON() antlr.TerminalNode {
	return s.GetToken(ProcedureParserSEMICOLON, 0)
}

func (s *Stmt_procedure_callContext) AllVariable_or_underscore() []IVariable_or_underscoreContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IVariable_or_underscoreContext); ok {
			len++
		}
	}

	tst := make([]IVariable_or_underscoreContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IVariable_or_underscoreContext); ok {
			tst[i] = t.(IVariable_or_underscoreContext)
			i++
		}
	}

	return tst
}

func (s *Stmt_procedure_callContext) Variable_or_underscore(i int) IVariable_or_underscoreContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IVariable_or_underscoreContext); ok {
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

	return t.(IVariable_or_underscoreContext)
}

func (s *Stmt_procedure_callContext) ASSIGN() antlr.TerminalNode {
	return s.GetToken(ProcedureParserASSIGN, 0)
}

func (s *Stmt_procedure_callContext) AllCOMMA() []antlr.TerminalNode {
	return s.GetTokens(ProcedureParserCOMMA)
}

func (s *Stmt_procedure_callContext) COMMA(i int) antlr.TerminalNode {
	return s.GetToken(ProcedureParserCOMMA, i)
}

func (s *Stmt_procedure_callContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case ProcedureParserVisitor:
		return t.VisitStmt_procedure_call(s)

	default:
		return t.VisitChildren(s)
	}
}

type Stmt_variable_assignmentContext struct {
	StatementContext
}

func NewStmt_variable_assignmentContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *Stmt_variable_assignmentContext {
	var p = new(Stmt_variable_assignmentContext)

	InitEmptyStatementContext(&p.StatementContext)
	p.parser = parser
	p.CopyAll(ctx.(*StatementContext))

	return p
}

func (s *Stmt_variable_assignmentContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Stmt_variable_assignmentContext) VARIABLE() antlr.TerminalNode {
	return s.GetToken(ProcedureParserVARIABLE, 0)
}

func (s *Stmt_variable_assignmentContext) ASSIGN() antlr.TerminalNode {
	return s.GetToken(ProcedureParserASSIGN, 0)
}

func (s *Stmt_variable_assignmentContext) Expression() IExpressionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *Stmt_variable_assignmentContext) SEMICOLON() antlr.TerminalNode {
	return s.GetToken(ProcedureParserSEMICOLON, 0)
}

func (s *Stmt_variable_assignmentContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case ProcedureParserVisitor:
		return t.VisitStmt_variable_assignment(s)

	default:
		return t.VisitChildren(s)
	}
}

type Stmt_sqlContext struct {
	StatementContext
}

func NewStmt_sqlContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *Stmt_sqlContext {
	var p = new(Stmt_sqlContext)

	InitEmptyStatementContext(&p.StatementContext)
	p.parser = parser
	p.CopyAll(ctx.(*StatementContext))

	return p
}

func (s *Stmt_sqlContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Stmt_sqlContext) ANY_SQL() antlr.TerminalNode {
	return s.GetToken(ProcedureParserANY_SQL, 0)
}

func (s *Stmt_sqlContext) SEMICOLON() antlr.TerminalNode {
	return s.GetToken(ProcedureParserSEMICOLON, 0)
}

func (s *Stmt_sqlContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case ProcedureParserVisitor:
		return t.VisitStmt_sql(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *ProcedureParser) Statement() (localctx IStatementContext) {
	localctx = NewStatementContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 2, ProcedureParserRULE_statement)
	var _la int

	p.SetState(108)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 9, p.GetParserRuleContext()) {
	case 1:
		localctx = NewStmt_variable_declarationContext(p, localctx)
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(26)
			p.Match(ProcedureParserVARIABLE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(27)
			p.Type_()
		}
		{
			p.SetState(28)
			p.Match(ProcedureParserSEMICOLON)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 2:
		localctx = NewStmt_procedure_callContext(p, localctx)
		p.EnterOuterAlt(localctx, 2)
		p.SetState(40)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == ProcedureParserUNDERSCORE || _la == ProcedureParserVARIABLE {
			{
				p.SetState(30)
				p.Variable_or_underscore()
			}
			p.SetState(35)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)

			for _la == ProcedureParserCOMMA {
				{
					p.SetState(31)
					p.Match(ProcedureParserCOMMA)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				{
					p.SetState(32)
					p.Variable_or_underscore()
				}

				p.SetState(37)
				p.GetErrorHandler().Sync(p)
				if p.HasError() {
					goto errorExit
				}
				_la = p.GetTokenStream().LA(1)
			}
			{
				p.SetState(38)
				p.Match(ProcedureParserASSIGN)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

		}
		{
			p.SetState(42)
			p.Call_expression()
		}
		{
			p.SetState(43)
			p.Match(ProcedureParserSEMICOLON)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 3:
		localctx = NewStmt_variable_assignmentContext(p, localctx)
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(45)
			p.Match(ProcedureParserVARIABLE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(46)
			p.Match(ProcedureParserASSIGN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(47)
			p.expression(0)
		}
		{
			p.SetState(48)
			p.Match(ProcedureParserSEMICOLON)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 4:
		localctx = NewStmt_variable_assignment_with_declarationContext(p, localctx)
		p.EnterOuterAlt(localctx, 4)
		{
			p.SetState(50)
			p.Match(ProcedureParserVARIABLE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(51)
			p.Type_()
		}
		{
			p.SetState(52)
			p.Match(ProcedureParserASSIGN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(53)
			p.expression(0)
		}
		{
			p.SetState(54)
			p.Match(ProcedureParserSEMICOLON)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 5:
		localctx = NewStmt_for_loopContext(p, localctx)
		p.EnterOuterAlt(localctx, 5)
		{
			p.SetState(56)
			p.Match(ProcedureParserFOR)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(57)
			p.Match(ProcedureParserVARIABLE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(58)
			p.Match(ProcedureParserIN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(63)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}

		switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 3, p.GetParserRuleContext()) {
		case 1:
			{
				p.SetState(59)
				p.Range_()
			}

		case 2:
			{
				p.SetState(60)
				p.Call_expression()
			}

		case 3:
			{
				p.SetState(61)
				p.Match(ProcedureParserVARIABLE)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

		case 4:
			{
				p.SetState(62)
				p.Match(ProcedureParserANY_SQL)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

		case antlr.ATNInvalidAltNumber:
			goto errorExit
		}
		{
			p.SetState(65)
			p.Match(ProcedureParserLBRACE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(69)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		for (int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&26494311137280) != 0 {
			{
				p.SetState(66)
				p.Statement()
			}

			p.SetState(71)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)
		}
		{
			p.SetState(72)
			p.Match(ProcedureParserRBRACE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 6:
		localctx = NewStmt_ifContext(p, localctx)
		p.EnterOuterAlt(localctx, 6)
		{
			p.SetState(73)
			p.Match(ProcedureParserIF)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(74)
			p.If_then_block()
		}
		p.SetState(79)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		for _la == ProcedureParserELSEIF {
			{
				p.SetState(75)
				p.Match(ProcedureParserELSEIF)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(76)
				p.If_then_block()
			}

			p.SetState(81)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)
		}
		p.SetState(91)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == ProcedureParserELSE {
			{
				p.SetState(82)
				p.Match(ProcedureParserELSE)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(83)
				p.Match(ProcedureParserLBRACE)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			p.SetState(87)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)

			for (int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&26494311137280) != 0 {
				{
					p.SetState(84)
					p.Statement()
				}

				p.SetState(89)
				p.GetErrorHandler().Sync(p)
				if p.HasError() {
					goto errorExit
				}
				_la = p.GetTokenStream().LA(1)
			}
			{
				p.SetState(90)
				p.Match(ProcedureParserRBRACE)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

		}

	case 7:
		localctx = NewStmt_sqlContext(p, localctx)
		p.EnterOuterAlt(localctx, 7)
		{
			p.SetState(93)
			p.Match(ProcedureParserANY_SQL)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(94)
			p.Match(ProcedureParserSEMICOLON)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 8:
		localctx = NewStmt_breakContext(p, localctx)
		p.EnterOuterAlt(localctx, 8)
		{
			p.SetState(95)
			p.Match(ProcedureParserBREAK)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(96)
			p.Match(ProcedureParserSEMICOLON)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 9:
		localctx = NewStmt_returnContext(p, localctx)
		p.EnterOuterAlt(localctx, 9)
		{
			p.SetState(97)
			p.Match(ProcedureParserRETURN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(100)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}

		switch p.GetTokenStream().LA(1) {
		case ProcedureParserLPAREN, ProcedureParserLBRACKET, ProcedureParserBOOLEAN_LITERAL, ProcedureParserINT_LITERAL, ProcedureParserBLOB_LITERAL, ProcedureParserTEXT_LITERAL, ProcedureParserNULL_LITERAL, ProcedureParserIDENTIFIER, ProcedureParserVARIABLE:
			{
				p.SetState(98)
				p.Expression_list()
			}

		case ProcedureParserANY_SQL:
			{
				p.SetState(99)
				p.Match(ProcedureParserANY_SQL)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

		default:
			p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
			goto errorExit
		}
		{
			p.SetState(102)
			p.Match(ProcedureParserSEMICOLON)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 10:
		localctx = NewStmt_return_nextContext(p, localctx)
		p.EnterOuterAlt(localctx, 10)
		{
			p.SetState(103)
			p.Match(ProcedureParserRETURN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(104)
			p.Match(ProcedureParserNEXT)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(105)
			p.Expression_list()
		}
		{
			p.SetState(106)
			p.Match(ProcedureParserSEMICOLON)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
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

// IVariable_or_underscoreContext is an interface to support dynamic dispatch.
type IVariable_or_underscoreContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	VARIABLE() antlr.TerminalNode
	UNDERSCORE() antlr.TerminalNode

	// IsVariable_or_underscoreContext differentiates from other interfaces.
	IsVariable_or_underscoreContext()
}

type Variable_or_underscoreContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyVariable_or_underscoreContext() *Variable_or_underscoreContext {
	var p = new(Variable_or_underscoreContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ProcedureParserRULE_variable_or_underscore
	return p
}

func InitEmptyVariable_or_underscoreContext(p *Variable_or_underscoreContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ProcedureParserRULE_variable_or_underscore
}

func (*Variable_or_underscoreContext) IsVariable_or_underscoreContext() {}

func NewVariable_or_underscoreContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Variable_or_underscoreContext {
	var p = new(Variable_or_underscoreContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = ProcedureParserRULE_variable_or_underscore

	return p
}

func (s *Variable_or_underscoreContext) GetParser() antlr.Parser { return s.parser }

func (s *Variable_or_underscoreContext) VARIABLE() antlr.TerminalNode {
	return s.GetToken(ProcedureParserVARIABLE, 0)
}

func (s *Variable_or_underscoreContext) UNDERSCORE() antlr.TerminalNode {
	return s.GetToken(ProcedureParserUNDERSCORE, 0)
}

func (s *Variable_or_underscoreContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Variable_or_underscoreContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Variable_or_underscoreContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case ProcedureParserVisitor:
		return t.VisitVariable_or_underscore(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *ProcedureParser) Variable_or_underscore() (localctx IVariable_or_underscoreContext) {
	localctx = NewVariable_or_underscoreContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 4, ProcedureParserRULE_variable_or_underscore)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(110)
		_la = p.GetTokenStream().LA(1)

		if !(_la == ProcedureParserUNDERSCORE || _la == ProcedureParserVARIABLE) {
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

// ITypeContext is an interface to support dynamic dispatch.
type ITypeContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	IDENTIFIER() antlr.TerminalNode
	LBRACKET() antlr.TerminalNode
	RBRACKET() antlr.TerminalNode

	// IsTypeContext differentiates from other interfaces.
	IsTypeContext()
}

type TypeContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyTypeContext() *TypeContext {
	var p = new(TypeContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ProcedureParserRULE_type
	return p
}

func InitEmptyTypeContext(p *TypeContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ProcedureParserRULE_type
}

func (*TypeContext) IsTypeContext() {}

func NewTypeContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *TypeContext {
	var p = new(TypeContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = ProcedureParserRULE_type

	return p
}

func (s *TypeContext) GetParser() antlr.Parser { return s.parser }

func (s *TypeContext) IDENTIFIER() antlr.TerminalNode {
	return s.GetToken(ProcedureParserIDENTIFIER, 0)
}

func (s *TypeContext) LBRACKET() antlr.TerminalNode {
	return s.GetToken(ProcedureParserLBRACKET, 0)
}

func (s *TypeContext) RBRACKET() antlr.TerminalNode {
	return s.GetToken(ProcedureParserRBRACKET, 0)
}

func (s *TypeContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *TypeContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *TypeContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case ProcedureParserVisitor:
		return t.VisitType(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *ProcedureParser) Type_() (localctx ITypeContext) {
	localctx = NewTypeContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 6, ProcedureParserRULE_type)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(112)
		p.Match(ProcedureParserIDENTIFIER)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(115)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == ProcedureParserLBRACKET {
		{
			p.SetState(113)
			p.Match(ProcedureParserLBRACKET)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(114)
			p.Match(ProcedureParserRBRACKET)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
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

// IExpressionContext is an interface to support dynamic dispatch.
type IExpressionContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser
	// IsExpressionContext differentiates from other interfaces.
	IsExpressionContext()
}

type ExpressionContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyExpressionContext() *ExpressionContext {
	var p = new(ExpressionContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ProcedureParserRULE_expression
	return p
}

func InitEmptyExpressionContext(p *ExpressionContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ProcedureParserRULE_expression
}

func (*ExpressionContext) IsExpressionContext() {}

func NewExpressionContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ExpressionContext {
	var p = new(ExpressionContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = ProcedureParserRULE_expression

	return p
}

func (s *ExpressionContext) GetParser() antlr.Parser { return s.parser }

func (s *ExpressionContext) CopyAll(ctx *ExpressionContext) {
	s.CopyFrom(&ctx.BaseParserRuleContext)
}

func (s *ExpressionContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ExpressionContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

type Expr_array_accessContext struct {
	ExpressionContext
}

func NewExpr_array_accessContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *Expr_array_accessContext {
	var p = new(Expr_array_accessContext)

	InitEmptyExpressionContext(&p.ExpressionContext)
	p.parser = parser
	p.CopyAll(ctx.(*ExpressionContext))

	return p
}

func (s *Expr_array_accessContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Expr_array_accessContext) AllExpression() []IExpressionContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IExpressionContext); ok {
			len++
		}
	}

	tst := make([]IExpressionContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IExpressionContext); ok {
			tst[i] = t.(IExpressionContext)
			i++
		}
	}

	return tst
}

func (s *Expr_array_accessContext) Expression(i int) IExpressionContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
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

	return t.(IExpressionContext)
}

func (s *Expr_array_accessContext) LBRACKET() antlr.TerminalNode {
	return s.GetToken(ProcedureParserLBRACKET, 0)
}

func (s *Expr_array_accessContext) RBRACKET() antlr.TerminalNode {
	return s.GetToken(ProcedureParserRBRACKET, 0)
}

func (s *Expr_array_accessContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case ProcedureParserVisitor:
		return t.VisitExpr_array_access(s)

	default:
		return t.VisitChildren(s)
	}
}

type Expr_arithmeticContext struct {
	ExpressionContext
}

func NewExpr_arithmeticContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *Expr_arithmeticContext {
	var p = new(Expr_arithmeticContext)

	InitEmptyExpressionContext(&p.ExpressionContext)
	p.parser = parser
	p.CopyAll(ctx.(*ExpressionContext))

	return p
}

func (s *Expr_arithmeticContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Expr_arithmeticContext) AllExpression() []IExpressionContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IExpressionContext); ok {
			len++
		}
	}

	tst := make([]IExpressionContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IExpressionContext); ok {
			tst[i] = t.(IExpressionContext)
			i++
		}
	}

	return tst
}

func (s *Expr_arithmeticContext) Expression(i int) IExpressionContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
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

	return t.(IExpressionContext)
}

func (s *Expr_arithmeticContext) MUL() antlr.TerminalNode {
	return s.GetToken(ProcedureParserMUL, 0)
}

func (s *Expr_arithmeticContext) DIV() antlr.TerminalNode {
	return s.GetToken(ProcedureParserDIV, 0)
}

func (s *Expr_arithmeticContext) MOD() antlr.TerminalNode {
	return s.GetToken(ProcedureParserMOD, 0)
}

func (s *Expr_arithmeticContext) PLUS() antlr.TerminalNode {
	return s.GetToken(ProcedureParserPLUS, 0)
}

func (s *Expr_arithmeticContext) MINUS() antlr.TerminalNode {
	return s.GetToken(ProcedureParserMINUS, 0)
}

func (s *Expr_arithmeticContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case ProcedureParserVisitor:
		return t.VisitExpr_arithmetic(s)

	default:
		return t.VisitChildren(s)
	}
}

type Expr_variableContext struct {
	ExpressionContext
}

func NewExpr_variableContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *Expr_variableContext {
	var p = new(Expr_variableContext)

	InitEmptyExpressionContext(&p.ExpressionContext)
	p.parser = parser
	p.CopyAll(ctx.(*ExpressionContext))

	return p
}

func (s *Expr_variableContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Expr_variableContext) VARIABLE() antlr.TerminalNode {
	return s.GetToken(ProcedureParserVARIABLE, 0)
}

func (s *Expr_variableContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case ProcedureParserVisitor:
		return t.VisitExpr_variable(s)

	default:
		return t.VisitChildren(s)
	}
}

type Expr_null_literalContext struct {
	ExpressionContext
}

func NewExpr_null_literalContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *Expr_null_literalContext {
	var p = new(Expr_null_literalContext)

	InitEmptyExpressionContext(&p.ExpressionContext)
	p.parser = parser
	p.CopyAll(ctx.(*ExpressionContext))

	return p
}

func (s *Expr_null_literalContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Expr_null_literalContext) NULL_LITERAL() antlr.TerminalNode {
	return s.GetToken(ProcedureParserNULL_LITERAL, 0)
}

func (s *Expr_null_literalContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case ProcedureParserVisitor:
		return t.VisitExpr_null_literal(s)

	default:
		return t.VisitChildren(s)
	}
}

type Expr_blob_literalContext struct {
	ExpressionContext
}

func NewExpr_blob_literalContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *Expr_blob_literalContext {
	var p = new(Expr_blob_literalContext)

	InitEmptyExpressionContext(&p.ExpressionContext)
	p.parser = parser
	p.CopyAll(ctx.(*ExpressionContext))

	return p
}

func (s *Expr_blob_literalContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Expr_blob_literalContext) BLOB_LITERAL() antlr.TerminalNode {
	return s.GetToken(ProcedureParserBLOB_LITERAL, 0)
}

func (s *Expr_blob_literalContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case ProcedureParserVisitor:
		return t.VisitExpr_blob_literal(s)

	default:
		return t.VisitChildren(s)
	}
}

type Expr_comparisonContext struct {
	ExpressionContext
	left     IExpressionContext
	operator antlr.Token
	right    IExpressionContext
}

func NewExpr_comparisonContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *Expr_comparisonContext {
	var p = new(Expr_comparisonContext)

	InitEmptyExpressionContext(&p.ExpressionContext)
	p.parser = parser
	p.CopyAll(ctx.(*ExpressionContext))

	return p
}

func (s *Expr_comparisonContext) GetOperator() antlr.Token { return s.operator }

func (s *Expr_comparisonContext) SetOperator(v antlr.Token) { s.operator = v }

func (s *Expr_comparisonContext) GetLeft() IExpressionContext { return s.left }

func (s *Expr_comparisonContext) GetRight() IExpressionContext { return s.right }

func (s *Expr_comparisonContext) SetLeft(v IExpressionContext) { s.left = v }

func (s *Expr_comparisonContext) SetRight(v IExpressionContext) { s.right = v }

func (s *Expr_comparisonContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Expr_comparisonContext) AllExpression() []IExpressionContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IExpressionContext); ok {
			len++
		}
	}

	tst := make([]IExpressionContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IExpressionContext); ok {
			tst[i] = t.(IExpressionContext)
			i++
		}
	}

	return tst
}

func (s *Expr_comparisonContext) Expression(i int) IExpressionContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
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

	return t.(IExpressionContext)
}

func (s *Expr_comparisonContext) LT() antlr.TerminalNode {
	return s.GetToken(ProcedureParserLT, 0)
}

func (s *Expr_comparisonContext) LT_EQ() antlr.TerminalNode {
	return s.GetToken(ProcedureParserLT_EQ, 0)
}

func (s *Expr_comparisonContext) GT() antlr.TerminalNode {
	return s.GetToken(ProcedureParserGT, 0)
}

func (s *Expr_comparisonContext) GT_EQ() antlr.TerminalNode {
	return s.GetToken(ProcedureParserGT_EQ, 0)
}

func (s *Expr_comparisonContext) NEQ() antlr.TerminalNode {
	return s.GetToken(ProcedureParserNEQ, 0)
}

func (s *Expr_comparisonContext) EQ() antlr.TerminalNode {
	return s.GetToken(ProcedureParserEQ, 0)
}

func (s *Expr_comparisonContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case ProcedureParserVisitor:
		return t.VisitExpr_comparison(s)

	default:
		return t.VisitChildren(s)
	}
}

type Expr_boolean_literalContext struct {
	ExpressionContext
}

func NewExpr_boolean_literalContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *Expr_boolean_literalContext {
	var p = new(Expr_boolean_literalContext)

	InitEmptyExpressionContext(&p.ExpressionContext)
	p.parser = parser
	p.CopyAll(ctx.(*ExpressionContext))

	return p
}

func (s *Expr_boolean_literalContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Expr_boolean_literalContext) BOOLEAN_LITERAL() antlr.TerminalNode {
	return s.GetToken(ProcedureParserBOOLEAN_LITERAL, 0)
}

func (s *Expr_boolean_literalContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case ProcedureParserVisitor:
		return t.VisitExpr_boolean_literal(s)

	default:
		return t.VisitChildren(s)
	}
}

type Expr_callContext struct {
	ExpressionContext
}

func NewExpr_callContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *Expr_callContext {
	var p = new(Expr_callContext)

	InitEmptyExpressionContext(&p.ExpressionContext)
	p.parser = parser
	p.CopyAll(ctx.(*ExpressionContext))

	return p
}

func (s *Expr_callContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Expr_callContext) Call_expression() ICall_expressionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ICall_expressionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ICall_expressionContext)
}

func (s *Expr_callContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case ProcedureParserVisitor:
		return t.VisitExpr_call(s)

	default:
		return t.VisitChildren(s)
	}
}

type Expr_make_arrayContext struct {
	ExpressionContext
}

func NewExpr_make_arrayContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *Expr_make_arrayContext {
	var p = new(Expr_make_arrayContext)

	InitEmptyExpressionContext(&p.ExpressionContext)
	p.parser = parser
	p.CopyAll(ctx.(*ExpressionContext))

	return p
}

func (s *Expr_make_arrayContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Expr_make_arrayContext) Expression_make_array() IExpression_make_arrayContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpression_make_arrayContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpression_make_arrayContext)
}

func (s *Expr_make_arrayContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case ProcedureParserVisitor:
		return t.VisitExpr_make_array(s)

	default:
		return t.VisitChildren(s)
	}
}

type Expr_field_accessContext struct {
	ExpressionContext
}

func NewExpr_field_accessContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *Expr_field_accessContext {
	var p = new(Expr_field_accessContext)

	InitEmptyExpressionContext(&p.ExpressionContext)
	p.parser = parser
	p.CopyAll(ctx.(*ExpressionContext))

	return p
}

func (s *Expr_field_accessContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Expr_field_accessContext) Expression() IExpressionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *Expr_field_accessContext) PERIOD() antlr.TerminalNode {
	return s.GetToken(ProcedureParserPERIOD, 0)
}

func (s *Expr_field_accessContext) IDENTIFIER() antlr.TerminalNode {
	return s.GetToken(ProcedureParserIDENTIFIER, 0)
}

func (s *Expr_field_accessContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case ProcedureParserVisitor:
		return t.VisitExpr_field_access(s)

	default:
		return t.VisitChildren(s)
	}
}

type Expr_int_literalContext struct {
	ExpressionContext
}

func NewExpr_int_literalContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *Expr_int_literalContext {
	var p = new(Expr_int_literalContext)

	InitEmptyExpressionContext(&p.ExpressionContext)
	p.parser = parser
	p.CopyAll(ctx.(*ExpressionContext))

	return p
}

func (s *Expr_int_literalContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Expr_int_literalContext) INT_LITERAL() antlr.TerminalNode {
	return s.GetToken(ProcedureParserINT_LITERAL, 0)
}

func (s *Expr_int_literalContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case ProcedureParserVisitor:
		return t.VisitExpr_int_literal(s)

	default:
		return t.VisitChildren(s)
	}
}

type Expr_text_literalContext struct {
	ExpressionContext
}

func NewExpr_text_literalContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *Expr_text_literalContext {
	var p = new(Expr_text_literalContext)

	InitEmptyExpressionContext(&p.ExpressionContext)
	p.parser = parser
	p.CopyAll(ctx.(*ExpressionContext))

	return p
}

func (s *Expr_text_literalContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Expr_text_literalContext) TEXT_LITERAL() antlr.TerminalNode {
	return s.GetToken(ProcedureParserTEXT_LITERAL, 0)
}

func (s *Expr_text_literalContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case ProcedureParserVisitor:
		return t.VisitExpr_text_literal(s)

	default:
		return t.VisitChildren(s)
	}
}

type Expr_parenthesizedContext struct {
	ExpressionContext
}

func NewExpr_parenthesizedContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *Expr_parenthesizedContext {
	var p = new(Expr_parenthesizedContext)

	InitEmptyExpressionContext(&p.ExpressionContext)
	p.parser = parser
	p.CopyAll(ctx.(*ExpressionContext))

	return p
}

func (s *Expr_parenthesizedContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Expr_parenthesizedContext) LPAREN() antlr.TerminalNode {
	return s.GetToken(ProcedureParserLPAREN, 0)
}

func (s *Expr_parenthesizedContext) Expression() IExpressionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *Expr_parenthesizedContext) RPAREN() antlr.TerminalNode {
	return s.GetToken(ProcedureParserRPAREN, 0)
}

func (s *Expr_parenthesizedContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case ProcedureParserVisitor:
		return t.VisitExpr_parenthesized(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *ProcedureParser) Expression() (localctx IExpressionContext) {
	return p.expression(0)
}

func (p *ProcedureParser) expression(_p int) (localctx IExpressionContext) {
	var _parentctx antlr.ParserRuleContext = p.GetParserRuleContext()

	_parentState := p.GetState()
	localctx = NewExpressionContext(p, p.GetParserRuleContext(), _parentState)
	var _prevctx IExpressionContext = localctx
	var _ antlr.ParserRuleContext = _prevctx // TODO: To prevent unused variable warning.
	_startState := 8
	p.EnterRecursionRule(localctx, 8, ProcedureParserRULE_expression, _p)
	var _la int

	var _alt int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(130)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case ProcedureParserTEXT_LITERAL:
		localctx = NewExpr_text_literalContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx

		{
			p.SetState(118)
			p.Match(ProcedureParserTEXT_LITERAL)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case ProcedureParserBOOLEAN_LITERAL:
		localctx = NewExpr_boolean_literalContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(119)
			p.Match(ProcedureParserBOOLEAN_LITERAL)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case ProcedureParserINT_LITERAL:
		localctx = NewExpr_int_literalContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(120)
			p.Match(ProcedureParserINT_LITERAL)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case ProcedureParserNULL_LITERAL:
		localctx = NewExpr_null_literalContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(121)
			p.Match(ProcedureParserNULL_LITERAL)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case ProcedureParserBLOB_LITERAL:
		localctx = NewExpr_blob_literalContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(122)
			p.Match(ProcedureParserBLOB_LITERAL)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case ProcedureParserLBRACKET:
		localctx = NewExpr_make_arrayContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(123)
			p.Expression_make_array()
		}

	case ProcedureParserIDENTIFIER:
		localctx = NewExpr_callContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(124)
			p.Call_expression()
		}

	case ProcedureParserVARIABLE:
		localctx = NewExpr_variableContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(125)
			p.Match(ProcedureParserVARIABLE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case ProcedureParserLPAREN:
		localctx = NewExpr_parenthesizedContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(126)
			p.Match(ProcedureParserLPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(127)
			p.expression(0)
		}
		{
			p.SetState(128)
			p.Match(ProcedureParserRPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	default:
		p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
		goto errorExit
	}
	p.GetParserRuleContext().SetStop(p.GetTokenStream().LT(-1))
	p.SetState(151)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 13, p.GetParserRuleContext())
	if p.HasError() {
		goto errorExit
	}
	for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
		if _alt == 1 {
			if p.GetParseListeners() != nil {
				p.TriggerExitRuleEvent()
			}
			_prevctx = localctx
			p.SetState(149)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}

			switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 12, p.GetParserRuleContext()) {
			case 1:
				localctx = NewExpr_comparisonContext(p, NewExpressionContext(p, _parentctx, _parentState))
				localctx.(*Expr_comparisonContext).left = _prevctx

				p.PushNewRecursionContext(localctx, _startState, ProcedureParserRULE_expression)
				p.SetState(132)

				if !(p.Precpred(p.GetParserRuleContext(), 3)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 3)", ""))
					goto errorExit
				}
				{
					p.SetState(133)

					var _lt = p.GetTokenStream().LT(1)

					localctx.(*Expr_comparisonContext).operator = _lt

					_la = p.GetTokenStream().LA(1)

					if !((int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&264241152) != 0) {
						var _ri = p.GetErrorHandler().RecoverInline(p)

						localctx.(*Expr_comparisonContext).operator = _ri
					} else {
						p.GetErrorHandler().ReportMatch(p)
						p.Consume()
					}
				}
				{
					p.SetState(134)

					var _x = p.expression(4)

					localctx.(*Expr_comparisonContext).right = _x
				}

			case 2:
				localctx = NewExpr_arithmeticContext(p, NewExpressionContext(p, _parentctx, _parentState))
				p.PushNewRecursionContext(localctx, _startState, ProcedureParserRULE_expression)
				p.SetState(135)

				if !(p.Precpred(p.GetParserRuleContext(), 2)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 2)", ""))
					goto errorExit
				}
				{
					p.SetState(136)
					_la = p.GetTokenStream().LA(1)

					if !((int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&3670016) != 0) {
						p.GetErrorHandler().RecoverInline(p)
					} else {
						p.GetErrorHandler().ReportMatch(p)
						p.Consume()
					}
				}
				{
					p.SetState(137)
					p.expression(3)
				}

			case 3:
				localctx = NewExpr_arithmeticContext(p, NewExpressionContext(p, _parentctx, _parentState))
				p.PushNewRecursionContext(localctx, _startState, ProcedureParserRULE_expression)
				p.SetState(138)

				if !(p.Precpred(p.GetParserRuleContext(), 1)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 1)", ""))
					goto errorExit
				}
				{
					p.SetState(139)
					_la = p.GetTokenStream().LA(1)

					if !(_la == ProcedureParserPLUS || _la == ProcedureParserMINUS) {
						p.GetErrorHandler().RecoverInline(p)
					} else {
						p.GetErrorHandler().ReportMatch(p)
						p.Consume()
					}
				}
				{
					p.SetState(140)
					p.expression(2)
				}

			case 4:
				localctx = NewExpr_array_accessContext(p, NewExpressionContext(p, _parentctx, _parentState))
				p.PushNewRecursionContext(localctx, _startState, ProcedureParserRULE_expression)
				p.SetState(141)

				if !(p.Precpred(p.GetParserRuleContext(), 6)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 6)", ""))
					goto errorExit
				}
				{
					p.SetState(142)
					p.Match(ProcedureParserLBRACKET)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				{
					p.SetState(143)
					p.expression(0)
				}
				{
					p.SetState(144)
					p.Match(ProcedureParserRBRACKET)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}

			case 5:
				localctx = NewExpr_field_accessContext(p, NewExpressionContext(p, _parentctx, _parentState))
				p.PushNewRecursionContext(localctx, _startState, ProcedureParserRULE_expression)
				p.SetState(146)

				if !(p.Precpred(p.GetParserRuleContext(), 5)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 5)", ""))
					goto errorExit
				}
				{
					p.SetState(147)
					p.Match(ProcedureParserPERIOD)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				{
					p.SetState(148)
					p.Match(ProcedureParserIDENTIFIER)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}

			case antlr.ATNInvalidAltNumber:
				goto errorExit
			}

		}
		p.SetState(153)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 13, p.GetParserRuleContext())
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

// IExpression_listContext is an interface to support dynamic dispatch.
type IExpression_listContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllExpression() []IExpressionContext
	Expression(i int) IExpressionContext
	AllCOMMA() []antlr.TerminalNode
	COMMA(i int) antlr.TerminalNode

	// IsExpression_listContext differentiates from other interfaces.
	IsExpression_listContext()
}

type Expression_listContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyExpression_listContext() *Expression_listContext {
	var p = new(Expression_listContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ProcedureParserRULE_expression_list
	return p
}

func InitEmptyExpression_listContext(p *Expression_listContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ProcedureParserRULE_expression_list
}

func (*Expression_listContext) IsExpression_listContext() {}

func NewExpression_listContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Expression_listContext {
	var p = new(Expression_listContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = ProcedureParserRULE_expression_list

	return p
}

func (s *Expression_listContext) GetParser() antlr.Parser { return s.parser }

func (s *Expression_listContext) AllExpression() []IExpressionContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IExpressionContext); ok {
			len++
		}
	}

	tst := make([]IExpressionContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IExpressionContext); ok {
			tst[i] = t.(IExpressionContext)
			i++
		}
	}

	return tst
}

func (s *Expression_listContext) Expression(i int) IExpressionContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
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

	return t.(IExpressionContext)
}

func (s *Expression_listContext) AllCOMMA() []antlr.TerminalNode {
	return s.GetTokens(ProcedureParserCOMMA)
}

func (s *Expression_listContext) COMMA(i int) antlr.TerminalNode {
	return s.GetToken(ProcedureParserCOMMA, i)
}

func (s *Expression_listContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Expression_listContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Expression_listContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case ProcedureParserVisitor:
		return t.VisitExpression_list(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *ProcedureParser) Expression_list() (localctx IExpression_listContext) {
	localctx = NewExpression_listContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 10, ProcedureParserRULE_expression_list)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(154)
		p.expression(0)
	}
	p.SetState(159)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == ProcedureParserCOMMA {
		{
			p.SetState(155)
			p.Match(ProcedureParserCOMMA)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(156)
			p.expression(0)
		}

		p.SetState(161)
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

// IExpression_make_arrayContext is an interface to support dynamic dispatch.
type IExpression_make_arrayContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	LBRACKET() antlr.TerminalNode
	RBRACKET() antlr.TerminalNode
	Expression_list() IExpression_listContext

	// IsExpression_make_arrayContext differentiates from other interfaces.
	IsExpression_make_arrayContext()
}

type Expression_make_arrayContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyExpression_make_arrayContext() *Expression_make_arrayContext {
	var p = new(Expression_make_arrayContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ProcedureParserRULE_expression_make_array
	return p
}

func InitEmptyExpression_make_arrayContext(p *Expression_make_arrayContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ProcedureParserRULE_expression_make_array
}

func (*Expression_make_arrayContext) IsExpression_make_arrayContext() {}

func NewExpression_make_arrayContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Expression_make_arrayContext {
	var p = new(Expression_make_arrayContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = ProcedureParserRULE_expression_make_array

	return p
}

func (s *Expression_make_arrayContext) GetParser() antlr.Parser { return s.parser }

func (s *Expression_make_arrayContext) LBRACKET() antlr.TerminalNode {
	return s.GetToken(ProcedureParserLBRACKET, 0)
}

func (s *Expression_make_arrayContext) RBRACKET() antlr.TerminalNode {
	return s.GetToken(ProcedureParserRBRACKET, 0)
}

func (s *Expression_make_arrayContext) Expression_list() IExpression_listContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpression_listContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpression_listContext)
}

func (s *Expression_make_arrayContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Expression_make_arrayContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Expression_make_arrayContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case ProcedureParserVisitor:
		return t.VisitExpression_make_array(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *ProcedureParser) Expression_make_array() (localctx IExpression_make_arrayContext) {
	localctx = NewExpression_make_arrayContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 12, ProcedureParserRULE_expression_make_array)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(162)
		p.Match(ProcedureParserLBRACKET)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(164)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if (int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&34909494190084) != 0 {
		{
			p.SetState(163)
			p.Expression_list()
		}

	}
	{
		p.SetState(166)
		p.Match(ProcedureParserRBRACKET)
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

// ICall_expressionContext is an interface to support dynamic dispatch.
type ICall_expressionContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	IDENTIFIER() antlr.TerminalNode
	LPAREN() antlr.TerminalNode
	RPAREN() antlr.TerminalNode
	Expression_list() IExpression_listContext

	// IsCall_expressionContext differentiates from other interfaces.
	IsCall_expressionContext()
}

type Call_expressionContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyCall_expressionContext() *Call_expressionContext {
	var p = new(Call_expressionContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ProcedureParserRULE_call_expression
	return p
}

func InitEmptyCall_expressionContext(p *Call_expressionContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ProcedureParserRULE_call_expression
}

func (*Call_expressionContext) IsCall_expressionContext() {}

func NewCall_expressionContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Call_expressionContext {
	var p = new(Call_expressionContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = ProcedureParserRULE_call_expression

	return p
}

func (s *Call_expressionContext) GetParser() antlr.Parser { return s.parser }

func (s *Call_expressionContext) IDENTIFIER() antlr.TerminalNode {
	return s.GetToken(ProcedureParserIDENTIFIER, 0)
}

func (s *Call_expressionContext) LPAREN() antlr.TerminalNode {
	return s.GetToken(ProcedureParserLPAREN, 0)
}

func (s *Call_expressionContext) RPAREN() antlr.TerminalNode {
	return s.GetToken(ProcedureParserRPAREN, 0)
}

func (s *Call_expressionContext) Expression_list() IExpression_listContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpression_listContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpression_listContext)
}

func (s *Call_expressionContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Call_expressionContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Call_expressionContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case ProcedureParserVisitor:
		return t.VisitCall_expression(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *ProcedureParser) Call_expression() (localctx ICall_expressionContext) {
	localctx = NewCall_expressionContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 14, ProcedureParserRULE_call_expression)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(168)
		p.Match(ProcedureParserIDENTIFIER)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(169)
		p.Match(ProcedureParserLPAREN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(171)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if (int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&34909494190084) != 0 {
		{
			p.SetState(170)
			p.Expression_list()
		}

	}
	{
		p.SetState(173)
		p.Match(ProcedureParserRPAREN)
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

// IRangeContext is an interface to support dynamic dispatch.
type IRangeContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllExpression() []IExpressionContext
	Expression(i int) IExpressionContext
	COLON() antlr.TerminalNode

	// IsRangeContext differentiates from other interfaces.
	IsRangeContext()
}

type RangeContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyRangeContext() *RangeContext {
	var p = new(RangeContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ProcedureParserRULE_range
	return p
}

func InitEmptyRangeContext(p *RangeContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ProcedureParserRULE_range
}

func (*RangeContext) IsRangeContext() {}

func NewRangeContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *RangeContext {
	var p = new(RangeContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = ProcedureParserRULE_range

	return p
}

func (s *RangeContext) GetParser() antlr.Parser { return s.parser }

func (s *RangeContext) AllExpression() []IExpressionContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IExpressionContext); ok {
			len++
		}
	}

	tst := make([]IExpressionContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IExpressionContext); ok {
			tst[i] = t.(IExpressionContext)
			i++
		}
	}

	return tst
}

func (s *RangeContext) Expression(i int) IExpressionContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
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

	return t.(IExpressionContext)
}

func (s *RangeContext) COLON() antlr.TerminalNode {
	return s.GetToken(ProcedureParserCOLON, 0)
}

func (s *RangeContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *RangeContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *RangeContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case ProcedureParserVisitor:
		return t.VisitRange(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *ProcedureParser) Range_() (localctx IRangeContext) {
	localctx = NewRangeContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 16, ProcedureParserRULE_range)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(175)
		p.expression(0)
	}
	{
		p.SetState(176)
		p.Match(ProcedureParserCOLON)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(177)
		p.expression(0)
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

// IIf_then_blockContext is an interface to support dynamic dispatch.
type IIf_then_blockContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Expression() IExpressionContext
	LBRACE() antlr.TerminalNode
	RBRACE() antlr.TerminalNode
	AllStatement() []IStatementContext
	Statement(i int) IStatementContext

	// IsIf_then_blockContext differentiates from other interfaces.
	IsIf_then_blockContext()
}

type If_then_blockContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyIf_then_blockContext() *If_then_blockContext {
	var p = new(If_then_blockContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ProcedureParserRULE_if_then_block
	return p
}

func InitEmptyIf_then_blockContext(p *If_then_blockContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ProcedureParserRULE_if_then_block
}

func (*If_then_blockContext) IsIf_then_blockContext() {}

func NewIf_then_blockContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *If_then_blockContext {
	var p = new(If_then_blockContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = ProcedureParserRULE_if_then_block

	return p
}

func (s *If_then_blockContext) GetParser() antlr.Parser { return s.parser }

func (s *If_then_blockContext) Expression() IExpressionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *If_then_blockContext) LBRACE() antlr.TerminalNode {
	return s.GetToken(ProcedureParserLBRACE, 0)
}

func (s *If_then_blockContext) RBRACE() antlr.TerminalNode {
	return s.GetToken(ProcedureParserRBRACE, 0)
}

func (s *If_then_blockContext) AllStatement() []IStatementContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IStatementContext); ok {
			len++
		}
	}

	tst := make([]IStatementContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IStatementContext); ok {
			tst[i] = t.(IStatementContext)
			i++
		}
	}

	return tst
}

func (s *If_then_blockContext) Statement(i int) IStatementContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IStatementContext); ok {
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

	return t.(IStatementContext)
}

func (s *If_then_blockContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *If_then_blockContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *If_then_blockContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case ProcedureParserVisitor:
		return t.VisitIf_then_block(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *ProcedureParser) If_then_block() (localctx IIf_then_blockContext) {
	localctx = NewIf_then_blockContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 18, ProcedureParserRULE_if_then_block)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(179)
		p.expression(0)
	}
	{
		p.SetState(180)
		p.Match(ProcedureParserLBRACE)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(184)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for (int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&26494311137280) != 0 {
		{
			p.SetState(181)
			p.Statement()
		}

		p.SetState(186)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(187)
		p.Match(ProcedureParserRBRACE)
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

func (p *ProcedureParser) Sempred(localctx antlr.RuleContext, ruleIndex, predIndex int) bool {
	switch ruleIndex {
	case 4:
		var t *ExpressionContext = nil
		if localctx != nil {
			t = localctx.(*ExpressionContext)
		}
		return p.Expression_Sempred(t, predIndex)

	default:
		panic("No predicate with index: " + fmt.Sprint(ruleIndex))
	}
}

func (p *ProcedureParser) Expression_Sempred(localctx antlr.RuleContext, predIndex int) bool {
	switch predIndex {
	case 0:
		return p.Precpred(p.GetParserRuleContext(), 3)

	case 1:
		return p.Precpred(p.GetParserRuleContext(), 2)

	case 2:
		return p.Precpred(p.GetParserRuleContext(), 1)

	case 3:
		return p.Precpred(p.GetParserRuleContext(), 6)

	case 4:
		return p.Precpred(p.GetParserRuleContext(), 5)

	default:
		panic("No predicate with index: " + fmt.Sprint(predIndex))
	}
}
