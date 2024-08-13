package planner

import (
	"fmt"
	"testing"
)

func Test_Rewrite(t *testing.T) {
	a := &ArithmeticOp{
		Left:  &Literal{Value: 1},
		Right: &Literal{Value: 2},
		Op:    Add,
	}

	for _, c := range a.Children() {
		switch n := c.(type) {
		case *Literal:
			n.Value = 3
		}
	}

	fmt.Println(a)
	panic("a")
}
