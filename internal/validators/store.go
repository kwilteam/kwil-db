package validators

import (
	"context"
	"fmt"
	"sync"

	"github.com/kwilteam/kwil-db/core/log"
)

type Datastore interface {
	Execute(ctx context.Context, stmt string, args map[string]any) ([]map[string]any, error)
	Query(ctx context.Context, query string, args map[string]any) ([]map[string]any, error)
}

// validatorStore provides persistent storage for validators and ongoing
// validator join votes/approval.
type validatorStore struct {
	// what's the caller's concurrency w.r.t. this store? is wrapping every method with this needed?
	rw sync.RWMutex

	db  Datastore
	log log.Logger
	// stmts *preparedStatements // may be useful, otherwise remove
}

// newValidatorStore constructs the validator storage with the provided
// SQL datastore.
func newValidatorStore(ctx context.Context, datastore Datastore, log log.Logger) (*validatorStore, error) {
	ar := &validatorStore{
		db:  datastore,
		log: log,
	}

	err := ar.initOrUpgradeDatabase(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database at version %d due to error: %w", valStoreVersion, err)
	}

	// err = ar.prepareStatements()
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to prepare statements: %w", err)
	// }

	return ar, nil
}

// CurrentValidators returns the current set of active validators.
func (vs *validatorStore) CurrentValidators(ctx context.Context) ([]*Validator, error) {
	vs.rw.RLock()
	defer vs.rw.RUnlock()

	return vs.currentValidators(ctx)
}

// ActiveVotes returns the currently ongoing join and removal requests. For join
// requests, this includes the candidate validator, the desired power, the set
// of validators who may approve the request, and if they did approve yet.
func (vs *validatorStore) ActiveVotes(ctx context.Context) ([]*JoinRequest, []*ValidatorRemoveProposal, error) {
	vs.rw.RLock()
	defer vs.rw.RUnlock()

	joins, err := vs.allActiveJoinReqs(ctx)
	if err != nil {
		return nil, nil, err
	}

	removals, err := vs.allActiveRemoveReqs(ctx)
	if err != nil {
		return nil, nil, err
	}

	return joins, removals, nil
}

// Init deletes all existing validator and join requests data, and inserts the
// provided validators as the initial set. This would be used at genesis.
func (vs *validatorStore) Init(ctx context.Context, vals []*Validator) error {
	vs.rw.Lock()
	defer vs.rw.Unlock()

	return vs.init(ctx, vals)
}

// UpdateValidatorPower modifies the power of an existing validator. This is
// typically done to punish a byzantine validator by reducing their power. This
// will not error if the validator does not exists, but it will not be added.
func (vs *validatorStore) UpdateValidatorPower(ctx context.Context, validator []byte, power int64) error {
	vs.rw.Lock()
	defer vs.rw.Unlock()

	// perhaps if power==0 we do removeValidator? depends if the row needs to
	// exist for other reasons, like if they should still be allowed to approve
	// join requests created when they were empowered. TODO: set the rules!

	return vs.updateValidatorPower(ctx, validator, power)
}

// RemoveValidator deletes a validator. NOTE: Both leave request handling and
// punishment involve reducing validator power, not removing them. As such, this
// may not be required since power 0 removes the validator from the "active"
// validators set returned by CurrentValidators. To support this, AddValidator
// performs an upsert operation to avoid a UNIQUE constraint violation if
// re-adding the validator later.
func (vs *validatorStore) RemoveValidator(ctx context.Context, validator []byte) error {
	vs.rw.Lock()
	defer vs.rw.Unlock()

	return vs.removeValidator(ctx, validator)
}

// AddValidator adds a new key to the validator set. This will also remove any
// join_requests that may have just finished.
func (vs *validatorStore) AddValidator(ctx context.Context, joiner []byte, power int64) error {
	vs.rw.Lock()
	defer vs.rw.Unlock()

	return vs.addValidator(ctx, joiner, power)
}

func (vs *validatorStore) StartJoinRequest(ctx context.Context, joiner []byte, approvers [][]byte, power int64, expiresAt int64) error {
	vs.rw.Lock()
	defer vs.rw.Unlock()

	return vs.startJoinRequest(ctx, joiner, approvers, power, expiresAt)
}

// AddApproval records that a certain validator has approved the join request
// for a candidate validator.
func (vs *validatorStore) AddApproval(ctx context.Context, joiner, approver []byte) error {
	vs.rw.Lock()
	defer vs.rw.Unlock()

	return vs.addApproval(ctx, joiner, approver)
}

// AddApproval records that a certain validator has requested removal of a
// validator from the validator set.
func (vs *validatorStore) AddRemoval(ctx context.Context, target, validator []byte) error {
	vs.rw.Lock()
	defer vs.rw.Unlock()

	return vs.addRemoval(ctx, target, validator)
}

// DeleteRemoval deletes a removal request. Note that when removing a validator
// with RemoveValidator, any and all removal requests for removed validator are
// deleted, so it is not necessary to use this method in that case.
func (vs *validatorStore) DeleteRemoval(ctx context.Context, target, validator []byte) error {
	vs.rw.Lock()
	defer vs.rw.Unlock()

	return vs.deleteRemoval(ctx, target, validator)
}

// Delete a join request
func (vs *validatorStore) DeleteJoinRequest(ctx context.Context, joiner []byte) error {
	vs.rw.Lock()
	defer vs.rw.Unlock()

	return vs.deleteJoinRequest(ctx, joiner)
}

// IsCurrent returns true if the validator is in the current validator set.
func (vs *validatorStore) IsCurrent(ctx context.Context, validator []byte) (bool, error) {
	vs.rw.RLock()
	defer vs.rw.RUnlock()

	power, err := vs.validatorPower(ctx, validator)
	if err == errUnknownValidator {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return power > 0, nil
}
