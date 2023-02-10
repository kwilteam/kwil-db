package big

import (
	"fmt"
	"math/big"
)

var (
	Big0 = big.NewInt(0)
)

// contains some utility functions for big.Int that I use in various places

type bigFunctionPicker interface {
	Add(bigint string) (*big.Int, error)
	Sub(bigint string) (*big.Int, error)
	Compare(bigint string) (int, error)
	AsBigInt() (*big.Int, error)
}

func BigStr(str string) bigFunctionPicker {
	return &bigStr{str}
}

func Big(i int64) bigFunctionPicker {
	return &bigStr{fmt.Sprintf("%d", i)}
}

type bigStr struct {
	str string
}

func (b *bigStr) Add(bigint string) (*big.Int, error) {
	aa, ok := new(big.Int).SetString(b.str, 10)
	if !ok {
		return nil, fmt.Errorf("could not convert %s to big.Int", b.str)
	}
	bb, ok := new(big.Int).SetString(bigint, 10)
	if !ok {
		return nil, fmt.Errorf("could not convert %s to big.Int", bigint)
	}
	return aa.Add(aa, bb), nil
}

func (b *bigStr) Sub(bigint string) (*big.Int, error) {
	aa, ok := new(big.Int).SetString(b.str, 10)
	if !ok {
		return nil, fmt.Errorf("could not convert %s to big.Int", b.str)
	}
	bb, ok := new(big.Int).SetString(bigint, 10)
	if !ok {
		return nil, fmt.Errorf("could not convert %s to big.Int", bigint)
	}
	return aa.Sub(aa, bb), nil
}

func (b *bigStr) Compare(bigint string) (int, error) {
	// convert to big.Int
	ai, ok := new(big.Int).SetString(b.str, 10)
	if !ok {
		return 0, fmt.Errorf("failed to convert %s to big.Int", b.str)
	}
	bi, ok := new(big.Int).SetString(bigint, 10)
	if !ok {
		return 0, fmt.Errorf("failed to convert %s to big.Int", bigint)
	}

	// compare
	return ai.Cmp(bi), nil
}

func (b *bigStr) AsBigInt() (*big.Int, error) {
	ai, ok := new(big.Int).SetString(b.str, 10)
	if !ok {
		return nil, fmt.Errorf("failed to convert %s to big.Int", b.str)
	}
	return ai, nil
}
