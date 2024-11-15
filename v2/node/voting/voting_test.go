package voting

import (
	"math"
	"testing"
)

func Test_intDivUpFraction(t *testing.T) {
	tests := []struct {
		val       int64
		numerator int64
		divisor   int64
		want      int64
	}{
		// General cases
		{10, 2, 3, 7},
		{100, 1, 2, 50},
		{5, 10, 2, 25}, // not a sensible ratio in our use case, but validate anyway

		// Edge cases
		{0, 1, 2, 0},   // val is 0
		{1, 0, 2, 0},   // numerator is 0, result should be 0
		{10, 1, 1, 10}, // divisor is 1, result should be the val

		// Boundary conditions
		{math.MaxInt64, 1, math.MaxInt64, 1}, // ensure no overflow

		// Cases to ensure rounding up
		{1, 1, 2, 1}, // 0.5 should round up to 1
		{3, 2, 5, 2}, // 1.2 should round up to 2
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			if got := intDivUpFraction(tt.val, tt.numerator, tt.divisor); got != tt.want {
				t.Errorf("intDivUpFraction() = %v, want %v", got, tt.want)
			}
		})
	}
}
