package store

import (
	"encoding/binary"
	"math/big"
)

func byteToBigInt(b []byte) *big.Int {
	return new(big.Int).SetBytes(b)
}

func int64ToBytes(i int64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(i))
	return b
}

func bytesToInt64(b []byte) int64 {
	return int64(binary.LittleEndian.Uint64(b))
}
