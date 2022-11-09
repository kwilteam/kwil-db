package ast

import (
	"fmt"
	"math/big"
	"strings"

	"ksl"
	"ksl/kslsyntax/lex"
)

var _ Node = (*Var)(nil)
var _ Node = (*Number)(nil)
var _ Node = (*Object)(nil)
var _ Node = (*Float)(nil)
var _ Node = (*Int)(nil)
var _ Node = (*Str)(nil)
var _ Node = (*QuotedStr)(nil)
var _ Node = (*Heredoc)(nil)
var _ Node = (*Bool)(nil)
var _ Node = (*Null)(nil)
var _ Node = (*List)(nil)

type Var struct {
	Dollar lex.Token
	Name   string

	SrcRange ksl.Range
}

func (v *Var) GetName() string {
	if v == nil {
		return ""
	}
	return v.Name
}

type Number struct {
	Value *big.Float

	SrcRange ksl.Range
}

type Object struct {
	LBrace     lex.Token
	Attributes Attributes
	RBrace     lex.Token

	SrcRange ksl.Range
}

func (o *Object) GetAttributes() Attributes {
	if o == nil {
		return nil
	}
	return o.Attributes
}

func (o *Object) AsMap() map[string]Expr {
	if o == nil {
		return nil
	}
	ret := make(map[string]Expr)
	for _, kv := range o.GetAttributes() {
		ret[kv.GetName()] = kv.GetValue()
	}
	return ret
}

type Float struct {
	Value float64

	SrcRange ksl.Range
}

func (s *Float) GetFloat() float64 {
	if s == nil {
		return 0
	}
	return s.Value
}

type Int struct {
	Value int

	SrcRange ksl.Range
}

func (s *Int) GetInt() int {
	if s == nil {
		return 0
	}
	return s.Value
}

func (s *Int) GetInt64() int64 {
	return int64(s.GetInt())
}

type Str struct {
	Value string

	SrcRange ksl.Range
}

func (s *Str) GetString() string {
	if s == nil {
		return ""
	}
	return s.Value
}

type QuotedStr struct {
	Char  rune
	Value string

	SrcRange ksl.Range
}

func (s *QuotedStr) GetString() string {
	if s == nil {
		return ""
	}
	return s.Value
}

func (s *QuotedStr) GetQuotedString() string {
	if s == nil {
		return "\"\""
	}
	return fmt.Sprintf("%c%s%c", s.Char, s.Value, s.Char)
}

type Heredoc struct {
	Begin       lex.Token
	Marker      string
	Values      []*Str
	StripIndent bool
	End         lex.Token

	SrcRange ksl.Range
}

func (n *Heredoc) GetString() string {
	if n == nil {
		return ""
	}

	lines := make([]string, 0, len(n.Values))
	for _, line := range n.Values {
		if n.StripIndent {
			lines = append(lines, strings.TrimLeft(line.Value, " \t"))
		} else {
			lines = append(lines, line.Value)
		}
	}
	return strings.Join(lines, "")
}

type Bool struct {
	Value bool

	SrcRange ksl.Range
}

func (b *Bool) GetBool() bool {
	if b == nil {
		return false
	}
	return b.Value
}

type Null struct {
	SrcRange ksl.Range
}

type List struct {
	LBrack lex.Token
	Values []Expr
	RBrack lex.Token

	SrcRange ksl.Range
}

func (l *List) GetValues() []Expr {
	if l == nil {
		return nil
	}
	return l.Values
}
