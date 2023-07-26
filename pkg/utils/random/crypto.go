package random

import (
	crand "crypto/rand"
	"encoding/binary"
	"math/rand"
)

// Source is a cryptographically secure random number source that satisfies the
// math/rand.Source and math/rand.Source64 interfaces, which are required to
// make a new math/rand.Rand instance that uses crypto/rand. See also New, for a
// new Rand instance using this source.
var Source source

type source struct{}

func (source) Uint64() uint64 {
	var b [8]byte
	crand.Read(b[:])
	return binary.LittleEndian.Uint64(b[:])
}

func (cs source) Int63() int64 {
	return int64(cs.Uint64() & ^uint64(1<<63)) // clear top bit, mask with (1<<63 - 1)
}

func (source) Seed(int64) {} // crypto/rand source is not seeded

var _ rand.Source = Source
var _ rand.Source64 = Source
var _ rand.Source = &Source
var _ rand.Source64 = &Source

// New creates a new math/rand.Rand number generator that uses the
// cryptographically secure source of randomness.
func New() *rand.Rand {
	return rand.New(Source)
}
