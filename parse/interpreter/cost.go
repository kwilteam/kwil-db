package interpreter

// CostTable is a table of cost values for different operations.
type CostTable struct {
	// AllocateVariableCost is the cost to allocate a variable.
	AllocateVariableCost int64
	// SetVariableCost is the cost to set a variable that is already allocated.
	SetVariableCost int64
	// GetVariableCost is the cost to get a variable.
	GetVariableCost int64
	// ArrayAccessCost is the cost to access an array.
	ArrayAccessCost int64
	// MakeArrayCost is the cost to make an array.
	MakeArrayCost int64
	// ComparisonCost is the cost to compare two values.
	ComparisonCost int64
	// IsCost is the cost to execute an IS operation.
	IsCost int64
	// UnaryCost is the cost to execute a unary operation.
	UnaryCost int64
	// LogicalCost is the cost to execute a logical operation.
	LogicalCost int64
	// ArithmeticCost is the cost to execute an arithmetic operation.
	ArithmeticCost int64
	// LoopCost is the cost to execute a loop.
	// It is added for each iteration.
	LoopCost int64
	// BreakCost is the cost to execute a break statement.
	BreakCost int64
	// ReturnCost is the cost to execute a return statement.
	ReturnCost int64
	// SizeCostConstant is the cost for each byte in a constant.
	SizeCostConstant int64
	// NullSize is the size of a null value.
	// It will be multiplied by SizeCostConstant for each null value.
	NullSize int64
	// CallBuiltInFunctionCost is the cost to call a built-in function.
	// Functions might also individually have additional costs.
	CallBuiltInFunctionCost int64
}

// DefaultCostTable is the default cost table.
func DefaultCostTable() *CostTable {
	return &CostTable{
		AllocateVariableCost:    1,
		SetVariableCost:         1,
		GetVariableCost:         1,
		ArrayAccessCost:         1,
		MakeArrayCost:           1,
		ComparisonCost:          1,
		IsCost:                  1,
		UnaryCost:               1,
		LogicalCost:             1,
		ArithmeticCost:          1,
		LoopCost:                1,
		BreakCost:               1,
		ReturnCost:              1,
		SizeCostConstant:        1,
		NullSize:                1,
		CallBuiltInFunctionCost: 1,
	}
}

// ZeroCostTable is a cost table with zero costs.
func ZeroCostTable() *CostTable {
	return &CostTable{}
}
