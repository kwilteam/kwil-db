package auth

import (
	"testing"
)

func Test_eip55ChecksumAddr(t *testing.T) {
	tests := []struct {
		name     string
		input    [20]byte
		expected string
	}{
		{
			name:     "Basic address",
			input:    [20]byte{0x5a, 0xAA, 0xfE, 0x6F, 0x8E, 0x4E, 0x44, 0xAA, 0x5d, 0x4c, 0xBd, 0x08, 0x7A, 0x63, 0x9B, 0x5E, 0x8A, 0x3E, 0xd3, 0x95},
			expected: "0x5aaaFe6F8e4E44aa5D4cBd087a639b5e8a3Ed395",
		},
		{
			name:     "All zeros",
			input:    [20]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			expected: "0x0000000000000000000000000000000000000000",
		},
		{
			name:     "Mixed case address",
			input:    [20]byte{0x00, 0x12, 0x34, 0x56, 0x78, 0x9a, 0xbc, 0xde, 0xf0, 0x12, 0x34, 0x56, 0x78, 0x9a, 0xbc, 0xde, 0xf0, 0x12, 0x34, 0x56},
			expected: "0x00123456789AbcdeF0123456789abCdef0123456",
		},
		{
			name:     "All F's",
			input:    [20]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			expected: "0xFFfFfFffFFfffFFfFFfFFFFFffFFFffffFfFFFfF",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := eip55ChecksumAddr(tt.input)
			if result != tt.expected {
				t.Errorf("checksumHex() = %v, want %v", result, tt.expected)
			}
		})
	}
}
