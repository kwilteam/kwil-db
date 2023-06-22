package dataset2

/*
func Builder() DatasetBuilder {
	return &datasetBuilder{
		tables:                map[string]*dto.Table{},
		procedures:            map[string]*Procedure{},
		extensionInitializers: map[string]Initializer{},
	}
}

type datasetBuilder struct {
	name                  string
	owner                 string
	tables                map[string]*dto.Table
	procedures            map[string]*Procedure
	extensionInitializers map[string]Initializer
	onDeploy              []*OpCodeExecution
	onLoad                []*OpCodeExecution
	db                    Datastore
	errs                  []error
}

// DatasetBuilder is a builder for new datasets.
type DatasetBuilder interface {
	Named(string) DatasetBuilder
	OwnedBy(string) DatasetBuilder
	WithTables(...*dto.Table) DatasetBuilder
	WithProcedures(...*Procedure) DatasetBuilder
	WithExtensionInitializers(map[string]Initializer) DatasetBuilder
	WithDatastore(Datastore) DatasetBuilder
	Build(context.Context) (*Dataset, error)
}

func (b *datasetBuilder) Named(name string) DatasetBuilder {
	b.name = name
	return b
}

func (b *datasetBuilder) OwnedBy(owner string) DatasetBuilder {
	b.owner = owner
	return b
}

func (b *datasetBuilder) WithTables(tables ...*dto.Table) DatasetBuilder {
	for _, table := range tables {
		lowerName := strings.ToLower(table.Name)
		if _, ok := b.tables[lowerName]; ok {
			b.errs = append(b.errs, fmt.Errorf("table %s already exists", lowerName))
			continue
		}

		b.tables[lowerName] = table
	}
	return b
}

func (b *datasetBuilder) WithProcedures(actions ...*Procedure) DatasetBuilder {
	for _, procedure := range actions {
		lowerName := strings.ToLower(procedure.Name)
		if _, ok := b.procedures[lowerName]; ok {
			b.errs = append(b.errs, fmt.Errorf("procedure %s already exists", lowerName))
			continue
		}

		b.procedures[lowerName] = procedure
	}
	return b
}

func (b *datasetBuilder) WithDatastore(db Datastore) DatasetBuilder {
	b.db = db
	return b
}

func (b *datasetBuilder) WithExtensionInitializers(initializers map[string]Initializer) DatasetBuilder {
	b.extensionInitializers = initializers
	return b
}

func (b *datasetBuilder) Build(ctx context.Context) (*Dataset, error) {
	if err := b.validate(); err != nil {
		return nil, err
	}

	err := b.persistMetadata(ctx)
	if err != nil {
		return nil, err
	}

	return OpenDataset(ctx, b.db,
		WithName(b.name),
		WithOwner(b.owner),
		WithExtensionInitializers(b.extensionInitializers),
	)
}

func (b *datasetBuilder) persistMetadata(ctx context.Context) error {
	sp, err := b.db.Savepoint()
	if err != nil {
		return fmt.Errorf("failed to create savepoint: %w", err)
	}
	defer sp.Rollback()

	err = b.persistTables(ctx)
	if err != nil {
		return fmt.Errorf("failed to persist tables: %w", err)
	}

	err = b.persistProcedures(ctx)
	if err != nil {
		return fmt.Errorf("failed to persist actions: %w", err)
	}

	err = b.persistExtensions(ctx)
	if err != nil {
		return fmt.Errorf("failed to persist extensions: %w", err)
	}

	err = sp.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit savepoint: %w", err)
	}

	return nil
}

func (b *datasetBuilder) persistTables(ctx context.Context) error {
	for _, table := range b.tables {
		err := table.Clean()
		if err != nil {
			return fmt.Errorf("failed to clean table %s: %w", table.Name, err)
		}

		err = b.db.CreateTable(ctx, table)
		if err != nil {
			return fmt.Errorf("failed to create table %s: %w", table.Name, err)
		}
	}

	return nil
}

func (b *datasetBuilder) persistProcedures(ctx context.Context) error {
	for _, procedure := range b.procedures {
		err := procedure.Clean()
		if err != nil {
			return fmt.Errorf("failed to clean procedure %s: %w", procedure.Name, err)
		}

		err = b.db.StoreProcedure(ctx, procedure)
		if err != nil {
			return fmt.Errorf("failed to create procedure %s: %w", procedure.Name, err)
		}
	}

	return nil
}

func (b *datasetBuilder) persistExtensions(ctx context.Context) error {
	for name, metadata := range b.extensionMetadata {
		dtoExtention := &dto.ExtensionInitialization{
			Name:     name,
			Metadata: metadata,
		}

		err := dtoExtention.Clean()
		if err != nil {
			return fmt.Errorf("failed to clean extension %s: %w", name, err)
		}

		err = b.db.StoreExtension(ctx, dtoExtention)
		if err != nil {
			return fmt.Errorf("failed to create extension %s: %w", name, err)
		}
	}

	return nil
}

func (b *datasetBuilder) validate() error {
	if len(b.errs) > 0 {
		return fmt.Errorf("failed to build dataset: %w", b.errs[0])
	}

	if b.name == "" {
		return fmt.Errorf("failed to build dataset: name is required")
	}

	if b.owner == "" {
		return fmt.Errorf("failed to build dataset: owner is required")
	}

	if b.db == nil {
		return fmt.Errorf("failed to build dataset: datastore is required")
	}

	if len(b.tables) == 0 {
		return fmt.Errorf("failed to build dataset: at least one table is required")
	}

	for extName := range b.extensionMetadata {
		if _, ok := b.extensionInitializers[extName]; !ok {
			return fmt.Errorf("failed to build dataset: extension %s is not registered on node", extName)
		}
	}

	return nil
}
*/
