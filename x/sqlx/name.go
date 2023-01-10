package spec

import (
	"fmt"
)

type NameType int

const (
	MAX_OWNER_NAME_LENGTH     = 44
	MAX_COLUMN_NAME_LENGTH    = 32
	MAX_TABLE_NAME_LENGTH     = 32
	MAX_INDEX_NAME_LENGTH     = 32
	MAX_ROLE_NAME_LENGTH      = 32
	MAX_COLUMNS_PER_TABLE     = 50
	MAX_DB_NAME_LENGTH        = 16
	MAX_SCHEMA_NAME_LENGTH    = 60
	MAX_QUERY_NAME_LENGTH     = 32
	MAX_ATTRIBUTE_NAME_LENGTH = 32
)

const (
	INVALID_NAME NameType = iota
	SCHEMA
	OWNER
	DATABASE
	ROLE
	TABLE
	COLUMN
	ATTRIBUTE
	INDEX
	QUERY
)

var (
	NamingParameters = namingParameters{}
)

type namingParameters struct{}

func (n *namingParameters) MaxLen(nameType NameType) int {
	switch nameType {
	case OWNER:
		return MAX_OWNER_NAME_LENGTH
	case DATABASE:
		return MAX_DB_NAME_LENGTH
	case ROLE:
		return MAX_ROLE_NAME_LENGTH
	case TABLE:
		return MAX_TABLE_NAME_LENGTH
	case COLUMN:
		return MAX_COLUMN_NAME_LENGTH
	case ATTRIBUTE:
		return MAX_ATTRIBUTE_NAME_LENGTH
	case INDEX:
		return MAX_INDEX_NAME_LENGTH
	case QUERY:
		return MAX_QUERY_NAME_LENGTH
	case SCHEMA:
		return MAX_SCHEMA_NAME_LENGTH
	}
	fmt.Println("unknown name type")
	return 0
}
