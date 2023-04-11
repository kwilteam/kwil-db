package validator

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

// 700-799
var (
	errorCode700 errorCode = ne("sql query names must be unique", 700)
	errorCode701           = ne(fmt.Sprintf("cannot have more than %d sql queries", MAX_QUERY_COUNT), 701)
)

// 800-899
var (
	errorCode800 errorCode = ne("sql query name must be valid", 800)
	errorCode801           = ne("sql query type must be valid", 801)
	errorCode802           = ne("table must exist for sql query", 802)
	errorCode803           = ne("insert and update sql queries must have at least one parameter", 803)
	errorCode804           = ne("update and delete sql queries must have at least one where clause", 804)
	errorCode805           = ne("insert sql queries cannot have where clauses", 805)
	errorCode806           = ne("delete sql queries cannot have parameters", 806)
	errorCode807           = ne("all not-null columns must have a parameter in insert queries", 807)
	errorCode808           = ne("name must not be keyword", 808)
)

// 900-999
var (
	errorCode900 errorCode = ne("parameter and where clause names must be unique within a sql query", 900)
	errorCode901           = ne("A column can only be used in one parameter per sql query", 901)
	errorCode902           = ne(fmt.Sprintf("cannot have more than %d parameters per sql query", MAX_PARAM_PER_QUERY), 902)
	errorCode903           = ne(fmt.Sprintf("cannot have more than %d where clauses per sql query", MAX_WHERE_PER_QUERY), 903)
)

// 1000-1099
var (
	errorCode1000 errorCode = ne("parameter and where clause name must be valid", 1000)
	errorCode1001           = ne("column must exist for parameter and where clause", 1001)
	errorCode1002           = ne("if not static, then default value must be empty", 1002)
	errorCode1003           = ne("if modifier is caller, then default value must be empty", 1003)
	errorCode1004           = ne("if modifier is caller, then parameter must be static", 1004)
	errorCode1005           = ne("if modifier is caller, then column must be a string", 1005)
	errorCode1006           = ne("if modifier is caller, then column must not have a minimum length of > 42", 1006)
	errorCode1007           = ne("if modifier is caller, then column must not have a maximum length of < 44", 1007)
	errorCode1008           = ne("operator type must be valid", 1008)
	errorCode1009           = ne("operator type must be valid for the column type", 1009)
	errorCode1010           = ne("modifier value must be valid", 1010)
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
	errorCode1203           = ne("index table must exist", 1203)
	errorCode1204           = ne("index column(s) must be exist", 1204)
	errorCode1205           = ne("index column(s) must be unique", 1205)
	errorCode1206           = ne("index must have at least one column", 1206)
	errorCode1207           = ne(fmt.Sprintf("cannot have more than %d columns per index", MAX_INDEX_COLUMNS), 1207)
)

// 1300-1399
var (
	errorCode1300 errorCode = ne("role names must be unique", 1300)
	errorCode1301           = ne(fmt.Sprintf("cannot have more than %d roles", MAX_ROLE_COUNT), 1301)
)

// 1400-1499
var (
	errorCode1400 errorCode = ne("role name must be valid", 1400)
	errorCode1401           = ne("role permissions must exist", 1401)
	errorCode1402           = ne("role permissions must be unique within the role", 1402)
	errorCode1403           = ne("role name must not be a reserved keyword", 1403)
)
