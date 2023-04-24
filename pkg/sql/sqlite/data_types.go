package sqlite

type DataType uint8

const (
	DataTypeNull DataType = iota
	DataTypeInteger
	DataTypeFloat
	DataTypeText
	DataTypeBlob
)
