package validators

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"slices"

	"github.com/kwilteam/kwil-db/pkg/log"
	"github.com/kwilteam/kwil-db/pkg/sessions"
	sqlSessions "github.com/kwilteam/kwil-db/pkg/sessions/sql-session"
)

type joinReq struct {
	pubkey     []byte
	power      int64
	expiresAt  int64
	validators map[string]bool // pubkey bytes as string for map key
}

func (jr *joinReq) votes() int {
	var n int
	for _, a := range jr.validators {
		if a {
			n++
		}
	}
	return n
}

func threshold(numValidators int) int {
	return int(intDivUp(2*int64(numValidators), 3)) // float64(valSet.Count*2) / 3.
}

func (jr *joinReq) requiredVotes() int {
	return threshold(len(jr.validators))
}

// func (jr *joinReq) approval(approver []byte) (approved, eligible bool) {
// 	approved, eligible = jr.validators[string(approver)]
// 	return
// }

func (jr *joinReq) approve(approver []byte) (repeat, eligible bool) {
	key := string(approver) // coerce to string for map key
	repeat, eligible = jr.validators[key]
	if !eligible || repeat {
		return
	}
	jr.validators[key] = true
	return
}

// ValidatorMgr defines specific validator join/approve/leave mechanics for a
// federated network.
type ValidatorMgr struct {
	db ValidatorStore

	lastBlockHeight int64
	// state - these maps are keyed by pubkey, just coerced to string for the map
	current    map[string]struct{}
	candidates map[string]*joinReq
	updates    []*Validator // updates are built in BeginBlock/DeliverTx and cleared in EndBlock

	// opts
	joinExpiry int64
	log        log.Logger
}

// NOTE: The SQLite validator/approval store is local and transparent to the
// consumer of ValidatorMgr. Given the necessarily tight coupling, the data
// store is unexported to minimize our API, and a concrete instance is assembled
// in NewValidatorMgr. If we decide that this can be used outside of
// ValidatorMgr (and thus constructed by the caller) it should accept this:

type ValidatorStore interface {
	Init(ctx context.Context, vals []*Validator) error
	CurrentValidators(ctx context.Context) ([]*Validator, error)
	UpdateValidatorPower(ctx context.Context, validator []byte, power int64) error
	ActiveVotes(ctx context.Context) ([]*JoinRequest, error)
	StartJoinRequest(ctx context.Context, joiner []byte, approvers [][]byte, power int64, expiresAt int64) error
	DeleteJoinRequest(ctx context.Context, joiner []byte) error
	AddApproval(ctx context.Context, joiner, approver []byte) error
	AddValidator(ctx context.Context, joiner []byte, power int64) error
}

// Committable is a wrapper around the SQL committable that provides a
// deterministic ID based on the state of the validator manager.
type Committable struct {
	sessions.Committable

	dbHash func() []byte
}

// ID overrides the base Committable. We do this so that we can have the
// persistence of the state be part of the 2pc process, but have the ID reflect
// the actual state free from SQL specifics.
func (c *Committable) ID(ctx context.Context) ([]byte, error) {
	return c.dbHash(), nil
}

// ValidatorDB state includes:
// - current validators
// - active join requests
// - approvers for each join request
func (vm *ValidatorMgr) validatorDbHash() []byte {
	hasher := sha256.New()

	// current validators  val1:val2:...
	var currentValidators []string
	for val := range vm.current {
		currentValidators = append(currentValidators, val)
	}
	slices.Sort(currentValidators)
	for _, val := range currentValidators {
		hasher.Write([]byte(val))
	}

	// active join requests & approvals
	// joinerPubkey:power:expiresAt:approver1:approver2:...
	var joiners []string
	for val := range vm.candidates {
		joiners = append(joiners, val)
	}
	slices.Sort(joiners)

	for _, joiner := range joiners {
		jr := vm.candidates[joiner]

		hasher.Write([]byte(joiner))
		binary.Write(hasher, binary.LittleEndian, jr.power)
		binary.Write(hasher, binary.LittleEndian, jr.expiresAt)

		var approvers []string
		for val, approved := range jr.validators {
			if approved {
				approvers = append(approvers, val)
			}
		}
		slices.Sort(approvers)
		for _, approver := range approvers {
			hasher.Write([]byte(approver))
		}
	}

	return hasher.Sum(nil)
}

func (vm *ValidatorMgr) isCurrent(val []byte) bool {
	_, have := vm.current[string(val)]
	return have
}

func (vm *ValidatorMgr) candidate(val []byte) *joinReq {
	return vm.candidates[string(val)]
}

// WrapCommittable wraps a SQL committable with methods that provide a
// deterministic ID based on the state of the validator manager.
func (vm *ValidatorMgr) WrapCommittable(committable *sqlSessions.SqlCommitable) *Committable {
	return &Committable{
		Committable: committable,
		dbHash:      vm.validatorDbHash,
	}
}

func NewValidatorMgr(ctx context.Context, datastore Datastore, opts ...ValidatorMgrOpt) (*ValidatorMgr, error) {
	vm := &ValidatorMgr{
		current:    make(map[string]struct{}),
		candidates: make(map[string]*joinReq),
		log:        log.NewNoOp(),
		joinExpiry: 86400,
	}
	for _, opt := range opts {
		opt(vm)
	}

	var err error
	vm.db, err = newValidatorStore(ctx, datastore, vm.log)
	if err != nil {
		return nil, err
	}

	if err = vm.init(); err != nil {
		return nil, err
	}
	return vm, nil
}

func (vm *ValidatorMgr) init() error {
	// Restore state: current validators
	current, err := vm.db.CurrentValidators(context.Background())
	if err != nil {
		return err
	}

	for _, vi := range current {
		vm.current[string(vi.PubKey)] = struct{}{}
	}

	// Restore state: active join requests
	joinReqs, err := vm.db.ActiveVotes(context.Background())
	if err != nil {
		return err
	}

	for _, dbj := range joinReqs {
		jr := &joinReq{
			pubkey:     dbj.Candidate,
			power:      dbj.Power,
			validators: make(map[string]bool, len(dbj.Board)),
		}

		for i, vi := range dbj.Board {
			jr.validators[string(vi)] = dbj.Approved[i]
		}

		vm.candidates[string(dbj.Candidate)] = jr
	}

	return nil
}

// CurrentValidators returns the current validator set.
func (vm *ValidatorMgr) CurrentValidators(ctx context.Context) ([]*Validator, error) {
	// NOTE: the DB is the simplest approach, but since this method may be
	// called on-demand method (e.g. by an RPC client), it is not synchronized
	// with the other methods that are intended to be utilized by the blockchain
	// application. The ValidatorStore is thread-safe, but updates to the store
	// are not deferred until Finalize like the updates to the tracking fields
	// in this struct. Thus, the atomicity of the underlying datastore is the
	// only thing guaranteeing that this method reflects the current state.
	//
	// Alternatively, we can rig a mutex on the `current` map field, using that
	// throughout the ValidatorMgr methods. That's not attractive.
	return vm.db.CurrentValidators(ctx)
}

// ActiveVotes returns the current validator join requests.
func (vm *ValidatorMgr) ActiveVotes(ctx context.Context) ([]*JoinRequest, error) {
	return vm.db.ActiveVotes(ctx)
}

// GenesisInit is called at the genesis block to set and initial list of
// validators.
func (vm *ValidatorMgr) GenesisInit(ctx context.Context, vals []*Validator, blockHeight int64) error {
	// Initialize the current validator map.
	vm.current = make(map[string]struct{}, len(vals))
	for _, vi := range vals {
		vm.current[string(vi.PubKey)] = struct{}{}
	}
	vm.candidates = make(map[string]*joinReq)
	vm.updates = nil
	vm.lastBlockHeight = blockHeight

	vm.log.Warn("Resetting validator store with genesis validators.")

	// Wipe DB (!) and store the provided set.
	return vm.db.Init(ctx, vals)
}

// CurrentSet returns the current validator list. This may be used on
// construction of a resuming application.
func (vm *ValidatorMgr) CurrentSet(ctx context.Context) ([]*Validator, error) {
	return vm.db.CurrentValidators(ctx)
}

// Update may be used at the start of block processing when byzantine validators
// are listed by the consensus client, or to process a leave request.
func (vm *ValidatorMgr) Update(ctx context.Context, validator []byte, newPower int64) error {
	if !vm.isCurrent(validator) {
		return errors.New("not a current validator")
	}
	vm.updates = append(vm.updates, &Validator{
		PubKey: validator,
		Power:  newPower,
	})
	// delete(vm.current, ..) // in Finalize

	// Record new validator power.
	return vm.db.UpdateValidatorPower(ctx, validator, newPower)
}

// BIG Q: in all of these methods, if spend worked, do we have to return the
// execution response with a fee, and put any subsequent execution error in that
// struct for

// Join creates a join request for a prospective validator.
func (vm *ValidatorMgr) Join(ctx context.Context, joiner []byte, power int64) error {
	if vm.isCurrent(joiner) {
		return errors.New("already a validator")
	}

	if vm.candidate(joiner) != nil {
		// they tried to join again... but we executed the tx... no error?
		return nil
	}

	approvers := make([][]byte, 0, len(vm.current))
	valMap := make(map[string]bool, len(vm.current))
	for pk := range vm.current {
		valMap[pk] = false // eligible, but no vote yet
		approvers = append(approvers, []byte(pk))
	}
	expiresAt := vm.lastBlockHeight + vm.joinExpiry

	vm.candidates[string(joiner)] = &joinReq{
		pubkey:     joiner,
		power:      power,
		validators: valMap,
		expiresAt:  expiresAt,
	}

	return vm.db.StartJoinRequest(ctx, joiner, approvers, power, expiresAt)
}

// Leave processes a leave request for a current validator.
func (vm *ValidatorMgr) Leave(ctx context.Context, leaver []byte) error {
	// TODO: decide if leave should be a hard removal from the database or just
	// set power to zero. Punish does update even to zero power, so probably
	// Leave should too.

	const leavePower = 0 // leave the entry, set power to zero
	return vm.Update(ctx, leaver, leavePower)
	// return vm.db.RemoveValidator(ctx, leaver)
}

// Approve records an approval transaction from a current validator.
func (vm *ValidatorMgr) Approve(ctx context.Context, joiner, approver []byte) error {
	candidate := vm.candidate(joiner)
	if candidate == nil {
		return errors.New("not a validator candidate")
	}
	dup, eligible := candidate.approve(approver)
	if !eligible {
		return errors.New("approver is not on the validator board for the candidate")
	}
	if dup {
		vm.log.Info("already voted") // fine, but don't touch our state... or error?
	} else {
		// Record the vote. Check threshold in Finalize.
		if err := vm.db.AddApproval(ctx, joiner, approver); err != nil {
			return fmt.Errorf("failed to record approval: %v", err)
		}
	}

	return nil
}

// Finalize is used at the end of block processing to retrieve the validator
// updates to be provided to the consensus client for the next block. This is
// not idempotent. The modules working list of updates is reset until subsequent
// join/approves are processed for the next block. end of block processing
// requires providing list of updates to the node's consensus client
func (vm *ValidatorMgr) Finalize(ctx context.Context) []*Validator {
	// Updates for approved (joining) validators.
	for candidate, join := range vm.candidates {
		if join.votes() < join.requiredVotes() {
			if isJoinExpired(join, vm.lastBlockHeight) {
				// Join request expired
				delete(vm.candidates, candidate)
				if err := vm.db.DeleteJoinRequest(ctx, join.pubkey); err != nil {
					panic(fmt.Sprintf("failed to delete expired join request: %v", err))
				}
			}

			continue

		}

		// Candidate is above vote threshold
		delete(vm.candidates, candidate) // further votes are not recorded!

		if err := vm.db.AddValidator(ctx, join.pubkey, join.power); err != nil {
			panic(fmt.Sprintf("failed to record approval: %v", err)) // ugh
		}

		vm.current[candidate] = struct{}{} // == join.pubkey

		vm.updates = append(vm.updates, &Validator{
			PubKey: join.pubkey,
			Power:  join.power,
		})
	}

	updates := make([]*Validator, len(vm.updates))
	for i, up := range vm.updates {
		updates[i] = &Validator{
			PubKey: up.PubKey,
			Power:  up.Power,
		}
		pk := string(up.PubKey)
		if up.Power > 0 {
			vm.current[pk] = struct{}{} // add or overwrite
		} else {
			delete(vm.current, pk) // bye
		}
	}
	vm.updates = nil

	return updates
}

func (mgr *ValidatorMgr) UpdateBlockHeight(height int64) {
	mgr.lastBlockHeight = height
}

func isJoinExpired(join *joinReq, blockHeight int64) bool {
	return join.expiresAt != -1 && blockHeight >= join.expiresAt
}
