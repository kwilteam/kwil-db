package common

type ComparisonOp uint8

const (
	Equal ComparisonOp = iota
	LessThan
	GreaterThan
	Is
	IsDistinctFrom
)

type UnaryOp uint8

const (
	Not UnaryOp = iota
	Neg
	Pos
)

type ArithmeticOp uint8

const (
	Add ArithmeticOp = iota
	Sub
	Mul
	Div
	Mod
	Concat
)
