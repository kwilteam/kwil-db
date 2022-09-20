package dbml

import "strings"

type Token int

//go:generate stringer -type=Token
const (
	// Special tokens
	ILLEGAL Token = iota
	EOF
	COMMENT
	NEWLINE
	WHITESPACE

	_literalBeg
	// Identifiers and basic type literals
	// (these tokens stand for classes of literals)
	IDENT   // main
	INT     // 12345
	FLOAT   // 123.45
	IMAG    // 123.45i
	STRING  // 'abc'
	DSTRING // "abc"
	TSTRING // '''abc'''

	EXPR // `now()`
	BLOCK

	_literalEnd

	_operatorBeg

	SUB  // -
	LT   // <
	GT   // >
	LTGT // <>

	LPAREN // (
	LBRACK // [
	LBRACE // {
	COMMA  // ,
	PERIOD // .

	RPAREN    // )
	RBRACK    // ]
	RBRACE    // }
	SEMICOLON // ;
	COLON     // :

	_operatorEnd

	_keywordBeg

	PROJECT
	TABLE
	QUERY
	ROLE
	ENUM
	REF
	AS
	TABLEGROUP

	_keywordEnd

	_miscBeg

	PRIMARY
	KEY
	PK
	NOTE
	UNIQUE
	NOT
	NULL
	INCREMENT
	DEFAULT

	INDEXES
	TYPE
	DELETE
	UPDATE
	NO
	ACTION
	RESTRICT
	SET

	_miscEnd
)

var Tokens = [...]string{
	ILLEGAL: "ILLEGAL",

	EOF:     "EOF",
	COMMENT: "COMMENT",

	IDENT:   "IDENT",
	INT:     "INT",
	FLOAT:   "FLOAT",
	IMAG:    "IMAG",
	STRING:  "STRING",
	DSTRING: "DSTRING",
	TSTRING: "TSTRING",
	EXPR:    "EXPR",

	SUB: "-",
	LT:  "<",
	GT:  ">",

	LPAREN: "(",
	LBRACK: "[",
	LBRACE: "{",

	RPAREN: ")",
	RBRACK: "]",
	RBRACE: "}",

	SEMICOLON: ";",
	COLON:     ":",
	COMMA:     ",",
	PERIOD:    ".",

	PROJECT:    "PROJECT",
	TABLE:      "TABLE",
	ENUM:       "ENUM",
	REF:        "REF",
	QUERY:      "QUERY",
	ROLE:       "ROLE",
	AS:         "AS",
	TABLEGROUP: "TABLEGROUP",

	PRIMARY:   "PRIMARY",
	KEY:       "KEY",
	PK:        "PK",
	NOTE:      "NOTE",
	UNIQUE:    "UNIQUE",
	NOT:       "NOT",
	NULL:      "NULL",
	INCREMENT: "INCREMENT",
	DEFAULT:   "DEFAULT",

	INDEXES:  "INDEXES",
	TYPE:     "TYPE",
	DELETE:   "DELETE",
	UPDATE:   "UPDATE",
	NO:       "NO",
	ACTION:   "ACTION",
	RESTRICT: "RESTRICT",
	SET:      "SET",
}

var keywords map[string]Token

func init() {
	keywords = make(map[string]Token)
	for i := _keywordBeg + 1; i < _miscEnd; i++ {
		keywords[Tokens[i]] = i
	}
}

func Lookup(ident string) Token {
	if tok, ok := keywords[strings.ToUpper(ident)]; ok {
		return tok
	}
	return IDENT
}

func IsIdent(t Token) bool {
	switch t {
	case IDENT, DSTRING, STRING:
		return true
	default:
		return _keywordBeg < t && t < _miscEnd
	}
}
