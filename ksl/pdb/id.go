package pdb

type NodeID interface{ nodeid() }
type ModelID int
type BlockID int
type EnumID int
type FieldID int
type IndexID int
type DirectiveID int
type RelationID int

type ModelFieldID struct {
	model ModelID
	field FieldID
}

func (mf ModelFieldID) Model() ModelID { return mf.model }
func (mf ModelFieldID) Field() FieldID { return mf.field }

type EnumValueID struct {
	enum  EnumID
	value IndexID
}

func (ev EnumValueID) Enum() EnumID   { return ev.enum }
func (ev EnumValueID) Value() IndexID { return ev.value }

type AnnotID struct {
	node  NodeID
	annot IndexID
}

func (a AnnotID) Node() NodeID   { return a.node }
func (a AnnotID) Index() IndexID { return a.annot }

func MakeAnnotID(node NodeID, annot IndexID) AnnotID {
	return AnnotID{node: node, annot: annot}
}

func MakeModelFieldID(model ModelID, field FieldID) ModelFieldID {
	return ModelFieldID{model: model, field: field}
}

func MakeEnumValueID(enum EnumID, value IndexID) EnumValueID {
	return EnumValueID{enum: enum, value: value}
}

func (ModelID) nodeid()      {}
func (BlockID) nodeid()      {}
func (EnumID) nodeid()       {}
func (ModelFieldID) nodeid() {}
func (EnumValueID) nodeid()  {}
func (AnnotID) nodeid()      {}
func (RelationID) nodeid()   {}
func (FieldID) nodeid()      {}
func (DirectiveID) nodeid()  {}
