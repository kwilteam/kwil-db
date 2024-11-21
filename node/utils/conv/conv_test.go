package conv_test

import (
	"testing"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/types/decimal"
	"github.com/kwilteam/kwil-db/node/utils/conv"
	"github.com/stretchr/testify/require"
)

func TestInt(t *testing.T) {
	tests := []struct {
		name    string
		arg     any
		want    int64
		wantErr bool
	}{
		{
			name: "int",
			arg:  1,
			want: 1,
		},
		{
			name: "int8",
			arg:  int8(1),
			want: 1,
		},
		{
			name: "int16",
			arg:  int16(1),
			want: 1,
		},
		{
			name: "string",
			arg:  "1",
			want: 1,
		},
		{
			name:    "string (invalid)",
			arg:     "hello",
			wantErr: true,
		},
		{
			name: "bool (true)",
			arg:  true,
			want: 1,
		},
		{
			name: "bool (false)",
			arg:  false,
			want: 0,
		},
		{
			name:    "struct",
			arg:     struct{}{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := conv.Int(tt.arg)
			if (err != nil) != tt.wantErr {
				t.Errorf("Int() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Int() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestString(t *testing.T) {
	tests := []struct {
		name    string
		arg     any
		want    string
		wantErr bool
	}{
		{
			name: "string",
			arg:  "hello",
			want: "hello",
		},
		{
			name: "int",
			arg:  1,
			want: "1",
		},
		{
			name: "int8",
			arg:  int8(1),
			want: "1",
		},
		{
			name:    "struct",
			arg:     struct{}{},
			wantErr: true,
		},
		{
			name: "bool",
			arg:  true,
			want: "true",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := conv.String(tt.arg)
			if (err != nil) != tt.wantErr {
				t.Errorf("String() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_Blob(t *testing.T) {
	tests := []struct {
		name    string
		arg     any
		want    []byte
		wantErr bool
	}{
		{
			name: "string",
			arg:  "hello",
			want: []byte("hello"),
		},
		{
			name: "[]byte",
			arg:  []byte("hello"),
			want: []byte("hello"),
		},
		{
			name:    "struct",
			arg:     struct{}{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := conv.Blob(tt.arg)
			if (err != nil) != tt.wantErr {
				t.Errorf("Blob() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if string(got) != string(tt.want) {
				t.Errorf("Blob() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_Bool(t *testing.T) {
	tests := []struct {
		name    string
		arg     any
		want    bool
		wantErr bool
	}{
		{
			name: "bool (true)",
			arg:  true,
			want: true,
		},
		{
			name: "bool (false)",
			arg:  false,
			want: false,
		},
		{
			name: "int (1)",
			arg:  1,
			want: true,
		},
		{
			name: "int (0)",
			arg:  0,
			want: false,
		},
		{
			name:    "string",
			arg:     "hello",
			wantErr: true,
		},
		{
			name:    "struct",
			arg:     struct{}{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := conv.Bool(tt.arg)
			if (err != nil) != tt.wantErr {
				t.Errorf("Bool() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Bool() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_UUID(t *testing.T) {
	tests := []struct {
		name    string
		arg     any
		want    *types.UUID
		wantErr bool
	}{
		{
			name: "string",
			arg:  "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
			want: mustParseUUID("6ba7b810-9dad-11d1-80b4-00c04fd430c8"),
		},
		{
			name: "[]byte",
			arg:  []byte("6ba7b810-9dad-11d1-80b4-00c04fd430c8"),
			want: mustParseUUID("6ba7b810-9dad-11d1-80b4-00c04fd430c8"),
		},
		{
			name: "uuid",
			arg:  mustParseUUID("6ba7b810-9dad-11d1-80b4-00c04fd430c8"),
			want: mustParseUUID("6ba7b810-9dad-11d1-80b4-00c04fd430c8"),
		},
		{
			name: "bytes",
			arg:  []byte{0x6b, 0xa7, 0xb8, 0x10, 0x9d, 0xad, 0x11, 0xd1, 0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8},
			want: mustParseUUID("6ba7b810-9dad-11d1-80b4-00c04fd430c8"),
		},
		{
			name:    "bool",
			arg:     true,
			wantErr: true,
		},
		{
			name:    "struct",
			arg:     struct{}{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := conv.UUID(tt.arg)
			hasErr := err != nil
			if hasErr != tt.wantErr {
				t.Errorf("UUID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if hasErr {
				return
			}

			if got.String() != tt.want.String() {
				t.Errorf("UUID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_Uint256(t *testing.T) {
	tests := []struct {
		name    string
		arg     any
		want    *types.Uint256
		wantErr bool
	}{
		{
			name:    "string - invalid",
			arg:     "6ba7b8109dad11d180b400c04fd430c8",
			wantErr: true,
		},
		{
			name: "string",
			arg:  "58292472827384374328378382394367238126421",
			want: mustParseUint256("58292472827384374328378382394367238126421"),
		},
		{
			name: "uint256",
			arg:  mustParseUint256("58292472827384374328378382394367238126421"),
			want: mustParseUint256("58292472827384374328378382394367238126421"),
		},
		{
			name: "bytes",
			arg:  []byte{0x00, 0x01},
			want: mustParseUint256("1"),
		},
		{
			name:    "bool",
			arg:     true,
			wantErr: true,
		},
		{
			name:    "struct",
			arg:     struct{}{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := conv.Uint256(tt.arg)
			hasErr := err != nil
			if hasErr != tt.wantErr {
				t.Errorf("Uint256() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if hasErr {
				return
			}

			if got.String() != tt.want.String() {
				t.Errorf("Uint256() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_Decimal(t *testing.T) {
	tests := []struct {
		name    string
		arg     any
		want    *decimal.Decimal
		wantErr bool
	}{
		{
			name: "string",
			arg:  "1.234",
			want: mustDecimal("1.234"),
		},
		{
			name: "int",
			arg:  1234,
			want: mustDecimal("1234"),
		},
		{
			name: "int8",
			arg:  int8(123),
			want: mustDecimal("123"),
		},
		{
			name:    "struct",
			arg:     struct{}{},
			wantErr: true,
		},
		{
			name:    "bool",
			arg:     true,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := conv.Decimal(tt.arg)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			if got.String() != tt.want.String() {
				t.Errorf("Decimal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func mustDecimal(s string) *decimal.Decimal {
	d, err := decimal.NewFromString(s)
	if err != nil {
		panic(err)
	}
	return d

}

func mustParseUint256(s string) *types.Uint256 {
	u, err := types.Uint256FromString(s)
	if err != nil {
		panic(err)
	}
	return u
}

func mustParseUUID(s string) *types.UUID {
	u, err := types.ParseUUID(s)
	if err != nil {
		panic(err)
	}
	return u
}
