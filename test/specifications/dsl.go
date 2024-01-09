package specifications

import (
	"context"
	"math/big"

	"github.com/kwilteam/kwil-db/core/client"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/types/transactions"
)

// A Dsl describes a set of interactions that could achieve a specific goal
// Whoever writes a Dsl doesn't need to know what is the underlying implementation
// When in testing, need to translate the DSL to driver protocol

// DatabaseIdentifier
// It's questionable whether this should be part of the DSL
type DatabaseIdentifier interface {
	DBID(name string) string
}

type DatabaseExister interface {
	// DatabaseExists checks if a database exists, impl should check
	// two APIs: ListDatabases and GetSchema
	DatabaseExists(ctx context.Context, dbid string) error
}

// DatabaseDeployDsl is dsl for database deployment specification
// This dsl could be used to deploy a database
type DatabaseDeployDsl interface {
	DatabaseIdentifier
	DatabaseExister
	TxQueryDsl
	DeployDatabase(ctx context.Context, db *transactions.Schema) (txHash []byte, err error)
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
	TransferAmt(ctx context.Context, to []byte, amt *big.Int) (txHash []byte, err error)
}

// DatabaseDropDsl is dsl for database drop specification
type DatabaseDropDsl interface {
	TxQueryDsl
	DropDatabase(ctx context.Context, dbName string) (txHash []byte, err error)
	DatabaseIdentifier
	DatabaseExister
}

// ExecuteCallDsl is dsl for call specification
type ExecuteCallDsl interface {
	DatabaseIdentifier
	Call(ctx context.Context, dbid, action string, inputs []any) (*client.Records, error)
}

// ExecuteExtensionDsl is dsl for extension specification
type ExecuteExtensionDsl interface {
	DatabaseIdentifier
	TxQueryDsl
	ExecuteCallDsl
	ExecuteAction(ctx context.Context, dbid string, actionName string, actionInputs ...[]any) ([]byte, error)
}

// ExecuteQueryDsl is dsl for query specification
type ExecuteQueryDsl interface {
	DatabaseIdentifier
	TxQueryDsl
	// ExecuteAction executes QUERY to a database
	ExecuteAction(ctx context.Context, dbid string, actionName string, actionInputs ...[]any) ([]byte, error)
	QueryDatabase(ctx context.Context, dbid, query string) (*client.Records, error)
	SupportBatch() bool
}

// ExecuteActionsDsl is dsl for executing any sort of action
type ExecuteActionsDsl interface {
	ExecuteQueryDsl
	ExecuteCallDsl
}

// TxQueryDsl is dsl for tx query specification
type TxQueryDsl interface {
	TxSuccess(ctx context.Context, txHash []byte) error
}

// InfoDsl is a dsl for information about the chain and node, according
// to usage in the TxSvc
type InfoDsl interface {
	ChainInfo(ctx context.Context) (*types.ChainInfo, error)
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
	ValidatorNodeRemove(ctx context.Context, target []byte) ([]byte, error)
}

// ValidatorOpsDsl is a DSL for validator set updates specification such as
// join, leave, approve, etc. TODO: split this up?
type ValidatorOpsDsl interface {
	ValidatorStatusDsl
	ValidatorNodeApprove(ctx context.Context, joinerPubKey []byte) ([]byte, error)
	ValidatorNodeJoin(ctx context.Context) ([]byte, error)
	ValidatorNodeLeave(ctx context.Context) ([]byte, error)
}
