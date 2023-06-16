package dataset2

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

type ExecutionOpt func(*executionContext)

func WithCaller(caller string) ExecutionOpt {
	return func(ec *executionContext) {
		ec.caller = caller
	}
}

func WithAction(action string) ExecutionOpt {
	return func(ec *executionContext) {
		ec.action = action
	}
}

func WithDataset(dataset string) ExecutionOpt {
	return func(ec *executionContext) {
		ec.dataset = dataset
	}
}
