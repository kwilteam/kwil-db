package validators

import "github.com/kwilteam/kwil-db/pkg/log"

type ValidatorMgrOpt func(*ValidatorMgr)

// WithLogger sets the logger
func WithLogger(logger log.Logger) ValidatorMgrOpt {
	return func(v *ValidatorMgr) {
		v.log = logger
	}
}
