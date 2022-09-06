package types

import (
	"math/big"
)

type Wal interface {
	BeginEthBlock(h *big.Int) error
	EndEthBlock(h *big.Int) error
	BeginTransaction(tx []byte) error
	EndTransaction(tx []byte) error
}
