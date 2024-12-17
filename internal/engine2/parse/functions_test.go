package parse_test

import (
	"testing"

	"github.com/kwilteam/kwil-db/internal/engine2/parse"
)

// tests that we have implemented all functions
func Test_AllFunctionsImplemented(t *testing.T) {
	for name, fn := range parse.Functions {
		switch fnt := fn.(type) {
		case *parse.ScalarFunctionDefinition:
			if fnt.PGFormatFunc == nil {
				t.Errorf("function %s has no PGFormatFunc", name)
			}

			if fnt.ValidateArgsFunc == nil {
				t.Errorf("function %s has no ValidateArgsFunc", name)
			}
		case *parse.AggregateFunctionDefinition:
			if fnt.PGFormatFunc == nil {
				t.Errorf("function %s has no PGFormatFunc", name)
			}

			if fnt.ValidateArgsFunc == nil {
				t.Errorf("function %s has no ValidateArgsFunc", name)
			}
		case *parse.WindowFunctionDefinition:
			if fnt.PGFormatFunc == nil {
				t.Errorf("function %s has no PGFormatFunc", name)
			}

			if fnt.ValidateArgsFunc == nil {
				t.Errorf("function %s has no ValidateArgsFunc", name)
			}
		default:
			t.Errorf("function %s is not a scalar, aggregate, or window function", name)
		}
	}
}
