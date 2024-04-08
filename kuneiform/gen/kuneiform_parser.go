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
		"procedure_declaration", "table_return", "stmt_typed_param_list", "stmt_type_list",
		"stmt_type_selector",
	}
	staticData.PredictionContextCache = antlr.NewPredictionContextCache()
	staticData.serializedATN = []int32{
		4, 1, 67, 258, 2, 0, 7, 0, 2, 1, 7, 1, 2, 2, 7, 2, 2, 3, 7, 3, 2, 4, 7,
		4, 2, 5, 7, 5, 2, 6, 7, 6, 2, 7, 7, 7, 2, 8, 7, 8, 2, 9, 7, 9, 2, 10, 7,
		10, 2, 11, 7, 11, 2, 12, 7, 12, 2, 13, 7, 13, 2, 14, 7, 14, 2, 15, 7, 15,
		2, 16, 7, 16, 2, 17, 7, 17, 2, 18, 7, 18, 1, 0, 1, 0, 1, 0, 1, 0, 5, 0,
		43, 8, 0, 10, 0, 12, 0, 46, 9, 0, 1, 0, 1, 0, 1, 1, 5, 1, 51, 8, 1, 10,
		1, 12, 1, 54, 9, 1, 1, 1, 1, 1, 3, 1, 58, 8, 1, 1, 2, 1, 2, 1, 2, 1, 2,
		1, 3, 1, 3, 1, 3, 1, 3, 1, 3, 1, 3, 1, 3, 1, 3, 1, 3, 1, 3, 3, 3, 74, 8,
		3, 1, 3, 1, 3, 3, 3, 78, 8, 3, 1, 3, 1, 3, 1, 3, 1, 3, 1, 4, 1, 4, 1, 4,
		1, 4, 1, 4, 1, 4, 1, 4, 1, 4, 3, 4, 92, 8, 4, 5, 4, 94, 8, 4, 10, 4, 12,
		4, 97, 9, 4, 1, 4, 1, 4, 1, 5, 1, 5, 1, 5, 5, 5, 104, 8, 5, 10, 5, 12,
		5, 107, 9, 5, 1, 6, 1, 6, 1, 6, 1, 6, 1, 6, 1, 6, 1, 7, 1, 7, 1, 7, 1,
		7, 1, 7, 1, 7, 1, 7, 1, 7, 1, 7, 1, 7, 5, 7, 125, 8, 7, 10, 7, 12, 7, 128,
		9, 7, 1, 8, 1, 8, 3, 8, 132, 8, 8, 1, 8, 1, 8, 1, 9, 1, 9, 1, 9, 5, 9,
		139, 8, 9, 10, 9, 12, 9, 142, 9, 9, 1, 10, 1, 10, 1, 11, 1, 11, 1, 11,
		3, 11, 149, 8, 11, 1, 12, 1, 12, 1, 12, 1, 12, 1, 12, 1, 12, 1, 12, 1,
		12, 1, 12, 1, 12, 1, 12, 1, 12, 1, 12, 1, 12, 1, 12, 1, 12, 1, 12, 1, 12,
		1, 12, 1, 12, 1, 12, 1, 12, 1, 12, 1, 12, 3, 12, 175, 8, 12, 1, 13, 1,
		13, 1, 13, 1, 13, 1, 13, 1, 13, 5, 13, 183, 8, 13, 10, 13, 12, 13, 186,
		9, 13, 3, 13, 188, 8, 13, 1, 13, 1, 13, 4, 13, 192, 8, 13, 11, 13, 12,
		13, 193, 1, 13, 1, 13, 1, 14, 1, 14, 1, 14, 1, 14, 3, 14, 202, 8, 14, 1,
		14, 1, 14, 4, 14, 206, 8, 14, 11, 14, 12, 14, 207, 1, 14, 1, 14, 1, 14,
		3, 14, 213, 8, 14, 3, 14, 215, 8, 14, 1, 14, 1, 14, 1, 15, 1, 15, 1, 15,
		1, 15, 1, 15, 1, 15, 1, 15, 5, 15, 226, 8, 15, 10, 15, 12, 15, 229, 9,
		15, 1, 15, 1, 15, 1, 16, 1, 16, 1, 16, 1, 16, 1, 16, 5, 16, 238, 8, 16,
		10, 16, 12, 16, 241, 9, 16, 1, 17, 1, 17, 1, 17, 1, 17, 5, 17, 247, 8,
		17, 10, 17, 12, 17, 250, 9, 17, 1, 17, 1, 17, 1, 18, 1, 18, 3, 18, 256,
		8, 18, 1, 18, 0, 0, 19, 0, 2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24,
		26, 28, 30, 32, 34, 36, 0, 4, 2, 0, 22, 22, 24, 25, 1, 0, 30, 31, 1, 0,
		32, 36, 1, 0, 40, 43, 271, 0, 38, 1, 0, 0, 0, 2, 52, 1, 0, 0, 0, 4, 59,
		1, 0, 0, 0, 6, 63, 1, 0, 0, 0, 8, 83, 1, 0, 0, 0, 10, 100, 1, 0, 0, 0,
		12, 108, 1, 0, 0, 0, 14, 114, 1, 0, 0, 0, 16, 129, 1, 0, 0, 0, 18, 135,
		1, 0, 0, 0, 20, 143, 1, 0, 0, 0, 22, 145, 1, 0, 0, 0, 24, 174, 1, 0, 0,
		0, 26, 176, 1, 0, 0, 0, 28, 197, 1, 0, 0, 0, 30, 218, 1, 0, 0, 0, 32, 232,
		1, 0, 0, 0, 34, 242, 1, 0, 0, 0, 36, 253, 1, 0, 0, 0, 38, 44, 3, 4, 2,
		0, 39, 43, 3, 6, 3, 0, 40, 43, 3, 8, 4, 0, 41, 43, 3, 2, 1, 0, 42, 39,
		1, 0, 0, 0, 42, 40, 1, 0, 0, 0, 42, 41, 1, 0, 0, 0, 43, 46, 1, 0, 0, 0,
		44, 42, 1, 0, 0, 0, 44, 45, 1, 0, 0, 0, 45, 47, 1, 0, 0, 0, 46, 44, 1,
		0, 0, 0, 47, 48, 5, 0, 0, 1, 48, 1, 1, 0, 0, 0, 49, 51, 5, 47, 0, 0, 50,
		49, 1, 0, 0, 0, 51, 54, 1, 0, 0, 0, 52, 50, 1, 0, 0, 0, 52, 53, 1, 0, 0,
		0, 53, 57, 1, 0, 0, 0, 54, 52, 1, 0, 0, 0, 55, 58, 3, 26, 13, 0, 56, 58,
		3, 28, 14, 0, 57, 55, 1, 0, 0, 0, 57, 56, 1, 0, 0, 0, 58, 3, 1, 0, 0, 0,
		59, 60, 5, 13, 0, 0, 60, 61, 5, 46, 0, 0, 61, 62, 5, 6, 0, 0, 62, 5, 1,
		0, 0, 0, 63, 64, 5, 14, 0, 0, 64, 77, 5, 46, 0, 0, 65, 66, 5, 1, 0, 0,
		66, 67, 5, 46, 0, 0, 67, 68, 5, 5, 0, 0, 68, 73, 3, 20, 10, 0, 69, 70,
		5, 9, 0, 0, 70, 71, 5, 46, 0, 0, 71, 72, 5, 5, 0, 0, 72, 74, 3, 20, 10,
		0, 73, 69, 1, 0, 0, 0, 73, 74, 1, 0, 0, 0, 74, 75, 1, 0, 0, 0, 75, 76,
		5, 2, 0, 0, 76, 78, 1, 0, 0, 0, 77, 65, 1, 0, 0, 0, 77, 78, 1, 0, 0, 0,
		78, 79, 1, 0, 0, 0, 79, 80, 5, 16, 0, 0, 80, 81, 5, 46, 0, 0, 81, 82, 5,
		6, 0, 0, 82, 7, 1, 0, 0, 0, 83, 84, 5, 26, 0, 0, 84, 85, 5, 46, 0, 0, 85,
		86, 5, 1, 0, 0, 86, 95, 3, 10, 5, 0, 87, 91, 5, 9, 0, 0, 88, 92, 3, 10,
		5, 0, 89, 92, 3, 12, 6, 0, 90, 92, 3, 14, 7, 0, 91, 88, 1, 0, 0, 0, 91,
		89, 1, 0, 0, 0, 91, 90, 1, 0, 0, 0, 92, 94, 1, 0, 0, 0, 93, 87, 1, 0, 0,
		0, 94, 97, 1, 0, 0, 0, 95, 93, 1, 0, 0, 0, 95, 96, 1, 0, 0, 0, 96, 98,
		1, 0, 0, 0, 97, 95, 1, 0, 0, 0, 98, 99, 5, 2, 0, 0, 99, 9, 1, 0, 0, 0,
		100, 101, 5, 46, 0, 0, 101, 105, 3, 22, 11, 0, 102, 104, 3, 24, 12, 0,
		103, 102, 1, 0, 0, 0, 104, 107, 1, 0, 0, 0, 105, 103, 1, 0, 0, 0, 105,
		106, 1, 0, 0, 0, 106, 11, 1, 0, 0, 0, 107, 105, 1, 0, 0, 0, 108, 109, 5,
		45, 0, 0, 109, 110, 7, 0, 0, 0, 110, 111, 5, 7, 0, 0, 111, 112, 3, 18,
		9, 0, 112, 113, 5, 8, 0, 0, 113, 13, 1, 0, 0, 0, 114, 115, 5, 28, 0, 0,
		115, 116, 5, 7, 0, 0, 116, 117, 3, 18, 9, 0, 117, 118, 5, 8, 0, 0, 118,
		119, 5, 29, 0, 0, 119, 120, 5, 46, 0, 0, 120, 121, 5, 7, 0, 0, 121, 122,
		3, 18, 9, 0, 122, 126, 5, 8, 0, 0, 123, 125, 3, 16, 8, 0, 124, 123, 1,
		0, 0, 0, 125, 128, 1, 0, 0, 0, 126, 124, 1, 0, 0, 0, 126, 127, 1, 0, 0,
		0, 127, 15, 1, 0, 0, 0, 128, 126, 1, 0, 0, 0, 129, 131, 7, 1, 0, 0, 130,
		132, 5, 37, 0, 0, 131, 130, 1, 0, 0, 0, 131, 132, 1, 0, 0, 0, 132, 133,
		1, 0, 0, 0, 133, 134, 7, 2, 0, 0, 134, 17, 1, 0, 0, 0, 135, 140, 5, 46,
		0, 0, 136, 137, 5, 9, 0, 0, 137, 139, 5, 46, 0, 0, 138, 136, 1, 0, 0, 0,
		139, 142, 1, 0, 0, 0, 140, 138, 1, 0, 0, 0, 140, 141, 1, 0, 0, 0, 141,
		19, 1, 0, 0, 0, 142, 140, 1, 0, 0, 0, 143, 144, 7, 3, 0, 0, 144, 21, 1,
		0, 0, 0, 145, 148, 5, 46, 0, 0, 146, 147, 5, 3, 0, 0, 147, 149, 5, 4, 0,
		0, 148, 146, 1, 0, 0, 0, 148, 149, 1, 0, 0, 0, 149, 23, 1, 0, 0, 0, 150,
		151, 5, 17, 0, 0, 151, 152, 5, 7, 0, 0, 152, 153, 5, 40, 0, 0, 153, 175,
		5, 8, 0, 0, 154, 155, 5, 18, 0, 0, 155, 156, 5, 7, 0, 0, 156, 157, 5, 40,
		0, 0, 157, 175, 5, 8, 0, 0, 158, 159, 5, 19, 0, 0, 159, 160, 5, 7, 0, 0,
		160, 161, 5, 40, 0, 0, 161, 175, 5, 8, 0, 0, 162, 163, 5, 20, 0, 0, 163,
		164, 5, 7, 0, 0, 164, 165, 5, 40, 0, 0, 165, 175, 5, 8, 0, 0, 166, 175,
		5, 21, 0, 0, 167, 175, 5, 22, 0, 0, 168, 169, 5, 23, 0, 0, 169, 170, 5,
		7, 0, 0, 170, 171, 3, 20, 10, 0, 171, 172, 5, 8, 0, 0, 172, 175, 1, 0,
		0, 0, 173, 175, 5, 24, 0, 0, 174, 150, 1, 0, 0, 0, 174, 154, 1, 0, 0, 0,
		174, 158, 1, 0, 0, 0, 174, 162, 1, 0, 0, 0, 174, 166, 1, 0, 0, 0, 174,
		167, 1, 0, 0, 0, 174, 168, 1, 0, 0, 0, 174, 173, 1, 0, 0, 0, 175, 25, 1,
		0, 0, 0, 176, 177, 5, 38, 0, 0, 177, 178, 5, 63, 0, 0, 178, 187, 5, 54,
		0, 0, 179, 184, 5, 61, 0, 0, 180, 181, 5, 56, 0, 0, 181, 183, 5, 61, 0,
		0, 182, 180, 1, 0, 0, 0, 183, 186, 1, 0, 0, 0, 184, 182, 1, 0, 0, 0, 184,
		185, 1, 0, 0, 0, 185, 188, 1, 0, 0, 0, 186, 184, 1, 0, 0, 0, 187, 179,
		1, 0, 0, 0, 187, 188, 1, 0, 0, 0, 188, 189, 1, 0, 0, 0, 189, 191, 5, 55,
		0, 0, 190, 192, 5, 62, 0, 0, 191, 190, 1, 0, 0, 0, 192, 193, 1, 0, 0, 0,
		193, 191, 1, 0, 0, 0, 193, 194, 1, 0, 0, 0, 194, 195, 1, 0, 0, 0, 195,
		196, 5, 52, 0, 0, 196, 27, 1, 0, 0, 0, 197, 198, 5, 39, 0, 0, 198, 199,
		5, 63, 0, 0, 199, 201, 5, 54, 0, 0, 200, 202, 3, 32, 16, 0, 201, 200, 1,
		0, 0, 0, 201, 202, 1, 0, 0, 0, 202, 203, 1, 0, 0, 0, 203, 205, 5, 55, 0,
		0, 204, 206, 5, 62, 0, 0, 205, 204, 1, 0, 0, 0, 206, 207, 1, 0, 0, 0, 207,
		205, 1, 0, 0, 0, 207, 208, 1, 0, 0, 0, 208, 214, 1, 0, 0, 0, 209, 212,
		5, 58, 0, 0, 210, 213, 3, 34, 17, 0, 211, 213, 3, 30, 15, 0, 212, 210,
		1, 0, 0, 0, 212, 211, 1, 0, 0, 0, 213, 215, 1, 0, 0, 0, 214, 209, 1, 0,
		0, 0, 214, 215, 1, 0, 0, 0, 215, 216, 1, 0, 0, 0, 216, 217, 5, 52, 0, 0,
		217, 29, 1, 0, 0, 0, 218, 219, 5, 59, 0, 0, 219, 220, 5, 54, 0, 0, 220,
		221, 5, 63, 0, 0, 221, 227, 3, 36, 18, 0, 222, 223, 5, 56, 0, 0, 223, 224,
		5, 63, 0, 0, 224, 226, 3, 36, 18, 0, 225, 222, 1, 0, 0, 0, 226, 229, 1,
		0, 0, 0, 227, 225, 1, 0, 0, 0, 227, 228, 1, 0, 0, 0, 228, 230, 1, 0, 0,
		0, 229, 227, 1, 0, 0, 0, 230, 231, 5, 55, 0, 0, 231, 31, 1, 0, 0, 0, 232,
		233, 5, 61, 0, 0, 233, 239, 3, 36, 18, 0, 234, 235, 5, 56, 0, 0, 235, 236,
		5, 61, 0, 0, 236, 238, 3, 36, 18, 0, 237, 234, 1, 0, 0, 0, 238, 241, 1,
		0, 0, 0, 239, 237, 1, 0, 0, 0, 239, 240, 1, 0, 0, 0, 240, 33, 1, 0, 0,
		0, 241, 239, 1, 0, 0, 0, 242, 243, 5, 54, 0, 0, 243, 248, 3, 36, 18, 0,
		244, 245, 5, 56, 0, 0, 245, 247, 3, 36, 18, 0, 246, 244, 1, 0, 0, 0, 247,
		250, 1, 0, 0, 0, 248, 246, 1, 0, 0, 0, 248, 249, 1, 0, 0, 0, 249, 251,
		1, 0, 0, 0, 250, 248, 1, 0, 0, 0, 251, 252, 5, 55, 0, 0, 252, 35, 1, 0,
		0, 0, 253, 255, 5, 63, 0, 0, 254, 256, 5, 60, 0, 0, 255, 254, 1, 0, 0,
		0, 255, 256, 1, 0, 0, 0, 256, 37, 1, 0, 0, 0, 25, 42, 44, 52, 57, 73, 77,
		91, 95, 105, 126, 131, 140, 148, 174, 184, 187, 193, 201, 207, 212, 214,
		227, 239, 248, 255,
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
	KuneiformParserRULE_table_return          = 15
	KuneiformParserRULE_stmt_typed_param_list = 16
	KuneiformParserRULE_stmt_type_list        = 17
	KuneiformParserRULE_stmt_type_selector    = 18
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
		p.SetState(38)
		p.Database_declaration()
	}
	p.SetState(44)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for (int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&141562189201408) != 0 {
		p.SetState(42)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}

		switch p.GetTokenStream().LA(1) {
		case KuneiformParserUSE:
			{
				p.SetState(39)
				p.Use_declaration()
			}

		case KuneiformParserTABLE:
			{
				p.SetState(40)
				p.Table_declaration()
			}

		case KuneiformParserSTART_ACTION, KuneiformParserSTART_PROCEDURE, KuneiformParserANNOTATION:
			{
				p.SetState(41)
				p.Stmt_mode()
			}

		default:
			p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
			goto errorExit
		}

		p.SetState(46)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(47)
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
	p.SetState(52)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == KuneiformParserANNOTATION {
		{
			p.SetState(49)
			p.Match(KuneiformParserANNOTATION)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

		p.SetState(54)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	p.SetState(57)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case KuneiformParserSTART_ACTION:
		{
			p.SetState(55)
			p.Action_declaration()
		}

	case KuneiformParserSTART_PROCEDURE:
		{
			p.SetState(56)
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
		p.SetState(59)
		p.Match(KuneiformParserDATABASE)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(60)
		p.Match(KuneiformParserIDENTIFIER)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(61)
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
		p.SetState(63)
		p.Match(KuneiformParserUSE)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(64)

		var _m = p.Match(KuneiformParserIDENTIFIER)

		localctx.(*Use_declarationContext).extension_name = _m
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

	if _la == KuneiformParserLBRACE {
		{
			p.SetState(65)
			p.Match(KuneiformParserLBRACE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(66)
			p.Match(KuneiformParserIDENTIFIER)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(67)
			p.Match(KuneiformParserCOL)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(68)
			p.Literal()
		}
		p.SetState(73)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == KuneiformParserCOMMA {
			{
				p.SetState(69)
				p.Match(KuneiformParserCOMMA)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(70)
				p.Match(KuneiformParserIDENTIFIER)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(71)
				p.Match(KuneiformParserCOL)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(72)
				p.Literal()
			}

		}
		{
			p.SetState(75)
			p.Match(KuneiformParserRBRACE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	}
	{
		p.SetState(79)
		p.Match(KuneiformParserAS)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(80)

		var _m = p.Match(KuneiformParserIDENTIFIER)

		localctx.(*Use_declarationContext).alias = _m
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(81)
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
		p.SetState(83)
		p.Match(KuneiformParserTABLE)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(84)
		p.Match(KuneiformParserIDENTIFIER)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(85)
		p.Match(KuneiformParserLBRACE)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(86)
		p.Column_def()
	}
	p.SetState(95)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == KuneiformParserCOMMA {
		{
			p.SetState(87)
			p.Match(KuneiformParserCOMMA)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(91)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}

		switch p.GetTokenStream().LA(1) {
		case KuneiformParserIDENTIFIER:
			{
				p.SetState(88)
				p.Column_def()
			}

		case KuneiformParserINDEX_NAME:
			{
				p.SetState(89)
				p.Index_def()
			}

		case KuneiformParserFOREIGN_KEY:
			{
				p.SetState(90)
				p.Foreign_key_def()
			}

		default:
			p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
			goto errorExit
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
		p.SetState(100)

		var _m = p.Match(KuneiformParserIDENTIFIER)

		localctx.(*Column_defContext).name = _m
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(101)

		var _x = p.Type_selector()

		localctx.(*Column_defContext).type_ = _x
	}
	p.SetState(105)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for (int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&33423360) != 0 {
		{
			p.SetState(102)
			p.Constraint()
		}

		p.SetState(107)
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
		p.SetState(108)
		p.Match(KuneiformParserINDEX_NAME)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(109)
		_la = p.GetTokenStream().LA(1)

		if !((int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&54525952) != 0) {
			p.GetErrorHandler().RecoverInline(p)
		} else {
			p.GetErrorHandler().ReportMatch(p)
			p.Consume()
		}
	}
	{
		p.SetState(110)
		p.Match(KuneiformParserLPAREN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(111)

		var _x = p.Identifier_list()

		localctx.(*Index_defContext).columns = _x
	}
	{
		p.SetState(112)
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
		p.SetState(114)
		p.Match(KuneiformParserFOREIGN_KEY)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(115)
		p.Match(KuneiformParserLPAREN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(116)

		var _x = p.Identifier_list()

		localctx.(*Foreign_key_defContext).child_keys = _x
	}
	{
		p.SetState(117)
		p.Match(KuneiformParserRPAREN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(118)
		p.Match(KuneiformParserREFERENCES)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(119)

		var _m = p.Match(KuneiformParserIDENTIFIER)

		localctx.(*Foreign_key_defContext).parent_table = _m
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(120)
		p.Match(KuneiformParserLPAREN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(121)

		var _x = p.Identifier_list()

		localctx.(*Foreign_key_defContext).parent_keys = _x
	}
	{
		p.SetState(122)
		p.Match(KuneiformParserRPAREN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(126)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == KuneiformParserON_UPDATE || _la == KuneiformParserON_DELETE {
		{
			p.SetState(123)
			p.Foreign_key_action()
		}

		p.SetState(128)
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
		p.SetState(129)
		_la = p.GetTokenStream().LA(1)

		if !(_la == KuneiformParserON_UPDATE || _la == KuneiformParserON_DELETE) {
			p.GetErrorHandler().RecoverInline(p)
		} else {
			p.GetErrorHandler().ReportMatch(p)
			p.Consume()
		}
	}
	p.SetState(131)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == KuneiformParserDO {
		{
			p.SetState(130)
			p.Match(KuneiformParserDO)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	}
	{
		p.SetState(133)
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

	for _la == KuneiformParserCOMMA {
		{
			p.SetState(136)
			p.Match(KuneiformParserCOMMA)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(137)
			p.Match(KuneiformParserIDENTIFIER)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

		p.SetState(142)
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
		p.SetState(143)
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
		p.SetState(145)

		var _m = p.Match(KuneiformParserIDENTIFIER)

		localctx.(*Type_selectorContext).type_ = _m
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(148)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == KuneiformParserLBRACKET {
		{
			p.SetState(146)
			p.Match(KuneiformParserLBRACKET)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(147)
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
	p.SetState(174)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case KuneiformParserMIN:
		localctx = NewMINContext(p, localctx)
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(150)
			p.Match(KuneiformParserMIN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(151)
			p.Match(KuneiformParserLPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(152)
			p.Match(KuneiformParserNUMERIC_LITERAL)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(153)
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
			p.SetState(154)
			p.Match(KuneiformParserMAX)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(155)
			p.Match(KuneiformParserLPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(156)
			p.Match(KuneiformParserNUMERIC_LITERAL)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(157)
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
			p.SetState(158)
			p.Match(KuneiformParserMIN_LEN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(159)
			p.Match(KuneiformParserLPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(160)
			p.Match(KuneiformParserNUMERIC_LITERAL)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(161)
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
			p.SetState(162)
			p.Match(KuneiformParserMAX_LEN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(163)
			p.Match(KuneiformParserLPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(164)
			p.Match(KuneiformParserNUMERIC_LITERAL)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(165)
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
			p.SetState(166)
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
			p.SetState(167)
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
			p.SetState(168)
			p.Match(KuneiformParserDEFAULT)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(169)
			p.Match(KuneiformParserLPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(170)
			p.Literal()
		}
		{
			p.SetState(171)
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
			p.SetState(173)
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
		p.SetState(176)
		p.Match(KuneiformParserSTART_ACTION)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(177)
		p.Match(KuneiformParserSTMT_IDENTIFIER)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(178)
		p.Match(KuneiformParserSTMT_LPAREN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(187)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == KuneiformParserSTMT_VAR {
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

		for _la == KuneiformParserSTMT_COMMA {
			{
				p.SetState(180)
				p.Match(KuneiformParserSTMT_COMMA)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(181)
				p.Match(KuneiformParserSTMT_VAR)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

			p.SetState(186)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)
		}

	}
	{
		p.SetState(189)
		p.Match(KuneiformParserSTMT_RPAREN)
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

	for ok := true; ok; ok = _la == KuneiformParserSTMT_ACCESS {
		{
			p.SetState(190)
			p.Match(KuneiformParserSTMT_ACCESS)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

		p.SetState(193)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(195)
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
	Stmt_type_list() IStmt_type_listContext
	Table_return() ITable_returnContext

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

func (s *Procedure_declarationContext) Stmt_type_list() IStmt_type_listContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IStmt_type_listContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IStmt_type_listContext)
}

func (s *Procedure_declarationContext) Table_return() ITable_returnContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITable_returnContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ITable_returnContext)
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
		p.SetState(197)
		p.Match(KuneiformParserSTART_PROCEDURE)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(198)

		var _m = p.Match(KuneiformParserSTMT_IDENTIFIER)

		localctx.(*Procedure_declarationContext).procedure_name = _m
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(199)
		p.Match(KuneiformParserSTMT_LPAREN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(201)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == KuneiformParserSTMT_VAR {
		{
			p.SetState(200)
			p.Stmt_typed_param_list()
		}

	}
	{
		p.SetState(203)
		p.Match(KuneiformParserSTMT_RPAREN)
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

	for ok := true; ok; ok = _la == KuneiformParserSTMT_ACCESS {
		{
			p.SetState(204)
			p.Match(KuneiformParserSTMT_ACCESS)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

		p.SetState(207)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	p.SetState(214)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == KuneiformParserSTMT_RETURNS {
		{
			p.SetState(209)
			p.Match(KuneiformParserSTMT_RETURNS)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(212)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}

		switch p.GetTokenStream().LA(1) {
		case KuneiformParserSTMT_LPAREN:
			{
				p.SetState(210)
				p.Stmt_type_list()
			}

		case KuneiformParserSTMT_TABLE:
			{
				p.SetState(211)
				p.Table_return()
			}

		default:
			p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
			goto errorExit
		}

	}
	{
		p.SetState(216)
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

// ITable_returnContext is an interface to support dynamic dispatch.
type ITable_returnContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	STMT_TABLE() antlr.TerminalNode
	STMT_LPAREN() antlr.TerminalNode
	AllSTMT_IDENTIFIER() []antlr.TerminalNode
	STMT_IDENTIFIER(i int) antlr.TerminalNode
	AllStmt_type_selector() []IStmt_type_selectorContext
	Stmt_type_selector(i int) IStmt_type_selectorContext
	STMT_RPAREN() antlr.TerminalNode
	AllSTMT_COMMA() []antlr.TerminalNode
	STMT_COMMA(i int) antlr.TerminalNode

	// IsTable_returnContext differentiates from other interfaces.
	IsTable_returnContext()
}

type Table_returnContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyTable_returnContext() *Table_returnContext {
	var p = new(Table_returnContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = KuneiformParserRULE_table_return
	return p
}

func InitEmptyTable_returnContext(p *Table_returnContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = KuneiformParserRULE_table_return
}

func (*Table_returnContext) IsTable_returnContext() {}

func NewTable_returnContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Table_returnContext {
	var p = new(Table_returnContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = KuneiformParserRULE_table_return

	return p
}

func (s *Table_returnContext) GetParser() antlr.Parser { return s.parser }

func (s *Table_returnContext) STMT_TABLE() antlr.TerminalNode {
	return s.GetToken(KuneiformParserSTMT_TABLE, 0)
}

func (s *Table_returnContext) STMT_LPAREN() antlr.TerminalNode {
	return s.GetToken(KuneiformParserSTMT_LPAREN, 0)
}

func (s *Table_returnContext) AllSTMT_IDENTIFIER() []antlr.TerminalNode {
	return s.GetTokens(KuneiformParserSTMT_IDENTIFIER)
}

func (s *Table_returnContext) STMT_IDENTIFIER(i int) antlr.TerminalNode {
	return s.GetToken(KuneiformParserSTMT_IDENTIFIER, i)
}

func (s *Table_returnContext) AllStmt_type_selector() []IStmt_type_selectorContext {
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

func (s *Table_returnContext) Stmt_type_selector(i int) IStmt_type_selectorContext {
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

func (s *Table_returnContext) STMT_RPAREN() antlr.TerminalNode {
	return s.GetToken(KuneiformParserSTMT_RPAREN, 0)
}

func (s *Table_returnContext) AllSTMT_COMMA() []antlr.TerminalNode {
	return s.GetTokens(KuneiformParserSTMT_COMMA)
}

func (s *Table_returnContext) STMT_COMMA(i int) antlr.TerminalNode {
	return s.GetToken(KuneiformParserSTMT_COMMA, i)
}

func (s *Table_returnContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Table_returnContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Table_returnContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case KuneiformParserVisitor:
		return t.VisitTable_return(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *KuneiformParser) Table_return() (localctx ITable_returnContext) {
	localctx = NewTable_returnContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 30, KuneiformParserRULE_table_return)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(218)
		p.Match(KuneiformParserSTMT_TABLE)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(219)
		p.Match(KuneiformParserSTMT_LPAREN)
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
	p.SetState(227)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == KuneiformParserSTMT_COMMA {
		{
			p.SetState(222)
			p.Match(KuneiformParserSTMT_COMMA)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(223)
			p.Match(KuneiformParserSTMT_IDENTIFIER)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(224)
			p.Stmt_type_selector()
		}

		p.SetState(229)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(230)
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
	p.SetState(239)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == KuneiformParserSTMT_COMMA {
		{
			p.SetState(234)
			p.Match(KuneiformParserSTMT_COMMA)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(235)
			p.Match(KuneiformParserSTMT_VAR)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(236)
			p.Stmt_type_selector()
		}

		p.SetState(241)
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

// IStmt_type_listContext is an interface to support dynamic dispatch.
type IStmt_type_listContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	STMT_LPAREN() antlr.TerminalNode
	AllStmt_type_selector() []IStmt_type_selectorContext
	Stmt_type_selector(i int) IStmt_type_selectorContext
	STMT_RPAREN() antlr.TerminalNode
	AllSTMT_COMMA() []antlr.TerminalNode
	STMT_COMMA(i int) antlr.TerminalNode

	// IsStmt_type_listContext differentiates from other interfaces.
	IsStmt_type_listContext()
}

type Stmt_type_listContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyStmt_type_listContext() *Stmt_type_listContext {
	var p = new(Stmt_type_listContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = KuneiformParserRULE_stmt_type_list
	return p
}

func InitEmptyStmt_type_listContext(p *Stmt_type_listContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = KuneiformParserRULE_stmt_type_list
}

func (*Stmt_type_listContext) IsStmt_type_listContext() {}

func NewStmt_type_listContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Stmt_type_listContext {
	var p = new(Stmt_type_listContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = KuneiformParserRULE_stmt_type_list

	return p
}

func (s *Stmt_type_listContext) GetParser() antlr.Parser { return s.parser }

func (s *Stmt_type_listContext) STMT_LPAREN() antlr.TerminalNode {
	return s.GetToken(KuneiformParserSTMT_LPAREN, 0)
}

func (s *Stmt_type_listContext) AllStmt_type_selector() []IStmt_type_selectorContext {
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

func (s *Stmt_type_listContext) Stmt_type_selector(i int) IStmt_type_selectorContext {
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

func (s *Stmt_type_listContext) STMT_RPAREN() antlr.TerminalNode {
	return s.GetToken(KuneiformParserSTMT_RPAREN, 0)
}

func (s *Stmt_type_listContext) AllSTMT_COMMA() []antlr.TerminalNode {
	return s.GetTokens(KuneiformParserSTMT_COMMA)
}

func (s *Stmt_type_listContext) STMT_COMMA(i int) antlr.TerminalNode {
	return s.GetToken(KuneiformParserSTMT_COMMA, i)
}

func (s *Stmt_type_listContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Stmt_type_listContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Stmt_type_listContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case KuneiformParserVisitor:
		return t.VisitStmt_type_list(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *KuneiformParser) Stmt_type_list() (localctx IStmt_type_listContext) {
	localctx = NewStmt_type_listContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 34, KuneiformParserRULE_stmt_type_list)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(242)
		p.Match(KuneiformParserSTMT_LPAREN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(243)
		p.Stmt_type_selector()
	}
	p.SetState(248)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == KuneiformParserSTMT_COMMA {
		{
			p.SetState(244)
			p.Match(KuneiformParserSTMT_COMMA)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(245)
			p.Stmt_type_selector()
		}

		p.SetState(250)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(251)
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
	p.EnterRule(localctx, 36, KuneiformParserRULE_stmt_type_selector)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(253)

		var _m = p.Match(KuneiformParserSTMT_IDENTIFIER)

		localctx.(*Stmt_type_selectorContext).type_ = _m
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(255)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == KuneiformParserSTMT_ARRAY {
		{
			p.SetState(254)
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
