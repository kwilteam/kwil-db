package ksl

import (
	"fmt"
)

type Pos struct {
	Line   int
	Column int
	Offset int
}

var InitialPos = Pos{Offset: 0, Line: 1, Column: 1}

type Range struct {
	Filename   string
	Start, End Pos
}

func RangeBetween(start, end Range) Range {
	return Range{
		Filename: start.Filename,
		Start:    start.Start,
		End:      end.End,
	}
}

func RangeOver(a, b Range) Range {
	if a.Empty() {
		return b
	}
	if b.Empty() {
		return a
	}

	var start, end Pos
	if a.Start.Offset < b.Start.Offset {
		start = a.Start
	} else {
		start = b.Start
	}
	if a.End.Offset > b.End.Offset {
		end = a.End
	} else {
		end = b.End
	}
	return Range{
		Filename: a.Filename,
		Start:    start,
		End:      end,
	}
}

func (r Range) ContainsPos(pos Pos) bool {
	return r.ContainsOffset(pos.Offset)
}

func (r Range) ContainsOffset(offset int) bool {
	return offset >= r.Start.Offset && offset < r.End.Offset
}

func (r Range) Ptr() *Range {
	return &r
}

func (r Range) String() string {
	if r.Start.Line == r.End.Line {
		return fmt.Sprintf(
			"%s:%d,%d-%d",
			r.Filename,
			r.Start.Line, r.Start.Column,
			r.End.Column,
		)
	} else {
		return fmt.Sprintf(
			"%s:%d,%d-%d,%d",
			r.Filename,
			r.Start.Line, r.Start.Column,
			r.End.Line, r.End.Column,
		)
	}
}

func (r Range) Empty() bool {
	return r.Start.Offset == r.End.Offset
}

func (r Range) SliceBytes(b []byte) []byte {
	start := r.Start.Offset
	end := r.End.Offset
	if start < 0 {
		start = 0
	} else if start > len(b) {
		start = len(b)
	}
	if end < 0 {
		end = 0
	} else if end > len(b) {
		end = len(b)
	}
	if end < start {
		end = start
	}
	return b[start:end]
}

func (r Range) Overlaps(other Range) bool {
	switch {
	case r.Filename != other.Filename:
		// If the ranges are in different files then they can't possibly overlap
		return false
	case r.Empty() || other.Empty():
		// Empty ranges can never overlap
		return false
	case r.ContainsOffset(other.Start.Offset) || r.ContainsOffset(other.End.Offset):
		return true
	case other.ContainsOffset(r.Start.Offset) || other.ContainsOffset(r.End.Offset):
		return true
	default:
		return false
	}
}

func (r Range) Overlap(other Range) Range {
	if !r.Overlaps(other) {
		return Range{
			Filename: r.Filename,
			Start:    r.Start,
			End:      r.Start,
		}
	}

	var start, end Pos
	if r.Start.Offset > other.Start.Offset {
		start = r.Start
	} else {
		start = other.Start
	}
	if r.End.Offset < other.End.Offset {
		end = r.End
	} else {
		end = other.End
	}

	return Range{
		Filename: r.Filename,
		Start:    start,
		End:      end,
	}
}

func (r Range) PartitionAround(other Range) (before, overlap, after Range) {
	overlap = r.Overlap(other)
	if overlap.Empty() {
		return overlap, overlap, overlap
	}

	before = Range{
		Filename: r.Filename,
		Start:    r.Start,
		End:      overlap.Start,
	}
	after = Range{
		Filename: r.Filename,
		Start:    overlap.End,
		End:      r.End,
	}

	return before, overlap, after
}
