package execution

type DataType int

// Data Types
const (
	INVALID_DATA_TYPE DataType = iota
	STRING
	INT32
	INT64
	BOOLEAN
	END_DATA_TYPE
)

func (d *DataType) String() string {
	switch *d {
	case STRING:
		return `string`
	case INT32:
		return `int32`
	case INT64:
		return `int64`
	case BOOLEAN:
		return `boolean`
	}
	return `unknown`
}

func (d *DataType) Int() int {
	return int(*d)
}

func (d *DataType) IsNumeric() bool {
	return *d == INT32 || *d == INT64
}

func (d *DataType) IsValid() bool {
	return *d > INVALID_DATA_TYPE && *d < END_DATA_TYPE
}
