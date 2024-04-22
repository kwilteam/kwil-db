// Code generated from KuneiformParser.g4 by ANTLR 4.13.1. DO NOT EDIT.

package gen // KuneiformParser
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

type KuneiformParser struct {
	*antlr.BaseParser
}

var KuneiformParserParserStaticData struct {
	once                   sync.Once
	serializedATN          []int32
	LiteralNames           []string
	SymbolicNames          []string
	RuleNames              []string
	PredictionContextCache *antlr.PredictionContextCache
	atn                    *antlr.ATN
	decisionToDFA          []*antlr.DFA
}

func kuneiformparserParserInit() {
	staticData := &KuneiformParserParserStaticData
	staticData.LiteralNames = []string{
		"", "'{'", "'}'", "'['", "']'", "':'", "';'", "'('", "')'", "','", "'@'",
		"'.'", "'='", "'database'", "'use'", "'import'", "'as'", "'min'", "'max'",
		"'minlen'", "'maxlen'", "", "", "'default'", "'unique'", "'index'",
		"'table'", "'type'", "", "", "", "", "", "'cascade'", "", "", "'restrict'",
		"'do'", "'action'", "'procedure'", "", "", "", "", "", "", "", "", "",
		"", "", "", "", "", "", "", "", "", "'returns'",
	}
	staticData.SymbolicNames = []string{
		"", "LBRACE", "RBRACE", "LBRACKET", "RBRACKET", "COL", "SCOL", "LPAREN",
		"RPAREN", "COMMA", "AT", "PERIOD", "EQUALS", "DATABASE", "USE", "IMPORT",
		"AS", "MIN", "MAX", "MIN_LEN", "MAX_LEN", "NOT_NULL", "PRIMARY", "DEFAULT",
		"UNIQUE", "INDEX", "TABLE", "TYPE", "FOREIGN_KEY", "REFERENCES", "ON_UPDATE",
		"ON_DELETE", "DO_NO_ACTION", "DO_CASCADE", "DO_SET_NULL", "DO_SET_DEFAULT",
		"DO_RESTRICT", "DO", "START_ACTION", "START_PROCEDURE", "NUMERIC_LITERAL",
		"TEXT_LITERAL", "BOOLEAN_LITERAL", "BLOB_LITERAL", "VAR", "INDEX_NAME",
		"IDENTIFIER", "ANNOTATION", "WS", "TERMINATOR", "BLOCK_COMMENT", "LINE_COMMENT",
		"STMT_BODY", "TEXT", "STMT_LPAREN", "STMT_RPAREN", "STMT_COMMA", "STMT_PERIOD",
		"STMT_RETURNS", "STMT_TABLE", "STMT_ARRAY", "STMT_VAR", "STMT_ACCESS",
		"STMT_IDENTIFIER", "STMT_WS", "STMT_TERMINATOR", "STMT_BLOCK_COMMENT",
		"STMT_LINE_COMMENT",
	}
	staticData.RuleNames = []string{
		"program", "stmt_mode", "database_declaration", "use_declaration", "table_declaration",
		"column_def", "index_def", "foreign_key_def", "foreign_key_action",
		"identifier_list", "literal", "type_selector", "constraint", "action_declaration",
		"procedure_declaration", "stmt_return", "stmt_typed_param_list", "stmt_type_selector",
	}
	staticData.PredictionContextCache = antlr.NewPredictionContextCache()
	staticData.serializedATN = []int32{
		4, 1, 67, 244, 2, 0, 7, 0, 2, 1, 7, 1, 2, 2, 7, 2, 2, 3, 7, 3, 2, 4, 7,
		4, 2, 5, 7, 5, 2, 6, 7, 6, 2, 7, 7, 7, 2, 8, 7, 8, 2, 9, 7, 9, 2, 10, 7,
		10, 2, 11, 7, 11, 2, 12, 7, 12, 2, 13, 7, 13, 2, 14, 7, 14, 2, 15, 7, 15,
		2, 16, 7, 16, 2, 17, 7, 17, 1, 0, 1, 0, 1, 0, 1, 0, 5, 0, 41, 8, 0, 10,
		0, 12, 0, 44, 9, 0, 1, 0, 1, 0, 1, 1, 5, 1, 49, 8, 1, 10, 1, 12, 1, 52,
		9, 1, 1, 1, 1, 1, 3, 1, 56, 8, 1, 1, 2, 1, 2, 1, 2, 1, 2, 1, 3, 1, 3, 1,
		3, 1, 3, 1, 3, 1, 3, 1, 3, 1, 3, 1, 3, 1, 3, 3, 3, 72, 8, 3, 1, 3, 1, 3,
		3, 3, 76, 8, 3, 1, 3, 1, 3, 1, 3, 1, 3, 1, 4, 1, 4, 1, 4, 1, 4, 1, 4, 1,
		4, 1, 4, 1, 4, 3, 4, 90, 8, 4, 5, 4, 92, 8, 4, 10, 4, 12, 4, 95, 9, 4,
		1, 4, 1, 4, 1, 5, 1, 5, 1, 5, 5, 5, 102, 8, 5, 10, 5, 12, 5, 105, 9, 5,
		1, 6, 1, 6, 1, 6, 1, 6, 1, 6, 1, 6, 1, 7, 1, 7, 1, 7, 1, 7, 1, 7, 1, 7,
		1, 7, 1, 7, 1, 7, 1, 7, 5, 7, 123, 8, 7, 10, 7, 12, 7, 126, 9, 7, 1, 8,
		1, 8, 3, 8, 130, 8, 8, 1, 8, 1, 8, 1, 9, 1, 9, 1, 9, 5, 9, 137, 8, 9, 10,
		9, 12, 9, 140, 9, 9, 1, 10, 1, 10, 1, 11, 1, 11, 1, 11, 3, 11, 147, 8,
		11, 1, 12, 1, 12, 1, 12, 1, 12, 1, 12, 1, 12, 1, 12, 1, 12, 1, 12, 1, 12,
		1, 12, 1, 12, 1, 12, 1, 12, 1, 12, 1, 12, 1, 12, 1, 12, 1, 12, 1, 12, 1,
		12, 1, 12, 1, 12, 1, 12, 3, 12, 173, 8, 12, 1, 13, 1, 13, 1, 13, 1, 13,
		1, 13, 1, 13, 5, 13, 181, 8, 13, 10, 13, 12, 13, 184, 9, 13, 3, 13, 186,
		8, 13, 1, 13, 1, 13, 4, 13, 190, 8, 13, 11, 13, 12, 13, 191, 1, 13, 1,
		13, 1, 14, 1, 14, 1, 14, 1, 14, 3, 14, 200, 8, 14, 1, 14, 1, 14, 4, 14,
		204, 8, 14, 11, 14, 12, 14, 205, 1, 14, 1, 14, 3, 14, 210, 8, 14, 1, 14,
		1, 14, 1, 15, 3, 15, 215, 8, 15, 1, 15, 1, 15, 1, 15, 1, 15, 1, 15, 1,
		15, 5, 15, 223, 8, 15, 10, 15, 12, 15, 226, 9, 15, 1, 15, 1, 15, 1, 16,
		1, 16, 1, 16, 1, 16, 1, 16, 5, 16, 235, 8, 16, 10, 16, 12, 16, 238, 9,
		16, 1, 17, 1, 17, 3, 17, 242, 8, 17, 1, 17, 0, 0, 18, 0, 2, 4, 6, 8, 10,
		12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32, 34, 0, 4, 2, 0, 22, 22, 24,
		25, 1, 0, 30, 31, 1, 0, 32, 36, 1, 0, 40, 43, 257, 0, 36, 1, 0, 0, 0, 2,
		50, 1, 0, 0, 0, 4, 57, 1, 0, 0, 0, 6, 61, 1, 0, 0, 0, 8, 81, 1, 0, 0, 0,
		10, 98, 1, 0, 0, 0, 12, 106, 1, 0, 0, 0, 14, 112, 1, 0, 0, 0, 16, 127,
		1, 0, 0, 0, 18, 133, 1, 0, 0, 0, 20, 141, 1, 0, 0, 0, 22, 143, 1, 0, 0,
		0, 24, 172, 1, 0, 0, 0, 26, 174, 1, 0, 0, 0, 28, 195, 1, 0, 0, 0, 30, 214,
		1, 0, 0, 0, 32, 229, 1, 0, 0, 0, 34, 239, 1, 0, 0, 0, 36, 42, 3, 4, 2,
		0, 37, 41, 3, 6, 3, 0, 38, 41, 3, 8, 4, 0, 39, 41, 3, 2, 1, 0, 40, 37,
		1, 0, 0, 0, 40, 38, 1, 0, 0, 0, 40, 39, 1, 0, 0, 0, 41, 44, 1, 0, 0, 0,
		42, 40, 1, 0, 0, 0, 42, 43, 1, 0, 0, 0, 43, 45, 1, 0, 0, 0, 44, 42, 1,
		0, 0, 0, 45, 46, 5, 0, 0, 1, 46, 1, 1, 0, 0, 0, 47, 49, 5, 47, 0, 0, 48,
		47, 1, 0, 0, 0, 49, 52, 1, 0, 0, 0, 50, 48, 1, 0, 0, 0, 50, 51, 1, 0, 0,
		0, 51, 55, 1, 0, 0, 0, 52, 50, 1, 0, 0, 0, 53, 56, 3, 26, 13, 0, 54, 56,
		3, 28, 14, 0, 55, 53, 1, 0, 0, 0, 55, 54, 1, 0, 0, 0, 56, 3, 1, 0, 0, 0,
		57, 58, 5, 13, 0, 0, 58, 59, 5, 46, 0, 0, 59, 60, 5, 6, 0, 0, 60, 5, 1,
		0, 0, 0, 61, 62, 5, 14, 0, 0, 62, 75, 5, 46, 0, 0, 63, 64, 5, 1, 0, 0,
		64, 65, 5, 46, 0, 0, 65, 66, 5, 5, 0, 0, 66, 71, 3, 20, 10, 0, 67, 68,
		5, 9, 0, 0, 68, 69, 5, 46, 0, 0, 69, 70, 5, 5, 0, 0, 70, 72, 3, 20, 10,
		0, 71, 67, 1, 0, 0, 0, 71, 72, 1, 0, 0, 0, 72, 73, 1, 0, 0, 0, 73, 74,
		5, 2, 0, 0, 74, 76, 1, 0, 0, 0, 75, 63, 1, 0, 0, 0, 75, 76, 1, 0, 0, 0,
		76, 77, 1, 0, 0, 0, 77, 78, 5, 16, 0, 0, 78, 79, 5, 46, 0, 0, 79, 80, 5,
		6, 0, 0, 80, 7, 1, 0, 0, 0, 81, 82, 5, 26, 0, 0, 82, 83, 5, 46, 0, 0, 83,
		84, 5, 1, 0, 0, 84, 93, 3, 10, 5, 0, 85, 89, 5, 9, 0, 0, 86, 90, 3, 10,
		5, 0, 87, 90, 3, 12, 6, 0, 88, 90, 3, 14, 7, 0, 89, 86, 1, 0, 0, 0, 89,
		87, 1, 0, 0, 0, 89, 88, 1, 0, 0, 0, 90, 92, 1, 0, 0, 0, 91, 85, 1, 0, 0,
		0, 92, 95, 1, 0, 0, 0, 93, 91, 1, 0, 0, 0, 93, 94, 1, 0, 0, 0, 94, 96,
		1, 0, 0, 0, 95, 93, 1, 0, 0, 0, 96, 97, 5, 2, 0, 0, 97, 9, 1, 0, 0, 0,
		98, 99, 5, 46, 0, 0, 99, 103, 3, 22, 11, 0, 100, 102, 3, 24, 12, 0, 101,
		100, 1, 0, 0, 0, 102, 105, 1, 0, 0, 0, 103, 101, 1, 0, 0, 0, 103, 104,
		1, 0, 0, 0, 104, 11, 1, 0, 0, 0, 105, 103, 1, 0, 0, 0, 106, 107, 5, 45,
		0, 0, 107, 108, 7, 0, 0, 0, 108, 109, 5, 7, 0, 0, 109, 110, 3, 18, 9, 0,
		110, 111, 5, 8, 0, 0, 111, 13, 1, 0, 0, 0, 112, 113, 5, 28, 0, 0, 113,
		114, 5, 7, 0, 0, 114, 115, 3, 18, 9, 0, 115, 116, 5, 8, 0, 0, 116, 117,
		5, 29, 0, 0, 117, 118, 5, 46, 0, 0, 118, 119, 5, 7, 0, 0, 119, 120, 3,
		18, 9, 0, 120, 124, 5, 8, 0, 0, 121, 123, 3, 16, 8, 0, 122, 121, 1, 0,
		0, 0, 123, 126, 1, 0, 0, 0, 124, 122, 1, 0, 0, 0, 124, 125, 1, 0, 0, 0,
		125, 15, 1, 0, 0, 0, 126, 124, 1, 0, 0, 0, 127, 129, 7, 1, 0, 0, 128, 130,
		5, 37, 0, 0, 129, 128, 1, 0, 0, 0, 129, 130, 1, 0, 0, 0, 130, 131, 1, 0,
		0, 0, 131, 132, 7, 2, 0, 0, 132, 17, 1, 0, 0, 0, 133, 138, 5, 46, 0, 0,
		134, 135, 5, 9, 0, 0, 135, 137, 5, 46, 0, 0, 136, 134, 1, 0, 0, 0, 137,
		140, 1, 0, 0, 0, 138, 136, 1, 0, 0, 0, 138, 139, 1, 0, 0, 0, 139, 19, 1,
		0, 0, 0, 140, 138, 1, 0, 0, 0, 141, 142, 7, 3, 0, 0, 142, 21, 1, 0, 0,
		0, 143, 146, 5, 46, 0, 0, 144, 145, 5, 3, 0, 0, 145, 147, 5, 4, 0, 0, 146,
		144, 1, 0, 0, 0, 146, 147, 1, 0, 0, 0, 147, 23, 1, 0, 0, 0, 148, 149, 5,
		17, 0, 0, 149, 150, 5, 7, 0, 0, 150, 151, 5, 40, 0, 0, 151, 173, 5, 8,
		0, 0, 152, 153, 5, 18, 0, 0, 153, 154, 5, 7, 0, 0, 154, 155, 5, 40, 0,
		0, 155, 173, 5, 8, 0, 0, 156, 157, 5, 19, 0, 0, 157, 158, 5, 7, 0, 0, 158,
		159, 5, 40, 0, 0, 159, 173, 5, 8, 0, 0, 160, 161, 5, 20, 0, 0, 161, 162,
		5, 7, 0, 0, 162, 163, 5, 40, 0, 0, 163, 173, 5, 8, 0, 0, 164, 173, 5, 21,
		0, 0, 165, 173, 5, 22, 0, 0, 166, 167, 5, 23, 0, 0, 167, 168, 5, 7, 0,
		0, 168, 169, 3, 20, 10, 0, 169, 170, 5, 8, 0, 0, 170, 173, 1, 0, 0, 0,
		171, 173, 5, 24, 0, 0, 172, 148, 1, 0, 0, 0, 172, 152, 1, 0, 0, 0, 172,
		156, 1, 0, 0, 0, 172, 160, 1, 0, 0, 0, 172, 164, 1, 0, 0, 0, 172, 165,
		1, 0, 0, 0, 172, 166, 1, 0, 0, 0, 172, 171, 1, 0, 0, 0, 173, 25, 1, 0,
		0, 0, 174, 175, 5, 38, 0, 0, 175, 176, 5, 63, 0, 0, 176, 185, 5, 54, 0,
		0, 177, 182, 5, 61, 0, 0, 178, 179, 5, 56, 0, 0, 179, 181, 5, 61, 0, 0,
		180, 178, 1, 0, 0, 0, 181, 184, 1, 0, 0, 0, 182, 180, 1, 0, 0, 0, 182,
		183, 1, 0, 0, 0, 183, 186, 1, 0, 0, 0, 184, 182, 1, 0, 0, 0, 185, 177,
		1, 0, 0, 0, 185, 186, 1, 0, 0, 0, 186, 187, 1, 0, 0, 0, 187, 189, 5, 55,
		0, 0, 188, 190, 5, 62, 0, 0, 189, 188, 1, 0, 0, 0, 190, 191, 1, 0, 0, 0,
		191, 189, 1, 0, 0, 0, 191, 192, 1, 0, 0, 0, 192, 193, 1, 0, 0, 0, 193,
		194, 5, 52, 0, 0, 194, 27, 1, 0, 0, 0, 195, 196, 5, 39, 0, 0, 196, 197,
		5, 63, 0, 0, 197, 199, 5, 54, 0, 0, 198, 200, 3, 32, 16, 0, 199, 198, 1,
		0, 0, 0, 199, 200, 1, 0, 0, 0, 200, 201, 1, 0, 0, 0, 201, 203, 5, 55, 0,
		0, 202, 204, 5, 62, 0, 0, 203, 202, 1, 0, 0, 0, 204, 205, 1, 0, 0, 0, 205,
		203, 1, 0, 0, 0, 205, 206, 1, 0, 0, 0, 206, 209, 1, 0, 0, 0, 207, 208,
		5, 58, 0, 0, 208, 210, 3, 30, 15, 0, 209, 207, 1, 0, 0, 0, 209, 210, 1,
		0, 0, 0, 210, 211, 1, 0, 0, 0, 211, 212, 5, 52, 0, 0, 212, 29, 1, 0, 0,
		0, 213, 215, 5, 59, 0, 0, 214, 213, 1, 0, 0, 0, 214, 215, 1, 0, 0, 0, 215,
		216, 1, 0, 0, 0, 216, 217, 5, 54, 0, 0, 217, 218, 5, 63, 0, 0, 218, 224,
		3, 34, 17, 0, 219, 220, 5, 56, 0, 0, 220, 221, 5, 63, 0, 0, 221, 223, 3,
		34, 17, 0, 222, 219, 1, 0, 0, 0, 223, 226, 1, 0, 0, 0, 224, 222, 1, 0,
		0, 0, 224, 225, 1, 0, 0, 0, 225, 227, 1, 0, 0, 0, 226, 224, 1, 0, 0, 0,
		227, 228, 5, 55, 0, 0, 228, 31, 1, 0, 0, 0, 229, 230, 5, 61, 0, 0, 230,
		236, 3, 34, 17, 0, 231, 232, 5, 56, 0, 0, 232, 233, 5, 61, 0, 0, 233, 235,
		3, 34, 17, 0, 234, 231, 1, 0, 0, 0, 235, 238, 1, 0, 0, 0, 236, 234, 1,
		0, 0, 0, 236, 237, 1, 0, 0, 0, 237, 33, 1, 0, 0, 0, 238, 236, 1, 0, 0,
		0, 239, 241, 5, 63, 0, 0, 240, 242, 5, 60, 0, 0, 241, 240, 1, 0, 0, 0,
		241, 242, 1, 0, 0, 0, 242, 35, 1, 0, 0, 0, 24, 40, 42, 50, 55, 71, 75,
		89, 93, 103, 124, 129, 138, 146, 172, 182, 185, 191, 199, 205, 209, 214,
		224, 236, 241,
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

// KuneiformParserInit initializes any static state used to implement KuneiformParser. By default the
// static state used to implement the parser is lazily initialized during the first call to
// NewKuneiformParser(). You can call this function if you wish to initialize the static state ahead
// of time.
func KuneiformParserInit() {
	staticData := &KuneiformParserParserStaticData
	staticData.once.Do(kuneiformparserParserInit)
}

// NewKuneiformParser produces a new parser instance for the optional input antlr.TokenStream.
func NewKuneiformParser(input antlr.TokenStream) *KuneiformParser {
	KuneiformParserInit()
	this := new(KuneiformParser)
	this.BaseParser = antlr.NewBaseParser(input)
	staticData := &KuneiformParserParserStaticData
	this.Interpreter = antlr.NewParserATNSimulator(this, staticData.atn, staticData.decisionToDFA, staticData.PredictionContextCache)
	this.RuleNames = staticData.RuleNames
	this.LiteralNames = staticData.LiteralNames
	this.SymbolicNames = staticData.SymbolicNames
	this.GrammarFileName = "KuneiformParser.g4"

	return this
}

// KuneiformParser tokens.
const (
	KuneiformParserEOF                = antlr.TokenEOF
	KuneiformParserLBRACE             = 1
	KuneiformParserRBRACE             = 2
	KuneiformParserLBRACKET           = 3
	KuneiformParserRBRACKET           = 4
	KuneiformParserCOL                = 5
	KuneiformParserSCOL               = 6
	KuneiformParserLPAREN             = 7
	KuneiformParserRPAREN             = 8
	KuneiformParserCOMMA              = 9
	KuneiformParserAT                 = 10
	KuneiformParserPERIOD             = 11
	KuneiformParserEQUALS             = 12
	KuneiformParserDATABASE           = 13
	KuneiformParserUSE                = 14
	KuneiformParserIMPORT             = 15
	KuneiformParserAS                 = 16
	KuneiformParserMIN                = 17
	KuneiformParserMAX                = 18
	KuneiformParserMIN_LEN            = 19
	KuneiformParserMAX_LEN            = 20
	KuneiformParserNOT_NULL           = 21
	KuneiformParserPRIMARY            = 22
	KuneiformParserDEFAULT            = 23
	KuneiformParserUNIQUE             = 24
	KuneiformParserINDEX              = 25
	KuneiformParserTABLE              = 26
	KuneiformParserTYPE               = 27
	KuneiformParserFOREIGN_KEY        = 28
	KuneiformParserREFERENCES         = 29
	KuneiformParserON_UPDATE          = 30
	KuneiformParserON_DELETE          = 31
	KuneiformParserDO_NO_ACTION       = 32
	KuneiformParserDO_CASCADE         = 33
	KuneiformParserDO_SET_NULL        = 34
	KuneiformParserDO_SET_DEFAULT     = 35
	KuneiformParserDO_RESTRICT        = 36
	KuneiformParserDO                 = 37
	KuneiformParserSTART_ACTION       = 38
	KuneiformParserSTART_PROCEDURE    = 39
	KuneiformParserNUMERIC_LITERAL    = 40
	KuneiformParserTEXT_LITERAL       = 41
	KuneiformParserBOOLEAN_LITERAL    = 42
	KuneiformParserBLOB_LITERAL       = 43
	KuneiformParserVAR                = 44
	KuneiformParserINDEX_NAME         = 45
	KuneiformParserIDENTIFIER         = 46
	KuneiformParserANNOTATION         = 47
	KuneiformParserWS                 = 48
	KuneiformParserTERMINATOR         = 49
	KuneiformParserBLOCK_COMMENT      = 50
	KuneiformParserLINE_COMMENT       = 51
	KuneiformParserSTMT_BODY          = 52
	KuneiformParserTEXT               = 53
	KuneiformParserSTMT_LPAREN        = 54
	KuneiformParserSTMT_RPAREN        = 55
	KuneiformParserSTMT_COMMA         = 56
	KuneiformParserSTMT_PERIOD        = 57
	KuneiformParserSTMT_RETURNS       = 58
	KuneiformParserSTMT_TABLE         = 59
	KuneiformParserSTMT_ARRAY         = 60
	KuneiformParserSTMT_VAR           = 61
	KuneiformParserSTMT_ACCESS        = 62
	KuneiformParserSTMT_IDENTIFIER    = 63
	KuneiformParserSTMT_WS            = 64
	KuneiformParserSTMT_TERMINATOR    = 65
	KuneiformParserSTMT_BLOCK_COMMENT = 66
	KuneiformParserSTMT_LINE_COMMENT  = 67
)

// KuneiformParser rules.
const (
	KuneiformParserRULE_program               = 0
	KuneiformParserRULE_stmt_mode             = 1
	KuneiformParserRULE_database_declaration  = 2
	KuneiformParserRULE_use_declaration       = 3
	KuneiformParserRULE_table_declaration     = 4
	KuneiformParserRULE_column_def            = 5
	KuneiformParserRULE_index_def             = 6
	KuneiformParserRULE_foreign_key_def       = 7
	KuneiformParserRULE_foreign_key_action    = 8
	KuneiformParserRULE_identifier_list       = 9
	KuneiformParserRULE_literal               = 10
	KuneiformParserRULE_type_selector         = 11
	KuneiformParserRULE_constraint            = 12
	KuneiformParserRULE_action_declaration    = 13
	KuneiformParserRULE_procedure_declaration = 14
	KuneiformParserRULE_stmt_return           = 15
	KuneiformParserRULE_stmt_typed_param_list = 16
	KuneiformParserRULE_stmt_type_selector    = 17
)

// IProgramContext is an interface to support dynamic dispatch.
type IProgramContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Database_declaration() IDatabase_declarationContext
	EOF() antlr.TerminalNode
	AllUse_declaration() []IUse_declarationContext
	Use_declaration(i int) IUse_declarationContext
	AllTable_declaration() []ITable_declarationContext
	Table_declaration(i int) ITable_declarationContext
	AllStmt_mode() []IStmt_modeContext
	Stmt_mode(i int) IStmt_modeContext

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
	p.RuleIndex = KuneiformParserRULE_program
	return p
}

func InitEmptyProgramContext(p *ProgramContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = KuneiformParserRULE_program
}

func (*ProgramContext) IsProgramContext() {}

func NewProgramContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ProgramContext {
	var p = new(ProgramContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = KuneiformParserRULE_program

	return p
}

func (s *ProgramContext) GetParser() antlr.Parser { return s.parser }

func (s *ProgramContext) Database_declaration() IDatabase_declarationContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IDatabase_declarationContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IDatabase_declarationContext)
}

func (s *ProgramContext) EOF() antlr.TerminalNode {
	return s.GetToken(KuneiformParserEOF, 0)
}

func (s *ProgramContext) AllUse_declaration() []IUse_declarationContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IUse_declarationContext); ok {
			len++
		}
	}

	tst := make([]IUse_declarationContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IUse_declarationContext); ok {
			tst[i] = t.(IUse_declarationContext)
			i++
		}
	}

	return tst
}

func (s *ProgramContext) Use_declaration(i int) IUse_declarationContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IUse_declarationContext); ok {
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

	return t.(IUse_declarationContext)
}

func (s *ProgramContext) AllTable_declaration() []ITable_declarationContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(ITable_declarationContext); ok {
			len++
		}
	}

	tst := make([]ITable_declarationContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(ITable_declarationContext); ok {
			tst[i] = t.(ITable_declarationContext)
			i++
		}
	}

	return tst
}

func (s *ProgramContext) Table_declaration(i int) ITable_declarationContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITable_declarationContext); ok {
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

	return t.(ITable_declarationContext)
}

func (s *ProgramContext) AllStmt_mode() []IStmt_modeContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IStmt_modeContext); ok {
			len++
		}
	}

	tst := make([]IStmt_modeContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IStmt_modeContext); ok {
			tst[i] = t.(IStmt_modeContext)
			i++
		}
	}

	return tst
}

func (s *ProgramContext) Stmt_mode(i int) IStmt_modeContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IStmt_modeContext); ok {
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

	return t.(IStmt_modeContext)
}

func (s *ProgramContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ProgramContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ProgramContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case KuneiformParserVisitor:
		return t.VisitProgram(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *KuneiformParser) Program() (localctx IProgramContext) {
	localctx = NewProgramContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 0, KuneiformParserRULE_program)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(36)
		p.Database_declaration()
	}
	p.SetState(42)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for (int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&141562189201408) != 0 {
		p.SetState(40)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}

		switch p.GetTokenStream().LA(1) {
		case KuneiformParserUSE:
			{
				p.SetState(37)
				p.Use_declaration()
			}

		case KuneiformParserTABLE:
			{
				p.SetState(38)
				p.Table_declaration()
			}

		case KuneiformParserSTART_ACTION, KuneiformParserSTART_PROCEDURE, KuneiformParserANNOTATION:
			{
				p.SetState(39)
				p.Stmt_mode()
			}

		default:
			p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
			goto errorExit
		}

		p.SetState(44)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(45)
		p.Match(KuneiformParserEOF)
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

// IStmt_modeContext is an interface to support dynamic dispatch.
type IStmt_modeContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Action_declaration() IAction_declarationContext
	Procedure_declaration() IProcedure_declarationContext
	AllANNOTATION() []antlr.TerminalNode
	ANNOTATION(i int) antlr.TerminalNode

	// IsStmt_modeContext differentiates from other interfaces.
	IsStmt_modeContext()
}

type Stmt_modeContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyStmt_modeContext() *Stmt_modeContext {
	var p = new(Stmt_modeContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = KuneiformParserRULE_stmt_mode
	return p
}

func InitEmptyStmt_modeContext(p *Stmt_modeContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = KuneiformParserRULE_stmt_mode
}

func (*Stmt_modeContext) IsStmt_modeContext() {}

func NewStmt_modeContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Stmt_modeContext {
	var p = new(Stmt_modeContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = KuneiformParserRULE_stmt_mode

	return p
}

func (s *Stmt_modeContext) GetParser() antlr.Parser { return s.parser }

func (s *Stmt_modeContext) Action_declaration() IAction_declarationContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IAction_declarationContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IAction_declarationContext)
}

func (s *Stmt_modeContext) Procedure_declaration() IProcedure_declarationContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IProcedure_declarationContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IProcedure_declarationContext)
}

func (s *Stmt_modeContext) AllANNOTATION() []antlr.TerminalNode {
	return s.GetTokens(KuneiformParserANNOTATION)
}

func (s *Stmt_modeContext) ANNOTATION(i int) antlr.TerminalNode {
	return s.GetToken(KuneiformParserANNOTATION, i)
}

func (s *Stmt_modeContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Stmt_modeContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Stmt_modeContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case KuneiformParserVisitor:
		return t.VisitStmt_mode(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *KuneiformParser) Stmt_mode() (localctx IStmt_modeContext) {
	localctx = NewStmt_modeContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 2, KuneiformParserRULE_stmt_mode)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(50)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == KuneiformParserANNOTATION {
		{
			p.SetState(47)
			p.Match(KuneiformParserANNOTATION)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

		p.SetState(52)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	p.SetState(55)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case KuneiformParserSTART_ACTION:
		{
			p.SetState(53)
			p.Action_declaration()
		}

	case KuneiformParserSTART_PROCEDURE:
		{
			p.SetState(54)
			p.Procedure_declaration()
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

// IDatabase_declarationContext is an interface to support dynamic dispatch.
type IDatabase_declarationContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	DATABASE() antlr.TerminalNode
	IDENTIFIER() antlr.TerminalNode
	SCOL() antlr.TerminalNode

	// IsDatabase_declarationContext differentiates from other interfaces.
	IsDatabase_declarationContext()
}

type Database_declarationContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyDatabase_declarationContext() *Database_declarationContext {
	var p = new(Database_declarationContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = KuneiformParserRULE_database_declaration
	return p
}

func InitEmptyDatabase_declarationContext(p *Database_declarationContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = KuneiformParserRULE_database_declaration
}

func (*Database_declarationContext) IsDatabase_declarationContext() {}

func NewDatabase_declarationContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Database_declarationContext {
	var p = new(Database_declarationContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = KuneiformParserRULE_database_declaration

	return p
}

func (s *Database_declarationContext) GetParser() antlr.Parser { return s.parser }

func (s *Database_declarationContext) DATABASE() antlr.TerminalNode {
	return s.GetToken(KuneiformParserDATABASE, 0)
}

func (s *Database_declarationContext) IDENTIFIER() antlr.TerminalNode {
	return s.GetToken(KuneiformParserIDENTIFIER, 0)
}

func (s *Database_declarationContext) SCOL() antlr.TerminalNode {
	return s.GetToken(KuneiformParserSCOL, 0)
}

func (s *Database_declarationContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Database_declarationContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Database_declarationContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case KuneiformParserVisitor:
		return t.VisitDatabase_declaration(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *KuneiformParser) Database_declaration() (localctx IDatabase_declarationContext) {
	localctx = NewDatabase_declarationContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 4, KuneiformParserRULE_database_declaration)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(57)
		p.Match(KuneiformParserDATABASE)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(58)
		p.Match(KuneiformParserIDENTIFIER)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(59)
		p.Match(KuneiformParserSCOL)
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

// IUse_declarationContext is an interface to support dynamic dispatch.
type IUse_declarationContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// GetExtension_name returns the extension_name token.
	GetExtension_name() antlr.Token

	// GetAlias returns the alias token.
	GetAlias() antlr.Token

	// SetExtension_name sets the extension_name token.
	SetExtension_name(antlr.Token)

	// SetAlias sets the alias token.
	SetAlias(antlr.Token)

	// Getter signatures
	USE() antlr.TerminalNode
	AS() antlr.TerminalNode
	SCOL() antlr.TerminalNode
	AllIDENTIFIER() []antlr.TerminalNode
	IDENTIFIER(i int) antlr.TerminalNode
	LBRACE() antlr.TerminalNode
	AllCOL() []antlr.TerminalNode
	COL(i int) antlr.TerminalNode
	AllLiteral() []ILiteralContext
	Literal(i int) ILiteralContext
	RBRACE() antlr.TerminalNode
	COMMA() antlr.TerminalNode

	// IsUse_declarationContext differentiates from other interfaces.
	IsUse_declarationContext()
}

type Use_declarationContext struct {
	antlr.BaseParserRuleContext
	parser         antlr.Parser
	extension_name antlr.Token
	alias          antlr.Token
}

func NewEmptyUse_declarationContext() *Use_declarationContext {
	var p = new(Use_declarationContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = KuneiformParserRULE_use_declaration
	return p
}

func InitEmptyUse_declarationContext(p *Use_declarationContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = KuneiformParserRULE_use_declaration
}

func (*Use_declarationContext) IsUse_declarationContext() {}

func NewUse_declarationContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Use_declarationContext {
	var p = new(Use_declarationContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = KuneiformParserRULE_use_declaration

	return p
}

func (s *Use_declarationContext) GetParser() antlr.Parser { return s.parser }

func (s *Use_declarationContext) GetExtension_name() antlr.Token { return s.extension_name }

func (s *Use_declarationContext) GetAlias() antlr.Token { return s.alias }

func (s *Use_declarationContext) SetExtension_name(v antlr.Token) { s.extension_name = v }

func (s *Use_declarationContext) SetAlias(v antlr.Token) { s.alias = v }

func (s *Use_declarationContext) USE() antlr.TerminalNode {
	return s.GetToken(KuneiformParserUSE, 0)
}

func (s *Use_declarationContext) AS() antlr.TerminalNode {
	return s.GetToken(KuneiformParserAS, 0)
}

func (s *Use_declarationContext) SCOL() antlr.TerminalNode {
	return s.GetToken(KuneiformParserSCOL, 0)
}

func (s *Use_declarationContext) AllIDENTIFIER() []antlr.TerminalNode {
	return s.GetTokens(KuneiformParserIDENTIFIER)
}

func (s *Use_declarationContext) IDENTIFIER(i int) antlr.TerminalNode {
	return s.GetToken(KuneiformParserIDENTIFIER, i)
}

func (s *Use_declarationContext) LBRACE() antlr.TerminalNode {
	return s.GetToken(KuneiformParserLBRACE, 0)
}

func (s *Use_declarationContext) AllCOL() []antlr.TerminalNode {
	return s.GetTokens(KuneiformParserCOL)
}

func (s *Use_declarationContext) COL(i int) antlr.TerminalNode {
	return s.GetToken(KuneiformParserCOL, i)
}

func (s *Use_declarationContext) AllLiteral() []ILiteralContext {
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

func (s *Use_declarationContext) Literal(i int) ILiteralContext {
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

func (s *Use_declarationContext) RBRACE() antlr.TerminalNode {
	return s.GetToken(KuneiformParserRBRACE, 0)
}

func (s *Use_declarationContext) COMMA() antlr.TerminalNode {
	return s.GetToken(KuneiformParserCOMMA, 0)
}

func (s *Use_declarationContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Use_declarationContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Use_declarationContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case KuneiformParserVisitor:
		return t.VisitUse_declaration(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *KuneiformParser) Use_declaration() (localctx IUse_declarationContext) {
	localctx = NewUse_declarationContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 6, KuneiformParserRULE_use_declaration)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(61)
		p.Match(KuneiformParserUSE)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(62)

		var _m = p.Match(KuneiformParserIDENTIFIER)

		localctx.(*Use_declarationContext).extension_name = _m
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(75)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == KuneiformParserLBRACE {
		{
			p.SetState(63)
			p.Match(KuneiformParserLBRACE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(64)
			p.Match(KuneiformParserIDENTIFIER)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(65)
			p.Match(KuneiformParserCOL)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(66)
			p.Literal()
		}
		p.SetState(71)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == KuneiformParserCOMMA {
			{
				p.SetState(67)
				p.Match(KuneiformParserCOMMA)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(68)
				p.Match(KuneiformParserIDENTIFIER)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(69)
				p.Match(KuneiformParserCOL)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(70)
				p.Literal()
			}

		}
		{
			p.SetState(73)
			p.Match(KuneiformParserRBRACE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	}
	{
		p.SetState(77)
		p.Match(KuneiformParserAS)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(78)

		var _m = p.Match(KuneiformParserIDENTIFIER)

		localctx.(*Use_declarationContext).alias = _m
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(79)
		p.Match(KuneiformParserSCOL)
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

// ITable_declarationContext is an interface to support dynamic dispatch.
type ITable_declarationContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	TABLE() antlr.TerminalNode
	IDENTIFIER() antlr.TerminalNode
	LBRACE() antlr.TerminalNode
	AllColumn_def() []IColumn_defContext
	Column_def(i int) IColumn_defContext
	RBRACE() antlr.TerminalNode
	AllCOMMA() []antlr.TerminalNode
	COMMA(i int) antlr.TerminalNode
	AllIndex_def() []IIndex_defContext
	Index_def(i int) IIndex_defContext
	AllForeign_key_def() []IForeign_key_defContext
	Foreign_key_def(i int) IForeign_key_defContext

	// IsTable_declarationContext differentiates from other interfaces.
	IsTable_declarationContext()
}

type Table_declarationContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyTable_declarationContext() *Table_declarationContext {
	var p = new(Table_declarationContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = KuneiformParserRULE_table_declaration
	return p
}

func InitEmptyTable_declarationContext(p *Table_declarationContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = KuneiformParserRULE_table_declaration
}

func (*Table_declarationContext) IsTable_declarationContext() {}

func NewTable_declarationContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Table_declarationContext {
	var p = new(Table_declarationContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = KuneiformParserRULE_table_declaration

	return p
}

func (s *Table_declarationContext) GetParser() antlr.Parser { return s.parser }

func (s *Table_declarationContext) TABLE() antlr.TerminalNode {
	return s.GetToken(KuneiformParserTABLE, 0)
}

func (s *Table_declarationContext) IDENTIFIER() antlr.TerminalNode {
	return s.GetToken(KuneiformParserIDENTIFIER, 0)
}

func (s *Table_declarationContext) LBRACE() antlr.TerminalNode {
	return s.GetToken(KuneiformParserLBRACE, 0)
}

func (s *Table_declarationContext) AllColumn_def() []IColumn_defContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IColumn_defContext); ok {
			len++
		}
	}

	tst := make([]IColumn_defContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IColumn_defContext); ok {
			tst[i] = t.(IColumn_defContext)
			i++
		}
	}

	return tst
}

func (s *Table_declarationContext) Column_def(i int) IColumn_defContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IColumn_defContext); ok {
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

	return t.(IColumn_defContext)
}

func (s *Table_declarationContext) RBRACE() antlr.TerminalNode {
	return s.GetToken(KuneiformParserRBRACE, 0)
}

func (s *Table_declarationContext) AllCOMMA() []antlr.TerminalNode {
	return s.GetTokens(KuneiformParserCOMMA)
}

func (s *Table_declarationContext) COMMA(i int) antlr.TerminalNode {
	return s.GetToken(KuneiformParserCOMMA, i)
}

func (s *Table_declarationContext) AllIndex_def() []IIndex_defContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IIndex_defContext); ok {
			len++
		}
	}

	tst := make([]IIndex_defContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IIndex_defContext); ok {
			tst[i] = t.(IIndex_defContext)
			i++
		}
	}

	return tst
}

func (s *Table_declarationContext) Index_def(i int) IIndex_defContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IIndex_defContext); ok {
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

	return t.(IIndex_defContext)
}

func (s *Table_declarationContext) AllForeign_key_def() []IForeign_key_defContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IForeign_key_defContext); ok {
			len++
		}
	}

	tst := make([]IForeign_key_defContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IForeign_key_defContext); ok {
			tst[i] = t.(IForeign_key_defContext)
			i++
		}
	}

	return tst
}

func (s *Table_declarationContext) Foreign_key_def(i int) IForeign_key_defContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IForeign_key_defContext); ok {
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

	return t.(IForeign_key_defContext)
}

func (s *Table_declarationContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Table_declarationContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Table_declarationContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case KuneiformParserVisitor:
		return t.VisitTable_declaration(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *KuneiformParser) Table_declaration() (localctx ITable_declarationContext) {
	localctx = NewTable_declarationContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 8, KuneiformParserRULE_table_declaration)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(81)
		p.Match(KuneiformParserTABLE)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(82)
		p.Match(KuneiformParserIDENTIFIER)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(83)
		p.Match(KuneiformParserLBRACE)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(84)
		p.Column_def()
	}
	p.SetState(93)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == KuneiformParserCOMMA {
		{
			p.SetState(85)
			p.Match(KuneiformParserCOMMA)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(89)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}

		switch p.GetTokenStream().LA(1) {
		case KuneiformParserIDENTIFIER:
			{
				p.SetState(86)
				p.Column_def()
			}

		case KuneiformParserINDEX_NAME:
			{
				p.SetState(87)
				p.Index_def()
			}

		case KuneiformParserFOREIGN_KEY:
			{
				p.SetState(88)
				p.Foreign_key_def()
			}

		default:
			p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
			goto errorExit
		}

		p.SetState(95)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(96)
		p.Match(KuneiformParserRBRACE)
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

// IColumn_defContext is an interface to support dynamic dispatch.
type IColumn_defContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// GetName returns the name token.
	GetName() antlr.Token

	// SetName sets the name token.
	SetName(antlr.Token)

	// GetType_ returns the type_ rule contexts.
	GetType_() IType_selectorContext

	// SetType_ sets the type_ rule contexts.
	SetType_(IType_selectorContext)

	// Getter signatures
	IDENTIFIER() antlr.TerminalNode
	Type_selector() IType_selectorContext
	AllConstraint() []IConstraintContext
	Constraint(i int) IConstraintContext

	// IsColumn_defContext differentiates from other interfaces.
	IsColumn_defContext()
}

type Column_defContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
	name   antlr.Token
	type_  IType_selectorContext
}

func NewEmptyColumn_defContext() *Column_defContext {
	var p = new(Column_defContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = KuneiformParserRULE_column_def
	return p
}

func InitEmptyColumn_defContext(p *Column_defContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = KuneiformParserRULE_column_def
}

func (*Column_defContext) IsColumn_defContext() {}

func NewColumn_defContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Column_defContext {
	var p = new(Column_defContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = KuneiformParserRULE_column_def

	return p
}

func (s *Column_defContext) GetParser() antlr.Parser { return s.parser }

func (s *Column_defContext) GetName() antlr.Token { return s.name }

func (s *Column_defContext) SetName(v antlr.Token) { s.name = v }

func (s *Column_defContext) GetType_() IType_selectorContext { return s.type_ }

func (s *Column_defContext) SetType_(v IType_selectorContext) { s.type_ = v }

func (s *Column_defContext) IDENTIFIER() antlr.TerminalNode {
	return s.GetToken(KuneiformParserIDENTIFIER, 0)
}

func (s *Column_defContext) Type_selector() IType_selectorContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IType_selectorContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IType_selectorContext)
}

func (s *Column_defContext) AllConstraint() []IConstraintContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IConstraintContext); ok {
			len++
		}
	}

	tst := make([]IConstraintContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IConstraintContext); ok {
			tst[i] = t.(IConstraintContext)
			i++
		}
	}

	return tst
}

func (s *Column_defContext) Constraint(i int) IConstraintContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IConstraintContext); ok {
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

	return t.(IConstraintContext)
}

func (s *Column_defContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Column_defContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Column_defContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case KuneiformParserVisitor:
		return t.VisitColumn_def(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *KuneiformParser) Column_def() (localctx IColumn_defContext) {
	localctx = NewColumn_defContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 10, KuneiformParserRULE_column_def)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(98)

		var _m = p.Match(KuneiformParserIDENTIFIER)

		localctx.(*Column_defContext).name = _m
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(99)

		var _x = p.Type_selector()

		localctx.(*Column_defContext).type_ = _x
	}
	p.SetState(103)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for (int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&33423360) != 0 {
		{
			p.SetState(100)
			p.Constraint()
		}

		p.SetState(105)
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

// IIndex_defContext is an interface to support dynamic dispatch.
type IIndex_defContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// GetColumns returns the columns rule contexts.
	GetColumns() IIdentifier_listContext

	// SetColumns sets the columns rule contexts.
	SetColumns(IIdentifier_listContext)

	// Getter signatures
	INDEX_NAME() antlr.TerminalNode
	LPAREN() antlr.TerminalNode
	RPAREN() antlr.TerminalNode
	UNIQUE() antlr.TerminalNode
	INDEX() antlr.TerminalNode
	PRIMARY() antlr.TerminalNode
	Identifier_list() IIdentifier_listContext

	// IsIndex_defContext differentiates from other interfaces.
	IsIndex_defContext()
}

type Index_defContext struct {
	antlr.BaseParserRuleContext
	parser  antlr.Parser
	columns IIdentifier_listContext
}

func NewEmptyIndex_defContext() *Index_defContext {
	var p = new(Index_defContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = KuneiformParserRULE_index_def
	return p
}

func InitEmptyIndex_defContext(p *Index_defContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = KuneiformParserRULE_index_def
}

func (*Index_defContext) IsIndex_defContext() {}

func NewIndex_defContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Index_defContext {
	var p = new(Index_defContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = KuneiformParserRULE_index_def

	return p
}

func (s *Index_defContext) GetParser() antlr.Parser { return s.parser }

func (s *Index_defContext) GetColumns() IIdentifier_listContext { return s.columns }

func (s *Index_defContext) SetColumns(v IIdentifier_listContext) { s.columns = v }

func (s *Index_defContext) INDEX_NAME() antlr.TerminalNode {
	return s.GetToken(KuneiformParserINDEX_NAME, 0)
}

func (s *Index_defContext) LPAREN() antlr.TerminalNode {
	return s.GetToken(KuneiformParserLPAREN, 0)
}

func (s *Index_defContext) RPAREN() antlr.TerminalNode {
	return s.GetToken(KuneiformParserRPAREN, 0)
}

func (s *Index_defContext) UNIQUE() antlr.TerminalNode {
	return s.GetToken(KuneiformParserUNIQUE, 0)
}

func (s *Index_defContext) INDEX() antlr.TerminalNode {
	return s.GetToken(KuneiformParserINDEX, 0)
}

func (s *Index_defContext) PRIMARY() antlr.TerminalNode {
	return s.GetToken(KuneiformParserPRIMARY, 0)
}

func (s *Index_defContext) Identifier_list() IIdentifier_listContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IIdentifier_listContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IIdentifier_listContext)
}

func (s *Index_defContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Index_defContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Index_defContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case KuneiformParserVisitor:
		return t.VisitIndex_def(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *KuneiformParser) Index_def() (localctx IIndex_defContext) {
	localctx = NewIndex_defContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 12, KuneiformParserRULE_index_def)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(106)
		p.Match(KuneiformParserINDEX_NAME)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(107)
		_la = p.GetTokenStream().LA(1)

		if !((int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&54525952) != 0) {
			p.GetErrorHandler().RecoverInline(p)
		} else {
			p.GetErrorHandler().ReportMatch(p)
			p.Consume()
		}
	}
	{
		p.SetState(108)
		p.Match(KuneiformParserLPAREN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(109)

		var _x = p.Identifier_list()

		localctx.(*Index_defContext).columns = _x
	}
	{
		p.SetState(110)
		p.Match(KuneiformParserRPAREN)
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

// IForeign_key_defContext is an interface to support dynamic dispatch.
type IForeign_key_defContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// GetParent_table returns the parent_table token.
	GetParent_table() antlr.Token

	// SetParent_table sets the parent_table token.
	SetParent_table(antlr.Token)

	// GetChild_keys returns the child_keys rule contexts.
	GetChild_keys() IIdentifier_listContext

	// GetParent_keys returns the parent_keys rule contexts.
	GetParent_keys() IIdentifier_listContext

	// SetChild_keys sets the child_keys rule contexts.
	SetChild_keys(IIdentifier_listContext)

	// SetParent_keys sets the parent_keys rule contexts.
	SetParent_keys(IIdentifier_listContext)

	// Getter signatures
	FOREIGN_KEY() antlr.TerminalNode
	AllLPAREN() []antlr.TerminalNode
	LPAREN(i int) antlr.TerminalNode
	AllRPAREN() []antlr.TerminalNode
	RPAREN(i int) antlr.TerminalNode
	REFERENCES() antlr.TerminalNode
	AllIdentifier_list() []IIdentifier_listContext
	Identifier_list(i int) IIdentifier_listContext
	IDENTIFIER() antlr.TerminalNode
	AllForeign_key_action() []IForeign_key_actionContext
	Foreign_key_action(i int) IForeign_key_actionContext

	// IsForeign_key_defContext differentiates from other interfaces.
	IsForeign_key_defContext()
}

type Foreign_key_defContext struct {
	antlr.BaseParserRuleContext
	parser       antlr.Parser
	child_keys   IIdentifier_listContext
	parent_table antlr.Token
	parent_keys  IIdentifier_listContext
}

func NewEmptyForeign_key_defContext() *Foreign_key_defContext {
	var p = new(Foreign_key_defContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = KuneiformParserRULE_foreign_key_def
	return p
}

func InitEmptyForeign_key_defContext(p *Foreign_key_defContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = KuneiformParserRULE_foreign_key_def
}

func (*Foreign_key_defContext) IsForeign_key_defContext() {}

func NewForeign_key_defContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Foreign_key_defContext {
	var p = new(Foreign_key_defContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = KuneiformParserRULE_foreign_key_def

	return p
}

func (s *Foreign_key_defContext) GetParser() antlr.Parser { return s.parser }

func (s *Foreign_key_defContext) GetParent_table() antlr.Token { return s.parent_table }

func (s *Foreign_key_defContext) SetParent_table(v antlr.Token) { s.parent_table = v }

func (s *Foreign_key_defContext) GetChild_keys() IIdentifier_listContext { return s.child_keys }

func (s *Foreign_key_defContext) GetParent_keys() IIdentifier_listContext { return s.parent_keys }

func (s *Foreign_key_defContext) SetChild_keys(v IIdentifier_listContext) { s.child_keys = v }

func (s *Foreign_key_defContext) SetParent_keys(v IIdentifier_listContext) { s.parent_keys = v }

func (s *Foreign_key_defContext) FOREIGN_KEY() antlr.TerminalNode {
	return s.GetToken(KuneiformParserFOREIGN_KEY, 0)
}

func (s *Foreign_key_defContext) AllLPAREN() []antlr.TerminalNode {
	return s.GetTokens(KuneiformParserLPAREN)
}

func (s *Foreign_key_defContext) LPAREN(i int) antlr.TerminalNode {
	return s.GetToken(KuneiformParserLPAREN, i)
}

func (s *Foreign_key_defContext) AllRPAREN() []antlr.TerminalNode {
	return s.GetTokens(KuneiformParserRPAREN)
}

func (s *Foreign_key_defContext) RPAREN(i int) antlr.TerminalNode {
	return s.GetToken(KuneiformParserRPAREN, i)
}

func (s *Foreign_key_defContext) REFERENCES() antlr.TerminalNode {
	return s.GetToken(KuneiformParserREFERENCES, 0)
}

func (s *Foreign_key_defContext) AllIdentifier_list() []IIdentifier_listContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IIdentifier_listContext); ok {
			len++
		}
	}

	tst := make([]IIdentifier_listContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IIdentifier_listContext); ok {
			tst[i] = t.(IIdentifier_listContext)
			i++
		}
	}

	return tst
}

func (s *Foreign_key_defContext) Identifier_list(i int) IIdentifier_listContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IIdentifier_listContext); ok {
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

	return t.(IIdentifier_listContext)
}

func (s *Foreign_key_defContext) IDENTIFIER() antlr.TerminalNode {
	return s.GetToken(KuneiformParserIDENTIFIER, 0)
}

func (s *Foreign_key_defContext) AllForeign_key_action() []IForeign_key_actionContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IForeign_key_actionContext); ok {
			len++
		}
	}

	tst := make([]IForeign_key_actionContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IForeign_key_actionContext); ok {
			tst[i] = t.(IForeign_key_actionContext)
			i++
		}
	}

	return tst
}

func (s *Foreign_key_defContext) Foreign_key_action(i int) IForeign_key_actionContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IForeign_key_actionContext); ok {
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

	return t.(IForeign_key_actionContext)
}

func (s *Foreign_key_defContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Foreign_key_defContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Foreign_key_defContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case KuneiformParserVisitor:
		return t.VisitForeign_key_def(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *KuneiformParser) Foreign_key_def() (localctx IForeign_key_defContext) {
	localctx = NewForeign_key_defContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 14, KuneiformParserRULE_foreign_key_def)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(112)
		p.Match(KuneiformParserFOREIGN_KEY)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(113)
		p.Match(KuneiformParserLPAREN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(114)

		var _x = p.Identifier_list()

		localctx.(*Foreign_key_defContext).child_keys = _x
	}
	{
		p.SetState(115)
		p.Match(KuneiformParserRPAREN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(116)
		p.Match(KuneiformParserREFERENCES)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(117)

		var _m = p.Match(KuneiformParserIDENTIFIER)

		localctx.(*Foreign_key_defContext).parent_table = _m
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(118)
		p.Match(KuneiformParserLPAREN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(119)

		var _x = p.Identifier_list()

		localctx.(*Foreign_key_defContext).parent_keys = _x
	}
	{
		p.SetState(120)
		p.Match(KuneiformParserRPAREN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(124)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == KuneiformParserON_UPDATE || _la == KuneiformParserON_DELETE {
		{
			p.SetState(121)
			p.Foreign_key_action()
		}

		p.SetState(126)
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

// IForeign_key_actionContext is an interface to support dynamic dispatch.
type IForeign_key_actionContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	ON_UPDATE() antlr.TerminalNode
	ON_DELETE() antlr.TerminalNode
	DO_NO_ACTION() antlr.TerminalNode
	DO_CASCADE() antlr.TerminalNode
	DO_SET_NULL() antlr.TerminalNode
	DO_SET_DEFAULT() antlr.TerminalNode
	DO_RESTRICT() antlr.TerminalNode
	DO() antlr.TerminalNode

	// IsForeign_key_actionContext differentiates from other interfaces.
	IsForeign_key_actionContext()
}

type Foreign_key_actionContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyForeign_key_actionContext() *Foreign_key_actionContext {
	var p = new(Foreign_key_actionContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = KuneiformParserRULE_foreign_key_action
	return p
}

func InitEmptyForeign_key_actionContext(p *Foreign_key_actionContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = KuneiformParserRULE_foreign_key_action
}

func (*Foreign_key_actionContext) IsForeign_key_actionContext() {}

func NewForeign_key_actionContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Foreign_key_actionContext {
	var p = new(Foreign_key_actionContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = KuneiformParserRULE_foreign_key_action

	return p
}

func (s *Foreign_key_actionContext) GetParser() antlr.Parser { return s.parser }

func (s *Foreign_key_actionContext) ON_UPDATE() antlr.TerminalNode {
	return s.GetToken(KuneiformParserON_UPDATE, 0)
}

func (s *Foreign_key_actionContext) ON_DELETE() antlr.TerminalNode {
	return s.GetToken(KuneiformParserON_DELETE, 0)
}

func (s *Foreign_key_actionContext) DO_NO_ACTION() antlr.TerminalNode {
	return s.GetToken(KuneiformParserDO_NO_ACTION, 0)
}

func (s *Foreign_key_actionContext) DO_CASCADE() antlr.TerminalNode {
	return s.GetToken(KuneiformParserDO_CASCADE, 0)
}

func (s *Foreign_key_actionContext) DO_SET_NULL() antlr.TerminalNode {
	return s.GetToken(KuneiformParserDO_SET_NULL, 0)
}

func (s *Foreign_key_actionContext) DO_SET_DEFAULT() antlr.TerminalNode {
	return s.GetToken(KuneiformParserDO_SET_DEFAULT, 0)
}

func (s *Foreign_key_actionContext) DO_RESTRICT() antlr.TerminalNode {
	return s.GetToken(KuneiformParserDO_RESTRICT, 0)
}

func (s *Foreign_key_actionContext) DO() antlr.TerminalNode {
	return s.GetToken(KuneiformParserDO, 0)
}

func (s *Foreign_key_actionContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Foreign_key_actionContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Foreign_key_actionContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case KuneiformParserVisitor:
		return t.VisitForeign_key_action(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *KuneiformParser) Foreign_key_action() (localctx IForeign_key_actionContext) {
	localctx = NewForeign_key_actionContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 16, KuneiformParserRULE_foreign_key_action)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(127)
		_la = p.GetTokenStream().LA(1)

		if !(_la == KuneiformParserON_UPDATE || _la == KuneiformParserON_DELETE) {
			p.GetErrorHandler().RecoverInline(p)
		} else {
			p.GetErrorHandler().ReportMatch(p)
			p.Consume()
		}
	}
	p.SetState(129)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == KuneiformParserDO {
		{
			p.SetState(128)
			p.Match(KuneiformParserDO)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	}
	{
		p.SetState(131)
		_la = p.GetTokenStream().LA(1)

		if !((int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&133143986176) != 0) {
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

// IIdentifier_listContext is an interface to support dynamic dispatch.
type IIdentifier_listContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllIDENTIFIER() []antlr.TerminalNode
	IDENTIFIER(i int) antlr.TerminalNode
	AllCOMMA() []antlr.TerminalNode
	COMMA(i int) antlr.TerminalNode

	// IsIdentifier_listContext differentiates from other interfaces.
	IsIdentifier_listContext()
}

type Identifier_listContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyIdentifier_listContext() *Identifier_listContext {
	var p = new(Identifier_listContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = KuneiformParserRULE_identifier_list
	return p
}

func InitEmptyIdentifier_listContext(p *Identifier_listContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = KuneiformParserRULE_identifier_list
}

func (*Identifier_listContext) IsIdentifier_listContext() {}

func NewIdentifier_listContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Identifier_listContext {
	var p = new(Identifier_listContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = KuneiformParserRULE_identifier_list

	return p
}

func (s *Identifier_listContext) GetParser() antlr.Parser { return s.parser }

func (s *Identifier_listContext) AllIDENTIFIER() []antlr.TerminalNode {
	return s.GetTokens(KuneiformParserIDENTIFIER)
}

func (s *Identifier_listContext) IDENTIFIER(i int) antlr.TerminalNode {
	return s.GetToken(KuneiformParserIDENTIFIER, i)
}

func (s *Identifier_listContext) AllCOMMA() []antlr.TerminalNode {
	return s.GetTokens(KuneiformParserCOMMA)
}

func (s *Identifier_listContext) COMMA(i int) antlr.TerminalNode {
	return s.GetToken(KuneiformParserCOMMA, i)
}

func (s *Identifier_listContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Identifier_listContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Identifier_listContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case KuneiformParserVisitor:
		return t.VisitIdentifier_list(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *KuneiformParser) Identifier_list() (localctx IIdentifier_listContext) {
	localctx = NewIdentifier_listContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 18, KuneiformParserRULE_identifier_list)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(133)
		p.Match(KuneiformParserIDENTIFIER)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(138)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == KuneiformParserCOMMA {
		{
			p.SetState(134)
			p.Match(KuneiformParserCOMMA)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(135)
			p.Match(KuneiformParserIDENTIFIER)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
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

// ILiteralContext is an interface to support dynamic dispatch.
type ILiteralContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	NUMERIC_LITERAL() antlr.TerminalNode
	BLOB_LITERAL() antlr.TerminalNode
	TEXT_LITERAL() antlr.TerminalNode
	BOOLEAN_LITERAL() antlr.TerminalNode

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
	p.RuleIndex = KuneiformParserRULE_literal
	return p
}

func InitEmptyLiteralContext(p *LiteralContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = KuneiformParserRULE_literal
}

func (*LiteralContext) IsLiteralContext() {}

func NewLiteralContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *LiteralContext {
	var p = new(LiteralContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = KuneiformParserRULE_literal

	return p
}

func (s *LiteralContext) GetParser() antlr.Parser { return s.parser }

func (s *LiteralContext) NUMERIC_LITERAL() antlr.TerminalNode {
	return s.GetToken(KuneiformParserNUMERIC_LITERAL, 0)
}

func (s *LiteralContext) BLOB_LITERAL() antlr.TerminalNode {
	return s.GetToken(KuneiformParserBLOB_LITERAL, 0)
}

func (s *LiteralContext) TEXT_LITERAL() antlr.TerminalNode {
	return s.GetToken(KuneiformParserTEXT_LITERAL, 0)
}

func (s *LiteralContext) BOOLEAN_LITERAL() antlr.TerminalNode {
	return s.GetToken(KuneiformParserBOOLEAN_LITERAL, 0)
}

func (s *LiteralContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *LiteralContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *LiteralContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case KuneiformParserVisitor:
		return t.VisitLiteral(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *KuneiformParser) Literal() (localctx ILiteralContext) {
	localctx = NewLiteralContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 20, KuneiformParserRULE_literal)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(141)
		_la = p.GetTokenStream().LA(1)

		if !((int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&16492674416640) != 0) {
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

// IType_selectorContext is an interface to support dynamic dispatch.
type IType_selectorContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// GetType_ returns the type_ token.
	GetType_() antlr.Token

	// SetType_ sets the type_ token.
	SetType_(antlr.Token)

	// Getter signatures
	IDENTIFIER() antlr.TerminalNode
	LBRACKET() antlr.TerminalNode
	RBRACKET() antlr.TerminalNode

	// IsType_selectorContext differentiates from other interfaces.
	IsType_selectorContext()
}

type Type_selectorContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
	type_  antlr.Token
}

func NewEmptyType_selectorContext() *Type_selectorContext {
	var p = new(Type_selectorContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = KuneiformParserRULE_type_selector
	return p
}

func InitEmptyType_selectorContext(p *Type_selectorContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = KuneiformParserRULE_type_selector
}

func (*Type_selectorContext) IsType_selectorContext() {}

func NewType_selectorContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Type_selectorContext {
	var p = new(Type_selectorContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = KuneiformParserRULE_type_selector

	return p
}

func (s *Type_selectorContext) GetParser() antlr.Parser { return s.parser }

func (s *Type_selectorContext) GetType_() antlr.Token { return s.type_ }

func (s *Type_selectorContext) SetType_(v antlr.Token) { s.type_ = v }

func (s *Type_selectorContext) IDENTIFIER() antlr.TerminalNode {
	return s.GetToken(KuneiformParserIDENTIFIER, 0)
}

func (s *Type_selectorContext) LBRACKET() antlr.TerminalNode {
	return s.GetToken(KuneiformParserLBRACKET, 0)
}

func (s *Type_selectorContext) RBRACKET() antlr.TerminalNode {
	return s.GetToken(KuneiformParserRBRACKET, 0)
}

func (s *Type_selectorContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Type_selectorContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Type_selectorContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case KuneiformParserVisitor:
		return t.VisitType_selector(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *KuneiformParser) Type_selector() (localctx IType_selectorContext) {
	localctx = NewType_selectorContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 22, KuneiformParserRULE_type_selector)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(143)

		var _m = p.Match(KuneiformParserIDENTIFIER)

		localctx.(*Type_selectorContext).type_ = _m
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(146)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == KuneiformParserLBRACKET {
		{
			p.SetState(144)
			p.Match(KuneiformParserLBRACKET)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(145)
			p.Match(KuneiformParserRBRACKET)
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

// IConstraintContext is an interface to support dynamic dispatch.
type IConstraintContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser
	// IsConstraintContext differentiates from other interfaces.
	IsConstraintContext()
}

type ConstraintContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyConstraintContext() *ConstraintContext {
	var p = new(ConstraintContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = KuneiformParserRULE_constraint
	return p
}

func InitEmptyConstraintContext(p *ConstraintContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = KuneiformParserRULE_constraint
}

func (*ConstraintContext) IsConstraintContext() {}

func NewConstraintContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ConstraintContext {
	var p = new(ConstraintContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = KuneiformParserRULE_constraint

	return p
}

func (s *ConstraintContext) GetParser() antlr.Parser { return s.parser }

func (s *ConstraintContext) CopyAll(ctx *ConstraintContext) {
	s.CopyFrom(&ctx.BaseParserRuleContext)
}

func (s *ConstraintContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ConstraintContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

type MIN_LENContext struct {
	ConstraintContext
}

func NewMIN_LENContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *MIN_LENContext {
	var p = new(MIN_LENContext)

	InitEmptyConstraintContext(&p.ConstraintContext)
	p.parser = parser
	p.CopyAll(ctx.(*ConstraintContext))

	return p
}

func (s *MIN_LENContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *MIN_LENContext) MIN_LEN() antlr.TerminalNode {
	return s.GetToken(KuneiformParserMIN_LEN, 0)
}

func (s *MIN_LENContext) LPAREN() antlr.TerminalNode {
	return s.GetToken(KuneiformParserLPAREN, 0)
}

func (s *MIN_LENContext) NUMERIC_LITERAL() antlr.TerminalNode {
	return s.GetToken(KuneiformParserNUMERIC_LITERAL, 0)
}

func (s *MIN_LENContext) RPAREN() antlr.TerminalNode {
	return s.GetToken(KuneiformParserRPAREN, 0)
}

func (s *MIN_LENContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case KuneiformParserVisitor:
		return t.VisitMIN_LEN(s)

	default:
		return t.VisitChildren(s)
	}
}

type MINContext struct {
	ConstraintContext
}

func NewMINContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *MINContext {
	var p = new(MINContext)

	InitEmptyConstraintContext(&p.ConstraintContext)
	p.parser = parser
	p.CopyAll(ctx.(*ConstraintContext))

	return p
}

func (s *MINContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *MINContext) MIN() antlr.TerminalNode {
	return s.GetToken(KuneiformParserMIN, 0)
}

func (s *MINContext) LPAREN() antlr.TerminalNode {
	return s.GetToken(KuneiformParserLPAREN, 0)
}

func (s *MINContext) NUMERIC_LITERAL() antlr.TerminalNode {
	return s.GetToken(KuneiformParserNUMERIC_LITERAL, 0)
}

func (s *MINContext) RPAREN() antlr.TerminalNode {
	return s.GetToken(KuneiformParserRPAREN, 0)
}

func (s *MINContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case KuneiformParserVisitor:
		return t.VisitMIN(s)

	default:
		return t.VisitChildren(s)
	}
}

type PRIMARY_KEYContext struct {
	ConstraintContext
}

func NewPRIMARY_KEYContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *PRIMARY_KEYContext {
	var p = new(PRIMARY_KEYContext)

	InitEmptyConstraintContext(&p.ConstraintContext)
	p.parser = parser
	p.CopyAll(ctx.(*ConstraintContext))

	return p
}

func (s *PRIMARY_KEYContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *PRIMARY_KEYContext) PRIMARY() antlr.TerminalNode {
	return s.GetToken(KuneiformParserPRIMARY, 0)
}

func (s *PRIMARY_KEYContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case KuneiformParserVisitor:
		return t.VisitPRIMARY_KEY(s)

	default:
		return t.VisitChildren(s)
	}
}

type MAXContext struct {
	ConstraintContext
}

func NewMAXContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *MAXContext {
	var p = new(MAXContext)

	InitEmptyConstraintContext(&p.ConstraintContext)
	p.parser = parser
	p.CopyAll(ctx.(*ConstraintContext))

	return p
}

func (s *MAXContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *MAXContext) MAX() antlr.TerminalNode {
	return s.GetToken(KuneiformParserMAX, 0)
}

func (s *MAXContext) LPAREN() antlr.TerminalNode {
	return s.GetToken(KuneiformParserLPAREN, 0)
}

func (s *MAXContext) NUMERIC_LITERAL() antlr.TerminalNode {
	return s.GetToken(KuneiformParserNUMERIC_LITERAL, 0)
}

func (s *MAXContext) RPAREN() antlr.TerminalNode {
	return s.GetToken(KuneiformParserRPAREN, 0)
}

func (s *MAXContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case KuneiformParserVisitor:
		return t.VisitMAX(s)

	default:
		return t.VisitChildren(s)
	}
}

type MAX_LENContext struct {
	ConstraintContext
}

func NewMAX_LENContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *MAX_LENContext {
	var p = new(MAX_LENContext)

	InitEmptyConstraintContext(&p.ConstraintContext)
	p.parser = parser
	p.CopyAll(ctx.(*ConstraintContext))

	return p
}

func (s *MAX_LENContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *MAX_LENContext) MAX_LEN() antlr.TerminalNode {
	return s.GetToken(KuneiformParserMAX_LEN, 0)
}

func (s *MAX_LENContext) LPAREN() antlr.TerminalNode {
	return s.GetToken(KuneiformParserLPAREN, 0)
}

func (s *MAX_LENContext) NUMERIC_LITERAL() antlr.TerminalNode {
	return s.GetToken(KuneiformParserNUMERIC_LITERAL, 0)
}

func (s *MAX_LENContext) RPAREN() antlr.TerminalNode {
	return s.GetToken(KuneiformParserRPAREN, 0)
}

func (s *MAX_LENContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case KuneiformParserVisitor:
		return t.VisitMAX_LEN(s)

	default:
		return t.VisitChildren(s)
	}
}

type UNIQUEContext struct {
	ConstraintContext
}

func NewUNIQUEContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *UNIQUEContext {
	var p = new(UNIQUEContext)

	InitEmptyConstraintContext(&p.ConstraintContext)
	p.parser = parser
	p.CopyAll(ctx.(*ConstraintContext))

	return p
}

func (s *UNIQUEContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *UNIQUEContext) UNIQUE() antlr.TerminalNode {
	return s.GetToken(KuneiformParserUNIQUE, 0)
}

func (s *UNIQUEContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case KuneiformParserVisitor:
		return t.VisitUNIQUE(s)

	default:
		return t.VisitChildren(s)
	}
}

type NOT_NULLContext struct {
	ConstraintContext
}

func NewNOT_NULLContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *NOT_NULLContext {
	var p = new(NOT_NULLContext)

	InitEmptyConstraintContext(&p.ConstraintContext)
	p.parser = parser
	p.CopyAll(ctx.(*ConstraintContext))

	return p
}

func (s *NOT_NULLContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *NOT_NULLContext) NOT_NULL() antlr.TerminalNode {
	return s.GetToken(KuneiformParserNOT_NULL, 0)
}

func (s *NOT_NULLContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case KuneiformParserVisitor:
		return t.VisitNOT_NULL(s)

	default:
		return t.VisitChildren(s)
	}
}

type DEFAULTContext struct {
	ConstraintContext
}

func NewDEFAULTContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *DEFAULTContext {
	var p = new(DEFAULTContext)

	InitEmptyConstraintContext(&p.ConstraintContext)
	p.parser = parser
	p.CopyAll(ctx.(*ConstraintContext))

	return p
}

func (s *DEFAULTContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *DEFAULTContext) DEFAULT() antlr.TerminalNode {
	return s.GetToken(KuneiformParserDEFAULT, 0)
}

func (s *DEFAULTContext) LPAREN() antlr.TerminalNode {
	return s.GetToken(KuneiformParserLPAREN, 0)
}

func (s *DEFAULTContext) Literal() ILiteralContext {
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

func (s *DEFAULTContext) RPAREN() antlr.TerminalNode {
	return s.GetToken(KuneiformParserRPAREN, 0)
}

func (s *DEFAULTContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case KuneiformParserVisitor:
		return t.VisitDEFAULT(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *KuneiformParser) Constraint() (localctx IConstraintContext) {
	localctx = NewConstraintContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 24, KuneiformParserRULE_constraint)
	p.SetState(172)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case KuneiformParserMIN:
		localctx = NewMINContext(p, localctx)
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(148)
			p.Match(KuneiformParserMIN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(149)
			p.Match(KuneiformParserLPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(150)
			p.Match(KuneiformParserNUMERIC_LITERAL)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(151)
			p.Match(KuneiformParserRPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case KuneiformParserMAX:
		localctx = NewMAXContext(p, localctx)
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(152)
			p.Match(KuneiformParserMAX)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(153)
			p.Match(KuneiformParserLPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(154)
			p.Match(KuneiformParserNUMERIC_LITERAL)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(155)
			p.Match(KuneiformParserRPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case KuneiformParserMIN_LEN:
		localctx = NewMIN_LENContext(p, localctx)
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(156)
			p.Match(KuneiformParserMIN_LEN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(157)
			p.Match(KuneiformParserLPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(158)
			p.Match(KuneiformParserNUMERIC_LITERAL)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(159)
			p.Match(KuneiformParserRPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case KuneiformParserMAX_LEN:
		localctx = NewMAX_LENContext(p, localctx)
		p.EnterOuterAlt(localctx, 4)
		{
			p.SetState(160)
			p.Match(KuneiformParserMAX_LEN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(161)
			p.Match(KuneiformParserLPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(162)
			p.Match(KuneiformParserNUMERIC_LITERAL)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(163)
			p.Match(KuneiformParserRPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case KuneiformParserNOT_NULL:
		localctx = NewNOT_NULLContext(p, localctx)
		p.EnterOuterAlt(localctx, 5)
		{
			p.SetState(164)
			p.Match(KuneiformParserNOT_NULL)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case KuneiformParserPRIMARY:
		localctx = NewPRIMARY_KEYContext(p, localctx)
		p.EnterOuterAlt(localctx, 6)
		{
			p.SetState(165)
			p.Match(KuneiformParserPRIMARY)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case KuneiformParserDEFAULT:
		localctx = NewDEFAULTContext(p, localctx)
		p.EnterOuterAlt(localctx, 7)
		{
			p.SetState(166)
			p.Match(KuneiformParserDEFAULT)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(167)
			p.Match(KuneiformParserLPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(168)
			p.Literal()
		}
		{
			p.SetState(169)
			p.Match(KuneiformParserRPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case KuneiformParserUNIQUE:
		localctx = NewUNIQUEContext(p, localctx)
		p.EnterOuterAlt(localctx, 8)
		{
			p.SetState(171)
			p.Match(KuneiformParserUNIQUE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
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

// IAction_declarationContext is an interface to support dynamic dispatch.
type IAction_declarationContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	START_ACTION() antlr.TerminalNode
	STMT_IDENTIFIER() antlr.TerminalNode
	STMT_LPAREN() antlr.TerminalNode
	STMT_RPAREN() antlr.TerminalNode
	STMT_BODY() antlr.TerminalNode
	AllSTMT_VAR() []antlr.TerminalNode
	STMT_VAR(i int) antlr.TerminalNode
	AllSTMT_ACCESS() []antlr.TerminalNode
	STMT_ACCESS(i int) antlr.TerminalNode
	AllSTMT_COMMA() []antlr.TerminalNode
	STMT_COMMA(i int) antlr.TerminalNode

	// IsAction_declarationContext differentiates from other interfaces.
	IsAction_declarationContext()
}

type Action_declarationContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyAction_declarationContext() *Action_declarationContext {
	var p = new(Action_declarationContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = KuneiformParserRULE_action_declaration
	return p
}

func InitEmptyAction_declarationContext(p *Action_declarationContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = KuneiformParserRULE_action_declaration
}

func (*Action_declarationContext) IsAction_declarationContext() {}

func NewAction_declarationContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Action_declarationContext {
	var p = new(Action_declarationContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = KuneiformParserRULE_action_declaration

	return p
}

func (s *Action_declarationContext) GetParser() antlr.Parser { return s.parser }

func (s *Action_declarationContext) START_ACTION() antlr.TerminalNode {
	return s.GetToken(KuneiformParserSTART_ACTION, 0)
}

func (s *Action_declarationContext) STMT_IDENTIFIER() antlr.TerminalNode {
	return s.GetToken(KuneiformParserSTMT_IDENTIFIER, 0)
}

func (s *Action_declarationContext) STMT_LPAREN() antlr.TerminalNode {
	return s.GetToken(KuneiformParserSTMT_LPAREN, 0)
}

func (s *Action_declarationContext) STMT_RPAREN() antlr.TerminalNode {
	return s.GetToken(KuneiformParserSTMT_RPAREN, 0)
}

func (s *Action_declarationContext) STMT_BODY() antlr.TerminalNode {
	return s.GetToken(KuneiformParserSTMT_BODY, 0)
}

func (s *Action_declarationContext) AllSTMT_VAR() []antlr.TerminalNode {
	return s.GetTokens(KuneiformParserSTMT_VAR)
}

func (s *Action_declarationContext) STMT_VAR(i int) antlr.TerminalNode {
	return s.GetToken(KuneiformParserSTMT_VAR, i)
}

func (s *Action_declarationContext) AllSTMT_ACCESS() []antlr.TerminalNode {
	return s.GetTokens(KuneiformParserSTMT_ACCESS)
}

func (s *Action_declarationContext) STMT_ACCESS(i int) antlr.TerminalNode {
	return s.GetToken(KuneiformParserSTMT_ACCESS, i)
}

func (s *Action_declarationContext) AllSTMT_COMMA() []antlr.TerminalNode {
	return s.GetTokens(KuneiformParserSTMT_COMMA)
}

func (s *Action_declarationContext) STMT_COMMA(i int) antlr.TerminalNode {
	return s.GetToken(KuneiformParserSTMT_COMMA, i)
}

func (s *Action_declarationContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Action_declarationContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Action_declarationContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case KuneiformParserVisitor:
		return t.VisitAction_declaration(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *KuneiformParser) Action_declaration() (localctx IAction_declarationContext) {
	localctx = NewAction_declarationContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 26, KuneiformParserRULE_action_declaration)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(174)
		p.Match(KuneiformParserSTART_ACTION)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(175)
		p.Match(KuneiformParserSTMT_IDENTIFIER)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(176)
		p.Match(KuneiformParserSTMT_LPAREN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(185)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == KuneiformParserSTMT_VAR {
		{
			p.SetState(177)
			p.Match(KuneiformParserSTMT_VAR)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(182)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		for _la == KuneiformParserSTMT_COMMA {
			{
				p.SetState(178)
				p.Match(KuneiformParserSTMT_COMMA)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(179)
				p.Match(KuneiformParserSTMT_VAR)
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
		}

	}
	{
		p.SetState(187)
		p.Match(KuneiformParserSTMT_RPAREN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(189)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for ok := true; ok; ok = _la == KuneiformParserSTMT_ACCESS {
		{
			p.SetState(188)
			p.Match(KuneiformParserSTMT_ACCESS)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

		p.SetState(191)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(193)
		p.Match(KuneiformParserSTMT_BODY)
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

// IProcedure_declarationContext is an interface to support dynamic dispatch.
type IProcedure_declarationContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// GetProcedure_name returns the procedure_name token.
	GetProcedure_name() antlr.Token

	// SetProcedure_name sets the procedure_name token.
	SetProcedure_name(antlr.Token)

	// Getter signatures
	START_PROCEDURE() antlr.TerminalNode
	STMT_LPAREN() antlr.TerminalNode
	STMT_RPAREN() antlr.TerminalNode
	STMT_BODY() antlr.TerminalNode
	STMT_IDENTIFIER() antlr.TerminalNode
	Stmt_typed_param_list() IStmt_typed_param_listContext
	AllSTMT_ACCESS() []antlr.TerminalNode
	STMT_ACCESS(i int) antlr.TerminalNode
	STMT_RETURNS() antlr.TerminalNode
	Stmt_return() IStmt_returnContext

	// IsProcedure_declarationContext differentiates from other interfaces.
	IsProcedure_declarationContext()
}

type Procedure_declarationContext struct {
	antlr.BaseParserRuleContext
	parser         antlr.Parser
	procedure_name antlr.Token
}

func NewEmptyProcedure_declarationContext() *Procedure_declarationContext {
	var p = new(Procedure_declarationContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = KuneiformParserRULE_procedure_declaration
	return p
}

func InitEmptyProcedure_declarationContext(p *Procedure_declarationContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = KuneiformParserRULE_procedure_declaration
}

func (*Procedure_declarationContext) IsProcedure_declarationContext() {}

func NewProcedure_declarationContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Procedure_declarationContext {
	var p = new(Procedure_declarationContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = KuneiformParserRULE_procedure_declaration

	return p
}

func (s *Procedure_declarationContext) GetParser() antlr.Parser { return s.parser }

func (s *Procedure_declarationContext) GetProcedure_name() antlr.Token { return s.procedure_name }

func (s *Procedure_declarationContext) SetProcedure_name(v antlr.Token) { s.procedure_name = v }

func (s *Procedure_declarationContext) START_PROCEDURE() antlr.TerminalNode {
	return s.GetToken(KuneiformParserSTART_PROCEDURE, 0)
}

func (s *Procedure_declarationContext) STMT_LPAREN() antlr.TerminalNode {
	return s.GetToken(KuneiformParserSTMT_LPAREN, 0)
}

func (s *Procedure_declarationContext) STMT_RPAREN() antlr.TerminalNode {
	return s.GetToken(KuneiformParserSTMT_RPAREN, 0)
}

func (s *Procedure_declarationContext) STMT_BODY() antlr.TerminalNode {
	return s.GetToken(KuneiformParserSTMT_BODY, 0)
}

func (s *Procedure_declarationContext) STMT_IDENTIFIER() antlr.TerminalNode {
	return s.GetToken(KuneiformParserSTMT_IDENTIFIER, 0)
}

func (s *Procedure_declarationContext) Stmt_typed_param_list() IStmt_typed_param_listContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IStmt_typed_param_listContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IStmt_typed_param_listContext)
}

func (s *Procedure_declarationContext) AllSTMT_ACCESS() []antlr.TerminalNode {
	return s.GetTokens(KuneiformParserSTMT_ACCESS)
}

func (s *Procedure_declarationContext) STMT_ACCESS(i int) antlr.TerminalNode {
	return s.GetToken(KuneiformParserSTMT_ACCESS, i)
}

func (s *Procedure_declarationContext) STMT_RETURNS() antlr.TerminalNode {
	return s.GetToken(KuneiformParserSTMT_RETURNS, 0)
}

func (s *Procedure_declarationContext) Stmt_return() IStmt_returnContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IStmt_returnContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IStmt_returnContext)
}

func (s *Procedure_declarationContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Procedure_declarationContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Procedure_declarationContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case KuneiformParserVisitor:
		return t.VisitProcedure_declaration(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *KuneiformParser) Procedure_declaration() (localctx IProcedure_declarationContext) {
	localctx = NewProcedure_declarationContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 28, KuneiformParserRULE_procedure_declaration)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(195)
		p.Match(KuneiformParserSTART_PROCEDURE)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(196)

		var _m = p.Match(KuneiformParserSTMT_IDENTIFIER)

		localctx.(*Procedure_declarationContext).procedure_name = _m
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(197)
		p.Match(KuneiformParserSTMT_LPAREN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(199)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == KuneiformParserSTMT_VAR {
		{
			p.SetState(198)
			p.Stmt_typed_param_list()
		}

	}
	{
		p.SetState(201)
		p.Match(KuneiformParserSTMT_RPAREN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(203)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for ok := true; ok; ok = _la == KuneiformParserSTMT_ACCESS {
		{
			p.SetState(202)
			p.Match(KuneiformParserSTMT_ACCESS)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

		p.SetState(205)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	p.SetState(209)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == KuneiformParserSTMT_RETURNS {
		{
			p.SetState(207)
			p.Match(KuneiformParserSTMT_RETURNS)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(208)
			p.Stmt_return()
		}

	}
	{
		p.SetState(211)
		p.Match(KuneiformParserSTMT_BODY)
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

// IStmt_returnContext is an interface to support dynamic dispatch.
type IStmt_returnContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	STMT_LPAREN() antlr.TerminalNode
	AllSTMT_IDENTIFIER() []antlr.TerminalNode
	STMT_IDENTIFIER(i int) antlr.TerminalNode
	AllStmt_type_selector() []IStmt_type_selectorContext
	Stmt_type_selector(i int) IStmt_type_selectorContext
	STMT_RPAREN() antlr.TerminalNode
	STMT_TABLE() antlr.TerminalNode
	AllSTMT_COMMA() []antlr.TerminalNode
	STMT_COMMA(i int) antlr.TerminalNode

	// IsStmt_returnContext differentiates from other interfaces.
	IsStmt_returnContext()
}

type Stmt_returnContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyStmt_returnContext() *Stmt_returnContext {
	var p = new(Stmt_returnContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = KuneiformParserRULE_stmt_return
	return p
}

func InitEmptyStmt_returnContext(p *Stmt_returnContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = KuneiformParserRULE_stmt_return
}

func (*Stmt_returnContext) IsStmt_returnContext() {}

func NewStmt_returnContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Stmt_returnContext {
	var p = new(Stmt_returnContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = KuneiformParserRULE_stmt_return

	return p
}

func (s *Stmt_returnContext) GetParser() antlr.Parser { return s.parser }

func (s *Stmt_returnContext) STMT_LPAREN() antlr.TerminalNode {
	return s.GetToken(KuneiformParserSTMT_LPAREN, 0)
}

func (s *Stmt_returnContext) AllSTMT_IDENTIFIER() []antlr.TerminalNode {
	return s.GetTokens(KuneiformParserSTMT_IDENTIFIER)
}

func (s *Stmt_returnContext) STMT_IDENTIFIER(i int) antlr.TerminalNode {
	return s.GetToken(KuneiformParserSTMT_IDENTIFIER, i)
}

func (s *Stmt_returnContext) AllStmt_type_selector() []IStmt_type_selectorContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IStmt_type_selectorContext); ok {
			len++
		}
	}

	tst := make([]IStmt_type_selectorContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IStmt_type_selectorContext); ok {
			tst[i] = t.(IStmt_type_selectorContext)
			i++
		}
	}

	return tst
}

func (s *Stmt_returnContext) Stmt_type_selector(i int) IStmt_type_selectorContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IStmt_type_selectorContext); ok {
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

	return t.(IStmt_type_selectorContext)
}

func (s *Stmt_returnContext) STMT_RPAREN() antlr.TerminalNode {
	return s.GetToken(KuneiformParserSTMT_RPAREN, 0)
}

func (s *Stmt_returnContext) STMT_TABLE() antlr.TerminalNode {
	return s.GetToken(KuneiformParserSTMT_TABLE, 0)
}

func (s *Stmt_returnContext) AllSTMT_COMMA() []antlr.TerminalNode {
	return s.GetTokens(KuneiformParserSTMT_COMMA)
}

func (s *Stmt_returnContext) STMT_COMMA(i int) antlr.TerminalNode {
	return s.GetToken(KuneiformParserSTMT_COMMA, i)
}

func (s *Stmt_returnContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Stmt_returnContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Stmt_returnContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case KuneiformParserVisitor:
		return t.VisitStmt_return(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *KuneiformParser) Stmt_return() (localctx IStmt_returnContext) {
	localctx = NewStmt_returnContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 30, KuneiformParserRULE_stmt_return)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(214)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == KuneiformParserSTMT_TABLE {
		{
			p.SetState(213)
			p.Match(KuneiformParserSTMT_TABLE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	}
	{
		p.SetState(216)
		p.Match(KuneiformParserSTMT_LPAREN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(217)
		p.Match(KuneiformParserSTMT_IDENTIFIER)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(218)
		p.Stmt_type_selector()
	}
	p.SetState(224)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == KuneiformParserSTMT_COMMA {
		{
			p.SetState(219)
			p.Match(KuneiformParserSTMT_COMMA)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(220)
			p.Match(KuneiformParserSTMT_IDENTIFIER)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(221)
			p.Stmt_type_selector()
		}

		p.SetState(226)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(227)
		p.Match(KuneiformParserSTMT_RPAREN)
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

// IStmt_typed_param_listContext is an interface to support dynamic dispatch.
type IStmt_typed_param_listContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllSTMT_VAR() []antlr.TerminalNode
	STMT_VAR(i int) antlr.TerminalNode
	AllStmt_type_selector() []IStmt_type_selectorContext
	Stmt_type_selector(i int) IStmt_type_selectorContext
	AllSTMT_COMMA() []antlr.TerminalNode
	STMT_COMMA(i int) antlr.TerminalNode

	// IsStmt_typed_param_listContext differentiates from other interfaces.
	IsStmt_typed_param_listContext()
}

type Stmt_typed_param_listContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyStmt_typed_param_listContext() *Stmt_typed_param_listContext {
	var p = new(Stmt_typed_param_listContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = KuneiformParserRULE_stmt_typed_param_list
	return p
}

func InitEmptyStmt_typed_param_listContext(p *Stmt_typed_param_listContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = KuneiformParserRULE_stmt_typed_param_list
}

func (*Stmt_typed_param_listContext) IsStmt_typed_param_listContext() {}

func NewStmt_typed_param_listContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Stmt_typed_param_listContext {
	var p = new(Stmt_typed_param_listContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = KuneiformParserRULE_stmt_typed_param_list

	return p
}

func (s *Stmt_typed_param_listContext) GetParser() antlr.Parser { return s.parser }

func (s *Stmt_typed_param_listContext) AllSTMT_VAR() []antlr.TerminalNode {
	return s.GetTokens(KuneiformParserSTMT_VAR)
}

func (s *Stmt_typed_param_listContext) STMT_VAR(i int) antlr.TerminalNode {
	return s.GetToken(KuneiformParserSTMT_VAR, i)
}

func (s *Stmt_typed_param_listContext) AllStmt_type_selector() []IStmt_type_selectorContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IStmt_type_selectorContext); ok {
			len++
		}
	}

	tst := make([]IStmt_type_selectorContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IStmt_type_selectorContext); ok {
			tst[i] = t.(IStmt_type_selectorContext)
			i++
		}
	}

	return tst
}

func (s *Stmt_typed_param_listContext) Stmt_type_selector(i int) IStmt_type_selectorContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IStmt_type_selectorContext); ok {
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

	return t.(IStmt_type_selectorContext)
}

func (s *Stmt_typed_param_listContext) AllSTMT_COMMA() []antlr.TerminalNode {
	return s.GetTokens(KuneiformParserSTMT_COMMA)
}

func (s *Stmt_typed_param_listContext) STMT_COMMA(i int) antlr.TerminalNode {
	return s.GetToken(KuneiformParserSTMT_COMMA, i)
}

func (s *Stmt_typed_param_listContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Stmt_typed_param_listContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Stmt_typed_param_listContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case KuneiformParserVisitor:
		return t.VisitStmt_typed_param_list(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *KuneiformParser) Stmt_typed_param_list() (localctx IStmt_typed_param_listContext) {
	localctx = NewStmt_typed_param_listContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 32, KuneiformParserRULE_stmt_typed_param_list)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(229)
		p.Match(KuneiformParserSTMT_VAR)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(230)
		p.Stmt_type_selector()
	}
	p.SetState(236)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == KuneiformParserSTMT_COMMA {
		{
			p.SetState(231)
			p.Match(KuneiformParserSTMT_COMMA)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(232)
			p.Match(KuneiformParserSTMT_VAR)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(233)
			p.Stmt_type_selector()
		}

		p.SetState(238)
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

// IStmt_type_selectorContext is an interface to support dynamic dispatch.
type IStmt_type_selectorContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// GetType_ returns the type_ token.
	GetType_() antlr.Token

	// SetType_ sets the type_ token.
	SetType_(antlr.Token)

	// Getter signatures
	STMT_IDENTIFIER() antlr.TerminalNode
	STMT_ARRAY() antlr.TerminalNode

	// IsStmt_type_selectorContext differentiates from other interfaces.
	IsStmt_type_selectorContext()
}

type Stmt_type_selectorContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
	type_  antlr.Token
}

func NewEmptyStmt_type_selectorContext() *Stmt_type_selectorContext {
	var p = new(Stmt_type_selectorContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = KuneiformParserRULE_stmt_type_selector
	return p
}

func InitEmptyStmt_type_selectorContext(p *Stmt_type_selectorContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = KuneiformParserRULE_stmt_type_selector
}

func (*Stmt_type_selectorContext) IsStmt_type_selectorContext() {}

func NewStmt_type_selectorContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Stmt_type_selectorContext {
	var p = new(Stmt_type_selectorContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = KuneiformParserRULE_stmt_type_selector

	return p
}

func (s *Stmt_type_selectorContext) GetParser() antlr.Parser { return s.parser }

func (s *Stmt_type_selectorContext) GetType_() antlr.Token { return s.type_ }

func (s *Stmt_type_selectorContext) SetType_(v antlr.Token) { s.type_ = v }

func (s *Stmt_type_selectorContext) STMT_IDENTIFIER() antlr.TerminalNode {
	return s.GetToken(KuneiformParserSTMT_IDENTIFIER, 0)
}

func (s *Stmt_type_selectorContext) STMT_ARRAY() antlr.TerminalNode {
	return s.GetToken(KuneiformParserSTMT_ARRAY, 0)
}

func (s *Stmt_type_selectorContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Stmt_type_selectorContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Stmt_type_selectorContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case KuneiformParserVisitor:
		return t.VisitStmt_type_selector(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *KuneiformParser) Stmt_type_selector() (localctx IStmt_type_selectorContext) {
	localctx = NewStmt_type_selectorContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 34, KuneiformParserRULE_stmt_type_selector)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(239)

		var _m = p.Match(KuneiformParserSTMT_IDENTIFIER)

		localctx.(*Stmt_type_selectorContext).type_ = _m
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(241)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == KuneiformParserSTMT_ARRAY {
		{
			p.SetState(240)
			p.Match(KuneiformParserSTMT_ARRAY)
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
