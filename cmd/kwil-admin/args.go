package main

import (
	"encoding"
	"encoding/hex"
)

// These types are defined to provide custom parsing logic to go-arg, which
// recognizes and uses the encoding.TextUnmarshaler interface to process
// arguments and flags.

// HexArg is used to decode hexadecimal text into a byte slice.
type HexArg []byte

// UnmarshalText satisfies the encoding.TextUnmarshaler interface.
func (ka *HexArg) UnmarshalText(b []byte) error {
	key, err := hex.DecodeString(string(b))
	if err != nil {
		return err
	}
	*ka = key
	return nil
}

var _ encoding.TextUnmarshaler = (*HexArg)(nil)
