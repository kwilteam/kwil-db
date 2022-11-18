package pdb

type IndexName struct {
	Model ModelID
	Name  string
}

type ConstraintNamespace struct {
	Global      map[GlobalConstraint]uint
	Local       map[LocalConstraint]uint
	LocalCustom map[LocalCustomConstraint]uint
}

type GlobalConstraint struct {
	Scope      ConstraintScope
	SchemaName string
	Name       string
}

type LocalConstraint struct {
	Model ModelID
	Scope ConstraintScope
	Name  string
}

type LocalCustomConstraint struct {
	Model ModelID
	Name  string
}

type ConstraintName interface {
	constraintname()
	String() string
}
type IndexConstraintName string

func (c IndexConstraintName) String() string { return string(c) }

type RelationConstraintName string

func (c RelationConstraintName) String() string { return string(c) }

type PrimaryKeyConstraintName string

func (c PrimaryKeyConstraintName) String() string { return string(c) }

type DefaultConstraintName string

func (c DefaultConstraintName) String() string { return string(c) }

func (IndexConstraintName) constraintname()      {}
func (RelationConstraintName) constraintname()   {}
func (PrimaryKeyConstraintName) constraintname() {}
func (DefaultConstraintName) constraintname()    {}

type ConstraintScope int

const (
	GlobalKeyIndex ConstraintScope = iota
	GlobalForeignKey
	GlobalPrimaryKeyKeyIndex
	GlobalPrimaryKeyForeignKeyDefault
	ModelKeyIndex
	ModelPrimaryKeyKeyIndex
	ModelPrimaryKeyKeyIndexForeignKey
)

func (c ConstraintScope) String() string {
	switch c {
	case GlobalKeyIndex:
		return "indexes and unique constraints globally"
	case GlobalForeignKey:
		return "foreign keys globally"
	case GlobalPrimaryKeyKeyIndex:
		return "primary key, indexes, and unique constraints globally"
	case GlobalPrimaryKeyForeignKeyDefault:
		return "primary keys, foreign keys, and default constraints globally"
	case ModelKeyIndex:
		return "indexes and unique constraints on a model"
	case ModelPrimaryKeyKeyIndex:
		return "primary key, indexes, and unique constraints on a model"
	case ModelPrimaryKeyKeyIndexForeignKey:
		return "primary key, indexes, unique constraints, and foreign keys on a model"
	default:
		panic("unreachable")
	}
}
