//go:build precompiles_math || ext_test

package precompiles

import (
	"fmt"
	"math/big"

	"github.com/kwilteam/kwil-db/common"
)

func init() {
	err := RegisterPrecompile("math", InitializeMath)
	if err != nil {
		panic(err)
	}
}

type MathExtension struct {
	roundUp bool // if true, round up.  If false, round down.
}

// this initialize function checks if round is set.  If not, it sets it to "up"
func InitializeMath(ctx *DeploymentContext, service *common.Service, metadata map[string]string) (Instance, error) {
	_, ok := metadata["round"]
	if !ok {
		metadata["round"] = "up"
	}

	roundVal := metadata["round"]
	if roundVal != "up" && roundVal != "down" {
		return nil, fmt.Errorf("round must be either 'up' or 'down'. default is 'up'")
	}

	roundUp := roundVal == "up"

	return &MathExtension{roundUp: roundUp}, nil
}

func (e *MathExtension) Call(ctx *ProcedureContext, app *common.App, method string, inputs []any) ([]any, error) {
	switch method {
	case "add":
		return e.add(inputs...)
	case "subtract":
		return e.subtract(inputs...)
	case "multiply":
		return e.multiply(inputs...)
	case "divide":
		return e.divide(inputs...)
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
func (e *MathExtension) divide(values ...any) ([]any, error) {
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
	if e.roundUp {
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
