package pg

// This file defines a generic histogram type, and many of the interpolation
// functions necessary to use it and create its bounds.

import (
	"bytes"
	"cmp"
	"encoding/binary"
	"encoding/gob"
	"errors"
	"fmt"
	"math"
	"math/big"
	"reflect"
	"slices"
	"strings"

	"github.com/kwilteam/kwil-db/common/sql"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/types/decimal"
)

type histo[T any] struct {
	// Normally given N boundaries, there would be N-1 bins/freqs.
	// In this histogram, there are two special bins that catch all
	// below the lowest bin boundary, and above the highest bin boundary.
	// This is done because we need to decide the boundaries before the
	// actual min and max are known, *and* we have to limit the number of bins.
	//
	// bounds:       h_0     h_1     h_2
	//                |       |       |
	// freqs:   f_0   |  f_1  |  f_2  |  f_3
	//
	// bins include values on (h_i,h_j]

	bounds []T   // len n
	freqs  []int // len n+1

	// Rather than having the comparison and interpolation functions as
	// package-level functions that accept interfaces, we store the generic
	// (compile-time typed) functions:

	comp func(a, b T) int // -1 for a<b, 0 for a==b, 1 for a>b

	// For more accurate summation up to a given value that is not also equal to
	// one of the bounds, linear interpolation can be used. (TODO)
	// interp func(f float64, a, b T) T
	// interpF func(v, a, b T) float64 // (b-v)/(b-a) -- [a...v...b]
}

// Equal is defined to satisfy the go-cmp package to assist with unit tests.
func (h histo[T]) Equal(g histo[T]) bool {
	if c := slices.Compare(h.freqs, g.freqs); c != 0 {
		return false
	}
	eq := slices.CompareFunc(h.bounds, g.bounds, func(a, b T) int {
		return h.comp(a, b) // NOTE: compare(a, b) if we ditch this as a field
	})
	return eq == 0
}

const histoBinVer uint16 = 0

// MarshalBinary serializes the data fields of the histogram.
func (h histo[T]) MarshalBinary() ([]byte, error) {
	var buf bytes.Buffer
	buf.Write(binary.BigEndian.AppendUint16(nil, histoBinVer))

	H := struct {
		Bounds []T
		Freqs  []int
	}{
		Bounds: h.bounds,
		Freqs:  h.freqs,
	}

	// marshalBinaryV0...
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(H) // for T non-native, calls into MarshalBinary of T
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
	// we lose the comp func. restoreHistoFuncs when reloading the statistics or
	// when they are next used.
}

// UnmarshalBinary deserializes the data fields of the histogram and restores
// the function fields based on the type parameter.
func (h *histo[T]) UnmarshalBinary(b []byte) error {
	if len(b) < 2 {
		return errors.New("insufficient data for histo")
	}

	switch ver := binary.BigEndian.Uint16(b); ver {
	case histoBinVer:
		b = b[2:]
	default:
		return fmt.Errorf("unsupported histogram serialization verion %d", ver)
	}

	// unmarshalBinaryV0...
	dec := gob.NewDecoder(bytes.NewReader(b))
	H := struct {
		Bounds []T
		Freqs  []int
	}{
		Bounds: []T{},
		Freqs:  []int{},
	}
	err := dec.Decode(&H)
	if err != nil {
		return err
	}
	h.bounds = H.Bounds
	h.freqs = H.Freqs
	restoreHistoFuncs(h)
	return nil
}

// setHistoCmpFunc is used like cs.Histogram = restoreHisto(cs.Histogram)
func setHistoCmpFunc(h any) any {
	switch ht := h.(type) {
	case nil: // it was never set
		return nil
	case histo[[]byte]:
		ht.comp = bytes.Compare
		return ht
	case histo[int64]:
		ht.comp = cmp.Compare[int64]
		return ht
	case histo[float64]:
		ht.comp = cmp.Compare[float64]
		return ht
	case histo[string]:
		ht.comp = strings.Compare
		return ht
	case histo[bool]:
		ht.comp = cmpBool
		return ht
	case histo[*decimal.Decimal]:
		ht.comp = cmpDecimal
		return ht
	case histo[*types.Uint256]:
		ht.comp = types.CmpUint256
		return ht
	case histo[*types.UUID]:
		ht.comp = func(a, b *types.UUID) int {
			return types.CmpUUID(*a, *b)
		}
		return ht
	case histo[types.UUID]:
		ht.comp = types.CmpUUID
		return ht

	case histo[decimal.DecimalArray]: // TODO
	case histo[types.Uint256Array]: // TODO
	case histo[[]string]:
	case histo[[]int64]:

	default:
		panic(fmt.Sprintf("unrecognized histogram type %T", h))
	}

	return h // unmodifed for TODO arrays
}

func restoreStatsHistoFuncs(cs *sql.ColumnStatistics) {
	cs.Histogram = setHistoCmpFunc(cs.Histogram)
}

// restoreHistoFuncs is similar to histoCmpFunc, but updates via a pointer to
// the histo instance so that the entire histo does not need to be cloned.
func restoreHistoFuncs(h any) {
	switch ht := h.(type) {
	case nil: // it was never set
		return
	case *histo[[]byte]:
		ht.comp = bytes.Compare
	case *histo[int64]:
		ht.comp = cmp.Compare[int64]
	case *histo[float64]:
		ht.comp = cmp.Compare[float64]
	case *histo[string]:
		ht.comp = strings.Compare
	case *histo[bool]:
		ht.comp = cmpBool
	case *histo[*decimal.Decimal]:
		ht.comp = cmpDecimal
	case *histo[*types.Uint256]:
		ht.comp = types.CmpUint256
	case *histo[types.UUID]:
		ht.comp = types.CmpUUID
	case *histo[*types.UUID]:
		ht.comp = func(a, b *types.UUID) int {
			return types.CmpUUID(*a, *b)
		}

	case histo[decimal.DecimalArray]: // TODO
	case histo[types.Uint256Array]: // TODO
	case histo[[]string]:
	case histo[[]int64]:

	default:
		panic(fmt.Sprintf("unrecognized histogram type %T", h))
	}
}

// ins adds an observed value into the appropriate bin and returns the index of
// the updated bin.
func (h histo[T]) ins(v T) int {
	loc, _ := slices.BinarySearchFunc(h.bounds, v, h.comp)
	h.freqs[loc]++
	return loc
}

// rm removes an observed value from the appropriate bin and returns the index
// of the updated bin.
func (h histo[T]) rm(v T) int {
	loc, _ := slices.BinarySearchFunc(h.bounds, v, h.comp)
	h.freqs[loc]--
	if f := h.freqs[loc]; f < 0 {
		panic("accounting error -- negative bin count on rm")
	}
	return loc
}

func (h *histo[T]) TotalCount() int {
	var total int
	for _, f := range h.freqs {
		total += f
	}
	return total
}

func (h histo[T]) String() string {
	totalFreq := h.TotalCount()
	if len(h.bounds) > 0 {
		// only print the bounds slice for basic scalars, not variable length types.
		switch reflect.TypeOf(h.bounds[0]).Kind() {
		case reflect.Bool, reflect.Float64, reflect.Int, reflect.Int16,
			reflect.Int32, reflect.Int64, reflect.Int8, reflect.Uint,
			reflect.Uint16, reflect.Uint8, reflect.Uint32, reflect.Uint64:
		default:
			return fmt.Sprintf("total = %d, bounds = (len %d []%T), freqs = %v, cmp = %s",
				totalFreq, len(h.bounds), h.bounds[0], h.freqs, reflect.TypeOf(h.comp))
		}
	}
	return fmt.Sprintf("total = %d, bounds = %v, freqs = %v, cmp = %s",
		totalFreq, h.bounds, h.freqs, reflect.TypeOf(h.comp))
}

// ltTotal returns the cumulative frequency for values less than (or equal) to
// the given value. There will be a host of methods along these lines (e.g.
// range, greater, equal) to support selectivity computation.
//
// Presently this is a simple summation, but it should be updated to perform
// interpolation when a value is not also exactly a boundary.
func (h histo[T]) ltTotal(v T) int { //nolint:unused
	loc, _ := slices.BinarySearchFunc(h.bounds, v, h.comp)

	var freq int
	for i, f := range h.freqs {

		if i == loc {
			/*if found { // no interp from next bin, just add this freq and break
				freq += f
			} else { // the value is somewhere between bins (before bin i) => linearly interpolate
				freq += int(float64(f) * h.interp(v, h.bounds[i-1], h.bounds[i]))
			}*/
			freq += f
			break
		}

		freq += f
	}
	return freq
}

func (h histo[T]) gtTotal(v T) int { //nolint:unused
	loc, _ := slices.BinarySearchFunc(h.bounds, v, h.comp)
	var freq int
	for _, f := range h.freqs[loc:] {
		freq += f
	}
	return freq
}

func (h histo[T]) rangeTotal(v0, v1 T) int { //nolint:unused
	loc0, _ := slices.BinarySearchFunc(h.bounds, v0, h.comp)
	loc1, _ := slices.BinarySearchFunc(h.bounds, v1, h.comp)
	var freq int
	for _, f := range h.freqs[loc0:loc1] {
		freq += f
	}
	return freq
}

type SignedInt interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

type UnsignedInt interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}

type Num interface {
	SignedInt | UnsignedInt | float32 | float64
}

// The following interpolation functions are used to create histogram bounds
// from min and max values. They will also be used for partial summation.

func interpNumF[T Num](f float64, a, b T) float64 {
	return f*float64(b) + (1-f)*float64(a)
	// return float64(a) + f*(float64(b)-float64(a))
}

func interpNum[T Num](f float64, a, b T) T {
	return T(interpNumF(f, a, b))
	// return a + T(f*(float64(b)-float64(a)))
}

func interpBig(f float64, a, b *big.Int) *big.Int {
	if b.Cmp(a) <= 0 {
		panic("b must not be less than a")
	}

	// return a number on [a,b] computed via interpolation with f on [0,1]
	// representing where between the two numbers.

	diff := new(big.Int).Sub(b, a)
	frac := new(big.Float).SetPrec(big.MaxPrec).SetInt(diff)
	frac = frac.Mul(frac, big.NewFloat(f).SetPrec(big.MaxPrec)) // f*(b-a)

	// a + frac
	fracInt, _ := frac.Int(nil)

	return new(big.Int).Add(a, fracInt)
}

// interpUint256 and interpDec below are both defined in terms of interpBig.

func interpUint256(f float64, a, b *types.Uint256) *types.Uint256 {
	c := interpBig(f, a.ToBig(), b.ToBig())
	d, err := types.Uint256FromBig(c)
	if err != nil { // shouldn't even possible if a and b were not NaNs
		panic(err.Error())
	}
	return d
}

func interpDec(f float64, a, b *decimal.Decimal) *decimal.Decimal {
	c := interpBig(f, a.BigInt(), b.BigInt())
	// d, err := decimal.NewFromBigInt(c, a.Exp())
	// d.SetPrecisionAndScale(a.Precision(), a.Scale())
	d, err := decimal.NewExplicit(c.String(), a.Precision(), a.Scale())
	if err != nil {
		panic(err.Error())
	}
	// This is messier with Decimal's math methods, and I don't yet know if
	// there'd be a benefit:
	// bma, err := decimal.Sub(b, a) // etc
	return d
}

func interpBts(f float64, a, b []byte) []byte {
	ai := big.NewInt(0).SetBytes(a)
	bi := big.NewInt(0).SetBytes(b)
	return interpBig(f, ai, bi).Bytes()
}

func interpUUID(f float64, a, b types.UUID) types.UUID {
	return types.UUID(interpBts(f, a[:], b[:]))
}

// interpBool is largely nonsense and should be unused as there should not ever
// be a boolean histogram, but it is here for completeness. Could just panic...
func interpBool(f float64, a, b bool) bool {
	if f < 0.5 {
		return a
	}
	return b
}

// interpString needs more consideration and testing. It MUST be consistent with
// lexicographic comparison a la strings.Compare (e.g. "x" > "abc"), so we can't
// just interpolate as bytes, which takes numerics semantics. For now we
// right-pad to make them the same length, then interpolate each character.
func interpString(f float64, a, b string) string {
	if f < 0 || f > 1 {
		panic("f out of range")
	}
	if a > b {
		panic("a > b")
	}
	// Ensure both strings are the same length by padding with \0
	maxLen := len(a)
	if len(b) > maxLen {
		maxLen = len(b)
	}

	a = padString(a, maxLen)
	b = padString(b, maxLen)

	result := make([]byte, maxLen)
	// Interpolate each character, wonky but remains legible
	for i := range result {
		charA, charB := a[i], b[i]
		diff := float64(charB) - float64(charA)
		result[i] = charA + byte(math.Round(f*diff))
	}

	// Convert the byte slice back to a string and trim any null characters
	return strings.TrimRight(string(result), "\x00")
}

func padString(s string, length int) string {
	if len(s) < length {
		return s + strings.Repeat("\x00", length-len(s))
	}
	return s
}

// makeBounds computes n+1 evenly spaced histogram bounds given the range [a,b].
// However, if two bounds would be equal, as is possible when T is an integer
// and n is too small, there will be fewer bounds.
func makeBounds[T any](n int, a, b T, comp func(a, b T) int, interp func(f float64, a, b T) T) []T {
	if comp(a, b) == 1 {
		panic("no good")
	}
	bounds := make([]T, 0, n+1)
	f := 1 / float64(n)
	for i := 0; i <= n; i++ {
		next := interp(f*float64(i), a, b)
		if i > 0 && comp(next, bounds[len(bounds)-1]) == 0 {
			continue // trying to over-subdivide, easy with integers
		}
		bounds = append(bounds, next)
	}
	return bounds
}

func newHisto[T any](bounds []T) histo[T] {
	h := histo[T]{
		bounds: bounds,
		freqs:  make([]int, len(bounds)+1),
	}
	restoreHistoFuncs(&h)
	return h
}

func makeHisto[T any](bounds []T, comp func(a, b T) int) histo[T] {
	return histo[T]{
		bounds: bounds,
		freqs:  make([]int, len(bounds)+1),
		comp:   comp,
	}
}

// interpNumRat provides a possibly more accurate approach, particularly when T
// is floating point. May remove.
func interpNumRat[T Num](f *big.Rat, a, b T) T {
	return T(interpNumRatF(f, a, b))
}

func interpNumRatF[T Num](f *big.Rat, a, b T) float64 {
	ra, _ := new(big.Float).SetPrec(big.MaxPrec).SetFloat64(float64(a)).Rat(nil) // new(big.Rat).SetFloat64(float64(a))
	rb, _ := new(big.Float).SetPrec(big.MaxPrec).SetFloat64(float64(b)).Rat(nil) // new(big.Rat).SetFloat64(float64(b))

	// a + f*(b-a)
	rbma := new(big.Rat).Sub(rb, ra)
	fab := new(big.Rat).Mul(f, rbma)
	res, exact := new(big.Rat).Add(ra, fab).Float64()

	if exact {
		return res
	}

	// just go with all float
	ff, _ := f.Float64()
	return interpNumF(ff, a, b)
}

func makeBoundsNum[T Num](n int, a, b T) []T {
	if b <= a {
		panic("no good")
	}
	bounds := make([]T, 0, n+1)
	// f := 1 / float64(n)
	for i := 0; i <= n; i++ {
		fi := big.NewRat(int64(i), int64(n)) // 1/n
		next := interpNumRat(fi, a, b)
		// next := interpNum(f*float64(i), a, b)
		if i > 0 && next == bounds[len(bounds)-1] {
			continue // trying to over-subdivide, easy with integers
		}
		bounds = append(bounds, next)
	}
	return bounds
}
