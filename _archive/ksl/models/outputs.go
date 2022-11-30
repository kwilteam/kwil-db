package query

import (
	"database/sql"
	"fmt"
)

type Column struct {
	Name  string
	Value sql.NullString
	Type  string
}

type Row struct {
	Columns []Column
	length  uint8
}

type Result struct {
	Rows []Row
}

// Load takes a SQLResult and returns a slice of Outputs
func (r *Result) Load(result *sql.Rows) error {
	cts, err := result.ColumnTypes()
	if err != nil {
		return err
	}

	row, err := DefineRow(cts)
	if err != nil {
		return err
	}

	err = r.loadResult(result, row)
	if err != nil {
		return err
	}

	return nil
}

func (r *Result) loadResult(result *sql.Rows, row *Row) error {
	for result.Next() {
		nr := row.Copy()                                          // create a new empty row
		if err := result.Scan(nr.GetScannable()...); err != nil { // get the values
			return err
		}
		r.Rows = append(r.Rows, nr) // append the new row
	}
	return nil
}

// Define row creates a template row for the result set
func DefineRow(cols []*sql.ColumnType) (*Row, error) {
	var row Row
	for _, ct := range cols {
		var coltp string
		switch ct.DatabaseTypeName() {
		case "INT4": // int32
			coltp = "int32"
		case "INT8": // int64
			coltp = "int64"
		case "VARCHAR": // string
			coltp = "string"
		case "TEXT": // string
			coltp = "string"
		case "NVARCHAR": // string
			coltp = "string"
		case "TIMESTAMP": // datetime
			coltp = "datetime"
		case "DATE": // date
			coltp = "date"
		case "TIME": // time
			coltp = "time"
		case "BYTEA": // []byte
			coltp = "bytes"
		case "BOOL": // bool
			coltp = "bool"
		case "NUMERIC": // decimal
			coltp = "decimal"
		default:
			return nil, fmt.Errorf("unsupported type: %s", ct.DatabaseTypeName())
		}
		row.Columns = append(row.Columns, Column{Name: ct.Name(), Type: coltp, Value: sql.NullString{}}) // Append the Output to the slice of Outputs
	}

	row.length = uint8(len(row.Columns))

	return &row, nil
}

// copy creates a copy of the row.  This is used to copy the template for rows
func (r *Row) Copy() Row {
	var nr Row
	nr.Columns = r.Columns
	nr.length = r.length
	return nr
}

// GetScannable returns a slice of pointers to the value in each column.
func (r *Row) GetScannable() []interface{} {
	var scns []interface{}
	// iterate for len(row.Column)
	// I use this instead of range because range returns a copy of the value
	for i := uint8(0); i < r.length; i++ {
		scns = append(scns, &r.Columns[i].Value)
	}

	return scns
}
