package schema

type KwilType string
type PGType string

const (
	KwilNull        KwilType = "null"
	KwilInt32       KwilType = "int32"
	KwilInt64       KwilType = "int64"
	KwilString      KwilType = "string"
	KwilBool        KwilType = "bool"
	KwilDate        KwilType = "date"
	KwilTime        KwilType = "time"
	KwilDatetime    KwilType = "datetime"
	KwilUnknownType KwilType = "unknown"
)

const (
	PGNull        PGType = "null"
	PGInt32       PGType = "int4"
	PGInt64       PGType = "int8"
	PGString      PGType = "varchar(1024)"
	PGBool        PGType = "bool"
	PGDate        PGType = "date"
	PGTime        PGType = "time"
	PGDatetime    PGType = "timestamp"
	PGUnknownType PGType = "unknown"
)

type KwilConstraint string
type PGConstraint string

const (
	KwilPrimaryKey        KwilConstraint = "primary_key"
	KwilNotNull           KwilConstraint = "not_null"
	KwilUnique            KwilConstraint = "unique"
	KwilUnknownConstraint KwilConstraint = "unknown"
)

const (
	PGPrimaryKey        PGConstraint = "primary key"
	PGNotNull           PGConstraint = "not null"
	PGUnique            PGConstraint = "unique"
	PGUnknownConstraint PGConstraint = "unknown"
)

type KwilIndex string
type PGIndex string

const (
	KwilBtree        KwilIndex = "btree"
	KwilUnknownIndex KwilIndex = "unknown"
)

const (
	PGBtree        PGIndex = "btree"
	PGUnknownIndex PGIndex = "unknown"
)

// Conversions for Types and Constraints
func (k KwilType) ToPG() PGType {
	switch k {
	case KwilNull:
		return PGNull
	case KwilInt32:
		return PGInt32
	case KwilInt64:
		return PGInt64
	case KwilString:
		return PGString
	case KwilBool:
		return PGBool
	case KwilDate:
		return PGDate
	case KwilTime:
		return PGTime
	case KwilDatetime:
		return PGDatetime
	default:
		return PGUnknownType
	}
}

func (p PGType) ToKwil() KwilType {
	switch p {
	case PGNull:
		return KwilNull
	case PGInt32:
		return KwilInt32
	case PGInt64:
		return KwilInt64
	case PGString:
		return KwilString
	case PGBool:
		return KwilBool
	case PGDate:
		return KwilDate
	case PGTime:
		return KwilTime
	case PGDatetime:
		return KwilDatetime
	default:
		return KwilUnknownType
	}
}

func (k KwilConstraint) ToPG() PGConstraint {
	switch k {
	case KwilPrimaryKey:
		return PGPrimaryKey
	case KwilNotNull:
		return PGNotNull
	case KwilUnique:
		return PGUnique
	default:
		return PGUnknownConstraint
	}
}

func (p PGConstraint) ToKwil() KwilConstraint {
	switch p {
	case PGPrimaryKey:
		return KwilPrimaryKey
	case PGNotNull:
		return KwilNotNull
	case PGUnique:
		return KwilUnique
	default:
		return KwilUnknownConstraint
	}
}

func (i KwilIndex) ToPG() PGIndex {
	switch i {
	case KwilBtree:
		return PGBtree
	default:
		return PGUnknownIndex
	}
}

func (i PGIndex) ToKwil() KwilIndex {
	switch i {
	case PGBtree:
		return KwilBtree
	default:
		return KwilUnknownIndex
	}
}

// Validity checks for Types and Constraints
func (t KwilType) Valid() bool {
	pgt := t.ToPG()
	return pgt != PGUnknownType
}

func (t PGType) Valid() bool {
	kwt := t.ToKwil()
	return kwt != KwilUnknownType
}

func (c KwilConstraint) Valid() bool {
	pgc := c.ToPG()
	return pgc != PGUnknownConstraint
}

func (c PGConstraint) Valid() bool {
	kwc := c.ToKwil()
	return kwc != KwilUnknownConstraint
}

// String methods
func (t KwilType) String() string {
	return string(t)
}

func (t PGType) String() string {
	return string(t)
}

func (c KwilConstraint) String() string {
	return string(c)
}

func (c PGConstraint) String() string {
	return string(c)
}

func (i KwilIndex) String() string {
	return string(i)
}

func (i PGIndex) String() string {
	return string(i)
}

// Interfaces for encompassing generics

type GType interface {
	KwilType | PGType
}

type GConstraint interface {
	KwilConstraint | PGConstraint
}

type GIndex interface {
	KwilIndex | PGIndex
}
