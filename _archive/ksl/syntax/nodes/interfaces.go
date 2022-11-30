package nodes

import "ksl"

type TopLevel interface {
	Node
	Identifiable
	Named
	top()
}

func GetTopLevelType(top TopLevel) string {
	switch top.(type) {
	case *Block:
		return "block"
	case *Model:
		return "model"
	case *Annotation:
		return "directive"
	case *Enum:
		return "enum"
	default:
		panic("unknown top level type")
	}
}

func (Block) top()      {}
func (Model) top()      {}
func (Annotation) top() {}
func (Enum) top()       {}

type NamedNode interface {
	Node
	Named
}

type Node interface{ Range() ksl.Range }
type NodeId string

var _ Node = (*Model)(nil)
var _ Node = (*Name)(nil)
var _ Node = (*Field)(nil)
var _ Node = (*Function)(nil)
var _ Node = (*Literal)(nil)
var _ Node = (*List)(nil)
var _ Node = (*String)(nil)
var _ Node = (*Number)(nil)
var _ Node = (*Enum)(nil)
var _ Node = (*EnumValue)(nil)
var _ Node = (*Comment)(nil)
var _ Node = (*Block)(nil)
var _ Node = (*Property)(nil)
var _ Node = (*ArgumentList)(nil)
var _ Node = (*Argument)(nil)
var _ Node = (*Annotation)(nil)

type Named interface{ GetName() string }

var _ Named = (*Model)(nil)
var _ Named = (*Annotation)(nil)
var _ Named = (*Argument)(nil)
var _ Named = (*Block)(nil)
var _ Named = (*Property)(nil)
var _ Named = (*Enum)(nil)
var _ Named = (*EnumValue)(nil)
var _ Named = (*Field)(nil)

type Documented interface{ Documentation() string }

var _ Documented = (*Model)(nil)
var _ Documented = (*Block)(nil)
var _ Documented = (*Enum)(nil)
var _ Documented = (*EnumValue)(nil)
var _ Documented = (*Field)(nil)

type Annotated interface {
	Node
	GetAnnotations() []*Annotation
}

var _ Annotated = (*Enum)(nil)
var _ Annotated = (*EnumValue)(nil)
var _ Annotated = (*Field)(nil)
var _ Annotated = (*Model)(nil)

type Identifiable interface{ GetNameNode() *Name }

var _ Identifiable = (*Model)(nil)
var _ Identifiable = (*Annotation)(nil)
var _ Identifiable = (*Argument)(nil)
var _ Identifiable = (*Block)(nil)
var _ Identifiable = (*Property)(nil)
var _ Identifiable = (*Enum)(nil)
var _ Identifiable = (*EnumValue)(nil)
var _ Identifiable = (*Field)(nil)

type ArgumentHolder interface {
	Node
	GetArgs() Arguments
}
