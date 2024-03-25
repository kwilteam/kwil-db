// Code generated from SQLParser.g4 by ANTLR 4.13.1. DO NOT EDIT.

package grammar // SQLParser
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

type SQLParser struct {
	*antlr.BaseParser
}

var SQLParserParserStaticData struct {
	once                   sync.Once
	serializedATN          []int32
	LiteralNames           []string
	SymbolicNames          []string
	RuleNames              []string
	PredictionContextCache *antlr.PredictionContextCache
	atn                    *antlr.ATN
	decisionToDFA          []*antlr.DFA
}

func sqlparserParserInit() {
	staticData := &SQLParserParserStaticData
	staticData.LiteralNames = []string{
		"", "';'", "'.'", "'('", "')'", "','", "'='", "'*'", "'+'", "'-'", "'/'",
		"'%'", "'<'", "'<='", "'>'", "'>='", "'!='", "'<>'", "'::'", "'ADD'",
		"'ALL'", "'AND'", "'ASC'", "'AS'", "'BETWEEN'", "'BY'", "'CASE'", "'COLLATE'",
		"'COMMIT'", "'CONFLICT'", "'CREATE'", "'CROSS'", "'DEFAULT'", "'DELETE'",
		"'DESC'", "'DISTINCT'", "'DO'", "'ELSE'", "'END'", "'ESCAPE'", "'EXCEPT'",
		"'EXISTS'", "'FALSE'", "'FILTER'", "'FIRST'", "'FROM'", "'FULL'", "'GROUPS'",
		"'GROUP'", "'HAVING'", "'INNER'", "'INSERT'", "'INTERSECT'", "'INTO'",
		"'IN'", "'ISNULL'", "'IS'", "'JOIN'", "'LAST'", "'LEFT'", "'LIKE'",
		"'LIMIT'", "'NOTHING'", "'NOTNULL'", "'NOT'", "'NULLS'", "'NULL'", "'OFFSET'",
		"'OF'", "'ON'", "'ORDER'", "'OR'", "'OUTER'", "'RAISE'", "'REPLACE'",
		"'RETURNING'", "'RIGHT'", "'SELECT'", "'SET'", "'THEN'", "'TRUE'", "'UNION'",
		"'UPDATE'", "'USING'", "'VALUES'", "'WHEN'", "'WHERE'", "'WITH'",
	}
	staticData.SymbolicNames = []string{
		"", "SCOL", "DOT", "OPEN_PAR", "CLOSE_PAR", "COMMA", "ASSIGN", "STAR",
		"PLUS", "MINUS", "DIV", "MOD", "LT", "LT_EQ", "GT", "GT_EQ", "NOT_EQ1",
		"NOT_EQ2", "TYPE_CAST", "ADD_", "ALL_", "AND_", "ASC_", "AS_", "BETWEEN_",
		"BY_", "CASE_", "COLLATE_", "COMMIT_", "CONFLICT_", "CREATE_", "CROSS_",
		"DEFAULT_", "DELETE_", "DESC_", "DISTINCT_", "DO_", "ELSE_", "END_",
		"ESCAPE_", "EXCEPT_", "EXISTS_", "FALSE_", "FILTER_", "FIRST_", "FROM_",
		"FULL_", "GROUPS_", "GROUP_", "HAVING_", "INNER_", "INSERT_", "INTERSECT_",
		"INTO_", "IN_", "ISNULL_", "IS_", "JOIN_", "LAST_", "LEFT_", "LIKE_",
		"LIMIT_", "NOTHING_", "NOTNULL_", "NOT_", "NULLS_", "NULL_", "OFFSET_",
		"OF_", "ON_", "ORDER_", "OR_", "OUTER_", "RAISE_", "REPLACE_", "RETURNING_",
		"RIGHT_", "SELECT_", "SET_", "THEN_", "TRUE_", "UNION_", "UPDATE_",
		"USING_", "VALUES_", "WHEN_", "WHERE_", "WITH_", "IDENTIFIER", "NUMERIC_LITERAL",
		"BIND_PARAMETER", "STRING_LITERAL", "SINGLE_LINE_COMMENT", "MULTILINE_COMMENT",
		"SPACES", "UNEXPECTED_CHAR",
	}
	staticData.RuleNames = []string{
		"statements", "sql_stmt_list", "sql_stmt", "indexed_column", "cte_table_name",
		"common_table_expression", "common_table_stmt", "delete_core", "delete_stmt",
		"variable", "function_call", "column_ref", "when_clause", "expr", "subquery",
		"expr_list", "comparisonOperator", "cast_type", "type_cast", "boolean_value",
		"string_value", "numeric_value", "literal", "value_row", "values_clause",
		"insert_core", "insert_stmt", "returning_clause", "upsert_update", "upsert_clause",
		"select_stmt_no_cte", "select_stmt", "join_relation", "relation", "select_core",
		"table_or_subquery", "result_column", "returning_clause_result_column",
		"join_operator", "join_constraint", "compound_operator", "update_set_subclause",
		"update_core", "update_stmt", "column_name_list", "qualified_table_name",
		"order_by_stmt", "limit_stmt", "ordering_term", "asc_desc", "function_keyword",
		"function_name", "table_name", "table_alias", "column_name", "column_alias",
		"collation_name", "index_name",
	}
	staticData.PredictionContextCache = antlr.NewPredictionContextCache()
	staticData.serializedATN = []int32{
		4, 1, 95, 713, 2, 0, 7, 0, 2, 1, 7, 1, 2, 2, 7, 2, 2, 3, 7, 3, 2, 4, 7,
		4, 2, 5, 7, 5, 2, 6, 7, 6, 2, 7, 7, 7, 2, 8, 7, 8, 2, 9, 7, 9, 2, 10, 7,
		10, 2, 11, 7, 11, 2, 12, 7, 12, 2, 13, 7, 13, 2, 14, 7, 14, 2, 15, 7, 15,
		2, 16, 7, 16, 2, 17, 7, 17, 2, 18, 7, 18, 2, 19, 7, 19, 2, 20, 7, 20, 2,
		21, 7, 21, 2, 22, 7, 22, 2, 23, 7, 23, 2, 24, 7, 24, 2, 25, 7, 25, 2, 26,
		7, 26, 2, 27, 7, 27, 2, 28, 7, 28, 2, 29, 7, 29, 2, 30, 7, 30, 2, 31, 7,
		31, 2, 32, 7, 32, 2, 33, 7, 33, 2, 34, 7, 34, 2, 35, 7, 35, 2, 36, 7, 36,
		2, 37, 7, 37, 2, 38, 7, 38, 2, 39, 7, 39, 2, 40, 7, 40, 2, 41, 7, 41, 2,
		42, 7, 42, 2, 43, 7, 43, 2, 44, 7, 44, 2, 45, 7, 45, 2, 46, 7, 46, 2, 47,
		7, 47, 2, 48, 7, 48, 2, 49, 7, 49, 2, 50, 7, 50, 2, 51, 7, 51, 2, 52, 7,
		52, 2, 53, 7, 53, 2, 54, 7, 54, 2, 55, 7, 55, 2, 56, 7, 56, 2, 57, 7, 57,
		1, 0, 5, 0, 118, 8, 0, 10, 0, 12, 0, 121, 9, 0, 1, 0, 1, 0, 1, 1, 5, 1,
		126, 8, 1, 10, 1, 12, 1, 129, 9, 1, 1, 1, 1, 1, 4, 1, 133, 8, 1, 11, 1,
		12, 1, 134, 1, 1, 5, 1, 138, 8, 1, 10, 1, 12, 1, 141, 9, 1, 1, 1, 5, 1,
		144, 8, 1, 10, 1, 12, 1, 147, 9, 1, 1, 2, 1, 2, 1, 2, 1, 2, 3, 2, 153,
		8, 2, 1, 3, 1, 3, 1, 4, 1, 4, 1, 4, 1, 4, 1, 4, 5, 4, 162, 8, 4, 10, 4,
		12, 4, 165, 9, 4, 1, 4, 1, 4, 3, 4, 169, 8, 4, 1, 5, 1, 5, 1, 5, 1, 5,
		1, 5, 1, 5, 1, 6, 1, 6, 1, 6, 1, 6, 5, 6, 181, 8, 6, 10, 6, 12, 6, 184,
		9, 6, 1, 7, 1, 7, 1, 7, 1, 7, 1, 7, 3, 7, 191, 8, 7, 1, 7, 3, 7, 194, 8,
		7, 1, 8, 3, 8, 197, 8, 8, 1, 8, 1, 8, 1, 9, 1, 9, 1, 10, 1, 10, 1, 10,
		3, 10, 206, 8, 10, 1, 10, 1, 10, 1, 10, 5, 10, 211, 8, 10, 10, 10, 12,
		10, 214, 9, 10, 1, 10, 3, 10, 217, 8, 10, 1, 10, 1, 10, 1, 11, 1, 11, 1,
		11, 3, 11, 224, 8, 11, 1, 11, 1, 11, 1, 12, 1, 12, 1, 12, 1, 12, 1, 12,
		1, 13, 1, 13, 1, 13, 3, 13, 236, 8, 13, 1, 13, 1, 13, 3, 13, 240, 8, 13,
		1, 13, 1, 13, 3, 13, 244, 8, 13, 1, 13, 1, 13, 1, 13, 1, 13, 1, 13, 1,
		13, 3, 13, 252, 8, 13, 1, 13, 3, 13, 255, 8, 13, 1, 13, 3, 13, 258, 8,
		13, 1, 13, 1, 13, 1, 13, 3, 13, 263, 8, 13, 1, 13, 4, 13, 266, 8, 13, 11,
		13, 12, 13, 267, 1, 13, 1, 13, 3, 13, 272, 8, 13, 1, 13, 1, 13, 1, 13,
		1, 13, 1, 13, 1, 13, 1, 13, 1, 13, 3, 13, 282, 8, 13, 1, 13, 1, 13, 3,
		13, 286, 8, 13, 1, 13, 1, 13, 1, 13, 1, 13, 1, 13, 1, 13, 1, 13, 1, 13,
		3, 13, 296, 8, 13, 1, 13, 1, 13, 1, 13, 1, 13, 1, 13, 1, 13, 1, 13, 1,
		13, 1, 13, 1, 13, 1, 13, 1, 13, 1, 13, 1, 13, 1, 13, 1, 13, 1, 13, 1, 13,
		1, 13, 1, 13, 3, 13, 318, 8, 13, 1, 13, 1, 13, 1, 13, 1, 13, 3, 13, 324,
		8, 13, 1, 13, 1, 13, 1, 13, 1, 13, 1, 13, 1, 13, 1, 13, 3, 13, 333, 8,
		13, 1, 13, 1, 13, 1, 13, 1, 13, 3, 13, 339, 8, 13, 1, 13, 1, 13, 1, 13,
		3, 13, 344, 8, 13, 1, 13, 1, 13, 1, 13, 1, 13, 1, 13, 3, 13, 351, 8, 13,
		1, 13, 1, 13, 5, 13, 355, 8, 13, 10, 13, 12, 13, 358, 9, 13, 1, 14, 1,
		14, 1, 14, 1, 14, 1, 15, 1, 15, 1, 15, 5, 15, 367, 8, 15, 10, 15, 12, 15,
		370, 9, 15, 1, 16, 1, 16, 1, 17, 1, 17, 1, 18, 1, 18, 1, 18, 1, 19, 1,
		19, 1, 20, 1, 20, 1, 21, 1, 21, 1, 22, 1, 22, 1, 22, 1, 22, 3, 22, 389,
		8, 22, 1, 23, 1, 23, 1, 23, 1, 23, 5, 23, 395, 8, 23, 10, 23, 12, 23, 398,
		9, 23, 1, 23, 1, 23, 1, 24, 1, 24, 1, 24, 1, 24, 5, 24, 406, 8, 24, 10,
		24, 12, 24, 409, 9, 24, 1, 25, 1, 25, 1, 25, 1, 25, 1, 25, 3, 25, 416,
		8, 25, 1, 25, 1, 25, 1, 25, 1, 25, 5, 25, 422, 8, 25, 10, 25, 12, 25, 425,
		9, 25, 1, 25, 1, 25, 3, 25, 429, 8, 25, 1, 25, 1, 25, 3, 25, 433, 8, 25,
		1, 25, 3, 25, 436, 8, 25, 1, 26, 3, 26, 439, 8, 26, 1, 26, 1, 26, 1, 27,
		1, 27, 1, 27, 1, 27, 5, 27, 447, 8, 27, 10, 27, 12, 27, 450, 9, 27, 1,
		28, 1, 28, 3, 28, 454, 8, 28, 1, 28, 1, 28, 1, 28, 1, 29, 1, 29, 1, 29,
		1, 29, 1, 29, 1, 29, 5, 29, 465, 8, 29, 10, 29, 12, 29, 468, 9, 29, 1,
		29, 1, 29, 1, 29, 3, 29, 473, 8, 29, 3, 29, 475, 8, 29, 1, 29, 1, 29, 1,
		29, 1, 29, 1, 29, 1, 29, 1, 29, 5, 29, 484, 8, 29, 10, 29, 12, 29, 487,
		9, 29, 1, 29, 1, 29, 3, 29, 491, 8, 29, 3, 29, 493, 8, 29, 1, 30, 1, 30,
		1, 30, 1, 30, 5, 30, 499, 8, 30, 10, 30, 12, 30, 502, 9, 30, 1, 30, 3,
		30, 505, 8, 30, 1, 30, 3, 30, 508, 8, 30, 1, 31, 3, 31, 511, 8, 31, 1,
		31, 1, 31, 1, 32, 1, 32, 1, 32, 1, 32, 1, 33, 1, 33, 5, 33, 521, 8, 33,
		10, 33, 12, 33, 524, 9, 33, 1, 34, 1, 34, 3, 34, 528, 8, 34, 1, 34, 1,
		34, 1, 34, 5, 34, 533, 8, 34, 10, 34, 12, 34, 536, 9, 34, 1, 34, 1, 34,
		3, 34, 540, 8, 34, 1, 34, 1, 34, 3, 34, 544, 8, 34, 1, 34, 1, 34, 1, 34,
		1, 34, 1, 34, 5, 34, 551, 8, 34, 10, 34, 12, 34, 554, 9, 34, 1, 34, 1,
		34, 3, 34, 558, 8, 34, 3, 34, 560, 8, 34, 1, 35, 1, 35, 1, 35, 3, 35, 565,
		8, 35, 1, 35, 1, 35, 1, 35, 1, 35, 1, 35, 3, 35, 572, 8, 35, 3, 35, 574,
		8, 35, 1, 36, 1, 36, 1, 36, 1, 36, 1, 36, 1, 36, 1, 36, 1, 36, 3, 36, 584,
		8, 36, 3, 36, 586, 8, 36, 1, 37, 1, 37, 1, 37, 1, 37, 3, 37, 592, 8, 37,
		3, 37, 594, 8, 37, 1, 38, 1, 38, 3, 38, 598, 8, 38, 1, 38, 3, 38, 601,
		8, 38, 1, 38, 1, 38, 1, 39, 1, 39, 1, 39, 1, 40, 1, 40, 3, 40, 610, 8,
		40, 1, 40, 1, 40, 3, 40, 614, 8, 40, 1, 41, 1, 41, 3, 41, 618, 8, 41, 1,
		41, 1, 41, 1, 41, 1, 42, 1, 42, 1, 42, 1, 42, 1, 42, 1, 42, 5, 42, 629,
		8, 42, 10, 42, 12, 42, 632, 9, 42, 1, 42, 1, 42, 3, 42, 636, 8, 42, 1,
		42, 1, 42, 3, 42, 640, 8, 42, 1, 42, 3, 42, 643, 8, 42, 1, 43, 3, 43, 646,
		8, 43, 1, 43, 1, 43, 1, 44, 1, 44, 1, 44, 1, 44, 5, 44, 654, 8, 44, 10,
		44, 12, 44, 657, 9, 44, 1, 44, 1, 44, 1, 45, 1, 45, 1, 45, 3, 45, 664,
		8, 45, 1, 46, 1, 46, 1, 46, 1, 46, 1, 46, 5, 46, 671, 8, 46, 10, 46, 12,
		46, 674, 9, 46, 1, 47, 1, 47, 1, 47, 1, 47, 3, 47, 680, 8, 47, 1, 48, 1,
		48, 3, 48, 684, 8, 48, 1, 48, 1, 48, 3, 48, 688, 8, 48, 1, 49, 1, 49, 1,
		50, 1, 50, 1, 50, 3, 50, 695, 8, 50, 1, 51, 1, 51, 3, 51, 699, 8, 51, 1,
		52, 1, 52, 1, 53, 1, 53, 1, 54, 1, 54, 1, 55, 1, 55, 1, 56, 1, 56, 1, 57,
		1, 57, 1, 57, 0, 1, 26, 58, 0, 2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22,
		24, 26, 28, 30, 32, 34, 36, 38, 40, 42, 44, 46, 48, 50, 52, 54, 56, 58,
		60, 62, 64, 66, 68, 70, 72, 74, 76, 78, 80, 82, 84, 86, 88, 90, 92, 94,
		96, 98, 100, 102, 104, 106, 108, 110, 112, 114, 0, 8, 1, 0, 8, 9, 2, 0,
		7, 7, 10, 11, 2, 0, 55, 55, 63, 63, 2, 0, 6, 6, 12, 17, 2, 0, 42, 42, 80,
		80, 3, 0, 46, 46, 59, 59, 76, 76, 2, 0, 44, 44, 58, 58, 2, 0, 22, 22, 34,
		34, 773, 0, 119, 1, 0, 0, 0, 2, 127, 1, 0, 0, 0, 4, 152, 1, 0, 0, 0, 6,
		154, 1, 0, 0, 0, 8, 156, 1, 0, 0, 0, 10, 170, 1, 0, 0, 0, 12, 176, 1, 0,
		0, 0, 14, 185, 1, 0, 0, 0, 16, 196, 1, 0, 0, 0, 18, 200, 1, 0, 0, 0, 20,
		202, 1, 0, 0, 0, 22, 223, 1, 0, 0, 0, 24, 227, 1, 0, 0, 0, 26, 285, 1,
		0, 0, 0, 28, 359, 1, 0, 0, 0, 30, 363, 1, 0, 0, 0, 32, 371, 1, 0, 0, 0,
		34, 373, 1, 0, 0, 0, 36, 375, 1, 0, 0, 0, 38, 378, 1, 0, 0, 0, 40, 380,
		1, 0, 0, 0, 42, 382, 1, 0, 0, 0, 44, 388, 1, 0, 0, 0, 46, 390, 1, 0, 0,
		0, 48, 401, 1, 0, 0, 0, 50, 410, 1, 0, 0, 0, 52, 438, 1, 0, 0, 0, 54, 442,
		1, 0, 0, 0, 56, 453, 1, 0, 0, 0, 58, 458, 1, 0, 0, 0, 60, 494, 1, 0, 0,
		0, 62, 510, 1, 0, 0, 0, 64, 514, 1, 0, 0, 0, 66, 518, 1, 0, 0, 0, 68, 525,
		1, 0, 0, 0, 70, 573, 1, 0, 0, 0, 72, 585, 1, 0, 0, 0, 74, 593, 1, 0, 0,
		0, 76, 600, 1, 0, 0, 0, 78, 604, 1, 0, 0, 0, 80, 613, 1, 0, 0, 0, 82, 617,
		1, 0, 0, 0, 84, 622, 1, 0, 0, 0, 86, 645, 1, 0, 0, 0, 88, 649, 1, 0, 0,
		0, 90, 660, 1, 0, 0, 0, 92, 665, 1, 0, 0, 0, 94, 675, 1, 0, 0, 0, 96, 681,
		1, 0, 0, 0, 98, 689, 1, 0, 0, 0, 100, 694, 1, 0, 0, 0, 102, 698, 1, 0,
		0, 0, 104, 700, 1, 0, 0, 0, 106, 702, 1, 0, 0, 0, 108, 704, 1, 0, 0, 0,
		110, 706, 1, 0, 0, 0, 112, 708, 1, 0, 0, 0, 114, 710, 1, 0, 0, 0, 116,
		118, 3, 2, 1, 0, 117, 116, 1, 0, 0, 0, 118, 121, 1, 0, 0, 0, 119, 117,
		1, 0, 0, 0, 119, 120, 1, 0, 0, 0, 120, 122, 1, 0, 0, 0, 121, 119, 1, 0,
		0, 0, 122, 123, 5, 0, 0, 1, 123, 1, 1, 0, 0, 0, 124, 126, 5, 1, 0, 0, 125,
		124, 1, 0, 0, 0, 126, 129, 1, 0, 0, 0, 127, 125, 1, 0, 0, 0, 127, 128,
		1, 0, 0, 0, 128, 130, 1, 0, 0, 0, 129, 127, 1, 0, 0, 0, 130, 139, 3, 4,
		2, 0, 131, 133, 5, 1, 0, 0, 132, 131, 1, 0, 0, 0, 133, 134, 1, 0, 0, 0,
		134, 132, 1, 0, 0, 0, 134, 135, 1, 0, 0, 0, 135, 136, 1, 0, 0, 0, 136,
		138, 3, 4, 2, 0, 137, 132, 1, 0, 0, 0, 138, 141, 1, 0, 0, 0, 139, 137,
		1, 0, 0, 0, 139, 140, 1, 0, 0, 0, 140, 145, 1, 0, 0, 0, 141, 139, 1, 0,
		0, 0, 142, 144, 5, 1, 0, 0, 143, 142, 1, 0, 0, 0, 144, 147, 1, 0, 0, 0,
		145, 143, 1, 0, 0, 0, 145, 146, 1, 0, 0, 0, 146, 3, 1, 0, 0, 0, 147, 145,
		1, 0, 0, 0, 148, 153, 3, 16, 8, 0, 149, 153, 3, 52, 26, 0, 150, 153, 3,
		62, 31, 0, 151, 153, 3, 86, 43, 0, 152, 148, 1, 0, 0, 0, 152, 149, 1, 0,
		0, 0, 152, 150, 1, 0, 0, 0, 152, 151, 1, 0, 0, 0, 153, 5, 1, 0, 0, 0, 154,
		155, 3, 108, 54, 0, 155, 7, 1, 0, 0, 0, 156, 168, 3, 104, 52, 0, 157, 158,
		5, 3, 0, 0, 158, 163, 3, 108, 54, 0, 159, 160, 5, 5, 0, 0, 160, 162, 3,
		108, 54, 0, 161, 159, 1, 0, 0, 0, 162, 165, 1, 0, 0, 0, 163, 161, 1, 0,
		0, 0, 163, 164, 1, 0, 0, 0, 164, 166, 1, 0, 0, 0, 165, 163, 1, 0, 0, 0,
		166, 167, 5, 4, 0, 0, 167, 169, 1, 0, 0, 0, 168, 157, 1, 0, 0, 0, 168,
		169, 1, 0, 0, 0, 169, 9, 1, 0, 0, 0, 170, 171, 3, 8, 4, 0, 171, 172, 5,
		23, 0, 0, 172, 173, 5, 3, 0, 0, 173, 174, 3, 60, 30, 0, 174, 175, 5, 4,
		0, 0, 175, 11, 1, 0, 0, 0, 176, 177, 5, 87, 0, 0, 177, 182, 3, 10, 5, 0,
		178, 179, 5, 5, 0, 0, 179, 181, 3, 10, 5, 0, 180, 178, 1, 0, 0, 0, 181,
		184, 1, 0, 0, 0, 182, 180, 1, 0, 0, 0, 182, 183, 1, 0, 0, 0, 183, 13, 1,
		0, 0, 0, 184, 182, 1, 0, 0, 0, 185, 186, 5, 33, 0, 0, 186, 187, 5, 45,
		0, 0, 187, 190, 3, 90, 45, 0, 188, 189, 5, 86, 0, 0, 189, 191, 3, 26, 13,
		0, 190, 188, 1, 0, 0, 0, 190, 191, 1, 0, 0, 0, 191, 193, 1, 0, 0, 0, 192,
		194, 3, 54, 27, 0, 193, 192, 1, 0, 0, 0, 193, 194, 1, 0, 0, 0, 194, 15,
		1, 0, 0, 0, 195, 197, 3, 12, 6, 0, 196, 195, 1, 0, 0, 0, 196, 197, 1, 0,
		0, 0, 197, 198, 1, 0, 0, 0, 198, 199, 3, 14, 7, 0, 199, 17, 1, 0, 0, 0,
		200, 201, 5, 90, 0, 0, 201, 19, 1, 0, 0, 0, 202, 203, 3, 102, 51, 0, 203,
		216, 5, 3, 0, 0, 204, 206, 5, 35, 0, 0, 205, 204, 1, 0, 0, 0, 205, 206,
		1, 0, 0, 0, 206, 207, 1, 0, 0, 0, 207, 212, 3, 26, 13, 0, 208, 209, 5,
		5, 0, 0, 209, 211, 3, 26, 13, 0, 210, 208, 1, 0, 0, 0, 211, 214, 1, 0,
		0, 0, 212, 210, 1, 0, 0, 0, 212, 213, 1, 0, 0, 0, 213, 217, 1, 0, 0, 0,
		214, 212, 1, 0, 0, 0, 215, 217, 5, 7, 0, 0, 216, 205, 1, 0, 0, 0, 216,
		215, 1, 0, 0, 0, 216, 217, 1, 0, 0, 0, 217, 218, 1, 0, 0, 0, 218, 219,
		5, 4, 0, 0, 219, 21, 1, 0, 0, 0, 220, 221, 3, 104, 52, 0, 221, 222, 5,
		2, 0, 0, 222, 224, 1, 0, 0, 0, 223, 220, 1, 0, 0, 0, 223, 224, 1, 0, 0,
		0, 224, 225, 1, 0, 0, 0, 225, 226, 3, 108, 54, 0, 226, 23, 1, 0, 0, 0,
		227, 228, 5, 85, 0, 0, 228, 229, 3, 26, 13, 0, 229, 230, 5, 79, 0, 0, 230,
		231, 3, 26, 13, 0, 231, 25, 1, 0, 0, 0, 232, 233, 6, 13, -1, 0, 233, 235,
		3, 44, 22, 0, 234, 236, 3, 36, 18, 0, 235, 234, 1, 0, 0, 0, 235, 236, 1,
		0, 0, 0, 236, 286, 1, 0, 0, 0, 237, 239, 3, 18, 9, 0, 238, 240, 3, 36,
		18, 0, 239, 238, 1, 0, 0, 0, 239, 240, 1, 0, 0, 0, 240, 286, 1, 0, 0, 0,
		241, 243, 3, 22, 11, 0, 242, 244, 3, 36, 18, 0, 243, 242, 1, 0, 0, 0, 243,
		244, 1, 0, 0, 0, 244, 286, 1, 0, 0, 0, 245, 246, 7, 0, 0, 0, 246, 286,
		3, 26, 13, 19, 247, 248, 5, 3, 0, 0, 248, 249, 3, 26, 13, 0, 249, 251,
		5, 4, 0, 0, 250, 252, 3, 36, 18, 0, 251, 250, 1, 0, 0, 0, 251, 252, 1,
		0, 0, 0, 252, 286, 1, 0, 0, 0, 253, 255, 5, 64, 0, 0, 254, 253, 1, 0, 0,
		0, 254, 255, 1, 0, 0, 0, 255, 256, 1, 0, 0, 0, 256, 258, 5, 41, 0, 0, 257,
		254, 1, 0, 0, 0, 257, 258, 1, 0, 0, 0, 258, 259, 1, 0, 0, 0, 259, 286,
		3, 28, 14, 0, 260, 262, 5, 26, 0, 0, 261, 263, 3, 26, 13, 0, 262, 261,
		1, 0, 0, 0, 262, 263, 1, 0, 0, 0, 263, 265, 1, 0, 0, 0, 264, 266, 3, 24,
		12, 0, 265, 264, 1, 0, 0, 0, 266, 267, 1, 0, 0, 0, 267, 265, 1, 0, 0, 0,
		267, 268, 1, 0, 0, 0, 268, 271, 1, 0, 0, 0, 269, 270, 5, 37, 0, 0, 270,
		272, 3, 26, 13, 0, 271, 269, 1, 0, 0, 0, 271, 272, 1, 0, 0, 0, 272, 273,
		1, 0, 0, 0, 273, 274, 5, 38, 0, 0, 274, 286, 1, 0, 0, 0, 275, 276, 5, 3,
		0, 0, 276, 277, 3, 30, 15, 0, 277, 278, 5, 4, 0, 0, 278, 286, 1, 0, 0,
		0, 279, 281, 3, 20, 10, 0, 280, 282, 3, 36, 18, 0, 281, 280, 1, 0, 0, 0,
		281, 282, 1, 0, 0, 0, 282, 286, 1, 0, 0, 0, 283, 284, 5, 64, 0, 0, 284,
		286, 3, 26, 13, 3, 285, 232, 1, 0, 0, 0, 285, 237, 1, 0, 0, 0, 285, 241,
		1, 0, 0, 0, 285, 245, 1, 0, 0, 0, 285, 247, 1, 0, 0, 0, 285, 257, 1, 0,
		0, 0, 285, 260, 1, 0, 0, 0, 285, 275, 1, 0, 0, 0, 285, 279, 1, 0, 0, 0,
		285, 283, 1, 0, 0, 0, 286, 356, 1, 0, 0, 0, 287, 288, 10, 12, 0, 0, 288,
		289, 7, 1, 0, 0, 289, 355, 3, 26, 13, 13, 290, 291, 10, 11, 0, 0, 291,
		292, 7, 0, 0, 0, 292, 355, 3, 26, 13, 12, 293, 295, 10, 8, 0, 0, 294, 296,
		5, 64, 0, 0, 295, 294, 1, 0, 0, 0, 295, 296, 1, 0, 0, 0, 296, 297, 1, 0,
		0, 0, 297, 298, 5, 24, 0, 0, 298, 299, 3, 26, 13, 0, 299, 300, 5, 21, 0,
		0, 300, 301, 3, 26, 13, 9, 301, 355, 1, 0, 0, 0, 302, 303, 10, 6, 0, 0,
		303, 304, 3, 32, 16, 0, 304, 305, 3, 26, 13, 7, 305, 355, 1, 0, 0, 0, 306,
		307, 10, 2, 0, 0, 307, 308, 5, 21, 0, 0, 308, 355, 3, 26, 13, 3, 309, 310,
		10, 1, 0, 0, 310, 311, 5, 71, 0, 0, 311, 355, 3, 26, 13, 2, 312, 313, 10,
		18, 0, 0, 313, 314, 5, 27, 0, 0, 314, 355, 3, 112, 56, 0, 315, 317, 10,
		10, 0, 0, 316, 318, 5, 64, 0, 0, 317, 316, 1, 0, 0, 0, 317, 318, 1, 0,
		0, 0, 318, 319, 1, 0, 0, 0, 319, 320, 5, 54, 0, 0, 320, 355, 3, 28, 14,
		0, 321, 323, 10, 9, 0, 0, 322, 324, 5, 64, 0, 0, 323, 322, 1, 0, 0, 0,
		323, 324, 1, 0, 0, 0, 324, 325, 1, 0, 0, 0, 325, 326, 5, 54, 0, 0, 326,
		327, 5, 3, 0, 0, 327, 328, 3, 30, 15, 0, 328, 329, 5, 4, 0, 0, 329, 355,
		1, 0, 0, 0, 330, 332, 10, 7, 0, 0, 331, 333, 5, 64, 0, 0, 332, 331, 1,
		0, 0, 0, 332, 333, 1, 0, 0, 0, 333, 334, 1, 0, 0, 0, 334, 335, 5, 60, 0,
		0, 335, 338, 3, 26, 13, 0, 336, 337, 5, 39, 0, 0, 337, 339, 3, 26, 13,
		0, 338, 336, 1, 0, 0, 0, 338, 339, 1, 0, 0, 0, 339, 355, 1, 0, 0, 0, 340,
		341, 10, 5, 0, 0, 341, 343, 5, 56, 0, 0, 342, 344, 5, 64, 0, 0, 343, 342,
		1, 0, 0, 0, 343, 344, 1, 0, 0, 0, 344, 350, 1, 0, 0, 0, 345, 346, 5, 35,
		0, 0, 346, 347, 5, 45, 0, 0, 347, 351, 3, 26, 13, 0, 348, 351, 3, 38, 19,
		0, 349, 351, 5, 66, 0, 0, 350, 345, 1, 0, 0, 0, 350, 348, 1, 0, 0, 0, 350,
		349, 1, 0, 0, 0, 351, 355, 1, 0, 0, 0, 352, 353, 10, 4, 0, 0, 353, 355,
		7, 2, 0, 0, 354, 287, 1, 0, 0, 0, 354, 290, 1, 0, 0, 0, 354, 293, 1, 0,
		0, 0, 354, 302, 1, 0, 0, 0, 354, 306, 1, 0, 0, 0, 354, 309, 1, 0, 0, 0,
		354, 312, 1, 0, 0, 0, 354, 315, 1, 0, 0, 0, 354, 321, 1, 0, 0, 0, 354,
		330, 1, 0, 0, 0, 354, 340, 1, 0, 0, 0, 354, 352, 1, 0, 0, 0, 355, 358,
		1, 0, 0, 0, 356, 354, 1, 0, 0, 0, 356, 357, 1, 0, 0, 0, 357, 27, 1, 0,
		0, 0, 358, 356, 1, 0, 0, 0, 359, 360, 5, 3, 0, 0, 360, 361, 3, 60, 30,
		0, 361, 362, 5, 4, 0, 0, 362, 29, 1, 0, 0, 0, 363, 368, 3, 26, 13, 0, 364,
		365, 5, 5, 0, 0, 365, 367, 3, 26, 13, 0, 366, 364, 1, 0, 0, 0, 367, 370,
		1, 0, 0, 0, 368, 366, 1, 0, 0, 0, 368, 369, 1, 0, 0, 0, 369, 31, 1, 0,
		0, 0, 370, 368, 1, 0, 0, 0, 371, 372, 7, 3, 0, 0, 372, 33, 1, 0, 0, 0,
		373, 374, 5, 88, 0, 0, 374, 35, 1, 0, 0, 0, 375, 376, 5, 18, 0, 0, 376,
		377, 3, 34, 17, 0, 377, 37, 1, 0, 0, 0, 378, 379, 7, 4, 0, 0, 379, 39,
		1, 0, 0, 0, 380, 381, 5, 91, 0, 0, 381, 41, 1, 0, 0, 0, 382, 383, 5, 89,
		0, 0, 383, 43, 1, 0, 0, 0, 384, 389, 5, 66, 0, 0, 385, 389, 3, 38, 19,
		0, 386, 389, 3, 40, 20, 0, 387, 389, 3, 42, 21, 0, 388, 384, 1, 0, 0, 0,
		388, 385, 1, 0, 0, 0, 388, 386, 1, 0, 0, 0, 388, 387, 1, 0, 0, 0, 389,
		45, 1, 0, 0, 0, 390, 391, 5, 3, 0, 0, 391, 396, 3, 26, 13, 0, 392, 393,
		5, 5, 0, 0, 393, 395, 3, 26, 13, 0, 394, 392, 1, 0, 0, 0, 395, 398, 1,
		0, 0, 0, 396, 394, 1, 0, 0, 0, 396, 397, 1, 0, 0, 0, 397, 399, 1, 0, 0,
		0, 398, 396, 1, 0, 0, 0, 399, 400, 5, 4, 0, 0, 400, 47, 1, 0, 0, 0, 401,
		402, 5, 84, 0, 0, 402, 407, 3, 46, 23, 0, 403, 404, 5, 5, 0, 0, 404, 406,
		3, 46, 23, 0, 405, 403, 1, 0, 0, 0, 406, 409, 1, 0, 0, 0, 407, 405, 1,
		0, 0, 0, 407, 408, 1, 0, 0, 0, 408, 49, 1, 0, 0, 0, 409, 407, 1, 0, 0,
		0, 410, 411, 5, 51, 0, 0, 411, 412, 5, 53, 0, 0, 412, 415, 3, 104, 52,
		0, 413, 414, 5, 23, 0, 0, 414, 416, 3, 106, 53, 0, 415, 413, 1, 0, 0, 0,
		415, 416, 1, 0, 0, 0, 416, 428, 1, 0, 0, 0, 417, 418, 5, 3, 0, 0, 418,
		423, 3, 108, 54, 0, 419, 420, 5, 5, 0, 0, 420, 422, 3, 108, 54, 0, 421,
		419, 1, 0, 0, 0, 422, 425, 1, 0, 0, 0, 423, 421, 1, 0, 0, 0, 423, 424,
		1, 0, 0, 0, 424, 426, 1, 0, 0, 0, 425, 423, 1, 0, 0, 0, 426, 427, 5, 4,
		0, 0, 427, 429, 1, 0, 0, 0, 428, 417, 1, 0, 0, 0, 428, 429, 1, 0, 0, 0,
		429, 430, 1, 0, 0, 0, 430, 432, 3, 48, 24, 0, 431, 433, 3, 58, 29, 0, 432,
		431, 1, 0, 0, 0, 432, 433, 1, 0, 0, 0, 433, 435, 1, 0, 0, 0, 434, 436,
		3, 54, 27, 0, 435, 434, 1, 0, 0, 0, 435, 436, 1, 0, 0, 0, 436, 51, 1, 0,
		0, 0, 437, 439, 3, 12, 6, 0, 438, 437, 1, 0, 0, 0, 438, 439, 1, 0, 0, 0,
		439, 440, 1, 0, 0, 0, 440, 441, 3, 50, 25, 0, 441, 53, 1, 0, 0, 0, 442,
		443, 5, 75, 0, 0, 443, 448, 3, 74, 37, 0, 444, 445, 5, 5, 0, 0, 445, 447,
		3, 74, 37, 0, 446, 444, 1, 0, 0, 0, 447, 450, 1, 0, 0, 0, 448, 446, 1,
		0, 0, 0, 448, 449, 1, 0, 0, 0, 449, 55, 1, 0, 0, 0, 450, 448, 1, 0, 0,
		0, 451, 454, 3, 108, 54, 0, 452, 454, 3, 88, 44, 0, 453, 451, 1, 0, 0,
		0, 453, 452, 1, 0, 0, 0, 454, 455, 1, 0, 0, 0, 455, 456, 5, 6, 0, 0, 456,
		457, 3, 26, 13, 0, 457, 57, 1, 0, 0, 0, 458, 459, 5, 69, 0, 0, 459, 474,
		5, 29, 0, 0, 460, 461, 5, 3, 0, 0, 461, 466, 3, 6, 3, 0, 462, 463, 5, 5,
		0, 0, 463, 465, 3, 6, 3, 0, 464, 462, 1, 0, 0, 0, 465, 468, 1, 0, 0, 0,
		466, 464, 1, 0, 0, 0, 466, 467, 1, 0, 0, 0, 467, 469, 1, 0, 0, 0, 468,
		466, 1, 0, 0, 0, 469, 472, 5, 4, 0, 0, 470, 471, 5, 86, 0, 0, 471, 473,
		3, 26, 13, 0, 472, 470, 1, 0, 0, 0, 472, 473, 1, 0, 0, 0, 473, 475, 1,
		0, 0, 0, 474, 460, 1, 0, 0, 0, 474, 475, 1, 0, 0, 0, 475, 476, 1, 0, 0,
		0, 476, 492, 5, 36, 0, 0, 477, 493, 5, 62, 0, 0, 478, 479, 5, 82, 0, 0,
		479, 480, 5, 78, 0, 0, 480, 485, 3, 56, 28, 0, 481, 482, 5, 5, 0, 0, 482,
		484, 3, 56, 28, 0, 483, 481, 1, 0, 0, 0, 484, 487, 1, 0, 0, 0, 485, 483,
		1, 0, 0, 0, 485, 486, 1, 0, 0, 0, 486, 490, 1, 0, 0, 0, 487, 485, 1, 0,
		0, 0, 488, 489, 5, 86, 0, 0, 489, 491, 3, 26, 13, 0, 490, 488, 1, 0, 0,
		0, 490, 491, 1, 0, 0, 0, 491, 493, 1, 0, 0, 0, 492, 477, 1, 0, 0, 0, 492,
		478, 1, 0, 0, 0, 493, 59, 1, 0, 0, 0, 494, 500, 3, 68, 34, 0, 495, 496,
		3, 80, 40, 0, 496, 497, 3, 68, 34, 0, 497, 499, 1, 0, 0, 0, 498, 495, 1,
		0, 0, 0, 499, 502, 1, 0, 0, 0, 500, 498, 1, 0, 0, 0, 500, 501, 1, 0, 0,
		0, 501, 504, 1, 0, 0, 0, 502, 500, 1, 0, 0, 0, 503, 505, 3, 92, 46, 0,
		504, 503, 1, 0, 0, 0, 504, 505, 1, 0, 0, 0, 505, 507, 1, 0, 0, 0, 506,
		508, 3, 94, 47, 0, 507, 506, 1, 0, 0, 0, 507, 508, 1, 0, 0, 0, 508, 61,
		1, 0, 0, 0, 509, 511, 3, 12, 6, 0, 510, 509, 1, 0, 0, 0, 510, 511, 1, 0,
		0, 0, 511, 512, 1, 0, 0, 0, 512, 513, 3, 60, 30, 0, 513, 63, 1, 0, 0, 0,
		514, 515, 3, 76, 38, 0, 515, 516, 3, 70, 35, 0, 516, 517, 3, 78, 39, 0,
		517, 65, 1, 0, 0, 0, 518, 522, 3, 70, 35, 0, 519, 521, 3, 64, 32, 0, 520,
		519, 1, 0, 0, 0, 521, 524, 1, 0, 0, 0, 522, 520, 1, 0, 0, 0, 522, 523,
		1, 0, 0, 0, 523, 67, 1, 0, 0, 0, 524, 522, 1, 0, 0, 0, 525, 527, 5, 77,
		0, 0, 526, 528, 5, 35, 0, 0, 527, 526, 1, 0, 0, 0, 527, 528, 1, 0, 0, 0,
		528, 529, 1, 0, 0, 0, 529, 534, 3, 72, 36, 0, 530, 531, 5, 5, 0, 0, 531,
		533, 3, 72, 36, 0, 532, 530, 1, 0, 0, 0, 533, 536, 1, 0, 0, 0, 534, 532,
		1, 0, 0, 0, 534, 535, 1, 0, 0, 0, 535, 539, 1, 0, 0, 0, 536, 534, 1, 0,
		0, 0, 537, 538, 5, 45, 0, 0, 538, 540, 3, 66, 33, 0, 539, 537, 1, 0, 0,
		0, 539, 540, 1, 0, 0, 0, 540, 543, 1, 0, 0, 0, 541, 542, 5, 86, 0, 0, 542,
		544, 3, 26, 13, 0, 543, 541, 1, 0, 0, 0, 543, 544, 1, 0, 0, 0, 544, 559,
		1, 0, 0, 0, 545, 546, 5, 48, 0, 0, 546, 547, 5, 25, 0, 0, 547, 552, 3,
		26, 13, 0, 548, 549, 5, 5, 0, 0, 549, 551, 3, 26, 13, 0, 550, 548, 1, 0,
		0, 0, 551, 554, 1, 0, 0, 0, 552, 550, 1, 0, 0, 0, 552, 553, 1, 0, 0, 0,
		553, 557, 1, 0, 0, 0, 554, 552, 1, 0, 0, 0, 555, 556, 5, 49, 0, 0, 556,
		558, 3, 26, 13, 0, 557, 555, 1, 0, 0, 0, 557, 558, 1, 0, 0, 0, 558, 560,
		1, 0, 0, 0, 559, 545, 1, 0, 0, 0, 559, 560, 1, 0, 0, 0, 560, 69, 1, 0,
		0, 0, 561, 564, 3, 104, 52, 0, 562, 563, 5, 23, 0, 0, 563, 565, 3, 106,
		53, 0, 564, 562, 1, 0, 0, 0, 564, 565, 1, 0, 0, 0, 565, 574, 1, 0, 0, 0,
		566, 567, 5, 3, 0, 0, 567, 568, 3, 60, 30, 0, 568, 571, 5, 4, 0, 0, 569,
		570, 5, 23, 0, 0, 570, 572, 3, 106, 53, 0, 571, 569, 1, 0, 0, 0, 571, 572,
		1, 0, 0, 0, 572, 574, 1, 0, 0, 0, 573, 561, 1, 0, 0, 0, 573, 566, 1, 0,
		0, 0, 574, 71, 1, 0, 0, 0, 575, 586, 5, 7, 0, 0, 576, 577, 3, 104, 52,
		0, 577, 578, 5, 2, 0, 0, 578, 579, 5, 7, 0, 0, 579, 586, 1, 0, 0, 0, 580,
		583, 3, 26, 13, 0, 581, 582, 5, 23, 0, 0, 582, 584, 3, 110, 55, 0, 583,
		581, 1, 0, 0, 0, 583, 584, 1, 0, 0, 0, 584, 586, 1, 0, 0, 0, 585, 575,
		1, 0, 0, 0, 585, 576, 1, 0, 0, 0, 585, 580, 1, 0, 0, 0, 586, 73, 1, 0,
		0, 0, 587, 594, 5, 7, 0, 0, 588, 591, 3, 26, 13, 0, 589, 590, 5, 23, 0,
		0, 590, 592, 3, 110, 55, 0, 591, 589, 1, 0, 0, 0, 591, 592, 1, 0, 0, 0,
		592, 594, 1, 0, 0, 0, 593, 587, 1, 0, 0, 0, 593, 588, 1, 0, 0, 0, 594,
		75, 1, 0, 0, 0, 595, 597, 7, 5, 0, 0, 596, 598, 5, 72, 0, 0, 597, 596,
		1, 0, 0, 0, 597, 598, 1, 0, 0, 0, 598, 601, 1, 0, 0, 0, 599, 601, 5, 50,
		0, 0, 600, 595, 1, 0, 0, 0, 600, 599, 1, 0, 0, 0, 600, 601, 1, 0, 0, 0,
		601, 602, 1, 0, 0, 0, 602, 603, 5, 57, 0, 0, 603, 77, 1, 0, 0, 0, 604,
		605, 5, 69, 0, 0, 605, 606, 3, 26, 13, 0, 606, 79, 1, 0, 0, 0, 607, 609,
		5, 81, 0, 0, 608, 610, 5, 20, 0, 0, 609, 608, 1, 0, 0, 0, 609, 610, 1,
		0, 0, 0, 610, 614, 1, 0, 0, 0, 611, 614, 5, 52, 0, 0, 612, 614, 5, 40,
		0, 0, 613, 607, 1, 0, 0, 0, 613, 611, 1, 0, 0, 0, 613, 612, 1, 0, 0, 0,
		614, 81, 1, 0, 0, 0, 615, 618, 3, 108, 54, 0, 616, 618, 3, 88, 44, 0, 617,
		615, 1, 0, 0, 0, 617, 616, 1, 0, 0, 0, 618, 619, 1, 0, 0, 0, 619, 620,
		5, 6, 0, 0, 620, 621, 3, 26, 13, 0, 621, 83, 1, 0, 0, 0, 622, 623, 5, 82,
		0, 0, 623, 624, 3, 90, 45, 0, 624, 625, 5, 78, 0, 0, 625, 630, 3, 82, 41,
		0, 626, 627, 5, 5, 0, 0, 627, 629, 3, 82, 41, 0, 628, 626, 1, 0, 0, 0,
		629, 632, 1, 0, 0, 0, 630, 628, 1, 0, 0, 0, 630, 631, 1, 0, 0, 0, 631,
		635, 1, 0, 0, 0, 632, 630, 1, 0, 0, 0, 633, 634, 5, 45, 0, 0, 634, 636,
		3, 66, 33, 0, 635, 633, 1, 0, 0, 0, 635, 636, 1, 0, 0, 0, 636, 639, 1,
		0, 0, 0, 637, 638, 5, 86, 0, 0, 638, 640, 3, 26, 13, 0, 639, 637, 1, 0,
		0, 0, 639, 640, 1, 0, 0, 0, 640, 642, 1, 0, 0, 0, 641, 643, 3, 54, 27,
		0, 642, 641, 1, 0, 0, 0, 642, 643, 1, 0, 0, 0, 643, 85, 1, 0, 0, 0, 644,
		646, 3, 12, 6, 0, 645, 644, 1, 0, 0, 0, 645, 646, 1, 0, 0, 0, 646, 647,
		1, 0, 0, 0, 647, 648, 3, 84, 42, 0, 648, 87, 1, 0, 0, 0, 649, 650, 5, 3,
		0, 0, 650, 655, 3, 108, 54, 0, 651, 652, 5, 5, 0, 0, 652, 654, 3, 108,
		54, 0, 653, 651, 1, 0, 0, 0, 654, 657, 1, 0, 0, 0, 655, 653, 1, 0, 0, 0,
		655, 656, 1, 0, 0, 0, 656, 658, 1, 0, 0, 0, 657, 655, 1, 0, 0, 0, 658,
		659, 5, 4, 0, 0, 659, 89, 1, 0, 0, 0, 660, 663, 3, 104, 52, 0, 661, 662,
		5, 23, 0, 0, 662, 664, 3, 106, 53, 0, 663, 661, 1, 0, 0, 0, 663, 664, 1,
		0, 0, 0, 664, 91, 1, 0, 0, 0, 665, 666, 5, 70, 0, 0, 666, 667, 5, 25, 0,
		0, 667, 672, 3, 96, 48, 0, 668, 669, 5, 5, 0, 0, 669, 671, 3, 96, 48, 0,
		670, 668, 1, 0, 0, 0, 671, 674, 1, 0, 0, 0, 672, 670, 1, 0, 0, 0, 672,
		673, 1, 0, 0, 0, 673, 93, 1, 0, 0, 0, 674, 672, 1, 0, 0, 0, 675, 676, 5,
		61, 0, 0, 676, 679, 3, 26, 13, 0, 677, 678, 5, 67, 0, 0, 678, 680, 3, 26,
		13, 0, 679, 677, 1, 0, 0, 0, 679, 680, 1, 0, 0, 0, 680, 95, 1, 0, 0, 0,
		681, 683, 3, 26, 13, 0, 682, 684, 3, 98, 49, 0, 683, 682, 1, 0, 0, 0, 683,
		684, 1, 0, 0, 0, 684, 687, 1, 0, 0, 0, 685, 686, 5, 65, 0, 0, 686, 688,
		7, 6, 0, 0, 687, 685, 1, 0, 0, 0, 687, 688, 1, 0, 0, 0, 688, 97, 1, 0,
		0, 0, 689, 690, 7, 7, 0, 0, 690, 99, 1, 0, 0, 0, 691, 695, 1, 0, 0, 0,
		692, 695, 5, 60, 0, 0, 693, 695, 5, 74, 0, 0, 694, 691, 1, 0, 0, 0, 694,
		692, 1, 0, 0, 0, 694, 693, 1, 0, 0, 0, 695, 101, 1, 0, 0, 0, 696, 699,
		5, 88, 0, 0, 697, 699, 3, 100, 50, 0, 698, 696, 1, 0, 0, 0, 698, 697, 1,
		0, 0, 0, 699, 103, 1, 0, 0, 0, 700, 701, 5, 88, 0, 0, 701, 105, 1, 0, 0,
		0, 702, 703, 5, 88, 0, 0, 703, 107, 1, 0, 0, 0, 704, 705, 5, 88, 0, 0,
		705, 109, 1, 0, 0, 0, 706, 707, 5, 88, 0, 0, 707, 111, 1, 0, 0, 0, 708,
		709, 5, 88, 0, 0, 709, 113, 1, 0, 0, 0, 710, 711, 5, 88, 0, 0, 711, 115,
		1, 0, 0, 0, 91, 119, 127, 134, 139, 145, 152, 163, 168, 182, 190, 193,
		196, 205, 212, 216, 223, 235, 239, 243, 251, 254, 257, 262, 267, 271, 281,
		285, 295, 317, 323, 332, 338, 343, 350, 354, 356, 368, 388, 396, 407, 415,
		423, 428, 432, 435, 438, 448, 453, 466, 472, 474, 485, 490, 492, 500, 504,
		507, 510, 522, 527, 534, 539, 543, 552, 557, 559, 564, 571, 573, 583, 585,
		591, 593, 597, 600, 609, 613, 617, 630, 635, 639, 642, 645, 655, 663, 672,
		679, 683, 687, 694, 698,
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

// SQLParserInit initializes any static state used to implement SQLParser. By default the
// static state used to implement the parser is lazily initialized during the first call to
// NewSQLParser(). You can call this function if you wish to initialize the static state ahead
// of time.
func SQLParserInit() {
	staticData := &SQLParserParserStaticData
	staticData.once.Do(sqlparserParserInit)
}

// NewSQLParser produces a new parser instance for the optional input antlr.TokenStream.
func NewSQLParser(input antlr.TokenStream) *SQLParser {
	SQLParserInit()
	this := new(SQLParser)
	this.BaseParser = antlr.NewBaseParser(input)
	staticData := &SQLParserParserStaticData
	this.Interpreter = antlr.NewParserATNSimulator(this, staticData.atn, staticData.decisionToDFA, staticData.PredictionContextCache)
	this.RuleNames = staticData.RuleNames
	this.LiteralNames = staticData.LiteralNames
	this.SymbolicNames = staticData.SymbolicNames
	this.GrammarFileName = "SQLParser.g4"

	return this
}

// SQLParser tokens.
const (
	SQLParserEOF                 = antlr.TokenEOF
	SQLParserSCOL                = 1
	SQLParserDOT                 = 2
	SQLParserOPEN_PAR            = 3
	SQLParserCLOSE_PAR           = 4
	SQLParserCOMMA               = 5
	SQLParserASSIGN              = 6
	SQLParserSTAR                = 7
	SQLParserPLUS                = 8
	SQLParserMINUS               = 9
	SQLParserDIV                 = 10
	SQLParserMOD                 = 11
	SQLParserLT                  = 12
	SQLParserLT_EQ               = 13
	SQLParserGT                  = 14
	SQLParserGT_EQ               = 15
	SQLParserNOT_EQ1             = 16
	SQLParserNOT_EQ2             = 17
	SQLParserTYPE_CAST           = 18
	SQLParserADD_                = 19
	SQLParserALL_                = 20
	SQLParserAND_                = 21
	SQLParserASC_                = 22
	SQLParserAS_                 = 23
	SQLParserBETWEEN_            = 24
	SQLParserBY_                 = 25
	SQLParserCASE_               = 26
	SQLParserCOLLATE_            = 27
	SQLParserCOMMIT_             = 28
	SQLParserCONFLICT_           = 29
	SQLParserCREATE_             = 30
	SQLParserCROSS_              = 31
	SQLParserDEFAULT_            = 32
	SQLParserDELETE_             = 33
	SQLParserDESC_               = 34
	SQLParserDISTINCT_           = 35
	SQLParserDO_                 = 36
	SQLParserELSE_               = 37
	SQLParserEND_                = 38
	SQLParserESCAPE_             = 39
	SQLParserEXCEPT_             = 40
	SQLParserEXISTS_             = 41
	SQLParserFALSE_              = 42
	SQLParserFILTER_             = 43
	SQLParserFIRST_              = 44
	SQLParserFROM_               = 45
	SQLParserFULL_               = 46
	SQLParserGROUPS_             = 47
	SQLParserGROUP_              = 48
	SQLParserHAVING_             = 49
	SQLParserINNER_              = 50
	SQLParserINSERT_             = 51
	SQLParserINTERSECT_          = 52
	SQLParserINTO_               = 53
	SQLParserIN_                 = 54
	SQLParserISNULL_             = 55
	SQLParserIS_                 = 56
	SQLParserJOIN_               = 57
	SQLParserLAST_               = 58
	SQLParserLEFT_               = 59
	SQLParserLIKE_               = 60
	SQLParserLIMIT_              = 61
	SQLParserNOTHING_            = 62
	SQLParserNOTNULL_            = 63
	SQLParserNOT_                = 64
	SQLParserNULLS_              = 65
	SQLParserNULL_               = 66
	SQLParserOFFSET_             = 67
	SQLParserOF_                 = 68
	SQLParserON_                 = 69
	SQLParserORDER_              = 70
	SQLParserOR_                 = 71
	SQLParserOUTER_              = 72
	SQLParserRAISE_              = 73
	SQLParserREPLACE_            = 74
	SQLParserRETURNING_          = 75
	SQLParserRIGHT_              = 76
	SQLParserSELECT_             = 77
	SQLParserSET_                = 78
	SQLParserTHEN_               = 79
	SQLParserTRUE_               = 80
	SQLParserUNION_              = 81
	SQLParserUPDATE_             = 82
	SQLParserUSING_              = 83
	SQLParserVALUES_             = 84
	SQLParserWHEN_               = 85
	SQLParserWHERE_              = 86
	SQLParserWITH_               = 87
	SQLParserIDENTIFIER          = 88
	SQLParserNUMERIC_LITERAL     = 89
	SQLParserBIND_PARAMETER      = 90
	SQLParserSTRING_LITERAL      = 91
	SQLParserSINGLE_LINE_COMMENT = 92
	SQLParserMULTILINE_COMMENT   = 93
	SQLParserSPACES              = 94
	SQLParserUNEXPECTED_CHAR     = 95
)

// SQLParser rules.
const (
	SQLParserRULE_statements                     = 0
	SQLParserRULE_sql_stmt_list                  = 1
	SQLParserRULE_sql_stmt                       = 2
	SQLParserRULE_indexed_column                 = 3
	SQLParserRULE_cte_table_name                 = 4
	SQLParserRULE_common_table_expression        = 5
	SQLParserRULE_common_table_stmt              = 6
	SQLParserRULE_delete_core                    = 7
	SQLParserRULE_delete_stmt                    = 8
	SQLParserRULE_variable                       = 9
	SQLParserRULE_function_call                  = 10
	SQLParserRULE_column_ref                     = 11
	SQLParserRULE_when_clause                    = 12
	SQLParserRULE_expr                           = 13
	SQLParserRULE_subquery                       = 14
	SQLParserRULE_expr_list                      = 15
	SQLParserRULE_comparisonOperator             = 16
	SQLParserRULE_cast_type                      = 17
	SQLParserRULE_type_cast                      = 18
	SQLParserRULE_boolean_value                  = 19
	SQLParserRULE_string_value                   = 20
	SQLParserRULE_numeric_value                  = 21
	SQLParserRULE_literal                        = 22
	SQLParserRULE_value_row                      = 23
	SQLParserRULE_values_clause                  = 24
	SQLParserRULE_insert_core                    = 25
	SQLParserRULE_insert_stmt                    = 26
	SQLParserRULE_returning_clause               = 27
	SQLParserRULE_upsert_update                  = 28
	SQLParserRULE_upsert_clause                  = 29
	SQLParserRULE_select_stmt_no_cte             = 30
	SQLParserRULE_select_stmt                    = 31
	SQLParserRULE_join_relation                  = 32
	SQLParserRULE_relation                       = 33
	SQLParserRULE_select_core                    = 34
	SQLParserRULE_table_or_subquery              = 35
	SQLParserRULE_result_column                  = 36
	SQLParserRULE_returning_clause_result_column = 37
	SQLParserRULE_join_operator                  = 38
	SQLParserRULE_join_constraint                = 39
	SQLParserRULE_compound_operator              = 40
	SQLParserRULE_update_set_subclause           = 41
	SQLParserRULE_update_core                    = 42
	SQLParserRULE_update_stmt                    = 43
	SQLParserRULE_column_name_list               = 44
	SQLParserRULE_qualified_table_name           = 45
	SQLParserRULE_order_by_stmt                  = 46
	SQLParserRULE_limit_stmt                     = 47
	SQLParserRULE_ordering_term                  = 48
	SQLParserRULE_asc_desc                       = 49
	SQLParserRULE_function_keyword               = 50
	SQLParserRULE_function_name                  = 51
	SQLParserRULE_table_name                     = 52
	SQLParserRULE_table_alias                    = 53
	SQLParserRULE_column_name                    = 54
	SQLParserRULE_column_alias                   = 55
	SQLParserRULE_collation_name                 = 56
	SQLParserRULE_index_name                     = 57
)

// IStatementsContext is an interface to support dynamic dispatch.
type IStatementsContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	EOF() antlr.TerminalNode
	AllSql_stmt_list() []ISql_stmt_listContext
	Sql_stmt_list(i int) ISql_stmt_listContext

	// IsStatementsContext differentiates from other interfaces.
	IsStatementsContext()
}

type StatementsContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyStatementsContext() *StatementsContext {
	var p = new(StatementsContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_statements
	return p
}

func InitEmptyStatementsContext(p *StatementsContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_statements
}

func (*StatementsContext) IsStatementsContext() {}

func NewStatementsContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *StatementsContext {
	var p = new(StatementsContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SQLParserRULE_statements

	return p
}

func (s *StatementsContext) GetParser() antlr.Parser { return s.parser }

func (s *StatementsContext) EOF() antlr.TerminalNode {
	return s.GetToken(SQLParserEOF, 0)
}

func (s *StatementsContext) AllSql_stmt_list() []ISql_stmt_listContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(ISql_stmt_listContext); ok {
			len++
		}
	}

	tst := make([]ISql_stmt_listContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(ISql_stmt_listContext); ok {
			tst[i] = t.(ISql_stmt_listContext)
			i++
		}
	}

	return tst
}

func (s *StatementsContext) Sql_stmt_list(i int) ISql_stmt_listContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ISql_stmt_listContext); ok {
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

	return t.(ISql_stmt_listContext)
}

func (s *StatementsContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *StatementsContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *StatementsContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SQLParserVisitor:
		return t.VisitStatements(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SQLParser) Statements() (localctx IStatementsContext) {
	localctx = NewStatementsContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 0, SQLParserRULE_statements)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(119)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for ((int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&2251808403619842) != 0) || ((int64((_la-77)) & ^0x3f) == 0 && ((int64(1)<<(_la-77))&1057) != 0) {
		{
			p.SetState(116)
			p.Sql_stmt_list()
		}

		p.SetState(121)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(122)
		p.Match(SQLParserEOF)
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

// ISql_stmt_listContext is an interface to support dynamic dispatch.
type ISql_stmt_listContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllSql_stmt() []ISql_stmtContext
	Sql_stmt(i int) ISql_stmtContext
	AllSCOL() []antlr.TerminalNode
	SCOL(i int) antlr.TerminalNode

	// IsSql_stmt_listContext differentiates from other interfaces.
	IsSql_stmt_listContext()
}

type Sql_stmt_listContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptySql_stmt_listContext() *Sql_stmt_listContext {
	var p = new(Sql_stmt_listContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_sql_stmt_list
	return p
}

func InitEmptySql_stmt_listContext(p *Sql_stmt_listContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_sql_stmt_list
}

func (*Sql_stmt_listContext) IsSql_stmt_listContext() {}

func NewSql_stmt_listContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Sql_stmt_listContext {
	var p = new(Sql_stmt_listContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SQLParserRULE_sql_stmt_list

	return p
}

func (s *Sql_stmt_listContext) GetParser() antlr.Parser { return s.parser }

func (s *Sql_stmt_listContext) AllSql_stmt() []ISql_stmtContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(ISql_stmtContext); ok {
			len++
		}
	}

	tst := make([]ISql_stmtContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(ISql_stmtContext); ok {
			tst[i] = t.(ISql_stmtContext)
			i++
		}
	}

	return tst
}

func (s *Sql_stmt_listContext) Sql_stmt(i int) ISql_stmtContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ISql_stmtContext); ok {
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

	return t.(ISql_stmtContext)
}

func (s *Sql_stmt_listContext) AllSCOL() []antlr.TerminalNode {
	return s.GetTokens(SQLParserSCOL)
}

func (s *Sql_stmt_listContext) SCOL(i int) antlr.TerminalNode {
	return s.GetToken(SQLParserSCOL, i)
}

func (s *Sql_stmt_listContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Sql_stmt_listContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Sql_stmt_listContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SQLParserVisitor:
		return t.VisitSql_stmt_list(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SQLParser) Sql_stmt_list() (localctx ISql_stmt_listContext) {
	localctx = NewSql_stmt_listContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 2, SQLParserRULE_sql_stmt_list)
	var _la int

	var _alt int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(127)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == SQLParserSCOL {
		{
			p.SetState(124)
			p.Match(SQLParserSCOL)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

		p.SetState(129)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(130)
		p.Sql_stmt()
	}
	p.SetState(139)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 3, p.GetParserRuleContext())
	if p.HasError() {
		goto errorExit
	}
	for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
		if _alt == 1 {
			p.SetState(132)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)

			for ok := true; ok; ok = _la == SQLParserSCOL {
				{
					p.SetState(131)
					p.Match(SQLParserSCOL)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}

				p.SetState(134)
				p.GetErrorHandler().Sync(p)
				if p.HasError() {
					goto errorExit
				}
				_la = p.GetTokenStream().LA(1)
			}
			{
				p.SetState(136)
				p.Sql_stmt()
			}

		}
		p.SetState(141)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 3, p.GetParserRuleContext())
		if p.HasError() {
			goto errorExit
		}
	}
	p.SetState(145)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 4, p.GetParserRuleContext())
	if p.HasError() {
		goto errorExit
	}
	for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
		if _alt == 1 {
			{
				p.SetState(142)
				p.Match(SQLParserSCOL)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

		}
		p.SetState(147)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 4, p.GetParserRuleContext())
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
	Delete_stmt() IDelete_stmtContext
	Insert_stmt() IInsert_stmtContext
	Select_stmt() ISelect_stmtContext
	Update_stmt() IUpdate_stmtContext

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
	p.RuleIndex = SQLParserRULE_sql_stmt
	return p
}

func InitEmptySql_stmtContext(p *Sql_stmtContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_sql_stmt
}

func (*Sql_stmtContext) IsSql_stmtContext() {}

func NewSql_stmtContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Sql_stmtContext {
	var p = new(Sql_stmtContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SQLParserRULE_sql_stmt

	return p
}

func (s *Sql_stmtContext) GetParser() antlr.Parser { return s.parser }

func (s *Sql_stmtContext) Delete_stmt() IDelete_stmtContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IDelete_stmtContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IDelete_stmtContext)
}

func (s *Sql_stmtContext) Insert_stmt() IInsert_stmtContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IInsert_stmtContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IInsert_stmtContext)
}

func (s *Sql_stmtContext) Select_stmt() ISelect_stmtContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ISelect_stmtContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ISelect_stmtContext)
}

func (s *Sql_stmtContext) Update_stmt() IUpdate_stmtContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IUpdate_stmtContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IUpdate_stmtContext)
}

func (s *Sql_stmtContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Sql_stmtContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Sql_stmtContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SQLParserVisitor:
		return t.VisitSql_stmt(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SQLParser) Sql_stmt() (localctx ISql_stmtContext) {
	localctx = NewSql_stmtContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 4, SQLParserRULE_sql_stmt)
	p.EnterOuterAlt(localctx, 1)
	p.SetState(152)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 5, p.GetParserRuleContext()) {
	case 1:
		{
			p.SetState(148)
			p.Delete_stmt()
		}

	case 2:
		{
			p.SetState(149)
			p.Insert_stmt()
		}

	case 3:
		{
			p.SetState(150)
			p.Select_stmt()
		}

	case 4:
		{
			p.SetState(151)
			p.Update_stmt()
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

// IIndexed_columnContext is an interface to support dynamic dispatch.
type IIndexed_columnContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Column_name() IColumn_nameContext

	// IsIndexed_columnContext differentiates from other interfaces.
	IsIndexed_columnContext()
}

type Indexed_columnContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyIndexed_columnContext() *Indexed_columnContext {
	var p = new(Indexed_columnContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_indexed_column
	return p
}

func InitEmptyIndexed_columnContext(p *Indexed_columnContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_indexed_column
}

func (*Indexed_columnContext) IsIndexed_columnContext() {}

func NewIndexed_columnContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Indexed_columnContext {
	var p = new(Indexed_columnContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SQLParserRULE_indexed_column

	return p
}

func (s *Indexed_columnContext) GetParser() antlr.Parser { return s.parser }

func (s *Indexed_columnContext) Column_name() IColumn_nameContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IColumn_nameContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IColumn_nameContext)
}

func (s *Indexed_columnContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Indexed_columnContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Indexed_columnContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SQLParserVisitor:
		return t.VisitIndexed_column(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SQLParser) Indexed_column() (localctx IIndexed_columnContext) {
	localctx = NewIndexed_columnContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 6, SQLParserRULE_indexed_column)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(154)
		p.Column_name()
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

// ICte_table_nameContext is an interface to support dynamic dispatch.
type ICte_table_nameContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Table_name() ITable_nameContext
	OPEN_PAR() antlr.TerminalNode
	AllColumn_name() []IColumn_nameContext
	Column_name(i int) IColumn_nameContext
	CLOSE_PAR() antlr.TerminalNode
	AllCOMMA() []antlr.TerminalNode
	COMMA(i int) antlr.TerminalNode

	// IsCte_table_nameContext differentiates from other interfaces.
	IsCte_table_nameContext()
}

type Cte_table_nameContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyCte_table_nameContext() *Cte_table_nameContext {
	var p = new(Cte_table_nameContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_cte_table_name
	return p
}

func InitEmptyCte_table_nameContext(p *Cte_table_nameContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_cte_table_name
}

func (*Cte_table_nameContext) IsCte_table_nameContext() {}

func NewCte_table_nameContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Cte_table_nameContext {
	var p = new(Cte_table_nameContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SQLParserRULE_cte_table_name

	return p
}

func (s *Cte_table_nameContext) GetParser() antlr.Parser { return s.parser }

func (s *Cte_table_nameContext) Table_name() ITable_nameContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITable_nameContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ITable_nameContext)
}

func (s *Cte_table_nameContext) OPEN_PAR() antlr.TerminalNode {
	return s.GetToken(SQLParserOPEN_PAR, 0)
}

func (s *Cte_table_nameContext) AllColumn_name() []IColumn_nameContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IColumn_nameContext); ok {
			len++
		}
	}

	tst := make([]IColumn_nameContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IColumn_nameContext); ok {
			tst[i] = t.(IColumn_nameContext)
			i++
		}
	}

	return tst
}

func (s *Cte_table_nameContext) Column_name(i int) IColumn_nameContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IColumn_nameContext); ok {
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

	return t.(IColumn_nameContext)
}

func (s *Cte_table_nameContext) CLOSE_PAR() antlr.TerminalNode {
	return s.GetToken(SQLParserCLOSE_PAR, 0)
}

func (s *Cte_table_nameContext) AllCOMMA() []antlr.TerminalNode {
	return s.GetTokens(SQLParserCOMMA)
}

func (s *Cte_table_nameContext) COMMA(i int) antlr.TerminalNode {
	return s.GetToken(SQLParserCOMMA, i)
}

func (s *Cte_table_nameContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Cte_table_nameContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Cte_table_nameContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SQLParserVisitor:
		return t.VisitCte_table_name(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SQLParser) Cte_table_name() (localctx ICte_table_nameContext) {
	localctx = NewCte_table_nameContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 8, SQLParserRULE_cte_table_name)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(156)
		p.Table_name()
	}
	p.SetState(168)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == SQLParserOPEN_PAR {
		{
			p.SetState(157)
			p.Match(SQLParserOPEN_PAR)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(158)
			p.Column_name()
		}
		p.SetState(163)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		for _la == SQLParserCOMMA {
			{
				p.SetState(159)
				p.Match(SQLParserCOMMA)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(160)
				p.Column_name()
			}

			p.SetState(165)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)
		}
		{
			p.SetState(166)
			p.Match(SQLParserCLOSE_PAR)
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

// ICommon_table_expressionContext is an interface to support dynamic dispatch.
type ICommon_table_expressionContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Cte_table_name() ICte_table_nameContext
	AS_() antlr.TerminalNode
	OPEN_PAR() antlr.TerminalNode
	Select_stmt_no_cte() ISelect_stmt_no_cteContext
	CLOSE_PAR() antlr.TerminalNode

	// IsCommon_table_expressionContext differentiates from other interfaces.
	IsCommon_table_expressionContext()
}

type Common_table_expressionContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyCommon_table_expressionContext() *Common_table_expressionContext {
	var p = new(Common_table_expressionContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_common_table_expression
	return p
}

func InitEmptyCommon_table_expressionContext(p *Common_table_expressionContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_common_table_expression
}

func (*Common_table_expressionContext) IsCommon_table_expressionContext() {}

func NewCommon_table_expressionContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Common_table_expressionContext {
	var p = new(Common_table_expressionContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SQLParserRULE_common_table_expression

	return p
}

func (s *Common_table_expressionContext) GetParser() antlr.Parser { return s.parser }

func (s *Common_table_expressionContext) Cte_table_name() ICte_table_nameContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ICte_table_nameContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ICte_table_nameContext)
}

func (s *Common_table_expressionContext) AS_() antlr.TerminalNode {
	return s.GetToken(SQLParserAS_, 0)
}

func (s *Common_table_expressionContext) OPEN_PAR() antlr.TerminalNode {
	return s.GetToken(SQLParserOPEN_PAR, 0)
}

func (s *Common_table_expressionContext) Select_stmt_no_cte() ISelect_stmt_no_cteContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ISelect_stmt_no_cteContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ISelect_stmt_no_cteContext)
}

func (s *Common_table_expressionContext) CLOSE_PAR() antlr.TerminalNode {
	return s.GetToken(SQLParserCLOSE_PAR, 0)
}

func (s *Common_table_expressionContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Common_table_expressionContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Common_table_expressionContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SQLParserVisitor:
		return t.VisitCommon_table_expression(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SQLParser) Common_table_expression() (localctx ICommon_table_expressionContext) {
	localctx = NewCommon_table_expressionContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 10, SQLParserRULE_common_table_expression)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(170)
		p.Cte_table_name()
	}
	{
		p.SetState(171)
		p.Match(SQLParserAS_)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(172)
		p.Match(SQLParserOPEN_PAR)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(173)
		p.Select_stmt_no_cte()
	}
	{
		p.SetState(174)
		p.Match(SQLParserCLOSE_PAR)
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

// ICommon_table_stmtContext is an interface to support dynamic dispatch.
type ICommon_table_stmtContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	WITH_() antlr.TerminalNode
	AllCommon_table_expression() []ICommon_table_expressionContext
	Common_table_expression(i int) ICommon_table_expressionContext
	AllCOMMA() []antlr.TerminalNode
	COMMA(i int) antlr.TerminalNode

	// IsCommon_table_stmtContext differentiates from other interfaces.
	IsCommon_table_stmtContext()
}

type Common_table_stmtContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyCommon_table_stmtContext() *Common_table_stmtContext {
	var p = new(Common_table_stmtContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_common_table_stmt
	return p
}

func InitEmptyCommon_table_stmtContext(p *Common_table_stmtContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_common_table_stmt
}

func (*Common_table_stmtContext) IsCommon_table_stmtContext() {}

func NewCommon_table_stmtContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Common_table_stmtContext {
	var p = new(Common_table_stmtContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SQLParserRULE_common_table_stmt

	return p
}

func (s *Common_table_stmtContext) GetParser() antlr.Parser { return s.parser }

func (s *Common_table_stmtContext) WITH_() antlr.TerminalNode {
	return s.GetToken(SQLParserWITH_, 0)
}

func (s *Common_table_stmtContext) AllCommon_table_expression() []ICommon_table_expressionContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(ICommon_table_expressionContext); ok {
			len++
		}
	}

	tst := make([]ICommon_table_expressionContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(ICommon_table_expressionContext); ok {
			tst[i] = t.(ICommon_table_expressionContext)
			i++
		}
	}

	return tst
}

func (s *Common_table_stmtContext) Common_table_expression(i int) ICommon_table_expressionContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ICommon_table_expressionContext); ok {
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

	return t.(ICommon_table_expressionContext)
}

func (s *Common_table_stmtContext) AllCOMMA() []antlr.TerminalNode {
	return s.GetTokens(SQLParserCOMMA)
}

func (s *Common_table_stmtContext) COMMA(i int) antlr.TerminalNode {
	return s.GetToken(SQLParserCOMMA, i)
}

func (s *Common_table_stmtContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Common_table_stmtContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Common_table_stmtContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SQLParserVisitor:
		return t.VisitCommon_table_stmt(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SQLParser) Common_table_stmt() (localctx ICommon_table_stmtContext) {
	localctx = NewCommon_table_stmtContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 12, SQLParserRULE_common_table_stmt)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(176)
		p.Match(SQLParserWITH_)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(177)
		p.Common_table_expression()
	}
	p.SetState(182)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == SQLParserCOMMA {
		{
			p.SetState(178)
			p.Match(SQLParserCOMMA)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(179)
			p.Common_table_expression()
		}

		p.SetState(184)
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

// IDelete_coreContext is an interface to support dynamic dispatch.
type IDelete_coreContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	DELETE_() antlr.TerminalNode
	FROM_() antlr.TerminalNode
	Qualified_table_name() IQualified_table_nameContext
	WHERE_() antlr.TerminalNode
	Expr() IExprContext
	Returning_clause() IReturning_clauseContext

	// IsDelete_coreContext differentiates from other interfaces.
	IsDelete_coreContext()
}

type Delete_coreContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyDelete_coreContext() *Delete_coreContext {
	var p = new(Delete_coreContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_delete_core
	return p
}

func InitEmptyDelete_coreContext(p *Delete_coreContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_delete_core
}

func (*Delete_coreContext) IsDelete_coreContext() {}

func NewDelete_coreContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Delete_coreContext {
	var p = new(Delete_coreContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SQLParserRULE_delete_core

	return p
}

func (s *Delete_coreContext) GetParser() antlr.Parser { return s.parser }

func (s *Delete_coreContext) DELETE_() antlr.TerminalNode {
	return s.GetToken(SQLParserDELETE_, 0)
}

func (s *Delete_coreContext) FROM_() antlr.TerminalNode {
	return s.GetToken(SQLParserFROM_, 0)
}

func (s *Delete_coreContext) Qualified_table_name() IQualified_table_nameContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IQualified_table_nameContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IQualified_table_nameContext)
}

func (s *Delete_coreContext) WHERE_() antlr.TerminalNode {
	return s.GetToken(SQLParserWHERE_, 0)
}

func (s *Delete_coreContext) Expr() IExprContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExprContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExprContext)
}

func (s *Delete_coreContext) Returning_clause() IReturning_clauseContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IReturning_clauseContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IReturning_clauseContext)
}

func (s *Delete_coreContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Delete_coreContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Delete_coreContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SQLParserVisitor:
		return t.VisitDelete_core(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SQLParser) Delete_core() (localctx IDelete_coreContext) {
	localctx = NewDelete_coreContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 14, SQLParserRULE_delete_core)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(185)
		p.Match(SQLParserDELETE_)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(186)
		p.Match(SQLParserFROM_)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(187)
		p.Qualified_table_name()
	}
	p.SetState(190)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == SQLParserWHERE_ {
		{
			p.SetState(188)
			p.Match(SQLParserWHERE_)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(189)
			p.expr(0)
		}

	}
	p.SetState(193)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == SQLParserRETURNING_ {
		{
			p.SetState(192)
			p.Returning_clause()
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

// IDelete_stmtContext is an interface to support dynamic dispatch.
type IDelete_stmtContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Delete_core() IDelete_coreContext
	Common_table_stmt() ICommon_table_stmtContext

	// IsDelete_stmtContext differentiates from other interfaces.
	IsDelete_stmtContext()
}

type Delete_stmtContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyDelete_stmtContext() *Delete_stmtContext {
	var p = new(Delete_stmtContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_delete_stmt
	return p
}

func InitEmptyDelete_stmtContext(p *Delete_stmtContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_delete_stmt
}

func (*Delete_stmtContext) IsDelete_stmtContext() {}

func NewDelete_stmtContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Delete_stmtContext {
	var p = new(Delete_stmtContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SQLParserRULE_delete_stmt

	return p
}

func (s *Delete_stmtContext) GetParser() antlr.Parser { return s.parser }

func (s *Delete_stmtContext) Delete_core() IDelete_coreContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IDelete_coreContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IDelete_coreContext)
}

func (s *Delete_stmtContext) Common_table_stmt() ICommon_table_stmtContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ICommon_table_stmtContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ICommon_table_stmtContext)
}

func (s *Delete_stmtContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Delete_stmtContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Delete_stmtContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SQLParserVisitor:
		return t.VisitDelete_stmt(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SQLParser) Delete_stmt() (localctx IDelete_stmtContext) {
	localctx = NewDelete_stmtContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 16, SQLParserRULE_delete_stmt)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(196)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == SQLParserWITH_ {
		{
			p.SetState(195)
			p.Common_table_stmt()
		}

	}
	{
		p.SetState(198)
		p.Delete_core()
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
	BIND_PARAMETER() antlr.TerminalNode

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
	p.RuleIndex = SQLParserRULE_variable
	return p
}

func InitEmptyVariableContext(p *VariableContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_variable
}

func (*VariableContext) IsVariableContext() {}

func NewVariableContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *VariableContext {
	var p = new(VariableContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SQLParserRULE_variable

	return p
}

func (s *VariableContext) GetParser() antlr.Parser { return s.parser }

func (s *VariableContext) BIND_PARAMETER() antlr.TerminalNode {
	return s.GetToken(SQLParserBIND_PARAMETER, 0)
}

func (s *VariableContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *VariableContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *VariableContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SQLParserVisitor:
		return t.VisitVariable(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SQLParser) Variable() (localctx IVariableContext) {
	localctx = NewVariableContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 18, SQLParserRULE_variable)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(200)
		p.Match(SQLParserBIND_PARAMETER)
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

// IFunction_callContext is an interface to support dynamic dispatch.
type IFunction_callContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Function_name() IFunction_nameContext
	OPEN_PAR() antlr.TerminalNode
	CLOSE_PAR() antlr.TerminalNode
	STAR() antlr.TerminalNode
	AllExpr() []IExprContext
	Expr(i int) IExprContext
	DISTINCT_() antlr.TerminalNode
	AllCOMMA() []antlr.TerminalNode
	COMMA(i int) antlr.TerminalNode

	// IsFunction_callContext differentiates from other interfaces.
	IsFunction_callContext()
}

type Function_callContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyFunction_callContext() *Function_callContext {
	var p = new(Function_callContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_function_call
	return p
}

func InitEmptyFunction_callContext(p *Function_callContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_function_call
}

func (*Function_callContext) IsFunction_callContext() {}

func NewFunction_callContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Function_callContext {
	var p = new(Function_callContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SQLParserRULE_function_call

	return p
}

func (s *Function_callContext) GetParser() antlr.Parser { return s.parser }

func (s *Function_callContext) Function_name() IFunction_nameContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IFunction_nameContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IFunction_nameContext)
}

func (s *Function_callContext) OPEN_PAR() antlr.TerminalNode {
	return s.GetToken(SQLParserOPEN_PAR, 0)
}

func (s *Function_callContext) CLOSE_PAR() antlr.TerminalNode {
	return s.GetToken(SQLParserCLOSE_PAR, 0)
}

func (s *Function_callContext) STAR() antlr.TerminalNode {
	return s.GetToken(SQLParserSTAR, 0)
}

func (s *Function_callContext) AllExpr() []IExprContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IExprContext); ok {
			len++
		}
	}

	tst := make([]IExprContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IExprContext); ok {
			tst[i] = t.(IExprContext)
			i++
		}
	}

	return tst
}

func (s *Function_callContext) Expr(i int) IExprContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExprContext); ok {
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

	return t.(IExprContext)
}

func (s *Function_callContext) DISTINCT_() antlr.TerminalNode {
	return s.GetToken(SQLParserDISTINCT_, 0)
}

func (s *Function_callContext) AllCOMMA() []antlr.TerminalNode {
	return s.GetTokens(SQLParserCOMMA)
}

func (s *Function_callContext) COMMA(i int) antlr.TerminalNode {
	return s.GetToken(SQLParserCOMMA, i)
}

func (s *Function_callContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Function_callContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Function_callContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SQLParserVisitor:
		return t.VisitFunction_call(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SQLParser) Function_call() (localctx IFunction_callContext) {
	localctx = NewFunction_callContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 20, SQLParserRULE_function_call)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(202)
		p.Function_name()
	}
	{
		p.SetState(203)
		p.Match(SQLParserOPEN_PAR)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(216)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	switch p.GetTokenStream().LA(1) {
	case SQLParserOPEN_PAR, SQLParserPLUS, SQLParserMINUS, SQLParserCASE_, SQLParserDISTINCT_, SQLParserEXISTS_, SQLParserFALSE_, SQLParserLIKE_, SQLParserNOT_, SQLParserNULL_, SQLParserREPLACE_, SQLParserTRUE_, SQLParserIDENTIFIER, SQLParserNUMERIC_LITERAL, SQLParserBIND_PARAMETER, SQLParserSTRING_LITERAL:
		p.SetState(205)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == SQLParserDISTINCT_ {
			{
				p.SetState(204)
				p.Match(SQLParserDISTINCT_)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

		}
		{
			p.SetState(207)
			p.expr(0)
		}
		p.SetState(212)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		for _la == SQLParserCOMMA {
			{
				p.SetState(208)
				p.Match(SQLParserCOMMA)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(209)
				p.expr(0)
			}

			p.SetState(214)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)
		}

	case SQLParserSTAR:
		{
			p.SetState(215)
			p.Match(SQLParserSTAR)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case SQLParserCLOSE_PAR:

	default:
	}
	{
		p.SetState(218)
		p.Match(SQLParserCLOSE_PAR)
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

// IColumn_refContext is an interface to support dynamic dispatch.
type IColumn_refContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Column_name() IColumn_nameContext
	Table_name() ITable_nameContext
	DOT() antlr.TerminalNode

	// IsColumn_refContext differentiates from other interfaces.
	IsColumn_refContext()
}

type Column_refContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyColumn_refContext() *Column_refContext {
	var p = new(Column_refContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_column_ref
	return p
}

func InitEmptyColumn_refContext(p *Column_refContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_column_ref
}

func (*Column_refContext) IsColumn_refContext() {}

func NewColumn_refContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Column_refContext {
	var p = new(Column_refContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SQLParserRULE_column_ref

	return p
}

func (s *Column_refContext) GetParser() antlr.Parser { return s.parser }

func (s *Column_refContext) Column_name() IColumn_nameContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IColumn_nameContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IColumn_nameContext)
}

func (s *Column_refContext) Table_name() ITable_nameContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITable_nameContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ITable_nameContext)
}

func (s *Column_refContext) DOT() antlr.TerminalNode {
	return s.GetToken(SQLParserDOT, 0)
}

func (s *Column_refContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Column_refContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Column_refContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SQLParserVisitor:
		return t.VisitColumn_ref(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SQLParser) Column_ref() (localctx IColumn_refContext) {
	localctx = NewColumn_refContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 22, SQLParserRULE_column_ref)
	p.EnterOuterAlt(localctx, 1)
	p.SetState(223)
	p.GetErrorHandler().Sync(p)

	if p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 15, p.GetParserRuleContext()) == 1 {
		{
			p.SetState(220)
			p.Table_name()
		}
		{
			p.SetState(221)
			p.Match(SQLParserDOT)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	} else if p.HasError() { // JIM
		goto errorExit
	}
	{
		p.SetState(225)
		p.Column_name()
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

// IWhen_clauseContext is an interface to support dynamic dispatch.
type IWhen_clauseContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// GetCondition returns the condition rule contexts.
	GetCondition() IExprContext

	// GetResult returns the result rule contexts.
	GetResult() IExprContext

	// SetCondition sets the condition rule contexts.
	SetCondition(IExprContext)

	// SetResult sets the result rule contexts.
	SetResult(IExprContext)

	// Getter signatures
	WHEN_() antlr.TerminalNode
	THEN_() antlr.TerminalNode
	AllExpr() []IExprContext
	Expr(i int) IExprContext

	// IsWhen_clauseContext differentiates from other interfaces.
	IsWhen_clauseContext()
}

type When_clauseContext struct {
	antlr.BaseParserRuleContext
	parser    antlr.Parser
	condition IExprContext
	result    IExprContext
}

func NewEmptyWhen_clauseContext() *When_clauseContext {
	var p = new(When_clauseContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_when_clause
	return p
}

func InitEmptyWhen_clauseContext(p *When_clauseContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_when_clause
}

func (*When_clauseContext) IsWhen_clauseContext() {}

func NewWhen_clauseContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *When_clauseContext {
	var p = new(When_clauseContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SQLParserRULE_when_clause

	return p
}

func (s *When_clauseContext) GetParser() antlr.Parser { return s.parser }

func (s *When_clauseContext) GetCondition() IExprContext { return s.condition }

func (s *When_clauseContext) GetResult() IExprContext { return s.result }

func (s *When_clauseContext) SetCondition(v IExprContext) { s.condition = v }

func (s *When_clauseContext) SetResult(v IExprContext) { s.result = v }

func (s *When_clauseContext) WHEN_() antlr.TerminalNode {
	return s.GetToken(SQLParserWHEN_, 0)
}

func (s *When_clauseContext) THEN_() antlr.TerminalNode {
	return s.GetToken(SQLParserTHEN_, 0)
}

func (s *When_clauseContext) AllExpr() []IExprContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IExprContext); ok {
			len++
		}
	}

	tst := make([]IExprContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IExprContext); ok {
			tst[i] = t.(IExprContext)
			i++
		}
	}

	return tst
}

func (s *When_clauseContext) Expr(i int) IExprContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExprContext); ok {
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

	return t.(IExprContext)
}

func (s *When_clauseContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *When_clauseContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *When_clauseContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SQLParserVisitor:
		return t.VisitWhen_clause(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SQLParser) When_clause() (localctx IWhen_clauseContext) {
	localctx = NewWhen_clauseContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 24, SQLParserRULE_when_clause)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(227)
		p.Match(SQLParserWHEN_)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(228)

		var _x = p.expr(0)

		localctx.(*When_clauseContext).condition = _x
	}
	{
		p.SetState(229)
		p.Match(SQLParserTHEN_)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(230)

		var _x = p.expr(0)

		localctx.(*When_clauseContext).result = _x
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

// IExprContext is an interface to support dynamic dispatch.
type IExprContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser
	// IsExprContext differentiates from other interfaces.
	IsExprContext()
}

type ExprContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyExprContext() *ExprContext {
	var p = new(ExprContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_expr
	return p
}

func InitEmptyExprContext(p *ExprContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_expr
}

func (*ExprContext) IsExprContext() {}

func NewExprContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ExprContext {
	var p = new(ExprContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SQLParserRULE_expr

	return p
}

func (s *ExprContext) GetParser() antlr.Parser { return s.parser }

func (s *ExprContext) CopyAll(ctx *ExprContext) {
	s.CopyFrom(&ctx.BaseParserRuleContext)
}

func (s *ExprContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ExprContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

type Subquery_exprContext struct {
	ExprContext
}

func NewSubquery_exprContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *Subquery_exprContext {
	var p = new(Subquery_exprContext)

	InitEmptyExprContext(&p.ExprContext)
	p.parser = parser
	p.CopyAll(ctx.(*ExprContext))

	return p
}

func (s *Subquery_exprContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Subquery_exprContext) Subquery() ISubqueryContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ISubqueryContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ISubqueryContext)
}

func (s *Subquery_exprContext) EXISTS_() antlr.TerminalNode {
	return s.GetToken(SQLParserEXISTS_, 0)
}

func (s *Subquery_exprContext) NOT_() antlr.TerminalNode {
	return s.GetToken(SQLParserNOT_, 0)
}

func (s *Subquery_exprContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SQLParserVisitor:
		return t.VisitSubquery_expr(s)

	default:
		return t.VisitChildren(s)
	}
}

type Logical_not_exprContext struct {
	ExprContext
}

func NewLogical_not_exprContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *Logical_not_exprContext {
	var p = new(Logical_not_exprContext)

	InitEmptyExprContext(&p.ExprContext)
	p.parser = parser
	p.CopyAll(ctx.(*ExprContext))

	return p
}

func (s *Logical_not_exprContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Logical_not_exprContext) NOT_() antlr.TerminalNode {
	return s.GetToken(SQLParserNOT_, 0)
}

func (s *Logical_not_exprContext) Expr() IExprContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExprContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExprContext)
}

func (s *Logical_not_exprContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SQLParserVisitor:
		return t.VisitLogical_not_expr(s)

	default:
		return t.VisitChildren(s)
	}
}

type Comparison_exprContext struct {
	ExprContext
	left  IExprContext
	right IExprContext
}

func NewComparison_exprContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *Comparison_exprContext {
	var p = new(Comparison_exprContext)

	InitEmptyExprContext(&p.ExprContext)
	p.parser = parser
	p.CopyAll(ctx.(*ExprContext))

	return p
}

func (s *Comparison_exprContext) GetLeft() IExprContext { return s.left }

func (s *Comparison_exprContext) GetRight() IExprContext { return s.right }

func (s *Comparison_exprContext) SetLeft(v IExprContext) { s.left = v }

func (s *Comparison_exprContext) SetRight(v IExprContext) { s.right = v }

func (s *Comparison_exprContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Comparison_exprContext) ComparisonOperator() IComparisonOperatorContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IComparisonOperatorContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IComparisonOperatorContext)
}

func (s *Comparison_exprContext) AllExpr() []IExprContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IExprContext); ok {
			len++
		}
	}

	tst := make([]IExprContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IExprContext); ok {
			tst[i] = t.(IExprContext)
			i++
		}
	}

	return tst
}

func (s *Comparison_exprContext) Expr(i int) IExprContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExprContext); ok {
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

	return t.(IExprContext)
}

func (s *Comparison_exprContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SQLParserVisitor:
		return t.VisitComparison_expr(s)

	default:
		return t.VisitChildren(s)
	}
}

type Like_exprContext struct {
	ExprContext
	elem     IExprContext
	operator antlr.Token
	pattern  IExprContext
	escape   IExprContext
}

func NewLike_exprContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *Like_exprContext {
	var p = new(Like_exprContext)

	InitEmptyExprContext(&p.ExprContext)
	p.parser = parser
	p.CopyAll(ctx.(*ExprContext))

	return p
}

func (s *Like_exprContext) GetOperator() antlr.Token { return s.operator }

func (s *Like_exprContext) SetOperator(v antlr.Token) { s.operator = v }

func (s *Like_exprContext) GetElem() IExprContext { return s.elem }

func (s *Like_exprContext) GetPattern() IExprContext { return s.pattern }

func (s *Like_exprContext) GetEscape() IExprContext { return s.escape }

func (s *Like_exprContext) SetElem(v IExprContext) { s.elem = v }

func (s *Like_exprContext) SetPattern(v IExprContext) { s.pattern = v }

func (s *Like_exprContext) SetEscape(v IExprContext) { s.escape = v }

func (s *Like_exprContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Like_exprContext) AllExpr() []IExprContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IExprContext); ok {
			len++
		}
	}

	tst := make([]IExprContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IExprContext); ok {
			tst[i] = t.(IExprContext)
			i++
		}
	}

	return tst
}

func (s *Like_exprContext) Expr(i int) IExprContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExprContext); ok {
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

	return t.(IExprContext)
}

func (s *Like_exprContext) LIKE_() antlr.TerminalNode {
	return s.GetToken(SQLParserLIKE_, 0)
}

func (s *Like_exprContext) NOT_() antlr.TerminalNode {
	return s.GetToken(SQLParserNOT_, 0)
}

func (s *Like_exprContext) ESCAPE_() antlr.TerminalNode {
	return s.GetToken(SQLParserESCAPE_, 0)
}

func (s *Like_exprContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SQLParserVisitor:
		return t.VisitLike_expr(s)

	default:
		return t.VisitChildren(s)
	}
}

type Null_exprContext struct {
	ExprContext
}

func NewNull_exprContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *Null_exprContext {
	var p = new(Null_exprContext)

	InitEmptyExprContext(&p.ExprContext)
	p.parser = parser
	p.CopyAll(ctx.(*ExprContext))

	return p
}

func (s *Null_exprContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Null_exprContext) Expr() IExprContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExprContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExprContext)
}

func (s *Null_exprContext) ISNULL_() antlr.TerminalNode {
	return s.GetToken(SQLParserISNULL_, 0)
}

func (s *Null_exprContext) NOTNULL_() antlr.TerminalNode {
	return s.GetToken(SQLParserNOTNULL_, 0)
}

func (s *Null_exprContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SQLParserVisitor:
		return t.VisitNull_expr(s)

	default:
		return t.VisitChildren(s)
	}
}

type Column_exprContext struct {
	ExprContext
}

func NewColumn_exprContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *Column_exprContext {
	var p = new(Column_exprContext)

	InitEmptyExprContext(&p.ExprContext)
	p.parser = parser
	p.CopyAll(ctx.(*ExprContext))

	return p
}

func (s *Column_exprContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Column_exprContext) Column_ref() IColumn_refContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IColumn_refContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IColumn_refContext)
}

func (s *Column_exprContext) Type_cast() IType_castContext {
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

func (s *Column_exprContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SQLParserVisitor:
		return t.VisitColumn_expr(s)

	default:
		return t.VisitChildren(s)
	}
}

type In_subquery_exprContext struct {
	ExprContext
	elem     IExprContext
	operator antlr.Token
}

func NewIn_subquery_exprContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *In_subquery_exprContext {
	var p = new(In_subquery_exprContext)

	InitEmptyExprContext(&p.ExprContext)
	p.parser = parser
	p.CopyAll(ctx.(*ExprContext))

	return p
}

func (s *In_subquery_exprContext) GetOperator() antlr.Token { return s.operator }

func (s *In_subquery_exprContext) SetOperator(v antlr.Token) { s.operator = v }

func (s *In_subquery_exprContext) GetElem() IExprContext { return s.elem }

func (s *In_subquery_exprContext) SetElem(v IExprContext) { s.elem = v }

func (s *In_subquery_exprContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *In_subquery_exprContext) Subquery() ISubqueryContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ISubqueryContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ISubqueryContext)
}

func (s *In_subquery_exprContext) Expr() IExprContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExprContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExprContext)
}

func (s *In_subquery_exprContext) IN_() antlr.TerminalNode {
	return s.GetToken(SQLParserIN_, 0)
}

func (s *In_subquery_exprContext) NOT_() antlr.TerminalNode {
	return s.GetToken(SQLParserNOT_, 0)
}

func (s *In_subquery_exprContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SQLParserVisitor:
		return t.VisitIn_subquery_expr(s)

	default:
		return t.VisitChildren(s)
	}
}

type Arithmetic_exprContext struct {
	ExprContext
	left     IExprContext
	operator antlr.Token
	right    IExprContext
}

func NewArithmetic_exprContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *Arithmetic_exprContext {
	var p = new(Arithmetic_exprContext)

	InitEmptyExprContext(&p.ExprContext)
	p.parser = parser
	p.CopyAll(ctx.(*ExprContext))

	return p
}

func (s *Arithmetic_exprContext) GetOperator() antlr.Token { return s.operator }

func (s *Arithmetic_exprContext) SetOperator(v antlr.Token) { s.operator = v }

func (s *Arithmetic_exprContext) GetLeft() IExprContext { return s.left }

func (s *Arithmetic_exprContext) GetRight() IExprContext { return s.right }

func (s *Arithmetic_exprContext) SetLeft(v IExprContext) { s.left = v }

func (s *Arithmetic_exprContext) SetRight(v IExprContext) { s.right = v }

func (s *Arithmetic_exprContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Arithmetic_exprContext) AllExpr() []IExprContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IExprContext); ok {
			len++
		}
	}

	tst := make([]IExprContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IExprContext); ok {
			tst[i] = t.(IExprContext)
			i++
		}
	}

	return tst
}

func (s *Arithmetic_exprContext) Expr(i int) IExprContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExprContext); ok {
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

	return t.(IExprContext)
}

func (s *Arithmetic_exprContext) STAR() antlr.TerminalNode {
	return s.GetToken(SQLParserSTAR, 0)
}

func (s *Arithmetic_exprContext) DIV() antlr.TerminalNode {
	return s.GetToken(SQLParserDIV, 0)
}

func (s *Arithmetic_exprContext) MOD() antlr.TerminalNode {
	return s.GetToken(SQLParserMOD, 0)
}

func (s *Arithmetic_exprContext) PLUS() antlr.TerminalNode {
	return s.GetToken(SQLParserPLUS, 0)
}

func (s *Arithmetic_exprContext) MINUS() antlr.TerminalNode {
	return s.GetToken(SQLParserMINUS, 0)
}

func (s *Arithmetic_exprContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SQLParserVisitor:
		return t.VisitArithmetic_expr(s)

	default:
		return t.VisitChildren(s)
	}
}

type Logical_binary_exprContext struct {
	ExprContext
	left     IExprContext
	operator antlr.Token
	right    IExprContext
}

func NewLogical_binary_exprContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *Logical_binary_exprContext {
	var p = new(Logical_binary_exprContext)

	InitEmptyExprContext(&p.ExprContext)
	p.parser = parser
	p.CopyAll(ctx.(*ExprContext))

	return p
}

func (s *Logical_binary_exprContext) GetOperator() antlr.Token { return s.operator }

func (s *Logical_binary_exprContext) SetOperator(v antlr.Token) { s.operator = v }

func (s *Logical_binary_exprContext) GetLeft() IExprContext { return s.left }

func (s *Logical_binary_exprContext) GetRight() IExprContext { return s.right }

func (s *Logical_binary_exprContext) SetLeft(v IExprContext) { s.left = v }

func (s *Logical_binary_exprContext) SetRight(v IExprContext) { s.right = v }

func (s *Logical_binary_exprContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Logical_binary_exprContext) AllExpr() []IExprContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IExprContext); ok {
			len++
		}
	}

	tst := make([]IExprContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IExprContext); ok {
			tst[i] = t.(IExprContext)
			i++
		}
	}

	return tst
}

func (s *Logical_binary_exprContext) Expr(i int) IExprContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExprContext); ok {
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

	return t.(IExprContext)
}

func (s *Logical_binary_exprContext) AND_() antlr.TerminalNode {
	return s.GetToken(SQLParserAND_, 0)
}

func (s *Logical_binary_exprContext) OR_() antlr.TerminalNode {
	return s.GetToken(SQLParserOR_, 0)
}

func (s *Logical_binary_exprContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SQLParserVisitor:
		return t.VisitLogical_binary_expr(s)

	default:
		return t.VisitChildren(s)
	}
}

type Variable_exprContext struct {
	ExprContext
}

func NewVariable_exprContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *Variable_exprContext {
	var p = new(Variable_exprContext)

	InitEmptyExprContext(&p.ExprContext)
	p.parser = parser
	p.CopyAll(ctx.(*ExprContext))

	return p
}

func (s *Variable_exprContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Variable_exprContext) Variable() IVariableContext {
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

func (s *Variable_exprContext) Type_cast() IType_castContext {
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

func (s *Variable_exprContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SQLParserVisitor:
		return t.VisitVariable_expr(s)

	default:
		return t.VisitChildren(s)
	}
}

type Unary_exprContext struct {
	ExprContext
	operator antlr.Token
}

func NewUnary_exprContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *Unary_exprContext {
	var p = new(Unary_exprContext)

	InitEmptyExprContext(&p.ExprContext)
	p.parser = parser
	p.CopyAll(ctx.(*ExprContext))

	return p
}

func (s *Unary_exprContext) GetOperator() antlr.Token { return s.operator }

func (s *Unary_exprContext) SetOperator(v antlr.Token) { s.operator = v }

func (s *Unary_exprContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Unary_exprContext) Expr() IExprContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExprContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExprContext)
}

func (s *Unary_exprContext) MINUS() antlr.TerminalNode {
	return s.GetToken(SQLParserMINUS, 0)
}

func (s *Unary_exprContext) PLUS() antlr.TerminalNode {
	return s.GetToken(SQLParserPLUS, 0)
}

func (s *Unary_exprContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SQLParserVisitor:
		return t.VisitUnary_expr(s)

	default:
		return t.VisitChildren(s)
	}
}

type Collate_exprContext struct {
	ExprContext
}

func NewCollate_exprContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *Collate_exprContext {
	var p = new(Collate_exprContext)

	InitEmptyExprContext(&p.ExprContext)
	p.parser = parser
	p.CopyAll(ctx.(*ExprContext))

	return p
}

func (s *Collate_exprContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Collate_exprContext) Expr() IExprContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExprContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExprContext)
}

func (s *Collate_exprContext) COLLATE_() antlr.TerminalNode {
	return s.GetToken(SQLParserCOLLATE_, 0)
}

func (s *Collate_exprContext) Collation_name() ICollation_nameContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ICollation_nameContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ICollation_nameContext)
}

func (s *Collate_exprContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SQLParserVisitor:
		return t.VisitCollate_expr(s)

	default:
		return t.VisitChildren(s)
	}
}

type Parenthesized_exprContext struct {
	ExprContext
}

func NewParenthesized_exprContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *Parenthesized_exprContext {
	var p = new(Parenthesized_exprContext)

	InitEmptyExprContext(&p.ExprContext)
	p.parser = parser
	p.CopyAll(ctx.(*ExprContext))

	return p
}

func (s *Parenthesized_exprContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Parenthesized_exprContext) OPEN_PAR() antlr.TerminalNode {
	return s.GetToken(SQLParserOPEN_PAR, 0)
}

func (s *Parenthesized_exprContext) Expr() IExprContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExprContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExprContext)
}

func (s *Parenthesized_exprContext) CLOSE_PAR() antlr.TerminalNode {
	return s.GetToken(SQLParserCLOSE_PAR, 0)
}

func (s *Parenthesized_exprContext) Type_cast() IType_castContext {
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

func (s *Parenthesized_exprContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SQLParserVisitor:
		return t.VisitParenthesized_expr(s)

	default:
		return t.VisitChildren(s)
	}
}

type Between_exprContext struct {
	ExprContext
	elem     IExprContext
	operator antlr.Token
	low      IExprContext
	high     IExprContext
}

func NewBetween_exprContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *Between_exprContext {
	var p = new(Between_exprContext)

	InitEmptyExprContext(&p.ExprContext)
	p.parser = parser
	p.CopyAll(ctx.(*ExprContext))

	return p
}

func (s *Between_exprContext) GetOperator() antlr.Token { return s.operator }

func (s *Between_exprContext) SetOperator(v antlr.Token) { s.operator = v }

func (s *Between_exprContext) GetElem() IExprContext { return s.elem }

func (s *Between_exprContext) GetLow() IExprContext { return s.low }

func (s *Between_exprContext) GetHigh() IExprContext { return s.high }

func (s *Between_exprContext) SetElem(v IExprContext) { s.elem = v }

func (s *Between_exprContext) SetLow(v IExprContext) { s.low = v }

func (s *Between_exprContext) SetHigh(v IExprContext) { s.high = v }

func (s *Between_exprContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Between_exprContext) AND_() antlr.TerminalNode {
	return s.GetToken(SQLParserAND_, 0)
}

func (s *Between_exprContext) AllExpr() []IExprContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IExprContext); ok {
			len++
		}
	}

	tst := make([]IExprContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IExprContext); ok {
			tst[i] = t.(IExprContext)
			i++
		}
	}

	return tst
}

func (s *Between_exprContext) Expr(i int) IExprContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExprContext); ok {
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

	return t.(IExprContext)
}

func (s *Between_exprContext) BETWEEN_() antlr.TerminalNode {
	return s.GetToken(SQLParserBETWEEN_, 0)
}

func (s *Between_exprContext) NOT_() antlr.TerminalNode {
	return s.GetToken(SQLParserNOT_, 0)
}

func (s *Between_exprContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SQLParserVisitor:
		return t.VisitBetween_expr(s)

	default:
		return t.VisitChildren(s)
	}
}

type Expr_list_exprContext struct {
	ExprContext
}

func NewExpr_list_exprContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *Expr_list_exprContext {
	var p = new(Expr_list_exprContext)

	InitEmptyExprContext(&p.ExprContext)
	p.parser = parser
	p.CopyAll(ctx.(*ExprContext))

	return p
}

func (s *Expr_list_exprContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Expr_list_exprContext) OPEN_PAR() antlr.TerminalNode {
	return s.GetToken(SQLParserOPEN_PAR, 0)
}

func (s *Expr_list_exprContext) Expr_list() IExpr_listContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpr_listContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpr_listContext)
}

func (s *Expr_list_exprContext) CLOSE_PAR() antlr.TerminalNode {
	return s.GetToken(SQLParserCLOSE_PAR, 0)
}

func (s *Expr_list_exprContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SQLParserVisitor:
		return t.VisitExpr_list_expr(s)

	default:
		return t.VisitChildren(s)
	}
}

type In_list_exprContext struct {
	ExprContext
	elem     IExprContext
	operator antlr.Token
}

func NewIn_list_exprContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *In_list_exprContext {
	var p = new(In_list_exprContext)

	InitEmptyExprContext(&p.ExprContext)
	p.parser = parser
	p.CopyAll(ctx.(*ExprContext))

	return p
}

func (s *In_list_exprContext) GetOperator() antlr.Token { return s.operator }

func (s *In_list_exprContext) SetOperator(v antlr.Token) { s.operator = v }

func (s *In_list_exprContext) GetElem() IExprContext { return s.elem }

func (s *In_list_exprContext) SetElem(v IExprContext) { s.elem = v }

func (s *In_list_exprContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *In_list_exprContext) OPEN_PAR() antlr.TerminalNode {
	return s.GetToken(SQLParserOPEN_PAR, 0)
}

func (s *In_list_exprContext) Expr_list() IExpr_listContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpr_listContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpr_listContext)
}

func (s *In_list_exprContext) CLOSE_PAR() antlr.TerminalNode {
	return s.GetToken(SQLParserCLOSE_PAR, 0)
}

func (s *In_list_exprContext) Expr() IExprContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExprContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExprContext)
}

func (s *In_list_exprContext) IN_() antlr.TerminalNode {
	return s.GetToken(SQLParserIN_, 0)
}

func (s *In_list_exprContext) NOT_() antlr.TerminalNode {
	return s.GetToken(SQLParserNOT_, 0)
}

func (s *In_list_exprContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SQLParserVisitor:
		return t.VisitIn_list_expr(s)

	default:
		return t.VisitChildren(s)
	}
}

type Literal_exprContext struct {
	ExprContext
}

func NewLiteral_exprContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *Literal_exprContext {
	var p = new(Literal_exprContext)

	InitEmptyExprContext(&p.ExprContext)
	p.parser = parser
	p.CopyAll(ctx.(*ExprContext))

	return p
}

func (s *Literal_exprContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Literal_exprContext) Literal() ILiteralContext {
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

func (s *Literal_exprContext) Type_cast() IType_castContext {
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

func (s *Literal_exprContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SQLParserVisitor:
		return t.VisitLiteral_expr(s)

	default:
		return t.VisitChildren(s)
	}
}

type Is_exprContext struct {
	ExprContext
}

func NewIs_exprContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *Is_exprContext {
	var p = new(Is_exprContext)

	InitEmptyExprContext(&p.ExprContext)
	p.parser = parser
	p.CopyAll(ctx.(*ExprContext))

	return p
}

func (s *Is_exprContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Is_exprContext) AllExpr() []IExprContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IExprContext); ok {
			len++
		}
	}

	tst := make([]IExprContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IExprContext); ok {
			tst[i] = t.(IExprContext)
			i++
		}
	}

	return tst
}

func (s *Is_exprContext) Expr(i int) IExprContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExprContext); ok {
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

	return t.(IExprContext)
}

func (s *Is_exprContext) IS_() antlr.TerminalNode {
	return s.GetToken(SQLParserIS_, 0)
}

func (s *Is_exprContext) DISTINCT_() antlr.TerminalNode {
	return s.GetToken(SQLParserDISTINCT_, 0)
}

func (s *Is_exprContext) FROM_() antlr.TerminalNode {
	return s.GetToken(SQLParserFROM_, 0)
}

func (s *Is_exprContext) Boolean_value() IBoolean_valueContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IBoolean_valueContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IBoolean_valueContext)
}

func (s *Is_exprContext) NULL_() antlr.TerminalNode {
	return s.GetToken(SQLParserNULL_, 0)
}

func (s *Is_exprContext) NOT_() antlr.TerminalNode {
	return s.GetToken(SQLParserNOT_, 0)
}

func (s *Is_exprContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SQLParserVisitor:
		return t.VisitIs_expr(s)

	default:
		return t.VisitChildren(s)
	}
}

type Case_exprContext struct {
	ExprContext
	case_clause IExprContext
	else_clause IExprContext
}

func NewCase_exprContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *Case_exprContext {
	var p = new(Case_exprContext)

	InitEmptyExprContext(&p.ExprContext)
	p.parser = parser
	p.CopyAll(ctx.(*ExprContext))

	return p
}

func (s *Case_exprContext) GetCase_clause() IExprContext { return s.case_clause }

func (s *Case_exprContext) GetElse_clause() IExprContext { return s.else_clause }

func (s *Case_exprContext) SetCase_clause(v IExprContext) { s.case_clause = v }

func (s *Case_exprContext) SetElse_clause(v IExprContext) { s.else_clause = v }

func (s *Case_exprContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Case_exprContext) CASE_() antlr.TerminalNode {
	return s.GetToken(SQLParserCASE_, 0)
}

func (s *Case_exprContext) END_() antlr.TerminalNode {
	return s.GetToken(SQLParserEND_, 0)
}

func (s *Case_exprContext) AllWhen_clause() []IWhen_clauseContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IWhen_clauseContext); ok {
			len++
		}
	}

	tst := make([]IWhen_clauseContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IWhen_clauseContext); ok {
			tst[i] = t.(IWhen_clauseContext)
			i++
		}
	}

	return tst
}

func (s *Case_exprContext) When_clause(i int) IWhen_clauseContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IWhen_clauseContext); ok {
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

	return t.(IWhen_clauseContext)
}

func (s *Case_exprContext) ELSE_() antlr.TerminalNode {
	return s.GetToken(SQLParserELSE_, 0)
}

func (s *Case_exprContext) AllExpr() []IExprContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IExprContext); ok {
			len++
		}
	}

	tst := make([]IExprContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IExprContext); ok {
			tst[i] = t.(IExprContext)
			i++
		}
	}

	return tst
}

func (s *Case_exprContext) Expr(i int) IExprContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExprContext); ok {
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

	return t.(IExprContext)
}

func (s *Case_exprContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SQLParserVisitor:
		return t.VisitCase_expr(s)

	default:
		return t.VisitChildren(s)
	}
}

type Function_exprContext struct {
	ExprContext
}

func NewFunction_exprContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *Function_exprContext {
	var p = new(Function_exprContext)

	InitEmptyExprContext(&p.ExprContext)
	p.parser = parser
	p.CopyAll(ctx.(*ExprContext))

	return p
}

func (s *Function_exprContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Function_exprContext) Function_call() IFunction_callContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IFunction_callContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IFunction_callContext)
}

func (s *Function_exprContext) Type_cast() IType_castContext {
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

func (s *Function_exprContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SQLParserVisitor:
		return t.VisitFunction_expr(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SQLParser) Expr() (localctx IExprContext) {
	return p.expr(0)
}

func (p *SQLParser) expr(_p int) (localctx IExprContext) {
	var _parentctx antlr.ParserRuleContext = p.GetParserRuleContext()

	_parentState := p.GetState()
	localctx = NewExprContext(p, p.GetParserRuleContext(), _parentState)
	var _prevctx IExprContext = localctx
	var _ antlr.ParserRuleContext = _prevctx // TODO: To prevent unused variable warning.
	_startState := 26
	p.EnterRecursionRule(localctx, 26, SQLParserRULE_expr, _p)
	var _la int

	var _alt int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(285)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 26, p.GetParserRuleContext()) {
	case 1:
		localctx = NewLiteral_exprContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx

		{
			p.SetState(233)
			p.Literal()
		}
		p.SetState(235)
		p.GetErrorHandler().Sync(p)

		if p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 16, p.GetParserRuleContext()) == 1 {
			{
				p.SetState(234)
				p.Type_cast()
			}

		} else if p.HasError() { // JIM
			goto errorExit
		}

	case 2:
		localctx = NewVariable_exprContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(237)
			p.Variable()
		}
		p.SetState(239)
		p.GetErrorHandler().Sync(p)

		if p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 17, p.GetParserRuleContext()) == 1 {
			{
				p.SetState(238)
				p.Type_cast()
			}

		} else if p.HasError() { // JIM
			goto errorExit
		}

	case 3:
		localctx = NewColumn_exprContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(241)
			p.Column_ref()
		}
		p.SetState(243)
		p.GetErrorHandler().Sync(p)

		if p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 18, p.GetParserRuleContext()) == 1 {
			{
				p.SetState(242)
				p.Type_cast()
			}

		} else if p.HasError() { // JIM
			goto errorExit
		}

	case 4:
		localctx = NewUnary_exprContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(245)

			var _lt = p.GetTokenStream().LT(1)

			localctx.(*Unary_exprContext).operator = _lt

			_la = p.GetTokenStream().LA(1)

			if !(_la == SQLParserPLUS || _la == SQLParserMINUS) {
				var _ri = p.GetErrorHandler().RecoverInline(p)

				localctx.(*Unary_exprContext).operator = _ri
			} else {
				p.GetErrorHandler().ReportMatch(p)
				p.Consume()
			}
		}
		{
			p.SetState(246)
			p.expr(19)
		}

	case 5:
		localctx = NewParenthesized_exprContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(247)
			p.Match(SQLParserOPEN_PAR)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(248)
			p.expr(0)
		}
		{
			p.SetState(249)
			p.Match(SQLParserCLOSE_PAR)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(251)
		p.GetErrorHandler().Sync(p)

		if p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 19, p.GetParserRuleContext()) == 1 {
			{
				p.SetState(250)
				p.Type_cast()
			}

		} else if p.HasError() { // JIM
			goto errorExit
		}

	case 6:
		localctx = NewSubquery_exprContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		p.SetState(257)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == SQLParserEXISTS_ || _la == SQLParserNOT_ {
			p.SetState(254)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)

			if _la == SQLParserNOT_ {
				{
					p.SetState(253)
					p.Match(SQLParserNOT_)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}

			}
			{
				p.SetState(256)
				p.Match(SQLParserEXISTS_)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

		}
		{
			p.SetState(259)
			p.Subquery()
		}

	case 7:
		localctx = NewCase_exprContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(260)
			p.Match(SQLParserCASE_)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(262)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if ((int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&1152928101743723272) != 0) || ((int64((_la-64)) & ^0x3f) == 0 && ((int64(1)<<(_la-64))&251724805) != 0) {
			{
				p.SetState(261)

				var _x = p.expr(0)

				localctx.(*Case_exprContext).case_clause = _x
			}

		}
		p.SetState(265)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		for ok := true; ok; ok = _la == SQLParserWHEN_ {
			{
				p.SetState(264)
				p.When_clause()
			}

			p.SetState(267)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)
		}
		p.SetState(271)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == SQLParserELSE_ {
			{
				p.SetState(269)
				p.Match(SQLParserELSE_)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(270)

				var _x = p.expr(0)

				localctx.(*Case_exprContext).else_clause = _x
			}

		}
		{
			p.SetState(273)
			p.Match(SQLParserEND_)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 8:
		localctx = NewExpr_list_exprContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(275)
			p.Match(SQLParserOPEN_PAR)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(276)
			p.Expr_list()
		}
		{
			p.SetState(277)
			p.Match(SQLParserCLOSE_PAR)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 9:
		localctx = NewFunction_exprContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(279)
			p.Function_call()
		}
		p.SetState(281)
		p.GetErrorHandler().Sync(p)

		if p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 25, p.GetParserRuleContext()) == 1 {
			{
				p.SetState(280)
				p.Type_cast()
			}

		} else if p.HasError() { // JIM
			goto errorExit
		}

	case 10:
		localctx = NewLogical_not_exprContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(283)
			p.Match(SQLParserNOT_)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(284)
			p.expr(3)
		}

	case antlr.ATNInvalidAltNumber:
		goto errorExit
	}
	p.GetParserRuleContext().SetStop(p.GetTokenStream().LT(-1))
	p.SetState(356)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 35, p.GetParserRuleContext())
	if p.HasError() {
		goto errorExit
	}
	for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
		if _alt == 1 {
			if p.GetParseListeners() != nil {
				p.TriggerExitRuleEvent()
			}
			_prevctx = localctx
			p.SetState(354)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}

			switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 34, p.GetParserRuleContext()) {
			case 1:
				localctx = NewArithmetic_exprContext(p, NewExprContext(p, _parentctx, _parentState))
				localctx.(*Arithmetic_exprContext).left = _prevctx

				p.PushNewRecursionContext(localctx, _startState, SQLParserRULE_expr)
				p.SetState(287)

				if !(p.Precpred(p.GetParserRuleContext(), 12)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 12)", ""))
					goto errorExit
				}
				{
					p.SetState(288)

					var _lt = p.GetTokenStream().LT(1)

					localctx.(*Arithmetic_exprContext).operator = _lt

					_la = p.GetTokenStream().LA(1)

					if !((int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&3200) != 0) {
						var _ri = p.GetErrorHandler().RecoverInline(p)

						localctx.(*Arithmetic_exprContext).operator = _ri
					} else {
						p.GetErrorHandler().ReportMatch(p)
						p.Consume()
					}
				}
				{
					p.SetState(289)

					var _x = p.expr(13)

					localctx.(*Arithmetic_exprContext).right = _x
				}

			case 2:
				localctx = NewArithmetic_exprContext(p, NewExprContext(p, _parentctx, _parentState))
				localctx.(*Arithmetic_exprContext).left = _prevctx

				p.PushNewRecursionContext(localctx, _startState, SQLParserRULE_expr)
				p.SetState(290)

				if !(p.Precpred(p.GetParserRuleContext(), 11)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 11)", ""))
					goto errorExit
				}
				{
					p.SetState(291)

					var _lt = p.GetTokenStream().LT(1)

					localctx.(*Arithmetic_exprContext).operator = _lt

					_la = p.GetTokenStream().LA(1)

					if !(_la == SQLParserPLUS || _la == SQLParserMINUS) {
						var _ri = p.GetErrorHandler().RecoverInline(p)

						localctx.(*Arithmetic_exprContext).operator = _ri
					} else {
						p.GetErrorHandler().ReportMatch(p)
						p.Consume()
					}
				}
				{
					p.SetState(292)

					var _x = p.expr(12)

					localctx.(*Arithmetic_exprContext).right = _x
				}

			case 3:
				localctx = NewBetween_exprContext(p, NewExprContext(p, _parentctx, _parentState))
				localctx.(*Between_exprContext).elem = _prevctx

				p.PushNewRecursionContext(localctx, _startState, SQLParserRULE_expr)
				p.SetState(293)

				if !(p.Precpred(p.GetParserRuleContext(), 8)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 8)", ""))
					goto errorExit
				}
				p.SetState(295)
				p.GetErrorHandler().Sync(p)
				if p.HasError() {
					goto errorExit
				}
				_la = p.GetTokenStream().LA(1)

				if _la == SQLParserNOT_ {
					{
						p.SetState(294)
						p.Match(SQLParserNOT_)
						if p.HasError() {
							// Recognition error - abort rule
							goto errorExit
						}
					}

				}
				{
					p.SetState(297)

					var _m = p.Match(SQLParserBETWEEN_)

					localctx.(*Between_exprContext).operator = _m
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				{
					p.SetState(298)

					var _x = p.expr(0)

					localctx.(*Between_exprContext).low = _x
				}
				{
					p.SetState(299)
					p.Match(SQLParserAND_)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				{
					p.SetState(300)

					var _x = p.expr(9)

					localctx.(*Between_exprContext).high = _x
				}

			case 4:
				localctx = NewComparison_exprContext(p, NewExprContext(p, _parentctx, _parentState))
				localctx.(*Comparison_exprContext).left = _prevctx

				p.PushNewRecursionContext(localctx, _startState, SQLParserRULE_expr)
				p.SetState(302)

				if !(p.Precpred(p.GetParserRuleContext(), 6)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 6)", ""))
					goto errorExit
				}
				{
					p.SetState(303)
					p.ComparisonOperator()
				}
				{
					p.SetState(304)

					var _x = p.expr(7)

					localctx.(*Comparison_exprContext).right = _x
				}

			case 5:
				localctx = NewLogical_binary_exprContext(p, NewExprContext(p, _parentctx, _parentState))
				localctx.(*Logical_binary_exprContext).left = _prevctx

				p.PushNewRecursionContext(localctx, _startState, SQLParserRULE_expr)
				p.SetState(306)

				if !(p.Precpred(p.GetParserRuleContext(), 2)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 2)", ""))
					goto errorExit
				}
				{
					p.SetState(307)

					var _m = p.Match(SQLParserAND_)

					localctx.(*Logical_binary_exprContext).operator = _m
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				{
					p.SetState(308)

					var _x = p.expr(3)

					localctx.(*Logical_binary_exprContext).right = _x
				}

			case 6:
				localctx = NewLogical_binary_exprContext(p, NewExprContext(p, _parentctx, _parentState))
				localctx.(*Logical_binary_exprContext).left = _prevctx

				p.PushNewRecursionContext(localctx, _startState, SQLParserRULE_expr)
				p.SetState(309)

				if !(p.Precpred(p.GetParserRuleContext(), 1)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 1)", ""))
					goto errorExit
				}
				{
					p.SetState(310)

					var _m = p.Match(SQLParserOR_)

					localctx.(*Logical_binary_exprContext).operator = _m
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				{
					p.SetState(311)

					var _x = p.expr(2)

					localctx.(*Logical_binary_exprContext).right = _x
				}

			case 7:
				localctx = NewCollate_exprContext(p, NewExprContext(p, _parentctx, _parentState))
				p.PushNewRecursionContext(localctx, _startState, SQLParserRULE_expr)
				p.SetState(312)

				if !(p.Precpred(p.GetParserRuleContext(), 18)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 18)", ""))
					goto errorExit
				}
				{
					p.SetState(313)
					p.Match(SQLParserCOLLATE_)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				{
					p.SetState(314)
					p.Collation_name()
				}

			case 8:
				localctx = NewIn_subquery_exprContext(p, NewExprContext(p, _parentctx, _parentState))
				localctx.(*In_subquery_exprContext).elem = _prevctx

				p.PushNewRecursionContext(localctx, _startState, SQLParserRULE_expr)
				p.SetState(315)

				if !(p.Precpred(p.GetParserRuleContext(), 10)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 10)", ""))
					goto errorExit
				}
				p.SetState(317)
				p.GetErrorHandler().Sync(p)
				if p.HasError() {
					goto errorExit
				}
				_la = p.GetTokenStream().LA(1)

				if _la == SQLParserNOT_ {
					{
						p.SetState(316)
						p.Match(SQLParserNOT_)
						if p.HasError() {
							// Recognition error - abort rule
							goto errorExit
						}
					}

				}
				{
					p.SetState(319)

					var _m = p.Match(SQLParserIN_)

					localctx.(*In_subquery_exprContext).operator = _m
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				{
					p.SetState(320)
					p.Subquery()
				}

			case 9:
				localctx = NewIn_list_exprContext(p, NewExprContext(p, _parentctx, _parentState))
				localctx.(*In_list_exprContext).elem = _prevctx

				p.PushNewRecursionContext(localctx, _startState, SQLParserRULE_expr)
				p.SetState(321)

				if !(p.Precpred(p.GetParserRuleContext(), 9)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 9)", ""))
					goto errorExit
				}
				p.SetState(323)
				p.GetErrorHandler().Sync(p)
				if p.HasError() {
					goto errorExit
				}
				_la = p.GetTokenStream().LA(1)

				if _la == SQLParserNOT_ {
					{
						p.SetState(322)
						p.Match(SQLParserNOT_)
						if p.HasError() {
							// Recognition error - abort rule
							goto errorExit
						}
					}

				}
				{
					p.SetState(325)

					var _m = p.Match(SQLParserIN_)

					localctx.(*In_list_exprContext).operator = _m
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				{
					p.SetState(326)
					p.Match(SQLParserOPEN_PAR)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				{
					p.SetState(327)
					p.Expr_list()
				}
				{
					p.SetState(328)
					p.Match(SQLParserCLOSE_PAR)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}

			case 10:
				localctx = NewLike_exprContext(p, NewExprContext(p, _parentctx, _parentState))
				localctx.(*Like_exprContext).elem = _prevctx

				p.PushNewRecursionContext(localctx, _startState, SQLParserRULE_expr)
				p.SetState(330)

				if !(p.Precpred(p.GetParserRuleContext(), 7)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 7)", ""))
					goto errorExit
				}
				p.SetState(332)
				p.GetErrorHandler().Sync(p)
				if p.HasError() {
					goto errorExit
				}
				_la = p.GetTokenStream().LA(1)

				if _la == SQLParserNOT_ {
					{
						p.SetState(331)
						p.Match(SQLParserNOT_)
						if p.HasError() {
							// Recognition error - abort rule
							goto errorExit
						}
					}

				}
				{
					p.SetState(334)

					var _m = p.Match(SQLParserLIKE_)

					localctx.(*Like_exprContext).operator = _m
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				{
					p.SetState(335)

					var _x = p.expr(0)

					localctx.(*Like_exprContext).pattern = _x
				}
				p.SetState(338)
				p.GetErrorHandler().Sync(p)

				if p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 31, p.GetParserRuleContext()) == 1 {
					{
						p.SetState(336)
						p.Match(SQLParserESCAPE_)
						if p.HasError() {
							// Recognition error - abort rule
							goto errorExit
						}
					}
					{
						p.SetState(337)

						var _x = p.expr(0)

						localctx.(*Like_exprContext).escape = _x
					}

				} else if p.HasError() { // JIM
					goto errorExit
				}

			case 11:
				localctx = NewIs_exprContext(p, NewExprContext(p, _parentctx, _parentState))
				p.PushNewRecursionContext(localctx, _startState, SQLParserRULE_expr)
				p.SetState(340)

				if !(p.Precpred(p.GetParserRuleContext(), 5)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 5)", ""))
					goto errorExit
				}
				{
					p.SetState(341)
					p.Match(SQLParserIS_)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				p.SetState(343)
				p.GetErrorHandler().Sync(p)
				if p.HasError() {
					goto errorExit
				}
				_la = p.GetTokenStream().LA(1)

				if _la == SQLParserNOT_ {
					{
						p.SetState(342)
						p.Match(SQLParserNOT_)
						if p.HasError() {
							// Recognition error - abort rule
							goto errorExit
						}
					}

				}
				p.SetState(350)
				p.GetErrorHandler().Sync(p)
				if p.HasError() {
					goto errorExit
				}

				switch p.GetTokenStream().LA(1) {
				case SQLParserDISTINCT_:
					{
						p.SetState(345)
						p.Match(SQLParserDISTINCT_)
						if p.HasError() {
							// Recognition error - abort rule
							goto errorExit
						}
					}
					{
						p.SetState(346)
						p.Match(SQLParserFROM_)
						if p.HasError() {
							// Recognition error - abort rule
							goto errorExit
						}
					}
					{
						p.SetState(347)
						p.expr(0)
					}

				case SQLParserFALSE_, SQLParserTRUE_:
					{
						p.SetState(348)
						p.Boolean_value()
					}

				case SQLParserNULL_:
					{
						p.SetState(349)
						p.Match(SQLParserNULL_)
						if p.HasError() {
							// Recognition error - abort rule
							goto errorExit
						}
					}

				default:
					p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
					goto errorExit
				}

			case 12:
				localctx = NewNull_exprContext(p, NewExprContext(p, _parentctx, _parentState))
				p.PushNewRecursionContext(localctx, _startState, SQLParserRULE_expr)
				p.SetState(352)

				if !(p.Precpred(p.GetParserRuleContext(), 4)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 4)", ""))
					goto errorExit
				}
				{
					p.SetState(353)
					_la = p.GetTokenStream().LA(1)

					if !(_la == SQLParserISNULL_ || _la == SQLParserNOTNULL_) {
						p.GetErrorHandler().RecoverInline(p)
					} else {
						p.GetErrorHandler().ReportMatch(p)
						p.Consume()
					}
				}

			case antlr.ATNInvalidAltNumber:
				goto errorExit
			}

		}
		p.SetState(358)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 35, p.GetParserRuleContext())
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

// ISubqueryContext is an interface to support dynamic dispatch.
type ISubqueryContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	OPEN_PAR() antlr.TerminalNode
	Select_stmt_no_cte() ISelect_stmt_no_cteContext
	CLOSE_PAR() antlr.TerminalNode

	// IsSubqueryContext differentiates from other interfaces.
	IsSubqueryContext()
}

type SubqueryContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptySubqueryContext() *SubqueryContext {
	var p = new(SubqueryContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_subquery
	return p
}

func InitEmptySubqueryContext(p *SubqueryContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_subquery
}

func (*SubqueryContext) IsSubqueryContext() {}

func NewSubqueryContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *SubqueryContext {
	var p = new(SubqueryContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SQLParserRULE_subquery

	return p
}

func (s *SubqueryContext) GetParser() antlr.Parser { return s.parser }

func (s *SubqueryContext) OPEN_PAR() antlr.TerminalNode {
	return s.GetToken(SQLParserOPEN_PAR, 0)
}

func (s *SubqueryContext) Select_stmt_no_cte() ISelect_stmt_no_cteContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ISelect_stmt_no_cteContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ISelect_stmt_no_cteContext)
}

func (s *SubqueryContext) CLOSE_PAR() antlr.TerminalNode {
	return s.GetToken(SQLParserCLOSE_PAR, 0)
}

func (s *SubqueryContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *SubqueryContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *SubqueryContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SQLParserVisitor:
		return t.VisitSubquery(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SQLParser) Subquery() (localctx ISubqueryContext) {
	localctx = NewSubqueryContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 28, SQLParserRULE_subquery)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(359)
		p.Match(SQLParserOPEN_PAR)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(360)
		p.Select_stmt_no_cte()
	}
	{
		p.SetState(361)
		p.Match(SQLParserCLOSE_PAR)
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

// IExpr_listContext is an interface to support dynamic dispatch.
type IExpr_listContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllExpr() []IExprContext
	Expr(i int) IExprContext
	AllCOMMA() []antlr.TerminalNode
	COMMA(i int) antlr.TerminalNode

	// IsExpr_listContext differentiates from other interfaces.
	IsExpr_listContext()
}

type Expr_listContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyExpr_listContext() *Expr_listContext {
	var p = new(Expr_listContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_expr_list
	return p
}

func InitEmptyExpr_listContext(p *Expr_listContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_expr_list
}

func (*Expr_listContext) IsExpr_listContext() {}

func NewExpr_listContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Expr_listContext {
	var p = new(Expr_listContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SQLParserRULE_expr_list

	return p
}

func (s *Expr_listContext) GetParser() antlr.Parser { return s.parser }

func (s *Expr_listContext) AllExpr() []IExprContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IExprContext); ok {
			len++
		}
	}

	tst := make([]IExprContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IExprContext); ok {
			tst[i] = t.(IExprContext)
			i++
		}
	}

	return tst
}

func (s *Expr_listContext) Expr(i int) IExprContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExprContext); ok {
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

	return t.(IExprContext)
}

func (s *Expr_listContext) AllCOMMA() []antlr.TerminalNode {
	return s.GetTokens(SQLParserCOMMA)
}

func (s *Expr_listContext) COMMA(i int) antlr.TerminalNode {
	return s.GetToken(SQLParserCOMMA, i)
}

func (s *Expr_listContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Expr_listContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Expr_listContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SQLParserVisitor:
		return t.VisitExpr_list(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SQLParser) Expr_list() (localctx IExpr_listContext) {
	localctx = NewExpr_listContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 30, SQLParserRULE_expr_list)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(363)
		p.expr(0)
	}
	p.SetState(368)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == SQLParserCOMMA {
		{
			p.SetState(364)
			p.Match(SQLParserCOMMA)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(365)
			p.expr(0)
		}

		p.SetState(370)
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

// IComparisonOperatorContext is an interface to support dynamic dispatch.
type IComparisonOperatorContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	LT() antlr.TerminalNode
	LT_EQ() antlr.TerminalNode
	GT() antlr.TerminalNode
	GT_EQ() antlr.TerminalNode
	ASSIGN() antlr.TerminalNode
	NOT_EQ1() antlr.TerminalNode
	NOT_EQ2() antlr.TerminalNode

	// IsComparisonOperatorContext differentiates from other interfaces.
	IsComparisonOperatorContext()
}

type ComparisonOperatorContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyComparisonOperatorContext() *ComparisonOperatorContext {
	var p = new(ComparisonOperatorContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_comparisonOperator
	return p
}

func InitEmptyComparisonOperatorContext(p *ComparisonOperatorContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_comparisonOperator
}

func (*ComparisonOperatorContext) IsComparisonOperatorContext() {}

func NewComparisonOperatorContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ComparisonOperatorContext {
	var p = new(ComparisonOperatorContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SQLParserRULE_comparisonOperator

	return p
}

func (s *ComparisonOperatorContext) GetParser() antlr.Parser { return s.parser }

func (s *ComparisonOperatorContext) LT() antlr.TerminalNode {
	return s.GetToken(SQLParserLT, 0)
}

func (s *ComparisonOperatorContext) LT_EQ() antlr.TerminalNode {
	return s.GetToken(SQLParserLT_EQ, 0)
}

func (s *ComparisonOperatorContext) GT() antlr.TerminalNode {
	return s.GetToken(SQLParserGT, 0)
}

func (s *ComparisonOperatorContext) GT_EQ() antlr.TerminalNode {
	return s.GetToken(SQLParserGT_EQ, 0)
}

func (s *ComparisonOperatorContext) ASSIGN() antlr.TerminalNode {
	return s.GetToken(SQLParserASSIGN, 0)
}

func (s *ComparisonOperatorContext) NOT_EQ1() antlr.TerminalNode {
	return s.GetToken(SQLParserNOT_EQ1, 0)
}

func (s *ComparisonOperatorContext) NOT_EQ2() antlr.TerminalNode {
	return s.GetToken(SQLParserNOT_EQ2, 0)
}

func (s *ComparisonOperatorContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ComparisonOperatorContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ComparisonOperatorContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SQLParserVisitor:
		return t.VisitComparisonOperator(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SQLParser) ComparisonOperator() (localctx IComparisonOperatorContext) {
	localctx = NewComparisonOperatorContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 32, SQLParserRULE_comparisonOperator)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(371)
		_la = p.GetTokenStream().LA(1)

		if !((int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&258112) != 0) {
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

// ICast_typeContext is an interface to support dynamic dispatch.
type ICast_typeContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	IDENTIFIER() antlr.TerminalNode

	// IsCast_typeContext differentiates from other interfaces.
	IsCast_typeContext()
}

type Cast_typeContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyCast_typeContext() *Cast_typeContext {
	var p = new(Cast_typeContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_cast_type
	return p
}

func InitEmptyCast_typeContext(p *Cast_typeContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_cast_type
}

func (*Cast_typeContext) IsCast_typeContext() {}

func NewCast_typeContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Cast_typeContext {
	var p = new(Cast_typeContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SQLParserRULE_cast_type

	return p
}

func (s *Cast_typeContext) GetParser() antlr.Parser { return s.parser }

func (s *Cast_typeContext) IDENTIFIER() antlr.TerminalNode {
	return s.GetToken(SQLParserIDENTIFIER, 0)
}

func (s *Cast_typeContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Cast_typeContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Cast_typeContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SQLParserVisitor:
		return t.VisitCast_type(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SQLParser) Cast_type() (localctx ICast_typeContext) {
	localctx = NewCast_typeContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 34, SQLParserRULE_cast_type)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(373)
		p.Match(SQLParserIDENTIFIER)
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

// IType_castContext is an interface to support dynamic dispatch.
type IType_castContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	TYPE_CAST() antlr.TerminalNode
	Cast_type() ICast_typeContext

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
	p.RuleIndex = SQLParserRULE_type_cast
	return p
}

func InitEmptyType_castContext(p *Type_castContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_type_cast
}

func (*Type_castContext) IsType_castContext() {}

func NewType_castContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Type_castContext {
	var p = new(Type_castContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SQLParserRULE_type_cast

	return p
}

func (s *Type_castContext) GetParser() antlr.Parser { return s.parser }

func (s *Type_castContext) TYPE_CAST() antlr.TerminalNode {
	return s.GetToken(SQLParserTYPE_CAST, 0)
}

func (s *Type_castContext) Cast_type() ICast_typeContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ICast_typeContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ICast_typeContext)
}

func (s *Type_castContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Type_castContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Type_castContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SQLParserVisitor:
		return t.VisitType_cast(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SQLParser) Type_cast() (localctx IType_castContext) {
	localctx = NewType_castContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 36, SQLParserRULE_type_cast)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(375)
		p.Match(SQLParserTYPE_CAST)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(376)
		p.Cast_type()
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

// IBoolean_valueContext is an interface to support dynamic dispatch.
type IBoolean_valueContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	TRUE_() antlr.TerminalNode
	FALSE_() antlr.TerminalNode

	// IsBoolean_valueContext differentiates from other interfaces.
	IsBoolean_valueContext()
}

type Boolean_valueContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyBoolean_valueContext() *Boolean_valueContext {
	var p = new(Boolean_valueContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_boolean_value
	return p
}

func InitEmptyBoolean_valueContext(p *Boolean_valueContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_boolean_value
}

func (*Boolean_valueContext) IsBoolean_valueContext() {}

func NewBoolean_valueContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Boolean_valueContext {
	var p = new(Boolean_valueContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SQLParserRULE_boolean_value

	return p
}

func (s *Boolean_valueContext) GetParser() antlr.Parser { return s.parser }

func (s *Boolean_valueContext) TRUE_() antlr.TerminalNode {
	return s.GetToken(SQLParserTRUE_, 0)
}

func (s *Boolean_valueContext) FALSE_() antlr.TerminalNode {
	return s.GetToken(SQLParserFALSE_, 0)
}

func (s *Boolean_valueContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Boolean_valueContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Boolean_valueContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SQLParserVisitor:
		return t.VisitBoolean_value(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SQLParser) Boolean_value() (localctx IBoolean_valueContext) {
	localctx = NewBoolean_valueContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 38, SQLParserRULE_boolean_value)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(378)
		_la = p.GetTokenStream().LA(1)

		if !(_la == SQLParserFALSE_ || _la == SQLParserTRUE_) {
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

// IString_valueContext is an interface to support dynamic dispatch.
type IString_valueContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	STRING_LITERAL() antlr.TerminalNode

	// IsString_valueContext differentiates from other interfaces.
	IsString_valueContext()
}

type String_valueContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyString_valueContext() *String_valueContext {
	var p = new(String_valueContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_string_value
	return p
}

func InitEmptyString_valueContext(p *String_valueContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_string_value
}

func (*String_valueContext) IsString_valueContext() {}

func NewString_valueContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *String_valueContext {
	var p = new(String_valueContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SQLParserRULE_string_value

	return p
}

func (s *String_valueContext) GetParser() antlr.Parser { return s.parser }

func (s *String_valueContext) STRING_LITERAL() antlr.TerminalNode {
	return s.GetToken(SQLParserSTRING_LITERAL, 0)
}

func (s *String_valueContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *String_valueContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *String_valueContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SQLParserVisitor:
		return t.VisitString_value(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SQLParser) String_value() (localctx IString_valueContext) {
	localctx = NewString_valueContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 40, SQLParserRULE_string_value)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(380)
		p.Match(SQLParserSTRING_LITERAL)
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

// INumeric_valueContext is an interface to support dynamic dispatch.
type INumeric_valueContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	NUMERIC_LITERAL() antlr.TerminalNode

	// IsNumeric_valueContext differentiates from other interfaces.
	IsNumeric_valueContext()
}

type Numeric_valueContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyNumeric_valueContext() *Numeric_valueContext {
	var p = new(Numeric_valueContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_numeric_value
	return p
}

func InitEmptyNumeric_valueContext(p *Numeric_valueContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_numeric_value
}

func (*Numeric_valueContext) IsNumeric_valueContext() {}

func NewNumeric_valueContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Numeric_valueContext {
	var p = new(Numeric_valueContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SQLParserRULE_numeric_value

	return p
}

func (s *Numeric_valueContext) GetParser() antlr.Parser { return s.parser }

func (s *Numeric_valueContext) NUMERIC_LITERAL() antlr.TerminalNode {
	return s.GetToken(SQLParserNUMERIC_LITERAL, 0)
}

func (s *Numeric_valueContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Numeric_valueContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Numeric_valueContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SQLParserVisitor:
		return t.VisitNumeric_value(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SQLParser) Numeric_value() (localctx INumeric_valueContext) {
	localctx = NewNumeric_valueContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 42, SQLParserRULE_numeric_value)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(382)
		p.Match(SQLParserNUMERIC_LITERAL)
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

// ILiteralContext is an interface to support dynamic dispatch.
type ILiteralContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	NULL_() antlr.TerminalNode
	Boolean_value() IBoolean_valueContext
	String_value() IString_valueContext
	Numeric_value() INumeric_valueContext

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
	p.RuleIndex = SQLParserRULE_literal
	return p
}

func InitEmptyLiteralContext(p *LiteralContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_literal
}

func (*LiteralContext) IsLiteralContext() {}

func NewLiteralContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *LiteralContext {
	var p = new(LiteralContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SQLParserRULE_literal

	return p
}

func (s *LiteralContext) GetParser() antlr.Parser { return s.parser }

func (s *LiteralContext) NULL_() antlr.TerminalNode {
	return s.GetToken(SQLParserNULL_, 0)
}

func (s *LiteralContext) Boolean_value() IBoolean_valueContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IBoolean_valueContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IBoolean_valueContext)
}

func (s *LiteralContext) String_value() IString_valueContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IString_valueContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IString_valueContext)
}

func (s *LiteralContext) Numeric_value() INumeric_valueContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(INumeric_valueContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(INumeric_valueContext)
}

func (s *LiteralContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *LiteralContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *LiteralContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SQLParserVisitor:
		return t.VisitLiteral(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SQLParser) Literal() (localctx ILiteralContext) {
	localctx = NewLiteralContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 44, SQLParserRULE_literal)
	p.SetState(388)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case SQLParserNULL_:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(384)
			p.Match(SQLParserNULL_)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case SQLParserFALSE_, SQLParserTRUE_:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(385)
			p.Boolean_value()
		}

	case SQLParserSTRING_LITERAL:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(386)
			p.String_value()
		}

	case SQLParserNUMERIC_LITERAL:
		p.EnterOuterAlt(localctx, 4)
		{
			p.SetState(387)
			p.Numeric_value()
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

// IValue_rowContext is an interface to support dynamic dispatch.
type IValue_rowContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	OPEN_PAR() antlr.TerminalNode
	AllExpr() []IExprContext
	Expr(i int) IExprContext
	CLOSE_PAR() antlr.TerminalNode
	AllCOMMA() []antlr.TerminalNode
	COMMA(i int) antlr.TerminalNode

	// IsValue_rowContext differentiates from other interfaces.
	IsValue_rowContext()
}

type Value_rowContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyValue_rowContext() *Value_rowContext {
	var p = new(Value_rowContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_value_row
	return p
}

func InitEmptyValue_rowContext(p *Value_rowContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_value_row
}

func (*Value_rowContext) IsValue_rowContext() {}

func NewValue_rowContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Value_rowContext {
	var p = new(Value_rowContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SQLParserRULE_value_row

	return p
}

func (s *Value_rowContext) GetParser() antlr.Parser { return s.parser }

func (s *Value_rowContext) OPEN_PAR() antlr.TerminalNode {
	return s.GetToken(SQLParserOPEN_PAR, 0)
}

func (s *Value_rowContext) AllExpr() []IExprContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IExprContext); ok {
			len++
		}
	}

	tst := make([]IExprContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IExprContext); ok {
			tst[i] = t.(IExprContext)
			i++
		}
	}

	return tst
}

func (s *Value_rowContext) Expr(i int) IExprContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExprContext); ok {
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

	return t.(IExprContext)
}

func (s *Value_rowContext) CLOSE_PAR() antlr.TerminalNode {
	return s.GetToken(SQLParserCLOSE_PAR, 0)
}

func (s *Value_rowContext) AllCOMMA() []antlr.TerminalNode {
	return s.GetTokens(SQLParserCOMMA)
}

func (s *Value_rowContext) COMMA(i int) antlr.TerminalNode {
	return s.GetToken(SQLParserCOMMA, i)
}

func (s *Value_rowContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Value_rowContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Value_rowContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SQLParserVisitor:
		return t.VisitValue_row(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SQLParser) Value_row() (localctx IValue_rowContext) {
	localctx = NewValue_rowContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 46, SQLParserRULE_value_row)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(390)
		p.Match(SQLParserOPEN_PAR)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(391)
		p.expr(0)
	}
	p.SetState(396)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == SQLParserCOMMA {
		{
			p.SetState(392)
			p.Match(SQLParserCOMMA)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(393)
			p.expr(0)
		}

		p.SetState(398)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(399)
		p.Match(SQLParserCLOSE_PAR)
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

// IValues_clauseContext is an interface to support dynamic dispatch.
type IValues_clauseContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	VALUES_() antlr.TerminalNode
	AllValue_row() []IValue_rowContext
	Value_row(i int) IValue_rowContext
	AllCOMMA() []antlr.TerminalNode
	COMMA(i int) antlr.TerminalNode

	// IsValues_clauseContext differentiates from other interfaces.
	IsValues_clauseContext()
}

type Values_clauseContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyValues_clauseContext() *Values_clauseContext {
	var p = new(Values_clauseContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_values_clause
	return p
}

func InitEmptyValues_clauseContext(p *Values_clauseContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_values_clause
}

func (*Values_clauseContext) IsValues_clauseContext() {}

func NewValues_clauseContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Values_clauseContext {
	var p = new(Values_clauseContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SQLParserRULE_values_clause

	return p
}

func (s *Values_clauseContext) GetParser() antlr.Parser { return s.parser }

func (s *Values_clauseContext) VALUES_() antlr.TerminalNode {
	return s.GetToken(SQLParserVALUES_, 0)
}

func (s *Values_clauseContext) AllValue_row() []IValue_rowContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IValue_rowContext); ok {
			len++
		}
	}

	tst := make([]IValue_rowContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IValue_rowContext); ok {
			tst[i] = t.(IValue_rowContext)
			i++
		}
	}

	return tst
}

func (s *Values_clauseContext) Value_row(i int) IValue_rowContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IValue_rowContext); ok {
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

	return t.(IValue_rowContext)
}

func (s *Values_clauseContext) AllCOMMA() []antlr.TerminalNode {
	return s.GetTokens(SQLParserCOMMA)
}

func (s *Values_clauseContext) COMMA(i int) antlr.TerminalNode {
	return s.GetToken(SQLParserCOMMA, i)
}

func (s *Values_clauseContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Values_clauseContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Values_clauseContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SQLParserVisitor:
		return t.VisitValues_clause(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SQLParser) Values_clause() (localctx IValues_clauseContext) {
	localctx = NewValues_clauseContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 48, SQLParserRULE_values_clause)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(401)
		p.Match(SQLParserVALUES_)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(402)
		p.Value_row()
	}
	p.SetState(407)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == SQLParserCOMMA {
		{
			p.SetState(403)
			p.Match(SQLParserCOMMA)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(404)
			p.Value_row()
		}

		p.SetState(409)
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

// IInsert_coreContext is an interface to support dynamic dispatch.
type IInsert_coreContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	INSERT_() antlr.TerminalNode
	INTO_() antlr.TerminalNode
	Table_name() ITable_nameContext
	Values_clause() IValues_clauseContext
	AS_() antlr.TerminalNode
	Table_alias() ITable_aliasContext
	OPEN_PAR() antlr.TerminalNode
	AllColumn_name() []IColumn_nameContext
	Column_name(i int) IColumn_nameContext
	CLOSE_PAR() antlr.TerminalNode
	Upsert_clause() IUpsert_clauseContext
	Returning_clause() IReturning_clauseContext
	AllCOMMA() []antlr.TerminalNode
	COMMA(i int) antlr.TerminalNode

	// IsInsert_coreContext differentiates from other interfaces.
	IsInsert_coreContext()
}

type Insert_coreContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyInsert_coreContext() *Insert_coreContext {
	var p = new(Insert_coreContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_insert_core
	return p
}

func InitEmptyInsert_coreContext(p *Insert_coreContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_insert_core
}

func (*Insert_coreContext) IsInsert_coreContext() {}

func NewInsert_coreContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Insert_coreContext {
	var p = new(Insert_coreContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SQLParserRULE_insert_core

	return p
}

func (s *Insert_coreContext) GetParser() antlr.Parser { return s.parser }

func (s *Insert_coreContext) INSERT_() antlr.TerminalNode {
	return s.GetToken(SQLParserINSERT_, 0)
}

func (s *Insert_coreContext) INTO_() antlr.TerminalNode {
	return s.GetToken(SQLParserINTO_, 0)
}

func (s *Insert_coreContext) Table_name() ITable_nameContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITable_nameContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ITable_nameContext)
}

func (s *Insert_coreContext) Values_clause() IValues_clauseContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IValues_clauseContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IValues_clauseContext)
}

func (s *Insert_coreContext) AS_() antlr.TerminalNode {
	return s.GetToken(SQLParserAS_, 0)
}

func (s *Insert_coreContext) Table_alias() ITable_aliasContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITable_aliasContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ITable_aliasContext)
}

func (s *Insert_coreContext) OPEN_PAR() antlr.TerminalNode {
	return s.GetToken(SQLParserOPEN_PAR, 0)
}

func (s *Insert_coreContext) AllColumn_name() []IColumn_nameContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IColumn_nameContext); ok {
			len++
		}
	}

	tst := make([]IColumn_nameContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IColumn_nameContext); ok {
			tst[i] = t.(IColumn_nameContext)
			i++
		}
	}

	return tst
}

func (s *Insert_coreContext) Column_name(i int) IColumn_nameContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IColumn_nameContext); ok {
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

	return t.(IColumn_nameContext)
}

func (s *Insert_coreContext) CLOSE_PAR() antlr.TerminalNode {
	return s.GetToken(SQLParserCLOSE_PAR, 0)
}

func (s *Insert_coreContext) Upsert_clause() IUpsert_clauseContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IUpsert_clauseContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IUpsert_clauseContext)
}

func (s *Insert_coreContext) Returning_clause() IReturning_clauseContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IReturning_clauseContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IReturning_clauseContext)
}

func (s *Insert_coreContext) AllCOMMA() []antlr.TerminalNode {
	return s.GetTokens(SQLParserCOMMA)
}

func (s *Insert_coreContext) COMMA(i int) antlr.TerminalNode {
	return s.GetToken(SQLParserCOMMA, i)
}

func (s *Insert_coreContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Insert_coreContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Insert_coreContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SQLParserVisitor:
		return t.VisitInsert_core(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SQLParser) Insert_core() (localctx IInsert_coreContext) {
	localctx = NewInsert_coreContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 50, SQLParserRULE_insert_core)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(410)
		p.Match(SQLParserINSERT_)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(411)
		p.Match(SQLParserINTO_)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(412)
		p.Table_name()
	}
	p.SetState(415)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == SQLParserAS_ {
		{
			p.SetState(413)
			p.Match(SQLParserAS_)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(414)
			p.Table_alias()
		}

	}
	p.SetState(428)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == SQLParserOPEN_PAR {
		{
			p.SetState(417)
			p.Match(SQLParserOPEN_PAR)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(418)
			p.Column_name()
		}
		p.SetState(423)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		for _la == SQLParserCOMMA {
			{
				p.SetState(419)
				p.Match(SQLParserCOMMA)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(420)
				p.Column_name()
			}

			p.SetState(425)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)
		}
		{
			p.SetState(426)
			p.Match(SQLParserCLOSE_PAR)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	}
	{
		p.SetState(430)
		p.Values_clause()
	}
	p.SetState(432)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == SQLParserON_ {
		{
			p.SetState(431)
			p.Upsert_clause()
		}

	}
	p.SetState(435)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == SQLParserRETURNING_ {
		{
			p.SetState(434)
			p.Returning_clause()
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

// IInsert_stmtContext is an interface to support dynamic dispatch.
type IInsert_stmtContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Insert_core() IInsert_coreContext
	Common_table_stmt() ICommon_table_stmtContext

	// IsInsert_stmtContext differentiates from other interfaces.
	IsInsert_stmtContext()
}

type Insert_stmtContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyInsert_stmtContext() *Insert_stmtContext {
	var p = new(Insert_stmtContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_insert_stmt
	return p
}

func InitEmptyInsert_stmtContext(p *Insert_stmtContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_insert_stmt
}

func (*Insert_stmtContext) IsInsert_stmtContext() {}

func NewInsert_stmtContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Insert_stmtContext {
	var p = new(Insert_stmtContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SQLParserRULE_insert_stmt

	return p
}

func (s *Insert_stmtContext) GetParser() antlr.Parser { return s.parser }

func (s *Insert_stmtContext) Insert_core() IInsert_coreContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IInsert_coreContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IInsert_coreContext)
}

func (s *Insert_stmtContext) Common_table_stmt() ICommon_table_stmtContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ICommon_table_stmtContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ICommon_table_stmtContext)
}

func (s *Insert_stmtContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Insert_stmtContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Insert_stmtContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SQLParserVisitor:
		return t.VisitInsert_stmt(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SQLParser) Insert_stmt() (localctx IInsert_stmtContext) {
	localctx = NewInsert_stmtContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 52, SQLParserRULE_insert_stmt)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(438)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == SQLParserWITH_ {
		{
			p.SetState(437)
			p.Common_table_stmt()
		}

	}
	{
		p.SetState(440)
		p.Insert_core()
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

// IReturning_clauseContext is an interface to support dynamic dispatch.
type IReturning_clauseContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	RETURNING_() antlr.TerminalNode
	AllReturning_clause_result_column() []IReturning_clause_result_columnContext
	Returning_clause_result_column(i int) IReturning_clause_result_columnContext
	AllCOMMA() []antlr.TerminalNode
	COMMA(i int) antlr.TerminalNode

	// IsReturning_clauseContext differentiates from other interfaces.
	IsReturning_clauseContext()
}

type Returning_clauseContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyReturning_clauseContext() *Returning_clauseContext {
	var p = new(Returning_clauseContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_returning_clause
	return p
}

func InitEmptyReturning_clauseContext(p *Returning_clauseContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_returning_clause
}

func (*Returning_clauseContext) IsReturning_clauseContext() {}

func NewReturning_clauseContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Returning_clauseContext {
	var p = new(Returning_clauseContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SQLParserRULE_returning_clause

	return p
}

func (s *Returning_clauseContext) GetParser() antlr.Parser { return s.parser }

func (s *Returning_clauseContext) RETURNING_() antlr.TerminalNode {
	return s.GetToken(SQLParserRETURNING_, 0)
}

func (s *Returning_clauseContext) AllReturning_clause_result_column() []IReturning_clause_result_columnContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IReturning_clause_result_columnContext); ok {
			len++
		}
	}

	tst := make([]IReturning_clause_result_columnContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IReturning_clause_result_columnContext); ok {
			tst[i] = t.(IReturning_clause_result_columnContext)
			i++
		}
	}

	return tst
}

func (s *Returning_clauseContext) Returning_clause_result_column(i int) IReturning_clause_result_columnContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IReturning_clause_result_columnContext); ok {
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

	return t.(IReturning_clause_result_columnContext)
}

func (s *Returning_clauseContext) AllCOMMA() []antlr.TerminalNode {
	return s.GetTokens(SQLParserCOMMA)
}

func (s *Returning_clauseContext) COMMA(i int) antlr.TerminalNode {
	return s.GetToken(SQLParserCOMMA, i)
}

func (s *Returning_clauseContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Returning_clauseContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Returning_clauseContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SQLParserVisitor:
		return t.VisitReturning_clause(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SQLParser) Returning_clause() (localctx IReturning_clauseContext) {
	localctx = NewReturning_clauseContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 54, SQLParserRULE_returning_clause)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(442)
		p.Match(SQLParserRETURNING_)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(443)
		p.Returning_clause_result_column()
	}
	p.SetState(448)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == SQLParserCOMMA {
		{
			p.SetState(444)
			p.Match(SQLParserCOMMA)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(445)
			p.Returning_clause_result_column()
		}

		p.SetState(450)
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

// IUpsert_updateContext is an interface to support dynamic dispatch.
type IUpsert_updateContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	ASSIGN() antlr.TerminalNode
	Expr() IExprContext
	Column_name() IColumn_nameContext
	Column_name_list() IColumn_name_listContext

	// IsUpsert_updateContext differentiates from other interfaces.
	IsUpsert_updateContext()
}

type Upsert_updateContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyUpsert_updateContext() *Upsert_updateContext {
	var p = new(Upsert_updateContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_upsert_update
	return p
}

func InitEmptyUpsert_updateContext(p *Upsert_updateContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_upsert_update
}

func (*Upsert_updateContext) IsUpsert_updateContext() {}

func NewUpsert_updateContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Upsert_updateContext {
	var p = new(Upsert_updateContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SQLParserRULE_upsert_update

	return p
}

func (s *Upsert_updateContext) GetParser() antlr.Parser { return s.parser }

func (s *Upsert_updateContext) ASSIGN() antlr.TerminalNode {
	return s.GetToken(SQLParserASSIGN, 0)
}

func (s *Upsert_updateContext) Expr() IExprContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExprContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExprContext)
}

func (s *Upsert_updateContext) Column_name() IColumn_nameContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IColumn_nameContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IColumn_nameContext)
}

func (s *Upsert_updateContext) Column_name_list() IColumn_name_listContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IColumn_name_listContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IColumn_name_listContext)
}

func (s *Upsert_updateContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Upsert_updateContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Upsert_updateContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SQLParserVisitor:
		return t.VisitUpsert_update(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SQLParser) Upsert_update() (localctx IUpsert_updateContext) {
	localctx = NewUpsert_updateContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 56, SQLParserRULE_upsert_update)
	p.EnterOuterAlt(localctx, 1)
	p.SetState(453)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case SQLParserIDENTIFIER:
		{
			p.SetState(451)
			p.Column_name()
		}

	case SQLParserOPEN_PAR:
		{
			p.SetState(452)
			p.Column_name_list()
		}

	default:
		p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
		goto errorExit
	}
	{
		p.SetState(455)
		p.Match(SQLParserASSIGN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(456)
		p.expr(0)
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

// IUpsert_clauseContext is an interface to support dynamic dispatch.
type IUpsert_clauseContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// GetTarget_expr returns the target_expr rule contexts.
	GetTarget_expr() IExprContext

	// GetUpdate_expr returns the update_expr rule contexts.
	GetUpdate_expr() IExprContext

	// SetTarget_expr sets the target_expr rule contexts.
	SetTarget_expr(IExprContext)

	// SetUpdate_expr sets the update_expr rule contexts.
	SetUpdate_expr(IExprContext)

	// Getter signatures
	ON_() antlr.TerminalNode
	CONFLICT_() antlr.TerminalNode
	DO_() antlr.TerminalNode
	NOTHING_() antlr.TerminalNode
	UPDATE_() antlr.TerminalNode
	SET_() antlr.TerminalNode
	OPEN_PAR() antlr.TerminalNode
	AllIndexed_column() []IIndexed_columnContext
	Indexed_column(i int) IIndexed_columnContext
	CLOSE_PAR() antlr.TerminalNode
	AllUpsert_update() []IUpsert_updateContext
	Upsert_update(i int) IUpsert_updateContext
	AllCOMMA() []antlr.TerminalNode
	COMMA(i int) antlr.TerminalNode
	AllWHERE_() []antlr.TerminalNode
	WHERE_(i int) antlr.TerminalNode
	AllExpr() []IExprContext
	Expr(i int) IExprContext

	// IsUpsert_clauseContext differentiates from other interfaces.
	IsUpsert_clauseContext()
}

type Upsert_clauseContext struct {
	antlr.BaseParserRuleContext
	parser      antlr.Parser
	target_expr IExprContext
	update_expr IExprContext
}

func NewEmptyUpsert_clauseContext() *Upsert_clauseContext {
	var p = new(Upsert_clauseContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_upsert_clause
	return p
}

func InitEmptyUpsert_clauseContext(p *Upsert_clauseContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_upsert_clause
}

func (*Upsert_clauseContext) IsUpsert_clauseContext() {}

func NewUpsert_clauseContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Upsert_clauseContext {
	var p = new(Upsert_clauseContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SQLParserRULE_upsert_clause

	return p
}

func (s *Upsert_clauseContext) GetParser() antlr.Parser { return s.parser }

func (s *Upsert_clauseContext) GetTarget_expr() IExprContext { return s.target_expr }

func (s *Upsert_clauseContext) GetUpdate_expr() IExprContext { return s.update_expr }

func (s *Upsert_clauseContext) SetTarget_expr(v IExprContext) { s.target_expr = v }

func (s *Upsert_clauseContext) SetUpdate_expr(v IExprContext) { s.update_expr = v }

func (s *Upsert_clauseContext) ON_() antlr.TerminalNode {
	return s.GetToken(SQLParserON_, 0)
}

func (s *Upsert_clauseContext) CONFLICT_() antlr.TerminalNode {
	return s.GetToken(SQLParserCONFLICT_, 0)
}

func (s *Upsert_clauseContext) DO_() antlr.TerminalNode {
	return s.GetToken(SQLParserDO_, 0)
}

func (s *Upsert_clauseContext) NOTHING_() antlr.TerminalNode {
	return s.GetToken(SQLParserNOTHING_, 0)
}

func (s *Upsert_clauseContext) UPDATE_() antlr.TerminalNode {
	return s.GetToken(SQLParserUPDATE_, 0)
}

func (s *Upsert_clauseContext) SET_() antlr.TerminalNode {
	return s.GetToken(SQLParserSET_, 0)
}

func (s *Upsert_clauseContext) OPEN_PAR() antlr.TerminalNode {
	return s.GetToken(SQLParserOPEN_PAR, 0)
}

func (s *Upsert_clauseContext) AllIndexed_column() []IIndexed_columnContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IIndexed_columnContext); ok {
			len++
		}
	}

	tst := make([]IIndexed_columnContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IIndexed_columnContext); ok {
			tst[i] = t.(IIndexed_columnContext)
			i++
		}
	}

	return tst
}

func (s *Upsert_clauseContext) Indexed_column(i int) IIndexed_columnContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IIndexed_columnContext); ok {
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

	return t.(IIndexed_columnContext)
}

func (s *Upsert_clauseContext) CLOSE_PAR() antlr.TerminalNode {
	return s.GetToken(SQLParserCLOSE_PAR, 0)
}

func (s *Upsert_clauseContext) AllUpsert_update() []IUpsert_updateContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IUpsert_updateContext); ok {
			len++
		}
	}

	tst := make([]IUpsert_updateContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IUpsert_updateContext); ok {
			tst[i] = t.(IUpsert_updateContext)
			i++
		}
	}

	return tst
}

func (s *Upsert_clauseContext) Upsert_update(i int) IUpsert_updateContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IUpsert_updateContext); ok {
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

	return t.(IUpsert_updateContext)
}

func (s *Upsert_clauseContext) AllCOMMA() []antlr.TerminalNode {
	return s.GetTokens(SQLParserCOMMA)
}

func (s *Upsert_clauseContext) COMMA(i int) antlr.TerminalNode {
	return s.GetToken(SQLParserCOMMA, i)
}

func (s *Upsert_clauseContext) AllWHERE_() []antlr.TerminalNode {
	return s.GetTokens(SQLParserWHERE_)
}

func (s *Upsert_clauseContext) WHERE_(i int) antlr.TerminalNode {
	return s.GetToken(SQLParserWHERE_, i)
}

func (s *Upsert_clauseContext) AllExpr() []IExprContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IExprContext); ok {
			len++
		}
	}

	tst := make([]IExprContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IExprContext); ok {
			tst[i] = t.(IExprContext)
			i++
		}
	}

	return tst
}

func (s *Upsert_clauseContext) Expr(i int) IExprContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExprContext); ok {
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

	return t.(IExprContext)
}

func (s *Upsert_clauseContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Upsert_clauseContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Upsert_clauseContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SQLParserVisitor:
		return t.VisitUpsert_clause(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SQLParser) Upsert_clause() (localctx IUpsert_clauseContext) {
	localctx = NewUpsert_clauseContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 58, SQLParserRULE_upsert_clause)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(458)
		p.Match(SQLParserON_)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(459)
		p.Match(SQLParserCONFLICT_)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(474)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == SQLParserOPEN_PAR {
		{
			p.SetState(460)
			p.Match(SQLParserOPEN_PAR)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(461)
			p.Indexed_column()
		}
		p.SetState(466)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		for _la == SQLParserCOMMA {
			{
				p.SetState(462)
				p.Match(SQLParserCOMMA)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(463)
				p.Indexed_column()
			}

			p.SetState(468)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)
		}
		{
			p.SetState(469)
			p.Match(SQLParserCLOSE_PAR)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(472)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == SQLParserWHERE_ {
			{
				p.SetState(470)
				p.Match(SQLParserWHERE_)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(471)

				var _x = p.expr(0)

				localctx.(*Upsert_clauseContext).target_expr = _x
			}

		}

	}
	{
		p.SetState(476)
		p.Match(SQLParserDO_)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(492)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case SQLParserNOTHING_:
		{
			p.SetState(477)
			p.Match(SQLParserNOTHING_)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case SQLParserUPDATE_:
		{
			p.SetState(478)
			p.Match(SQLParserUPDATE_)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(479)
			p.Match(SQLParserSET_)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

		{
			p.SetState(480)
			p.Upsert_update()
		}
		p.SetState(485)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		for _la == SQLParserCOMMA {
			{
				p.SetState(481)
				p.Match(SQLParserCOMMA)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(482)
				p.Upsert_update()
			}

			p.SetState(487)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)
		}
		p.SetState(490)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == SQLParserWHERE_ {
			{
				p.SetState(488)
				p.Match(SQLParserWHERE_)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(489)

				var _x = p.expr(0)

				localctx.(*Upsert_clauseContext).update_expr = _x
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

// ISelect_stmt_no_cteContext is an interface to support dynamic dispatch.
type ISelect_stmt_no_cteContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllSelect_core() []ISelect_coreContext
	Select_core(i int) ISelect_coreContext
	AllCompound_operator() []ICompound_operatorContext
	Compound_operator(i int) ICompound_operatorContext
	Order_by_stmt() IOrder_by_stmtContext
	Limit_stmt() ILimit_stmtContext

	// IsSelect_stmt_no_cteContext differentiates from other interfaces.
	IsSelect_stmt_no_cteContext()
}

type Select_stmt_no_cteContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptySelect_stmt_no_cteContext() *Select_stmt_no_cteContext {
	var p = new(Select_stmt_no_cteContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_select_stmt_no_cte
	return p
}

func InitEmptySelect_stmt_no_cteContext(p *Select_stmt_no_cteContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_select_stmt_no_cte
}

func (*Select_stmt_no_cteContext) IsSelect_stmt_no_cteContext() {}

func NewSelect_stmt_no_cteContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Select_stmt_no_cteContext {
	var p = new(Select_stmt_no_cteContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SQLParserRULE_select_stmt_no_cte

	return p
}

func (s *Select_stmt_no_cteContext) GetParser() antlr.Parser { return s.parser }

func (s *Select_stmt_no_cteContext) AllSelect_core() []ISelect_coreContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(ISelect_coreContext); ok {
			len++
		}
	}

	tst := make([]ISelect_coreContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(ISelect_coreContext); ok {
			tst[i] = t.(ISelect_coreContext)
			i++
		}
	}

	return tst
}

func (s *Select_stmt_no_cteContext) Select_core(i int) ISelect_coreContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ISelect_coreContext); ok {
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

	return t.(ISelect_coreContext)
}

func (s *Select_stmt_no_cteContext) AllCompound_operator() []ICompound_operatorContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(ICompound_operatorContext); ok {
			len++
		}
	}

	tst := make([]ICompound_operatorContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(ICompound_operatorContext); ok {
			tst[i] = t.(ICompound_operatorContext)
			i++
		}
	}

	return tst
}

func (s *Select_stmt_no_cteContext) Compound_operator(i int) ICompound_operatorContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ICompound_operatorContext); ok {
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

	return t.(ICompound_operatorContext)
}

func (s *Select_stmt_no_cteContext) Order_by_stmt() IOrder_by_stmtContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IOrder_by_stmtContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IOrder_by_stmtContext)
}

func (s *Select_stmt_no_cteContext) Limit_stmt() ILimit_stmtContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ILimit_stmtContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ILimit_stmtContext)
}

func (s *Select_stmt_no_cteContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Select_stmt_no_cteContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Select_stmt_no_cteContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SQLParserVisitor:
		return t.VisitSelect_stmt_no_cte(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SQLParser) Select_stmt_no_cte() (localctx ISelect_stmt_no_cteContext) {
	localctx = NewSelect_stmt_no_cteContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 60, SQLParserRULE_select_stmt_no_cte)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(494)
		p.Select_core()
	}
	p.SetState(500)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for (int64((_la-40)) & ^0x3f) == 0 && ((int64(1)<<(_la-40))&2199023259649) != 0 {
		{
			p.SetState(495)
			p.Compound_operator()
		}
		{
			p.SetState(496)
			p.Select_core()
		}

		p.SetState(502)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	p.SetState(504)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == SQLParserORDER_ {
		{
			p.SetState(503)
			p.Order_by_stmt()
		}

	}
	p.SetState(507)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == SQLParserLIMIT_ {
		{
			p.SetState(506)
			p.Limit_stmt()
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

// ISelect_stmtContext is an interface to support dynamic dispatch.
type ISelect_stmtContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Select_stmt_no_cte() ISelect_stmt_no_cteContext
	Common_table_stmt() ICommon_table_stmtContext

	// IsSelect_stmtContext differentiates from other interfaces.
	IsSelect_stmtContext()
}

type Select_stmtContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptySelect_stmtContext() *Select_stmtContext {
	var p = new(Select_stmtContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_select_stmt
	return p
}

func InitEmptySelect_stmtContext(p *Select_stmtContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_select_stmt
}

func (*Select_stmtContext) IsSelect_stmtContext() {}

func NewSelect_stmtContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Select_stmtContext {
	var p = new(Select_stmtContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SQLParserRULE_select_stmt

	return p
}

func (s *Select_stmtContext) GetParser() antlr.Parser { return s.parser }

func (s *Select_stmtContext) Select_stmt_no_cte() ISelect_stmt_no_cteContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ISelect_stmt_no_cteContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ISelect_stmt_no_cteContext)
}

func (s *Select_stmtContext) Common_table_stmt() ICommon_table_stmtContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ICommon_table_stmtContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ICommon_table_stmtContext)
}

func (s *Select_stmtContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Select_stmtContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Select_stmtContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SQLParserVisitor:
		return t.VisitSelect_stmt(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SQLParser) Select_stmt() (localctx ISelect_stmtContext) {
	localctx = NewSelect_stmtContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 62, SQLParserRULE_select_stmt)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(510)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == SQLParserWITH_ {
		{
			p.SetState(509)
			p.Common_table_stmt()
		}

	}
	{
		p.SetState(512)
		p.Select_stmt_no_cte()
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

// IJoin_relationContext is an interface to support dynamic dispatch.
type IJoin_relationContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// GetRight_relation returns the right_relation rule contexts.
	GetRight_relation() ITable_or_subqueryContext

	// SetRight_relation sets the right_relation rule contexts.
	SetRight_relation(ITable_or_subqueryContext)

	// Getter signatures
	Join_operator() IJoin_operatorContext
	Join_constraint() IJoin_constraintContext
	Table_or_subquery() ITable_or_subqueryContext

	// IsJoin_relationContext differentiates from other interfaces.
	IsJoin_relationContext()
}

type Join_relationContext struct {
	antlr.BaseParserRuleContext
	parser         antlr.Parser
	right_relation ITable_or_subqueryContext
}

func NewEmptyJoin_relationContext() *Join_relationContext {
	var p = new(Join_relationContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_join_relation
	return p
}

func InitEmptyJoin_relationContext(p *Join_relationContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_join_relation
}

func (*Join_relationContext) IsJoin_relationContext() {}

func NewJoin_relationContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Join_relationContext {
	var p = new(Join_relationContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SQLParserRULE_join_relation

	return p
}

func (s *Join_relationContext) GetParser() antlr.Parser { return s.parser }

func (s *Join_relationContext) GetRight_relation() ITable_or_subqueryContext { return s.right_relation }

func (s *Join_relationContext) SetRight_relation(v ITable_or_subqueryContext) { s.right_relation = v }

func (s *Join_relationContext) Join_operator() IJoin_operatorContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IJoin_operatorContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IJoin_operatorContext)
}

func (s *Join_relationContext) Join_constraint() IJoin_constraintContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IJoin_constraintContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IJoin_constraintContext)
}

func (s *Join_relationContext) Table_or_subquery() ITable_or_subqueryContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITable_or_subqueryContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ITable_or_subqueryContext)
}

func (s *Join_relationContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Join_relationContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Join_relationContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SQLParserVisitor:
		return t.VisitJoin_relation(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SQLParser) Join_relation() (localctx IJoin_relationContext) {
	localctx = NewJoin_relationContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 64, SQLParserRULE_join_relation)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(514)
		p.Join_operator()
	}
	{
		p.SetState(515)

		var _x = p.Table_or_subquery()

		localctx.(*Join_relationContext).right_relation = _x
	}
	{
		p.SetState(516)
		p.Join_constraint()
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

// IRelationContext is an interface to support dynamic dispatch.
type IRelationContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Table_or_subquery() ITable_or_subqueryContext
	AllJoin_relation() []IJoin_relationContext
	Join_relation(i int) IJoin_relationContext

	// IsRelationContext differentiates from other interfaces.
	IsRelationContext()
}

type RelationContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyRelationContext() *RelationContext {
	var p = new(RelationContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_relation
	return p
}

func InitEmptyRelationContext(p *RelationContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_relation
}

func (*RelationContext) IsRelationContext() {}

func NewRelationContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *RelationContext {
	var p = new(RelationContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SQLParserRULE_relation

	return p
}

func (s *RelationContext) GetParser() antlr.Parser { return s.parser }

func (s *RelationContext) Table_or_subquery() ITable_or_subqueryContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITable_or_subqueryContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ITable_or_subqueryContext)
}

func (s *RelationContext) AllJoin_relation() []IJoin_relationContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IJoin_relationContext); ok {
			len++
		}
	}

	tst := make([]IJoin_relationContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IJoin_relationContext); ok {
			tst[i] = t.(IJoin_relationContext)
			i++
		}
	}

	return tst
}

func (s *RelationContext) Join_relation(i int) IJoin_relationContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IJoin_relationContext); ok {
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

	return t.(IJoin_relationContext)
}

func (s *RelationContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *RelationContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *RelationContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SQLParserVisitor:
		return t.VisitRelation(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SQLParser) Relation() (localctx IRelationContext) {
	localctx = NewRelationContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 66, SQLParserRULE_relation)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(518)
		p.Table_or_subquery()
	}
	p.SetState(522)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for (int64((_la-46)) & ^0x3f) == 0 && ((int64(1)<<(_la-46))&1073752081) != 0 {
		{
			p.SetState(519)
			p.Join_relation()
		}

		p.SetState(524)
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

// ISelect_coreContext is an interface to support dynamic dispatch.
type ISelect_coreContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// GetWhereExpr returns the whereExpr rule contexts.
	GetWhereExpr() IExprContext

	// Get_expr returns the _expr rule contexts.
	Get_expr() IExprContext

	// GetHavingExpr returns the havingExpr rule contexts.
	GetHavingExpr() IExprContext

	// SetWhereExpr sets the whereExpr rule contexts.
	SetWhereExpr(IExprContext)

	// Set_expr sets the _expr rule contexts.
	Set_expr(IExprContext)

	// SetHavingExpr sets the havingExpr rule contexts.
	SetHavingExpr(IExprContext)

	// GetGroupByExpr returns the groupByExpr rule context list.
	GetGroupByExpr() []IExprContext

	// SetGroupByExpr sets the groupByExpr rule context list.
	SetGroupByExpr([]IExprContext)

	// Getter signatures
	SELECT_() antlr.TerminalNode
	AllResult_column() []IResult_columnContext
	Result_column(i int) IResult_columnContext
	DISTINCT_() antlr.TerminalNode
	AllCOMMA() []antlr.TerminalNode
	COMMA(i int) antlr.TerminalNode
	FROM_() antlr.TerminalNode
	Relation() IRelationContext
	WHERE_() antlr.TerminalNode
	GROUP_() antlr.TerminalNode
	BY_() antlr.TerminalNode
	AllExpr() []IExprContext
	Expr(i int) IExprContext
	HAVING_() antlr.TerminalNode

	// IsSelect_coreContext differentiates from other interfaces.
	IsSelect_coreContext()
}

type Select_coreContext struct {
	antlr.BaseParserRuleContext
	parser      antlr.Parser
	whereExpr   IExprContext
	_expr       IExprContext
	groupByExpr []IExprContext
	havingExpr  IExprContext
}

func NewEmptySelect_coreContext() *Select_coreContext {
	var p = new(Select_coreContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_select_core
	return p
}

func InitEmptySelect_coreContext(p *Select_coreContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_select_core
}

func (*Select_coreContext) IsSelect_coreContext() {}

func NewSelect_coreContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Select_coreContext {
	var p = new(Select_coreContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SQLParserRULE_select_core

	return p
}

func (s *Select_coreContext) GetParser() antlr.Parser { return s.parser }

func (s *Select_coreContext) GetWhereExpr() IExprContext { return s.whereExpr }

func (s *Select_coreContext) Get_expr() IExprContext { return s._expr }

func (s *Select_coreContext) GetHavingExpr() IExprContext { return s.havingExpr }

func (s *Select_coreContext) SetWhereExpr(v IExprContext) { s.whereExpr = v }

func (s *Select_coreContext) Set_expr(v IExprContext) { s._expr = v }

func (s *Select_coreContext) SetHavingExpr(v IExprContext) { s.havingExpr = v }

func (s *Select_coreContext) GetGroupByExpr() []IExprContext { return s.groupByExpr }

func (s *Select_coreContext) SetGroupByExpr(v []IExprContext) { s.groupByExpr = v }

func (s *Select_coreContext) SELECT_() antlr.TerminalNode {
	return s.GetToken(SQLParserSELECT_, 0)
}

func (s *Select_coreContext) AllResult_column() []IResult_columnContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IResult_columnContext); ok {
			len++
		}
	}

	tst := make([]IResult_columnContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IResult_columnContext); ok {
			tst[i] = t.(IResult_columnContext)
			i++
		}
	}

	return tst
}

func (s *Select_coreContext) Result_column(i int) IResult_columnContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IResult_columnContext); ok {
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

	return t.(IResult_columnContext)
}

func (s *Select_coreContext) DISTINCT_() antlr.TerminalNode {
	return s.GetToken(SQLParserDISTINCT_, 0)
}

func (s *Select_coreContext) AllCOMMA() []antlr.TerminalNode {
	return s.GetTokens(SQLParserCOMMA)
}

func (s *Select_coreContext) COMMA(i int) antlr.TerminalNode {
	return s.GetToken(SQLParserCOMMA, i)
}

func (s *Select_coreContext) FROM_() antlr.TerminalNode {
	return s.GetToken(SQLParserFROM_, 0)
}

func (s *Select_coreContext) Relation() IRelationContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IRelationContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IRelationContext)
}

func (s *Select_coreContext) WHERE_() antlr.TerminalNode {
	return s.GetToken(SQLParserWHERE_, 0)
}

func (s *Select_coreContext) GROUP_() antlr.TerminalNode {
	return s.GetToken(SQLParserGROUP_, 0)
}

func (s *Select_coreContext) BY_() antlr.TerminalNode {
	return s.GetToken(SQLParserBY_, 0)
}

func (s *Select_coreContext) AllExpr() []IExprContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IExprContext); ok {
			len++
		}
	}

	tst := make([]IExprContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IExprContext); ok {
			tst[i] = t.(IExprContext)
			i++
		}
	}

	return tst
}

func (s *Select_coreContext) Expr(i int) IExprContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExprContext); ok {
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

	return t.(IExprContext)
}

func (s *Select_coreContext) HAVING_() antlr.TerminalNode {
	return s.GetToken(SQLParserHAVING_, 0)
}

func (s *Select_coreContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Select_coreContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Select_coreContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SQLParserVisitor:
		return t.VisitSelect_core(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SQLParser) Select_core() (localctx ISelect_coreContext) {
	localctx = NewSelect_coreContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 68, SQLParserRULE_select_core)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(525)
		p.Match(SQLParserSELECT_)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(527)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == SQLParserDISTINCT_ {
		{
			p.SetState(526)
			p.Match(SQLParserDISTINCT_)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	}
	{
		p.SetState(529)
		p.Result_column()
	}
	p.SetState(534)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == SQLParserCOMMA {
		{
			p.SetState(530)
			p.Match(SQLParserCOMMA)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(531)
			p.Result_column()
		}

		p.SetState(536)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	p.SetState(539)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == SQLParserFROM_ {
		{
			p.SetState(537)
			p.Match(SQLParserFROM_)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(538)
			p.Relation()
		}

	}
	p.SetState(543)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == SQLParserWHERE_ {
		{
			p.SetState(541)
			p.Match(SQLParserWHERE_)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(542)

			var _x = p.expr(0)

			localctx.(*Select_coreContext).whereExpr = _x
		}

	}
	p.SetState(559)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == SQLParserGROUP_ {
		{
			p.SetState(545)
			p.Match(SQLParserGROUP_)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(546)
			p.Match(SQLParserBY_)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(547)

			var _x = p.expr(0)

			localctx.(*Select_coreContext)._expr = _x
		}
		localctx.(*Select_coreContext).groupByExpr = append(localctx.(*Select_coreContext).groupByExpr, localctx.(*Select_coreContext)._expr)
		p.SetState(552)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		for _la == SQLParserCOMMA {
			{
				p.SetState(548)
				p.Match(SQLParserCOMMA)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(549)

				var _x = p.expr(0)

				localctx.(*Select_coreContext)._expr = _x
			}
			localctx.(*Select_coreContext).groupByExpr = append(localctx.(*Select_coreContext).groupByExpr, localctx.(*Select_coreContext)._expr)

			p.SetState(554)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)
		}
		p.SetState(557)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == SQLParserHAVING_ {
			{
				p.SetState(555)
				p.Match(SQLParserHAVING_)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(556)

				var _x = p.expr(0)

				localctx.(*Select_coreContext).havingExpr = _x
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

// ITable_or_subqueryContext is an interface to support dynamic dispatch.
type ITable_or_subqueryContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Table_name() ITable_nameContext
	AS_() antlr.TerminalNode
	Table_alias() ITable_aliasContext
	OPEN_PAR() antlr.TerminalNode
	Select_stmt_no_cte() ISelect_stmt_no_cteContext
	CLOSE_PAR() antlr.TerminalNode

	// IsTable_or_subqueryContext differentiates from other interfaces.
	IsTable_or_subqueryContext()
}

type Table_or_subqueryContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyTable_or_subqueryContext() *Table_or_subqueryContext {
	var p = new(Table_or_subqueryContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_table_or_subquery
	return p
}

func InitEmptyTable_or_subqueryContext(p *Table_or_subqueryContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_table_or_subquery
}

func (*Table_or_subqueryContext) IsTable_or_subqueryContext() {}

func NewTable_or_subqueryContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Table_or_subqueryContext {
	var p = new(Table_or_subqueryContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SQLParserRULE_table_or_subquery

	return p
}

func (s *Table_or_subqueryContext) GetParser() antlr.Parser { return s.parser }

func (s *Table_or_subqueryContext) Table_name() ITable_nameContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITable_nameContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ITable_nameContext)
}

func (s *Table_or_subqueryContext) AS_() antlr.TerminalNode {
	return s.GetToken(SQLParserAS_, 0)
}

func (s *Table_or_subqueryContext) Table_alias() ITable_aliasContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITable_aliasContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ITable_aliasContext)
}

func (s *Table_or_subqueryContext) OPEN_PAR() antlr.TerminalNode {
	return s.GetToken(SQLParserOPEN_PAR, 0)
}

func (s *Table_or_subqueryContext) Select_stmt_no_cte() ISelect_stmt_no_cteContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ISelect_stmt_no_cteContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ISelect_stmt_no_cteContext)
}

func (s *Table_or_subqueryContext) CLOSE_PAR() antlr.TerminalNode {
	return s.GetToken(SQLParserCLOSE_PAR, 0)
}

func (s *Table_or_subqueryContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Table_or_subqueryContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Table_or_subqueryContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SQLParserVisitor:
		return t.VisitTable_or_subquery(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SQLParser) Table_or_subquery() (localctx ITable_or_subqueryContext) {
	localctx = NewTable_or_subqueryContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 70, SQLParserRULE_table_or_subquery)
	var _la int

	p.SetState(573)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case SQLParserIDENTIFIER:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(561)
			p.Table_name()
		}
		p.SetState(564)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == SQLParserAS_ {
			{
				p.SetState(562)
				p.Match(SQLParserAS_)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(563)
				p.Table_alias()
			}

		}

	case SQLParserOPEN_PAR:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(566)
			p.Match(SQLParserOPEN_PAR)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(567)
			p.Select_stmt_no_cte()
		}
		{
			p.SetState(568)
			p.Match(SQLParserCLOSE_PAR)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(571)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == SQLParserAS_ {
			{
				p.SetState(569)
				p.Match(SQLParserAS_)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(570)
				p.Table_alias()
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

// IResult_columnContext is an interface to support dynamic dispatch.
type IResult_columnContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	STAR() antlr.TerminalNode
	Table_name() ITable_nameContext
	DOT() antlr.TerminalNode
	Expr() IExprContext
	AS_() antlr.TerminalNode
	Column_alias() IColumn_aliasContext

	// IsResult_columnContext differentiates from other interfaces.
	IsResult_columnContext()
}

type Result_columnContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyResult_columnContext() *Result_columnContext {
	var p = new(Result_columnContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_result_column
	return p
}

func InitEmptyResult_columnContext(p *Result_columnContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_result_column
}

func (*Result_columnContext) IsResult_columnContext() {}

func NewResult_columnContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Result_columnContext {
	var p = new(Result_columnContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SQLParserRULE_result_column

	return p
}

func (s *Result_columnContext) GetParser() antlr.Parser { return s.parser }

func (s *Result_columnContext) STAR() antlr.TerminalNode {
	return s.GetToken(SQLParserSTAR, 0)
}

func (s *Result_columnContext) Table_name() ITable_nameContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITable_nameContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ITable_nameContext)
}

func (s *Result_columnContext) DOT() antlr.TerminalNode {
	return s.GetToken(SQLParserDOT, 0)
}

func (s *Result_columnContext) Expr() IExprContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExprContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExprContext)
}

func (s *Result_columnContext) AS_() antlr.TerminalNode {
	return s.GetToken(SQLParserAS_, 0)
}

func (s *Result_columnContext) Column_alias() IColumn_aliasContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IColumn_aliasContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IColumn_aliasContext)
}

func (s *Result_columnContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Result_columnContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Result_columnContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SQLParserVisitor:
		return t.VisitResult_column(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SQLParser) Result_column() (localctx IResult_columnContext) {
	localctx = NewResult_columnContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 72, SQLParserRULE_result_column)
	var _la int

	p.SetState(585)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 70, p.GetParserRuleContext()) {
	case 1:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(575)
			p.Match(SQLParserSTAR)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 2:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(576)
			p.Table_name()
		}
		{
			p.SetState(577)
			p.Match(SQLParserDOT)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(578)
			p.Match(SQLParserSTAR)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 3:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(580)
			p.expr(0)
		}
		p.SetState(583)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == SQLParserAS_ {
			{
				p.SetState(581)
				p.Match(SQLParserAS_)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(582)
				p.Column_alias()
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

// IReturning_clause_result_columnContext is an interface to support dynamic dispatch.
type IReturning_clause_result_columnContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	STAR() antlr.TerminalNode
	Expr() IExprContext
	AS_() antlr.TerminalNode
	Column_alias() IColumn_aliasContext

	// IsReturning_clause_result_columnContext differentiates from other interfaces.
	IsReturning_clause_result_columnContext()
}

type Returning_clause_result_columnContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyReturning_clause_result_columnContext() *Returning_clause_result_columnContext {
	var p = new(Returning_clause_result_columnContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_returning_clause_result_column
	return p
}

func InitEmptyReturning_clause_result_columnContext(p *Returning_clause_result_columnContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_returning_clause_result_column
}

func (*Returning_clause_result_columnContext) IsReturning_clause_result_columnContext() {}

func NewReturning_clause_result_columnContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Returning_clause_result_columnContext {
	var p = new(Returning_clause_result_columnContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SQLParserRULE_returning_clause_result_column

	return p
}

func (s *Returning_clause_result_columnContext) GetParser() antlr.Parser { return s.parser }

func (s *Returning_clause_result_columnContext) STAR() antlr.TerminalNode {
	return s.GetToken(SQLParserSTAR, 0)
}

func (s *Returning_clause_result_columnContext) Expr() IExprContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExprContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExprContext)
}

func (s *Returning_clause_result_columnContext) AS_() antlr.TerminalNode {
	return s.GetToken(SQLParserAS_, 0)
}

func (s *Returning_clause_result_columnContext) Column_alias() IColumn_aliasContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IColumn_aliasContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IColumn_aliasContext)
}

func (s *Returning_clause_result_columnContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Returning_clause_result_columnContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Returning_clause_result_columnContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SQLParserVisitor:
		return t.VisitReturning_clause_result_column(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SQLParser) Returning_clause_result_column() (localctx IReturning_clause_result_columnContext) {
	localctx = NewReturning_clause_result_columnContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 74, SQLParserRULE_returning_clause_result_column)
	var _la int

	p.SetState(593)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case SQLParserSTAR:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(587)
			p.Match(SQLParserSTAR)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case SQLParserOPEN_PAR, SQLParserPLUS, SQLParserMINUS, SQLParserCASE_, SQLParserEXISTS_, SQLParserFALSE_, SQLParserLIKE_, SQLParserNOT_, SQLParserNULL_, SQLParserREPLACE_, SQLParserTRUE_, SQLParserIDENTIFIER, SQLParserNUMERIC_LITERAL, SQLParserBIND_PARAMETER, SQLParserSTRING_LITERAL:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(588)
			p.expr(0)
		}
		p.SetState(591)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == SQLParserAS_ {
			{
				p.SetState(589)
				p.Match(SQLParserAS_)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(590)
				p.Column_alias()
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

// IJoin_operatorContext is an interface to support dynamic dispatch.
type IJoin_operatorContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	JOIN_() antlr.TerminalNode
	INNER_() antlr.TerminalNode
	LEFT_() antlr.TerminalNode
	RIGHT_() antlr.TerminalNode
	FULL_() antlr.TerminalNode
	OUTER_() antlr.TerminalNode

	// IsJoin_operatorContext differentiates from other interfaces.
	IsJoin_operatorContext()
}

type Join_operatorContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyJoin_operatorContext() *Join_operatorContext {
	var p = new(Join_operatorContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_join_operator
	return p
}

func InitEmptyJoin_operatorContext(p *Join_operatorContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_join_operator
}

func (*Join_operatorContext) IsJoin_operatorContext() {}

func NewJoin_operatorContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Join_operatorContext {
	var p = new(Join_operatorContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SQLParserRULE_join_operator

	return p
}

func (s *Join_operatorContext) GetParser() antlr.Parser { return s.parser }

func (s *Join_operatorContext) JOIN_() antlr.TerminalNode {
	return s.GetToken(SQLParserJOIN_, 0)
}

func (s *Join_operatorContext) INNER_() antlr.TerminalNode {
	return s.GetToken(SQLParserINNER_, 0)
}

func (s *Join_operatorContext) LEFT_() antlr.TerminalNode {
	return s.GetToken(SQLParserLEFT_, 0)
}

func (s *Join_operatorContext) RIGHT_() antlr.TerminalNode {
	return s.GetToken(SQLParserRIGHT_, 0)
}

func (s *Join_operatorContext) FULL_() antlr.TerminalNode {
	return s.GetToken(SQLParserFULL_, 0)
}

func (s *Join_operatorContext) OUTER_() antlr.TerminalNode {
	return s.GetToken(SQLParserOUTER_, 0)
}

func (s *Join_operatorContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Join_operatorContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Join_operatorContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SQLParserVisitor:
		return t.VisitJoin_operator(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SQLParser) Join_operator() (localctx IJoin_operatorContext) {
	localctx = NewJoin_operatorContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 76, SQLParserRULE_join_operator)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(600)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	switch p.GetTokenStream().LA(1) {
	case SQLParserFULL_, SQLParserLEFT_, SQLParserRIGHT_:
		{
			p.SetState(595)
			_la = p.GetTokenStream().LA(1)

			if !((int64((_la-46)) & ^0x3f) == 0 && ((int64(1)<<(_la-46))&1073750017) != 0) {
				p.GetErrorHandler().RecoverInline(p)
			} else {
				p.GetErrorHandler().ReportMatch(p)
				p.Consume()
			}
		}
		p.SetState(597)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == SQLParserOUTER_ {
			{
				p.SetState(596)
				p.Match(SQLParserOUTER_)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

		}

	case SQLParserINNER_:
		{
			p.SetState(599)
			p.Match(SQLParserINNER_)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case SQLParserJOIN_:

	default:
	}
	{
		p.SetState(602)
		p.Match(SQLParserJOIN_)
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

// IJoin_constraintContext is an interface to support dynamic dispatch.
type IJoin_constraintContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	ON_() antlr.TerminalNode
	Expr() IExprContext

	// IsJoin_constraintContext differentiates from other interfaces.
	IsJoin_constraintContext()
}

type Join_constraintContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyJoin_constraintContext() *Join_constraintContext {
	var p = new(Join_constraintContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_join_constraint
	return p
}

func InitEmptyJoin_constraintContext(p *Join_constraintContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_join_constraint
}

func (*Join_constraintContext) IsJoin_constraintContext() {}

func NewJoin_constraintContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Join_constraintContext {
	var p = new(Join_constraintContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SQLParserRULE_join_constraint

	return p
}

func (s *Join_constraintContext) GetParser() antlr.Parser { return s.parser }

func (s *Join_constraintContext) ON_() antlr.TerminalNode {
	return s.GetToken(SQLParserON_, 0)
}

func (s *Join_constraintContext) Expr() IExprContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExprContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExprContext)
}

func (s *Join_constraintContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Join_constraintContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Join_constraintContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SQLParserVisitor:
		return t.VisitJoin_constraint(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SQLParser) Join_constraint() (localctx IJoin_constraintContext) {
	localctx = NewJoin_constraintContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 78, SQLParserRULE_join_constraint)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(604)
		p.Match(SQLParserON_)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(605)
		p.expr(0)
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

// ICompound_operatorContext is an interface to support dynamic dispatch.
type ICompound_operatorContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	UNION_() antlr.TerminalNode
	ALL_() antlr.TerminalNode
	INTERSECT_() antlr.TerminalNode
	EXCEPT_() antlr.TerminalNode

	// IsCompound_operatorContext differentiates from other interfaces.
	IsCompound_operatorContext()
}

type Compound_operatorContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyCompound_operatorContext() *Compound_operatorContext {
	var p = new(Compound_operatorContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_compound_operator
	return p
}

func InitEmptyCompound_operatorContext(p *Compound_operatorContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_compound_operator
}

func (*Compound_operatorContext) IsCompound_operatorContext() {}

func NewCompound_operatorContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Compound_operatorContext {
	var p = new(Compound_operatorContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SQLParserRULE_compound_operator

	return p
}

func (s *Compound_operatorContext) GetParser() antlr.Parser { return s.parser }

func (s *Compound_operatorContext) UNION_() antlr.TerminalNode {
	return s.GetToken(SQLParserUNION_, 0)
}

func (s *Compound_operatorContext) ALL_() antlr.TerminalNode {
	return s.GetToken(SQLParserALL_, 0)
}

func (s *Compound_operatorContext) INTERSECT_() antlr.TerminalNode {
	return s.GetToken(SQLParserINTERSECT_, 0)
}

func (s *Compound_operatorContext) EXCEPT_() antlr.TerminalNode {
	return s.GetToken(SQLParserEXCEPT_, 0)
}

func (s *Compound_operatorContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Compound_operatorContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Compound_operatorContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SQLParserVisitor:
		return t.VisitCompound_operator(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SQLParser) Compound_operator() (localctx ICompound_operatorContext) {
	localctx = NewCompound_operatorContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 80, SQLParserRULE_compound_operator)
	var _la int

	p.SetState(613)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case SQLParserUNION_:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(607)
			p.Match(SQLParserUNION_)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(609)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == SQLParserALL_ {
			{
				p.SetState(608)
				p.Match(SQLParserALL_)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

		}

	case SQLParserINTERSECT_:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(611)
			p.Match(SQLParserINTERSECT_)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case SQLParserEXCEPT_:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(612)
			p.Match(SQLParserEXCEPT_)
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

// IUpdate_set_subclauseContext is an interface to support dynamic dispatch.
type IUpdate_set_subclauseContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	ASSIGN() antlr.TerminalNode
	Expr() IExprContext
	Column_name() IColumn_nameContext
	Column_name_list() IColumn_name_listContext

	// IsUpdate_set_subclauseContext differentiates from other interfaces.
	IsUpdate_set_subclauseContext()
}

type Update_set_subclauseContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyUpdate_set_subclauseContext() *Update_set_subclauseContext {
	var p = new(Update_set_subclauseContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_update_set_subclause
	return p
}

func InitEmptyUpdate_set_subclauseContext(p *Update_set_subclauseContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_update_set_subclause
}

func (*Update_set_subclauseContext) IsUpdate_set_subclauseContext() {}

func NewUpdate_set_subclauseContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Update_set_subclauseContext {
	var p = new(Update_set_subclauseContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SQLParserRULE_update_set_subclause

	return p
}

func (s *Update_set_subclauseContext) GetParser() antlr.Parser { return s.parser }

func (s *Update_set_subclauseContext) ASSIGN() antlr.TerminalNode {
	return s.GetToken(SQLParserASSIGN, 0)
}

func (s *Update_set_subclauseContext) Expr() IExprContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExprContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExprContext)
}

func (s *Update_set_subclauseContext) Column_name() IColumn_nameContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IColumn_nameContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IColumn_nameContext)
}

func (s *Update_set_subclauseContext) Column_name_list() IColumn_name_listContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IColumn_name_listContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IColumn_name_listContext)
}

func (s *Update_set_subclauseContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Update_set_subclauseContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Update_set_subclauseContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SQLParserVisitor:
		return t.VisitUpdate_set_subclause(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SQLParser) Update_set_subclause() (localctx IUpdate_set_subclauseContext) {
	localctx = NewUpdate_set_subclauseContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 82, SQLParserRULE_update_set_subclause)
	p.EnterOuterAlt(localctx, 1)
	p.SetState(617)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case SQLParserIDENTIFIER:
		{
			p.SetState(615)
			p.Column_name()
		}

	case SQLParserOPEN_PAR:
		{
			p.SetState(616)
			p.Column_name_list()
		}

	default:
		p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
		goto errorExit
	}
	{
		p.SetState(619)
		p.Match(SQLParserASSIGN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(620)
		p.expr(0)
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

// IUpdate_coreContext is an interface to support dynamic dispatch.
type IUpdate_coreContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	UPDATE_() antlr.TerminalNode
	Qualified_table_name() IQualified_table_nameContext
	SET_() antlr.TerminalNode
	AllUpdate_set_subclause() []IUpdate_set_subclauseContext
	Update_set_subclause(i int) IUpdate_set_subclauseContext
	AllCOMMA() []antlr.TerminalNode
	COMMA(i int) antlr.TerminalNode
	FROM_() antlr.TerminalNode
	Relation() IRelationContext
	WHERE_() antlr.TerminalNode
	Expr() IExprContext
	Returning_clause() IReturning_clauseContext

	// IsUpdate_coreContext differentiates from other interfaces.
	IsUpdate_coreContext()
}

type Update_coreContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyUpdate_coreContext() *Update_coreContext {
	var p = new(Update_coreContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_update_core
	return p
}

func InitEmptyUpdate_coreContext(p *Update_coreContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_update_core
}

func (*Update_coreContext) IsUpdate_coreContext() {}

func NewUpdate_coreContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Update_coreContext {
	var p = new(Update_coreContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SQLParserRULE_update_core

	return p
}

func (s *Update_coreContext) GetParser() antlr.Parser { return s.parser }

func (s *Update_coreContext) UPDATE_() antlr.TerminalNode {
	return s.GetToken(SQLParserUPDATE_, 0)
}

func (s *Update_coreContext) Qualified_table_name() IQualified_table_nameContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IQualified_table_nameContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IQualified_table_nameContext)
}

func (s *Update_coreContext) SET_() antlr.TerminalNode {
	return s.GetToken(SQLParserSET_, 0)
}

func (s *Update_coreContext) AllUpdate_set_subclause() []IUpdate_set_subclauseContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IUpdate_set_subclauseContext); ok {
			len++
		}
	}

	tst := make([]IUpdate_set_subclauseContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IUpdate_set_subclauseContext); ok {
			tst[i] = t.(IUpdate_set_subclauseContext)
			i++
		}
	}

	return tst
}

func (s *Update_coreContext) Update_set_subclause(i int) IUpdate_set_subclauseContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IUpdate_set_subclauseContext); ok {
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

	return t.(IUpdate_set_subclauseContext)
}

func (s *Update_coreContext) AllCOMMA() []antlr.TerminalNode {
	return s.GetTokens(SQLParserCOMMA)
}

func (s *Update_coreContext) COMMA(i int) antlr.TerminalNode {
	return s.GetToken(SQLParserCOMMA, i)
}

func (s *Update_coreContext) FROM_() antlr.TerminalNode {
	return s.GetToken(SQLParserFROM_, 0)
}

func (s *Update_coreContext) Relation() IRelationContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IRelationContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IRelationContext)
}

func (s *Update_coreContext) WHERE_() antlr.TerminalNode {
	return s.GetToken(SQLParserWHERE_, 0)
}

func (s *Update_coreContext) Expr() IExprContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExprContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExprContext)
}

func (s *Update_coreContext) Returning_clause() IReturning_clauseContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IReturning_clauseContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IReturning_clauseContext)
}

func (s *Update_coreContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Update_coreContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Update_coreContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SQLParserVisitor:
		return t.VisitUpdate_core(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SQLParser) Update_core() (localctx IUpdate_coreContext) {
	localctx = NewUpdate_coreContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 84, SQLParserRULE_update_core)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(622)
		p.Match(SQLParserUPDATE_)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(623)
		p.Qualified_table_name()
	}
	{
		p.SetState(624)
		p.Match(SQLParserSET_)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(625)
		p.Update_set_subclause()
	}
	p.SetState(630)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == SQLParserCOMMA {
		{
			p.SetState(626)
			p.Match(SQLParserCOMMA)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(627)
			p.Update_set_subclause()
		}

		p.SetState(632)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	p.SetState(635)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == SQLParserFROM_ {
		{
			p.SetState(633)
			p.Match(SQLParserFROM_)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(634)
			p.Relation()
		}

	}
	p.SetState(639)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == SQLParserWHERE_ {
		{
			p.SetState(637)
			p.Match(SQLParserWHERE_)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(638)
			p.expr(0)
		}

	}
	p.SetState(642)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == SQLParserRETURNING_ {
		{
			p.SetState(641)
			p.Returning_clause()
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

// IUpdate_stmtContext is an interface to support dynamic dispatch.
type IUpdate_stmtContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Update_core() IUpdate_coreContext
	Common_table_stmt() ICommon_table_stmtContext

	// IsUpdate_stmtContext differentiates from other interfaces.
	IsUpdate_stmtContext()
}

type Update_stmtContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyUpdate_stmtContext() *Update_stmtContext {
	var p = new(Update_stmtContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_update_stmt
	return p
}

func InitEmptyUpdate_stmtContext(p *Update_stmtContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_update_stmt
}

func (*Update_stmtContext) IsUpdate_stmtContext() {}

func NewUpdate_stmtContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Update_stmtContext {
	var p = new(Update_stmtContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SQLParserRULE_update_stmt

	return p
}

func (s *Update_stmtContext) GetParser() antlr.Parser { return s.parser }

func (s *Update_stmtContext) Update_core() IUpdate_coreContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IUpdate_coreContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IUpdate_coreContext)
}

func (s *Update_stmtContext) Common_table_stmt() ICommon_table_stmtContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ICommon_table_stmtContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ICommon_table_stmtContext)
}

func (s *Update_stmtContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Update_stmtContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Update_stmtContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SQLParserVisitor:
		return t.VisitUpdate_stmt(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SQLParser) Update_stmt() (localctx IUpdate_stmtContext) {
	localctx = NewUpdate_stmtContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 86, SQLParserRULE_update_stmt)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(645)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == SQLParserWITH_ {
		{
			p.SetState(644)
			p.Common_table_stmt()
		}

	}
	{
		p.SetState(647)
		p.Update_core()
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

// IColumn_name_listContext is an interface to support dynamic dispatch.
type IColumn_name_listContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	OPEN_PAR() antlr.TerminalNode
	AllColumn_name() []IColumn_nameContext
	Column_name(i int) IColumn_nameContext
	CLOSE_PAR() antlr.TerminalNode
	AllCOMMA() []antlr.TerminalNode
	COMMA(i int) antlr.TerminalNode

	// IsColumn_name_listContext differentiates from other interfaces.
	IsColumn_name_listContext()
}

type Column_name_listContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyColumn_name_listContext() *Column_name_listContext {
	var p = new(Column_name_listContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_column_name_list
	return p
}

func InitEmptyColumn_name_listContext(p *Column_name_listContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_column_name_list
}

func (*Column_name_listContext) IsColumn_name_listContext() {}

func NewColumn_name_listContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Column_name_listContext {
	var p = new(Column_name_listContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SQLParserRULE_column_name_list

	return p
}

func (s *Column_name_listContext) GetParser() antlr.Parser { return s.parser }

func (s *Column_name_listContext) OPEN_PAR() antlr.TerminalNode {
	return s.GetToken(SQLParserOPEN_PAR, 0)
}

func (s *Column_name_listContext) AllColumn_name() []IColumn_nameContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IColumn_nameContext); ok {
			len++
		}
	}

	tst := make([]IColumn_nameContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IColumn_nameContext); ok {
			tst[i] = t.(IColumn_nameContext)
			i++
		}
	}

	return tst
}

func (s *Column_name_listContext) Column_name(i int) IColumn_nameContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IColumn_nameContext); ok {
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

	return t.(IColumn_nameContext)
}

func (s *Column_name_listContext) CLOSE_PAR() antlr.TerminalNode {
	return s.GetToken(SQLParserCLOSE_PAR, 0)
}

func (s *Column_name_listContext) AllCOMMA() []antlr.TerminalNode {
	return s.GetTokens(SQLParserCOMMA)
}

func (s *Column_name_listContext) COMMA(i int) antlr.TerminalNode {
	return s.GetToken(SQLParserCOMMA, i)
}

func (s *Column_name_listContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Column_name_listContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Column_name_listContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SQLParserVisitor:
		return t.VisitColumn_name_list(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SQLParser) Column_name_list() (localctx IColumn_name_listContext) {
	localctx = NewColumn_name_listContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 88, SQLParserRULE_column_name_list)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(649)
		p.Match(SQLParserOPEN_PAR)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(650)
		p.Column_name()
	}
	p.SetState(655)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == SQLParserCOMMA {
		{
			p.SetState(651)
			p.Match(SQLParserCOMMA)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(652)
			p.Column_name()
		}

		p.SetState(657)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(658)
		p.Match(SQLParserCLOSE_PAR)
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

// IQualified_table_nameContext is an interface to support dynamic dispatch.
type IQualified_table_nameContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Table_name() ITable_nameContext
	AS_() antlr.TerminalNode
	Table_alias() ITable_aliasContext

	// IsQualified_table_nameContext differentiates from other interfaces.
	IsQualified_table_nameContext()
}

type Qualified_table_nameContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyQualified_table_nameContext() *Qualified_table_nameContext {
	var p = new(Qualified_table_nameContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_qualified_table_name
	return p
}

func InitEmptyQualified_table_nameContext(p *Qualified_table_nameContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_qualified_table_name
}

func (*Qualified_table_nameContext) IsQualified_table_nameContext() {}

func NewQualified_table_nameContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Qualified_table_nameContext {
	var p = new(Qualified_table_nameContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SQLParserRULE_qualified_table_name

	return p
}

func (s *Qualified_table_nameContext) GetParser() antlr.Parser { return s.parser }

func (s *Qualified_table_nameContext) Table_name() ITable_nameContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITable_nameContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ITable_nameContext)
}

func (s *Qualified_table_nameContext) AS_() antlr.TerminalNode {
	return s.GetToken(SQLParserAS_, 0)
}

func (s *Qualified_table_nameContext) Table_alias() ITable_aliasContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITable_aliasContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ITable_aliasContext)
}

func (s *Qualified_table_nameContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Qualified_table_nameContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Qualified_table_nameContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SQLParserVisitor:
		return t.VisitQualified_table_name(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SQLParser) Qualified_table_name() (localctx IQualified_table_nameContext) {
	localctx = NewQualified_table_nameContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 90, SQLParserRULE_qualified_table_name)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(660)
		p.Table_name()
	}
	p.SetState(663)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == SQLParserAS_ {
		{
			p.SetState(661)
			p.Match(SQLParserAS_)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(662)
			p.Table_alias()
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

// IOrder_by_stmtContext is an interface to support dynamic dispatch.
type IOrder_by_stmtContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	ORDER_() antlr.TerminalNode
	BY_() antlr.TerminalNode
	AllOrdering_term() []IOrdering_termContext
	Ordering_term(i int) IOrdering_termContext
	AllCOMMA() []antlr.TerminalNode
	COMMA(i int) antlr.TerminalNode

	// IsOrder_by_stmtContext differentiates from other interfaces.
	IsOrder_by_stmtContext()
}

type Order_by_stmtContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyOrder_by_stmtContext() *Order_by_stmtContext {
	var p = new(Order_by_stmtContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_order_by_stmt
	return p
}

func InitEmptyOrder_by_stmtContext(p *Order_by_stmtContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_order_by_stmt
}

func (*Order_by_stmtContext) IsOrder_by_stmtContext() {}

func NewOrder_by_stmtContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Order_by_stmtContext {
	var p = new(Order_by_stmtContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SQLParserRULE_order_by_stmt

	return p
}

func (s *Order_by_stmtContext) GetParser() antlr.Parser { return s.parser }

func (s *Order_by_stmtContext) ORDER_() antlr.TerminalNode {
	return s.GetToken(SQLParserORDER_, 0)
}

func (s *Order_by_stmtContext) BY_() antlr.TerminalNode {
	return s.GetToken(SQLParserBY_, 0)
}

func (s *Order_by_stmtContext) AllOrdering_term() []IOrdering_termContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IOrdering_termContext); ok {
			len++
		}
	}

	tst := make([]IOrdering_termContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IOrdering_termContext); ok {
			tst[i] = t.(IOrdering_termContext)
			i++
		}
	}

	return tst
}

func (s *Order_by_stmtContext) Ordering_term(i int) IOrdering_termContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IOrdering_termContext); ok {
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

	return t.(IOrdering_termContext)
}

func (s *Order_by_stmtContext) AllCOMMA() []antlr.TerminalNode {
	return s.GetTokens(SQLParserCOMMA)
}

func (s *Order_by_stmtContext) COMMA(i int) antlr.TerminalNode {
	return s.GetToken(SQLParserCOMMA, i)
}

func (s *Order_by_stmtContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Order_by_stmtContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Order_by_stmtContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SQLParserVisitor:
		return t.VisitOrder_by_stmt(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SQLParser) Order_by_stmt() (localctx IOrder_by_stmtContext) {
	localctx = NewOrder_by_stmtContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 92, SQLParserRULE_order_by_stmt)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(665)
		p.Match(SQLParserORDER_)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(666)
		p.Match(SQLParserBY_)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(667)
		p.Ordering_term()
	}
	p.SetState(672)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == SQLParserCOMMA {
		{
			p.SetState(668)
			p.Match(SQLParserCOMMA)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(669)
			p.Ordering_term()
		}

		p.SetState(674)
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

// ILimit_stmtContext is an interface to support dynamic dispatch.
type ILimit_stmtContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	LIMIT_() antlr.TerminalNode
	AllExpr() []IExprContext
	Expr(i int) IExprContext
	OFFSET_() antlr.TerminalNode

	// IsLimit_stmtContext differentiates from other interfaces.
	IsLimit_stmtContext()
}

type Limit_stmtContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyLimit_stmtContext() *Limit_stmtContext {
	var p = new(Limit_stmtContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_limit_stmt
	return p
}

func InitEmptyLimit_stmtContext(p *Limit_stmtContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_limit_stmt
}

func (*Limit_stmtContext) IsLimit_stmtContext() {}

func NewLimit_stmtContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Limit_stmtContext {
	var p = new(Limit_stmtContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SQLParserRULE_limit_stmt

	return p
}

func (s *Limit_stmtContext) GetParser() antlr.Parser { return s.parser }

func (s *Limit_stmtContext) LIMIT_() antlr.TerminalNode {
	return s.GetToken(SQLParserLIMIT_, 0)
}

func (s *Limit_stmtContext) AllExpr() []IExprContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IExprContext); ok {
			len++
		}
	}

	tst := make([]IExprContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IExprContext); ok {
			tst[i] = t.(IExprContext)
			i++
		}
	}

	return tst
}

func (s *Limit_stmtContext) Expr(i int) IExprContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExprContext); ok {
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

	return t.(IExprContext)
}

func (s *Limit_stmtContext) OFFSET_() antlr.TerminalNode {
	return s.GetToken(SQLParserOFFSET_, 0)
}

func (s *Limit_stmtContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Limit_stmtContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Limit_stmtContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SQLParserVisitor:
		return t.VisitLimit_stmt(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SQLParser) Limit_stmt() (localctx ILimit_stmtContext) {
	localctx = NewLimit_stmtContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 94, SQLParserRULE_limit_stmt)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(675)
		p.Match(SQLParserLIMIT_)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(676)
		p.expr(0)
	}
	p.SetState(679)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == SQLParserOFFSET_ {
		{
			p.SetState(677)
			p.Match(SQLParserOFFSET_)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(678)
			p.expr(0)
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

// IOrdering_termContext is an interface to support dynamic dispatch.
type IOrdering_termContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Expr() IExprContext
	Asc_desc() IAsc_descContext
	NULLS_() antlr.TerminalNode
	FIRST_() antlr.TerminalNode
	LAST_() antlr.TerminalNode

	// IsOrdering_termContext differentiates from other interfaces.
	IsOrdering_termContext()
}

type Ordering_termContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyOrdering_termContext() *Ordering_termContext {
	var p = new(Ordering_termContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_ordering_term
	return p
}

func InitEmptyOrdering_termContext(p *Ordering_termContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_ordering_term
}

func (*Ordering_termContext) IsOrdering_termContext() {}

func NewOrdering_termContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Ordering_termContext {
	var p = new(Ordering_termContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SQLParserRULE_ordering_term

	return p
}

func (s *Ordering_termContext) GetParser() antlr.Parser { return s.parser }

func (s *Ordering_termContext) Expr() IExprContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExprContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExprContext)
}

func (s *Ordering_termContext) Asc_desc() IAsc_descContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IAsc_descContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IAsc_descContext)
}

func (s *Ordering_termContext) NULLS_() antlr.TerminalNode {
	return s.GetToken(SQLParserNULLS_, 0)
}

func (s *Ordering_termContext) FIRST_() antlr.TerminalNode {
	return s.GetToken(SQLParserFIRST_, 0)
}

func (s *Ordering_termContext) LAST_() antlr.TerminalNode {
	return s.GetToken(SQLParserLAST_, 0)
}

func (s *Ordering_termContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Ordering_termContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Ordering_termContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SQLParserVisitor:
		return t.VisitOrdering_term(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SQLParser) Ordering_term() (localctx IOrdering_termContext) {
	localctx = NewOrdering_termContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 96, SQLParserRULE_ordering_term)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(681)
		p.expr(0)
	}
	p.SetState(683)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == SQLParserASC_ || _la == SQLParserDESC_ {
		{
			p.SetState(682)
			p.Asc_desc()
		}

	}
	p.SetState(687)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == SQLParserNULLS_ {
		{
			p.SetState(685)
			p.Match(SQLParserNULLS_)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(686)
			_la = p.GetTokenStream().LA(1)

			if !(_la == SQLParserFIRST_ || _la == SQLParserLAST_) {
				p.GetErrorHandler().RecoverInline(p)
			} else {
				p.GetErrorHandler().ReportMatch(p)
				p.Consume()
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

// IAsc_descContext is an interface to support dynamic dispatch.
type IAsc_descContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	ASC_() antlr.TerminalNode
	DESC_() antlr.TerminalNode

	// IsAsc_descContext differentiates from other interfaces.
	IsAsc_descContext()
}

type Asc_descContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyAsc_descContext() *Asc_descContext {
	var p = new(Asc_descContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_asc_desc
	return p
}

func InitEmptyAsc_descContext(p *Asc_descContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_asc_desc
}

func (*Asc_descContext) IsAsc_descContext() {}

func NewAsc_descContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Asc_descContext {
	var p = new(Asc_descContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SQLParserRULE_asc_desc

	return p
}

func (s *Asc_descContext) GetParser() antlr.Parser { return s.parser }

func (s *Asc_descContext) ASC_() antlr.TerminalNode {
	return s.GetToken(SQLParserASC_, 0)
}

func (s *Asc_descContext) DESC_() antlr.TerminalNode {
	return s.GetToken(SQLParserDESC_, 0)
}

func (s *Asc_descContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Asc_descContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Asc_descContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SQLParserVisitor:
		return t.VisitAsc_desc(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SQLParser) Asc_desc() (localctx IAsc_descContext) {
	localctx = NewAsc_descContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 98, SQLParserRULE_asc_desc)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(689)
		_la = p.GetTokenStream().LA(1)

		if !(_la == SQLParserASC_ || _la == SQLParserDESC_) {
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

// IFunction_keywordContext is an interface to support dynamic dispatch.
type IFunction_keywordContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	LIKE_() antlr.TerminalNode
	REPLACE_() antlr.TerminalNode

	// IsFunction_keywordContext differentiates from other interfaces.
	IsFunction_keywordContext()
}

type Function_keywordContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyFunction_keywordContext() *Function_keywordContext {
	var p = new(Function_keywordContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_function_keyword
	return p
}

func InitEmptyFunction_keywordContext(p *Function_keywordContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_function_keyword
}

func (*Function_keywordContext) IsFunction_keywordContext() {}

func NewFunction_keywordContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Function_keywordContext {
	var p = new(Function_keywordContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SQLParserRULE_function_keyword

	return p
}

func (s *Function_keywordContext) GetParser() antlr.Parser { return s.parser }

func (s *Function_keywordContext) LIKE_() antlr.TerminalNode {
	return s.GetToken(SQLParserLIKE_, 0)
}

func (s *Function_keywordContext) REPLACE_() antlr.TerminalNode {
	return s.GetToken(SQLParserREPLACE_, 0)
}

func (s *Function_keywordContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Function_keywordContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Function_keywordContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SQLParserVisitor:
		return t.VisitFunction_keyword(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SQLParser) Function_keyword() (localctx IFunction_keywordContext) {
	localctx = NewFunction_keywordContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 100, SQLParserRULE_function_keyword)
	p.SetState(694)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case SQLParserOPEN_PAR:
		p.EnterOuterAlt(localctx, 1)

	case SQLParserLIKE_:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(692)
			p.Match(SQLParserLIKE_)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case SQLParserREPLACE_:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(693)
			p.Match(SQLParserREPLACE_)
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

// IFunction_nameContext is an interface to support dynamic dispatch.
type IFunction_nameContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	IDENTIFIER() antlr.TerminalNode
	Function_keyword() IFunction_keywordContext

	// IsFunction_nameContext differentiates from other interfaces.
	IsFunction_nameContext()
}

type Function_nameContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyFunction_nameContext() *Function_nameContext {
	var p = new(Function_nameContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_function_name
	return p
}

func InitEmptyFunction_nameContext(p *Function_nameContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_function_name
}

func (*Function_nameContext) IsFunction_nameContext() {}

func NewFunction_nameContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Function_nameContext {
	var p = new(Function_nameContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SQLParserRULE_function_name

	return p
}

func (s *Function_nameContext) GetParser() antlr.Parser { return s.parser }

func (s *Function_nameContext) IDENTIFIER() antlr.TerminalNode {
	return s.GetToken(SQLParserIDENTIFIER, 0)
}

func (s *Function_nameContext) Function_keyword() IFunction_keywordContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IFunction_keywordContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IFunction_keywordContext)
}

func (s *Function_nameContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Function_nameContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Function_nameContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SQLParserVisitor:
		return t.VisitFunction_name(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SQLParser) Function_name() (localctx IFunction_nameContext) {
	localctx = NewFunction_nameContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 102, SQLParserRULE_function_name)
	p.SetState(698)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case SQLParserIDENTIFIER:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(696)
			p.Match(SQLParserIDENTIFIER)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case SQLParserOPEN_PAR, SQLParserLIKE_, SQLParserREPLACE_:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(697)
			p.Function_keyword()
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

// ITable_nameContext is an interface to support dynamic dispatch.
type ITable_nameContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	IDENTIFIER() antlr.TerminalNode

	// IsTable_nameContext differentiates from other interfaces.
	IsTable_nameContext()
}

type Table_nameContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyTable_nameContext() *Table_nameContext {
	var p = new(Table_nameContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_table_name
	return p
}

func InitEmptyTable_nameContext(p *Table_nameContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_table_name
}

func (*Table_nameContext) IsTable_nameContext() {}

func NewTable_nameContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Table_nameContext {
	var p = new(Table_nameContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SQLParserRULE_table_name

	return p
}

func (s *Table_nameContext) GetParser() antlr.Parser { return s.parser }

func (s *Table_nameContext) IDENTIFIER() antlr.TerminalNode {
	return s.GetToken(SQLParserIDENTIFIER, 0)
}

func (s *Table_nameContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Table_nameContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Table_nameContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SQLParserVisitor:
		return t.VisitTable_name(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SQLParser) Table_name() (localctx ITable_nameContext) {
	localctx = NewTable_nameContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 104, SQLParserRULE_table_name)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(700)
		p.Match(SQLParserIDENTIFIER)
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

// ITable_aliasContext is an interface to support dynamic dispatch.
type ITable_aliasContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	IDENTIFIER() antlr.TerminalNode

	// IsTable_aliasContext differentiates from other interfaces.
	IsTable_aliasContext()
}

type Table_aliasContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyTable_aliasContext() *Table_aliasContext {
	var p = new(Table_aliasContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_table_alias
	return p
}

func InitEmptyTable_aliasContext(p *Table_aliasContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_table_alias
}

func (*Table_aliasContext) IsTable_aliasContext() {}

func NewTable_aliasContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Table_aliasContext {
	var p = new(Table_aliasContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SQLParserRULE_table_alias

	return p
}

func (s *Table_aliasContext) GetParser() antlr.Parser { return s.parser }

func (s *Table_aliasContext) IDENTIFIER() antlr.TerminalNode {
	return s.GetToken(SQLParserIDENTIFIER, 0)
}

func (s *Table_aliasContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Table_aliasContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Table_aliasContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SQLParserVisitor:
		return t.VisitTable_alias(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SQLParser) Table_alias() (localctx ITable_aliasContext) {
	localctx = NewTable_aliasContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 106, SQLParserRULE_table_alias)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(702)
		p.Match(SQLParserIDENTIFIER)
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

// IColumn_nameContext is an interface to support dynamic dispatch.
type IColumn_nameContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	IDENTIFIER() antlr.TerminalNode

	// IsColumn_nameContext differentiates from other interfaces.
	IsColumn_nameContext()
}

type Column_nameContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyColumn_nameContext() *Column_nameContext {
	var p = new(Column_nameContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_column_name
	return p
}

func InitEmptyColumn_nameContext(p *Column_nameContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_column_name
}

func (*Column_nameContext) IsColumn_nameContext() {}

func NewColumn_nameContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Column_nameContext {
	var p = new(Column_nameContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SQLParserRULE_column_name

	return p
}

func (s *Column_nameContext) GetParser() antlr.Parser { return s.parser }

func (s *Column_nameContext) IDENTIFIER() antlr.TerminalNode {
	return s.GetToken(SQLParserIDENTIFIER, 0)
}

func (s *Column_nameContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Column_nameContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Column_nameContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SQLParserVisitor:
		return t.VisitColumn_name(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SQLParser) Column_name() (localctx IColumn_nameContext) {
	localctx = NewColumn_nameContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 108, SQLParserRULE_column_name)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(704)
		p.Match(SQLParserIDENTIFIER)
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

// IColumn_aliasContext is an interface to support dynamic dispatch.
type IColumn_aliasContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	IDENTIFIER() antlr.TerminalNode

	// IsColumn_aliasContext differentiates from other interfaces.
	IsColumn_aliasContext()
}

type Column_aliasContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyColumn_aliasContext() *Column_aliasContext {
	var p = new(Column_aliasContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_column_alias
	return p
}

func InitEmptyColumn_aliasContext(p *Column_aliasContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_column_alias
}

func (*Column_aliasContext) IsColumn_aliasContext() {}

func NewColumn_aliasContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Column_aliasContext {
	var p = new(Column_aliasContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SQLParserRULE_column_alias

	return p
}

func (s *Column_aliasContext) GetParser() antlr.Parser { return s.parser }

func (s *Column_aliasContext) IDENTIFIER() antlr.TerminalNode {
	return s.GetToken(SQLParserIDENTIFIER, 0)
}

func (s *Column_aliasContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Column_aliasContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Column_aliasContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SQLParserVisitor:
		return t.VisitColumn_alias(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SQLParser) Column_alias() (localctx IColumn_aliasContext) {
	localctx = NewColumn_aliasContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 110, SQLParserRULE_column_alias)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(706)
		p.Match(SQLParserIDENTIFIER)
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

// ICollation_nameContext is an interface to support dynamic dispatch.
type ICollation_nameContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	IDENTIFIER() antlr.TerminalNode

	// IsCollation_nameContext differentiates from other interfaces.
	IsCollation_nameContext()
}

type Collation_nameContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyCollation_nameContext() *Collation_nameContext {
	var p = new(Collation_nameContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_collation_name
	return p
}

func InitEmptyCollation_nameContext(p *Collation_nameContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_collation_name
}

func (*Collation_nameContext) IsCollation_nameContext() {}

func NewCollation_nameContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Collation_nameContext {
	var p = new(Collation_nameContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SQLParserRULE_collation_name

	return p
}

func (s *Collation_nameContext) GetParser() antlr.Parser { return s.parser }

func (s *Collation_nameContext) IDENTIFIER() antlr.TerminalNode {
	return s.GetToken(SQLParserIDENTIFIER, 0)
}

func (s *Collation_nameContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Collation_nameContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Collation_nameContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SQLParserVisitor:
		return t.VisitCollation_name(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SQLParser) Collation_name() (localctx ICollation_nameContext) {
	localctx = NewCollation_nameContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 112, SQLParserRULE_collation_name)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(708)
		p.Match(SQLParserIDENTIFIER)
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

// IIndex_nameContext is an interface to support dynamic dispatch.
type IIndex_nameContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	IDENTIFIER() antlr.TerminalNode

	// IsIndex_nameContext differentiates from other interfaces.
	IsIndex_nameContext()
}

type Index_nameContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyIndex_nameContext() *Index_nameContext {
	var p = new(Index_nameContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_index_name
	return p
}

func InitEmptyIndex_nameContext(p *Index_nameContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SQLParserRULE_index_name
}

func (*Index_nameContext) IsIndex_nameContext() {}

func NewIndex_nameContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Index_nameContext {
	var p = new(Index_nameContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SQLParserRULE_index_name

	return p
}

func (s *Index_nameContext) GetParser() antlr.Parser { return s.parser }

func (s *Index_nameContext) IDENTIFIER() antlr.TerminalNode {
	return s.GetToken(SQLParserIDENTIFIER, 0)
}

func (s *Index_nameContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Index_nameContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Index_nameContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case SQLParserVisitor:
		return t.VisitIndex_name(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *SQLParser) Index_name() (localctx IIndex_nameContext) {
	localctx = NewIndex_nameContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 114, SQLParserRULE_index_name)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(710)
		p.Match(SQLParserIDENTIFIER)
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

func (p *SQLParser) Sempred(localctx antlr.RuleContext, ruleIndex, predIndex int) bool {
	switch ruleIndex {
	case 13:
		var t *ExprContext = nil
		if localctx != nil {
			t = localctx.(*ExprContext)
		}
		return p.Expr_Sempred(t, predIndex)

	default:
		panic("No predicate with index: " + fmt.Sprint(ruleIndex))
	}
}

func (p *SQLParser) Expr_Sempred(localctx antlr.RuleContext, predIndex int) bool {
	switch predIndex {
	case 0:
		return p.Precpred(p.GetParserRuleContext(), 12)

	case 1:
		return p.Precpred(p.GetParserRuleContext(), 11)

	case 2:
		return p.Precpred(p.GetParserRuleContext(), 8)

	case 3:
		return p.Precpred(p.GetParserRuleContext(), 6)

	case 4:
		return p.Precpred(p.GetParserRuleContext(), 2)

	case 5:
		return p.Precpred(p.GetParserRuleContext(), 1)

	case 6:
		return p.Precpred(p.GetParserRuleContext(), 18)

	case 7:
		return p.Precpred(p.GetParserRuleContext(), 10)

	case 8:
		return p.Precpred(p.GetParserRuleContext(), 9)

	case 9:
		return p.Precpred(p.GetParserRuleContext(), 7)

	case 10:
		return p.Precpred(p.GetParserRuleContext(), 5)

	case 11:
		return p.Precpred(p.GetParserRuleContext(), 4)

	default:
		panic("No predicate with index: " + fmt.Sprint(predIndex))
	}
}
