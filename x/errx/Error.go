package errx

import (
	"fmt"
	"runtime"
)

//ref: https://mazux.medium.com/golang-error-handling-the-neat-way-4b7ec3ac5d6

type AppErr struct {
	msg   string
	code  int
	trace string
}

func (e AppErr) Error() string {
	if e.code < 0 {
		return fmt.Sprintf("Msg: %s, trace:\n %s", e.msg, e.trace)
	}
	return fmt.Sprintf("Msg: %s, code: %d, trace:\n %s", e.msg, e.code, e.trace)
}

func NewCodedErr(msg string, code int) AppErr {
	stackSlice := make([]byte, 512)
	s := runtime.Stack(stackSlice, false)
	return AppErr{msg, code, fmt.Sprintf("\n%s", stackSlice[0:s])}
}

func NewErr(msg string) AppErr {
	stackSlice := make([]byte, 512)
	s := runtime.Stack(stackSlice, false)
	return AppErr{msg, 0, fmt.Sprintf("\n%s", stackSlice[0:s])}
}
