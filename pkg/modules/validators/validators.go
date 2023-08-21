// Package validators provides a module for processing validator requests from a
// blockchain application using a pluggable validator manager and account store.
package validators

import "github.com/kwilteam/kwil-db/pkg/log"

// NOTE: currently there is no pricing. Any fee is accepted (nonce update only)
// if their account has the sufficient balance.

// ValidatorModule separates validator update and state persistence details from
// the processing of validator related transactions (pricing and account updates
// i.e. "spending").
type ValidatorModule struct {
	mgr   ValidatorMgr
	accts Spender
	addr  Addresser

	log log.Logger
}

// NewValidatorModule constructs a validator module. The ValidatorMgr handles
// the details of computing validator updates to be included in a block, while
// the Spender provides handles account balance updates when processing the
// transactions.
func NewValidatorModule(mgr ValidatorMgr, accts Spender, addr Addresser, opts ...ValidatorModuleOpt) *ValidatorModule {
	d := &ValidatorModule{
		mgr:   mgr,
		accts: accts,
		addr:  addr,
	}

	for _, opt := range opts {
		opt(d)
	}

	return d
}

type ValidatorModuleOpt func(*ValidatorModule)

// WithLogger sets the logger
func WithLogger(logger log.Logger) ValidatorModuleOpt {
	return func(u *ValidatorModule) {
		u.log = logger
	}
}
