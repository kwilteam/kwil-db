package accounts

import (
	"errors"
	"math/big"
	"testing"
)

func TestSpendErrorAs(t *testing.T) {
	// Only a pointer (*SpendError) satisfies the error interface.
	sep := &SpendError{
		Balance: big.NewInt(1),
		Nonce:   2,
	}

	err := error(sep)

	// Only valid syntax is with a pointer to a *SpendError.
	sep0 := new(SpendError)
	if !errors.As(err, &sep0) {
		t.Fatal("should have been a SpendError")
	}
	// Changing either the receiver or the second argument indirection results
	// in "second argument to errors.As must be a non-nil pointer to either a
	// type that implements error, or to any interface type" at compile time.
	// This avoids runtime gotchas if it were a value method receiver.
}

func TestSpendErrorUnwrap(t *testing.T) {
	sep := &SpendError{
		Err:     ErrInsufficientFunds,
		Balance: big.NewInt(1),
		Nonce:   2,
	}

	err := error(sep)

	sep0 := new(SpendError)
	if !errors.As(err, &sep0) {
		t.Fatal("should have been a SpendError")
	}

	if !errors.Is(err, ErrInsufficientFunds) {
		t.Fatal("should have been an ErrInsufficientFunds")
	}
}

func TestSpendError_Error(t *testing.T) {
	type fields struct {
		Err     error
		Balance *big.Int
		Nonce   int64
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			"ok",
			fields{
				Err:     errors.New("fail"),
				Balance: big.NewInt(1),
				Nonce:   2,
			},
			"fail: account balance 1, nonce 2",
		},
		{
			"ok nil bal",
			fields{
				Err:     errors.New("fail"),
				Balance: nil,
				Nonce:   2,
			},
			"fail: account balance <nil>, nonce 2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			se := NewSpendError(tt.fields.Err, tt.fields.Balance, tt.fields.Nonce)
			if got := se.Error(); got != tt.want {
				t.Errorf("SpendError.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}
