package db

import (
	"context"

	"github.com/kwilteam/kwil-db/pkg/engine/types"
)

// StoreProcedure stores a procedure in the database
func (d *DB) StoreProcedure(ctx context.Context, procedure *types.Procedure) error {
	return serdes[types.Procedure]{
		db: d,
	}.persistSerializable(ctx, procedure)
}

// ListProcedures lists all procedures in the database
func (d *DB) ListProcedures(ctx context.Context) ([]*types.Procedure, error) {
	return serdes[types.Procedure]{db: d}.listDeserialized(ctx)
}
