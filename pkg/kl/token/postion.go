package token

import (
	"fmt"
	"sort"
)

type Pos uint

type Position struct {
	Line   Pos // Line number, starting at 1.
	Column Pos // Column number, starting at 1 (character count per line).
}

func (p Position) String() string {
	return fmt.Sprintf("%d:%d", p.Line, p.Column)
}

type File struct {
	Size  int   // file size
	Lines []int // start offset of each line
}

func (f *File) Position(pos Pos) Position {
	offset := int(pos)

	if offset < 0 || offset > f.Size {
		panic(fmt.Sprintf("invalid offset %d, range: [0, %d]", offset, f.Size))
	}

	p := sort.Search(len(f.Lines), func(i int) bool { return f.Lines[i] > offset })

	return Position{Line: Pos(p), Column: Pos(offset - f.Lines[p-1] + 1)}
}

func (f *File) AddLine(offset int) {
	f.Lines = append(f.Lines, offset)
}
