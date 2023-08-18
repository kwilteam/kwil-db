package validators

import "testing"

func Test_intDivUp(t *testing.T) {
	tests := []struct {
		val  int64
		div  int64
		want int64
	}{
		{0, 1, 0},
		{1, 1, 1},
		{1, 2, 1},
		{1, 3, 1},
		{1, 4, 1},
		{2, 1, 2},
		{2, 2, 1},
		{3, 2, 2},
		{3, 3, 1},
		{3, 1, 3},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			if got := intDivUp(tt.val, tt.div); got != tt.want {
				t.Errorf("intDivUp() = %v, want %v", got, tt.want)
			}
		})
	}
}
