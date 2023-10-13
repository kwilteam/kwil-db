package types

import (
	"fmt"
	"strings"
)

type DataType string

// Data Types
const (
	NULL DataType = "NULL"
	TEXT DataType = "TEXT"
	INT  DataType = "INT"
	BLOB DataType = "BLOB"
)

func (d DataType) String() string {
	return string(d)
}

func (d *DataType) IsNumeric() bool {
	return *d == INT
}

func (d *DataType) IsValid() bool {
	upper := strings.ToUpper(d.String())

	return upper == NULL.String() ||
		upper == TEXT.String() ||
		upper == INT.String() ||
		upper == BLOB.String()

}

// will check if the data type is a text type
func (d *DataType) IsText() bool {
	return *d == TEXT
}

func (d *DataType) Clean() error {
	if !d.IsValid() {
		return fmt.Errorf("invalid data type: %s", d.String())
	}

	*d = DataType(strings.ToUpper(d.String()))

	return nil
}
