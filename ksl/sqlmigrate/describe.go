package sqlmigrate

import (
	"context"
	"ksl/sqlschema"
)

type Describer interface {
	Describe(string) (sqlschema.Database, error)
	DescribeContext(context.Context, string) (sqlschema.Database, error)
}
