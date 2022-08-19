package utils

import (
	"errors"
	"fmt"
	"log"
	"os"
)

type Error struct {
	inner error
	outer string
}

func (e Error) Print() Error {
	Print(e)
	return e
}

func (e Error) Println() Error {
	Println(e)
	return e
}

func (e Error) Fatal() {
	Fatal(e)
}

func (e Error) Fatalln() {
	Fatalln(e)
}

func (e Error) Panic() {
	panic(e)
}

func (e Error) Error() string {
	return e.outer
}

func (e Error) Unwrap() error {
	return UnwrapR(e)
}

func PanicIfError(err error) {
	if err != nil {
		panic(err)
	}
}

func PanicIfErrorMsg(err error, msg string) {
	if err != nil {
		panic(err)
	}
}

func PanicIfErrorMsgF(err error, format string, args ...interface{}) {
	if err != nil {
		panic(fmt.Sprintf(format, args...))
	}
}

func PanicIfErrorT[T any](v T, err error) T {
	PanicIfError(err)
	return v
}

func PanicIfErrorFn[T any](fn func() (T, error)) T {
	return PanicIfErrorT(fn())
}

func PanicIfErrorFn1[T any, U any](u U, fn func(u U) (T, error)) T {
	return PanicIfErrorT(fn(u))
}

func PanicIfErrorFn2[T any, U any, V any](u U, v V, fn func(u U, v V) (T, error)) T {
	return PanicIfErrorT(fn(u, v))
}

func PanicIfErrorAc(action func() error) {
	PanicIfError(action())
}

func PanicIfErrorAc1[T any](t T, action func(t T) error) {
	PanicIfError(action(t))
}

func PanicIfErrorAc2[T any, U any](t T, u U, action func(t T, u U) error) {
	PanicIfError(action(t, u))
}

func Print(err error) {
	e := errors.Unwrap(err)
	if e != nil {
		log.Print(e)
	}

	fmt.Print(err.Error())
}

func Println(err error) {
	e := errors.Unwrap(err)
	if e != nil {
		log.Println(e)
	}
	fmt.Println(err.Error())
}

func PrintIfError(err error) {
	if err != nil {
		Print(err)
	}
}

func PrintIfErrorLn(err error) {
	if err != nil {
		Println(err)
	}
}

func Fatal(err error) {
	Print(err)
	os.Exit(1)
}

func Fatalln(err error) {
	Println(err)
	os.Exit(1)
}

func FatalIfError(err error) {
	if err != nil {
		Fatal(err)
	}
}

func FatalIfErrorLn(err error) {
	if err != nil {
		Fatalln(err)
	}
}

func NewError(inner error, outer string) *Error {
	return &Error{inner, outer}
}

func NewErrorf(inner error, format string, args ...interface{}) *Error {
	return NewError(inner, fmt.Sprintf(format, args...))
}

func TryUnwrapR(err error) error {
	tmp := UnwrapR(err)
	if tmp == nil {
		return err
	}

	return tmp
}

func UnwrapR(err error) error {
	if err == nil {
		return nil
	}

	err = errors.Unwrap(err)
	if err == nil {
		return nil
	}

	for {
		tmp := errors.Unwrap(err)
		if tmp == nil {
			return err //return last non nil error
		}
		err = tmp
	}
}
