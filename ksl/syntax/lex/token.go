package lex

import (
	"bytes"
	"fmt"
	"unsafe"

	"ksl"

	"github.com/apparentlymart/go-textseg/v13/textseg"
)

// Token represents a sequence of bytes from some KSL code that has been
// tagged with a type and its range within the source file.
type Token struct {
	Type  TokenType
	Value string
	Range ksl.Range
}

func (t Token) IsAny(ty ...TokenType) bool {
	for _, t2 := range ty {
		if t.Type == t2 {
			return true
		}
	}
	return false
}

// Tokens is a slice of Token.
type Tokens []Token

// TokenType is an enumeration used for the Type field on Token.
type TokenType rune

const (
	// Single-character tokens are represented by their own character, for
	// convenience in producing these within the scanner. However, the values
	// are otherwise arbitrary and just intended to be mnemonic for humans
	// who might see them in debug output.

	TokenLBrace       TokenType = '{'
	TokenRBrace       TokenType = '}'
	TokenLBrack       TokenType = '['
	TokenRBrack       TokenType = ']'
	TokenLParen       TokenType = '('
	TokenRParen       TokenType = ')'
	TokenHeredocBegin TokenType = 'H'
	TokenHeredocEnd   TokenType = 'h'
	TokenEqual        TokenType = '='
	TokenBang         TokenType = '!'
	TokenDot          TokenType = '.'
	TokenComma        TokenType = ','
	TokenQuestion     TokenType = '?'
	TokenColon        TokenType = ':'
	TokenAt           TokenType = '@'
	TokenDollar       TokenType = '$'
	TokenModel        TokenType = 'M'
	TokenEnum         TokenType = 'E'

	TokenBoolLit        TokenType = 'B'
	TokenNullLit        TokenType = 'n'
	TokenQuotedLit      TokenType = 'Q' // might contain backslash escapes
	TokenStringLit      TokenType = 'S' // cannot contain backslash escapes
	TokenNumberLit      TokenType = 'N'
	TokenIntegerLit     TokenType = 'd'
	TokenFloatLit       TokenType = 'f'
	TokenIdent          TokenType = 'i'
	TokenQualifiedIdent TokenType = 'I'
	TokenComment        TokenType = 'C'
	TokenDocComment     TokenType = 'D'

	TokenNewline TokenType = '\n'
	TokenEOF     TokenType = '␄'

	// The rest are not used in the language but recognized by the scanner so
	// we can generate good diagnostics in the parser when users try to write
	// things that might work in other languages they are familiar with, or
	// simply make incorrect assumptions about the KSL language.

	TokenApostrophe    TokenType = '\''
	TokenBacktick      TokenType = '`'
	TokenSemicolon     TokenType = ';'
	TokenTabs          TokenType = '␉'
	TokenInvalid       TokenType = '�'
	TokenBadUTF8       TokenType = '💩'
	TokenQuotedNewline TokenType = '␤'

	// TokenNil is a placeholder for when a token is required but none is
	// available, e.g. when reporting errors. The scanner will never produce
	// this as part of a token stream.
	TokenNil TokenType = '\x00'
)

func (t TokenType) GoString() string {
	return fmt.Sprintf("syntax.%s", t.String())
}

type tokenAccum struct {
	Filename  string
	Bytes     []byte
	Pos       ksl.Pos
	Tokens    []Token
	StartByte int
}

func (f *tokenAccum) emitToken(ty TokenType, startOfs, endOfs int) {
	// Walk through our buffer to figure out how much we need to adjust
	// the start pos to get our end pos.

	start := f.Pos
	start.Column += startOfs + f.StartByte - f.Pos.Offset // Safe because only ASCII spaces can be in the offset
	start.Offset = startOfs + f.StartByte

	end := start
	end.Offset = endOfs + f.StartByte
	b := f.Bytes[startOfs:endOfs]
	data := b
	for len(b) > 0 {
		advance, seq, _ := textseg.ScanGraphemeClusters(b, true)
		if (len(seq) == 1 && seq[0] == '\n') || (len(seq) == 2 && seq[0] == '\r' && seq[1] == '\n') {
			end.Line++
			end.Column = 1
		} else {
			end.Column++
		}
		b = b[advance:]
	}

	f.Pos = end
	f.Tokens = append(f.Tokens, Token{
		Type:  ty,
		Value: *(*string)(unsafe.Pointer(&data)),
		Range: ksl.Range{
			Filename: f.Filename,
			Start:    start,
			End:      end,
		},
	})
}

type heredocInProgress struct {
	Marker      []byte
	StartOfLine bool
}

// checkInvalidTokens does a simple pass across the given tokens and generates
// diagnostics for tokens that should _never_ appear in KSL source. This
// is intended to avoid the need for the parser to have special support
// for them all over.
//
// Returns a diagnostics with no errors if everything seems acceptable.
// Otherwise, returns zero or more error diagnostics, though tries to limit
// repetition of the same information.
func CheckInvalidTokens(tokens Tokens) ksl.Diagnostics {
	var diags ksl.Diagnostics

	toldBacktick := 0
	toldApostrophe := 0
	toldSemicolon := 0
	toldTabs := 0
	toldBadUTF8 := 0

	for _, tok := range tokens {
		tokRange := func() *ksl.Range {
			r := tok.Range
			return &r
		}

		switch tok.Type {
		case TokenBacktick:
			// Only report for alternating (even) backticks, so we won't report both start and ends of the same
			// backtick-quoted string.
			if (toldBacktick % 2) == 0 {
				diags = append(diags, &ksl.Diagnostic{
					Severity: ksl.DiagError,
					Summary:  "Invalid character",
					Detail:   "The \"`\" character is not valid. To create a multi-line string, use the \"heredoc\" syntax, like \"<<EOT\".",
					Subject:  tokRange(),
				})
			}
			if toldBacktick <= 2 {
				toldBacktick++
			}
		case TokenApostrophe:
			if (toldApostrophe % 2) == 0 {
				newDiag := &ksl.Diagnostic{
					Severity: ksl.DiagError,
					Summary:  "Invalid character",
					Detail:   "Single quotes are not valid. Use double quotes (\") to enclose strings.",
					Subject:  tokRange(),
				}
				diags = append(diags, newDiag)
			}
			if toldApostrophe <= 2 {
				toldApostrophe++
			}
		case TokenSemicolon:
			if toldSemicolon < 1 {
				diags = append(diags, &ksl.Diagnostic{
					Severity: ksl.DiagError,
					Summary:  "Invalid character",
					Detail:   "The \";\" character is not valid. Use newlines to separate arguments and blocks, and commas to separate items in collection values.",
					Subject:  tokRange(),
				})

				toldSemicolon++
			}
		case TokenTabs:
			if toldTabs < 1 {
				diags = append(diags, &ksl.Diagnostic{
					Severity: ksl.DiagError,
					Summary:  "Invalid character",
					Detail:   "Tab characters may not be used. The recommended indentation style is four spaces per indent.",
					Subject:  tokRange(),
				})

				toldTabs++
			}
		case TokenBadUTF8:
			if toldBadUTF8 < 1 {
				diags = append(diags, &ksl.Diagnostic{
					Severity: ksl.DiagError,
					Summary:  "Invalid character encoding",
					Detail:   "All input files must be UTF-8 encoded. Ensure that UTF-8 encoding is selected in your editor.",
					Subject:  tokRange(),
				})

				toldBadUTF8++
			}
		case TokenQuotedNewline:
			diags = append(diags, &ksl.Diagnostic{
				Severity: ksl.DiagError,
				Summary:  "Invalid multi-line string",
				Detail:   "Quoted strings may not be split over multiple lines. To produce a multi-line string, either use the \\n escape to represent a newline character or use the \"heredoc\" multi-line template syntax.",
				Subject:  tokRange(),
			})
		case TokenInvalid:
			chars := string(tok.Value)
			switch chars {
			case "“", "”":
				diags = append(diags, &ksl.Diagnostic{
					Severity: ksl.DiagError,
					Summary:  "Invalid character",
					Detail:   "\"Curly quotes\" are not valid here. These can sometimes be inadvertently introduced when sharing code via documents or discussion forums. It might help to replace the character with a \"straight quote\".",
					Subject:  tokRange(),
				})
			default:
				diags = append(diags, &ksl.Diagnostic{
					Severity: ksl.DiagError,
					Summary:  "Invalid character",
					Detail:   "This character is not used within the language.",
					Subject:  tokRange(),
				})
			}
		}
	}
	return diags
}

var utf8BOM = []byte{0xef, 0xbb, 0xbf}

// stripUTF8BOM checks whether the given buffer begins with a UTF-8 byte order
// mark (0xEF 0xBB 0xBF) and, if so, returns a truncated slice with the same
// backing array but with the BOM skipped.
//
// If there is no BOM present, the given slice is returned verbatim.
func stripUTF8BOM(src []byte) []byte {
	if bytes.HasPrefix(src, utf8BOM) {
		return src[3:]
	}
	return src
}
