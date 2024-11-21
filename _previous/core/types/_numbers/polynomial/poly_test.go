package polynomial_test

import (
	"math/big"
	"testing"

	poly "github.com/kwilteam/kwil-db/core/types/_numbers/polynomial"
)

func Test_Polynomial(t *testing.T) {
	a := poly.WeightedVar("a", poly.NewFloat(1))
	b := poly.WeightedVar("b", poly.NewFloat(2))

	// val = a + 2b
	p := poly.Add(a, b)

	vars := map[string]*big.Float{
		"a": poly.NewFloat(1),
		"b": poly.NewFloat(2),
	}

	res, err := p.Evaluate(vars)
	if err != nil {
		t.Error(err)
	}

	// val = 1 + 2*2 = 5
	if res.Cmp(poly.NewFloat(5)) != 0 {
		t.Errorf("expected 5, got %v", res)
	}

	variables := p.Variables()
	if len(variables) != 2 {
		t.Error("expected 2 variables")
	}
}

func Test_Polynomial_2(t *testing.T) {
	// going to build val = 2a + (3b+5a*2c)+2c*log2(2c)
	// a = 103, b = 43, c = 24
	// 2 * 103 = 206
	// 3 * 43 = 129
	// 5 * 103 * 2 * 24 = 24720
	// 2 * 24 = 48
	//log2(2*24) = 5.5849625007
	// 206 +(129+24720)+48*5.5849625007 = 25213.5849625007
	// 25055 + 268.08 = 25323.08

	firstVar := poly.WeightedVar("a", poly.NewFloat(2))
	secondVar := poly.WeightedVar("b", poly.NewFloat(3))
	thirdVar := poly.WeightedVar("a", poly.NewFloat(5))
	fourthVar := poly.WeightedVar("c", poly.NewFloat(2))
	fifthVar := poly.WeightedVar("c", poly.NewFloat(2))
	sixthVar := poly.WeightedVar("c", poly.NewFloat(2))

	// evaluate 5a*2c
	expr2Multiply := poly.Mul(thirdVar, fourthVar)

	// evaluate 3b+5a*2c
	expr2 := poly.Add(secondVar, expr2Multiply)

	// evaluate log2(2c)
	expr3Log := poly.Log(fifthVar)

	// evaluate 2c*log2(2c)
	expr3 := poly.Mul(sixthVar, expr3Log)

	// evaluate 2a + (3b+5a*2c)
	expr1and2 := poly.Add(firstVar, expr2)

	// evaluate 2a + (3b+5a*2c)+2c*log2(2c)
	expr := poly.Add(expr1and2, expr3)

	vars := map[string]*big.Float{
		"a": poly.NewFloat(103),
		"b": poly.NewFloat(43),
		"c": poly.NewFloat(24),
	}

	res, err := expr.Evaluate(vars)
	if err != nil {
		t.Error(err)
	}

	intNum, _ := res.Int64()
	if intNum != 25323 {
		t.Errorf("expected 25323, got %v", res)
	}
}
