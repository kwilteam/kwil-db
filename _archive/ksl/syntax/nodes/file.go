package nodes

import "ksl"

type File struct {
	Name     string
	Entries  []TopLevel
	Contents []byte
	Span     ksl.Range
}

func (f *File) Range() ksl.Range { return f.Span }
