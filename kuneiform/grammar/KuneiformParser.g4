/*
 * A ANTLR4 grammar for Kuneiform.
 * Developed by the Kuneiform team.
*/

parser grammar KuneiformParser;

options {
    tokenVocab=KuneiformLexer;
}

// program is the parser entrypoint
program:
    database_declaration
    (use_declaration | table_declaration
    | stmt_mode
    )* EOF
;

stmt_mode:
    ANNOTATION*
    (action_declaration|procedure_declaration)
;

database_declaration:
    DATABASE IDENTIFIER SCOL
;

use_declaration:
    USE extension_name=IDENTIFIER
    (LBRACE IDENTIFIER COL literal (COMMA IDENTIFIER COL literal)? RBRACE)?
    AS alias=IDENTIFIER SCOL
;

table_declaration:
     TABLE IDENTIFIER LBRACE
     column_def (COMMA (column_def | index_def | foreign_key_def))*
     RBRACE
 ;

column_def:
    name=IDENTIFIER type=type_selector constraint*
;

index_def:
    INDEX_NAME
    (UNIQUE | INDEX | PRIMARY)
    LPAREN  columns=identifier_list RPAREN
;

foreign_key_def:
    FOREIGN_KEY
    LPAREN child_keys=identifier_list RPAREN
    REFERENCES parent_table=IDENTIFIER LPAREN parent_keys=identifier_list RPAREN
    foreign_key_action*
;

foreign_key_action:
    (ON_UPDATE|ON_DELETE) DO? (DO_NO_ACTION|DO_CASCADE|DO_SET_NULL|DO_SET_DEFAULT|DO_RESTRICT)
;

identifier_list:
    IDENTIFIER (COMMA IDENTIFIER)*
;

literal:
    NUMERIC_LITERAL
    | BLOB_LITERAL
    | TEXT_LITERAL
    | BOOLEAN_LITERAL
;

type_selector:
    type=IDENTIFIER
    (LBRACKET RBRACKET)?
;

constraint:
    MIN LPAREN NUMERIC_LITERAL RPAREN # MIN
    | MAX LPAREN NUMERIC_LITERAL RPAREN # MAX
    | MIN_LEN LPAREN NUMERIC_LITERAL RPAREN # MIN_LEN
    | MAX_LEN LPAREN NUMERIC_LITERAL RPAREN # MAX_LEN
    | NOT_NULL # NOT_NULL
    | PRIMARY # PRIMARY_KEY
    | DEFAULT LPAREN literal RPAREN # DEFAULT
    | UNIQUE # UNIQUE
;

// STMT_MODE parsing:

action_declaration:
    START_ACTION STMT_IDENTIFIER
    STMT_LPAREN (STMT_VAR (STMT_COMMA STMT_VAR)*)? STMT_RPAREN
    STMT_ACCESS+
    STMT_BODY
;

procedure_declaration:
    START_PROCEDURE procedure_name=STMT_IDENTIFIER
    STMT_LPAREN stmt_typed_param_list? STMT_RPAREN
    STMT_ACCESS+
    (STMT_RETURNS stmt_return)?
    STMT_BODY
;

stmt_return:
    STMT_TABLE? STMT_LPAREN STMT_IDENTIFIER stmt_type_selector (STMT_COMMA STMT_IDENTIFIER stmt_type_selector)* STMT_RPAREN
;

stmt_typed_param_list:
    STMT_VAR stmt_type_selector (STMT_COMMA STMT_VAR stmt_type_selector)*
;

stmt_type_selector:
    type=STMT_IDENTIFIER
    (STMT_ARRAY)?
;