package polynomial

import (
	"fmt"
	"math"
	"math/big"
)

// a basic library for working with polynomials
// we use big float to guarantee precision

const precision uint = 128

func newPrecise() *big.Float {
	return new(big.Float).SetPrec(precision)
}

func NewFloat(f float64) *big.Float {
	return newPrecise().SetFloat64(f)
}

func NewFloatFromInt(i int64) *big.Float {
	return newPrecise().SetInt64(i)
}

func NewWeight(f float64) *Weight {
	return &Weight{Value: NewFloat(f)}
}

type Expression interface {
	Evaluate(variables map[string]*big.Float) (*big.Float, error)
	String() string

	// Variables returns a map of all variables used in the expression.
	// For example, the expression "4a + 2b" would return a map with keys "a" and "b".
	Variables() map[string]struct{}
}

type Addition struct {
	Left, Right Expression
}

func (a *Addition) Evaluate(variables map[string]*big.Float) (*big.Float, error) {
	l, err := a.Left.Evaluate(variables)
	if err != nil {
		return nil, err
	}
	r, err := a.Right.Evaluate(variables)
	if err != nil {
		return nil, err
	}
	return newPrecise().Add(l, r), nil
}

type Multiply struct {
	Left, Right Expression
}

func (m *Multiply) Evaluate(variables map[string]*big.Float) (*big.Float, error) {
	l, err := m.Left.Evaluate(variables)
	if err != nil {
		return nil, err
	}
	r, err := m.Right.Evaluate(variables)
	if err != nil {
		return nil, err
	}
	return newPrecise().Mul(l, r), nil
}

type Weight struct {
	Value *big.Float
}

func (w *Weight) Evaluate(_ map[string]*big.Float) (*big.Float, error) {
	return w.Value, nil
}

type Logarithm struct {
	Base, Expression Expression
}

func (l *Logarithm) Evaluate(variables map[string]*big.Float) (*big.Float, error) {
	base, err := l.Base.Evaluate(variables)
	if err != nil {
		return nil, err
	}
	expression, err := l.Expression.Evaluate(variables)
	if err != nil {
		return nil, err
	}

	smallExprFloat, _ := expression.Float64()
	if smallExprFloat <= 0 {
		return nil, fmt.Errorf("logarithm of non-positive number")
	}
	smallBaseFloat, _ := base.Float64()
	if smallBaseFloat <= 0 {
		return nil, fmt.Errorf("logarithm of non-positive base")
	}

	logResult := new(big.Float).Quo(
		newPrecise().SetFloat64(math.Log(smallExprFloat)),
		newPrecise().SetFloat64(math.Log(smallBaseFloat)))
	return logResult, nil
}

type Variable struct {
	Name string
}

func (v *Variable) Evaluate(variables map[string]*big.Float) (*big.Float, error) {
	value, ok := variables[v.Name]
	if !ok {
		return nil, fmt.Errorf("variable %s not found", v.Name)
	}
	return value, nil
}

type Division struct {
	Left, Right Expression
}

func (d *Division) Evaluate(variables map[string]*big.Float) (*big.Float, error) {
	l, err := d.Left.Evaluate(variables)
	if err != nil {
		return nil, err
	}
	r, err := d.Right.Evaluate(variables)
	if err != nil {
		return nil, err
	}

	// Check for division by zero
	if isZero(r) {
		return nil, fmt.Errorf("division by zero")
	}

	return new(big.Float).SetPrec(precision).Quo(l, r), nil
}

func isZero(f *big.Float) bool {
	return f.Cmp(newPrecise()) == 0
}

// print
// Add
func (a *Addition) String() string {
	return fmt.Sprintf("(%s + %s)", a.Left, a.Right)
}

// Multiply
func (m *Multiply) String() string {
	return fmt.Sprintf("(%s * %s)", m.Left, m.Right)
}

// Division
func (d *Division) String() string {
	return fmt.Sprintf("(%s / %s)", d.Left, d.Right)
}

// Weight
func (w *Weight) String() string {
	return w.Value.Text('g', -1)
}

// Logarithm
func (l *Logarithm) String() string {
	return fmt.Sprintf("log_%s(%s)", l.Base, l.Expression)
}

// Variable
func (v *Variable) String() string {
	return v.Name
}

// Variables
func (a *Addition) Variables() map[string]struct{} {
	return union(a.Left.Variables(), a.Right.Variables())
}

func (m *Multiply) Variables() map[string]struct{} {
	return union(m.Left.Variables(), m.Right.Variables())
}

func (d *Division) Variables() map[string]struct{} {
	return union(d.Left.Variables(), d.Right.Variables())
}

func (l *Logarithm) Variables() map[string]struct{} {
	return union(l.Base.Variables(), l.Expression.Variables())
}

func (v *Variable) Variables() map[string]struct{} {
	return map[string]struct{}{v.Name: {}}
}

func (w *Weight) Variables() map[string]struct{} {
	return map[string]struct{}{}
}

func union(a, b map[string]struct{}) map[string]struct{} {
	result := make(map[string]struct{})
	for k := range a {
		result[k] = struct{}{}
	}
	for k := range b {
		result[k] = struct{}{}
	}
	return result
}
