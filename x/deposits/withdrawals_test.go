package deposits

import "testing"

func Test_validateNonce(t *testing.T) {
	type args struct {
		n    string
		low  int64
		high int64
		l    uint8
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "valid nonce",
			args: args{
				n:    "1453052:cm+r3",
				low:  1453045,
				high: 1453145,
				l:    5,
			},
			want: true,
		},
		{
			name: "range too low",
			args: args{
				n:    "1453052:cm+r3",
				low:  1453045,
				high: 1453050,
				l:    5,
			},
			want: false,
		},
		{
			name: "range too high",
			args: args{
				n:    "1453052:cm+r3",
				low:  1453100,
				high: 1453200,
				l:    5,
			},
			want: false,
		},
		{
			name: "no delimeter",
			args: args{
				n:    "1453052cm+r3",
				low:  1453045,
				high: 1453145,
				l:    5,
			},
			want: false,
		},
		{
			name: "too many delimeters",
			args: args{
				n:    "1453052:cm+r3:cm+r3",
				low:  1453045,
				high: 1453145,
				l:    5,
			},
			want: false,
		},
		{
			name: "invalid block expiration",
			args: args{
				n:    "1453a52:cm+r3",
				low:  1453045,
				high: 1453145,
				l:    5,
			},
			want: false,
		},
		{
			name: "second half of nonce too long",
			args: args{
				n:    "1453052:cm+r3+",
				low:  1453045,
				high: 1453145,
				l:    5,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := validateNonce(tt.args.n, tt.args.low, tt.args.high, tt.args.l); got != tt.want {
				t.Errorf("validateNonce() = %v, want %v", got, tt.want)
			}
		})
	}
}
