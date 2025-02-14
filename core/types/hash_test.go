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

	t.Run("unmarshal null to *Hash", func(t *testing.T) {
		h := new(Hash)
		input := `null`
		err := json.Unmarshal([]byte(input), &h)
		if err != nil {
			t.Fatal(err)
		}
		if h != nil {
			t.Errorf("got %v, want nil", h)
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

func TestHashTextMarshaling(t *testing.T) {
	t.Run("marshal text valid hash", func(t *testing.T) {
		h := Hash{0xff, 0xee, 0xdd, 0xcc}
		data, err := h.MarshalText()
		if err != nil {
			t.Fatal(err)
		}
		expected := "ffeeddcc00000000000000000000000000000000000000000000000000000000"
		if string(data) != expected {
			t.Errorf("got %s, want %s", string(data), expected)
		}

		// now test that it can be unmarshaled
		var h2 Hash
		err = h2.UnmarshalText(data)
		if err != nil {
			t.Fatal(err)
		}
		if h != h2 {
			t.Errorf("got %v, want %v", h2, h)
		}
	})

	t.Run("unmarshal text valid hash", func(t *testing.T) {
		input := "ffeeddcc00000000000000000000000000000000000000000000000000000000"
		var h Hash
		err := h.UnmarshalText([]byte(input))
		if err != nil {
			t.Fatal(err)
		}
		expected := Hash{0xff, 0xee, 0xdd, 0xcc}
		if h != expected {
			t.Errorf("got %v, want %v", h, expected)
		}
	})

	t.Run("unmarshal text odd length", func(t *testing.T) {
		input := "ffeeddccc"
		var h Hash
		err := h.UnmarshalText([]byte(input))
		if err == nil {
			t.Error("expected error for odd length hex string")
		}
	})

	t.Run("unmarshal text invalid characters", func(t *testing.T) {
		input := "gghhiijj00000000000000000000000000000000000000000000000000000000"
		var h Hash
		err := h.UnmarshalText([]byte(input))
		if err == nil {
			t.Error("expected error for invalid hex characters")
		}
	})

	t.Run("marshal empty hash", func(t *testing.T) {
		var h Hash
		data, err := h.MarshalText()
		if err != nil {
			t.Fatal(err)
		}
		expected := "0000000000000000000000000000000000000000000000000000000000000000"
		if string(data) != expected {
			t.Errorf("got %s, want %s", string(data), expected)
		}
	})

	t.Run("unmarshal text with whitespace", func(t *testing.T) {
		input := "  ffeeddcc00000000000000000000000000000000000000000000000000000000  "
		var h Hash
		err := h.UnmarshalText([]byte(input))
		if err == nil {
			t.Error("expected error for input with whitespace")
		}
	})
}
