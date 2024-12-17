package types

import (
	"encoding/json"
	"testing"
)

func TestHashJSONMarshaling(t *testing.T) {
	t.Run("marshal valid hash", func(t *testing.T) {
		h := Hash{0x1, 0x2, 0x3, 0x4}
		data, err := json.Marshal(h)
		if err != nil {
			t.Fatal(err)
		}
		expected := `"0102030400000000000000000000000000000000000000000000000000000000"`
		if string(data) != expected {
			t.Errorf("got %s, want %s", string(data), expected)
		}
	})

	t.Run("unmarshal valid hash", func(t *testing.T) {
		input := `"0102030400000000000000000000000000000000000000000000000000000000"`
		var h Hash
		err := json.Unmarshal([]byte(input), &h)
		if err != nil {
			t.Fatal(err)
		}
		expected := Hash{0x1, 0x2, 0x3, 0x4}
		if h != expected {
			t.Errorf("got %v, want %v", h, expected)
		}
	})

	t.Run("unmarshal invalid json", func(t *testing.T) {
		input := `invalid`
		var h Hash
		err := json.Unmarshal([]byte(input), &h)
		if err == nil {
			t.Error("expected error for invalid JSON")
		}
	})

	t.Run("unmarshal invalid hex string", func(t *testing.T) {
		input := `"xyz"`
		var h Hash
		err := json.Unmarshal([]byte(input), &h)
		if err == nil {
			t.Error("expected error for invalid hex string")
		}
	})

	t.Run("unmarshal wrong length", func(t *testing.T) {
		input := `"0102"`
		var h Hash
		err := json.Unmarshal([]byte(input), &h)
		if err == nil {
			t.Error("expected error for wrong length hex string")
		}
	})

	t.Run("marshal zero hash", func(t *testing.T) {
		h := Hash{}
		data, err := json.Marshal(h)
		if err != nil {
			t.Fatal(err)
		}
		expected := `"0000000000000000000000000000000000000000000000000000000000000000"`
		if string(data) != expected {
			t.Errorf("got %s, want %s", string(data), expected)
		}
	})
}

func TestNewHashFromBytes(t *testing.T) {
	t.Run("valid byte slice", func(t *testing.T) {
		input := []byte{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0xb, 0xc, 0xd, 0xe, 0xf, 0x10,
			0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f, 0x20}
		h, err := NewHashFromBytes(input)
		if err != nil {
			t.Fatal(err)
		}
		for i := range input {
			if h[i] != input[i] {
				t.Errorf("byte at index %d: got %x, want %x", i, h[i], input[i])
			}
		}
	})

	t.Run("empty byte slice", func(t *testing.T) {
		_, err := NewHashFromBytes([]byte{})
		if err == nil {
			t.Error("expected error for empty byte slice")
		}
	})

	t.Run("byte slice too short", func(t *testing.T) {
		_, err := NewHashFromBytes([]byte{0x1, 0x2, 0x3})
		if err == nil {
			t.Error("expected error for short byte slice")
		}
	})

	t.Run("byte slice too long", func(t *testing.T) {
		input := make([]byte, 33)
		_, err := NewHashFromBytes(input)
		if err == nil {
			t.Error("expected error for long byte slice")
		}
	})
}

func TestHashIsZero(t *testing.T) {
	t.Run("zero hash", func(t *testing.T) {
		h := Hash{}
		if !h.IsZero() {
			t.Error("expected zero hash to be zero")
		}
	})

	t.Run("non-zero hash", func(t *testing.T) {
		h := Hash{0x1}
		if h.IsZero() {
			t.Error("expected non-zero hash to not be zero")
		}
	})

	t.Run("compare with ZeroHash", func(t *testing.T) {
		if !ZeroHash.IsZero() {
			t.Error("expected ZeroHash to be zero")
		}
	})
}
