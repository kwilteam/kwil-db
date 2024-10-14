package parse_test

import (
	"testing"

	"github.com/kwilteam/kwil-db/parse"
)

// tests that we have implemented all functions
func Test_AllFunctionsImplemented(t *testing.T) {
	for name, fn := range parse.Functions {
		scalar, ok := fn.(*parse.ScalarFunctionDefinition)
		if ok {
			if scalar.PGFormatFunc == nil {
				t.Errorf("function %s has no PGFormatFunc", name)
			}
			if scalar.ValidateArgsFunc == nil {
				t.Errorf("function %s has no ValidateArgsFunc", name)
			}
		} else {
			agg, ok := fn.(*parse.AggregateFunctionDefinition)
			if !ok {
				t.Errorf("function %s is not a scalar or aggregate function", name)
			}
			if agg.PGFormatFunc == nil {
				t.Errorf("function %s has no PGFormatFunc", name)
			}
			if agg.ValidateArgsFunc == nil {
				t.Errorf("function %s has no ValidateArgsFunc", name)
			}
		}
	}
}
