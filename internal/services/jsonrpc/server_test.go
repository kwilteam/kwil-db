package rpcserver

import "testing"

func ptrTo[T any](x T) *T {
	return &x
}

func Test_zeroID(t *testing.T) {
	var i any = (*int)(nil) // i != nil, it's a non-nil interface with nil data
	tests := []struct {
		name string
		id   any
		want bool
	}{
		{"int 0", int(0), true},
		{"int64 0", int64(0), true},
		{"float64 0", float64(0), true},
		{"ptr to int 0", ptrTo(0), true},
		{"nil ptr", (*int)(nil), true},
		{"non-interface to nil", i, true},
		{"nil", nil, true},
		{"empty string", "", true},
		{"int 1`", int(1), false},
		{"float64 1.1", float64(1.1), false},
		{"ptr to int 1", ptrTo(1), false},
		{"string a", "a", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := zeroID(tt.id); got != tt.want {
				t.Errorf("zeroID() = %v, want %v", got, tt.want)
			}
		})
	}
}
