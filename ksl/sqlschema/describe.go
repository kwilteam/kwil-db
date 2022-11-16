package sqlschema

import (
	"context"
)

type Describer interface {
	Describe(string) (Database, error)
	DescribeContext(context.Context, string) (Database, error)
}
