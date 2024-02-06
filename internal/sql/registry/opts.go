package registry

import (
	"github.com/kwilteam/kwil-db/core/log"
)

type RegistryOpt func(*Registry)

func WithLogger(l log.Logger) RegistryOpt {
	return func(r *Registry) {
		r.log = l
	}
}
