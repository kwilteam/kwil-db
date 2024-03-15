/*
 * A ANTLR4 grammar for Action. ONLY temporary.
 * Developed by the Kuneiform team.
*/

parser grammar ActionParser;

options {
    tokenVocab=ActionLexer;
}

statement:
    stmt+
;

literal_value:
    STRING_LITERAL
    | UNSIGNED_NUMBER_LITERAL
;

// NOTE: this is temporary, this will be same as `ext_call_name`
// but we don't support call external action yet.
action_name:
    IDENTIFIER
;

stmt:
    sql_stmt
    | call_stmt
;

sql_stmt:
    SQL_STMT SCOL
;

call_stmt:
    (call_receivers ASSIGN)?
    call_body SCOL
;

call_receivers:
    variable (COMMA variable)*
;

// use expr in the future, limit syntax for now
call_body:
    fn_name
    L_PAREN fn_arg_list R_PAREN
;

variable:
    VARIABLE
;

block_var:
    BLOCK_VARIABLE
;

extension_call_name:
    IDENTIFIER PERIOD IDENTIFIER
;

//external_action_name:
//    IDENTIFIER PERIOD IDENTIFIER
//;

// function name
fn_name:
    extension_call_name
    | action_name
//    | external_action_name
;

// scalar function, it is meant to be same as SQL scalar function name
sfn_name:
    IDENTIFIER
;

//fn_arg:
//    literal_value
//    | variable_name
//    | block_variable_name
//;

fn_arg_list:
//    fn_arg? (COMMA fn_arg)*
    fn_arg_expr? (COMMA fn_arg_expr)*
;

// NOTE: this will only be used inside fn_arg_list, his is based on sqlparser's expr.
// This is only meant to support the most basic expressions.
//
// binary operators precedence: highest to lowest:
//   * / %
//   + -
//   < <= > >=
//   == != <>
//   =
//   AND
//   OR
fn_arg_expr:
    // primary expressions(don't fit in operator pattern), order is irrelevant
    literal_value
    | variable
    | block_var
    // scalar functions
    | sfn_name L_PAREN ( (fn_arg_expr (COMMA fn_arg_expr)*) | STAR )? R_PAREN
    // order is relevant for below expressions
    | L_PAREN elevate_expr=fn_arg_expr R_PAREN
    | ( MINUS | PLUS ) unary_expr=fn_arg_expr
    // binary operators
    | fn_arg_expr ( STAR | DIV | MOD ) fn_arg_expr
    | fn_arg_expr ( PLUS | MINUS) fn_arg_expr
    | fn_arg_expr ( LT | LT_EQ | GT | GT_EQ ) fn_arg_expr
    | fn_arg_expr ( ASSIGN | SQL_NOT_EQ1 | SQL_NOT_EQ2 ) fn_arg_expr
    // logical operators
    | NOT_ unary_expr=fn_arg_expr
    | fn_arg_expr AND_ fn_arg_expr
    | fn_arg_expr OR_ fn_arg_expr
;

// future expr, replace whole `call_body`?