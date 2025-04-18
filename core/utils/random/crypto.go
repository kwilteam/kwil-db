package random

import (
	crand "crypto/rand"
	"encoding/binary"
	"math/rand/v2"
)

// Source is a cryptographically secure random number source that satisfies the
// math/rand.Source and math/rand.Source64 interfaces, which are required to
// make a new math/rand.Rand instance that uses crypto/rand. See also New, for a
// new Rand instance using this source.
var Source source

type source struct{}

// Uint64 is part of the math/rand/v2.Source and math/rand.Source64 interfaces.
func (source) Uint64() uint64 {
	var b [8]byte
	crand.Read(b[:])
	return binary.LittleEndian.Uint64(b[:])
}

// Int63 is part of the math/rand.Source interface.
func (cs source) Int63() int64 {
	return int64(cs.Uint64() & ^uint64(1<<63)) // clear top bit, mask with (1<<63 - 1)
}

func (source) Seed(int64) {} // crypto/rand source is not seeded

// math/rand/v2.Source is just interface { Uint64() uint64 }
// so it does not need the Int64 method at all.
var _ rand.Source = Source
var _ rand.Source = &Source

// New creates a new math/rand.Rand number generator that uses the
// cryptographically secure source of randomness. This is helpful for the
// versatile methods like Intn, Float64, etc., which crypto/rand does not
// provide. If you just need bytes, use the standard library's crypto/rand.Read.
func New() *rand.Rand {
	return rand.New(Source)
}

var rng2 *rand.Rand = New()
