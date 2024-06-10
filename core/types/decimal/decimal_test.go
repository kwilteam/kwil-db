package decimal_test

import (
	"math/big"
	"testing"

	"github.com/kwilteam/kwil-db/core/types/decimal"
	"github.com/stretchr/testify/require"
	"gotest.tools/assert"
)

func Test_NewParsedDecimal(t *testing.T) {
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
		{
			name:    "<1",
			decimal: "0.000123",
			prec:    6,
			scale:   6,
			want:    "0.000123",
		},
	}

	// test cases for decimal creation
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, err := decimal.NewExplicit(tt.decimal, tt.prec, tt.scale)
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

func Test_DecimalParsing(t *testing.T) {
	type testcase struct {
		name  string
		in    string
		prec  uint16
		scale uint16
		err   bool
	}

	tests := []testcase{
		{
			name:  "basic",
			in:    "123.456",
			prec:  6,
			scale: 3,
		},
		{
			name:  "no decimal",
			in:    "1",
			prec:  1,
			scale: 0,
		},
		{
			name:  "no int",
			in:    "0.456",
			prec:  3,
			scale: 3,
		},
		{
			name: "no decimal or int",
			in:   "",
			err:  true,
		},
		{
			name:  "negative",
			in:    "-123.456",
			prec:  6,
			scale: 3,
		},
		{
			name:  "positive",
			in:    "+123.456",
			prec:  6,
			scale: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, err := decimal.NewFromString(tt.in)
			if tt.err {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			require.Equal(t, tt.prec, d.Precision())
			require.Equal(t, tt.scale, d.Scale())
		})

	}
}

func Test_MulDecimal(t *testing.T) {
	// happy path
	a := "123.456"
	b := "2.000"

	decA, err := decimal.NewFromString(a)
	require.NoError(t, err)

	decB, err := decimal.NewFromString(b)
	require.NoError(t, err)

	decMul, err := decA.Mul(decA, decB)
	require.NoError(t, err)

	assert.Equal(t, "246.912", decMul.String())

	// overflow
	decA, err = decimal.NewFromString("123.456")
	require.NoError(t, err)

	decB, err = decimal.NewFromString("10.000")
	require.NoError(t, err)

	_, err = decA.Mul(decA, decB)
	require.Error(t, err)

	// handle the overflow error
	decA, err = decimal.NewFromString("123.456")
	require.NoError(t, err)

	decB, err = decimal.NewFromString("10.000")
	require.NoError(t, err)

	res := decimal.Decimal{}
	err = res.SetPrecisionAndScale(6, 2)
	require.NoError(t, err)

	_, err = res.Mul(decA, decB)
	require.NoError(t, err)

	require.Equal(t, "1234.56", res.String())
}

func Test_DecimalMath(t *testing.T) {
	type testcase struct {
		name string
		a    string
		b    string
		add  string
		sub  string
		div  string
		mod  string
	}

	tests := []testcase{
		{
			name: "basic",
			a:    "111.111",
			b:    "222.222",
			add:  "333.333",
			sub:  "-111.111",
			div:  "0.500",
			mod:  "111.111",
		},
		{
			name: "negative",
			a:    "-111.111",
			b:    "222.222",
			add:  "111.111",
			sub:  "-333.333",
			div:  "-0.500",
			mod:  "-111.111",
		},
		{
			name: "different scale",
			a:    "111.111",
			b:    "222.222222",
			add:  "333.333222",
			sub:  "-111.111222",
			div:  "0.500000",
			mod:  "111.111000",
		},
		{
			name: "different precision",
			a:    "1.111",
			b:    "222.222",
			add:  "223.333",
			sub:  "-221.111",
			div:  "0.005",
			mod:  "1.111",
		},
		{
			name: "different precision and scale",
			a:    "11.11",
			b:    "2.2222",
			add:  "13.3322",
			sub:  "8.8878",
			div:  "4.9995",
			mod:  "2.2212",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var a *decimal.Decimal
			var b *decimal.Decimal
			// greatestScale is the greatest scale of the two decimals
			var greatestScale uint16

			// reset resets the a and b variables,
			// since their pointers get shared between tests.
			reset := func() {
				var err error
				a, err = decimal.NewFromString(tt.a)
				require.NoError(t, err)

				b, err = decimal.NewFromString(tt.b)
				require.NoError(t, err)

				if a.Scale() > b.Scale() {
					greatestScale = a.Scale()
				} else {
					greatestScale = b.Scale()
				}
			}
			reset()

			add, err := decimal.Add(a, b)
			require.NoError(t, err)
			eq(t, add, tt.add, greatestScale)

			reset()

			sub, err := decimal.Sub(a, b)
			require.NoError(t, err)
			eq(t, sub, tt.sub, greatestScale)

			reset()

			// we dont test mul here since it would likely overflow

			div, err := decimal.Div(a, b)
			require.NoError(t, err)
			d := div.String()
			_ = d
			eq(t, div, tt.div, greatestScale)

			reset()

			mod, err := decimal.Mod(a, b)
			require.NoError(t, err)
			eq(t, mod, tt.mod, greatestScale)
		})
	}
}

// eq checks that a decimal is equal to a string.
// It will round the decimal to the given scale.
func eq(t *testing.T, dec *decimal.Decimal, want string, round uint16) {
	dec2, err := decimal.NewFromString(want)
	require.NoError(t, err)

	old := dec.String()

	err = dec.Round(round)
	require.NoError(t, err)

	err = dec2.Round(round)
	require.NoError(t, err)

	// since dec will get overwritten by Cmp
	got := dec.String()

	cmp, err := dec.Cmp(dec2)
	require.NoError(t, err)

	require.Equalf(t, 0, cmp, "want: %s, got: %s, rounded from: %s", dec2.String(), got, old)
}

func Test_AdjustPrecAndScale(t *testing.T) {
	a, err := decimal.NewFromString("111.111")
	require.NoError(t, err)

	err = a.SetPrecisionAndScale(9, 6)
	require.NoError(t, err)

	require.Equal(t, "111.111000", a.String())

	// set prec/scale back
	err = a.SetPrecisionAndScale(6, 3)
	require.NoError(t, err)

	require.Equal(t, "111.111", a.String())

	// set prec/scale too low
	err = a.SetPrecisionAndScale(3, 2)
	require.Error(t, err)
}

func Test_AdjustScaleMath(t *testing.T) {
	a, err := decimal.NewFromString("111.111")
	require.NoError(t, err)

	err = a.SetPrecisionAndScale(6, 3)
	require.NoError(t, err)

	b, err := decimal.NewFromString("222.22")
	require.NoError(t, err)

	_, err = a.Add(a, b)
	require.NoError(t, err)

	require.Equal(t, "333.331", a.String())

	// set prec/scale back
	err = a.SetPrecisionAndScale(6, 2)
	require.NoError(t, err)

	require.Equal(t, "333.33", a.String())

	c, err := decimal.NewFromString("30.22")
	require.NoError(t, err)

	_, err = a.Sub(a, c)
	require.NoError(t, err)

	require.Equal(t, "303.11", a.String())
}

func Test_RemoveScale(t *testing.T) {
	a, err := decimal.NewFromString("111.111")
	require.NoError(t, err)

	err = a.SetPrecisionAndScale(6, 2)
	require.NoError(t, err)

	require.Equal(t, "111.11", a.String())

	err = a.SetPrecisionAndScale(6, 3)
	require.NoError(t, err)

	require.Equal(t, "111.110", a.String())
}

func Test_DecimalCmp(t *testing.T) {
	type testcase struct {
		name string
		a    string
		b    string
		want int
	}

	tests := []testcase{
		{
			name: "equal",
			a:    "123.456",
			b:    "123.456",
			want: 0,
		},
		{
			name: "equal values, different scale",
			a:    "0123.456",
			b:    "123.456000",
			want: 0,
		},
		{
			name: "different values, different scale",
			a:    "123.456001",
			b:    "123.456",
			want: 1,
		},
		{
			name: "different values, different precision",
			a:    "123.456",
			b:    "1123.456",
			want: -1,
		},
		{
			name: "negative",
			a:    "-123.456",
			b:    "123.456",
			want: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, err := decimal.NewFromString(tt.a)
			require.NoError(t, err)

			b, err := decimal.NewFromString(tt.b)
			require.NoError(t, err)

			cmp, err := a.Cmp(b)
			require.NoError(t, err)

			require.Equal(t, tt.want, cmp)
		})
	}
}

// Testing setting a decimal from a big int and an exponent
func Test_BigAndExp(t *testing.T) {
	type testcase struct {
		name     string
		big      string // will be converted to a big.Int
		exp      int32
		out      string
		outPrec  uint16
		outScale uint16
		wantErr  bool
	}

	tests := []testcase{
		{
			name:     "basic",
			big:      "123456",
			exp:      -3,
			out:      "123.456",
			outPrec:  6,
			outScale: 3,
		},
		{
			name:     "negative",
			big:      "-123456",
			exp:      -2,
			out:      "-1234.56",
			outPrec:  6,
			outScale: 2,
		},
		{
			name:     "0 exponent",
			big:      "123456",
			exp:      0,
			out:      "123456",
			outPrec:  6,
			outScale: 0,
		},
		{
			name:    "positive exp",
			big:     "123",
			exp:     4,
			wantErr: true,
		},
		{
			name:     "exp less than precision properly adjusts precision",
			big:      "123",
			exp:      -4,
			out:      "0.0123",
			outPrec:  4,
			outScale: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bigInt, ok := new(big.Int).SetString(tt.big, 10)
			require.True(t, ok)

			d, err := decimal.NewFromBigInt(bigInt, tt.exp)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			require.Equal(t, tt.out, d.String())
			require.Equal(t, tt.outPrec, d.Precision())
			require.Equal(t, tt.outScale, d.Scale())
		})
	}
}
