package registry

import (
	"time"

	"github.com/kwilteam/kwil-db/core/log"
)

type RegistryOpt func(*Registry)

func WithLogger(l log.Logger) RegistryOpt {
	return func(r *Registry) {
		r.log = l
	}
}

// WithReaderWaitTimeout sets the timeout that the `Commit` function will wait for readers before
// forcing them to close.
// Default is 1000ms.
func WithReaderWaitTimeout(t time.Duration) RegistryOpt {
	return func(r *Registry) {
		r.readerCloseTime = t
	}
}

// WithFilesystem sets the filesystem that the registry will use.
// Default is the local filesystem.
func WithFilesystem(fs Filesystem) RegistryOpt {
	return func(r *Registry) {
		r.filesystem = fs
	}
}
