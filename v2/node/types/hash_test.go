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
