package ast

import (
	"ksl"
)

var _ Node = (*Type)(nil)

type Type struct {
	IsArray  bool
	Name     *Str
	Nullable bool

	SrcRange ksl.Range
}
