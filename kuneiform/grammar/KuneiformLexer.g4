/*
 * A ANTLR4 grammar for Kuneiform.
 * Developed by the Kuneiform team.
*/

lexer grammar KuneiformLexer;

options { caseInsensitive = true; }

// symbols
LBRACE: '{';
RBRACE: '}';
LBRACKET: '[';
RBRACKET: ']';
COL:       ':';
SCOL:      ';';
LPAREN:   '(';
RPAREN:   ')';
COMMA:     ',';
AT:        '@';
PERIOD:    '.';
EQUALS: '=';


// keywords
DATABASE: 'database';
USE:      'use';
IMPORT:   'import';
AS:       'as';
//// column attrs
MIN:      'min';
MAX:      'max';
MIN_LEN:  'minlen';
MAX_LEN:  'maxlen';
NOT_NULL: 'not' WSNL? 'null';
PRIMARY:  'primary' ('_'|WSNL)? 'key'?;
DEFAULT:  'default';
UNIQUE:   'unique';
INDEX:    'index';
TABLE: 'table';
TYPE: 'type';
//// foreign key
FOREIGN_KEY:           'foreign' ('_'|WSNL) 'key'|'fk';
REFERENCES:            'references'|'ref';
ON_UPDATE:      'on' ('_'|WSNL) 'update';
ON_DELETE:      'on' ('_'|WSNL) 'delete';
DO_NO_ACTION:   'no' ('_'|WSNL) 'action';
DO_CASCADE:     'cascade';
DO_SET_NULL:    'set' ('_'|WSNL) 'null';
DO_SET_DEFAULT: 'set' ('_'|WSNL) 'default';
DO_RESTRICT:    'restrict';

START_ACTION: 'action' -> pushMode(STMT_MODE);
START_PROCEDURE: 'procedure' -> pushMode(STMT_MODE);

// literals
NUMERIC_LITERAL:
    ('+'|'-')?[0-9]+
;

TEXT_LITERAL:
    '\'' ( ~['\r\n\\] | ('\\' .) )* '\''
;

BOOLEAN_LITERAL:
    'true'
    | 'false'
;

BLOB_LITERAL:
    '0x' [0-9a-f]+
;

VAR:
    '$' IDENTIFIER
;

INDEX_NAME:
    '#' IDENTIFIER
;

IDENTIFIER:
    [a-z] [a-z_0-9]*
;

ANNOTATION:
    '@' ~[\n]+
;

WS:            [ \t\r\n]        -> channel(HIDDEN);
TERMINATOR:    [\r\n]+       -> channel(HIDDEN);
BLOCK_COMMENT: '/*' .*? '*/' -> channel(HIDDEN);
LINE_COMMENT:  '//' ~[\r\n]* -> channel(HIDDEN);

// fragments
fragment WSNL: [ \t\r\n]+; // whitespace with new line
fragment DIGIT: [0-9];

// STMT_MODE captures actions and procedures.
// it is used to allow us to correctly parse action / procedure bodies
mode STMT_MODE;
    // STMT_BODY captures everything between two curly braces.
    // It is defined recusively to only exit on the final curly brace.
    STMT_BODY: LBRACE ( ANY | STMT_BODY | TEXT )* RBRACE -> popMode;

    // we don't make TEXT fragmented because we want anything
    // textual to be ignore and not tokenized.
    TEXT: '\'' ( '\\' '\'' | ~('\'' | '\\') )* '\'';

    STMT_LPAREN: LPAREN;
    STMT_RPAREN: RPAREN;
    STMT_COMMA: COMMA;
    STMT_PERIOD: PERIOD;
    STMT_RETURNS: 'returns';
    STMT_TABLE: TABLE;
    STMT_ARRAY: LBRACKET RBRACKET;
    STMT_VAR: '$' IDENTIFIER;
    STMT_ACCESS: 'public'|'private'|'view'|'owner';

    // keep IDENTIFIER last
    STMT_IDENTIFIER: IDENTIFIER;

    // RECUR tokenizes braces recusively, allowing us to
    // only exit the mode once enough  right braces have been
    // used

    STMT_WS: WS -> channel(HIDDEN);
    STMT_TERMINATOR: TERMINATOR -> channel(HIDDEN);
    STMT_BLOCK_COMMENT: BLOCK_COMMENT -> channel(HIDDEN);
    STMT_LINE_COMMENT: LINE_COMMENT -> channel(HIDDEN);

    // text to not tokenize braces in text literals
    fragment ANY: ~[{}]+;