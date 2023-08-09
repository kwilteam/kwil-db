package tx

import "math/big"

// ExecutionResponse is the response from any interaction that modifies state.
type ExecutionResponse struct {
	// Fee is the amount of tokens spent on the execution
	Fee *big.Int
}
