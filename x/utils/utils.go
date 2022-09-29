package utils

import (
	"context"
	"encoding/binary"
	"errors"
	"math/big"
	"os"
	"path"
	"runtime"
	"strings"
	"time"

	typ "kwil/x/types"
)

// Mod e.g., modulo returns the remainder of x/y
func Mod[T typ.Integer](x, y T) T {
	return (x%y + y) % y
}

// Max returns the largest of x or y.
func Max[T typ.Integer](x, y T) T {
	if x > y {
		return x
	}
	return y
}

// Min returns the smallest of x or y.
func Min[T typ.Integer](x, y T) T {
	if x < y {
		return x
	}
	return y
}

// IsContextTimeout returns true if the underlying error is a timeout. This function
// counts canceled contexts as timeouts.
// source article:
// https://blog.afoolishmanifesto.com/posts/context-deadlines-in-golang/
func IsContextTimeout(err error) bool {
	if err == nil {
		return false
	}

	switch {
	case errors.Is(err, context.Canceled):
		return true
	case errors.Is(err, context.DeadlineExceeded):
		return true
	default:
		return false
	}
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

func PanicIfError(err error) {
	if err != nil {
		panic(err)
	}
}

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

func AsDefault[T any]() T {
	var t T
	return t
}

func ExpandHomeDirAndEnv(path string) string {
	path = Ignore(path, TryExpandHomeDir)
	return os.ExpandEnv(path)
}

func ExpandHomeDir(path string) string {
	return Ignore(path, TryExpandHomeDir)
}

func TryExpandHomeDir(path string) (expandedPath string, expanded bool) {
	if !strings.HasPrefix(path, "~") {
		return path, false
	}

	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	return home + path[1:], true
}

func Ignore[T any, R1 any, R2 any](arg T, fn func(arg T) (R1, R2)) R1 {
	r, _ := fn(arg)
	return r
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

func Uint64ToBytes(i uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, i)
	return b
}

func GetGoFilePathOfCaller() string {
	_, filename, _, _ := runtime.Caller(1)

	return path.Dir(filename)
}

// Deadline implements the deadline/timeout resiliency pattern.
type Deadline struct {
	deadline time.Time
}

// Expiry will return the time at which the deadline will expire
func (d *Deadline) Expiry() time.Time {
	return d.deadline
}

// HasExpired will return true if the deadline has expired
func (d *Deadline) HasExpired() bool {
	return !d.deadline.Before(time.Now())
}

// NewDeadline constructs a new Deadline with the given timeout.
func NewDeadline(timeout time.Duration) *Deadline {
	return &Deadline{time.Now().Add(time.Millisecond * timeout)}
}
