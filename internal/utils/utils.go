package utils

import (
	"encoding/binary"
	"math/big"
	"os"
)

func Coalesce[T comparable](check T, alt T) T {
	var d T
	if check == d {
		return check
	}
	return alt
}

func CoalesceP[T any](check *T, alt *T) *T {
	if !IsNil(check) {
		return check
	}
	return alt
}

func Any[T comparable](compare T, params ...T) bool {
	for _, param := range params {
		if compare == param {
			return true
		}
	}
	return false
}

func All[T comparable](compare T, params ...T) bool {
	for _, param := range params {
		if compare != param {
			return false
		}
	}
	return true
}

func First[T any](values []T, comparer func(v T) bool) (bool, T) {
	for _, v := range values {
		if comparer(v) {
			return true, v
		}
	}

	var t T

	return false, t
}

func FirstOrDefault[T any](values []T, comparer func(v T) bool) (defaultValue T) {
	for _, v := range values {
		if comparer(v) {
			defaultValue = v
			return
		}
	}
	return
}

func IfElse[T any](predicate bool, ifReturn, elseReturn T) T {
	if predicate {
		return ifReturn
	}
	return elseReturn
}

func IsDefault[T comparable](compare T) bool {
	var d T
	return compare == d
}

func IsNotDefault[T comparable](compare T) bool {
	var d T
	return compare != d
}

func IsNil[T any](v *T) bool {
	return v == nil
}

func IsNotNil[T any](v *T) bool {
	return v != nil
}

func DEFAULT[T any]() T {
	var t T
	return t
}

func FileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil && os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func DirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil && os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

// AppendByteArrLength Will append the slice length as uint16 to the end of the byte slice
// Use this function instead of doing it manually since this uses uint16 instead of int64
func AppendByteArrLength(b []byte, a []byte) []byte {
	return append(b, Uint16ToBytes(uint16(len(a)))...)
}

// BigInt2Bytes This function converts a big int to bytes.  The result will always be a byte slice of length 16.
func BigInt2Bytes(h *big.Int) []byte {
	b := make([]byte, 16)
	k := h.FillBytes(b)
	return k
}

func Uint16ToBytes(i uint16) []byte {
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, i)
	return b
}
