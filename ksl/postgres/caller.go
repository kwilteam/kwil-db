package postgres

import (
	"context"
	"ksl/sqldriver"
	"ksl/sqlschema"
)

type Caller struct {
	Db   sqlschema.Database
	Conn sqldriver.ExecQuerier
}

func (c *Caller) Call(ctx context.Context, name string, args ...any) (any, error) {
	panic("not implemented")
}
