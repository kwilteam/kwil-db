package dataset3

import (
	"context"

	"github.com/kwilteam/kwil-db/pkg/engine/types"
)

func Builder() DatasetBuilder {
	return &datasetBuilder{
		avalableExtensions: map[string]Initializer{},
	}
}

type DatasetBuilder interface {
	WithProcedures(procedures ...*types.Procedure) DatasetBuilder
	WithTables(tables ...*types.Table) DatasetBuilder
	WithDatastore(datastore Datastore) DatasetBuilder
	WithExtensions(extensions map[string]Initializer) DatasetBuilder
	Build(context.Context) (*Dataset, error)
}

type datasetBuilder struct {
	procedures         []*types.Procedure
	tables             []*types.Table
	datastore          Datastore
	avalableExtensions map[string]Initializer
}

func (b *datasetBuilder) WithProcedures(procedures ...*types.Procedure) DatasetBuilder {
	b.procedures = procedures
	return b
}

func (b *datasetBuilder) WithTables(tables ...*types.Table) DatasetBuilder {
	b.tables = tables
	return b
}

func (b *datasetBuilder) WithDatastore(datastore Datastore) DatasetBuilder {
	b.datastore = datastore
	return b
}

func (b *datasetBuilder) WithExtensions(extensions map[string]Initializer) DatasetBuilder {
	b.avalableExtensions = extensions
	return b
}

func (b *datasetBuilder) Build(ctx context.Context) (*Dataset, error) {
	savepoint, err := b.datastore.Savepoint()
	if err != nil {
		return nil, err
	}
	defer savepoint.Rollback()

	for _, table := range b.tables {
		err = b.datastore.CreateTable(ctx, table)
		if err != nil {
			return nil, err
		}
	}

	for _, procedure := range b.procedures {
		err = b.datastore.StoreProcedure(ctx, procedure)
		if err != nil {
			return nil, err
		}
	}

	err = savepoint.Commit()
	if err != nil {
		return nil, err
	}

	return &Dataset{}, nil
}
