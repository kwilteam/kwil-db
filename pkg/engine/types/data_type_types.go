package types

type DataType int

// Data Types
const (
	INVALID_DATA_TYPE DataType = iota + 100
	NULL
	TEXT
	INT
	END_DATA_TYPE
)

func (d DataType) String() string {
	switch d {
	case NULL:
		return `null`
	case TEXT:
		return `text`
	case INT:
		return `int`
	}
	return `unknown`
}

func (d *DataType) Int() int {
	return int(*d)
}

func (d *DataType) IsNumeric() bool {
	return *d == INT
}

func (d *DataType) IsValid() bool {
	return *d > INVALID_DATA_TYPE && *d < END_DATA_TYPE
}

// will check if the data type is a text type
func (d *DataType) IsText() bool {
	return *d == TEXT
}
