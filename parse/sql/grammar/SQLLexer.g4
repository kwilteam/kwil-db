/*
 * The MIT License (MIT)
 *
 * Copyright (c) 2020 by Martin Mirchev
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
 * Developed by : Bart Kiers, bart@big-o.nl
 */

// $antlr-format alignTrailingComments on, columnLimit 150, maxEmptyLinesToKeep 1, reflowComments off, useTab off
// $antlr-format allowShortRulesOnASingleLine on, alignSemicolons ownLine

lexer grammar SQLLexer;

options { caseInsensitive = true; }

SCOL:      ';';
DOT:       '.';
OPEN_PAR:  '(';
CLOSE_PAR: ')';
L_BRACKET: '[';
R_BRACKET: ']';
COMMA:     ',';
ASSIGN:    '=';
STAR:      '*';
PLUS:      '+';
MINUS:     '-';
DIV:       '/';
MOD:       '%';
LT:        '<';
LT_EQ:     '<=';
GT:        '>';
GT_EQ:     '>=';
NOT_EQ1:   '!=';
NOT_EQ2:   '<>';
TYPE_CAST: '::';

// http://www.sqlite.org/lang_keywords.html
ADD_:               'ADD';
ALL_:               'ALL';
AND_:               'AND';
ASC_:               'ASC';
AS_:                'AS';
BETWEEN_:           'BETWEEN';
BY_:                'BY';
CASE_:              'CASE';
COLLATE_:           'COLLATE';
COMMIT_:            'COMMIT';
CONFLICT_:          'CONFLICT';
CREATE_:            'CREATE';
CROSS_:             'CROSS';
DEFAULT_:           'DEFAULT';
DELETE_:            'DELETE';
DESC_:              'DESC';
DISTINCT_:          'DISTINCT';
DO_:                'DO';
ELSE_:              'ELSE';
END_:               'END';
ESCAPE_:            'ESCAPE';
EXCEPT_:            'EXCEPT';
EXISTS_:            'EXISTS';
FILTER_:            'FILTER';
FIRST_:             'FIRST';
FROM_:              'FROM';
FULL_:              'FULL';
GROUPS_:            'GROUPS';
GROUP_:             'GROUP';
HAVING_:            'HAVING';
INNER_:             'INNER';
INSERT_:            'INSERT';
INTERSECT_:         'INTERSECT';
INTO_:              'INTO';
IN_:                'IN';
ISNULL_:            'ISNULL';
IS_:                'IS';
JOIN_:              'JOIN';
LAST_:              'LAST';
LEFT_:              'LEFT';
LIKE_:              'LIKE';
LIMIT_:             'LIMIT';
NOTHING_:           'NOTHING';
NOTNULL_:           'NOTNULL';
NOT_:               'NOT';
NULLS_:             'NULLS';
OFFSET_:            'OFFSET';
OF_:                'OF';
ON_:                'ON';
ORDER_:             'ORDER';
OR_:                'OR';
OUTER_:             'OUTER';
RAISE_:             'RAISE';
REPLACE_:           'REPLACE';
RETURNING_:         'RETURNING';
RIGHT_:             'RIGHT';
SELECT_:            'SELECT';
SET_:               'SET';
THEN_:              'THEN';
UNION_:             'UNION';
UPDATE_:            'UPDATE';
USING_:             'USING';
VALUES_:            'VALUES';
WHEN_:              'WHEN';
WHERE_:             'WHERE';
WITH_:              'WITH';

// literals

BOOLEAN_LITERAL:
    'true'
    | 'false'
;

NUMERIC_LITERAL:
    [0-9]+
;

BLOB_LITERAL:
    '0x' [0-9a-f]+
;

TEXT_LITERAL:
    '\'' ( ~'\'' | '\'\'')* '\''
;

NULL_LITERAL: 'null';

IDENTIFIER:
    '"' (~'"' | '""')* '"' // Delimited identifiers
    | '`' (~'`' | '``')* '`'
    | [A-Z_] [A-Z_0-9]* // Ordinary identifiers
;

BIND_PARAMETER: [@$] IDENTIFIER;

SINGLE_LINE_COMMENT: '--' ~[\r\n]* (('\r'? '\n') | EOF) -> channel(HIDDEN);

MULTILINE_COMMENT: '/*' .*? '*/' -> channel(HIDDEN);

SPACES: [ \u000B\t\r\n] -> channel(HIDDEN);

UNEXPECTED_CHAR: .;
