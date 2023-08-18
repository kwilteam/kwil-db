package validators

import (
	"context"
	"fmt"
	"sync"

	"github.com/kwilteam/kwil-db/pkg/log"
	"github.com/kwilteam/kwil-db/pkg/sql"
)

type Datastore interface {
	Execute(ctx context.Context, stmt string, args map[string]any) error
	Query(ctx context.Context, query string, args map[string]any) ([]map[string]any, error)
	Prepare(stmt string) (sql.Statement, error)
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

	err := ar.initTables(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize tables: %w", err)
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

// ActiveVotes returns the currently ongoing join requests, which includes the
// candidate validator, the desired power, the set of validators who may
// approve the request, and if they did approve yet.
func (vs *validatorStore) ActiveVotes(ctx context.Context) ([]*JoinRequest, error) {
	vs.rw.RLock()
	defer vs.rw.RUnlock()

	return vs.allActiveJoinReqs(ctx)
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

// RemoveValidator deletes a validator. This is normally done in response to a
// leave request from the same validator. On the other hand, punishment involves
// reducing validator power. NOTE: This may not be required since power 0
// removes the validator from the "active" validators set returned by
// CurrentValidators.
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

func (vs *validatorStore) StartJoinRequest(ctx context.Context, joiner []byte, approvers [][]byte, power int64) error {
	vs.rw.Lock()
	defer vs.rw.Unlock()

	return vs.startJoinRequest(ctx, joiner, approvers, power)
}

// AddApproval records that a certain validator has approved the join request
// for a candidate validator.
func (vs *validatorStore) AddApproval(ctx context.Context, joiner, approver []byte) error {
	vs.rw.Lock()
	defer vs.rw.Unlock()

	return vs.addApproval(ctx, joiner, approver)
}
