lexer grammar ProcedureLexer;

options { caseInsensitive = true; }

// symbols
SEMICOLON: ';';
LPAREN:   '(';
RPAREN:   ')';
LBRACE:   '{';
RBRACE:   '}';
COMMA:     ',';
COLON:     ':';
DOLLAR:    '$';
AT:        '@';
ASSIGN:    ':=';
PERIOD:    '.';
LBRACKET : '[';
RBRACKET : ']';
SINGLE_QUOTE: '\'';

// arithmetic operators
PLUS:      '+';
MINUS:     '-';
MUL:       '*';
DIV:       '/';
MOD:       '%';

// comparison operators
LT:        '<';
LT_EQ:     '<=';
GT:        '>';
GT_EQ:     '>=';
NEQ:       '!=';
EQ:        '==';

// we only need sql statement as a whole, sql-parser will parse it
ANY_SQL: (SELECT_ | INSERT_ | UPDATE_ | DELETE_ | WITH_) WSNL ~[;{]+;

// Keywords
FOR: 'for';
IN: 'in';
IF: 'if';
ELSEIF: 'elseif';
ELSE: 'else';
TO: 'to';
RETURN: 'return';
BREAK: 'break';
NEXT: 'next';

// literals

BOOLEAN_LITERAL:
    'true' | 'false'
;

INT_LITERAL:
    [0-9]+
;

BLOB_LITERAL:
    '0x' [0-9a-f]+
;

TEXT_LITERAL:
    SINGLE_QUOTE (~['\r\n\\] | ('\\' .))* SINGLE_QUOTE
;

NULL_LITERAL: 'null';

// vars
IDENTIFIER:
    [a-z] [a-z_0-9]*
;

VARIABLE: (DOLLAR|AT) IDENTIFIER;

WS:            WSNL        -> channel(HIDDEN);
TERMINATOR:    [\r\n]+       -> channel(HIDDEN);
BLOCK_COMMENT: '/*' .*? '*/' -> channel(HIDDEN);
LINE_COMMENT:  '//' ~[\r\n]* -> channel(HIDDEN);

// fragments
fragment WSNL: [ \t\r\n]+; // whitespace with new line
fragment SELECT_:   'select';
fragment INSERT_:   'insert';
fragment UPDATE_:   'update';
fragment DELETE_:   'delete';
fragment WITH_:     'with';