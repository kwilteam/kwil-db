package specifications

import (
	"context"
	"crypto/ecdsa"
	"math/big"

	"github.com/kwilteam/kwil-db/core/types"
)

// A Dsl describes a set of interactions that could achieve a specific goal
// Whoever writes a Dsl doesn't need to know what is the underlying implementation
// When in testing, need to translate the DSL to driver protocol

type DatabaseExister interface {
	// DatabaseExists checks if a database exists, impl should check
	// two APIs: ListDatabases and GetSchema
	DatabaseExists(ctx context.Context, dbid string) error
}

// AccountBalanceDsl is the dsl for checking an confirmed account balance. This
// is likely to be useful for most other specifications when gas is enabled.
type AccountBalanceDsl interface {
	AccountBalance(ctx context.Context, acctID []byte) (*big.Int, error)
}

// TransferAmountDsl is the dsl for the account-to-account transfer
// specification.
type TransferAmountDsl interface {
	TxQueryDsl
	AccountBalanceDsl
	TransferAmt(ctx context.Context, to []byte, amt *big.Int) (txHash types.Hash, err error)
}

// ExecuteCallDsl is dsl for call specification
type ExecuteCallDsl interface {
	Call(ctx context.Context, dbid, action string, inputs []any) (*types.CallResult, error)
}

// ExecuteExtensionDsl is dsl for extension specification
type ExecuteExtensionDsl interface {
	TxQueryDsl
	ExecuteCallDsl
	Execute(ctx context.Context, dbid string, actionName string, actionInputs ...[]any) (types.Hash, error)
}

// ExecuteQueryDsl is dsl for query specification
type ExecuteQueryDsl interface {
	TxQueryDsl
	// ExecuteAction executes QUERY to a database
	Execute(ctx context.Context, dbid string, actionName string, actionInputs ...[]any) (types.Hash, error)
	ExecuteSQL(ctx context.Context, sql string, params map[string]any) (types.Hash, error)
	QueryDatabase(ctx context.Context, query string) (*types.QueryResult, error)
	SupportBatch() bool
}

// ExecuteActionsDsl is dsl for executing any sort of action
type ExecuteActionsDsl interface {
	ExecuteQueryDsl
	ExecuteCallDsl
}

// TxQueryDsl is dsl for tx query specification
type TxQueryDsl interface {
	TxSuccess(ctx context.Context, txHash types.Hash) error
}

// InfoDsl is a dsl for information about the chain and node, according
// to usage in the TxSvc
type InfoDsl interface {
	ChainInfo(ctx context.Context) (*types.ChainInfo, error)
	// Signer returns the wallet's signer. This is the bytes address for eth signers.
	Signer() []byte
	// Identifier returns the identifier derived from the authenticator.
	// This is a hex address for eth signers.
	Identifier() (string, error)
}

// ValidatorStatusDsl is the dsl for checking validator status, including
// current validator set and active join requests.
type ValidatorStatusDsl interface {
	TxQueryDsl
	ValidatorJoinStatus(ctx context.Context, pubKey []byte) (*types.JoinRequest, error)
	ValidatorsList(ctx context.Context) ([]*types.Validator, error)
}

// ValidatorRemoveDsl is the dsl for the validator remove procedure.
type ValidatorRemoveDsl interface {
	ValidatorStatusDsl
	ValidatorNodeRemove(ctx context.Context, target []byte) (types.Hash, error)
}

// ValidatorOpsDsl is a DSL for validator set updates specification such as
// join, leave, approve, etc. TODO: split this up?
type ValidatorOpsDsl interface {
	ValidatorStatusDsl
	ValidatorNodeApprove(ctx context.Context, joinerPubKey []byte) (types.Hash, error)
	ValidatorNodeJoin(ctx context.Context) (types.Hash, error)
	ValidatorNodeLeave(ctx context.Context) (types.Hash, error)
}

type DeployerDsl interface {
	AccountBalanceDsl
	TransferAmountDsl

	Approve(ctx context.Context, sender *ecdsa.PrivateKey, amount *big.Int) error
	Deposit(ctx context.Context, sender *ecdsa.PrivateKey, amount *big.Int) error
	EscrowBalance(ctx context.Context, sender *ecdsa.PrivateKey) (*big.Int, error)
	UserBalance(ctx context.Context, sender *ecdsa.PrivateKey) (*big.Int, error)
	Allowance(ctx context.Context, sender *ecdsa.PrivateKey) (*big.Int, error)
}

type TxInfoer interface {
	TxInfo(ctx context.Context, hash types.Hash) (*types.TxQueryResponse, error)
}

type PeersDsl interface {
	ListPeers(ctx context.Context) ([]string, error)
	AddPeer(ctx context.Context, peerID string) error
	RemovePeer(ctx context.Context, peerID string) error

	ConnectedPeers(ctx context.Context) ([]string, error)
}

// type ResolutionDsl interface {
// 	CreateResolution(ctx context.Context, resolutionType string, resolutionData []byte) ([]byte, error)
// 	ApproveResolution(ctx context.Context, resolutionID string) ([]byte, error)
// 	DeleteResolution(ctx context.Context, resolutionID string) ([]byte, error)
// }

type MigrationOpsDsl interface {
	TxQueryDsl
	SubmitMigrationProposal(ctx context.Context, activationHeight *big.Int, migrationDuration *big.Int) (types.Hash, error)
	ApproveMigration(ctx context.Context, migrationResolutionID *types.UUID) (types.Hash, error)
	ListMigrations(ctx context.Context) ([]*types.Migration, error)
	GenesisState(ctx context.Context) (*types.MigrationMetadata, error)
	GenesisSnapshotChunk(ctx context.Context, height uint64, chunkIdx uint32) ([]byte, error)
}
