package typing

import (
	"fmt"
)

var (
	ErrInvalidType     = fmt.Errorf("invalid type")
	ErrCompoundShape   = fmt.Errorf("compound shape mismatch")
	errColumnNotFound  = fmt.Errorf("column not found")
	errAmbiguousColumn = fmt.Errorf("ambiguous column")
)
