parser grammar ProcedureParser;

options {
    tokenVocab=ProcedureLexer;
}

// top-level rule
program:
    statement* EOF
;

statement:
    VARIABLE type SEMICOLON # stmt_variable_declaration
    // stmt_procedure_call must go above stmt_variable_assignment
    | (variable_or_underscore (COMMA variable_or_underscore)* ASSIGN)? call_expression SEMICOLON # stmt_procedure_call
    | VARIABLE ASSIGN expression SEMICOLON # stmt_variable_assignment
    | VARIABLE type ASSIGN expression SEMICOLON # stmt_variable_assignment_with_declaration
    | FOR VARIABLE IN (range|call_expression|VARIABLE|ANY_SQL) LBRACE statement* RBRACE # stmt_for_loop
    | IF if_then_block (ELSEIF if_then_block)* (ELSE LBRACE statement* RBRACE)? # stmt_if
    | ANY_SQL SEMICOLON # stmt_sql
    | BREAK SEMICOLON # stmt_break
    | RETURN (expression_list|ANY_SQL) SEMICOLON # stmt_return
    | RETURN NEXT expression_list SEMICOLON # stmt_return_next
;

variable_or_underscore:
   (VARIABLE|UNDERSCORE)
;

type:
    IDENTIFIER (LPAREN literal_list RPAREN)? (LBRACKET RBRACKET)? // Handles arrays of any type, including nested arrays
;

literal:
    TEXT_LITERAL
    | BOOLEAN_LITERAL
    | INT_LITERAL
    | NULL_LITERAL
    | BLOB_LITERAL
    | FIXED_LITERAL
;

literal_list:
    literal (COMMA literal)*
;

// expressions
expression:
    literal type_cast? # expr_literal
    | expression_make_array type_cast? # expr_make_array
    | call_expression type_cast? # expr_call
    | VARIABLE type_cast? # expr_variable
    | expression LBRACKET expression RBRACKET type_cast? # expr_array_access
    | expression PERIOD IDENTIFIER type_cast? # expr_field_access
    | LPAREN expression RPAREN type_cast? # expr_parenthesized
    // the below do not have type casts, but can be wrapped in parens
    // which can be type casted
    | left=expression operator=(LT|LT_EQ|GT|GT_EQ|NEQ|EQ) right=expression # expr_comparison
    // logical operators, separated for precedence
    | expression (MUL|DIV|MOD) expression # expr_arithmetic
    | expression (PLUS|MINUS) expression # expr_arithmetic
;

type_cast:
    TYPE_CAST type
;

expression_list:
    expression (COMMA expression)*
;

expression_make_array:
    LBRACKET (expression_list)? RBRACKET
;

call_expression:
    IDENTIFIER LPAREN (expression_list)? RPAREN #normal_call
    | IDENTIFIER LBRACKET dbid=expression COMMA procedure=expression RBRACKET LPAREN (expression_list)? RPAREN #foreign_call
;

// range used for for loops
range:
    expression COLON expression
;

if_then_block:
    expression LBRACE statement* RBRACE
;