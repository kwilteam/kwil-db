package validators

// intDivUp divides two integers, rounding up, without using floating point
// conversion. This will panic if the denominator is zero, just like regular
// integer division.
func intDivUp(val, div int64) int64 {
	// https://github.com/rust-lang/rust/blob/343889b7234bf786e2bc673029467052f22fca08/library/core/src/num/uint_macros.rs#L2061
	q, rem := val/div, val%div
	if (rem > 0 && div > 0) || (rem < 0 && div < 0) {
		q++
	}
	return q
	// rumor is that this is just as good: (val + div - 1) / div
}
