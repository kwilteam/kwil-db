/*
 * The MIT License (MIT)
 *
 * Copyright (c) 2014 by Bart Kiers
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy of this software and
 * associated documentation files (the "Software"), to deal in the Software without restriction,
 * including without limitation the rights to use, copy, modify, merge, publish, distribute,
 * sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all copies or
 * substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT
 * NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
 * NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM,
 * DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 *
 * Project : sqlite-parser; an ANTLR4 grammar for SQLite https://github.com/bkiers/sqlite-parser
 * Developed by:
 *     Bart Kiers, bart@big-o.nl
 *     Martin Mirchev, marti_2203@abv.bg
 *     Mike Lische, mike@lischke-online.de
 */

// $antlr-format alignTrailingComments on, columnLimit 130, minEmptyLines 1, maxEmptyLinesToKeep 1, reflowComments off
// $antlr-format useTab off, allowShortRulesOnASingleLine off, allowShortBlocksOnASingleLine on, alignSemicolons ownLine

parser grammar SQLParser;

options {
    tokenVocab = SQLLexer;
}

statements: (sql_stmt_list)* EOF
;

sql_stmt_list:
    SCOL* sql_stmt (SCOL+ sql_stmt)* SCOL*
;

sql_stmt: (
        delete_stmt
        | insert_stmt
        | select_stmt
        | update_stmt
    )
;

indexed_column: column_name
;

cte_table_name:
    table_name (OPEN_PAR column_name (COMMA column_name)* CLOSE_PAR)?
;

common_table_expression:
    cte_table_name AS_ OPEN_PAR select_core CLOSE_PAR
;

common_table_stmt: //additional structures
    WITH_ common_table_expression (COMMA common_table_expression)*
;

delete_core:
    DELETE_ FROM_ qualified_table_name
    (WHERE_ expr)?
    returning_clause?
;

delete_stmt:
    common_table_stmt?
    delete_core
;

variable:
    BIND_PARAMETER
;

function_call:
    function_name OPEN_PAR ((DISTINCT_? expr_list) | STAR)? CLOSE_PAR #normal_function_call
    | IDENTIFIER L_BRACKET dbid=expr COMMA procedure=expr R_BRACKET OPEN_PAR expr_list? CLOSE_PAR #foreign_function_call
;


column_ref:
    (table_name DOT)? column_name
;

when_clause:
    WHEN_ condition=expr THEN_ result=expr
;

/*
 https://www.postgresql.org/docs/16/sql-syntax-lexical.html#SQL-PRECEDENCE

 Operator/Element	        Associativity	Description
 .                              left	        table/column name separator
 ::                             left	        PostgreSQL-style typecast
 [ ]                            left	        array element selection
 + -                            right	        unary plus, unary minus
 COLLATE                        left	        collation selection
 AT                             left	        AT TIME ZONE
 ^                              left	        exponentiation
 * / %                          left	        multiplication, division, modulo
 + -                            left	        addition, subtraction
 (any other operator)           left	        all other native and user-defined operators
 BETWEEN IN LIKE ILIKE SIMILAR                  range containment, set membership, string matching
 < > = <= >= <>                                 comparison operators
 IS ISNULL NOTNULL                              IS TRUE, IS FALSE, IS NULL, IS DISTINCT FROM, etc.
 NOT                            right	        logical negation
 AND                            left	        logical conjunction
 OR                             left	        logical disjunction

===========
Another way to layout expr rules is to group them by the same level of precedence,
expr:
    bool_expr
    | ...
;
bool_expr:
    predicate_expr
    | ...
;
predicate_expr:
    arithmatic_expr
    | ...
;
arithmatic_expr:
    primary_expr
    | ...
;
primary_expr:
    literal_expr
    | ...
;
============
Type cast can only be applied to:
    literal_value
    BIND_PARAMETER
    column_name
    () parenthesesed expr
    function_call
*/
expr:
    TEXT_LITERAL type_cast?                                              #text_literal_expr
    | BOOLEAN_LITERAL type_cast?                                         #boolean_literal_expr
    | INT_LITERAL type_cast?                                             #int_literal_expr
    | NULL_LITERAL type_cast?                                            #null_literal_expr
    | BLOB_LITERAL type_cast?                                            #blob_literal_expr
    | FIXED_LITERAL type_cast?                                           #fixed_literal_expr
    | variable type_cast?                                                #variable_expr
    | column_ref type_cast?                                              #column_expr
    | <assoc=right>  operator=(MINUS | PLUS) expr                        #unary_expr
    | expr COLLATE_ collation_name                                       #collate_expr
    | OPEN_PAR expr CLOSE_PAR type_cast?                                 #parenthesized_expr
    | ((NOT_)? EXISTS_)? subquery                                        #subquery_expr
    |  CASE_ case_clause=expr?
        when_clause+
        (ELSE_ else_clause=expr)? END_                                   #case_expr
    | OPEN_PAR expr_list CLOSE_PAR                                       #expr_list_expr
    | function_call type_cast?                                           #function_expr
    // arithmetic expressions
    // exponentiation
    | left=expr operator=(STAR|DIV|MOD) right=expr                       #arithmetic_expr
    | left=expr operator=(PLUS|MINUS) right=expr                         #arithmetic_expr
    // boolean expressions
    // predicate
    | elem=expr NOT_? operator=IN_ subquery                              #in_subquery_expr
    | elem=expr NOT_? operator=IN_ OPEN_PAR expr_list CLOSE_PAR          #in_list_expr
    | elem=expr NOT_? operator=BETWEEN_ low=expr AND_ high=expr          #between_expr
    | elem=expr NOT_? operator=LIKE_ pattern=expr (ESCAPE_ escape=expr)? #like_expr
    // comparison
    | left=expr comparisonOperator right=expr                            #comparison_expr
    //| left=expr comparisonOperator right=subquery                      #scalar_subquery_expr
    | expr IS_ NOT_? ((DISTINCT_ FROM_ expr)|BOOLEAN_LITERAL|NULL_LITERAL)#is_expr
    | expr (ISNULL_ | NOTNULL_)                                          #null_expr
    // logical expressions
    | <assoc=right> NOT_ expr                                             #logical_not_expr
    | left=expr operator=AND_ right=expr                                 #logical_binary_expr
    | left=expr operator=OR_ right=expr                                  #logical_binary_expr
;

subquery:
    OPEN_PAR select_core CLOSE_PAR // note: don't support with clause in subquery
;

expr_list:
    expr (COMMA expr)*
;

comparisonOperator:
    LT|LT_EQ|GT|GT_EQ|ASSIGN|NOT_EQ1|NOT_EQ2
;

cast_type:
    IDENTIFIER (L_BRACKET R_BRACKET)?
;

type_cast:
    TYPE_CAST cast_type
;

value_row:
    OPEN_PAR expr (COMMA expr)* CLOSE_PAR
;

values_clause:
    VALUES_ value_row (COMMA value_row)*
;

insert_core:
    INSERT_ INTO_ table_name
    (AS_ table_alias)?
    (OPEN_PAR column_name ( COMMA column_name)* CLOSE_PAR)?
    values_clause
    upsert_clause?
    returning_clause?
;

insert_stmt:
    common_table_stmt?
    insert_core
;

returning_clause:
    RETURNING_ returning_clause_result_column (COMMA returning_clause_result_column)*
;

// @yaiba eaiser to parse this way
upsert_update:
    (column_name | column_name_list) ASSIGN expr
;

upsert_clause:
    ON_ CONFLICT_
    (OPEN_PAR indexed_column (COMMA indexed_column)* CLOSE_PAR (WHERE_ target_expr=expr)?)?
    DO_
    (
        NOTHING_
        | UPDATE_ SET_
            (
                upsert_update (COMMA upsert_update)*
                (WHERE_ update_expr=expr)?
            )
    )
;

select_core:
    simple_select
    (compound_operator simple_select)*
    order_by_stmt?
    limit_stmt?
;

select_stmt:
    common_table_stmt?
    select_core
;

join_relation:
    join_operator right_relation=table_or_subquery join_constraint
;

relation:
    table_or_subquery join_relation*
;

simple_select:
    SELECT_ DISTINCT_?
    result_column (COMMA result_column)*
    (FROM_ relation)?
    (WHERE_ whereExpr=expr)?
    (
      GROUP_ BY_ groupByExpr+=expr (COMMA groupByExpr+=expr)*
      (HAVING_ havingExpr=expr)?
    )?
;

table_or_subquery:
    function_call (AS_ table_alias)?
    | table_name (AS_ table_alias)?
    | OPEN_PAR select_core CLOSE_PAR (AS_ table_alias)?
;

result_column:
    STAR
    | table_name DOT STAR
    | expr (AS_ column_alias)?
;

returning_clause_result_column:
    STAR
    | expr (AS_ column_alias)?
;

join_operator:
    ((LEFT_ | RIGHT_ | FULL_) OUTER_? | INNER_)?
    JOIN_
;

join_constraint:
    ON_ expr
;

compound_operator:
    UNION_ ALL_?
    | INTERSECT_
    | EXCEPT_
;

update_set_subclause:
    (column_name | column_name_list) ASSIGN expr
;

update_core:
    UPDATE_
    qualified_table_name
    SET_ update_set_subclause (COMMA update_set_subclause)*
    (FROM_ relation)?
    (WHERE_ expr)?
    returning_clause?
;

update_stmt:
    common_table_stmt?
    update_core
;

column_name_list:
    OPEN_PAR column_name (COMMA column_name)* CLOSE_PAR
;

qualified_table_name:
    table_name (AS_ table_alias)?
;

order_by_stmt:
    ORDER_ BY_ ordering_term (COMMA ordering_term)*
;

limit_stmt:
    LIMIT_ expr (OFFSET_ expr)?
;

ordering_term:
    expr
    asc_desc?
    (NULLS_ (FIRST_ | LAST_))?
;

asc_desc:
    ASC_
    | DESC_
;


// function_keywords are keywords also function names
function_keyword:
    LIKE_
    | REPLACE_
;

function_name:
    IDENTIFIER
    | function_keyword
;

table_name:
    IDENTIFIER
;

table_alias:
    IDENTIFIER
;

column_name:
    IDENTIFIER
;

column_alias:
    IDENTIFIER
;

collation_name:
    IDENTIFIER      // back-compatible with sqlite, NOCASE
;

index_name:
    IDENTIFIER
;

