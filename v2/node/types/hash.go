package types

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

const (
	HashLen = 32
)

type Hash [HashLen]byte

func HashBytes(b []byte) Hash {
	return sha256.Sum256(b)
}

// String returns the hexadecimal representation of the hash (always 64 characters)
func (h Hash) String() string {
	return hex.EncodeToString(h[:])
}

var _ json.Marshaler = Hash{}
var _ json.Marshaler = (*Hash)(nil)

// MarshalJSON ensures the hash marshals to JSON as a hexadecimal string.
func (h Hash) MarshalJSON() ([]byte, error) {
	return json.Marshal(h.String()) // i.e. `"` + h.String() + `"`
}

var _ json.Unmarshaler = (*Hash)(nil)

// UnmarshalJSON unmarshals a hash from a hexadecimal JSON string.
func (h *Hash) UnmarshalJSON(data []byte) error {
	var hexStr string
	if err := json.Unmarshal(data, &hexStr); err != nil {
		return err
	}

	parsed, err := NewHashFromString(hexStr)
	if err != nil {
		return err
	}

	*h = parsed
	return nil
}

// NewHashFromString parses a hexadecimal string into a Hash.
func NewHashFromString(s string) (Hash, error) {
	var h Hash
	if len(s) != HashLen*2 {
		return h, fmt.Errorf("invalid hash length: expected %d, got %d", HashLen*2, len(s))
	}
	_, err := hex.Decode(h[:], []byte(s))
	return h, err
}

// NewHashFromBytes creates a Hash from a byte slice.
func NewHashFromBytes(b []byte) (Hash, error) {
	var h Hash
	if len(b) != HashLen {
		return h, fmt.Errorf("invalid byte slice length: expected %d, got %d", HashLen, len(b))
	}
	copy(h[:], b)
	return h, nil
}

var ZeroHash Hash

func (h Hash) IsZero() bool {
	return h == ZeroHash
}
