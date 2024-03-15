package tree

import "fmt"

/*
	'type cast' works as below:
	- it can directly be applied to literal, bind parameter, column, and function expression
    - case expression does not have type cast
	- other expressions can have type cast only when wrapped

    This logic is also enforced in the sqlparser.
*/

type TypeCastType string

const (
	TypeCastInt  TypeCastType = "INT"
	TypeCastText TypeCastType = "TEXT"
)

func (t TypeCastType) String() string {
	return string(t)
}

func (t TypeCastType) Valid() error {
	switch t {
	case TypeCastInt, TypeCastText:
		return nil
	default:
		return fmt.Errorf("invalid type cast: %s", t)
	}
}

// suffixTypeCast adds a type cast to `expression` if it is not empty
// NOTE: `::` is used to indicate type cast in SQL
func suffixTypeCast(expr string, typeCast TypeCastType) string {
	if err := typeCast.Valid(); err == nil {
		return fmt.Sprintf("%s ::%s", expr, typeCast)
	}
	return expr
}
