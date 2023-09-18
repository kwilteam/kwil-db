package types

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
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

var _ json.Marshaler = HexBytes{}
var _ json.Unmarshaler = (*HexBytes)(nil)
