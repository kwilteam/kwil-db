package dataset

import (
	"context"

	"github.com/kwilteam/kwil-db/pkg/engine/types"
	"github.com/kwilteam/kwil-db/pkg/log"
)

func Builder() DatasetBuilder {
	return &datasetBuilder{
		avalableExtensions: map[string]Initializer{},
		procedures:         []*types.Procedure{},
		tables:             []*types.Table{},
		extensions:         []*types.Extension{},
		log:                log.NewNoOp(),
	}
}

type DatasetBuilder interface {
	WithProcedures(procedures ...*types.Procedure) DatasetBuilder
	WithTables(tables ...*types.Table) DatasetBuilder
	WithDatastore(datastore Datastore) DatasetBuilder
	WithInitializers(extensions map[string]Initializer) DatasetBuilder
	WithExtensions(extensions ...*types.Extension) DatasetBuilder
	OwnedBy(owner string) DatasetBuilder
	Named(name string) DatasetBuilder
	WithLogger(l log.Logger) DatasetBuilder
	Build(context.Context) (*Dataset, error)
}

type datasetBuilder struct {
	procedures         []*types.Procedure
	tables             []*types.Table
	datastore          Datastore
	avalableExtensions map[string]Initializer
	extensions         []*types.Extension
	owner              string
	name               string
	log                log.Logger
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

func (b *datasetBuilder) WithInitializers(extensions map[string]Initializer) DatasetBuilder {
	b.avalableExtensions = extensions
	return b
}

func (b *datasetBuilder) WithExtensions(extensions ...*types.Extension) DatasetBuilder {
	b.extensions = extensions
	return b
}

func (b *datasetBuilder) OwnedBy(owner string) DatasetBuilder {
	b.owner = owner
	return b
}

func (b *datasetBuilder) Named(name string) DatasetBuilder {
	b.name = name
	return b
}

func (b *datasetBuilder) WithLogger(l log.Logger) DatasetBuilder {
	b.log = l
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

	for _, extension := range b.extensions {
		err = b.datastore.StoreExtension(ctx, extension)
		if err != nil {
			return nil, err
		}
	}

	err = savepoint.Commit()
	if err != nil {
		return nil, err
	}

	ds, err := OpenDataset(ctx, b.datastore,
		WithAvailableExtensions(b.avalableExtensions),
		OwnedBy(b.owner),
		Named(b.name),
		WithLogger(b.log),
	)
	if err != nil {
		return nil, err
	}

	err = ds.execConstructor(ctx, &TxOpts{
		Caller: b.owner,
	})
	if err != nil {
		return nil, err
	}

	return ds, nil
}
