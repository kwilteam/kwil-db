// Package muhash implements the MuHash algorithm, to allow order-independent
// set hashing.
package muhash

import (
	"crypto/sha256"
	"math/big"
)

var (
	// P is the secp256k1 prime 2^256 - 2^32 - 977
	// https://www.secg.org/sec2-v2.pdf section 2.4.1
	P, _ = new(big.Int).SetString("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEFFFFFC2F", 16)
)

// MuHash is a struct that represents a MuHash value. It uses the secp256k1
// prime, and the SHA256 hash function. Use New() to create a new MuHash.
type MuHash struct {
	value *big.Int
}

// New creates a new MuHash value.
func New() *MuHash {
	return &MuHash{value: big.NewInt(1)} // multiplicative identity
}

// Add adds a new element to the MuHash value.
func (m *MuHash) Add(data []byte) {
	if m.value == nil {
		m.value = big.NewInt(1)
	}
	h := hashToBigInt(data)
	m.value.Mul(m.value, h)
	m.value.Mod(m.value, P)
}

// Digest returns the MuHash value as a big.Int.
func (m *MuHash) Digest() *big.Int {
	return new(big.Int).Set(m.value)
}

// DigestHash returns the MuHash value as a 32-byte array.
func (m *MuHash) DigestHash() [32]byte {
	return sha256.Sum256(m.value.Bytes())
}

// Remove removes an element from the MuHash value.
// This is commented since it is not required for the use case of this project,
// but it is shown here for completeness.
// func (m *MuHash) Remove(data []byte) {
// 	if m.value == nil {
// 		m.value = big.NewInt(1)
// 	}
// 	h := hashToBigInt(data)
// 	inv := new(big.Int).ModInverse(h, P)
// 	m.value.Mul(m.value, inv)
// 	m.value.Mod(m.value, P)
// }

// Reset resets the MuHash value to its initial state.
func (m *MuHash) Reset() {
	m.value = big.NewInt(1)
}

func hashToBigInt(data []byte) *big.Int {
	h := sha256.Sum256(data)
	return new(big.Int).SetBytes(h[:])
}
