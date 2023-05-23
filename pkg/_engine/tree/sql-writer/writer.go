package sqlwriter

import (
	"fmt"
	"strings"
)

type SqlWriter struct {
	stmt  *strings.Builder
	Token *tokenWriter
}

func NewWriter() *SqlWriter {
	builder := &strings.Builder{}
	return &SqlWriter{
		stmt:  builder,
		Token: newTokenWriter(builder),
	}
}

func (s *SqlWriter) String() string {
	return s.stmt.String()
}

func (s *SqlWriter) write(str string) {
	s.stmt.WriteString(" ")
	s.stmt.WriteString(str)
	s.stmt.WriteString(" ")
}

func (s *SqlWriter) WriteString(str string) {
	s.write(str)
}

func (s *SqlWriter) WriteIdent(str string) {
	s.stmt.WriteString(` "`)
	s.stmt.WriteString(str)
	s.stmt.WriteString(`" `)
}

func (s *SqlWriter) WriteIdentNoSpace(str string) {
	s.stmt.WriteString(`"`)
	s.stmt.WriteString(str)
	s.stmt.WriteString(`"`)
}

func (s *SqlWriter) WriteInt64(i int64) {
	s.write(fmt.Sprint(i))
}

// WriteList writes a comma-separated list of strings, using the provided function to generate each string.
func (s *SqlWriter) WriteList(length int, fn func(i int)) {
	for i := 0; i < length; i++ {
		if i > 0 && i < length {
			s.Token.Comma()
		}
		fn(i)
	}
}

// WriteParenList writes a comma-separated list of strings, using the provided function to generate each string.
// The list is wrapped in parentheses.
// The first argument is the length of the list.
func (s *SqlWriter) WriteParenList(length int, fn func(i int)) {
	s.Token.Lparen()
	s.WriteList(length, fn)
	s.Token.Rparen()
}
