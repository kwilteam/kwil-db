package tree

import (
	"fmt"
	"strings"
)

type sqlBuilder struct {
	stmt strings.Builder
}

func newSQLBuilder() *sqlBuilder {
	return &sqlBuilder{
		stmt: strings.Builder{},
	}
}

func (b *sqlBuilder) Write(tokens ...SqliteKeywords) {
	for _, token := range tokens {
		b.stmt.WriteString(string(token))
	}
}
func (b *sqlBuilder) WriteString(s ...string) {
	for _, str := range s {
		b.stmt.WriteString(str)
	}
}

func (b *sqlBuilder) WriteInt64(is ...int64) {
	for _, i := range is {
		b.stmt.WriteString(fmt.Sprint(i))
	}
}

func (b *sqlBuilder) WriteIdent(s string) {
	b.stmt.WriteString(`"`)
	b.stmt.WriteString(s)
	b.stmt.WriteString(`"`)
}

func (b *sqlBuilder) String() string {
	return b.stmt.String()
}

// maybe rename this, since it includes things like parentheses, commas, etc.
type SqliteKeywords string

const (
	LPAREN                SqliteKeywords = "("
	RPAREN                SqliteKeywords = ")"
	COMMA                 SqliteKeywords = ","
	SEMICOLON             SqliteKeywords = ";"
	PERIOD                SqliteKeywords = "."
	ASTERISK              SqliteKeywords = "*"
	PLUS                  SqliteKeywords = "+"
	MINUS                 SqliteKeywords = "-"
	DIVIDE                SqliteKeywords = "/"
	MOD                   SqliteKeywords = "%"
	EQUALS                SqliteKeywords = "="
	NOT_EQUALS            SqliteKeywords = "!="
	LESS_THAN             SqliteKeywords = "<"
	GREATER_THAN          SqliteKeywords = ">"
	LESS_THAN_OR_EQUAL    SqliteKeywords = "<="
	GREATER_THAN_OR_EQUAL SqliteKeywords = ">="
	NOT                   SqliteKeywords = "NOT"
	AND                   SqliteKeywords = "AND"
	OR                    SqliteKeywords = "OR"
	IS                    SqliteKeywords = "IS"
	IS_NOT                SqliteKeywords = "IS NOT"
	IN                    SqliteKeywords = "IN"
	NOT_IN                SqliteKeywords = "NOT IN"
	LIKE                  SqliteKeywords = "LIKE"
	GLOB                  SqliteKeywords = "GLOB"
	REGEXP                SqliteKeywords = "REGEXP"
	MATCH                 SqliteKeywords = "MATCH"
	ESCAPE                SqliteKeywords = "ESCAPE"
	INSERT                SqliteKeywords = "INSERT"
	INTO                  SqliteKeywords = "INTO"
	REPLACE               SqliteKeywords = "REPLACE"
	VALUES                SqliteKeywords = "VALUES"
	SELECT                SqliteKeywords = "SELECT"
	FROM                  SqliteKeywords = "FROM"
	WHERE                 SqliteKeywords = "WHERE"
	GROUP_BY              SqliteKeywords = "GROUP BY"
	HAVING                SqliteKeywords = "HAVING"
	ORDER_BY              SqliteKeywords = "ORDER BY"
	ASC                   SqliteKeywords = "ASC"
	DESC                  SqliteKeywords = "DESC"
	LIMIT                 SqliteKeywords = "LIMIT"
	OFFSET                SqliteKeywords = "OFFSET"
	ALL                   SqliteKeywords = "ALL"
	DISTINCT              SqliteKeywords = "DISTINCT"
	AS                    SqliteKeywords = "AS"
	EXISTS                SqliteKeywords = "EXISTS"
	CAST                  SqliteKeywords = "CAST"
	CASE                  SqliteKeywords = "CASE"
	WHEN                  SqliteKeywords = "WHEN"
	THEN                  SqliteKeywords = "THEN"
	ELSE                  SqliteKeywords = "ELSE"
	END                   SqliteKeywords = "END"
	EXTRACT               SqliteKeywords = "EXTRACT"
	SPACE                 SqliteKeywords = " "
	ON                    SqliteKeywords = "ON"
	CONFLICT              SqliteKeywords = "CONFLICT"
	DO                    SqliteKeywords = "DO"
	NOTHING               SqliteKeywords = "NOTHING"
	UPDATE                SqliteKeywords = "UPDATE"
	SET                   SqliteKeywords = "SET"
	COLLATE               SqliteKeywords = "COLLATE"
	RETURNING             SqliteKeywords = "RETURNING"
	ORDER                 SqliteKeywords = "ORDER"
	GROUP                 SqliteKeywords = "GROUP"
	BY                    SqliteKeywords = "BY"
	NULL                  SqliteKeywords = "NULL"
	BETWEEN               SqliteKeywords = "BETWEEN"
	NATURAL               SqliteKeywords = "NATURAL"
	INNER                 SqliteKeywords = "INNER"
	LEFT                  SqliteKeywords = "LEFT"
	RIGHT                 SqliteKeywords = "RIGHT"
	FULL                  SqliteKeywords = "FULL"
	OUTER                 SqliteKeywords = "OUTER"
	JOIN                  SqliteKeywords = "JOIN"
)
