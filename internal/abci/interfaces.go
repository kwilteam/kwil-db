package abci

import (
	"context"
	"math/big"

	"github.com/kwilteam/kwil-db/core/types/transactions"
	modDataset "github.com/kwilteam/kwil-db/internal/modules/datasets"
	modVal "github.com/kwilteam/kwil-db/internal/modules/validators"

	"github.com/kwilteam/kwil-db/internal/abci/snapshots"
	"github.com/kwilteam/kwil-db/internal/accounts"
	"github.com/kwilteam/kwil-db/internal/engine/types"
	"github.com/kwilteam/kwil-db/internal/validators"
)

type DatasetsModule interface {
	Deploy(ctx context.Context, schema *types.Schema, tx *transactions.Transaction) (*modDataset.ExecutionResponse, error)
	Drop(ctx context.Context, dbid string, tx *transactions.Transaction) (*modDataset.ExecutionResponse, error)
	Execute(ctx context.Context, dbid string, action string, args [][]any, tx *transactions.Transaction) (*modDataset.ExecutionResponse, error)

	PriceDeploy(ctx context.Context, schema *types.Schema) (*big.Int, error)
	PriceDrop(ctx context.Context, dbid string) (*big.Int, error)
	PriceExecute(ctx context.Context, dbid string, action string, args [][]any) (*big.Int, error)
}

// ValidatorModule handles the processing of validator approve/join/leave
// transactions, punishment, preparation of validator updates to be applied when
// a block is finalized, and performing transaction accounting (e.g. fee and
// nonce checks).
//
// NOTE: this may be premature abstraction since we are designing this function
// for the needs for an abci/types.Application, yet using standard or Kwil
// types. But if a different blockchain package is used, this is unlikely to be
// what it needs, but it is as generic as possible.
type ValidatorModule interface {
	// GenesisInit configures the initial set of validators for the genesis
	// block. This is only called once for a new chain.
	GenesisInit(ctx context.Context, vals []*validators.Validator, blockHeight int64) error

	// CurrentSet returns the current validator set. This is used on app
	// construction to initialize the addr=>pubkey mapping.
	CurrentSet(ctx context.Context) ([]*validators.Validator, error)

	// Punish may be used at the start of block processing when byzantine
	// validators are listed by the consensus client (no transaction).
	Punish(ctx context.Context, validator []byte, power int64) error

	// Join creates a join request for a prospective validator.
	Join(ctx context.Context, power int64, tx *transactions.Transaction) (*modVal.ExecutionResponse, error)
	// Leave processes a leave request for a validator.
	Leave(ctx context.Context, tx *transactions.Transaction) (*modVal.ExecutionResponse, error)
	// Approve records an approval transaction from a current validator. The
	// approver is the tx Sender.
	Approve(ctx context.Context, joiner []byte, tx *transactions.Transaction) (*modVal.ExecutionResponse, error)
	// Remove removes a validator from the validator set, if the sender is a
	// current validator.
	Remove(ctx context.Context, validator []byte, tx *transactions.Transaction) (*modVal.ExecutionResponse, error)

	// Finalize is used at the end of block processing to retrieve the validator
	// updates to be provided to the consensus client for the next block. This
	// is not idempotent. The modules working list of updates is reset until
	// subsequent join/approves are processed for the next block.
	Finalize(ctx context.Context) ([]*validators.Validator, error) // end of block processing requires providing list of updates to the node's consensus client

	// Updates block height stored by the validator manager. Called in the abci Commit
	UpdateBlockHeight(ctx context.Context, blockHeight int64)

	// PriceJoin returns the price of a join transaction.
	PriceJoin(ctx context.Context) (*big.Int, error)

	// PriceApprove returns the price of an approve transaction.
	PriceApprove(ctx context.Context) (*big.Int, error)

	// PriceLeave returns the price of a leave transaction.
	PriceLeave(ctx context.Context) (*big.Int, error)

	// PriceRemove returns the price of a remove transaction.
	PriceRemove(ctx context.Context) (*big.Int, error)
}

// AtomicCommitter is an interface for a struct that implements atomic commits across multiple stores
type AtomicCommitter interface {
	Begin(ctx context.Context, idempotencyKey []byte) error
	Commit(ctx context.Context, idempotencyKey []byte) ([]byte, error)
}

// KVStore is an interface for a basic key-value store
type KVStore interface {
	Get(key []byte) ([]byte, error)
	Set(key []byte, value []byte) error
	Delete(key []byte) error
}

// SnapshotModule is an interface for a struct that implements snapshotting
type SnapshotModule interface {
	// Checks if databases are to be snapshotted at a particular height
	IsSnapshotDue(height uint64) bool

	// Starts the snapshotting process, Locking databases need to be handled outside this fn
	CreateSnapshot(height uint64) error

	// Lists all the available snapshots in the snapshotstore and returns the snapshot metadata
	ListSnapshots() ([]snapshots.Snapshot, error)

	// Returns the snapshot chunk of index chunkId at a given height
	LoadSnapshotChunk(height uint64, format uint32, chunkID uint32) []byte
}

// DBBootstrapModule is an interface for a struct that implements bootstrapping
type DBBootstrapModule interface {
	// Offers a snapshot (metadata) to the bootstrapper and decides whether to accept the snapshot or not
	OfferSnapshot(snapshot *snapshots.Snapshot) error

	// Offers a snapshot Chunk to the bootstrapper, once all the chunks corresponding to the snapshot are received, the databases are restored from the chunks
	ApplySnapshotChunk(chunk []byte, index uint32) ([]uint32, snapshots.Status, error)

	// Signifies the end of the db restoration
	IsDBRestored() bool
}

type AccountsModule interface {
	GetAccount(ctx context.Context, pubKey []byte) (*accounts.Account, error)
	Credit(ctx context.Context, addr string, acct []byte, amt *big.Int) error
	// Passing an opaque chain event would be awkward here I think.
}

// BridgeEventsModule is used to report deposit attestations and retrieve
// deposit events on the cross chain bridge contract. The only event is
// presently a deposit, and the events pertain to specific chain and contract.
//
// NOTE: the implementation will depend on an event store, which is continually
// updated with deposit events for contracts configured with the bridge client.
type BridgeEventsModule interface {
	// DepositsToReport returns witnessed deposit events that should be
	// broadcast to other validators via vote extensions.
	DepositsToReport() (eventID, amt, acct []string)

	// RecordDepositAttestation stores a validator's attestation to a deposit
	// event. i.e. we have received another validator's vote extension
	// referencing the event in VerifyVoteExtension or PrepareProposal, or we
	// have reported a deposit with ExtendVote.
	//
	// The amount and account are stored so that a proposer may author an
	// account credit transaction without themselves having observed the event.
	// Do we need the full event data to validate the eventID was correct for
	// the given amt and account?
	RecordDepositAttestation(eventID, amt, acct string, validator []byte)

	// DepositEvents is used when preparing a block to determine which events
	// should be acted upon (i.e. by creation of a governance transaction that
	// credits and account's balance). This give the application enough
	// information to decide if an event has sufficient attestation, and to
	// create a transaction to credit the account.
	DepositEvents() map[string]struct { // eventID => details
		Account string
		Amount  string
		// Attestations int // number of attestations including local
		Attestations [][]byte // the attester identities, including self
	}

	// MarkDepositActuated is used to mark a deposit as applied to an account
	// (when a relevant governance transaction referencing it is executed). This
	// may mean removing all entries for the event. NOTE: would this also be
	// used as the block proposer at the time the proposal is submitted rather
	// than later in execution?
	MarkDepositActuated(eventID string)

	// Address returns the address for an account public key. Observed on-chain
	// contract transaction events report a chain-specific address of the
	// depositor. This method is used so that Kwil accounts and thus transaction
	// senders, which are presently public keys, can be matched with deposit
	// events.
	Address(pubkey []byte) string

	// **** possibly useless methods below for consideration

	// RemoveDepositAttestation is used to remove a leaving/removed validator's
	// attestation to an event. NOTE: maybe we don't need this if the module
	// returns the source/observer identity of each event, which allows the
	// application to only count attestations for current validators.
	//RemoveDepositAttestation(eventID string, validator []byte)
}
