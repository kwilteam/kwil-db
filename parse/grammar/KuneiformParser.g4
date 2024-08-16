/*
 * A ANTLR4 grammar for Kuneiform.
 * Developed by the Kwil team.
*/
parser grammar KuneiformParser;

options {
    tokenVocab = KuneiformLexer;
}

// there are 4 top-level entry points for the parser:
// 1. schema_entry
// 2. sql_entry
// 3. action_entry
// 4. procedure_entry
// It is necessary to keep each type of entry separate, since some statements
// can be ambiguous between the different types of entries. Callers will know
// which entry to use based on when they are parsing.

schema_entry:
    schema EOF
;

sql_entry:
    sql EOF
;

action_entry:
    action_block EOF
;

procedure_entry:
    procedure_block EOF
;

/*
    The following section includes the parser rules that are commonly
    used among all sections of the grammar. These include literals,
*/

literal:
    STRING_                                     # string_literal
    | (PLUS | MINUS)? DIGITS_                   # integer_literal
    | (PLUS | MINUS)? DIGITS_ PERIOD DIGITS_    # decimal_literal
    | (TRUE | FALSE)                            # boolean_literal
    | NULL                                      # null_literal
    | BINARY_                                   # binary_literal
;

// identifier is used for table / column names
identifier:
    (DOUBLE_QUOTE IDENTIFIER DOUBLE_QUOTE) | IDENTIFIER
;

identifier_list:
    identifier (COMMA identifier)*
;

type:
    IDENTIFIER (LPAREN DIGITS_ COMMA DIGITS_ RPAREN)? (LBRACKET RBRACKET)? // Handles arrays of any type, including nested arrays
;

type_cast:
    TYPE_CAST type
;

variable:
    VARIABLE | CONTEXTUAL_VARIABLE
;

variable_list:
    variable (COMMA variable)*
;

/*
    The following section includes parser rules for top-level Kuneiform.
    These are the rules that parse the schema / DDL, and are used pre-consensus.
*/

// schema is the parser entrypoint for an entire
// Kuneiform schema.
schema:
    database_declaration
    (use_declaration | table_declaration
     | action_declaration | procedure_declaration
     | foreign_procedure_declaration
    )*
;

annotation:
    // sort've a hack; annotations don't technically use contextual variables, but they have
    // the same syntax of @identifier
    CONTEXTUAL_VARIABLE LPAREN (IDENTIFIER EQUALS literal (COMMA IDENTIFIER EQUALS literal)*)? RPAREN
;

database_declaration:
    DATABASE IDENTIFIER SCOL
;

use_declaration:
    USE IDENTIFIER
    (LBRACE IDENTIFIER COL literal (COMMA IDENTIFIER COL literal)* RBRACE)?
    AS IDENTIFIER SCOL
;

table_declaration:
     TABLE IDENTIFIER LBRACE
     column_def (COMMA (column_def | index_def | foreign_key_def))*
     RBRACE
 ;

column_def:
    name=IDENTIFIER type constraint*
;

index_def:
    HASH_IDENTIFIER
    (UNIQUE | INDEX | PRIMARY)
    LPAREN  columns=identifier_list RPAREN
;

foreign_key_def:
    (FOREIGN KEY|LEGACY_FOREIGN_KEY) // for backwards compatibility
    LPAREN child_keys=identifier_list RPAREN
    (REFERENCES|REF) parent_table=IDENTIFIER LPAREN parent_keys=identifier_list RPAREN
    foreign_key_action*
;

// variability here is to support legacy syntax
foreign_key_action:
    ((ON UPDATE|LEGACY_ON_UPDATE)|(ON DELETE|LEGACY_ON_DELETE)) DO? ((NO ACTION|LEGACY_NO_ACTION)|CASCADE|(SET NULL|LEGACY_SET_NULL)|(SET DEFAULT|LEGACY_SET_DEFAULT)|RESTRICT)
;

type_list:
    type (COMMA type)*
;

named_type_list:
    IDENTIFIER type (COMMA IDENTIFIER type)*
;

typed_variable_list:
    variable type (COMMA variable type)*
;

constraint:
    // conditionally allow some tokens, since they are used elsewhere
    (IDENTIFIER| PRIMARY KEY? | NOT NULL | DEFAULT | UNIQUE) (LPAREN literal RPAREN)?
;

access_modifier:
    PUBLIC | PRIVATE | VIEW | OWNER
;

action_declaration:
    annotation*
    ACTION IDENTIFIER
    LPAREN variable_list? RPAREN
    (access_modifier)+
    LBRACE action_block RBRACE
;

procedure_declaration:
    annotation*
    PROCEDURE IDENTIFIER
    LPAREN (typed_variable_list)? RPAREN
    (access_modifier)+
    (procedure_return)?
    LBRACE procedure_block RBRACE
;


foreign_procedure_declaration:
    FOREIGN PROCEDURE IDENTIFIER
    LPAREN (unnamed_params=type_list|named_params=typed_variable_list)? RPAREN
    (procedure_return)?
;

procedure_return:
    RETURNS (TABLE? LPAREN return_columns=named_type_list RPAREN
    | LPAREN unnamed_return_types=type_list RPAREN)
;

/*
    The following section includes parser rules for SQL.
*/

// sql is a top-level SQL statement.
sql:
    sql_statement SCOL
;

sql_statement:
    (WITH common_table_expression (COMMA common_table_expression)*)?
    (select_statement | update_statement | insert_statement | delete_statement)
;

common_table_expression:
    identifier (LPAREN (identifier (COMMA identifier)*)? RPAREN)? AS LPAREN select_statement RPAREN
;

select_statement:
    select_core
    (compound_operator select_core)*
    (ORDER BY ordering_term (COMMA ordering_term)*)?
    (LIMIT limit=sql_expr)?
    (OFFSET offset=sql_expr)?
;

compound_operator:
    UNION ALL? | INTERSECT | EXCEPT
;

ordering_term:
    sql_expr (ASC | DESC)? (NULLS (FIRST | LAST))?
;

select_core:
    SELECT DISTINCT?
    result_column (COMMA result_column)*
    (FROM relation join*)?
    (WHERE where=sql_expr)?
    (
        GROUP BY group_by=sql_expr_list
        (HAVING having=sql_expr)?
    )?
;

relation:
    table_name=identifier (AS? alias=identifier)?               # table_relation
    // aliases are technically required in Kuneiform for subquery and function calls,
    // but we allow it to pass here since it is standard SQL to not require it, and
    // we can throw a better error message after parsing.
    | LPAREN select_statement RPAREN (AS? alias=identifier)?    # subquery_relation
    | sql_function_call (AS? alias=identifier?)                 # function_relation
;

join:
    (INNER| LEFT | RIGHT | FULL)? JOIN
    relation ON sql_expr
;

result_column:
    sql_expr (AS? identifier)?              # expression_result_column
    | (table_name=identifier PERIOD)? STAR  # wildcard_result_column
;

update_statement:
    UPDATE table_name=identifier (AS? alias=identifier)?
    SET update_set_clause (COMMA update_set_clause)*
    (FROM relation join*)?
    (WHERE where=sql_expr)?
;

update_set_clause:
   column=identifier EQUALS sql_expr
;

insert_statement:
    INSERT INTO table_name=identifier (AS? alias=identifier)?
    (LPAREN target_columns=identifier_list RPAREN)?
    VALUES LPAREN sql_expr_list RPAREN (COMMA LPAREN sql_expr_list RPAREN)*
    upsert_clause?
;

upsert_clause:
    ON CONFLICT
    (LPAREN conflict_columns=identifier_list RPAREN (WHERE conflict_where=sql_expr)?)?
    DO (
        NOTHING
        | UPDATE SET update_set_clause (COMMA update_set_clause)*
        (WHERE update_where=sql_expr)?
    )
;

delete_statement:
    DELETE FROM table_name=identifier (AS? alias=identifier)?
    // (USING relation join*)?
    (WHERE where=sql_expr)?
;

// https://docs.kwil.com/docs/kuneiform/operators
sql_expr:
    // highest precedence:
    LPAREN sql_expr RPAREN type_cast?                                               # paren_sql_expr
    | sql_expr PERIOD identifier type_cast?                                         # field_access_sql_expr
    | sql_expr LBRACKET sql_expr RBRACKET type_cast?                                # array_access_sql_expr
    | <assoc=right> (PLUS|MINUS) sql_expr                                           # unary_sql_expr
    | sql_expr COLLATE identifier                                                   # collate_sql_expr
    | left=sql_expr (STAR | DIV | MOD) right=sql_expr                               # arithmetic_sql_expr
    | left=sql_expr (PLUS | MINUS) right=sql_expr                                   # arithmetic_sql_expr

    // any unspecified operator:
    | literal type_cast?                                                            # literal_sql_expr
    | sql_function_call type_cast?                                                  # function_call_sql_expr
    | variable type_cast?                                                           # variable_sql_expr
    | (table=identifier PERIOD)? column=identifier type_cast?                       # column_sql_expr 
    | CASE case_clause=sql_expr?
        (when_then_clause)+
        (ELSE else_clause=sql_expr)? END                                            # case_expr
    | (NOT? EXISTS)? LPAREN select_statement RPAREN type_cast?                      # subquery_sql_expr
    // setting precedence for arithmetic operations:
    | left=sql_expr CONCAT right=sql_expr                                           # arithmetic_sql_expr

    // the rest:
    | sql_expr NOT? IN LPAREN (sql_expr_list|select_statement) RPAREN               # in_sql_expr
    | left=sql_expr NOT? (LIKE|ILIKE) right=sql_expr                                # like_sql_expr
    | element=sql_expr (NOT)? BETWEEN lower=sql_expr AND upper=sql_expr             # between_sql_expr
    | left=sql_expr (EQUALS | EQUATE | NEQ | LT | LTE | GT | GTE) right=sql_expr    # comparison_sql_expr
    | left=sql_expr IS NOT? ((DISTINCT FROM right=sql_expr) | NULL | TRUE | FALSE)  # is_sql_expr
    | <assoc=right> (NOT) sql_expr                                                  # unary_sql_expr
    | left=sql_expr AND right=sql_expr                                              # logical_sql_expr
    | left=sql_expr OR right=sql_expr                                               # logical_sql_expr
;


when_then_clause:
    WHEN when_condition=sql_expr THEN then=sql_expr
;

sql_expr_list:
    sql_expr (COMMA sql_expr)*
;

sql_function_call:
    identifier LPAREN (DISTINCT? sql_expr_list|STAR)? RPAREN                                                #normal_call_sql
    | identifier LBRACKET dbid=sql_expr COMMA procedure=sql_expr RBRACKET LPAREN (sql_expr_list)? RPAREN    #foreign_call_sql
;

/*
    The following section includes parser rules for action blocks.
*/
// action_block is the top-level rule for an action block.
action_block:
    (action_statement SCOL)*
;

// action statements can only be 3 things:
// 1. a sql statement
// 2. a local action/procedure call.
// 3. an extension call
action_statement:
    sql_statement                                                                               # sql_action
    | IDENTIFIER LPAREN (procedure_expr_list)? RPAREN                                           # local_action
    | (variable_list EQUALS)? IDENTIFIER PERIOD IDENTIFIER LPAREN (procedure_expr_list)? RPAREN # extension_action
;

/*
    This section includes parser rules for procedures
*/

// procedure_block is the top-level rule for a procedure.
procedure_block:
    proc_statement*
;

// https://docs.kwil.com/docs/kuneiform/operators
procedure_expr:
    // highest precedence:
    LPAREN procedure_expr RPAREN type_cast?                                                     # paren_procedure_expr
    | procedure_expr PERIOD IDENTIFIER type_cast?                                               # field_access_procedure_expr
    | procedure_expr LBRACKET procedure_expr RBRACKET type_cast?                                # array_access_procedure_expr
    | <assoc=right> (PLUS|MINUS|EXCL) procedure_expr                                            # unary_procedure_expr
    | procedure_expr (STAR | DIV | MOD) procedure_expr                                          # procedure_expr_arithmetic
    | procedure_expr (PLUS | MINUS) procedure_expr                                              # procedure_expr_arithmetic

    // any unspecified operator:
    | literal type_cast?                                                                        # literal_procedure_expr
    | procedure_function_call type_cast?                                                        # function_call_procedure_expr
    | variable type_cast?                                                                       # variable_procedure_expr
    | LBRACKET (procedure_expr_list)? RBRACKET type_cast?                                       # make_array_procedure_expr
    | procedure_expr CONCAT procedure_expr                                                      # procedure_expr_arithmetic

    // the rest:
    | procedure_expr (EQUALS | EQUATE | NEQ | LT | LTE | GT | GTE) procedure_expr               # comparison_procedure_expr
    | left=procedure_expr IS NOT? ((DISTINCT FROM right=procedure_expr) | NULL | TRUE | FALSE)  # is_procedure_expr
    | <assoc=right> (NOT) procedure_expr                                                        # unary_procedure_expr
    | procedure_expr AND procedure_expr                                                         # logical_procedure_expr
    | procedure_expr OR procedure_expr                                                          # logical_procedure_expr
;

procedure_expr_list:
    procedure_expr (COMMA procedure_expr)*
;

proc_statement:
    VARIABLE type SCOL                                                                                  # stmt_variable_declaration
    // stmt_procedure_call must go above stmt_variable_assignment due to lexer ambiguity
    | ((variable_or_underscore) (COMMA (variable_or_underscore))* ASSIGN)? procedure_function_call SCOL # stmt_procedure_call
    | procedure_expr type? ASSIGN procedure_expr SCOL                                                         # stmt_variable_assignment
    | FOR receiver=VARIABLE IN (range|target_variable=variable|sql_statement) LBRACE proc_statement* RBRACE  # stmt_for_loop
    | IF if_then_block (ELSEIF if_then_block)* (ELSE LBRACE proc_statement* RBRACE)?                         # stmt_if
    | sql_statement SCOL                                                                                # stmt_sql
    | BREAK SCOL                                                                                        # stmt_break
    | RETURN (procedure_expr_list|sql_statement)? SCOL                                                   # stmt_return
    | RETURN NEXT procedure_expr_list SCOL                                                              # stmt_return_next
;

variable_or_underscore:
    VARIABLE | UNDERSCORE
;

procedure_function_call:
    IDENTIFIER LPAREN (procedure_expr_list)? RPAREN                                                                         #normal_call_procedure
    | IDENTIFIER LBRACKET dbid=procedure_expr COMMA procedure=procedure_expr RBRACKET LPAREN (procedure_expr_list)? RPAREN  #foreign_call_procedure
;

if_then_block:
    procedure_expr LBRACE proc_statement* RBRACE
;

// range used for for loops
range:
    procedure_expr RANGE procedure_expr
;