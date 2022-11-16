package syntax

import (
	"fmt"

	"ksl"
	"ksl/syntax/lex"
)

type peeker struct {
	Tokens    lex.Tokens
	NextIndex int

	IncludeComments      bool
	IncludeNewlinesStack []bool
}

func newPeeker(tokens lex.Tokens, includeComments bool) *peeker {
	return &peeker{
		Tokens:               tokens,
		IncludeComments:      includeComments,
		IncludeNewlinesStack: []bool{true},
	}
}

func (p *peeker) Peek() lex.Token {
	ret, _ := p.nextToken(p.NextIndex)
	return ret
}

func (p *peeker) Peek2() (lex.Token, lex.Token) {
	toks := p.PeekN(2)
	return toks[0], toks[1]
}

func (p *peeker) Peek3() (lex.Token, lex.Token, lex.Token) {
	toks := p.PeekN(3)
	return toks[0], toks[1], toks[2]
}

func (p *peeker) ReadN(n int) lex.Tokens {
	ret := make(lex.Tokens, n)
	for i := 0; i < n; i++ {
		ret[i] = p.Read()
	}
	return ret
}

func (p *peeker) PeekN(n int) lex.Tokens {
	if n <= 0 {
		return nil
	}

	ret := make(lex.Tokens, n)

	idx := p.NextIndex
	var cur int
	for cur = 0; cur < n && idx < len(p.Tokens); cur++ {
		ret[cur], idx = p.nextToken(idx)
	}
	for ; cur < n; cur++ {
		ret[cur] = lex.Token{Type: lex.TokenEOF}
	}
	return ret
}

func (p *peeker) Read() (ret lex.Token) {
	ret, p.NextIndex = p.nextToken(p.NextIndex)
	return
}

func (p *peeker) NextRange() ksl.Range {
	return p.Peek().Range
}

func (p *peeker) PrevRange() ksl.Range {
	if p.NextIndex == 0 {
		return p.NextRange()
	}

	return p.Tokens[p.NextIndex-1].Range
}

func (p *peeker) nextToken(idx int) (lex.Token, int) {
	for i := idx; i < len(p.Tokens); i++ {
		tok := p.Tokens[i]
		switch tok.Type {
		case lex.TokenComment:
			if !p.IncludeComments {
				if p.includingNewlines() {
					if len(tok.Value) > 0 && tok.Value[len(tok.Value)-1] == '\n' {
						fakeNewline := lex.Token{
							Type:  lex.TokenNewline,
							Value: tok.Value[len(tok.Value)-1 : len(tok.Value)],
							Range: tok.Range,
						}
						return fakeNewline, i + 1
					}
				}

				continue
			}
		case lex.TokenNewline:
			if !p.includingNewlines() {
				continue
			}
		}

		return tok, i + 1
	}

	return p.Tokens[len(p.Tokens)-1], len(p.Tokens)
}

func (p *peeker) includingNewlines() bool {
	return p.IncludeNewlinesStack[len(p.IncludeNewlinesStack)-1]
}

func (p *peeker) PushIncludeNewlines(include bool) {
	p.IncludeNewlinesStack = append(p.IncludeNewlinesStack, include)
}

func (p *peeker) PopIncludeNewlines() bool {
	stack := p.IncludeNewlinesStack
	remain, ret := stack[:len(stack)-1], stack[len(stack)-1]
	p.IncludeNewlinesStack = remain
	return ret
}

// AssertEmptyNewlinesStack checks if the IncludeNewlinesStack is empty, doing
// panicking if it is not. This can be used to catch stack mismanagement that
// might otherwise just cause confusing downstream errors.
func (p *peeker) AssertEmptyIncludeNewlinesStack() {
	if len(p.IncludeNewlinesStack) != 1 {
		panic(fmt.Errorf("non-empty IncludeNewlinesStack after parse: %#v", p.IncludeNewlinesStack))
	}
}
