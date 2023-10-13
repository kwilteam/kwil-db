package abci

import (
	"fmt"
	"math"
)

type Gas uint64

// GasMeter is an interface to track gas consumption. It's expected to be used by the modules or store.
type GasMeter interface {
	// GasConsumed returns the amount of gas consumed.
	GasConsumed() Gas
	// GasRemaining returns the amount of gas remaining.
	GasRemaining() Gas
	// ConsumeGas consumes the given amount of gas from the meter.
	ConsumeGas(amount Gas, caller string) error
}

type simpleGasMeter struct {
	consumed Gas
	limit    Gas
}

func NewSimpleGasMeter(limit Gas) GasMeter {
	return &simpleGasMeter{
		limit:    limit,
		consumed: 0,
	}
}

func (s *simpleGasMeter) GasConsumed() Gas {
	return s.consumed
}

func (s *simpleGasMeter) GasRemaining() Gas {
	return s.limit - s.consumed
}

func (s *simpleGasMeter) ConsumeGas(amount Gas, caller string) error {
	if math.MaxUint64-s.consumed < amount {
		return fmt.Errorf("gas overflow, %s", caller)
	}

	if s.consumed+amount > s.limit {
		return fmt.Errorf("gas limit exceeded, %s", caller)
	}
	s.consumed += amount
	return nil
}
