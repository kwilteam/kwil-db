package dataset

import "github.com/kwilteam/kwil-db/pkg/log"

type TxOpts struct {
	Caller string
}

type engineOptions struct {
	initializers map[string]Initializer
	owner        string
	name         string
	log          log.Logger
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

func WithLogger(logger log.Logger) OpenOpt {
	return func(opts *engineOptions) {
		opts.log = logger
	}
}
