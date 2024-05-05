// package Decimal implements a fixed-point decimal number.
// It is mostly a wrapper around github.com/cockroachdb/apd/v3, with some
// functionality that makes it easier to use in the context of Kwil. It enforces
// certain semantics of Kwil's decimal, such as precision and scale maxing out
// at 1000 (which is the maximum that Postgres supports).
package decimal

import (
	"fmt"

	"github.com/cockroachdb/apd/v3"
)

// Decimal is a decimal number.
type Decimal struct {
	dec       apd.Decimal
	context   apd.Context
	scale     uint16
	precision uint16
}

// New creates a new Decimal from a string.
// The precision must be between 1 and 1000, and the scale must be between 0 and precision.
func New(s string, precision, scale uint16) (*Decimal, error) {
	dec := &Decimal{
		context:   defaultContext(),
		scale:     scale,
		precision: precision,
	}

	if precision > 1000 {
		return nil, fmt.Errorf("precision too large: %d", precision)
	}
	if scale > 1000 {
		return nil, fmt.Errorf("scale too large: %d", scale)
	}
	if precision < 1 {
		return nil, fmt.Errorf("precision must be at least 1: %d", precision)
	}
	if scale > precision {
		return nil, fmt.Errorf("scale must be less than or equal to precision: %d > %d", scale, precision)
	}

	dec.context.Precision = uint32(precision)

	res, _, err := dec.context.NewFromString(s)
	if err != nil {
		return nil, err
	}

	dec.dec = *res

	// Enforce scale
	if err := dec.enforceScale(); err != nil {
		return nil, err
	}

	return dec, nil
}

// Scale returns the scale of the decimal.
// This is the number of digits to the right of the decimal point.
// It will be a value between 0 and 1000
func (d *Decimal) Scale() uint16 {
	return d.scale
}

// Precision returns the precision of the decimal.
// This is the number of significant digits in the decimal.
// It will be a value between 1 and 1000
func (d *Decimal) Precision() uint16 {
	return d.precision
}

// IsNegative returns true if the decimal is negative.
func (d *Decimal) IsNegative() bool {
	return d.dec.Negative
}

// String returns the string representation of the decimal.
func (d *Decimal) String() string {
	return d.dec.String()
}

// Add adds two decimals together.
func (d *Decimal) Add(d2 *Decimal) (*Decimal, error) {
	res := &Decimal{
		context: defaultContext(),
	}

	_, err := d.context.Add(&res.dec, &d.dec, &d2.dec)

	if err != nil {
		return nil, err
	}

	if err := res.overflowed(); err != nil {
		return nil, err
	}

	return res, nil
}

// Float64 returns the decimal as a float64.
func (d *Decimal) Float64() (float64, error) {
	return d.dec.Float64()
}

var ErrOverflow = fmt.Errorf("overflow")

// overflowed returns true if the decimal overflowed.
// apd can handle arbitrary precision, but we want to enforce a maximum
// precision of 1000.
// TODO: I think we can get rid of this since apd will handle this for us.
func (d *Decimal) overflowed() error {
	if d.Precision() > 1000 {
		return fmt.Errorf("%w: precision %d", ErrOverflow, d.Precision())
	}

	return nil
}

// defaultContext returns a copy of the default context.
func defaultContext() apd.Context {
	return apd.Context{
		Precision:   1000,
		MaxExponent: 1000,
		MinExponent: -1000,
		Traps:       apd.DefaultTraps,
		Rounding:    apd.RoundHalfUp,
	}
}

// enforceScale enforces scale on a decimal.
func (d *Decimal) enforceScale() error {
	_, err := d.context.Quantize(&d.dec, &d.dec, -int32(d.scale))
	return err
}
