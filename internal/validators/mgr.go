package validators

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"slices"

	"github.com/kwilteam/kwil-db/core/log"

	"go.uber.org/zap"
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

	// removals is a map of validators to the set of validators that have
	// proposed to remove them.
	removals map[string]map[string]bool

	// opts
	joinExpiry int64
	log        log.Logger

	// pricing
	feeMultiplier int64

	committable Committable
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
	ActiveVotes(ctx context.Context) ([]*JoinRequest, []*ValidatorRemoveProposal, error)
	StartJoinRequest(ctx context.Context, joiner []byte, approvers [][]byte, power int64, expiresAt int64) error
	DeleteJoinRequest(ctx context.Context, joiner []byte) error
	AddApproval(ctx context.Context, joiner, approver []byte) error
	AddRemoval(ctx context.Context, target, validator []byte) error
	DeleteRemoval(ctx context.Context, target, validator []byte) error
	AddValidator(ctx context.Context, joiner []byte, power int64) error
	RemoveValidator(ctx context.Context, validator []byte) error
	IsCurrent(ctx context.Context, validator []byte) (bool, error)
}

// ValidatorDB state includes:
// - current validators
// - active join requests
// - approvers for each join request
// - removals
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
func NewValidatorMgr(ctx context.Context, datastore Datastore, committable Committable, opts ...ValidatorMgrOpt) (*ValidatorMgr, error) {
	vm := &ValidatorMgr{
		current:     make(map[string]struct{}),
		candidates:  make(map[string]*joinReq),
		removals:    make(map[string]map[string]bool),
		log:         log.NewNoOp(),
		joinExpiry:  14400, // really should *always* come from opts in production to match consensus config
		committable: committable,
	}
	for _, opt := range opts {
		opt(vm)
	}
	vm.committable.SetIDFunc(func() ([]byte, error) {
		return vm.validatorDbHash(), nil
	})

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
	joinReqs, removals, err := vm.db.ActiveVotes(context.Background())
	if err != nil {
		return err
	}

	for _, dbj := range joinReqs {
		jr := &joinReq{
			pubkey:     dbj.Candidate,
			power:      dbj.Power,
			validators: make(map[string]bool, len(dbj.Board)),
			expiresAt:  dbj.ExpiresAt,
		}

		for i, vi := range dbj.Board {
			jr.validators[string(vi)] = dbj.Approved[i]
		}

		vm.candidates[string(dbj.Candidate)] = jr
	}

	for _, dbr := range removals {
		rKey, tKey := string(dbr.Remover), string(dbr.Target)
		if rems, ok := vm.removals[tKey]; ok {
			rems[rKey] = true
		} else {
			vm.removals[tKey] = map[string]bool{rKey: true}
		}
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

// ActiveVotes returns the current committed validator join requests.
func (vm *ValidatorMgr) ActiveVotes(ctx context.Context) ([]*JoinRequest, []*ValidatorRemoveProposal, error) {
	return vm.db.ActiveVotes(ctx)
}

// GenesisInit is called at the genesis block to set an initial list of
// validators.
func (vm *ValidatorMgr) GenesisInit(ctx context.Context, vals []*Validator, blockHeight int64) error {
	// Initialize the current validator map.
	vm.current = make(map[string]struct{}, len(vals))
	for _, vi := range vals {
		vm.current[string(vi.PubKey)] = struct{}{}
	}
	vm.candidates = make(map[string]*joinReq)
	vm.removals = make(map[string]map[string]bool)
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
	if vm.committable.Skip() {
		return nil
	}

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
	if vm.committable.Skip() {
		return nil
	}

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
	if vm.committable.Skip() {
		return nil
	}

	const leavePower = 0 // leave the entry, set power to zero
	return vm.Update(ctx, leaver, leavePower)
	// return vm.db.RemoveValidator(ctx, leaver)
}

// Approve records an approval transaction from a current validator.
func (vm *ValidatorMgr) Approve(ctx context.Context, joiner, approver []byte) error {
	if vm.committable.Skip() {
		return nil
	}

	candidate := vm.candidate(joiner)
	if candidate == nil {
		return errors.New("not a validator candidate")
	}
	dup, eligible := candidate.approve(approver)
	if !eligible {
		return errors.New("approver is not on the validator board for the candidate")
	}
	if dup { // I think would be better as an error so that tx exec fails
		vm.log.Info("already voted") // fine, but don't touch our state... or error? errors.New("already voted")
	} else {
		// Record the vote. Check threshold in Finalize.
		if err := vm.db.AddApproval(ctx, joiner, approver); err != nil {
			candidate.validators[string(approver)] = false
			return fmt.Errorf("failed to record approval: %v", err)
		}
	}

	return nil
}

// Remove stores a remove proposal. It should check that both the remover and
// target are both current validators. There should not already be a removal
// proposal recorded from this remover for this target. It should insert a row
// into the removals table. Finalize should do the removal counting and if the
// threshold is reached, the target validator should be removed and all the
// entries in the removals table for this target validator deleted.
func (vm *ValidatorMgr) Remove(ctx context.Context, target, remover []byte) error {
	if vm.committable.Skip() {
		return nil
	}

	if !vm.isCurrent(target) {
		return errors.New("target is not a current validator")
	}
	if !vm.isCurrent(remover) {
		return errors.New("remover is not a current validator")
	}

	// Check if we already have a removal proposal from this remover for this
	// targeted validator.
	tKey, rKey := string(target), string(remover)
	removals, have := vm.removals[tKey]
	if !have { // first!
		vm.removals[tKey] = map[string]bool{
			rKey: true,
		}
	} else if removals[rKey] {
		return errors.New("already proposed a removal")
	} else { // existing removal proposal
		removals[rKey] = true
	}

	return vm.db.AddRemoval(ctx, target, remover)
}

// Finalize is used at the end of block processing to retrieve the validator
// updates to be provided to the consensus client for the next block. This is
// not idempotent. The modules working list of updates is reset until subsequent
// join/approves are processed for the next block. end of block processing
// requires providing list of updates to the node's consensus client
func (vm *ValidatorMgr) Finalize(ctx context.Context) ([]*Validator, error) {
	if vm.committable.Skip() {
		return nil, nil
	}

	// Updates for approved (joining) validators.
	for candidate, join := range vm.candidates {
		if join.votes() < join.requiredVotes() {
			if isJoinExpired(join, vm.lastBlockHeight) {
				// Join request expired
				delete(vm.candidates, candidate)
				if err := vm.db.DeleteJoinRequest(ctx, join.pubkey); err != nil {
					return nil, fmt.Errorf("failed to delete expired join request: %v", err)
				}
			}

			continue

		}

		// Candidate is above vote threshold
		delete(vm.candidates, candidate) // further votes are not recorded!

		if err := vm.db.AddValidator(ctx, join.pubkey, join.power); err != nil {
			return nil, fmt.Errorf("failed to record approval: %v", err)
		}

		vm.current[candidate] = struct{}{} // == join.pubkey

		vm.updates = append(vm.updates, &Validator{
			PubKey: join.pubkey,
			Power:  join.power,
		})
	}

	// Updates for removals.
	for target, removals := range vm.removals {
		// Check if we have enough removals to remove the target. We compute the
		// threshold based on the size of the current validator set, and prune
		// the removals when validators are removed.
		if len(removals) < threshold(len(vm.current)) {
			continue
		}
		delete(vm.removals, target) // no further removals for this target
		targetPubKey := []byte(target)
		vm.log.Info("removing validator", zap.String("target", hex.EncodeToString(targetPubKey)))
		vm.updates = append(vm.updates, &Validator{
			PubKey: targetPubKey,
			Power:  0,
		})
		if err := vm.db.RemoveValidator(ctx, targetPubKey); err != nil {
			return nil, fmt.Errorf("failed to record removal: %v", err)
		}
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
		} else { // removed
			delete(vm.current, pk) // bye

			// Delete all removals targeting this now-gone validator.
			delete(vm.removals, pk)

			// Delete any removals coming from this validator.
			for target, r := range vm.removals {
				if r[pk] { // the removal from this validator no longer counts (after this block)
					delete(r, pk)
					if err := vm.db.DeleteRemoval(ctx, []byte(target), up.PubKey); err != nil {
						return nil, fmt.Errorf("failed to delete removal: %v", err)
					}
				}
			}
		}
	}
	vm.updates = nil

	return updates, nil
}

func (mgr *ValidatorMgr) UpdateBlockHeight(height int64) {
	mgr.lastBlockHeight = height
}

func isJoinExpired(join *joinReq, blockHeight int64) bool {
	return join.expiresAt != -1 && blockHeight >= join.expiresAt
}

// IsCurrent returns true if the given validator is in the current validator set.
// It reads directly from the DB, and does not consider pending updates.
func (mgr *ValidatorMgr) IsCurrent(ctx context.Context, validator []byte) (bool, error) {
	return mgr.db.IsCurrent(ctx, validator)
}
