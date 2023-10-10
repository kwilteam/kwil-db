//go:build actions_math || ext_test

package extensions

import (
	"context"
	"fmt"
	"math/big"
)

func init() {
	ext, err := NewMathExtension()
	if err != nil {
		panic(err)
	}

	err = RegisterExtension("math", ext)
	if err != nil {
		panic(err)
	}
}

type MathExtension struct{}

func NewMathExtension() (*Extension, error) {
	mathExt := &MathExtension{}
	methods := map[string]MethodFunc{
		"add":      mathExt.add,
		"subtract": mathExt.subtract,
		"multiply": mathExt.multiply,
		"divide":   mathExt.divide,
	}

	ext, err := Builder().Named("math").WithMethods(methods).WithInitializer(initialize).Build()
	if err != nil {
		return nil, err
	}
	return ext, nil
}

func (e *MathExtension) Name() string {
	return "math"
}

// this initialize function checks if round is set.  If not, it sets it to "up"
func initialize(ctx context.Context, metadata map[string]string) (map[string]string, error) {
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

func (e *MathExtension) add(ctx *ExecutionContext, values ...*ScalarValue) ([]*ScalarValue, error) {
	if len(values) != 2 {
		return nil, fmt.Errorf("expected 2 values for method Add, got %d", len(values))
	}

	val0Int, err := values[0].Int()
	if err != nil {
		return nil, fmt.Errorf("failed to convert value to int: %w. \nreceived value: %v", err, val0Int)
	}

	val1Int, err := values[1].Int()
	if err != nil {
		return nil, fmt.Errorf("failed to convert value to int: %w. \nreceived value: %v", err, val1Int)
	}

	return encodeScalarValues(val0Int + val1Int)
}

func (e *MathExtension) subtract(ctx *ExecutionContext, values ...*ScalarValue) ([]*ScalarValue, error) {
	if len(values) != 2 {
		return nil, fmt.Errorf("expected 2 values for method Subtract, got %d", len(values))
	}

	val0Int, err := values[0].Int()
	if err != nil {
		return nil, fmt.Errorf("failed to convert value to int: %w. \nreceived value: %v", err, val0Int)
	}

	val1Int, err := values[1].Int()
	if err != nil {
		return nil, fmt.Errorf("failed to convert value to int: %w. \nreceived value: %v", err, val1Int)
	}

	return encodeScalarValues(val0Int - val1Int)
}

func (e *MathExtension) multiply(ctx *ExecutionContext, values ...*ScalarValue) ([]*ScalarValue, error) {
	if len(values) != 2 {
		return nil, fmt.Errorf("expected 2 values for method Multiply, got %d", len(values))
	}

	val0Int, err := values[0].Int()
	if err != nil {
		return nil, fmt.Errorf("failed to convert value to int: %w. \nreceived value: %v", err, val0Int)
	}

	val1Int, err := values[1].Int()
	if err != nil {
		return nil, fmt.Errorf("failed to convert value to int: %w. \nreceived value: %v", err, val1Int)
	}

	return encodeScalarValues(val0Int * val1Int)
}

func (e *MathExtension) divide(ctx *ExecutionContext, values ...*ScalarValue) ([]*ScalarValue, error) {
	if len(values) != 2 {
		return nil, fmt.Errorf("expected 2 values for method Divide, got %d", len(values))
	}

	val0Int, err := values[0].Int()
	if err != nil {
		return nil, fmt.Errorf("failed to convert value to int: %w. \nreceived value: %v", err, val0Int)
	}

	val1Int, err := values[1].Int()
	if err != nil {
		return nil, fmt.Errorf("failed to convert value to int: %w. \nreceived value: %v", err, val1Int)
	}

	bigVal1 := newBigFloat(float64(val0Int))

	bigVal2 := newBigFloat(float64(val1Int))

	result := new(big.Float).Quo(bigVal1, bigVal2)

	var IntResult *big.Int
	if ctx.Metadata["round"] == "up" {
		IntResult = roundUp(result)
	} else {
		IntResult = roundDown(result)
	}

	return encodeScalarValues(IntResult.Int64())
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

func encodeScalarValues(values ...any) ([]*ScalarValue, error) {
	scalarValues := make([]*ScalarValue, len(values))
	for i, v := range values {
		scalarValue, err := NewScalarValue(v)
		if err != nil {
			return nil, err
		}

		scalarValues[i] = scalarValue
	}

	return scalarValues, nil
}

const (
	precision = 128
)

func newBigFloat(num float64) *big.Float {
	bg := new(big.Float).SetPrec(precision)

	return bg.SetFloat64(num)
}
