package decimal_test

import (
	"testing"

	"github.com/kwilteam/kwil-db/core/types/decimal"
	"github.com/stretchr/testify/require"
)

func Test_Decimal(t *testing.T) {

	type testcase struct {
		name    string
		decimal string
		prec    uint16
		scale   uint16
		want    string
		err     bool
	}

	tests := []testcase{
		{
			name:    "basic",
			decimal: "123.456",
			prec:    6,
			scale:   3,
			want:    "123.456",
		},
		{
			name:    "no scale",
			decimal: "1.456",
			prec:    1,
			scale:   0,
			want:    "1",
		},
		{
			name:    "overflow",
			decimal: "123.456",
			prec:    5,
			scale:   3,
			err:     true,
		},
		{
			name:    "rounding",
			decimal: "123.456",
			prec:    5,
			scale:   2,
			want:    "123.46",
		},
		{
			name:    "negative",
			decimal: "-123.456",
			prec:    6,
			scale:   3,
			want:    "-123.456",
		},
		{
			name:    "round down",
			decimal: "123.44",
			prec:    4,
			scale:   1,
			want:    "123.4",
		},
		{
			name:    "round up",
			decimal: "123.45",
			prec:    4,
			scale:   1,
			want:    "123.5",
		},
		{
			// while this is sort've unideal, it is expected, so keeping
			// it as a test case.
			name:    "second-digit round with enough precision",
			decimal: "123.449",
			prec:    5,
			scale:   1,
			want:    "123.5",
		},
		{
			name:    "second-digit round with not enough precision",
			decimal: "123.449",
			prec:    4,
			scale:   1,
			want:    "123.4",
		},
	}

	// test cases for decimal creation
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, err := decimal.New(tt.decimal, tt.prec, tt.scale)
			if tt.err {
				require.Errorf(t, err, "result: %v", d)
				return
			}
			if tt.err {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			require.Equal(t, tt.want, d.String())
		})
	}
}
