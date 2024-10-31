package types

import (
	"bytes"
	"testing"
)

func TestHexBytes_MarshalText(t *testing.T) {
	tests := []struct {
		name    string
		hb      HexBytes
		want    []byte
		wantErr bool
	}{
		{
			name:    "empty bytes",
			hb:      HexBytes{},
			want:    []byte(""),
			wantErr: false,
		},
		{
			name:    "single byte",
			hb:      HexBytes{0xff},
			want:    []byte("ff"),
			wantErr: false,
		},
		{
			name:    "multiple bytes",
			hb:      HexBytes{0x12, 0x34, 0x56},
			want:    []byte("123456"),
			wantErr: false,
		},
		{
			name:    "zero byte",
			hb:      HexBytes{0x00},
			want:    []byte("00"),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.hb.MarshalText()
			if (err != nil) != tt.wantErr {
				t.Errorf("HexBytes.MarshalText() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !bytes.Equal(got, tt.want) {
				t.Errorf("HexBytes.MarshalText() = %v, want %v", string(got), string(tt.want))
			}
		})
	}
}

func TestHexBytes_UnmarshalText(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		want    HexBytes
		wantErr bool
	}{
		{
			name:    "empty string",
			input:   []byte(""),
			want:    HexBytes{},
			wantErr: false,
		},
		{
			name:    "valid hex string",
			input:   []byte("0102030405"),
			want:    HexBytes{0x01, 0x02, 0x03, 0x04, 0x05},
			wantErr: false,
		},
		{
			name:    "invalid hex string",
			input:   []byte("xyz"),
			want:    nil,
			wantErr: true,
		},
		{
			name:    "odd length hex string",
			input:   []byte("123"),
			want:    nil,
			wantErr: true,
		},
		{
			name:    "with spaces",
			input:   []byte("12 34"),
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var hb HexBytes
			err := hb.UnmarshalText(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("HexBytes.UnmarshalText() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !bytes.Equal(hb, tt.want) {
				t.Errorf("HexBytes.UnmarshalText() = %v, want %v", hb, tt.want)
			}
		})
	}
}

func TestHexBytes_JSON(t *testing.T) {
	tests := []struct {
		name    string
		hb      HexBytes
		json    []byte
		wantErr bool
	}{
		{
			name:    "nil bytes (ok)",
			hb:      nil,
			json:    []byte(`""`),
			wantErr: false,
		},
		{
			name:    "missing quotes",
			json:    []byte(`1234`),
			wantErr: true,
		},
		{
			name:    "missing end quote",
			json:    []byte(`"1234`),
			wantErr: true,
		},
		{
			name:    "missing start quote",
			json:    []byte(`1234"`),
			wantErr: true,
		},
		{
			name:    "invalid hex chars",
			json:    []byte(`"gh12"`),
			wantErr: true,
		},
		{
			name:    "large byte array (valid json)",
			hb:      HexBytes{0xff, 0xee, 0xdd, 0xcc, 0xbb, 0xaa, 0x99, 0x88},
			json:    []byte(`"ffeeddccbbaa9988"`),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal never errors, round trip in next t.Run below.

			// Test Unmarshal
			var decoded HexBytes
			err := decoded.UnmarshalJSON(tt.json)
			if (err != nil) != tt.wantErr {
				t.Errorf("HexBytes.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !bytes.Equal(decoded, tt.hb) {
				t.Errorf("HexBytes.UnmarshalJSON() = %v, want %v", decoded, tt.hb)
			}
		})
	}

	t.Run("roundtrip", func(t *testing.T) {
		original := HexBytes{0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef}
		encoded, err := original.MarshalJSON()
		if err != nil {
			t.Fatalf("MarshalJSON failed: %v", err)
		}

		var decoded HexBytes
		err = decoded.UnmarshalJSON(encoded)
		if err != nil {
			t.Fatalf("UnmarshalJSON failed: %v", err)
		}

		if !bytes.Equal(original, decoded) {
			t.Errorf("roundtrip failed: got %v, want %v", decoded, original)
		}
	})
}
