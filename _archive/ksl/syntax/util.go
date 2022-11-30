package syntax

import (
	"fmt"
	"strconv"
	"unicode/utf8"

	"ksl"
	"ksl/syntax/lex"

	"github.com/apparentlymart/go-textseg/v13/textseg"
)

// ParseStringLiteralToken processes the given token, which must be either a
// TokenQuotedLit or a TokenStringLit, returning the string resulting from
// resolving any escape sequences.
func ParseStringLiteralToken(tok lex.Token) (string, ksl.Diagnostics) {
	var quoted bool
	switch tok.Type {
	case lex.TokenQuotedLit:
		quoted = true
	case lex.TokenStringLit:
		quoted = false
	default:
		panic("ParseStringLiteralToken can only be used with TokenStringLit and TokenQuotedLit tokens")
	}
	var diags ksl.Diagnostics

	ret := make([]byte, 0, len(tok.Value))
	slices := lex.ScanStringLit([]byte(tok.Value), quoted)

	// We will mutate rng constantly as we walk through our token slices below.
	// Any diagnostics must take a copy of this rng rather than simply pointing
	// to it, e.g. by using rng.Ptr() rather than &rng.
	rng := tok.Range
	rng.End = rng.Start

Slices:
	for _, slice := range slices {
		if len(slice) == 0 {
			continue
		}

		// Advance the start of our range to where the previous token ended
		rng.Start = rng.End

		// Advance the end of our range to after our token.
		b := slice
		for len(b) > 0 {
			adv, ch, _ := textseg.ScanGraphemeClusters(b, true)
			rng.End.Offset += adv
			switch ch[0] {
			case '\r', '\n':
				rng.End.Line++
				rng.End.Column = 1
			default:
				rng.End.Column++
			}
			b = b[adv:]
		}

	TokenType:
		switch slice[0] {
		case '\\':
			if !quoted {
				// If we're not in quoted mode then just treat this token as
				// normal. (Slices can still start with backslash even if we're
				// not specifically looking for backslash sequences.)
				break TokenType
			}
			if len(slice) < 2 {
				diags = append(diags, &ksl.Diagnostic{
					Severity: ksl.DiagError,
					Summary:  "Invalid escape sequence",
					Detail:   "Backslash must be followed by an escape sequence selector character.",
					Subject:  rng.Ptr(),
				})
				break TokenType
			}

			switch slice[1] {

			case 'n':
				ret = append(ret, '\n')
				continue Slices
			case 'r':
				ret = append(ret, '\r')
				continue Slices
			case 't':
				ret = append(ret, '\t')
				continue Slices
			case '"':
				ret = append(ret, '"')
				continue Slices
			case '\\':
				ret = append(ret, '\\')
				continue Slices
			case 'u', 'U':
				if slice[1] == 'u' && len(slice) != 6 {
					diags = append(diags, &ksl.Diagnostic{
						Severity: ksl.DiagError,
						Summary:  "Invalid escape sequence",
						Detail:   "The \\u escape sequence must be followed by four hexadecimal digits.",
						Subject:  rng.Ptr(),
					})
					break TokenType
				} else if slice[1] == 'U' && len(slice) != 10 {
					diags = append(diags, &ksl.Diagnostic{
						Severity: ksl.DiagError,
						Summary:  "Invalid escape sequence",
						Detail:   "The \\U escape sequence must be followed by eight hexadecimal digits.",
						Subject:  rng.Ptr(),
					})
					break TokenType
				}

				numHex := string(slice[2:])
				num, err := strconv.ParseUint(numHex, 16, 32)
				if err != nil {
					// Should never happen because the scanner won't match
					// a sequence of digits that isn't valid.
					panic(err)
				}

				r := rune(num)
				l := utf8.RuneLen(r)
				if l == -1 {
					diags = append(diags, &ksl.Diagnostic{
						Severity: ksl.DiagError,
						Summary:  "Invalid escape sequence",
						Detail:   fmt.Sprintf("Cannot encode character U+%04x in UTF-8.", num),
						Subject:  rng.Ptr(),
					})
					break TokenType
				}
				for i := 0; i < l; i++ {
					ret = append(ret, 0)
				}
				rb := ret[len(ret)-l:]
				utf8.EncodeRune(rb, r)

				continue Slices

			default:
				diags = append(diags, &ksl.Diagnostic{
					Severity: ksl.DiagError,
					Summary:  "Invalid escape sequence",
					Detail:   fmt.Sprintf("The symbol %q is not a valid escape sequence selector.", slice[1:]),
					Subject:  rng.Ptr(),
				})
				ret = append(ret, slice[1:]...)
				continue Slices
			}
		}

		// If we fall out here or break out of here from the switch above
		// then this slice is just a literal.
		ret = append(ret, slice...)
	}

	return string(ret), diags
}
