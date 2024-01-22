package mathutil

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/kwilteam/kwil-db/internal/engine/execution"
)

func InitializeMathUtil(ctx *execution.DeploymentContext, metadata map[string]string) (execution.ExtensionNamespace, error) {
	if len(metadata) != 0 {
		return nil, fmt.Errorf("mathutil does not take any configs")
	}

	return &mathUtilExt{}, nil
}

var _ = execution.ExtensionInitializer(InitializeMathUtil)

type mathUtilExt struct{}

var _ = execution.ExtensionNamespace(&mathUtilExt{})

func (m *mathUtilExt) Call(scoper *execution.ProcedureContext, method string, inputs []any) ([]any, error) {
	switch strings.ToLower(method) {
	case knownMethodFraction:
		if len(inputs) != 3 {
			return nil, fmt.Errorf("expected 3 inputs, got %d", len(inputs))
		}

		number, ok := inputs[0].(int64)
		if !ok {
			return nil, fmt.Errorf("expected int64 for arg 1, got %T", inputs[0])
		}

		numerator, ok := inputs[1].(int64)
		if !ok {
			return nil, fmt.Errorf("expected int64 for arg 2, got %T", inputs[1])
		}

		denominator, ok := inputs[2].(int64)
		if !ok {
			return nil, fmt.Errorf("expected int64 for arg 3, got %T", inputs[2])
		}

		return fraction(number, numerator, denominator)
	default:
		return nil, fmt.Errorf("unknown method '%s'", method)
	}
}

func fraction(number, numerator, denominator int64) ([]any, error) {
	if denominator == 0 {
		return nil, fmt.Errorf("denominator cannot be zero")
	}

	// we will simply rely on go's integer division to truncate (round down)
	// we will use big math to avoid overflow
	bigNumber := big.NewInt(number)
	bigNumerator := big.NewInt(numerator)
	bigDenominator := big.NewInt(denominator)

	// (numerator/denominator) * number

	// numerator * number
	bigProduct := new(big.Int).Mul(bigNumerator, bigNumber)

	// numerator * number / denominator
	bigQuotient := new(big.Int).Div(bigProduct, bigDenominator)

	return []any{bigQuotient.Int64()}, nil
}

const (
	knownMethodFraction = "fraction"
)
