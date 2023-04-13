package scanner

import (
	"fmt"
	"kwil/pkg/kuneiform/token"
)

type Error struct {
	Pos token.Position
	Msg string
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Pos.String(), e.Msg)
}

type ErrorList []*Error

func (e *ErrorList) Add(pos token.Position, msg string) {
	*e = append(*e, &Error{pos, msg})
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

func (e ErrorList) Err() error {
	if len(e) == 0 {
		return nil
	}
	return e
}
