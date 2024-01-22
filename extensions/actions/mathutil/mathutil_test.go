package mathutil

import "testing"

func Test_Fraction(t *testing.T) {
	type testcase struct {
		name        string
		numerator   int64
		denominator int64
		number      int64
		want        int64
	}
	// function does (numerator/denominator) * number, and rounds down
	tests := []testcase{
		{
			name:        "1/2 * 2",
			numerator:   1,
			denominator: 2,
			number:      2,
			want:        1,
		},
		{
			name:        "1/2 * 1",
			numerator:   1,
			denominator: 2,
			number:      1,
			want:        0,
		},
		{
			name:        "104892/32034 * 6932", // arbitrarily big numbers 1
			numerator:   104892,
			denominator: 32034,
			number:      6932,
			want:        22698,
		},
		{
			name:        "13/3234454 * 15734567318", // arbitrarily big numbers 2
			numerator:   13,
			denominator: 3234454,
			number:      15734567318,
			want:        63240,
		},
		{
			name:        "largest int64s",
			numerator:   9223372036854775807,
			denominator: 9223372036854775807,
			number:      9223372036854775807,
			want:        9223372036854775807,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := fraction(tt.number, tt.numerator, tt.denominator)
			if err != nil {
				t.Errorf("fraction() error = %v", err)
				return
			}
			if got[0] != tt.want {
				t.Errorf("fraction() = %v, want %v", got, tt.want)
			}
		})
	}
}
