package token

import "strings"

type Token int

const (
	// Special tokens
	ILLEGAL Token = iota
	EOF
	COMMENT

	literalBeg
	// Identifiers and basic type literals
	// (these tokens stand for classes of literals)
	IDENT   // variables names
	INTEGER // 12345
	//FLOAT  // 123.45
	STRING // "abc"
	literalEnd

	keywordBeg
	DATABASE // database
	TABLE    // table
	ACTION
	PUBLIC // public
	PRIVATE

	WITH
	REPLACE
	INSERT // insert
	//INTO   // into
	//VALUES // values
	//WHERE    // where
	//AND      // and
	//OR       // or
	SELECT // select
	UPDATE // update
	DELETE
	DROP    // drop
	UNIQUE  // unique
	INDEX   // index
	PRIMARY // primary

	//CONST      // const
	//ACTION     // action

	attrBeg
	MIN    // min
	MAX    // max
	MINLEN // minlen
	MAXLEN // maxlen
	NULL
	NOTNULL // notnull
	attrEnd
	keywordEnd // keywordEnd

	symbolBeg
	ADD
	SUB
	DIV
	MUL
	MOD

	ASSIGN // =
	//EQL    // ==
	LSS // <
	GTR // >
	NOT // !
	NEQ // !=
	LEQ // <=
	GEQ // >=

	LPAREN // (
	LBRACK // [
	LBRACE // {
	RPAREN // )
	RBRACK // ]
	RBRACE // }

	COMMA     // ,
	PERIOD    // .
	SEMICOLON // ;
	//COLON     // :
	symbolEnd
)

var tokens = [...]string{
	ILLEGAL: "ILLEGAL",
	EOF:     "EOF",
	COMMENT: "COMMENT",
	//
	IDENT:   "IDENT",
	INTEGER: "INTEGER",
	STRING:  "STRING",
	//
	DATABASE: "database",
	TABLE:    "table",
	ACTION:   "action",
	PUBLIC:   "public",
	PRIVATE:  "private",
	WITH:     "with",
	REPLACE:  "replace",
	INSERT:   "insert",
	//INTO:     "into",
	//VALUES:   "values",
	SELECT: "select",
	UPDATE: "update",
	DELETE: "delete",
	DROP:   "drop",

	UNIQUE:  "unique",
	INDEX:   "index",
	PRIMARY: "primary",

	MIN:     "min",
	MAX:     "max",
	MINLEN:  "minlen",
	MAXLEN:  "maxlen",
	NULL:    "null",
	NOTNULL: "notnull",
	//

	ADD: "+",
	SUB: "-",
	MUL: "*",
	DIV: "/",
	MOD: "%",

	LSS:       "<",
	GTR:       ">",
	ASSIGN:    "=",
	NOT:       "!",
	NEQ:       "!=",
	LEQ:       "<=",
	GEQ:       ">=",
	LPAREN:    "(",
	LBRACK:    "[",
	LBRACE:    "{",
	RPAREN:    ")",
	RBRACK:    "]",
	RBRACE:    "}",
	COMMA:     ",",
	PERIOD:    ".",
	SEMICOLON: ";",
	//COLON:     ":",
}

func (t Token) ToInt() int {
	return int(t)
}

func (t Token) String() string {
	return tokens[t]
}
func (t Token) IsAttr() bool {
	return attrBeg < t && t < attrEnd
}

func (t Token) IsLiteral() bool {
	return literalBeg < t && t < literalEnd
}

//func (t TokenType) IsColumnType() bool {
//	return t == INT || t == TEXT || t == UUID
//}

var keywords map[string]Token
var symbols map[string]Token

func init() {
	keywords = make(map[string]Token, keywordEnd-(keywordBeg+1))
	for i := keywordBeg + 1; i < keywordEnd; i++ {
		keywords[tokens[i]] = i
	}

	symbols = make(map[string]Token, symbolEnd-(symbolBeg+1))
	for i := symbolBeg + 1; i < symbolEnd; i++ {
		symbols[tokens[i]] = i
	}
}

func Lookup(ident string) Token {
	ident = strings.ToLower(ident)

	if len(ident) == 1 {
		return IDENT
	}

	if tok, isKeyword := keywords[ident]; isKeyword {
		return tok
	}
	return IDENT
}
