//go:build actions_math || ext_test

package mathexample

import (
	"context"
	"fmt"
	"math/big"

	"github.com/kwilteam/kwil-db/core/types/extensions"
	"github.com/kwilteam/kwil-db/extensions/actions"
)

func init() {
	mathExt := &MathExtension{}
	err := actions.RegisterExtension("math", mathExt)
	if err != nil {
		panic(err)
	}
}

type MathExtension struct{}

func (e *MathExtension) Name() string {
	return "math"
}

// this initialize function checks if round is set.  If not, it sets it to "up"
func (e *MathExtension) Initialize(ctx context.Context, metadata map[string]string) (map[string]string, error) {
	_, ok := metadata["round"]
	if !ok {
		metadata["round"] = "up"
	}

	roundVal := metadata["round"]
	if roundVal != "up" && roundVal != "down" {
		return nil, fmt.Errorf("round must be either 'up' or 'down'. default is 'up'")
	}

	return metadata, nil
}

func (e *MathExtension) Execute(ctx extensions.CallContext, metadata map[string]string, method string, args ...any) ([]any, error) {
	switch method {
	case "add":
		return e.add(args...)
	case "subtract":
		return e.subtract(args...)
	case "multiply":
		return e.multiply(args...)
	case "divide":
		return e.divide(metadata, args...)
	default:
		return nil, fmt.Errorf("method %s not found", method)
	}
}

// add takes two integers and returns their sum
func (e *MathExtension) add(values ...any) ([]any, error) {
	if len(values) != 2 {
		return nil, fmt.Errorf("expected 2 values for method Add, got %d", len(values))
	}

	val0Int, ok := values[0].(int)
	if !ok {
		return nil, fmt.Errorf("argument 1 is not an int")
	}

	val1Int, ok := values[1].(int)
	if !ok {
		return nil, fmt.Errorf("argument 2 is not an int")
	}

	var results []any
	results = append(results, val0Int+val1Int)
	return results, nil
}

// subtract takes two integers and returns their difference
func (e *MathExtension) subtract(values ...any) ([]any, error) {
	if len(values) != 2 {
		return nil, fmt.Errorf("expected 2 values for method Add, got %d", len(values))
	}

	val0Int, ok := values[0].(int)
	if !ok {
		return nil, fmt.Errorf("argument 1 is not an int")
	}

	val1Int, ok := values[1].(int)
	if !ok {
		return nil, fmt.Errorf("argument 2 is not an int")
	}

	var results []any
	results = append(results, val0Int-val1Int)
	return results, nil
}

// multiply takes two integers and returns their product
func (e *MathExtension) multiply(values ...any) ([]any, error) {
	if len(values) != 2 {
		return nil, fmt.Errorf("expected 2 values for method Add, got %d", len(values))
	}

	val0Int, ok := values[0].(int)
	if !ok {
		return nil, fmt.Errorf("argument 1 is not an int")
	}

	val1Int, ok := values[1].(int)
	if !ok {
		return nil, fmt.Errorf("argument 2 is not an int")
	}

	var results []any
	results = append(results, val0Int*val1Int)
	return results, nil
}

// divide takes two integers and returns their quotient rounded up or down depending on how the extension was initialized
func (e *MathExtension) divide(metadata map[string]string, values ...any) ([]any, error) {
	if len(values) != 2 {
		return nil, fmt.Errorf("expected 2 values for method Divide, got %d", len(values))
	}

	val0Int, ok := values[0].(int)
	if !ok {
		return nil, fmt.Errorf("argument 1 is not an int")
	}

	val1Int, ok := values[1].(int)
	if !ok {
		return nil, fmt.Errorf("argument 2 is not an int")
	}

	bigVal1 := newBigFloat(float64(val0Int))

	bigVal2 := newBigFloat(float64(val1Int))

	result := new(big.Float).Quo(bigVal1, bigVal2)

	var IntResult *big.Int
	var results []any
	if metadata["round"] == "up" {
		IntResult = roundUp(result)
	} else {
		IntResult = roundDown(result)
	}
	results = append(results, IntResult)
	return results, nil
}

// roundUp takes a big.Float and returns a new big.Float rounded up.
func roundUp(f *big.Float) *big.Int {
	c := new(big.Float).SetPrec(precision).Copy(f)
	r := new(big.Int)
	f.Int(r)

	if c.Sub(c, new(big.Float).SetPrec(precision).SetInt(r)).Sign() > 0 {
		r.Add(r, big.NewInt(1))
	}

	return r
}

// roundDown takes a big.Float and returns a new big.Float rounded down.
func roundDown(f *big.Float) *big.Int {
	r := new(big.Int)
	f.Int(r)

	return r
}

const (
	precision = 128
)

func newBigFloat(num float64) *big.Float {
	bg := new(big.Float).SetPrec(precision)

	return bg.SetFloat64(num)
}
