package types

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
)

// HexBytes is used to decode hexadecimal text into a byte slice.
type HexBytes []byte

func (hb HexBytes) String() string {
	return hex.EncodeToString(hb)
}

// UnmarshalText satisfies the json.Unmarshaler interface.
func (hb *HexBytes) UnmarshalJSON(b []byte) error {
	if len(b) < 2 || b[0] != '"' || b[len(b)-1] != '"' {
		return fmt.Errorf("invalid hex string: %s", b)
	}
	sub := b[1 : len(b)-1] // strip the quotes
	dec := make([]byte, hex.DecodedLen(len(sub)))
	_, err := hex.Decode(dec, sub)
	if err != nil {
		return err
	}
	*hb = dec
	return nil
}

// MarshalJSON satisfies the json.Marshaler interface.
func (hb HexBytes) MarshalJSON() ([]byte, error) {
	s := make([]byte, 2+hex.EncodedLen(len(hb)))
	s[0], s[len(s)-1] = '"', '"'
	hex.Encode(s[1:], hb)
	return s, nil
}

func (hb HexBytes) MarshalText() ([]byte, error) {
	return []byte(hb.String()), nil
}

func (hb *HexBytes) UnmarshalText(b []byte) error {
	dec := make([]byte, hex.DecodedLen(len(b)))
	_, err := hex.Decode(dec, b)
	if err != nil {
		return err
	}
	*hb = dec
	return nil
}

var _ json.Marshaler = HexBytes{}
var _ json.Unmarshaler = (*HexBytes)(nil)

func (hb *HexBytes) Equals(other HexBytes) bool {
	return bytes.Equal(*hb, other)
}

var _ fmt.Formatter = HexBytes{}

// Format writes either address of 0th element in a slice in base 16 notation,
// with leading 0x (%p), or casts HexBytes to bytes and writes as hexadecimal
// string to s.
func (hb HexBytes) Format(s fmt.State, verb rune) {
	switch verb {
	case 'p':
		s.Write([]byte(fmt.Sprintf("%p", hb)))
	case 'X':
		s.Write([]byte(strings.ToUpper(hb.String())))
	default:
		s.Write([]byte(hb.String()))
	}
}
