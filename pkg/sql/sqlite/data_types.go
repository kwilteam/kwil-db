package sqlite

import "github.com/kwilteam/go-sqlite"

type DataType uint8

const (
	DataTypeNull DataType = iota
	DataTypeInteger
	DataTypeFloat
	DataTypeText
	DataTypeBlob
)

var (
	innerSqliteTypeMap = map[sqlite.ColumnType]DataType{
		sqlite.TypeInteger: DataTypeInteger,
		sqlite.TypeFloat:   DataTypeFloat,
		sqlite.TypeText:    DataTypeText,
		sqlite.TypeBlob:    DataTypeBlob,
		sqlite.TypeNull:    DataTypeNull,
	}
)
