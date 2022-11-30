package sqlschema

import "ksl"

type SchemaIdentifier interface{ id() }

type TableID int
type EnumID int
type ColumnID int
type IndexID int
type IndexPartID int
type ForeignKeyID int
type ViewID int
type ExtensionID int

func (TableID) id()      {}
func (ColumnID) id()     {}
func (ForeignKeyID) id() {}
func (EnumID) id()       {}
func (IndexID) id()      {}
func (IndexPartID) id()  {}
func (ViewID) id()       {}
func (ExtensionID) id()  {}

type ConnectorData interface{ cdata() }

type QualName struct {
	Namespace string
	Name      string
}

type IndexType int

const (
	NormalIndex IndexType = iota
	UniqueIndex
	PrimaryKeyIndex
)

type ForeignKeyAction int

const (
	NoAction ForeignKeyAction = iota
	Restrict
	Cascade
	SetNull
	SetDefault
)

func (f ForeignKeyAction) DDL() string {
	switch f {
	case NoAction:
		return "NO ACTION"
	case Restrict:
		return "RESTRICT"
	case Cascade:
		return "CASCADE"
	case SetNull:
		return "SET NULL"
	case SetDefault:
		return "SET DEFAULT"
	default:
		panic("unknown foreign key action")
	}
}

type IndexAlgorithm int

const (
	UnspecifiedAlgo IndexAlgorithm = iota
	BTreeAlgo
	GinAlgo
	GistAlgo
	HashAlgo
	BrinAlgo
	SpGistAlgo
)

func (a IndexAlgorithm) String() string {
	switch a {
	case BTreeAlgo:
		return "BTREE"
	case GinAlgo:
		return "GIN"
	case GistAlgo:
		return "GIST"
	case HashAlgo:
		return "HASH"
	case BrinAlgo:
		return "BRIN"
	case SpGistAlgo:
		return "SPGIST"
	default:
		return ""
	}
}

type SortOrder int

const (
	Ascending SortOrder = iota
	Descending
)

type ColumnArity int

const (
	Required ColumnArity = iota
	Nullable
	List
)

type Value interface{ val() }

type Sequence struct {
	Sequence string
}

func (Sequence) val() {}

type UniqueRowID struct{}

func (UniqueRowID) val() {}

type DBGenerated struct {
	Value string
}

func (DBGenerated) val() {}

type Extension struct {
	Name        string
	Namespace   string
	Version     string
	Relocatable bool
}

type EnumType struct {
	ksl.Type
	Name string
	ID   EnumID
}
