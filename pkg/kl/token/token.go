package token

type TokenType int

const (
	// Special tokens
	ILLEGAL TokenType = iota
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
	//INSERT   // insert
	//INTO     // into
	//VALUES   // values
	//WHERE    // where
	//AND      // and
	//OR       // or
	//SELECT   // select
	//FROM     // from
	//UPDATE   // update
	//DROP     // drop
	//
	//INT  // int
	//UUID // uuid
	//TEXT // string
	//
	//PRIMARY // primary
	//UNIQUE  // unique
	//MIN     // min
	//MAX     // max
	//MINLEN  // minlen
	//MAXLEN  // maxlen
	NULL
	NOTNULL // notnull
	//
	//CONST      // const
	//ACTION     // action
	keywordEnd // keywordEnd

	symbolBeg
	ASSIGN // =
	EQL    // ==
	LSS    // <
	GTR    // >
	//NOT    // !
	//NEQ // !=
	//LEQ // <=
	//GEQ // >=

	LPAREN // (
	LBRACK // [
	LBRACE // {
	RPAREN // )
	RBRACK // ]
	RBRACE // }

	COMMA // ,
	//PERIOD    // .
	SEMICOLON // ;
	//COLON     // :
	symbolEnd
)

var tokens = [...]string{
	ILLEGAL: "ILLEGAL",
	EOF:     "EOF",
	COMMENT: "COMMENT",
	//
	IDENT: "IDENT",
	//
	DATABASE: "database",
	TABLE:    "table",
	//INT:      "int",
	//UUID:     "uuid",
	//TEXT:     "text",
	NULL:    "null",
	NOTNULL: "notnull",
	//
	//EQL:       "==",
	//LSS:       "<",
	//GTR:       ">",
	//ASSIGN:    "=",
	//NOT:       "!",
	//NEQ:       "!=",
	//LEQ:       "<=",
	//GEQ:       ">=",
	LPAREN: "(",
	LBRACK: "[",
	LBRACE: "{",
	RPAREN: ")",
	RBRACK: "]",
	RBRACE: "}",
	COMMA:  ",",
	//PERIOD:    ".",
	SEMICOLON: ";",
	//COLON:     ":",
}

func (t TokenType) ToInt() int {
	return int(t)
}

func (t TokenType) IsLiteral() bool {
	return literalBeg < t && t < literalEnd
}

func (t TokenType) String() string {
	return tokens[t]
}

//func (t TokenType) IsColumnType() bool {
//	return t == INT || t == TEXT || t == UUID
//}

func (t TokenType) IsAttrType() bool {
	return t == NULL || t == NOTNULL
}

var keywords map[string]TokenType
var symbols map[string]TokenType

func init() {
	keywords = make(map[string]TokenType, keywordEnd-(keywordBeg+1))
	for i := keywordBeg + 1; i < keywordEnd; i++ {
		keywords[tokens[i]] = i
	}

	symbols = make(map[string]TokenType, symbolEnd-(symbolBeg+1))
	for i := symbolBeg + 1; i < symbolEnd; i++ {
		symbols[tokens[i]] = i
	}
}

func Lookup(ident string) TokenType {
	if len(ident) == 1 {
		return IDENT
	}

	if tok, isKeyword := keywords[ident]; isKeyword {
		return tok
	}
	return IDENT
}

type Token struct {
	Type    TokenType
	Literal string
	//pos
}
