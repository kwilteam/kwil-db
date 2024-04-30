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
	TypeCastInt  TypeCastType = "int"
	TypeCastText TypeCastType = "text"
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
