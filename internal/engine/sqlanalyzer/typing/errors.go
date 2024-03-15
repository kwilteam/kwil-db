package typing

import (
	"fmt"
)

var (
	ErrInvalidType   = fmt.Errorf("invalid type")
	ErrCompoundShape = fmt.Errorf("compound shape mismatch")
)
