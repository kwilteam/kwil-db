package dbml

import (
	"strings"
	"testing"
)

func sc(str string) *Scanner {
	return NewScanner(strings.NewReader(str))
}

func TestScanForNumber(t *testing.T) {
	s := sc("123.456")
	if tok, lit := s.Read(); tok != FLOAT {
		t.Fatalf("token %s, should be %s, lit %s", tok, FLOAT, lit)
	}

	s = sc("123.456i")
	if tok, lit := s.Read(); tok != FLOAT {
		t.Fatalf("token %s, should be FLOAT, lit %s", tok, lit)
		if tok, lit := s.Read(); tok != IDENT {
			t.Fatalf("token %s, should be %s, lit %s", tok, IDENT, lit)
		}
	}

	s = sc("123")
	if tok, lit := s.Read(); tok != INT {
		t.Fatalf("token %s, should be %s, lit %s", tok, INT, lit)
	}

	s = sc("123.2.3")
	if tok, lit := s.Read(); tok != ILLEGAL {
		t.Fatalf("token %s, should be %s, lit %s", tok, ILLEGAL, lit)
	}
}
