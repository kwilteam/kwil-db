package dataset

type TxOpts struct {
	Caller string
}

type engineOptions struct {
	initializers map[string]Initializer
	owner        string
	name         string
}

type OpenOpt func(*engineOptions)

func WithAvailableExtensions(exts map[string]Initializer) OpenOpt {
	return func(opts *engineOptions) {
		opts.initializers = exts
	}
}

func OwnedBy(owner string) OpenOpt {
	return func(opts *engineOptions) {
		opts.owner = owner
	}
}

func Named(name string) OpenOpt {
	return func(opts *engineOptions) {
		opts.name = name
	}
}
