// Package validators provides a module for processing validator requests from a
// blockchain application using a pluggable validator manager and account store.
package validators

import (
	"context"

	"github.com/kwilteam/kwil-db/internal/validators"
)

// GenesisInit is called at the genesis block to set and initial list of
// validators.
func (vm *ValidatorModule) GenesisInit(ctx context.Context, vals []*validators.Validator, blockHeight int64) error {
	vm.log.Warn("Resetting validator store with genesis validators.")
	return vm.mgr.GenesisInit(ctx, vals, blockHeight)
}

// CurrentSet returns the current validator list. This may be used on
// construction of a resuming application.
func (vm *ValidatorModule) CurrentSet(ctx context.Context) ([]*validators.Validator, error) {
	return vm.mgr.CurrentSet(ctx)
}

// Punish may be used at the start of block processing when byzantine
// validators are listed by the consensus client.
func (vm *ValidatorModule) Punish(ctx context.Context, validator []byte, newPower int64) error {
	// Record new validator power.
	return vm.mgr.Update(ctx, validator, newPower)
}

// Finalize is used at the end of block processing to retrieve the validator
// updates to be provided to the consensus client for the next block. This is
// not idempotent. The modules working list of updates is reset until subsequent
// join/approves are processed for the next block. end of block processing
// requires providing list of updates to the node's consensus client
func (vm *ValidatorModule) Finalize(ctx context.Context) []*validators.Validator {
	return vm.mgr.Finalize(ctx)
}

// Updates block height stored by the validator manager. Called in the abci Commit
func (vm *ValidatorModule) UpdateBlockHeight(ctx context.Context, blockHeight int64) {
	vm.mgr.UpdateBlockHeight(blockHeight)
}
