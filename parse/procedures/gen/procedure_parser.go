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
		"'next'", "", "", "", "", "", "'null'",
	}
	staticData.SymbolicNames = []string{
		"", "SEMICOLON", "LPAREN", "RPAREN", "LBRACE", "RBRACE", "COMMA", "TYPE_CAST",
		"COLON", "DOLLAR", "AT", "ASSIGN", "PERIOD", "LBRACKET", "RBRACKET",
		"SINGLE_QUOTE", "UNDERSCORE", "PLUS", "MINUS", "MUL", "DIV", "MOD",
		"LT", "LT_EQ", "GT", "GT_EQ", "NEQ", "EQ", "ANY_SQL", "FOR", "IN", "IF",
		"ELSEIF", "ELSE", "TO", "RETURN", "BREAK", "NEXT", "BOOLEAN_LITERAL",
		"INT_LITERAL", "FIXED_LITERAL", "BLOB_LITERAL", "TEXT_LITERAL", "NULL_LITERAL",
		"IDENTIFIER", "VARIABLE", "WS", "TERMINATOR", "BLOCK_COMMENT", "LINE_COMMENT",
	}
	staticData.RuleNames = []string{
		"program", "statement", "variable_or_underscore", "type", "literal",
		"literal_list", "expression", "type_cast", "expression_list", "expression_make_array",
		"call_expression", "range", "if_then_block",
	}
	staticData.PredictionContextCache = antlr.NewPredictionContextCache()
	staticData.serializedATN = []int32{
		4, 1, 49, 245, 2, 0, 7, 0, 2, 1, 7, 1, 2, 2, 7, 2, 2, 3, 7, 3, 2, 4, 7,
		4, 2, 5, 7, 5, 2, 6, 7, 6, 2, 7, 7, 7, 2, 8, 7, 8, 2, 9, 7, 9, 2, 10, 7,
		10, 2, 11, 7, 11, 2, 12, 7, 12, 1, 0, 5, 0, 28, 8, 0, 10, 0, 12, 0, 31,
		9, 0, 1, 0, 1, 0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 5, 1, 42, 8,
		1, 10, 1, 12, 1, 45, 9, 1, 1, 1, 1, 1, 3, 1, 49, 8, 1, 1, 1, 1, 1, 1, 1,
		1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
		1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 3, 1, 72, 8, 1, 1, 1, 1, 1, 5, 1, 76,
		8, 1, 10, 1, 12, 1, 79, 9, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 5, 1, 86, 8,
		1, 10, 1, 12, 1, 89, 9, 1, 1, 1, 1, 1, 1, 1, 5, 1, 94, 8, 1, 10, 1, 12,
		1, 97, 9, 1, 1, 1, 3, 1, 100, 8, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
		1, 1, 3, 1, 109, 8, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 3, 1, 117, 8,
		1, 1, 2, 1, 2, 1, 3, 1, 3, 1, 3, 1, 3, 1, 3, 3, 3, 126, 8, 3, 1, 3, 1,
		3, 3, 3, 130, 8, 3, 1, 4, 1, 4, 1, 5, 1, 5, 1, 5, 5, 5, 137, 8, 5, 10,
		5, 12, 5, 140, 9, 5, 1, 6, 1, 6, 1, 6, 3, 6, 145, 8, 6, 1, 6, 1, 6, 3,
		6, 149, 8, 6, 1, 6, 1, 6, 3, 6, 153, 8, 6, 1, 6, 1, 6, 3, 6, 157, 8, 6,
		1, 6, 1, 6, 1, 6, 1, 6, 3, 6, 163, 8, 6, 3, 6, 165, 8, 6, 1, 6, 1, 6, 1,
		6, 1, 6, 1, 6, 1, 6, 1, 6, 1, 6, 1, 6, 1, 6, 1, 6, 1, 6, 1, 6, 1, 6, 3,
		6, 181, 8, 6, 1, 6, 1, 6, 1, 6, 1, 6, 3, 6, 187, 8, 6, 5, 6, 189, 8, 6,
		10, 6, 12, 6, 192, 9, 6, 1, 7, 1, 7, 1, 7, 1, 8, 1, 8, 1, 8, 5, 8, 200,
		8, 8, 10, 8, 12, 8, 203, 9, 8, 1, 9, 1, 9, 3, 9, 207, 8, 9, 1, 9, 1, 9,
		1, 10, 1, 10, 1, 10, 3, 10, 214, 8, 10, 1, 10, 1, 10, 1, 10, 1, 10, 1,
		10, 1, 10, 1, 10, 1, 10, 1, 10, 3, 10, 225, 8, 10, 1, 10, 1, 10, 3, 10,
		229, 8, 10, 1, 11, 1, 11, 1, 11, 1, 11, 1, 12, 1, 12, 1, 12, 5, 12, 238,
		8, 12, 10, 12, 12, 12, 241, 9, 12, 1, 12, 1, 12, 1, 12, 0, 1, 12, 13, 0,
		2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 0, 5, 2, 0, 16, 16, 45, 45,
		1, 0, 38, 43, 1, 0, 22, 27, 1, 0, 19, 21, 1, 0, 17, 18, 276, 0, 29, 1,
		0, 0, 0, 2, 116, 1, 0, 0, 0, 4, 118, 1, 0, 0, 0, 6, 120, 1, 0, 0, 0, 8,
		131, 1, 0, 0, 0, 10, 133, 1, 0, 0, 0, 12, 164, 1, 0, 0, 0, 14, 193, 1,
		0, 0, 0, 16, 196, 1, 0, 0, 0, 18, 204, 1, 0, 0, 0, 20, 228, 1, 0, 0, 0,
		22, 230, 1, 0, 0, 0, 24, 234, 1, 0, 0, 0, 26, 28, 3, 2, 1, 0, 27, 26, 1,
		0, 0, 0, 28, 31, 1, 0, 0, 0, 29, 27, 1, 0, 0, 0, 29, 30, 1, 0, 0, 0, 30,
		32, 1, 0, 0, 0, 31, 29, 1, 0, 0, 0, 32, 33, 5, 0, 0, 1, 33, 1, 1, 0, 0,
		0, 34, 35, 5, 45, 0, 0, 35, 36, 3, 6, 3, 0, 36, 37, 5, 1, 0, 0, 37, 117,
		1, 0, 0, 0, 38, 43, 3, 4, 2, 0, 39, 40, 5, 6, 0, 0, 40, 42, 3, 4, 2, 0,
		41, 39, 1, 0, 0, 0, 42, 45, 1, 0, 0, 0, 43, 41, 1, 0, 0, 0, 43, 44, 1,
		0, 0, 0, 44, 46, 1, 0, 0, 0, 45, 43, 1, 0, 0, 0, 46, 47, 5, 11, 0, 0, 47,
		49, 1, 0, 0, 0, 48, 38, 1, 0, 0, 0, 48, 49, 1, 0, 0, 0, 49, 50, 1, 0, 0,
		0, 50, 51, 3, 20, 10, 0, 51, 52, 5, 1, 0, 0, 52, 117, 1, 0, 0, 0, 53, 54,
		5, 45, 0, 0, 54, 55, 5, 11, 0, 0, 55, 56, 3, 12, 6, 0, 56, 57, 5, 1, 0,
		0, 57, 117, 1, 0, 0, 0, 58, 59, 5, 45, 0, 0, 59, 60, 3, 6, 3, 0, 60, 61,
		5, 11, 0, 0, 61, 62, 3, 12, 6, 0, 62, 63, 5, 1, 0, 0, 63, 117, 1, 0, 0,
		0, 64, 65, 5, 29, 0, 0, 65, 66, 5, 45, 0, 0, 66, 71, 5, 30, 0, 0, 67, 72,
		3, 22, 11, 0, 68, 72, 3, 20, 10, 0, 69, 72, 5, 45, 0, 0, 70, 72, 5, 28,
		0, 0, 71, 67, 1, 0, 0, 0, 71, 68, 1, 0, 0, 0, 71, 69, 1, 0, 0, 0, 71, 70,
		1, 0, 0, 0, 72, 73, 1, 0, 0, 0, 73, 77, 5, 4, 0, 0, 74, 76, 3, 2, 1, 0,
		75, 74, 1, 0, 0, 0, 76, 79, 1, 0, 0, 0, 77, 75, 1, 0, 0, 0, 77, 78, 1,
		0, 0, 0, 78, 80, 1, 0, 0, 0, 79, 77, 1, 0, 0, 0, 80, 117, 5, 5, 0, 0, 81,
		82, 5, 31, 0, 0, 82, 87, 3, 24, 12, 0, 83, 84, 5, 32, 0, 0, 84, 86, 3,
		24, 12, 0, 85, 83, 1, 0, 0, 0, 86, 89, 1, 0, 0, 0, 87, 85, 1, 0, 0, 0,
		87, 88, 1, 0, 0, 0, 88, 99, 1, 0, 0, 0, 89, 87, 1, 0, 0, 0, 90, 91, 5,
		33, 0, 0, 91, 95, 5, 4, 0, 0, 92, 94, 3, 2, 1, 0, 93, 92, 1, 0, 0, 0, 94,
		97, 1, 0, 0, 0, 95, 93, 1, 0, 0, 0, 95, 96, 1, 0, 0, 0, 96, 98, 1, 0, 0,
		0, 97, 95, 1, 0, 0, 0, 98, 100, 5, 5, 0, 0, 99, 90, 1, 0, 0, 0, 99, 100,
		1, 0, 0, 0, 100, 117, 1, 0, 0, 0, 101, 102, 5, 28, 0, 0, 102, 117, 5, 1,
		0, 0, 103, 104, 5, 36, 0, 0, 104, 117, 5, 1, 0, 0, 105, 108, 5, 35, 0,
		0, 106, 109, 3, 16, 8, 0, 107, 109, 5, 28, 0, 0, 108, 106, 1, 0, 0, 0,
		108, 107, 1, 0, 0, 0, 109, 110, 1, 0, 0, 0, 110, 117, 5, 1, 0, 0, 111,
		112, 5, 35, 0, 0, 112, 113, 5, 37, 0, 0, 113, 114, 3, 16, 8, 0, 114, 115,
		5, 1, 0, 0, 115, 117, 1, 0, 0, 0, 116, 34, 1, 0, 0, 0, 116, 48, 1, 0, 0,
		0, 116, 53, 1, 0, 0, 0, 116, 58, 1, 0, 0, 0, 116, 64, 1, 0, 0, 0, 116,
		81, 1, 0, 0, 0, 116, 101, 1, 0, 0, 0, 116, 103, 1, 0, 0, 0, 116, 105, 1,
		0, 0, 0, 116, 111, 1, 0, 0, 0, 117, 3, 1, 0, 0, 0, 118, 119, 7, 0, 0, 0,
		119, 5, 1, 0, 0, 0, 120, 125, 5, 44, 0, 0, 121, 122, 5, 2, 0, 0, 122, 123,
		3, 10, 5, 0, 123, 124, 5, 3, 0, 0, 124, 126, 1, 0, 0, 0, 125, 121, 1, 0,
		0, 0, 125, 126, 1, 0, 0, 0, 126, 129, 1, 0, 0, 0, 127, 128, 5, 13, 0, 0,
		128, 130, 5, 14, 0, 0, 129, 127, 1, 0, 0, 0, 129, 130, 1, 0, 0, 0, 130,
		7, 1, 0, 0, 0, 131, 132, 7, 1, 0, 0, 132, 9, 1, 0, 0, 0, 133, 138, 3, 8,
		4, 0, 134, 135, 5, 6, 0, 0, 135, 137, 3, 8, 4, 0, 136, 134, 1, 0, 0, 0,
		137, 140, 1, 0, 0, 0, 138, 136, 1, 0, 0, 0, 138, 139, 1, 0, 0, 0, 139,
		11, 1, 0, 0, 0, 140, 138, 1, 0, 0, 0, 141, 142, 6, 6, -1, 0, 142, 144,
		3, 8, 4, 0, 143, 145, 3, 14, 7, 0, 144, 143, 1, 0, 0, 0, 144, 145, 1, 0,
		0, 0, 145, 165, 1, 0, 0, 0, 146, 148, 3, 18, 9, 0, 147, 149, 3, 14, 7,
		0, 148, 147, 1, 0, 0, 0, 148, 149, 1, 0, 0, 0, 149, 165, 1, 0, 0, 0, 150,
		152, 3, 20, 10, 0, 151, 153, 3, 14, 7, 0, 152, 151, 1, 0, 0, 0, 152, 153,
		1, 0, 0, 0, 153, 165, 1, 0, 0, 0, 154, 156, 5, 45, 0, 0, 155, 157, 3, 14,
		7, 0, 156, 155, 1, 0, 0, 0, 156, 157, 1, 0, 0, 0, 157, 165, 1, 0, 0, 0,
		158, 159, 5, 2, 0, 0, 159, 160, 3, 12, 6, 0, 160, 162, 5, 3, 0, 0, 161,
		163, 3, 14, 7, 0, 162, 161, 1, 0, 0, 0, 162, 163, 1, 0, 0, 0, 163, 165,
		1, 0, 0, 0, 164, 141, 1, 0, 0, 0, 164, 146, 1, 0, 0, 0, 164, 150, 1, 0,
		0, 0, 164, 154, 1, 0, 0, 0, 164, 158, 1, 0, 0, 0, 165, 190, 1, 0, 0, 0,
		166, 167, 10, 3, 0, 0, 167, 168, 7, 2, 0, 0, 168, 189, 3, 12, 6, 4, 169,
		170, 10, 2, 0, 0, 170, 171, 7, 3, 0, 0, 171, 189, 3, 12, 6, 3, 172, 173,
		10, 1, 0, 0, 173, 174, 7, 4, 0, 0, 174, 189, 3, 12, 6, 2, 175, 176, 10,
		6, 0, 0, 176, 177, 5, 13, 0, 0, 177, 178, 3, 12, 6, 0, 178, 180, 5, 14,
		0, 0, 179, 181, 3, 14, 7, 0, 180, 179, 1, 0, 0, 0, 180, 181, 1, 0, 0, 0,
		181, 189, 1, 0, 0, 0, 182, 183, 10, 5, 0, 0, 183, 184, 5, 12, 0, 0, 184,
		186, 5, 44, 0, 0, 185, 187, 3, 14, 7, 0, 186, 185, 1, 0, 0, 0, 186, 187,
		1, 0, 0, 0, 187, 189, 1, 0, 0, 0, 188, 166, 1, 0, 0, 0, 188, 169, 1, 0,
		0, 0, 188, 172, 1, 0, 0, 0, 188, 175, 1, 0, 0, 0, 188, 182, 1, 0, 0, 0,
		189, 192, 1, 0, 0, 0, 190, 188, 1, 0, 0, 0, 190, 191, 1, 0, 0, 0, 191,
		13, 1, 0, 0, 0, 192, 190, 1, 0, 0, 0, 193, 194, 5, 7, 0, 0, 194, 195, 3,
		6, 3, 0, 195, 15, 1, 0, 0, 0, 196, 201, 3, 12, 6, 0, 197, 198, 5, 6, 0,
		0, 198, 200, 3, 12, 6, 0, 199, 197, 1, 0, 0, 0, 200, 203, 1, 0, 0, 0, 201,
		199, 1, 0, 0, 0, 201, 202, 1, 0, 0, 0, 202, 17, 1, 0, 0, 0, 203, 201, 1,
		0, 0, 0, 204, 206, 5, 13, 0, 0, 205, 207, 3, 16, 8, 0, 206, 205, 1, 0,
		0, 0, 206, 207, 1, 0, 0, 0, 207, 208, 1, 0, 0, 0, 208, 209, 5, 14, 0, 0,
		209, 19, 1, 0, 0, 0, 210, 211, 5, 44, 0, 0, 211, 213, 5, 2, 0, 0, 212,
		214, 3, 16, 8, 0, 213, 212, 1, 0, 0, 0, 213, 214, 1, 0, 0, 0, 214, 215,
		1, 0, 0, 0, 215, 229, 5, 3, 0, 0, 216, 217, 5, 44, 0, 0, 217, 218, 5, 13,
		0, 0, 218, 219, 3, 12, 6, 0, 219, 220, 5, 6, 0, 0, 220, 221, 3, 12, 6,
		0, 221, 222, 5, 14, 0, 0, 222, 224, 5, 2, 0, 0, 223, 225, 3, 16, 8, 0,
		224, 223, 1, 0, 0, 0, 224, 225, 1, 0, 0, 0, 225, 226, 1, 0, 0, 0, 226,
		227, 5, 3, 0, 0, 227, 229, 1, 0, 0, 0, 228, 210, 1, 0, 0, 0, 228, 216,
		1, 0, 0, 0, 229, 21, 1, 0, 0, 0, 230, 231, 3, 12, 6, 0, 231, 232, 5, 8,
		0, 0, 232, 233, 3, 12, 6, 0, 233, 23, 1, 0, 0, 0, 234, 235, 3, 12, 6, 0,
		235, 239, 5, 4, 0, 0, 236, 238, 3, 2, 1, 0, 237, 236, 1, 0, 0, 0, 238,
		241, 1, 0, 0, 0, 239, 237, 1, 0, 0, 0, 239, 240, 1, 0, 0, 0, 240, 242,
		1, 0, 0, 0, 241, 239, 1, 0, 0, 0, 242, 243, 5, 5, 0, 0, 243, 25, 1, 0,
		0, 0, 29, 29, 43, 48, 71, 77, 87, 95, 99, 108, 116, 125, 129, 138, 144,
		148, 152, 156, 162, 164, 180, 186, 188, 190, 201, 206, 213, 224, 228, 239,
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
	ProcedureParserFIXED_LITERAL   = 40
	ProcedureParserBLOB_LITERAL    = 41
	ProcedureParserTEXT_LITERAL    = 42
	ProcedureParserNULL_LITERAL    = 43
	ProcedureParserIDENTIFIER      = 44
	ProcedureParserVARIABLE        = 45
	ProcedureParserWS              = 46
	ProcedureParserTERMINATOR      = 47
	ProcedureParserBLOCK_COMMENT   = 48
	ProcedureParserLINE_COMMENT    = 49
)

// ProcedureParser rules.
const (
	ProcedureParserRULE_program                = 0
	ProcedureParserRULE_statement              = 1
	ProcedureParserRULE_variable_or_underscore = 2
	ProcedureParserRULE_type                   = 3
	ProcedureParserRULE_literal                = 4
	ProcedureParserRULE_literal_list           = 5
	ProcedureParserRULE_expression             = 6
	ProcedureParserRULE_type_cast              = 7
	ProcedureParserRULE_expression_list        = 8
	ProcedureParserRULE_expression_make_array  = 9
	ProcedureParserRULE_call_expression        = 10
	ProcedureParserRULE_range                  = 11
	ProcedureParserRULE_if_then_block          = 12
)

// IProgramContext is an interface to support dynamic dispatch.
type IProgramContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	EOF() antlr.TerminalNode
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

func (s *ProgramContext) EOF() antlr.TerminalNode {
	return s.GetToken(ProcedureParserEOF, 0)
}

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
	p.SetState(29)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for (int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&52882590203904) != 0 {
		{
			p.SetState(26)
			p.Statement()
		}

		p.SetState(31)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(32)
		p.Match(ProcedureParserEOF)
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

	p.SetState(116)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 9, p.GetParserRuleContext()) {
	case 1:
		localctx = NewStmt_variable_declarationContext(p, localctx)
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(34)
			p.Match(ProcedureParserVARIABLE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(35)
			p.Type_()
		}
		{
			p.SetState(36)
			p.Match(ProcedureParserSEMICOLON)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 2:
		localctx = NewStmt_procedure_callContext(p, localctx)
		p.EnterOuterAlt(localctx, 2)
		p.SetState(48)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == ProcedureParserUNDERSCORE || _la == ProcedureParserVARIABLE {
			{
				p.SetState(38)
				p.Variable_or_underscore()
			}
			p.SetState(43)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)

			for _la == ProcedureParserCOMMA {
				{
					p.SetState(39)
					p.Match(ProcedureParserCOMMA)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				{
					p.SetState(40)
					p.Variable_or_underscore()
				}

				p.SetState(45)
				p.GetErrorHandler().Sync(p)
				if p.HasError() {
					goto errorExit
				}
				_la = p.GetTokenStream().LA(1)
			}
			{
				p.SetState(46)
				p.Match(ProcedureParserASSIGN)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

		}
		{
			p.SetState(50)
			p.Call_expression()
		}
		{
			p.SetState(51)
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
			p.SetState(53)
			p.Match(ProcedureParserVARIABLE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(54)
			p.Match(ProcedureParserASSIGN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(55)
			p.expression(0)
		}
		{
			p.SetState(56)
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
			p.SetState(58)
			p.Match(ProcedureParserVARIABLE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(59)
			p.Type_()
		}
		{
			p.SetState(60)
			p.Match(ProcedureParserASSIGN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(61)
			p.expression(0)
		}
		{
			p.SetState(62)
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
			p.SetState(64)
			p.Match(ProcedureParserFOR)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(65)
			p.Match(ProcedureParserVARIABLE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(66)
			p.Match(ProcedureParserIN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(71)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}

		switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 3, p.GetParserRuleContext()) {
		case 1:
			{
				p.SetState(67)
				p.Range_()
			}

		case 2:
			{
				p.SetState(68)
				p.Call_expression()
			}

		case 3:
			{
				p.SetState(69)
				p.Match(ProcedureParserVARIABLE)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

		case 4:
			{
				p.SetState(70)
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
			p.SetState(73)
			p.Match(ProcedureParserLBRACE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(77)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		for (int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&52882590203904) != 0 {
			{
				p.SetState(74)
				p.Statement()
			}

			p.SetState(79)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)
		}
		{
			p.SetState(80)
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
			p.SetState(81)
			p.Match(ProcedureParserIF)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(82)
			p.If_then_block()
		}
		p.SetState(87)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		for _la == ProcedureParserELSEIF {
			{
				p.SetState(83)
				p.Match(ProcedureParserELSEIF)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(84)
				p.If_then_block()
			}

			p.SetState(89)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)
		}
		p.SetState(99)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == ProcedureParserELSE {
			{
				p.SetState(90)
				p.Match(ProcedureParserELSE)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(91)
				p.Match(ProcedureParserLBRACE)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			p.SetState(95)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)

			for (int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&52882590203904) != 0 {
				{
					p.SetState(92)
					p.Statement()
				}

				p.SetState(97)
				p.GetErrorHandler().Sync(p)
				if p.HasError() {
					goto errorExit
				}
				_la = p.GetTokenStream().LA(1)
			}
			{
				p.SetState(98)
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
			p.SetState(101)
			p.Match(ProcedureParserANY_SQL)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(102)
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
			p.SetState(103)
			p.Match(ProcedureParserBREAK)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(104)
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
			p.SetState(105)
			p.Match(ProcedureParserRETURN)
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
		case ProcedureParserLPAREN, ProcedureParserLBRACKET, ProcedureParserBOOLEAN_LITERAL, ProcedureParserINT_LITERAL, ProcedureParserFIXED_LITERAL, ProcedureParserBLOB_LITERAL, ProcedureParserTEXT_LITERAL, ProcedureParserNULL_LITERAL, ProcedureParserIDENTIFIER, ProcedureParserVARIABLE:
			{
				p.SetState(106)
				p.Expression_list()
			}

		case ProcedureParserANY_SQL:
			{
				p.SetState(107)
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
			p.SetState(110)
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
			p.SetState(111)
			p.Match(ProcedureParserRETURN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(112)
			p.Match(ProcedureParserNEXT)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(113)
			p.Expression_list()
		}
		{
			p.SetState(114)
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
		p.SetState(118)
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
	LPAREN() antlr.TerminalNode
	Literal_list() ILiteral_listContext
	RPAREN() antlr.TerminalNode
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

func (s *TypeContext) LPAREN() antlr.TerminalNode {
	return s.GetToken(ProcedureParserLPAREN, 0)
}

func (s *TypeContext) Literal_list() ILiteral_listContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ILiteral_listContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ILiteral_listContext)
}

func (s *TypeContext) RPAREN() antlr.TerminalNode {
	return s.GetToken(ProcedureParserRPAREN, 0)
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
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(120)
		p.Match(ProcedureParserIDENTIFIER)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(125)
	p.GetErrorHandler().Sync(p)

	if p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 10, p.GetParserRuleContext()) == 1 {
		{
			p.SetState(121)
			p.Match(ProcedureParserLPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(122)
			p.Literal_list()
		}
		{
			p.SetState(123)
			p.Match(ProcedureParserRPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	} else if p.HasError() { // JIM
		goto errorExit
	}
	p.SetState(129)
	p.GetErrorHandler().Sync(p)

	if p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 11, p.GetParserRuleContext()) == 1 {
		{
			p.SetState(127)
			p.Match(ProcedureParserLBRACKET)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(128)
			p.Match(ProcedureParserRBRACKET)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	} else if p.HasError() { // JIM
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

// ILiteralContext is an interface to support dynamic dispatch.
type ILiteralContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	TEXT_LITERAL() antlr.TerminalNode
	BOOLEAN_LITERAL() antlr.TerminalNode
	INT_LITERAL() antlr.TerminalNode
	NULL_LITERAL() antlr.TerminalNode
	BLOB_LITERAL() antlr.TerminalNode
	FIXED_LITERAL() antlr.TerminalNode

	// IsLiteralContext differentiates from other interfaces.
	IsLiteralContext()
}

type LiteralContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyLiteralContext() *LiteralContext {
	var p = new(LiteralContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ProcedureParserRULE_literal
	return p
}

func InitEmptyLiteralContext(p *LiteralContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ProcedureParserRULE_literal
}

func (*LiteralContext) IsLiteralContext() {}

func NewLiteralContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *LiteralContext {
	var p = new(LiteralContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = ProcedureParserRULE_literal

	return p
}

func (s *LiteralContext) GetParser() antlr.Parser { return s.parser }

func (s *LiteralContext) TEXT_LITERAL() antlr.TerminalNode {
	return s.GetToken(ProcedureParserTEXT_LITERAL, 0)
}

func (s *LiteralContext) BOOLEAN_LITERAL() antlr.TerminalNode {
	return s.GetToken(ProcedureParserBOOLEAN_LITERAL, 0)
}

func (s *LiteralContext) INT_LITERAL() antlr.TerminalNode {
	return s.GetToken(ProcedureParserINT_LITERAL, 0)
}

func (s *LiteralContext) NULL_LITERAL() antlr.TerminalNode {
	return s.GetToken(ProcedureParserNULL_LITERAL, 0)
}

func (s *LiteralContext) BLOB_LITERAL() antlr.TerminalNode {
	return s.GetToken(ProcedureParserBLOB_LITERAL, 0)
}

func (s *LiteralContext) FIXED_LITERAL() antlr.TerminalNode {
	return s.GetToken(ProcedureParserFIXED_LITERAL, 0)
}

func (s *LiteralContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *LiteralContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *LiteralContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case ProcedureParserVisitor:
		return t.VisitLiteral(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *ProcedureParser) Literal() (localctx ILiteralContext) {
	localctx = NewLiteralContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 8, ProcedureParserRULE_literal)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(131)
		_la = p.GetTokenStream().LA(1)

		if !((int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&17317308137472) != 0) {
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

// ILiteral_listContext is an interface to support dynamic dispatch.
type ILiteral_listContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllLiteral() []ILiteralContext
	Literal(i int) ILiteralContext
	AllCOMMA() []antlr.TerminalNode
	COMMA(i int) antlr.TerminalNode

	// IsLiteral_listContext differentiates from other interfaces.
	IsLiteral_listContext()
}

type Literal_listContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyLiteral_listContext() *Literal_listContext {
	var p = new(Literal_listContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ProcedureParserRULE_literal_list
	return p
}

func InitEmptyLiteral_listContext(p *Literal_listContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ProcedureParserRULE_literal_list
}

func (*Literal_listContext) IsLiteral_listContext() {}

func NewLiteral_listContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Literal_listContext {
	var p = new(Literal_listContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = ProcedureParserRULE_literal_list

	return p
}

func (s *Literal_listContext) GetParser() antlr.Parser { return s.parser }

func (s *Literal_listContext) AllLiteral() []ILiteralContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(ILiteralContext); ok {
			len++
		}
	}

	tst := make([]ILiteralContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(ILiteralContext); ok {
			tst[i] = t.(ILiteralContext)
			i++
		}
	}

	return tst
}

func (s *Literal_listContext) Literal(i int) ILiteralContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ILiteralContext); ok {
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

	return t.(ILiteralContext)
}

func (s *Literal_listContext) AllCOMMA() []antlr.TerminalNode {
	return s.GetTokens(ProcedureParserCOMMA)
}

func (s *Literal_listContext) COMMA(i int) antlr.TerminalNode {
	return s.GetToken(ProcedureParserCOMMA, i)
}

func (s *Literal_listContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Literal_listContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Literal_listContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case ProcedureParserVisitor:
		return t.VisitLiteral_list(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *ProcedureParser) Literal_list() (localctx ILiteral_listContext) {
	localctx = NewLiteral_listContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 10, ProcedureParserRULE_literal_list)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(133)
		p.Literal()
	}
	p.SetState(138)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == ProcedureParserCOMMA {
		{
			p.SetState(134)
			p.Match(ProcedureParserCOMMA)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(135)
			p.Literal()
		}

		p.SetState(140)
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

func (s *Expr_callContext) Type_cast() IType_castContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IType_castContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IType_castContext)
}

func (s *Expr_callContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case ProcedureParserVisitor:
		return t.VisitExpr_call(s)

	default:
		return t.VisitChildren(s)
	}
}

type Expr_literalContext struct {
	ExpressionContext
}

func NewExpr_literalContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *Expr_literalContext {
	var p = new(Expr_literalContext)

	InitEmptyExpressionContext(&p.ExpressionContext)
	p.parser = parser
	p.CopyAll(ctx.(*ExpressionContext))

	return p
}

func (s *Expr_literalContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Expr_literalContext) Literal() ILiteralContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ILiteralContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ILiteralContext)
}

func (s *Expr_literalContext) Type_cast() IType_castContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IType_castContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IType_castContext)
}

func (s *Expr_literalContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case ProcedureParserVisitor:
		return t.VisitExpr_literal(s)

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

func (s *Expr_make_arrayContext) Type_cast() IType_castContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IType_castContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IType_castContext)
}

func (s *Expr_make_arrayContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case ProcedureParserVisitor:
		return t.VisitExpr_make_array(s)

	default:
		return t.VisitChildren(s)
	}
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

func (s *Expr_array_accessContext) Type_cast() IType_castContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IType_castContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IType_castContext)
}

func (s *Expr_array_accessContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case ProcedureParserVisitor:
		return t.VisitExpr_array_access(s)

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

func (s *Expr_field_accessContext) Type_cast() IType_castContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IType_castContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IType_castContext)
}

func (s *Expr_field_accessContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case ProcedureParserVisitor:
		return t.VisitExpr_field_access(s)

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

func (s *Expr_variableContext) Type_cast() IType_castContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IType_castContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IType_castContext)
}

func (s *Expr_variableContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case ProcedureParserVisitor:
		return t.VisitExpr_variable(s)

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

func (s *Expr_parenthesizedContext) Type_cast() IType_castContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IType_castContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IType_castContext)
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
	_startState := 12
	p.EnterRecursionRule(localctx, 12, ProcedureParserRULE_expression, _p)
	var _la int

	var _alt int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(164)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case ProcedureParserBOOLEAN_LITERAL, ProcedureParserINT_LITERAL, ProcedureParserFIXED_LITERAL, ProcedureParserBLOB_LITERAL, ProcedureParserTEXT_LITERAL, ProcedureParserNULL_LITERAL:
		localctx = NewExpr_literalContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx

		{
			p.SetState(142)
			p.Literal()
		}
		p.SetState(144)
		p.GetErrorHandler().Sync(p)

		if p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 13, p.GetParserRuleContext()) == 1 {
			{
				p.SetState(143)
				p.Type_cast()
			}

		} else if p.HasError() { // JIM
			goto errorExit
		}

	case ProcedureParserLBRACKET:
		localctx = NewExpr_make_arrayContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(146)
			p.Expression_make_array()
		}
		p.SetState(148)
		p.GetErrorHandler().Sync(p)

		if p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 14, p.GetParserRuleContext()) == 1 {
			{
				p.SetState(147)
				p.Type_cast()
			}

		} else if p.HasError() { // JIM
			goto errorExit
		}

	case ProcedureParserIDENTIFIER:
		localctx = NewExpr_callContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(150)
			p.Call_expression()
		}
		p.SetState(152)
		p.GetErrorHandler().Sync(p)

		if p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 15, p.GetParserRuleContext()) == 1 {
			{
				p.SetState(151)
				p.Type_cast()
			}

		} else if p.HasError() { // JIM
			goto errorExit
		}

	case ProcedureParserVARIABLE:
		localctx = NewExpr_variableContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(154)
			p.Match(ProcedureParserVARIABLE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(156)
		p.GetErrorHandler().Sync(p)

		if p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 16, p.GetParserRuleContext()) == 1 {
			{
				p.SetState(155)
				p.Type_cast()
			}

		} else if p.HasError() { // JIM
			goto errorExit
		}

	case ProcedureParserLPAREN:
		localctx = NewExpr_parenthesizedContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(158)
			p.Match(ProcedureParserLPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(159)
			p.expression(0)
		}
		{
			p.SetState(160)
			p.Match(ProcedureParserRPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(162)
		p.GetErrorHandler().Sync(p)

		if p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 17, p.GetParserRuleContext()) == 1 {
			{
				p.SetState(161)
				p.Type_cast()
			}

		} else if p.HasError() { // JIM
			goto errorExit
		}

	default:
		p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
		goto errorExit
	}
	p.GetParserRuleContext().SetStop(p.GetTokenStream().LT(-1))
	p.SetState(190)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 22, p.GetParserRuleContext())
	if p.HasError() {
		goto errorExit
	}
	for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
		if _alt == 1 {
			if p.GetParseListeners() != nil {
				p.TriggerExitRuleEvent()
			}
			_prevctx = localctx
			p.SetState(188)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}

			switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 21, p.GetParserRuleContext()) {
			case 1:
				localctx = NewExpr_comparisonContext(p, NewExpressionContext(p, _parentctx, _parentState))
				localctx.(*Expr_comparisonContext).left = _prevctx

				p.PushNewRecursionContext(localctx, _startState, ProcedureParserRULE_expression)
				p.SetState(166)

				if !(p.Precpred(p.GetParserRuleContext(), 3)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 3)", ""))
					goto errorExit
				}
				{
					p.SetState(167)

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
					p.SetState(168)

					var _x = p.expression(4)

					localctx.(*Expr_comparisonContext).right = _x
				}

			case 2:
				localctx = NewExpr_arithmeticContext(p, NewExpressionContext(p, _parentctx, _parentState))
				p.PushNewRecursionContext(localctx, _startState, ProcedureParserRULE_expression)
				p.SetState(169)

				if !(p.Precpred(p.GetParserRuleContext(), 2)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 2)", ""))
					goto errorExit
				}
				{
					p.SetState(170)
					_la = p.GetTokenStream().LA(1)

					if !((int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&3670016) != 0) {
						p.GetErrorHandler().RecoverInline(p)
					} else {
						p.GetErrorHandler().ReportMatch(p)
						p.Consume()
					}
				}
				{
					p.SetState(171)
					p.expression(3)
				}

			case 3:
				localctx = NewExpr_arithmeticContext(p, NewExpressionContext(p, _parentctx, _parentState))
				p.PushNewRecursionContext(localctx, _startState, ProcedureParserRULE_expression)
				p.SetState(172)

				if !(p.Precpred(p.GetParserRuleContext(), 1)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 1)", ""))
					goto errorExit
				}
				{
					p.SetState(173)
					_la = p.GetTokenStream().LA(1)

					if !(_la == ProcedureParserPLUS || _la == ProcedureParserMINUS) {
						p.GetErrorHandler().RecoverInline(p)
					} else {
						p.GetErrorHandler().ReportMatch(p)
						p.Consume()
					}
				}
				{
					p.SetState(174)
					p.expression(2)
				}

			case 4:
				localctx = NewExpr_array_accessContext(p, NewExpressionContext(p, _parentctx, _parentState))
				p.PushNewRecursionContext(localctx, _startState, ProcedureParserRULE_expression)
				p.SetState(175)

				if !(p.Precpred(p.GetParserRuleContext(), 6)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 6)", ""))
					goto errorExit
				}
				{
					p.SetState(176)
					p.Match(ProcedureParserLBRACKET)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				{
					p.SetState(177)
					p.expression(0)
				}
				{
					p.SetState(178)
					p.Match(ProcedureParserRBRACKET)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				p.SetState(180)
				p.GetErrorHandler().Sync(p)

				if p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 19, p.GetParserRuleContext()) == 1 {
					{
						p.SetState(179)
						p.Type_cast()
					}

				} else if p.HasError() { // JIM
					goto errorExit
				}

			case 5:
				localctx = NewExpr_field_accessContext(p, NewExpressionContext(p, _parentctx, _parentState))
				p.PushNewRecursionContext(localctx, _startState, ProcedureParserRULE_expression)
				p.SetState(182)

				if !(p.Precpred(p.GetParserRuleContext(), 5)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 5)", ""))
					goto errorExit
				}
				{
					p.SetState(183)
					p.Match(ProcedureParserPERIOD)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				{
					p.SetState(184)
					p.Match(ProcedureParserIDENTIFIER)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				p.SetState(186)
				p.GetErrorHandler().Sync(p)

				if p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 20, p.GetParserRuleContext()) == 1 {
					{
						p.SetState(185)
						p.Type_cast()
					}

				} else if p.HasError() { // JIM
					goto errorExit
				}

			case antlr.ATNInvalidAltNumber:
				goto errorExit
			}

		}
		p.SetState(192)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 22, p.GetParserRuleContext())
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

// IType_castContext is an interface to support dynamic dispatch.
type IType_castContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	TYPE_CAST() antlr.TerminalNode
	Type_() ITypeContext

	// IsType_castContext differentiates from other interfaces.
	IsType_castContext()
}

type Type_castContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyType_castContext() *Type_castContext {
	var p = new(Type_castContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ProcedureParserRULE_type_cast
	return p
}

func InitEmptyType_castContext(p *Type_castContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ProcedureParserRULE_type_cast
}

func (*Type_castContext) IsType_castContext() {}

func NewType_castContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Type_castContext {
	var p = new(Type_castContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = ProcedureParserRULE_type_cast

	return p
}

func (s *Type_castContext) GetParser() antlr.Parser { return s.parser }

func (s *Type_castContext) TYPE_CAST() antlr.TerminalNode {
	return s.GetToken(ProcedureParserTYPE_CAST, 0)
}

func (s *Type_castContext) Type_() ITypeContext {
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

func (s *Type_castContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Type_castContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Type_castContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case ProcedureParserVisitor:
		return t.VisitType_cast(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *ProcedureParser) Type_cast() (localctx IType_castContext) {
	localctx = NewType_castContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 14, ProcedureParserRULE_type_cast)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(193)
		p.Match(ProcedureParserTYPE_CAST)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(194)
		p.Type_()
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
	p.EnterRule(localctx, 16, ProcedureParserRULE_expression_list)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(196)
		p.expression(0)
	}
	p.SetState(201)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == ProcedureParserCOMMA {
		{
			p.SetState(197)
			p.Match(ProcedureParserCOMMA)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(198)
			p.expression(0)
		}

		p.SetState(203)
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
	p.EnterRule(localctx, 18, ProcedureParserRULE_expression_make_array)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(204)
		p.Match(ProcedureParserLBRACKET)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(206)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if (int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&70093866278916) != 0 {
		{
			p.SetState(205)
			p.Expression_list()
		}

	}
	{
		p.SetState(208)
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

func (s *Call_expressionContext) CopyAll(ctx *Call_expressionContext) {
	s.CopyFrom(&ctx.BaseParserRuleContext)
}

func (s *Call_expressionContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Call_expressionContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

type Normal_callContext struct {
	Call_expressionContext
}

func NewNormal_callContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *Normal_callContext {
	var p = new(Normal_callContext)

	InitEmptyCall_expressionContext(&p.Call_expressionContext)
	p.parser = parser
	p.CopyAll(ctx.(*Call_expressionContext))

	return p
}

func (s *Normal_callContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Normal_callContext) IDENTIFIER() antlr.TerminalNode {
	return s.GetToken(ProcedureParserIDENTIFIER, 0)
}

func (s *Normal_callContext) LPAREN() antlr.TerminalNode {
	return s.GetToken(ProcedureParserLPAREN, 0)
}

func (s *Normal_callContext) RPAREN() antlr.TerminalNode {
	return s.GetToken(ProcedureParserRPAREN, 0)
}

func (s *Normal_callContext) Expression_list() IExpression_listContext {
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

func (s *Normal_callContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case ProcedureParserVisitor:
		return t.VisitNormal_call(s)

	default:
		return t.VisitChildren(s)
	}
}

type Foreign_callContext struct {
	Call_expressionContext
	dbid      IExpressionContext
	procedure IExpressionContext
}

func NewForeign_callContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *Foreign_callContext {
	var p = new(Foreign_callContext)

	InitEmptyCall_expressionContext(&p.Call_expressionContext)
	p.parser = parser
	p.CopyAll(ctx.(*Call_expressionContext))

	return p
}

func (s *Foreign_callContext) GetDbid() IExpressionContext { return s.dbid }

func (s *Foreign_callContext) GetProcedure() IExpressionContext { return s.procedure }

func (s *Foreign_callContext) SetDbid(v IExpressionContext) { s.dbid = v }

func (s *Foreign_callContext) SetProcedure(v IExpressionContext) { s.procedure = v }

func (s *Foreign_callContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Foreign_callContext) IDENTIFIER() antlr.TerminalNode {
	return s.GetToken(ProcedureParserIDENTIFIER, 0)
}

func (s *Foreign_callContext) LBRACKET() antlr.TerminalNode {
	return s.GetToken(ProcedureParserLBRACKET, 0)
}

func (s *Foreign_callContext) COMMA() antlr.TerminalNode {
	return s.GetToken(ProcedureParserCOMMA, 0)
}

func (s *Foreign_callContext) RBRACKET() antlr.TerminalNode {
	return s.GetToken(ProcedureParserRBRACKET, 0)
}

func (s *Foreign_callContext) LPAREN() antlr.TerminalNode {
	return s.GetToken(ProcedureParserLPAREN, 0)
}

func (s *Foreign_callContext) RPAREN() antlr.TerminalNode {
	return s.GetToken(ProcedureParserRPAREN, 0)
}

func (s *Foreign_callContext) AllExpression() []IExpressionContext {
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

func (s *Foreign_callContext) Expression(i int) IExpressionContext {
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

func (s *Foreign_callContext) Expression_list() IExpression_listContext {
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

func (s *Foreign_callContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case ProcedureParserVisitor:
		return t.VisitForeign_call(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *ProcedureParser) Call_expression() (localctx ICall_expressionContext) {
	localctx = NewCall_expressionContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 20, ProcedureParserRULE_call_expression)
	var _la int

	p.SetState(228)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 27, p.GetParserRuleContext()) {
	case 1:
		localctx = NewNormal_callContext(p, localctx)
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(210)
			p.Match(ProcedureParserIDENTIFIER)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(211)
			p.Match(ProcedureParserLPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(213)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if (int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&70093866278916) != 0 {
			{
				p.SetState(212)
				p.Expression_list()
			}

		}
		{
			p.SetState(215)
			p.Match(ProcedureParserRPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 2:
		localctx = NewForeign_callContext(p, localctx)
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(216)
			p.Match(ProcedureParserIDENTIFIER)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(217)
			p.Match(ProcedureParserLBRACKET)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(218)

			var _x = p.expression(0)

			localctx.(*Foreign_callContext).dbid = _x
		}
		{
			p.SetState(219)
			p.Match(ProcedureParserCOMMA)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(220)

			var _x = p.expression(0)

			localctx.(*Foreign_callContext).procedure = _x
		}
		{
			p.SetState(221)
			p.Match(ProcedureParserRBRACKET)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(222)
			p.Match(ProcedureParserLPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(224)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if (int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&70093866278916) != 0 {
			{
				p.SetState(223)
				p.Expression_list()
			}

		}
		{
			p.SetState(226)
			p.Match(ProcedureParserRPAREN)
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
	p.EnterRule(localctx, 22, ProcedureParserRULE_range)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(230)
		p.expression(0)
	}
	{
		p.SetState(231)
		p.Match(ProcedureParserCOLON)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(232)
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
	p.EnterRule(localctx, 24, ProcedureParserRULE_if_then_block)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(234)
		p.expression(0)
	}
	{
		p.SetState(235)
		p.Match(ProcedureParserLBRACE)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(239)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for (int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&52882590203904) != 0 {
		{
			p.SetState(236)
			p.Statement()
		}

		p.SetState(241)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(242)
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
	case 6:
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
