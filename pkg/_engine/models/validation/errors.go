package validation

import "fmt"

// all errorCodes should not be exported so we can see if they are being used (and therefore implemented)
// this isn't a perfect solution, but it at least helps to identify missing errorCodes

// violation returns a new error that annotates err with a errorCode number
func violation(errorCode errorCode, err error) error {
	return fmt.Errorf("errorCode %d: %s: %w", errorCode.Int, errorCode.String, err)
}

type errorCode struct {
	String string
	Int    int
}

// for creating new errorCodes
func ne(s string, i int) errorCode {
	return errorCode{
		String: s,
		Int:    i,
	}
}

// 0-99
var (
	errorCode0 errorCode = ne("database name must be valid", 0)
	errorCode1           = ne("database owner address must be valid", 1)
)

// 100-199
var (
	errorCode100 errorCode = ne("table names must be unique", 100)
	errorCode101           = ne("database must have at least one table", 101)
	errorCode102           = ne(fmt.Sprintf("cannot have more than %d tables", MAX_TABLE_COUNT), 102)
)

// 200-299
var (
	errorCode200 errorCode = ne("table name must be valid", 200)
	errorCode201           = ne("table name must not be a reserved keyword", 201)
)

// 300-399
var (
	errorCode300 errorCode = ne("column names must be unique", 300)
	errorCode301           = ne(fmt.Sprintf("cannot have more than %d columns", MAX_COLUMNS_PER_TABLE), 301)
)

// 400-499
var (
	errorCode400 errorCode = ne("column name must be valid", 400)
	errorCode401           = ne("column name must not be a reserved keyword", 401)
	errorCode402           = ne("column type must be valid", 402)
)

// 500-599
var (
	errorCode500 errorCode = ne("attribute types must be unique within the column", 500)
	errorCode501           = ne(fmt.Sprintf("cannot have more than %d attributes", MAX_ATTRIBUTES_PER_COLUMN), 501)
	errorCode502           = ne("cannot have unique and default attributes on the same column", 502)
)

// 600-699
var (
	errorCode600 errorCode = ne("attribute type must be valid", 600)
	errorCode601           = ne("attribute value must be valid for the attribute type", 601)
	errorCode602           = ne("default attribute value must be valid for the column type", 602)
	errorCode603           = ne("attribute must be applicable to column type", 603)
)

// 900-999
var (
	errorCode900 errorCode = ne("action names must be unique", 900)
)

// 1000-1099
var (
	errorCode1000 errorCode = ne("action inputs must begin with a $", 1000)
	errorCode1001           = ne("action inputs musdt be unique", 1001)
	errorCode1002           = ne("action inputs must only contain letters, numbers, and underscores", 1002)
	errorCode1003           = ne("action must have at least 1 statement", 1003)
)

// 1100-1199
var (
	errorCode1100 errorCode = ne("index names must be unique", 1100)
	errorCode1101           = ne(fmt.Sprintf("cannot have more than %d indexes", MAX_INDEX_COUNT), 1101)
	errorCode1102           = ne("cannot have two indexes of the same type on the same columns", 1102)
)

// 1200-1299
var (
	errorCode1200 errorCode = ne("index name must be valid", 1200)
	errorCode1201           = ne("index name must not be a reserved keyword", 1201)
	errorCode1202           = ne("index type must be valid", 1202)
	errorCode1204           = ne("index column(s) must be exist", 1204)
	errorCode1205           = ne("index column(s) must be unique", 1205)
	errorCode1206           = ne("index must have at least one column", 1206)
	errorCode1207           = ne(fmt.Sprintf("cannot have more than %d columns per index", MAX_INDEX_COLUMNS), 1207)
)
