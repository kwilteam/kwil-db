package source

import (
	"fmt"

	"github.com/kwilteam/kwil-db/internal/sql/sqlerr"
)

type Error struct {
	Filename string
	Line     int
	Column   int
	Err      error
}

func NewError(fileName string, source string, loc int, err error) error {
	line := 1
	column := 1
	if lerr, ok := err.(*sqlerr.Error); ok {
		if lerr.Location != 0 {
			loc = lerr.Location
		} else if lerr.Line != 0 && lerr.Column != 0 {
			line = lerr.Line
			column = lerr.Column
		}
	}
	if source != "" && loc != 0 {
		line, column = LineNumber(source, loc)
	}

	return &Error{
		Filename: fileName,
		Line:     line,
		Column:   column,
		Err:      err,
	}
}

func (e *Error) Unwrap() error {
	return e.Err
}

func (e *Error) Error() string {
	if e.Filename != "" {
		return fmt.Sprintf("%s:%d:%d: %s", e.Filename, e.Line, e.Column, e.Err.Error())
	}
	return fmt.Sprintf("%d:%d: %s", e.Line, e.Column, e.Err.Error())
}
