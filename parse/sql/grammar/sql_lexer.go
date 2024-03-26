// Code generated from SQLLexer.g4 by ANTLR 4.13.1. DO NOT EDIT.

package grammar

import (
	"fmt"
	"github.com/antlr4-go/antlr/v4"
	"sync"
	"unicode"
)

// Suppress unused import error
var _ = fmt.Printf
var _ = sync.Once{}
var _ = unicode.IsLetter

type SQLLexer struct {
	*antlr.BaseLexer
	channelNames []string
	modeNames    []string
	// TODO: EOF string
}

var SQLLexerLexerStaticData struct {
	once                   sync.Once
	serializedATN          []int32
	ChannelNames           []string
	ModeNames              []string
	LiteralNames           []string
	SymbolicNames          []string
	RuleNames              []string
	PredictionContextCache *antlr.PredictionContextCache
	atn                    *antlr.ATN
	decisionToDFA          []*antlr.DFA
}

func sqllexerLexerInit() {
	staticData := &SQLLexerLexerStaticData
	staticData.ChannelNames = []string{
		"DEFAULT_TOKEN_CHANNEL", "HIDDEN",
	}
	staticData.ModeNames = []string{
		"DEFAULT_MODE",
	}
	staticData.LiteralNames = []string{
		"", "';'", "'.'", "'('", "')'", "','", "'='", "'*'", "'+'", "'-'", "'/'",
		"'%'", "'<'", "'<='", "'>'", "'>='", "'!='", "'<>'", "'::'", "'ADD'",
		"'ALL'", "'AND'", "'ASC'", "'AS'", "'BETWEEN'", "'BY'", "'CASE'", "'COLLATE'",
		"'COMMIT'", "'CONFLICT'", "'CREATE'", "'CROSS'", "'DEFAULT'", "'DELETE'",
		"'DESC'", "'DISTINCT'", "'DO'", "'ELSE'", "'END'", "'ESCAPE'", "'EXCEPT'",
		"'EXISTS'", "'FILTER'", "'FIRST'", "'FROM'", "'FULL'", "'GROUPS'", "'GROUP'",
		"'HAVING'", "'INNER'", "'INSERT'", "'INTERSECT'", "'INTO'", "'IN'",
		"'ISNULL'", "'IS'", "'JOIN'", "'LAST'", "'LEFT'", "'LIKE'", "'LIMIT'",
		"'NOTHING'", "'NOTNULL'", "'NOT'", "'NULLS'", "'OFFSET'", "'OF'", "'ON'",
		"'ORDER'", "'OR'", "'OUTER'", "'RAISE'", "'REPLACE'", "'RETURNING'",
		"'RIGHT'", "'SELECT'", "'SET'", "'THEN'", "'UNION'", "'UPDATE'", "'USING'",
		"'VALUES'", "'WHEN'", "'WHERE'", "'WITH'", "", "", "", "", "'null'",
	}
	staticData.SymbolicNames = []string{
		"", "SCOL", "DOT", "OPEN_PAR", "CLOSE_PAR", "COMMA", "ASSIGN", "STAR",
		"PLUS", "MINUS", "DIV", "MOD", "LT", "LT_EQ", "GT", "GT_EQ", "NOT_EQ1",
		"NOT_EQ2", "TYPE_CAST", "ADD_", "ALL_", "AND_", "ASC_", "AS_", "BETWEEN_",
		"BY_", "CASE_", "COLLATE_", "COMMIT_", "CONFLICT_", "CREATE_", "CROSS_",
		"DEFAULT_", "DELETE_", "DESC_", "DISTINCT_", "DO_", "ELSE_", "END_",
		"ESCAPE_", "EXCEPT_", "EXISTS_", "FILTER_", "FIRST_", "FROM_", "FULL_",
		"GROUPS_", "GROUP_", "HAVING_", "INNER_", "INSERT_", "INTERSECT_", "INTO_",
		"IN_", "ISNULL_", "IS_", "JOIN_", "LAST_", "LEFT_", "LIKE_", "LIMIT_",
		"NOTHING_", "NOTNULL_", "NOT_", "NULLS_", "OFFSET_", "OF_", "ON_", "ORDER_",
		"OR_", "OUTER_", "RAISE_", "REPLACE_", "RETURNING_", "RIGHT_", "SELECT_",
		"SET_", "THEN_", "UNION_", "UPDATE_", "USING_", "VALUES_", "WHEN_",
		"WHERE_", "WITH_", "BOOLEAN_LITERAL", "NUMERIC_LITERAL", "BLOB_LITERAL",
		"TEXT_LITERAL", "NULL_LITERAL", "IDENTIFIER", "BIND_PARAMETER", "SINGLE_LINE_COMMENT",
		"MULTILINE_COMMENT", "SPACES",
	}
	staticData.RuleNames = []string{
		"SCOL", "DOT", "OPEN_PAR", "CLOSE_PAR", "COMMA", "ASSIGN", "STAR", "PLUS",
		"MINUS", "DIV", "MOD", "LT", "LT_EQ", "GT", "GT_EQ", "NOT_EQ1", "NOT_EQ2",
		"TYPE_CAST", "ADD_", "ALL_", "AND_", "ASC_", "AS_", "BETWEEN_", "BY_",
		"CASE_", "COLLATE_", "COMMIT_", "CONFLICT_", "CREATE_", "CROSS_", "DEFAULT_",
		"DELETE_", "DESC_", "DISTINCT_", "DO_", "ELSE_", "END_", "ESCAPE_",
		"EXCEPT_", "EXISTS_", "FILTER_", "FIRST_", "FROM_", "FULL_", "GROUPS_",
		"GROUP_", "HAVING_", "INNER_", "INSERT_", "INTERSECT_", "INTO_", "IN_",
		"ISNULL_", "IS_", "JOIN_", "LAST_", "LEFT_", "LIKE_", "LIMIT_", "NOTHING_",
		"NOTNULL_", "NOT_", "NULLS_", "OFFSET_", "OF_", "ON_", "ORDER_", "OR_",
		"OUTER_", "RAISE_", "REPLACE_", "RETURNING_", "RIGHT_", "SELECT_", "SET_",
		"THEN_", "UNION_", "UPDATE_", "USING_", "VALUES_", "WHEN_", "WHERE_",
		"WITH_", "BOOLEAN_LITERAL", "NUMERIC_LITERAL", "BLOB_LITERAL", "TEXT_LITERAL",
		"NULL_LITERAL", "IDENTIFIER", "BIND_PARAMETER", "SINGLE_LINE_COMMENT",
		"MULTILINE_COMMENT", "SPACES",
	}
	staticData.PredictionContextCache = antlr.NewPredictionContextCache()
	staticData.serializedATN = []int32{
		4, 0, 94, 732, 6, -1, 2, 0, 7, 0, 2, 1, 7, 1, 2, 2, 7, 2, 2, 3, 7, 3, 2,
		4, 7, 4, 2, 5, 7, 5, 2, 6, 7, 6, 2, 7, 7, 7, 2, 8, 7, 8, 2, 9, 7, 9, 2,
		10, 7, 10, 2, 11, 7, 11, 2, 12, 7, 12, 2, 13, 7, 13, 2, 14, 7, 14, 2, 15,
		7, 15, 2, 16, 7, 16, 2, 17, 7, 17, 2, 18, 7, 18, 2, 19, 7, 19, 2, 20, 7,
		20, 2, 21, 7, 21, 2, 22, 7, 22, 2, 23, 7, 23, 2, 24, 7, 24, 2, 25, 7, 25,
		2, 26, 7, 26, 2, 27, 7, 27, 2, 28, 7, 28, 2, 29, 7, 29, 2, 30, 7, 30, 2,
		31, 7, 31, 2, 32, 7, 32, 2, 33, 7, 33, 2, 34, 7, 34, 2, 35, 7, 35, 2, 36,
		7, 36, 2, 37, 7, 37, 2, 38, 7, 38, 2, 39, 7, 39, 2, 40, 7, 40, 2, 41, 7,
		41, 2, 42, 7, 42, 2, 43, 7, 43, 2, 44, 7, 44, 2, 45, 7, 45, 2, 46, 7, 46,
		2, 47, 7, 47, 2, 48, 7, 48, 2, 49, 7, 49, 2, 50, 7, 50, 2, 51, 7, 51, 2,
		52, 7, 52, 2, 53, 7, 53, 2, 54, 7, 54, 2, 55, 7, 55, 2, 56, 7, 56, 2, 57,
		7, 57, 2, 58, 7, 58, 2, 59, 7, 59, 2, 60, 7, 60, 2, 61, 7, 61, 2, 62, 7,
		62, 2, 63, 7, 63, 2, 64, 7, 64, 2, 65, 7, 65, 2, 66, 7, 66, 2, 67, 7, 67,
		2, 68, 7, 68, 2, 69, 7, 69, 2, 70, 7, 70, 2, 71, 7, 71, 2, 72, 7, 72, 2,
		73, 7, 73, 2, 74, 7, 74, 2, 75, 7, 75, 2, 76, 7, 76, 2, 77, 7, 77, 2, 78,
		7, 78, 2, 79, 7, 79, 2, 80, 7, 80, 2, 81, 7, 81, 2, 82, 7, 82, 2, 83, 7,
		83, 2, 84, 7, 84, 2, 85, 7, 85, 2, 86, 7, 86, 2, 87, 7, 87, 2, 88, 7, 88,
		2, 89, 7, 89, 2, 90, 7, 90, 2, 91, 7, 91, 2, 92, 7, 92, 2, 93, 7, 93, 1,
		0, 1, 0, 1, 1, 1, 1, 1, 2, 1, 2, 1, 3, 1, 3, 1, 4, 1, 4, 1, 5, 1, 5, 1,
		6, 1, 6, 1, 7, 1, 7, 1, 8, 1, 8, 1, 9, 1, 9, 1, 10, 1, 10, 1, 11, 1, 11,
		1, 12, 1, 12, 1, 12, 1, 13, 1, 13, 1, 14, 1, 14, 1, 14, 1, 15, 1, 15, 1,
		15, 1, 16, 1, 16, 1, 16, 1, 17, 1, 17, 1, 17, 1, 18, 1, 18, 1, 18, 1, 18,
		1, 19, 1, 19, 1, 19, 1, 19, 1, 20, 1, 20, 1, 20, 1, 20, 1, 21, 1, 21, 1,
		21, 1, 21, 1, 22, 1, 22, 1, 22, 1, 23, 1, 23, 1, 23, 1, 23, 1, 23, 1, 23,
		1, 23, 1, 23, 1, 24, 1, 24, 1, 24, 1, 25, 1, 25, 1, 25, 1, 25, 1, 25, 1,
		26, 1, 26, 1, 26, 1, 26, 1, 26, 1, 26, 1, 26, 1, 26, 1, 27, 1, 27, 1, 27,
		1, 27, 1, 27, 1, 27, 1, 27, 1, 28, 1, 28, 1, 28, 1, 28, 1, 28, 1, 28, 1,
		28, 1, 28, 1, 28, 1, 29, 1, 29, 1, 29, 1, 29, 1, 29, 1, 29, 1, 29, 1, 30,
		1, 30, 1, 30, 1, 30, 1, 30, 1, 30, 1, 31, 1, 31, 1, 31, 1, 31, 1, 31, 1,
		31, 1, 31, 1, 31, 1, 32, 1, 32, 1, 32, 1, 32, 1, 32, 1, 32, 1, 32, 1, 33,
		1, 33, 1, 33, 1, 33, 1, 33, 1, 34, 1, 34, 1, 34, 1, 34, 1, 34, 1, 34, 1,
		34, 1, 34, 1, 34, 1, 35, 1, 35, 1, 35, 1, 36, 1, 36, 1, 36, 1, 36, 1, 36,
		1, 37, 1, 37, 1, 37, 1, 37, 1, 38, 1, 38, 1, 38, 1, 38, 1, 38, 1, 38, 1,
		38, 1, 39, 1, 39, 1, 39, 1, 39, 1, 39, 1, 39, 1, 39, 1, 40, 1, 40, 1, 40,
		1, 40, 1, 40, 1, 40, 1, 40, 1, 41, 1, 41, 1, 41, 1, 41, 1, 41, 1, 41, 1,
		41, 1, 42, 1, 42, 1, 42, 1, 42, 1, 42, 1, 42, 1, 43, 1, 43, 1, 43, 1, 43,
		1, 43, 1, 44, 1, 44, 1, 44, 1, 44, 1, 44, 1, 45, 1, 45, 1, 45, 1, 45, 1,
		45, 1, 45, 1, 45, 1, 46, 1, 46, 1, 46, 1, 46, 1, 46, 1, 46, 1, 47, 1, 47,
		1, 47, 1, 47, 1, 47, 1, 47, 1, 47, 1, 48, 1, 48, 1, 48, 1, 48, 1, 48, 1,
		48, 1, 49, 1, 49, 1, 49, 1, 49, 1, 49, 1, 49, 1, 49, 1, 50, 1, 50, 1, 50,
		1, 50, 1, 50, 1, 50, 1, 50, 1, 50, 1, 50, 1, 50, 1, 51, 1, 51, 1, 51, 1,
		51, 1, 51, 1, 52, 1, 52, 1, 52, 1, 53, 1, 53, 1, 53, 1, 53, 1, 53, 1, 53,
		1, 53, 1, 54, 1, 54, 1, 54, 1, 55, 1, 55, 1, 55, 1, 55, 1, 55, 1, 56, 1,
		56, 1, 56, 1, 56, 1, 56, 1, 57, 1, 57, 1, 57, 1, 57, 1, 57, 1, 58, 1, 58,
		1, 58, 1, 58, 1, 58, 1, 59, 1, 59, 1, 59, 1, 59, 1, 59, 1, 59, 1, 60, 1,
		60, 1, 60, 1, 60, 1, 60, 1, 60, 1, 60, 1, 60, 1, 61, 1, 61, 1, 61, 1, 61,
		1, 61, 1, 61, 1, 61, 1, 61, 1, 62, 1, 62, 1, 62, 1, 62, 1, 63, 1, 63, 1,
		63, 1, 63, 1, 63, 1, 63, 1, 64, 1, 64, 1, 64, 1, 64, 1, 64, 1, 64, 1, 64,
		1, 65, 1, 65, 1, 65, 1, 66, 1, 66, 1, 66, 1, 67, 1, 67, 1, 67, 1, 67, 1,
		67, 1, 67, 1, 68, 1, 68, 1, 68, 1, 69, 1, 69, 1, 69, 1, 69, 1, 69, 1, 69,
		1, 70, 1, 70, 1, 70, 1, 70, 1, 70, 1, 70, 1, 71, 1, 71, 1, 71, 1, 71, 1,
		71, 1, 71, 1, 71, 1, 71, 1, 72, 1, 72, 1, 72, 1, 72, 1, 72, 1, 72, 1, 72,
		1, 72, 1, 72, 1, 72, 1, 73, 1, 73, 1, 73, 1, 73, 1, 73, 1, 73, 1, 74, 1,
		74, 1, 74, 1, 74, 1, 74, 1, 74, 1, 74, 1, 75, 1, 75, 1, 75, 1, 75, 1, 76,
		1, 76, 1, 76, 1, 76, 1, 76, 1, 77, 1, 77, 1, 77, 1, 77, 1, 77, 1, 77, 1,
		78, 1, 78, 1, 78, 1, 78, 1, 78, 1, 78, 1, 78, 1, 79, 1, 79, 1, 79, 1, 79,
		1, 79, 1, 79, 1, 80, 1, 80, 1, 80, 1, 80, 1, 80, 1, 80, 1, 80, 1, 81, 1,
		81, 1, 81, 1, 81, 1, 81, 1, 82, 1, 82, 1, 82, 1, 82, 1, 82, 1, 82, 1, 83,
		1, 83, 1, 83, 1, 83, 1, 83, 1, 84, 1, 84, 1, 84, 1, 84, 1, 84, 1, 84, 1,
		84, 1, 84, 1, 84, 3, 84, 626, 8, 84, 1, 85, 4, 85, 629, 8, 85, 11, 85,
		12, 85, 630, 1, 86, 1, 86, 1, 86, 1, 86, 4, 86, 637, 8, 86, 11, 86, 12,
		86, 638, 1, 87, 1, 87, 1, 87, 1, 87, 5, 87, 645, 8, 87, 10, 87, 12, 87,
		648, 9, 87, 1, 87, 1, 87, 1, 88, 1, 88, 1, 88, 1, 88, 1, 88, 1, 89, 1,
		89, 1, 89, 1, 89, 5, 89, 661, 8, 89, 10, 89, 12, 89, 664, 9, 89, 1, 89,
		1, 89, 1, 89, 1, 89, 1, 89, 5, 89, 671, 8, 89, 10, 89, 12, 89, 674, 9,
		89, 1, 89, 1, 89, 1, 89, 5, 89, 679, 8, 89, 10, 89, 12, 89, 682, 9, 89,
		1, 89, 1, 89, 1, 89, 5, 89, 687, 8, 89, 10, 89, 12, 89, 690, 9, 89, 3,
		89, 692, 8, 89, 1, 90, 1, 90, 1, 90, 1, 91, 1, 91, 1, 91, 1, 91, 5, 91,
		701, 8, 91, 10, 91, 12, 91, 704, 9, 91, 1, 91, 3, 91, 707, 8, 91, 1, 91,
		1, 91, 3, 91, 711, 8, 91, 1, 91, 1, 91, 1, 92, 1, 92, 1, 92, 1, 92, 5,
		92, 719, 8, 92, 10, 92, 12, 92, 722, 9, 92, 1, 92, 1, 92, 1, 92, 1, 92,
		1, 92, 1, 93, 1, 93, 1, 93, 1, 93, 1, 720, 0, 94, 1, 1, 3, 2, 5, 3, 7,
		4, 9, 5, 11, 6, 13, 7, 15, 8, 17, 9, 19, 10, 21, 11, 23, 12, 25, 13, 27,
		14, 29, 15, 31, 16, 33, 17, 35, 18, 37, 19, 39, 20, 41, 21, 43, 22, 45,
		23, 47, 24, 49, 25, 51, 26, 53, 27, 55, 28, 57, 29, 59, 30, 61, 31, 63,
		32, 65, 33, 67, 34, 69, 35, 71, 36, 73, 37, 75, 38, 77, 39, 79, 40, 81,
		41, 83, 42, 85, 43, 87, 44, 89, 45, 91, 46, 93, 47, 95, 48, 97, 49, 99,
		50, 101, 51, 103, 52, 105, 53, 107, 54, 109, 55, 111, 56, 113, 57, 115,
		58, 117, 59, 119, 60, 121, 61, 123, 62, 125, 63, 127, 64, 129, 65, 131,
		66, 133, 67, 135, 68, 137, 69, 139, 70, 141, 71, 143, 72, 145, 73, 147,
		74, 149, 75, 151, 76, 153, 77, 155, 78, 157, 79, 159, 80, 161, 81, 163,
		82, 165, 83, 167, 84, 169, 85, 171, 86, 173, 87, 175, 88, 177, 89, 179,
		90, 181, 91, 183, 92, 185, 93, 187, 94, 1, 0, 35, 2, 0, 65, 65, 97, 97,
		2, 0, 68, 68, 100, 100, 2, 0, 76, 76, 108, 108, 2, 0, 78, 78, 110, 110,
		2, 0, 83, 83, 115, 115, 2, 0, 67, 67, 99, 99, 2, 0, 66, 66, 98, 98, 2,
		0, 69, 69, 101, 101, 2, 0, 84, 84, 116, 116, 2, 0, 87, 87, 119, 119, 2,
		0, 89, 89, 121, 121, 2, 0, 79, 79, 111, 111, 2, 0, 77, 77, 109, 109, 2,
		0, 73, 73, 105, 105, 2, 0, 70, 70, 102, 102, 2, 0, 82, 82, 114, 114, 2,
		0, 85, 85, 117, 117, 2, 0, 80, 80, 112, 112, 2, 0, 88, 88, 120, 120, 2,
		0, 71, 71, 103, 103, 2, 0, 72, 72, 104, 104, 2, 0, 86, 86, 118, 118, 2,
		0, 74, 74, 106, 106, 2, 0, 75, 75, 107, 107, 1, 0, 48, 57, 3, 0, 48, 57,
		65, 70, 97, 102, 1, 0, 39, 39, 1, 0, 34, 34, 1, 0, 96, 96, 1, 0, 93, 93,
		3, 0, 65, 90, 95, 95, 97, 122, 4, 0, 48, 57, 65, 90, 95, 95, 97, 122, 2,
		0, 36, 36, 64, 64, 2, 0, 10, 10, 13, 13, 3, 0, 9, 11, 13, 13, 32, 32, 749,
		0, 1, 1, 0, 0, 0, 0, 3, 1, 0, 0, 0, 0, 5, 1, 0, 0, 0, 0, 7, 1, 0, 0, 0,
		0, 9, 1, 0, 0, 0, 0, 11, 1, 0, 0, 0, 0, 13, 1, 0, 0, 0, 0, 15, 1, 0, 0,
		0, 0, 17, 1, 0, 0, 0, 0, 19, 1, 0, 0, 0, 0, 21, 1, 0, 0, 0, 0, 23, 1, 0,
		0, 0, 0, 25, 1, 0, 0, 0, 0, 27, 1, 0, 0, 0, 0, 29, 1, 0, 0, 0, 0, 31, 1,
		0, 0, 0, 0, 33, 1, 0, 0, 0, 0, 35, 1, 0, 0, 0, 0, 37, 1, 0, 0, 0, 0, 39,
		1, 0, 0, 0, 0, 41, 1, 0, 0, 0, 0, 43, 1, 0, 0, 0, 0, 45, 1, 0, 0, 0, 0,
		47, 1, 0, 0, 0, 0, 49, 1, 0, 0, 0, 0, 51, 1, 0, 0, 0, 0, 53, 1, 0, 0, 0,
		0, 55, 1, 0, 0, 0, 0, 57, 1, 0, 0, 0, 0, 59, 1, 0, 0, 0, 0, 61, 1, 0, 0,
		0, 0, 63, 1, 0, 0, 0, 0, 65, 1, 0, 0, 0, 0, 67, 1, 0, 0, 0, 0, 69, 1, 0,
		0, 0, 0, 71, 1, 0, 0, 0, 0, 73, 1, 0, 0, 0, 0, 75, 1, 0, 0, 0, 0, 77, 1,
		0, 0, 0, 0, 79, 1, 0, 0, 0, 0, 81, 1, 0, 0, 0, 0, 83, 1, 0, 0, 0, 0, 85,
		1, 0, 0, 0, 0, 87, 1, 0, 0, 0, 0, 89, 1, 0, 0, 0, 0, 91, 1, 0, 0, 0, 0,
		93, 1, 0, 0, 0, 0, 95, 1, 0, 0, 0, 0, 97, 1, 0, 0, 0, 0, 99, 1, 0, 0, 0,
		0, 101, 1, 0, 0, 0, 0, 103, 1, 0, 0, 0, 0, 105, 1, 0, 0, 0, 0, 107, 1,
		0, 0, 0, 0, 109, 1, 0, 0, 0, 0, 111, 1, 0, 0, 0, 0, 113, 1, 0, 0, 0, 0,
		115, 1, 0, 0, 0, 0, 117, 1, 0, 0, 0, 0, 119, 1, 0, 0, 0, 0, 121, 1, 0,
		0, 0, 0, 123, 1, 0, 0, 0, 0, 125, 1, 0, 0, 0, 0, 127, 1, 0, 0, 0, 0, 129,
		1, 0, 0, 0, 0, 131, 1, 0, 0, 0, 0, 133, 1, 0, 0, 0, 0, 135, 1, 0, 0, 0,
		0, 137, 1, 0, 0, 0, 0, 139, 1, 0, 0, 0, 0, 141, 1, 0, 0, 0, 0, 143, 1,
		0, 0, 0, 0, 145, 1, 0, 0, 0, 0, 147, 1, 0, 0, 0, 0, 149, 1, 0, 0, 0, 0,
		151, 1, 0, 0, 0, 0, 153, 1, 0, 0, 0, 0, 155, 1, 0, 0, 0, 0, 157, 1, 0,
		0, 0, 0, 159, 1, 0, 0, 0, 0, 161, 1, 0, 0, 0, 0, 163, 1, 0, 0, 0, 0, 165,
		1, 0, 0, 0, 0, 167, 1, 0, 0, 0, 0, 169, 1, 0, 0, 0, 0, 171, 1, 0, 0, 0,
		0, 173, 1, 0, 0, 0, 0, 175, 1, 0, 0, 0, 0, 177, 1, 0, 0, 0, 0, 179, 1,
		0, 0, 0, 0, 181, 1, 0, 0, 0, 0, 183, 1, 0, 0, 0, 0, 185, 1, 0, 0, 0, 0,
		187, 1, 0, 0, 0, 1, 189, 1, 0, 0, 0, 3, 191, 1, 0, 0, 0, 5, 193, 1, 0,
		0, 0, 7, 195, 1, 0, 0, 0, 9, 197, 1, 0, 0, 0, 11, 199, 1, 0, 0, 0, 13,
		201, 1, 0, 0, 0, 15, 203, 1, 0, 0, 0, 17, 205, 1, 0, 0, 0, 19, 207, 1,
		0, 0, 0, 21, 209, 1, 0, 0, 0, 23, 211, 1, 0, 0, 0, 25, 213, 1, 0, 0, 0,
		27, 216, 1, 0, 0, 0, 29, 218, 1, 0, 0, 0, 31, 221, 1, 0, 0, 0, 33, 224,
		1, 0, 0, 0, 35, 227, 1, 0, 0, 0, 37, 230, 1, 0, 0, 0, 39, 234, 1, 0, 0,
		0, 41, 238, 1, 0, 0, 0, 43, 242, 1, 0, 0, 0, 45, 246, 1, 0, 0, 0, 47, 249,
		1, 0, 0, 0, 49, 257, 1, 0, 0, 0, 51, 260, 1, 0, 0, 0, 53, 265, 1, 0, 0,
		0, 55, 273, 1, 0, 0, 0, 57, 280, 1, 0, 0, 0, 59, 289, 1, 0, 0, 0, 61, 296,
		1, 0, 0, 0, 63, 302, 1, 0, 0, 0, 65, 310, 1, 0, 0, 0, 67, 317, 1, 0, 0,
		0, 69, 322, 1, 0, 0, 0, 71, 331, 1, 0, 0, 0, 73, 334, 1, 0, 0, 0, 75, 339,
		1, 0, 0, 0, 77, 343, 1, 0, 0, 0, 79, 350, 1, 0, 0, 0, 81, 357, 1, 0, 0,
		0, 83, 364, 1, 0, 0, 0, 85, 371, 1, 0, 0, 0, 87, 377, 1, 0, 0, 0, 89, 382,
		1, 0, 0, 0, 91, 387, 1, 0, 0, 0, 93, 394, 1, 0, 0, 0, 95, 400, 1, 0, 0,
		0, 97, 407, 1, 0, 0, 0, 99, 413, 1, 0, 0, 0, 101, 420, 1, 0, 0, 0, 103,
		430, 1, 0, 0, 0, 105, 435, 1, 0, 0, 0, 107, 438, 1, 0, 0, 0, 109, 445,
		1, 0, 0, 0, 111, 448, 1, 0, 0, 0, 113, 453, 1, 0, 0, 0, 115, 458, 1, 0,
		0, 0, 117, 463, 1, 0, 0, 0, 119, 468, 1, 0, 0, 0, 121, 474, 1, 0, 0, 0,
		123, 482, 1, 0, 0, 0, 125, 490, 1, 0, 0, 0, 127, 494, 1, 0, 0, 0, 129,
		500, 1, 0, 0, 0, 131, 507, 1, 0, 0, 0, 133, 510, 1, 0, 0, 0, 135, 513,
		1, 0, 0, 0, 137, 519, 1, 0, 0, 0, 139, 522, 1, 0, 0, 0, 141, 528, 1, 0,
		0, 0, 143, 534, 1, 0, 0, 0, 145, 542, 1, 0, 0, 0, 147, 552, 1, 0, 0, 0,
		149, 558, 1, 0, 0, 0, 151, 565, 1, 0, 0, 0, 153, 569, 1, 0, 0, 0, 155,
		574, 1, 0, 0, 0, 157, 580, 1, 0, 0, 0, 159, 587, 1, 0, 0, 0, 161, 593,
		1, 0, 0, 0, 163, 600, 1, 0, 0, 0, 165, 605, 1, 0, 0, 0, 167, 611, 1, 0,
		0, 0, 169, 625, 1, 0, 0, 0, 171, 628, 1, 0, 0, 0, 173, 632, 1, 0, 0, 0,
		175, 640, 1, 0, 0, 0, 177, 651, 1, 0, 0, 0, 179, 691, 1, 0, 0, 0, 181,
		693, 1, 0, 0, 0, 183, 696, 1, 0, 0, 0, 185, 714, 1, 0, 0, 0, 187, 728,
		1, 0, 0, 0, 189, 190, 5, 59, 0, 0, 190, 2, 1, 0, 0, 0, 191, 192, 5, 46,
		0, 0, 192, 4, 1, 0, 0, 0, 193, 194, 5, 40, 0, 0, 194, 6, 1, 0, 0, 0, 195,
		196, 5, 41, 0, 0, 196, 8, 1, 0, 0, 0, 197, 198, 5, 44, 0, 0, 198, 10, 1,
		0, 0, 0, 199, 200, 5, 61, 0, 0, 200, 12, 1, 0, 0, 0, 201, 202, 5, 42, 0,
		0, 202, 14, 1, 0, 0, 0, 203, 204, 5, 43, 0, 0, 204, 16, 1, 0, 0, 0, 205,
		206, 5, 45, 0, 0, 206, 18, 1, 0, 0, 0, 207, 208, 5, 47, 0, 0, 208, 20,
		1, 0, 0, 0, 209, 210, 5, 37, 0, 0, 210, 22, 1, 0, 0, 0, 211, 212, 5, 60,
		0, 0, 212, 24, 1, 0, 0, 0, 213, 214, 5, 60, 0, 0, 214, 215, 5, 61, 0, 0,
		215, 26, 1, 0, 0, 0, 216, 217, 5, 62, 0, 0, 217, 28, 1, 0, 0, 0, 218, 219,
		5, 62, 0, 0, 219, 220, 5, 61, 0, 0, 220, 30, 1, 0, 0, 0, 221, 222, 5, 33,
		0, 0, 222, 223, 5, 61, 0, 0, 223, 32, 1, 0, 0, 0, 224, 225, 5, 60, 0, 0,
		225, 226, 5, 62, 0, 0, 226, 34, 1, 0, 0, 0, 227, 228, 5, 58, 0, 0, 228,
		229, 5, 58, 0, 0, 229, 36, 1, 0, 0, 0, 230, 231, 7, 0, 0, 0, 231, 232,
		7, 1, 0, 0, 232, 233, 7, 1, 0, 0, 233, 38, 1, 0, 0, 0, 234, 235, 7, 0,
		0, 0, 235, 236, 7, 2, 0, 0, 236, 237, 7, 2, 0, 0, 237, 40, 1, 0, 0, 0,
		238, 239, 7, 0, 0, 0, 239, 240, 7, 3, 0, 0, 240, 241, 7, 1, 0, 0, 241,
		42, 1, 0, 0, 0, 242, 243, 7, 0, 0, 0, 243, 244, 7, 4, 0, 0, 244, 245, 7,
		5, 0, 0, 245, 44, 1, 0, 0, 0, 246, 247, 7, 0, 0, 0, 247, 248, 7, 4, 0,
		0, 248, 46, 1, 0, 0, 0, 249, 250, 7, 6, 0, 0, 250, 251, 7, 7, 0, 0, 251,
		252, 7, 8, 0, 0, 252, 253, 7, 9, 0, 0, 253, 254, 7, 7, 0, 0, 254, 255,
		7, 7, 0, 0, 255, 256, 7, 3, 0, 0, 256, 48, 1, 0, 0, 0, 257, 258, 7, 6,
		0, 0, 258, 259, 7, 10, 0, 0, 259, 50, 1, 0, 0, 0, 260, 261, 7, 5, 0, 0,
		261, 262, 7, 0, 0, 0, 262, 263, 7, 4, 0, 0, 263, 264, 7, 7, 0, 0, 264,
		52, 1, 0, 0, 0, 265, 266, 7, 5, 0, 0, 266, 267, 7, 11, 0, 0, 267, 268,
		7, 2, 0, 0, 268, 269, 7, 2, 0, 0, 269, 270, 7, 0, 0, 0, 270, 271, 7, 8,
		0, 0, 271, 272, 7, 7, 0, 0, 272, 54, 1, 0, 0, 0, 273, 274, 7, 5, 0, 0,
		274, 275, 7, 11, 0, 0, 275, 276, 7, 12, 0, 0, 276, 277, 7, 12, 0, 0, 277,
		278, 7, 13, 0, 0, 278, 279, 7, 8, 0, 0, 279, 56, 1, 0, 0, 0, 280, 281,
		7, 5, 0, 0, 281, 282, 7, 11, 0, 0, 282, 283, 7, 3, 0, 0, 283, 284, 7, 14,
		0, 0, 284, 285, 7, 2, 0, 0, 285, 286, 7, 13, 0, 0, 286, 287, 7, 5, 0, 0,
		287, 288, 7, 8, 0, 0, 288, 58, 1, 0, 0, 0, 289, 290, 7, 5, 0, 0, 290, 291,
		7, 15, 0, 0, 291, 292, 7, 7, 0, 0, 292, 293, 7, 0, 0, 0, 293, 294, 7, 8,
		0, 0, 294, 295, 7, 7, 0, 0, 295, 60, 1, 0, 0, 0, 296, 297, 7, 5, 0, 0,
		297, 298, 7, 15, 0, 0, 298, 299, 7, 11, 0, 0, 299, 300, 7, 4, 0, 0, 300,
		301, 7, 4, 0, 0, 301, 62, 1, 0, 0, 0, 302, 303, 7, 1, 0, 0, 303, 304, 7,
		7, 0, 0, 304, 305, 7, 14, 0, 0, 305, 306, 7, 0, 0, 0, 306, 307, 7, 16,
		0, 0, 307, 308, 7, 2, 0, 0, 308, 309, 7, 8, 0, 0, 309, 64, 1, 0, 0, 0,
		310, 311, 7, 1, 0, 0, 311, 312, 7, 7, 0, 0, 312, 313, 7, 2, 0, 0, 313,
		314, 7, 7, 0, 0, 314, 315, 7, 8, 0, 0, 315, 316, 7, 7, 0, 0, 316, 66, 1,
		0, 0, 0, 317, 318, 7, 1, 0, 0, 318, 319, 7, 7, 0, 0, 319, 320, 7, 4, 0,
		0, 320, 321, 7, 5, 0, 0, 321, 68, 1, 0, 0, 0, 322, 323, 7, 1, 0, 0, 323,
		324, 7, 13, 0, 0, 324, 325, 7, 4, 0, 0, 325, 326, 7, 8, 0, 0, 326, 327,
		7, 13, 0, 0, 327, 328, 7, 3, 0, 0, 328, 329, 7, 5, 0, 0, 329, 330, 7, 8,
		0, 0, 330, 70, 1, 0, 0, 0, 331, 332, 7, 1, 0, 0, 332, 333, 7, 11, 0, 0,
		333, 72, 1, 0, 0, 0, 334, 335, 7, 7, 0, 0, 335, 336, 7, 2, 0, 0, 336, 337,
		7, 4, 0, 0, 337, 338, 7, 7, 0, 0, 338, 74, 1, 0, 0, 0, 339, 340, 7, 7,
		0, 0, 340, 341, 7, 3, 0, 0, 341, 342, 7, 1, 0, 0, 342, 76, 1, 0, 0, 0,
		343, 344, 7, 7, 0, 0, 344, 345, 7, 4, 0, 0, 345, 346, 7, 5, 0, 0, 346,
		347, 7, 0, 0, 0, 347, 348, 7, 17, 0, 0, 348, 349, 7, 7, 0, 0, 349, 78,
		1, 0, 0, 0, 350, 351, 7, 7, 0, 0, 351, 352, 7, 18, 0, 0, 352, 353, 7, 5,
		0, 0, 353, 354, 7, 7, 0, 0, 354, 355, 7, 17, 0, 0, 355, 356, 7, 8, 0, 0,
		356, 80, 1, 0, 0, 0, 357, 358, 7, 7, 0, 0, 358, 359, 7, 18, 0, 0, 359,
		360, 7, 13, 0, 0, 360, 361, 7, 4, 0, 0, 361, 362, 7, 8, 0, 0, 362, 363,
		7, 4, 0, 0, 363, 82, 1, 0, 0, 0, 364, 365, 7, 14, 0, 0, 365, 366, 7, 13,
		0, 0, 366, 367, 7, 2, 0, 0, 367, 368, 7, 8, 0, 0, 368, 369, 7, 7, 0, 0,
		369, 370, 7, 15, 0, 0, 370, 84, 1, 0, 0, 0, 371, 372, 7, 14, 0, 0, 372,
		373, 7, 13, 0, 0, 373, 374, 7, 15, 0, 0, 374, 375, 7, 4, 0, 0, 375, 376,
		7, 8, 0, 0, 376, 86, 1, 0, 0, 0, 377, 378, 7, 14, 0, 0, 378, 379, 7, 15,
		0, 0, 379, 380, 7, 11, 0, 0, 380, 381, 7, 12, 0, 0, 381, 88, 1, 0, 0, 0,
		382, 383, 7, 14, 0, 0, 383, 384, 7, 16, 0, 0, 384, 385, 7, 2, 0, 0, 385,
		386, 7, 2, 0, 0, 386, 90, 1, 0, 0, 0, 387, 388, 7, 19, 0, 0, 388, 389,
		7, 15, 0, 0, 389, 390, 7, 11, 0, 0, 390, 391, 7, 16, 0, 0, 391, 392, 7,
		17, 0, 0, 392, 393, 7, 4, 0, 0, 393, 92, 1, 0, 0, 0, 394, 395, 7, 19, 0,
		0, 395, 396, 7, 15, 0, 0, 396, 397, 7, 11, 0, 0, 397, 398, 7, 16, 0, 0,
		398, 399, 7, 17, 0, 0, 399, 94, 1, 0, 0, 0, 400, 401, 7, 20, 0, 0, 401,
		402, 7, 0, 0, 0, 402, 403, 7, 21, 0, 0, 403, 404, 7, 13, 0, 0, 404, 405,
		7, 3, 0, 0, 405, 406, 7, 19, 0, 0, 406, 96, 1, 0, 0, 0, 407, 408, 7, 13,
		0, 0, 408, 409, 7, 3, 0, 0, 409, 410, 7, 3, 0, 0, 410, 411, 7, 7, 0, 0,
		411, 412, 7, 15, 0, 0, 412, 98, 1, 0, 0, 0, 413, 414, 7, 13, 0, 0, 414,
		415, 7, 3, 0, 0, 415, 416, 7, 4, 0, 0, 416, 417, 7, 7, 0, 0, 417, 418,
		7, 15, 0, 0, 418, 419, 7, 8, 0, 0, 419, 100, 1, 0, 0, 0, 420, 421, 7, 13,
		0, 0, 421, 422, 7, 3, 0, 0, 422, 423, 7, 8, 0, 0, 423, 424, 7, 7, 0, 0,
		424, 425, 7, 15, 0, 0, 425, 426, 7, 4, 0, 0, 426, 427, 7, 7, 0, 0, 427,
		428, 7, 5, 0, 0, 428, 429, 7, 8, 0, 0, 429, 102, 1, 0, 0, 0, 430, 431,
		7, 13, 0, 0, 431, 432, 7, 3, 0, 0, 432, 433, 7, 8, 0, 0, 433, 434, 7, 11,
		0, 0, 434, 104, 1, 0, 0, 0, 435, 436, 7, 13, 0, 0, 436, 437, 7, 3, 0, 0,
		437, 106, 1, 0, 0, 0, 438, 439, 7, 13, 0, 0, 439, 440, 7, 4, 0, 0, 440,
		441, 7, 3, 0, 0, 441, 442, 7, 16, 0, 0, 442, 443, 7, 2, 0, 0, 443, 444,
		7, 2, 0, 0, 444, 108, 1, 0, 0, 0, 445, 446, 7, 13, 0, 0, 446, 447, 7, 4,
		0, 0, 447, 110, 1, 0, 0, 0, 448, 449, 7, 22, 0, 0, 449, 450, 7, 11, 0,
		0, 450, 451, 7, 13, 0, 0, 451, 452, 7, 3, 0, 0, 452, 112, 1, 0, 0, 0, 453,
		454, 7, 2, 0, 0, 454, 455, 7, 0, 0, 0, 455, 456, 7, 4, 0, 0, 456, 457,
		7, 8, 0, 0, 457, 114, 1, 0, 0, 0, 458, 459, 7, 2, 0, 0, 459, 460, 7, 7,
		0, 0, 460, 461, 7, 14, 0, 0, 461, 462, 7, 8, 0, 0, 462, 116, 1, 0, 0, 0,
		463, 464, 7, 2, 0, 0, 464, 465, 7, 13, 0, 0, 465, 466, 7, 23, 0, 0, 466,
		467, 7, 7, 0, 0, 467, 118, 1, 0, 0, 0, 468, 469, 7, 2, 0, 0, 469, 470,
		7, 13, 0, 0, 470, 471, 7, 12, 0, 0, 471, 472, 7, 13, 0, 0, 472, 473, 7,
		8, 0, 0, 473, 120, 1, 0, 0, 0, 474, 475, 7, 3, 0, 0, 475, 476, 7, 11, 0,
		0, 476, 477, 7, 8, 0, 0, 477, 478, 7, 20, 0, 0, 478, 479, 7, 13, 0, 0,
		479, 480, 7, 3, 0, 0, 480, 481, 7, 19, 0, 0, 481, 122, 1, 0, 0, 0, 482,
		483, 7, 3, 0, 0, 483, 484, 7, 11, 0, 0, 484, 485, 7, 8, 0, 0, 485, 486,
		7, 3, 0, 0, 486, 487, 7, 16, 0, 0, 487, 488, 7, 2, 0, 0, 488, 489, 7, 2,
		0, 0, 489, 124, 1, 0, 0, 0, 490, 491, 7, 3, 0, 0, 491, 492, 7, 11, 0, 0,
		492, 493, 7, 8, 0, 0, 493, 126, 1, 0, 0, 0, 494, 495, 7, 3, 0, 0, 495,
		496, 7, 16, 0, 0, 496, 497, 7, 2, 0, 0, 497, 498, 7, 2, 0, 0, 498, 499,
		7, 4, 0, 0, 499, 128, 1, 0, 0, 0, 500, 501, 7, 11, 0, 0, 501, 502, 7, 14,
		0, 0, 502, 503, 7, 14, 0, 0, 503, 504, 7, 4, 0, 0, 504, 505, 7, 7, 0, 0,
		505, 506, 7, 8, 0, 0, 506, 130, 1, 0, 0, 0, 507, 508, 7, 11, 0, 0, 508,
		509, 7, 14, 0, 0, 509, 132, 1, 0, 0, 0, 510, 511, 7, 11, 0, 0, 511, 512,
		7, 3, 0, 0, 512, 134, 1, 0, 0, 0, 513, 514, 7, 11, 0, 0, 514, 515, 7, 15,
		0, 0, 515, 516, 7, 1, 0, 0, 516, 517, 7, 7, 0, 0, 517, 518, 7, 15, 0, 0,
		518, 136, 1, 0, 0, 0, 519, 520, 7, 11, 0, 0, 520, 521, 7, 15, 0, 0, 521,
		138, 1, 0, 0, 0, 522, 523, 7, 11, 0, 0, 523, 524, 7, 16, 0, 0, 524, 525,
		7, 8, 0, 0, 525, 526, 7, 7, 0, 0, 526, 527, 7, 15, 0, 0, 527, 140, 1, 0,
		0, 0, 528, 529, 7, 15, 0, 0, 529, 530, 7, 0, 0, 0, 530, 531, 7, 13, 0,
		0, 531, 532, 7, 4, 0, 0, 532, 533, 7, 7, 0, 0, 533, 142, 1, 0, 0, 0, 534,
		535, 7, 15, 0, 0, 535, 536, 7, 7, 0, 0, 536, 537, 7, 17, 0, 0, 537, 538,
		7, 2, 0, 0, 538, 539, 7, 0, 0, 0, 539, 540, 7, 5, 0, 0, 540, 541, 7, 7,
		0, 0, 541, 144, 1, 0, 0, 0, 542, 543, 7, 15, 0, 0, 543, 544, 7, 7, 0, 0,
		544, 545, 7, 8, 0, 0, 545, 546, 7, 16, 0, 0, 546, 547, 7, 15, 0, 0, 547,
		548, 7, 3, 0, 0, 548, 549, 7, 13, 0, 0, 549, 550, 7, 3, 0, 0, 550, 551,
		7, 19, 0, 0, 551, 146, 1, 0, 0, 0, 552, 553, 7, 15, 0, 0, 553, 554, 7,
		13, 0, 0, 554, 555, 7, 19, 0, 0, 555, 556, 7, 20, 0, 0, 556, 557, 7, 8,
		0, 0, 557, 148, 1, 0, 0, 0, 558, 559, 7, 4, 0, 0, 559, 560, 7, 7, 0, 0,
		560, 561, 7, 2, 0, 0, 561, 562, 7, 7, 0, 0, 562, 563, 7, 5, 0, 0, 563,
		564, 7, 8, 0, 0, 564, 150, 1, 0, 0, 0, 565, 566, 7, 4, 0, 0, 566, 567,
		7, 7, 0, 0, 567, 568, 7, 8, 0, 0, 568, 152, 1, 0, 0, 0, 569, 570, 7, 8,
		0, 0, 570, 571, 7, 20, 0, 0, 571, 572, 7, 7, 0, 0, 572, 573, 7, 3, 0, 0,
		573, 154, 1, 0, 0, 0, 574, 575, 7, 16, 0, 0, 575, 576, 7, 3, 0, 0, 576,
		577, 7, 13, 0, 0, 577, 578, 7, 11, 0, 0, 578, 579, 7, 3, 0, 0, 579, 156,
		1, 0, 0, 0, 580, 581, 7, 16, 0, 0, 581, 582, 7, 17, 0, 0, 582, 583, 7,
		1, 0, 0, 583, 584, 7, 0, 0, 0, 584, 585, 7, 8, 0, 0, 585, 586, 7, 7, 0,
		0, 586, 158, 1, 0, 0, 0, 587, 588, 7, 16, 0, 0, 588, 589, 7, 4, 0, 0, 589,
		590, 7, 13, 0, 0, 590, 591, 7, 3, 0, 0, 591, 592, 7, 19, 0, 0, 592, 160,
		1, 0, 0, 0, 593, 594, 7, 21, 0, 0, 594, 595, 7, 0, 0, 0, 595, 596, 7, 2,
		0, 0, 596, 597, 7, 16, 0, 0, 597, 598, 7, 7, 0, 0, 598, 599, 7, 4, 0, 0,
		599, 162, 1, 0, 0, 0, 600, 601, 7, 9, 0, 0, 601, 602, 7, 20, 0, 0, 602,
		603, 7, 7, 0, 0, 603, 604, 7, 3, 0, 0, 604, 164, 1, 0, 0, 0, 605, 606,
		7, 9, 0, 0, 606, 607, 7, 20, 0, 0, 607, 608, 7, 7, 0, 0, 608, 609, 7, 15,
		0, 0, 609, 610, 7, 7, 0, 0, 610, 166, 1, 0, 0, 0, 611, 612, 7, 9, 0, 0,
		612, 613, 7, 13, 0, 0, 613, 614, 7, 8, 0, 0, 614, 615, 7, 20, 0, 0, 615,
		168, 1, 0, 0, 0, 616, 617, 7, 8, 0, 0, 617, 618, 7, 15, 0, 0, 618, 619,
		7, 16, 0, 0, 619, 626, 7, 7, 0, 0, 620, 621, 7, 14, 0, 0, 621, 622, 7,
		0, 0, 0, 622, 623, 7, 2, 0, 0, 623, 624, 7, 4, 0, 0, 624, 626, 7, 7, 0,
		0, 625, 616, 1, 0, 0, 0, 625, 620, 1, 0, 0, 0, 626, 170, 1, 0, 0, 0, 627,
		629, 7, 24, 0, 0, 628, 627, 1, 0, 0, 0, 629, 630, 1, 0, 0, 0, 630, 628,
		1, 0, 0, 0, 630, 631, 1, 0, 0, 0, 631, 172, 1, 0, 0, 0, 632, 633, 5, 48,
		0, 0, 633, 634, 7, 18, 0, 0, 634, 636, 1, 0, 0, 0, 635, 637, 7, 25, 0,
		0, 636, 635, 1, 0, 0, 0, 637, 638, 1, 0, 0, 0, 638, 636, 1, 0, 0, 0, 638,
		639, 1, 0, 0, 0, 639, 174, 1, 0, 0, 0, 640, 646, 5, 39, 0, 0, 641, 645,
		8, 26, 0, 0, 642, 643, 5, 39, 0, 0, 643, 645, 5, 39, 0, 0, 644, 641, 1,
		0, 0, 0, 644, 642, 1, 0, 0, 0, 645, 648, 1, 0, 0, 0, 646, 644, 1, 0, 0,
		0, 646, 647, 1, 0, 0, 0, 647, 649, 1, 0, 0, 0, 648, 646, 1, 0, 0, 0, 649,
		650, 5, 39, 0, 0, 650, 176, 1, 0, 0, 0, 651, 652, 7, 3, 0, 0, 652, 653,
		7, 16, 0, 0, 653, 654, 7, 2, 0, 0, 654, 655, 7, 2, 0, 0, 655, 178, 1, 0,
		0, 0, 656, 662, 5, 34, 0, 0, 657, 661, 8, 27, 0, 0, 658, 659, 5, 34, 0,
		0, 659, 661, 5, 34, 0, 0, 660, 657, 1, 0, 0, 0, 660, 658, 1, 0, 0, 0, 661,
		664, 1, 0, 0, 0, 662, 660, 1, 0, 0, 0, 662, 663, 1, 0, 0, 0, 663, 665,
		1, 0, 0, 0, 664, 662, 1, 0, 0, 0, 665, 692, 5, 34, 0, 0, 666, 672, 5, 96,
		0, 0, 667, 671, 8, 28, 0, 0, 668, 669, 5, 96, 0, 0, 669, 671, 5, 96, 0,
		0, 670, 667, 1, 0, 0, 0, 670, 668, 1, 0, 0, 0, 671, 674, 1, 0, 0, 0, 672,
		670, 1, 0, 0, 0, 672, 673, 1, 0, 0, 0, 673, 675, 1, 0, 0, 0, 674, 672,
		1, 0, 0, 0, 675, 692, 5, 96, 0, 0, 676, 680, 5, 91, 0, 0, 677, 679, 8,
		29, 0, 0, 678, 677, 1, 0, 0, 0, 679, 682, 1, 0, 0, 0, 680, 678, 1, 0, 0,
		0, 680, 681, 1, 0, 0, 0, 681, 683, 1, 0, 0, 0, 682, 680, 1, 0, 0, 0, 683,
		692, 5, 93, 0, 0, 684, 688, 7, 30, 0, 0, 685, 687, 7, 31, 0, 0, 686, 685,
		1, 0, 0, 0, 687, 690, 1, 0, 0, 0, 688, 686, 1, 0, 0, 0, 688, 689, 1, 0,
		0, 0, 689, 692, 1, 0, 0, 0, 690, 688, 1, 0, 0, 0, 691, 656, 1, 0, 0, 0,
		691, 666, 1, 0, 0, 0, 691, 676, 1, 0, 0, 0, 691, 684, 1, 0, 0, 0, 692,
		180, 1, 0, 0, 0, 693, 694, 7, 32, 0, 0, 694, 695, 3, 179, 89, 0, 695, 182,
		1, 0, 0, 0, 696, 697, 5, 45, 0, 0, 697, 698, 5, 45, 0, 0, 698, 702, 1,
		0, 0, 0, 699, 701, 8, 33, 0, 0, 700, 699, 1, 0, 0, 0, 701, 704, 1, 0, 0,
		0, 702, 700, 1, 0, 0, 0, 702, 703, 1, 0, 0, 0, 703, 710, 1, 0, 0, 0, 704,
		702, 1, 0, 0, 0, 705, 707, 5, 13, 0, 0, 706, 705, 1, 0, 0, 0, 706, 707,
		1, 0, 0, 0, 707, 708, 1, 0, 0, 0, 708, 711, 5, 10, 0, 0, 709, 711, 5, 0,
		0, 1, 710, 706, 1, 0, 0, 0, 710, 709, 1, 0, 0, 0, 711, 712, 1, 0, 0, 0,
		712, 713, 6, 91, 0, 0, 713, 184, 1, 0, 0, 0, 714, 715, 5, 47, 0, 0, 715,
		716, 5, 42, 0, 0, 716, 720, 1, 0, 0, 0, 717, 719, 9, 0, 0, 0, 718, 717,
		1, 0, 0, 0, 719, 722, 1, 0, 0, 0, 720, 721, 1, 0, 0, 0, 720, 718, 1, 0,
		0, 0, 721, 723, 1, 0, 0, 0, 722, 720, 1, 0, 0, 0, 723, 724, 5, 42, 0, 0,
		724, 725, 5, 47, 0, 0, 725, 726, 1, 0, 0, 0, 726, 727, 6, 92, 0, 0, 727,
		186, 1, 0, 0, 0, 728, 729, 7, 34, 0, 0, 729, 730, 1, 0, 0, 0, 730, 731,
		6, 93, 0, 0, 731, 188, 1, 0, 0, 0, 17, 0, 625, 630, 638, 644, 646, 660,
		662, 670, 672, 680, 688, 691, 702, 706, 710, 720, 1, 0, 1, 0,
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

// SQLLexerInit initializes any static state used to implement SQLLexer. By default the
// static state used to implement the lexer is lazily initialized during the first call to
// NewSQLLexer(). You can call this function if you wish to initialize the static state ahead
// of time.
func SQLLexerInit() {
	staticData := &SQLLexerLexerStaticData
	staticData.once.Do(sqllexerLexerInit)
}

// NewSQLLexer produces a new lexer instance for the optional input antlr.CharStream.
func NewSQLLexer(input antlr.CharStream) *SQLLexer {
	SQLLexerInit()
	l := new(SQLLexer)
	l.BaseLexer = antlr.NewBaseLexer(input)
	staticData := &SQLLexerLexerStaticData
	l.Interpreter = antlr.NewLexerATNSimulator(l, staticData.atn, staticData.decisionToDFA, staticData.PredictionContextCache)
	l.channelNames = staticData.ChannelNames
	l.modeNames = staticData.ModeNames
	l.RuleNames = staticData.RuleNames
	l.LiteralNames = staticData.LiteralNames
	l.SymbolicNames = staticData.SymbolicNames
	l.GrammarFileName = "SQLLexer.g4"
	// TODO: l.EOF = antlr.TokenEOF

	return l
}

// SQLLexer tokens.
const (
	SQLLexerSCOL                = 1
	SQLLexerDOT                 = 2
	SQLLexerOPEN_PAR            = 3
	SQLLexerCLOSE_PAR           = 4
	SQLLexerCOMMA               = 5
	SQLLexerASSIGN              = 6
	SQLLexerSTAR                = 7
	SQLLexerPLUS                = 8
	SQLLexerMINUS               = 9
	SQLLexerDIV                 = 10
	SQLLexerMOD                 = 11
	SQLLexerLT                  = 12
	SQLLexerLT_EQ               = 13
	SQLLexerGT                  = 14
	SQLLexerGT_EQ               = 15
	SQLLexerNOT_EQ1             = 16
	SQLLexerNOT_EQ2             = 17
	SQLLexerTYPE_CAST           = 18
	SQLLexerADD_                = 19
	SQLLexerALL_                = 20
	SQLLexerAND_                = 21
	SQLLexerASC_                = 22
	SQLLexerAS_                 = 23
	SQLLexerBETWEEN_            = 24
	SQLLexerBY_                 = 25
	SQLLexerCASE_               = 26
	SQLLexerCOLLATE_            = 27
	SQLLexerCOMMIT_             = 28
	SQLLexerCONFLICT_           = 29
	SQLLexerCREATE_             = 30
	SQLLexerCROSS_              = 31
	SQLLexerDEFAULT_            = 32
	SQLLexerDELETE_             = 33
	SQLLexerDESC_               = 34
	SQLLexerDISTINCT_           = 35
	SQLLexerDO_                 = 36
	SQLLexerELSE_               = 37
	SQLLexerEND_                = 38
	SQLLexerESCAPE_             = 39
	SQLLexerEXCEPT_             = 40
	SQLLexerEXISTS_             = 41
	SQLLexerFILTER_             = 42
	SQLLexerFIRST_              = 43
	SQLLexerFROM_               = 44
	SQLLexerFULL_               = 45
	SQLLexerGROUPS_             = 46
	SQLLexerGROUP_              = 47
	SQLLexerHAVING_             = 48
	SQLLexerINNER_              = 49
	SQLLexerINSERT_             = 50
	SQLLexerINTERSECT_          = 51
	SQLLexerINTO_               = 52
	SQLLexerIN_                 = 53
	SQLLexerISNULL_             = 54
	SQLLexerIS_                 = 55
	SQLLexerJOIN_               = 56
	SQLLexerLAST_               = 57
	SQLLexerLEFT_               = 58
	SQLLexerLIKE_               = 59
	SQLLexerLIMIT_              = 60
	SQLLexerNOTHING_            = 61
	SQLLexerNOTNULL_            = 62
	SQLLexerNOT_                = 63
	SQLLexerNULLS_              = 64
	SQLLexerOFFSET_             = 65
	SQLLexerOF_                 = 66
	SQLLexerON_                 = 67
	SQLLexerORDER_              = 68
	SQLLexerOR_                 = 69
	SQLLexerOUTER_              = 70
	SQLLexerRAISE_              = 71
	SQLLexerREPLACE_            = 72
	SQLLexerRETURNING_          = 73
	SQLLexerRIGHT_              = 74
	SQLLexerSELECT_             = 75
	SQLLexerSET_                = 76
	SQLLexerTHEN_               = 77
	SQLLexerUNION_              = 78
	SQLLexerUPDATE_             = 79
	SQLLexerUSING_              = 80
	SQLLexerVALUES_             = 81
	SQLLexerWHEN_               = 82
	SQLLexerWHERE_              = 83
	SQLLexerWITH_               = 84
	SQLLexerBOOLEAN_LITERAL     = 85
	SQLLexerNUMERIC_LITERAL     = 86
	SQLLexerBLOB_LITERAL        = 87
	SQLLexerTEXT_LITERAL        = 88
	SQLLexerNULL_LITERAL        = 89
	SQLLexerIDENTIFIER          = 90
	SQLLexerBIND_PARAMETER      = 91
	SQLLexerSINGLE_LINE_COMMENT = 92
	SQLLexerMULTILINE_COMMENT   = 93
	SQLLexerSPACES              = 94
)
