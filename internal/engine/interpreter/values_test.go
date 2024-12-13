package interpreter

import (
	"testing"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/types/decimal"
	"github.com/stretchr/testify/require"
)

func Test_Arithmetic(t *testing.T) {
	type testcase struct {
		name     string
		a        any
		aIsNull  *types.DataType
		b        any
		operator ArithmeticOp
		want     any
		wantErr  error
	}

	tests := []testcase{
		{
			name:     "add int",
			a:        1,
			b:        2,
			operator: add,
			want:     int64(3),
		},
		{
			name:     "sub int",
			a:        1,
			b:        2,
			operator: sub,
			want:     int64(-1),
		},
		{
			name:     "div int",
			a:        10,
			b:        5,
			operator: div,
			want:     int64(2),
		},
		{
			name:     "mul int",
			a:        2,
			b:        2,
			operator: mul,
			want:     int64(4),
		},
		{
			name:     "mod int",
			a:        10,
			b:        3,
			operator: mod,
			want:     int64(1),
		},
		{
			name:     "concat string",
			a:        "hello",
			b:        "world",
			operator: concat,
			want:     "helloworld",
		},
		{
			name:     "add decimal",
			a:        mustDec("1.1"),
			b:        mustDec("2.2"),
			operator: add,
			want:     mustDec("3.3"),
		},
		{
			name:     "sub decimal",
			a:        mustDec("1.1"),
			b:        mustDec("2.2"),
			operator: sub,
			want:     mustDec("-1.1"),
		},
		{
			name:     "div decimal",
			a:        mustDec("10.2"),
			b:        mustDec("5.1"),
			operator: div,
			want:     mustDec("2"),
		},
		{
			name:     "mul decimal",
			a:        mustDec("2.2"),
			b:        mustDec("2.2"),
			operator: mul,
			want:     mustDec("4.84"),
		},
		{
			name:     "mod decimal",
			a:        mustDec("10.3"),
			b:        mustDec("3"),
			operator: mod,
			want:     mustDec("1.3"),
		},
		{
			name:    "add mixed",
			a:       1,
			b:       mustDec("2.2"),
			wantErr: ErrTypeMismatch,
		},
		{
			name:    "cannot add blob",
			a:       []byte("hello"),
			b:       []byte("world"),
			wantErr: ErrArithmetic,
		},
		{
			name:    "cannot add bool",
			a:       true,
			b:       false,
			wantErr: ErrArithmetic,
		},
		{
			name:    "cannot add uuid",
			a:       mustUUID("550e8400-e29b-41d4-a716-446655440000"),
			b:       mustUUID("550e8400-e29b-41d4-a716-446655440000"),
			wantErr: ErrArithmetic,
		},
		{
			name:    "adding nulls yields null",
			aIsNull: types.IntType,
			b:       5,
			want:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var a Value
			var err error
			if tt.aIsNull == nil {
				a, err = NewValue(tt.a)
				require.NoError(t, err)
			} else {
				a = newNull(tt.aIsNull)
			}

			b, err := NewValue(tt.b)
			require.NoError(t, err)

			res, err := a.(ScalarValue).Arithmetic(b.(ScalarValue), tt.operator)
			if tt.wantErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)

				// if the results is a decimal, we should manually compare it
				raw := res.RawValue()
				if rawDec, ok := raw.(*decimal.Decimal); ok {
					wantDec, ok := tt.want.(*decimal.Decimal)
					require.True(t, ok)

					rec, err := rawDec.Cmp(wantDec)
					require.NoError(t, err)

					if rec != 0 {
						t.Fatalf("expected %v, got %v", wantDec.String(), rawDec.String())
					}

					return
				}

				require.EqualValues(t, tt.want, raw)
			}
		})
	}
}

func mustDec(dec string) *decimal.Decimal {
	d, err := decimal.NewFromString(dec)
	if err != nil {
		panic(err)
	}
	return d
}

func mustUUID(s string) *types.UUID {
	u, err := types.ParseUUID(s)
	if err != nil {
		panic(err)
	}
	return u
}
