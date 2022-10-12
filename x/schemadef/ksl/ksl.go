package ksl

import (
	"io"
	"os"
	"strings"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

type (
	Document struct {
		Entries []Entry `parser:"@@*"`
	}

	Directive struct {
		LeadingTrivia  *LeadingTrivia  `parser:"@@"`
		Kind           *Ident          `parser:"'@' @@"`
		Name           *Ident          `parser:"(@@ '=')?"`
		Value          Value           `parser:"@@"`
		TrailingTrivia *TrailingTrivia `parser:"@@"`
	}

	Resource struct {
		Pos           lexer.Position
		LeadingTrivia *LeadingTrivia `parser:"@@"`
		Kind          *Ident         `parser:"@@"`
		Name          *Name          `parser:"@@?"` // names can be optional for some resources
		Modifier      *Modifier      `parser:"@@?"` // optional modifier
		Labels        *Labels        `parser:"@@?"` // optional labels
		Fields        []Field        `parser:"'{' Newline? @@* '}' Newline?"`
	}

	Property struct {
		LeadingTrivia  *LeadingTrivia  `parser:"@@"`
		Key            *Ident          `parser:"@@"`
		Value          Value           `parser:"('=' @@)?"`
		TrailingTrivia *TrailingTrivia `parser:"@@"`
	}

	Declaration struct {
		LeadingTrivia  *LeadingTrivia  `parser:"@@"`
		Name           *Ident          `parser:"@@ ':'"`
		Type           *Type           `parser:"@@"`
		Annotations    []*Annotation   `parser:"@@*"`
		TrailingTrivia *TrailingTrivia `parser:"@@"`
	}

	BlockAnnotation struct {
		LeadingTrivia  *LeadingTrivia  `parser:"@@"`
		Annotation     *Annotation     `parser:"'@' @@"`
		TrailingTrivia *TrailingTrivia `parser:"@@"`
	}

	Labels struct {
		Values []*KeyValue `parser:"'[' @@ (',' @@)* ']'"`
	}

	Modifier struct {
		Keyword *Ident `parser:"@@"`
		Target  *Ident `parser:"@@"`
	}
	KeyValue struct {
		Key   *Ident `parser:"@@"`
		Value Value  `parser:"('=' @@)?"`
	}

	Type struct {
		IsArray  bool  `parser:"@'[]'?"`
		Name     *Name `parser:"@@"`
		Size     *Int  `parser:"('(' @@ ')')?"`
		Nullable bool  `parser:"@'?'?"`
	}

	Var struct {
		Name *Ident `parser:"'$' @@"`
	}

	Name struct {
		Qualifiers []*Ident `parser:"(@@ '.')* (?=Ident)"`
		Value      *Ident   `parser:"@@"`
	}

	Ident struct {
		Value string `parser:"@Ident"`
	}

	Float struct {
		Value float64 `parser:"@Float"`
	}

	Int struct {
		Value int `parser:"@Int"`
	}

	Str struct {
		Value string `parser:"@String"`
	}

	Expr struct {
		Value string `parser:"@Expr"`
	}

	Heredoc struct {
		Token string `parser:"@Heredoc"`
		Value string `parser:"@Block End"`
	}

	Bool struct {
		Value Boolean `parser:"@('true' | 'false')"`
	}

	FunctionCall struct {
		Name   *Ident   `parser:"@@ '('"`
		Args   []*Arg   `parser:"((@@? (',' @@)*)"`
		Kwargs []*Kwarg `parser:"@@? (',' @@)*)? ')'"`
	}

	Slice struct {
		Values []Value `parser:"'[' @@ (',' @@)* ']'"`
	}

	Annotation struct {
		Name   *Ident   `parser:"'@' @@"`
		Args   []*Arg   `parser:"('(' (@@? (',' @@)*)"`
		Kwargs []*Kwarg `parser:"@@? (',' @@)* ')')?"`
	}

	Arg struct {
		Value Value `parser:"@@ (?!'=')"`
	}

	Kwarg struct {
		Name  *Ident `parser:"@@ '='"`
		Value Value  `parser:"@@"`
	}

	LeadingTrivia struct {
		Trivia []SyntaxTrivia `parser:"@@*"`
	}

	TrailingTrivia struct {
		Comment *Comment `parser:"@@? Newline?"`
	}

	Comment struct {
		Value string `parser:"@Comment"`
	}

	Newline struct {
		Value string `parser:"@Newline"`
	}
)

type Boolean bool

func (b *Boolean) Capture(values []string) error {
	*b = values[0] == "true"
	return nil
}

type String interface{ str() }
type Entry interface{ entry() }
type Value interface{ value() }
type Field interface{ field() }
type SyntaxTrivia interface{ trivia() }

func (*Resource) entry()  {}
func (*Directive) entry() {}

func (*Ident) value()        {}
func (*Float) value()        {}
func (*Int) value()          {}
func (*Str) value()          {}
func (*Expr) value()         {}
func (*Heredoc) value()      {}
func (*Bool) value()         {}
func (*Slice) value()        {}
func (*Name) value()         {}
func (*FunctionCall) value() {}
func (*Var) value()          {}

func (*Comment) trivia() {}
func (*Newline) trivia() {}

func (*Resource) field()        {}
func (*Property) field()        {}
func (*Declaration) field()     {}
func (*BlockAnnotation) field() {}

func ParseFile(path string) (*Document, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return Parse(path, file)
}

func Parse(filename string, rd io.Reader) (*Document, error) {
	l, err := lexer.New(lexer.Rules{
		"Root": {
			{Name: "Heredoc", Pattern: `<<(\w+)\b`, Action: lexer.Push("Heredoc")},
			{Name: "Comment", Pattern: `///([^\n]*)\n?`},
			{Name: "Note", Pattern: `(?:#|//)[^\n]*\n?`},
			{Name: "Ident", Pattern: `[a-zA-Z]\w*`},
			{Name: "Expr", Pattern: "`[^`]*`"},
			{Name: "String", Pattern: `"[^"]*"`},
			{Name: "Float", Pattern: `[-+]?\d*\.\d+`},
			{Name: "Int", Pattern: `\d+`},
			{Name: "ArrayType", Pattern: `\[\]`},
			{Name: "Punct", Pattern: `[-[!@#$%^&*()+_={}\|:;"'<,>.?/]|]`},
			{Name: "Whitespace", Pattern: `[ \t\r]+`},
			{Name: "Newline", Pattern: `\n`},
		},
		"Heredoc": {
			{Name: "ws", Pattern: `\s+`},
			{Name: "End", Pattern: `\b\1\b`, Action: lexer.Pop()},
			{Name: "Block", Pattern: `.+`},
		},
	})

	if err != nil {
		return nil, err
	}

	p := participle.MustBuild[Document](
		participle.Lexer(l),
		participle.Elide("Note", "Whitespace"),
		participle.Unquote("String", "Expr"),
		participle.Map(mapComment, "Comment"),
		participle.Map(mapHeredoc, "Heredoc"),
		participle.Union[Value](&Var{}, &FunctionCall{}, &Float{}, &Int{}, &Str{}, &Expr{}, &Heredoc{}, &Bool{}, &Name{}, &Ident{}, &Slice{}),
		participle.Union[Entry](&Directive{}, &Resource{}),
		participle.Union[SyntaxTrivia](&Comment{}, &Newline{}),
		participle.Union[Field](&Declaration{}, &Resource{}, &Property{}, &BlockAnnotation{}),
		participle.UseLookahead(2),
	)
	schema, err := p.Parse(filename, rd)
	if err != nil {
		return nil, err
	}
	return schema, nil
}

func mapComment(t lexer.Token) (lexer.Token, error) {
	t.Value = strings.TrimSpace(strings.TrimPrefix(t.Value, "///"))
	return t, nil
}

func mapHeredoc(t lexer.Token) (lexer.Token, error) {
	t.Value = strings.TrimPrefix(t.Value, "<<")
	return t, nil
}
