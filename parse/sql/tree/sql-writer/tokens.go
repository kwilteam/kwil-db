package sqlwriter

import "strings"

type sqliteSymbol string

type tokenWriter struct {
	stmt *strings.Builder
}

func newTokenWriter(s *strings.Builder) *tokenWriter {
	return &tokenWriter{
		stmt: s,
	}
}

func (t *tokenWriter) token(s sqliteSymbol) {
	t.stmt.WriteString(" ")
	t.stmt.WriteString(string(s))
	t.stmt.WriteString(" ")
}

func (t *tokenWriter) Lparen() *tokenWriter {
	t.token(lparen)
	return t
}

func (t *tokenWriter) Rparen() *tokenWriter {
	t.token(rparen)
	return t
}

func (t *tokenWriter) Comma() *tokenWriter {
	t.token(comma)
	return t
}

func (t *tokenWriter) Semicolon() *tokenWriter {
	t.token(semicolon)
	return t
}

func (t *tokenWriter) Period() *tokenWriter {
	t.token(period)
	return t
}

func (t *tokenWriter) Asterisk() *tokenWriter {
	t.token(asterisk)
	return t
}

func (t *tokenWriter) Equals() *tokenWriter {
	t.token(equals)
	return t
}

func (t *tokenWriter) Not() *tokenWriter {
	t.token(not)
	return t
}

func (t *tokenWriter) And() *tokenWriter {
	t.token(and)
	return t
}

func (t *tokenWriter) Or() *tokenWriter {
	t.token(or)
	return t
}

func (t *tokenWriter) Is() *tokenWriter {
	t.token(is)
	return t
}

func (t *tokenWriter) IsNot() *tokenWriter {
	t.token(is_not)
	return t
}

func (t *tokenWriter) In() *tokenWriter {
	t.token(in)
	return t
}

func (t *tokenWriter) Escape() *tokenWriter {
	t.token(escape)
	return t
}

func (t *tokenWriter) Insert() *tokenWriter {
	t.token(insert)
	return t
}

func (t *tokenWriter) Into() *tokenWriter {
	t.token(into)
	return t
}

func (t *tokenWriter) Replace() *tokenWriter {
	t.token(replace)
	return t
}

func (t *tokenWriter) Values() *tokenWriter {
	t.token(values)
	return t
}

func (t *tokenWriter) Select() *tokenWriter {
	t.token(selectToken)
	return t
}

func (t *tokenWriter) From() *tokenWriter {
	t.token(from)
	return t
}

func (t *tokenWriter) Where() *tokenWriter {
	t.token(where)
	return t
}

func (t *tokenWriter) Having() *tokenWriter {
	t.token(having)
	return t
}

func (t *tokenWriter) Asc() *tokenWriter {
	t.token(asc)
	return t
}

func (t *tokenWriter) Desc() *tokenWriter {
	t.token(desc)
	return t
}

func (t *tokenWriter) Limit() *tokenWriter {
	t.token(limit)
	return t
}

func (t *tokenWriter) Offset() *tokenWriter {
	t.token(offset)
	return t
}

func (t *tokenWriter) All() *tokenWriter {
	t.token(all)
	return t
}

func (t *tokenWriter) Distinct() *tokenWriter {
	t.token(distinct)
	return t
}

func (t *tokenWriter) As() *tokenWriter {
	t.token(as)
	return t
}

func (t *tokenWriter) Exists() *tokenWriter {
	t.token(exists)
	return t
}

func (t *tokenWriter) Case() *tokenWriter {
	t.token(case_)
	return t
}

func (t *tokenWriter) When() *tokenWriter {
	t.token(when)
	return t
}

func (t *tokenWriter) Then() *tokenWriter {
	t.token(then)
	return t
}

func (t *tokenWriter) Else() *tokenWriter {
	t.token(else_)
	return t
}

func (t *tokenWriter) End() *tokenWriter {
	t.token(end)
	return t
}

func (t *tokenWriter) Space() *tokenWriter {
	t.token(space)
	return t
}

func (t *tokenWriter) On() *tokenWriter {
	t.token(on)
	return t
}

func (t *tokenWriter) Conflict() *tokenWriter {
	t.token(conflict)
	return t
}

func (t *tokenWriter) Do() *tokenWriter {
	t.token(do)
	return t
}

func (t *tokenWriter) Nothing() *tokenWriter {
	t.token(nothing)
	return t
}

func (t *tokenWriter) Update() *tokenWriter {
	t.token(update)
	return t
}

func (t *tokenWriter) Set() *tokenWriter {
	t.token(set)
	return t
}

func (t *tokenWriter) Collate() *tokenWriter {
	t.token(collate)
	return t
}

func (t *tokenWriter) Returning() *tokenWriter {
	t.token(returning)
	return t
}

func (t *tokenWriter) Order() *tokenWriter {
	t.token(order)
	return t
}

func (t *tokenWriter) Group() *tokenWriter {
	t.token(group)
	return t
}

func (t *tokenWriter) By() *tokenWriter {
	t.token(by)
	return t
}

func (t *tokenWriter) Null() *tokenWriter {
	t.token(null)
	return t
}

func (t *tokenWriter) Between() *tokenWriter {
	t.token(between)
	return t
}

func (t *tokenWriter) Natural() *tokenWriter {
	t.token(natural)
	return t
}

func (t *tokenWriter) Inner() *tokenWriter {
	t.token(inner)
	return t
}

func (t *tokenWriter) Left() *tokenWriter {
	t.token(left)
	return t
}

func (t *tokenWriter) Right() *tokenWriter {
	t.token(right)
	return t
}

func (t *tokenWriter) Full() *tokenWriter {
	t.token(full)
	return t
}

func (t *tokenWriter) Outer() *tokenWriter {
	t.token(outer)
	return t
}

func (t *tokenWriter) Join() *tokenWriter {
	t.token(join)
	return t
}

func (t *tokenWriter) Delete() *tokenWriter {
	t.token(delete)
	return t
}

func (t *tokenWriter) Indexed() *tokenWriter {
	t.token(indexed)
	return t
}

func (t *tokenWriter) With() *tokenWriter {
	t.token(with)
	return t
}

func (t *tokenWriter) Raise() *tokenWriter {
	t.token(raise)
	return t
}

const (
	lparen      sqliteSymbol = "("
	rparen      sqliteSymbol = ")"
	comma       sqliteSymbol = ","
	semicolon   sqliteSymbol = ";"
	period      sqliteSymbol = "."
	asterisk    sqliteSymbol = "*"
	equals      sqliteSymbol = "="
	not         sqliteSymbol = "NOT"
	and         sqliteSymbol = "AND"
	or          sqliteSymbol = "OR"
	is          sqliteSymbol = "IS"
	is_not      sqliteSymbol = "IS NOT"
	in          sqliteSymbol = "IN"
	escape      sqliteSymbol = "ESCAPE"
	insert      sqliteSymbol = "INSERT"
	into        sqliteSymbol = "INTO"
	replace     sqliteSymbol = "REPLACE"
	values      sqliteSymbol = "VALUES"
	selectToken sqliteSymbol = "SELECT" // select is a reserved word in Go
	from        sqliteSymbol = "FROM"
	where       sqliteSymbol = "WHERE"
	having      sqliteSymbol = "HAVING"
	asc         sqliteSymbol = "ASC"
	desc        sqliteSymbol = "DESC"
	limit       sqliteSymbol = "LIMIT"
	offset      sqliteSymbol = "OFFSET"
	all         sqliteSymbol = "ALL"
	distinct    sqliteSymbol = "DISTINCT"
	as          sqliteSymbol = "AS"
	exists      sqliteSymbol = "EXISTS"
	case_       sqliteSymbol = "CASE"
	when        sqliteSymbol = "WHEN"
	then        sqliteSymbol = "THEN"
	else_       sqliteSymbol = "ELSE"
	end         sqliteSymbol = "END"
	space       sqliteSymbol = " "
	on          sqliteSymbol = "ON"
	conflict    sqliteSymbol = "CONFLICT"
	do          sqliteSymbol = "DO"
	nothing     sqliteSymbol = "NOTHING"
	update      sqliteSymbol = "UPDATE"
	set         sqliteSymbol = "SET"
	collate     sqliteSymbol = "COLLATE"
	returning   sqliteSymbol = "RETURNING"
	order       sqliteSymbol = "ORDER"
	group       sqliteSymbol = "GROUP"
	by          sqliteSymbol = "BY"
	null        sqliteSymbol = "NULL"
	between     sqliteSymbol = "BETWEEN"
	natural     sqliteSymbol = "NATURAL"
	inner       sqliteSymbol = "INNER"
	left        sqliteSymbol = "LEFT"
	right       sqliteSymbol = "RIGHT"
	full        sqliteSymbol = "FULL"
	outer       sqliteSymbol = "OUTER"
	join        sqliteSymbol = "JOIN"
	delete      sqliteSymbol = "DELETE"
	indexed     sqliteSymbol = "INDEXED"
	with        sqliteSymbol = "WITH"
	raise       sqliteSymbol = "RAISE"
)
