package polynomial

import "math/big"

func WeightedVar(name string, weight *big.Float) Expression {
	return &Multiply{
		Left:  &Weight{Value: weight},
		Right: &Variable{Name: name},
	}
}

func Mul(a, b Expression) Expression {
	return &Multiply{
		Left:  a,
		Right: b,
	}
}

func Add(a, b Expression) Expression {
	return &Addition{
		Left:  a,
		Right: b,
	}
}

func Sub(a, b Expression) Expression {
	return &Addition{
		Left:  a,
		Right: Mul(b, NewWeight(-1)),
	}
}

func Div(a, b Expression) Expression {
	return &Division{
		Left:  a,
		Right: b,
	}
}

// Log returns the logarithm of a with base "base".
// If base is not provided, it defaults to 2.
func Log(a Expression, base ...Expression) Expression {
	var b Expression
	b = NewWeight(2)
	if len(base) > 0 {
		b = base[0]
	}
	return &Logarithm{
		Base:       b,
		Expression: a,
	}
}
