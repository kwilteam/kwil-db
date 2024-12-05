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

entry:
    statement* EOF  // entry point for the parser
;

statement:
    (
        sql_statement
        | create_table_statement
        | alter_table_statement
        | drop_table_statement
        | create_index_statement
        | drop_index_statement
        | create_role_statement
        | drop_role_statement
        | grant_statement
        | revoke_statement
        | transfer_ownership_statement
        | create_action_statement
        | drop_action_statement
    ) SCOL
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
    identifier (LPAREN DIGITS_ COMMA DIGITS_ RPAREN)? (LBRACKET RBRACKET)? // Handles arrays of any type, including nested arrays
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

column_def:
    name=identifier type constraint*
;

table_column_def:
    name=identifier type inline_constraint*
;

type_list:
    type (COMMA type)*
;

named_type_list:
    identifier type (COMMA identifier type)*
;

typed_variable_list:
    variable type (COMMA variable type)*
;

constraint:
    // conditionally allow some tokens, since they are used elsewhere
    (identifier| PRIMARY KEY? | NOT NULL | DEFAULT | UNIQUE) (LPAREN literal RPAREN)?
;

inline_constraint:
    PRIMARY KEY
    | UNIQUE
    | NOT NULL
    | DEFAULT literal
    | fk_constraint
    | CHECK (LPAREN sql_expr RPAREN)
;

fk_action:
    ON (UPDATE|DELETE)
    (SET NULL | SET DEFAULT | RESTRICT | NO ACTION | CASCADE)
;

fk_constraint:
    REFERENCES table=identifier LPAREN identifier_list RPAREN (fk_action (fk_action)?)? // can be up to 0-2 actions
;

procedure_return:
    RETURNS (TABLE? LPAREN return_columns=named_type_list RPAREN
    | LPAREN unnamed_return_types=type_list RPAREN)
;

/*
    The following section includes parser rules for SQL.
*/

// sql_stmt is a top-level SQL statement, it maps to SQLStmt interface in AST.
sql_stmt:
    (sql_statement | ddl_stmt) SCOL
;

ddl_stmt:
    create_table_statement
    | alter_table_statement
    | drop_table_statement
    | create_index_statement
    | drop_index_statement
    | create_role_statement
    | drop_role_statement
    | grant_statement
    | revoke_statement
    | transfer_ownership_statement
    | create_action_statement
    | drop_action_statement
;

sql_statement: // NOTE: This is only DDL. We should combine ddl and dml into sql_stmt in the future.
    (WITH RECURSIVE? common_table_expression (COMMA common_table_expression)*)?
    (select_statement | update_statement | insert_statement | delete_statement)
;

common_table_expression:
    identifier (LPAREN (identifier (COMMA identifier)*)? RPAREN)? AS LPAREN select_statement RPAREN
;

create_table_statement:
    CREATE TABLE (IF NOT EXISTS)? name=identifier
    LPAREN
    (table_column_def | table_constraint_def)
    (COMMA  (table_column_def | table_constraint_def))*
    RPAREN
;

table_constraint_def:
    (CONSTRAINT name=identifier)?
    (
    UNIQUE LPAREN identifier_list RPAREN
     | CHECK LPAREN sql_expr RPAREN
     | FOREIGN KEY LPAREN identifier_list RPAREN fk_constraint
     | PRIMARY KEY LPAREN identifier_list RPAREN
     )
;

opt_drop_behavior:
    CASCADE
    | RESTRICT
;

drop_table_statement:
    DROP TABLE (IF EXISTS)? tables=identifier_list opt_drop_behavior?
;

alter_table_statement:
    ALTER TABLE table=identifier
    alter_table_action
;

alter_table_action:
      ALTER COLUMN column=identifier SET (NOT NULL | DEFAULT literal)   # add_column_constraint
    | ALTER COLUMN column=identifier DROP (NOT NULL | DEFAULT)          # drop_column_constraint
    | ADD COLUMN column=identifier type                                 # add_column
    | DROP COLUMN column=identifier                                     # drop_column
    | RENAME COLUMN old_column=identifier TO new_column=identifier      # rename_column
    | RENAME TO new_table=identifier                                    # rename_table
    | ADD table_constraint_def                                          # add_table_constraint
    | DROP CONSTRAINT identifier                                        # drop_table_constraint
;

create_index_statement:
    CREATE UNIQUE? INDEX (IF NOT EXISTS)? name=identifier?
    ON table=identifier LPAREN  columns=identifier_list RPAREN
;

drop_index_statement:
    DROP INDEX (IF EXISTS)? name=identifier
;

create_role_statement:
    CREATE ROLE (IF NOT EXISTS)? identifier
;

drop_role_statement:
    DROP ROLE (IF EXISTS)? identifier
;

grant_statement:
    GRANT (privilege_list|grant_role=identifier) (ON namespace=identifier) TO (role=identifier|user=STRING_)
;

revoke_statement:
    REVOKE (privilege_list|grant_role=identifier) (ON namespace=identifier) FROM (role=identifier|user=STRING_)
;

privilege_list:
    privilege (COMMA privilege)*
;

privilege:
    SELECT | INSERT | UPDATE | DELETE | CREATE | DROP | ALTER | ROLES | CALL
;

transfer_ownership_statement:
    TRANSFER OWNERSHIP TO identifier
;

create_action_statement:
    CREATE ACTION ((IF NOT EXISTS)|(OR REPLACE))? identifier
    LPAREN (VARIABLE type (COMMA VARIABLE type)*)? RPAREN
    (PUBLIC | PRIVATE) (OWNER | VIEW)*
    procedure_return?
    LBRACE proc_statement* RBRACE
;

drop_action_statement:
    DROP ACTION (IF EXISTS)? identifier
;

use_extension_statement:
    USE extension_name=identifier (IF NOT EXISTS)?
    (LBRACE (identifier COL literal (COMMA identifier COL literal)*)? RBRACE)?
    AS alias=identifier
;

unuse_extension_statement:
    UNUSE alias=identifier (IF EXISTS)?
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
    (WINDOW identifier AS window (COMMA identifier AS window)*)?
;

relation:
    table_name=identifier (AS? alias=identifier)?               # table_relation
    // aliases are technically required in Kuneiform for subquery and function calls,
    // but we allow it to pass here since it is standard SQL to not require it, and
    // we can throw a better error message after parsing.
    | LPAREN select_statement RPAREN (AS? alias=identifier)?    # subquery_relation
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
    (
        (VALUES LPAREN sql_expr_list RPAREN (COMMA LPAREN sql_expr_list RPAREN)*)
        | (select_statement)
    )
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
    LPAREN sql_expr RPAREN type_cast?                                                       # paren_sql_expr
    | sql_expr PERIOD identifier type_cast?                                                 # field_access_sql_expr
    | array_element=sql_expr LBRACKET (
        // can be arr[1], arr[1:2], arr[1:], arr[:2], arr[:]
            single=sql_expr
            | (left=sql_expr? COL right=sql_expr?)
        ) RBRACKET type_cast?                                                               # array_access_sql_expr
    | <assoc=right> (PLUS|MINUS) sql_expr                                                   # unary_sql_expr
    | sql_expr COLLATE identifier                                                           # collate_sql_expr
    | left=sql_expr (STAR | DIV | MOD) right=sql_expr                                       # arithmetic_sql_expr
    | left=sql_expr (PLUS | MINUS) right=sql_expr                                           # arithmetic_sql_expr

    // any unspecified operator:
    | literal type_cast?                                                                    # literal_sql_expr
    // direct function calls can have a type cast, but window functions cannot
    | sql_function_call (FILTER LPAREN WHERE sql_expr RPAREN)? OVER (window|identifier)     # window_function_call_sql_expr
    | sql_function_call type_cast?                                                          # function_call_sql_expr
    | variable type_cast?                                                                   # variable_sql_expr
    | (table=identifier PERIOD)? column=identifier type_cast?                               # column_sql_expr
    | CASE case_clause=sql_expr?
        (when_then_clause)+
        (ELSE else_clause=sql_expr)? END                                                    # case_expr
    | (NOT? EXISTS)? LPAREN select_statement RPAREN type_cast?                              # subquery_sql_expr
    // setting precedence for arithmetic operations:
    | left=sql_expr CONCAT right=sql_expr                                                   # arithmetic_sql_expr

    // the rest:
    | sql_expr NOT? IN LPAREN (sql_expr_list|select_statement) RPAREN                       # in_sql_expr
    | left=sql_expr NOT? (LIKE|ILIKE) right=sql_expr                                        # like_sql_expr
    | element=sql_expr (NOT)? BETWEEN lower=sql_expr AND upper=sql_expr                     # between_sql_expr
    | left=sql_expr (EQUALS | EQUATE | NEQ | LT | LTE | GT | GTE) right=sql_expr            # comparison_sql_expr
    | left=sql_expr IS NOT? ((DISTINCT FROM right=sql_expr) | NULL | TRUE | FALSE)          # is_sql_expr
    | <assoc=right> (NOT) sql_expr                                                          # unary_sql_expr
    | left=sql_expr AND right=sql_expr                                                      # logical_sql_expr
    | left=sql_expr OR right=sql_expr                                                       # logical_sql_expr
;

window:
    LPAREN
        (PARTITION BY partition=sql_expr_list)?
        (ORDER BY ordering_term (COMMA ordering_term)*)?
    RPAREN
;


when_then_clause:
    WHEN when_condition=sql_expr THEN then=sql_expr
;

sql_expr_list:
    sql_expr (COMMA sql_expr)*
;

sql_function_call:
    identifier LPAREN (DISTINCT? sql_expr_list|STAR)? RPAREN                                                #normal_call_sql
;

/*
    The following section includes parser rules for action blocks.
*/

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
    | procedure_expr PERIOD identifier type_cast?                                               # field_access_procedure_expr
    | array_element=procedure_expr LBRACKET (
            // can be arr[1], arr[1:2], arr[1:], arr[:2], arr[:]
            single=procedure_expr
            | (left=procedure_expr? COL right=procedure_expr?)
        ) RBRACKET type_cast?                                                                   # array_access_procedure_expr
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
    (identifier PERIOD)? identifier LPAREN (procedure_expr_list)? RPAREN                                #normal_call_procedure
;

if_then_block:
    procedure_expr LBRACE proc_statement* RBRACE
;

// range used for for loops
range:
    procedure_expr RANGE procedure_expr
;