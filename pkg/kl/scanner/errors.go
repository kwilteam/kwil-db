package scanner

import (
	"errors"
	"fmt"
)

type ErrorList []error

func (e *ErrorList) Add(msg string) {
	*e = append(*e, errors.New(msg))
}

func (e ErrorList) Error() string {
	switch len(e) {
	case 0:
		return "no errors"
	case 1:
		return e[0].Error()
	default:
		return fmt.Sprintf("%s (with %d+ errors)", e[0], len(e)-1)
	}
}
