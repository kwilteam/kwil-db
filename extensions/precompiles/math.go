//go:build precompiles_math || ext_test

package precompiles

import (
	"context"
	"fmt"
	"math/big"

	"github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/node/types/sql"
)

func init() {
	err := RegisterPrecompile("math-precompile", PrecompileExtension[MathExtension]{
		Initialize: func(ctx context.Context, service *common.Service, db sql.DB, alias string, metadata map[string]any) (*MathExtension, error) {
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
		},
		Methods: []Method[MathExtension]{
			{
				Name:            "add",
				AccessModifiers: []Modifier{SYSTEM},
				Handler: func(ctx *common.EngineContext, app *common.App, inputs []any, resultFn func([]any) error, t *MathExtension) error {
					a, b, err := getArgs(inputs)
					if err != nil {
						return err
					}

					return resultFn([]any{a + b})
				},
			},
			{
				Name:            "subtract",
				AccessModifiers: []Modifier{SYSTEM},
				Handler: func(ctx *common.EngineContext, app *common.App, inputs []any, resultFn func([]any) error, t *MathExtension) error {
					a, b, err := getArgs(inputs)
					if err != nil {
						return err
					}

					return resultFn([]any{a - b})
				},
			},
			{
				Name:            "multiply",
				AccessModifiers: []Modifier{SYSTEM},
				Handler: func(ctx *common.EngineContext, app *common.App, inputs []any, resultFn func([]any) error, t *MathExtension) error {
					a, b, err := getArgs(inputs)
					if err != nil {
						return err
					}

					return resultFn([]any{a * b})
				},
			},
			{
				Name:            "divide",
				AccessModifiers: []Modifier{SYSTEM},
				Handler: func(ctx *common.EngineContext, app *common.App, inputs []any, resultFn func([]any) error, t *MathExtension) error {
					a, b, err := getArgs(inputs)
					if err != nil {
						return err
					}

					bigVal1 := newBigFloat(float64(a))

					bigVal2 := newBigFloat(float64(b))

					result := new(big.Float).Quo(bigVal1, bigVal2)

					var IntResult *big.Int
					var results []any
					if t.roundUp {
						IntResult = roundUp(result)
					} else {
						IntResult = roundDown(result)
					}
					results = append(results, IntResult)
					return resultFn(results)
				},
			},
		},
	})
	if err != nil {
		panic(err)
	}
}

// getArgs is a helper function that takes a slice of any and returns two integers and an error
func getArgs(args []any) (a, b int64, err error) {
	if len(args) != 2 {
		err = fmt.Errorf("expected 2 values, got %d", len(args))
		return
	}

	a, ok := args[0].(int64)
	if !ok {
		err = fmt.Errorf("argument 1 is not an int")
		return
	}

	b, ok = args[1].(int64)
	if !ok {
		err = fmt.Errorf("argument 2 is not an int")
		return
	}

	return a, b, nil
}

type MathExtension struct {
	roundUp bool // if true, round up.  If false, round down.
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
