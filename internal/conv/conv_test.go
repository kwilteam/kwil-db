package conv_test

import (
	"testing"

	"github.com/kwilteam/kwil-db/internal/conv"
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
