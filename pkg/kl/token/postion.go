package token

import "fmt"

type Pos uint

type Position struct {
	Line   Pos
	Column Pos
}

func (p Position) String() string {
	return fmt.Sprintf("%d:%d", p.Line, p.Column)
}

type File struct {
	Name  string
	lines []int
}

func (f *File) Position(pos Pos) Position {
	return Position{Line: 1, Column: 1}
}
