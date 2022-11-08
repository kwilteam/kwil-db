package sqlutil

import (
	"bytes"
	"reflect"
	"strings"
)

// A Builder provides a syntactic sugar API for writing SQL statements.
type Builder struct {
	bytes.Buffer
	QuoteChar byte    // quoting identifiers
	Schema    *string // schema qualifier
}

// P writes a list of phrases to the builder separated and
// suffixed with whitespace.
func (b *Builder) P(phrases ...string) *Builder {
	for _, p := range phrases {
		if p == "" {
			continue
		}
		if b.Len() > 0 && b.lastByte() != ' ' && b.lastByte() != '(' {
			b.WriteByte(' ')
		}
		b.WriteString(p)
		if p[len(p)-1] != ' ' {
			b.WriteByte(' ')
		}
	}
	return b
}

// Ident writes the given string quoted as an SQL identifier.
func (b *Builder) Ident(s string) *Builder {
	if s != "" {
		b.WriteByte(b.QuoteChar)
		b.WriteString(s)
		b.WriteByte(b.QuoteChar)
		b.WriteByte(' ')
	}
	return b
}

// Table writes the table identifier to the builder, prefixed
// with the schema name if exists.
func (b *Builder) Table(schemaName, tableName string) *Builder {
	switch {
	// Custom qualifier.
	case b.Schema != nil:
		// Empty means skip prefix.
		if *b.Schema != "" {
			b.Ident(*b.Schema)
			b.rewriteLastByte('.')
		}
	// Default schema qualifier.
	case schemaName != "":
		b.Ident(schemaName)
		b.rewriteLastByte('.')
	}
	b.Ident(tableName)
	return b
}

// Comma writes a comma in case the buffer is not empty, or
// replaces the last char if it is a whitespace.
func (b *Builder) Comma() *Builder {
	switch {
	case b.Len() == 0:
	case b.lastByte() == ' ':
		b.rewriteLastByte(',')
		b.WriteByte(' ')
	default:
		b.WriteString(", ")
	}
	return b
}

// MapComma maps the slice x using the function f and joins the result with
// a comma separating between the written elements.
func (b *Builder) MapComma(x any, f func(i int, b *Builder)) *Builder {
	s := reflect.ValueOf(x)
	for i := 0; i < s.Len(); i++ {
		if i > 0 {
			b.Comma()
		}
		f(i, b)
	}
	return b
}

// MapCommaErr is like MapComma, but returns an error if f returns an error.
func (b *Builder) MapCommaErr(x any, f func(i int, b *Builder) error) error {
	s := reflect.ValueOf(x)
	for i := 0; i < s.Len(); i++ {
		if i > 0 {
			b.Comma()
		}
		if err := f(i, b); err != nil {
			return err
		}
	}
	return nil
}

// Wrap wraps the written string with parentheses.
func (b *Builder) Wrap(f func(b *Builder)) *Builder {
	b.WriteByte('(')
	f(b)
	if b.lastByte() != ' ' {
		b.WriteByte(')')
	} else {
		b.rewriteLastByte(')')
	}
	return b
}

// Clone returns a duplicate of the builder.
func (b *Builder) Clone() *Builder {
	return &Builder{
		QuoteChar: b.QuoteChar,
		Buffer:    *bytes.NewBufferString(b.Buffer.String()),
	}
}

// String overrides the Buffer.String method and ensure no spaces pad the returned statement.
func (b *Builder) String() string {
	return strings.TrimSpace(b.Buffer.String())
}

func (b *Builder) lastByte() byte {
	if b.Len() == 0 {
		return 0
	}
	buf := b.Buffer.Bytes()
	return buf[len(buf)-1]
}

func (b *Builder) rewriteLastByte(c byte) {
	if b.Len() == 0 {
		return
	}
	buf := b.Buffer.Bytes()
	buf[len(buf)-1] = c
}
