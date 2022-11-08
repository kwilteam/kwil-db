package ast

import "ksl"

type Document struct {
	Directives Directives
	Blocks     Blocks

	SrcRange ksl.Range
}

var _ Node = (*Document)(nil)
