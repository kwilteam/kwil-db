package dbml

import (
	"bufio"
	"bytes"
	"io"
)

const eof = rune(0)

const (
	DefaultMode = iota
	BlockMode
)

type Scanner struct {
	r    *bufio.Reader
	ch   rune
	l    uint
	c    uint
	mode int
}

func NewScanner(r io.Reader) *Scanner {
	s := &Scanner{r: bufio.NewReader(r), l: 1, c: 0}
	s.next()
	return s
}

func (s *Scanner) BlockMode(fn func()) {
	mode := s.mode
	s.mode = BlockMode
	fn()
	s.mode = mode
}

func (s *Scanner) Read() (tok Token, lit string) {
	switch {
	case isLetter(s.ch):
		return s.scanIdent()

	case isDigit(s.ch):
		return s.scanNumber()

	case s.ch == ' ' || s.ch == '\t':
		var buf bytes.Buffer
		for s.ch == ' ' || s.ch == '\t' {
			buf.WriteRune(s.ch)
			s.next()
		}
		return WHITESPACE, buf.String()

	case s.ch == '\n':
		s.next()
		return NEWLINE, "\n"

	default:
		ch := s.ch
		lit := string(ch)
		s.next()

		switch ch {
		case eof:
			return EOF, ""
		case '-':
			return SUB, lit
		case '<':
			if s.ch == '>' {
				s.next()
				return LTGT, "<>"
			}
			return LT, lit
		case '>':
			return GT, lit
		case '(':
			return LPAREN, lit
		case '[':
			return LBRACK, lit
		case '{':
			if s.mode == BlockMode {
				lit, ok := s.scanTo('}', true)
				if !ok {
					return ILLEGAL, lit
				}
				return BLOCK, lit
			}
			return LBRACE, lit
		case ')':
			return RPAREN, lit
		case ']':
			return RBRACK, lit
		case '}':
			return RBRACE, lit
		case ';':
			return SEMICOLON, lit
		case ':':
			return COLON, lit
		case ',':
			return COMMA, lit
		case '.':
			return PERIOD, lit
		case '`':
			return s.scanExpression()
		case '\'', '"':
			return s.scanString(ch)
		case '/':
			if s.ch == '/' {
				return COMMENT, s.scanComment()
			}
			return ILLEGAL, string(ch)
		}
		return ILLEGAL, string(ch)
	}
}

func (s *Scanner) scanComment() string {
	var buf bytes.Buffer
	buf.WriteString("/")
	for s.ch != '\n' && s.ch != eof {
		buf.WriteRune(s.ch)
		s.next()
	}
	return buf.String()
}

func (s *Scanner) scanNumber() (Token, string) {
	var buf bytes.Buffer
	countDot := 0
	for isDigit(s.ch) || (s.ch == '.' && countDot < 2) {
		if s.ch == '.' {
			countDot++
		}
		buf.WriteRune(s.ch)
		s.next()
	}
	if countDot < 1 {
		return INT, buf.String()
	} else if countDot > 1 {
		return ILLEGAL, buf.String()
	}
	return FLOAT, buf.String()
}

func (s *Scanner) scanString(quo rune) (Token, string) {
	switch quo {
	case '"':
		lit, ok := s.scanTo(quo, false)
		if ok {
			return DSTRING, lit
		}
		return ILLEGAL, lit
	case '\'':
		if s.ch != '\'' {
			lit, ok := s.scanTo(quo, false)
			if ok {
				return STRING, lit
			}
			return ILLEGAL, lit
		}
		// Handle Triple quote string
		var buf bytes.Buffer
		s.next()
		if s.ch == '\'' { // triple quote string
			s.next()
			count := 0
			for count < 3 {
				switch s.ch {
				case '\'':
					count++
				case eof:
					return ILLEGAL, buf.String()
				}
				buf.WriteRune(s.ch)
				s.next()
			}
			return TSTRING, buf.String()[:buf.Len()-count]
		}
		return ILLEGAL, buf.String()
	default:
		return ILLEGAL, string(eof)
	}
}

func (s *Scanner) scanExpression() (Token, string) {
	lit, ok := s.scanTo('`', true)
	if ok {
		return EXPR, lit
	}
	return ILLEGAL, lit
}

func (s *Scanner) scanTo(stop rune, multiline bool) (string, bool) {
	var buf bytes.Buffer
	for {
		switch s.ch {
		case stop:
			s.next()
			return buf.String(), true
		case '\n':
			if !multiline {
				return buf.String(), false
			}
			buf.WriteRune(s.ch)
			s.next()
		case eof:
			return buf.String(), false
		default:
			buf.WriteRune(s.ch)
			s.next()
		}
	}
}

func (s *Scanner) scanIdent() (tok Token, lit string) {
	var buf bytes.Buffer
	for {
		buf.WriteRune(s.ch)
		s.next()
		if !isLetter(s.ch) && !isDigit(s.ch) && s.ch != '_' {
			break
		}
	}
	return Lookup(buf.String()), buf.String()
}

func (s *Scanner) next() {
	ch, _, err := s.r.ReadRune()
	if err != nil {
		s.ch = eof
		return
	}
	if ch == '\n' {
		s.l++
		s.c = 0
	}
	s.c++
	s.ch = ch
}

func (s *Scanner) LineInfo() (uint, uint) {
	return s.l, s.c
}
