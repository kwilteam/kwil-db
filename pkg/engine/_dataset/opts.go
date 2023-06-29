package dataset

type OpenOpt func(*Dataset)

func WithName(name string) OpenOpt {
	return func(ds *Dataset) {
		ds.name = name
	}
}

func WithOwner(owner string) OpenOpt {
	return func(ds *Dataset) {
		ds.owner = owner
	}
}

func WithExtensionInitializers(initializers map[string]Initializer) OpenOpt {
	return func(ds *Dataset) {
		ds.extensionInitializers = initializers
	}
}
