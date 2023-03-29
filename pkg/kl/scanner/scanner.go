package scanner

import (
	"fmt"
	"kwil/pkg/kl/token"
	"unicode"
	"unicode/utf8"
)

const (
	bom = 0xFEFF // byte order mark, only permitted as very first character
	eof = -1     // end of file
)

type ErrorHandler func(pos token.Position, msg string)

// Scanner is a lexical scanner. It takes a []byte as input and produces a stream of tokens.
type Scanner struct {
	file *token.File  // source file handle
	src  []byte       // source
	err  ErrorHandler // error reporting; or nil

	// scanning state
	ch       rune // current character
	offset   int  // current character offset
	rdOffset int  // reading offset (position after current character)

	ErrorCount int // number of errors encountered
}

func New(src []byte, err ErrorHandler, file *token.File) *Scanner {
	var s Scanner

	s.file = file
	s.err = err
	s.src = src
	s.offset = 0
	s.rdOffset = 0
	s.ch = ' '
	s.ErrorCount = 0

	// point to the first character
	s.nextChar()
	if s.ch == bom {
		s.nextChar() // ignore BOM at file beginning
	}

	return &s
}

// nextChar reads the next Unicode character
func (s *Scanner) nextChar() {
	if s.rdOffset < len(s.src) {
		s.offset = s.rdOffset

		if s.ch == '\n' {
			s.file.AddLine(s.offset)
		}

		// support unicode
		// convert byte to rune
		r, w := rune(s.src[s.rdOffset]), 1
		switch {
		case r == 0:
			s.error(s.offset, "illegal character NUL")
		case r >= utf8.RuneSelf:
			// not ASCII
			r, w = utf8.DecodeRune(s.src[s.rdOffset:])
			if r == utf8.RuneError && w == 1 {
				s.error(s.offset, "illegal UTF-8 encoding")
			} else if r == bom && s.offset > 0 {
				s.error(s.offset, "illegal byte order mark")
			}
		}
		s.rdOffset += w
		s.ch = r
	} else {
		s.offset = len(s.src)
		if s.ch == '\n' {
			s.file.AddLine(s.offset)
		}
		s.ch = eof
	}
}

func (s *Scanner) error(offs int, msg string) {
	if s.err != nil {
		s.err(s.file.Position(token.Pos(offs)), msg)
	}
	s.ErrorCount++
}

func (s *Scanner) errorf(offs int, format string, args ...any) {
	s.error(offs, fmt.Sprintf(format, args...))
}

func (s *Scanner) Init(src []byte) {
	s.src = src
	s.offset = 0
	s.rdOffset = 0
	s.ch = ' '
	s.ErrorCount = 0

	// point to the first character
	s.nextChar()
	if s.ch == bom {
		s.nextChar() // ignore BOM at file beginning
	}
}

func (s *Scanner) skipWhitespace() {
	for s.ch == ' ' || s.ch == '\t' || s.ch == '\n' || s.ch == '\r' {
		s.nextChar()
	}
}

func lower(ch rune) rune     { return ('a' - 'A') | ch } // set lower bit to get lower case letter
func isDecimal(ch rune) bool { return '0' <= ch && ch <= '9' }
func isHex(ch rune) bool     { return '0' <= ch && ch <= '9' || 'a' <= lower(ch) && lower(ch) <= 'f' }

func isLetter(ch rune) bool {
	return 'a' <= lower(ch) && lower(ch) <= 'z' || ch == '_' || ch >= utf8.RuneSelf && unicode.IsLetter(ch)
}

func isDigit(ch rune) bool {
	return isDecimal(ch) || ch >= utf8.RuneSelf && unicode.IsDigit(ch)
}

func (s *Scanner) scanIdentifier() string {
	offs := s.offset
	for isLetter(s.ch) || isDecimal(s.ch) || s.ch == '_' {
		s.nextChar()
	}
	return string(s.src[offs:s.offset])
}

func (s *Scanner) scanNumber() (tok token.Token, lit string) {
	offs := s.offset
	for isDecimal(s.ch) {
		s.nextChar()
	}
	return token.INTEGER, string(s.src[offs:s.offset])
}

func (s *Scanner) scanEscape(quote rune) bool {
	offs := s.offset

	switch s.ch {
	case 'n', '\\', quote:
		s.nextChar()
		return true
	default:
		msg := "unknown escape sequence"
		if s.ch < 0 {
			msg = "escape sequence not terminated"
		}
		s.error(offs, msg)
		return false
	}
}

func (s *Scanner) scanString(quote rune) string {
	offs := s.offset

	s.nextChar()
	for {
		ch := s.ch
		if ch == '\n' || ch < 0 {
			s.error(offs, "string literal not terminated")
			break
		}
		s.nextChar()
		if ch == quote {
			break
		}
		if ch == '\\' {
			s.scanEscape('"')
		}
	}

	return string(s.src[offs:s.offset])
}

func (s *Scanner) Next() (tok token.Token, lit string, pos token.Pos) {
	s.skipWhitespace()

	pos = token.Pos(s.offset)

	ch := s.ch

	switch ch {
	case eof:
		lit = "EOF"
		tok = token.EOF
	case '+':
		lit = "+"
		tok = token.ADD
	case '-':
		lit = "-"
		tok = token.SUB
	case '/':
		lit = "/"
		tok = token.DIV
	case '*':
		lit = "*"
		tok = token.MUL
	case '%':
		lit = "%"
		tok = token.MOD
	case '<':
		//lit = "<"
		//tok = token.LSS
		s.nextChar()
		if s.ch == '=' {
			lit = "<="
			tok = token.LEQ
		} else if s.ch == '>' {
			lit = "<>"
			tok = token.NEQ // !=
		} else {
			lit = "<"
			tok = token.LSS
			return
		}
	case '>':
		s.nextChar()
		if s.ch == '=' {
			lit = ">="
			tok = token.GEQ
		} else {
			lit = ">"
			tok = token.GTR
			return
		}
	case '=':
		lit = "="
		tok = token.ASSIGN
	case '!':
		s.nextChar()
		if s.ch == '=' {
			lit = "!="
			tok = token.NEQ
		} else {
			lit = "!"
			tok = token.NOT
			return
		}
	case '{':
		lit = "{"
		tok = token.LBRACE
	case '}':
		lit = "}"
		tok = token.RBRACE
	case '(':
		lit = "("
		tok = token.LPAREN
	case ')':
		lit = ")"
		tok = token.RPAREN
	case ',':
		lit = ","
		tok = token.COMMA
	case '.':
		lit = "."
		tok = token.PERIOD
	case ';':
		lit = ";"
		tok = token.SEMICOLON
	case '@':
		// treat @xx as identifier
		s.nextChar()
		lit = s.scanIdentifier()
		lit = "@" + lit
		tok = token.Lookup(lit)
		return
	case '$':
		// treat $xx as identifier
		s.nextChar()
		lit = s.scanIdentifier()
		lit = "$" + lit
		tok = token.Lookup(lit)
		return
	case '"':
		tok = token.STRING
		lit = s.scanString('"')
		return
	case '\'':
		tok = token.STRING
		lit = s.scanString('\'')
		return

	default:
		switch {
		case isLetter(ch):
			lit = s.scanIdentifier()
			tok = token.Lookup(lit)
			// no need to forward to nextChar, scanIdentifier did it
			return
		case isDecimal(ch):
			tok, lit = s.scanNumber()
			// no need to forward to nextChar, scanNumber did it
			return
		default:
			s.error(s.offset, fmt.Sprintf("illegal character %#U", ch))
			tok = token.ILLEGAL
			lit = string(ch)
		}
	}

	s.nextChar()
	return
}
