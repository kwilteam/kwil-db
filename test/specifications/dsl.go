package specifications

import (
	"context"

	"github.com/kwilteam/kwil-db/pkg/client"
	"github.com/kwilteam/kwil-db/pkg/transactions"
	"github.com/kwilteam/kwil-db/pkg/validators"
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
	Call(ctx context.Context, dbid, action string, inputs []any, withSignature bool) (*client.Records, error)
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

// ExecuteOwnerActionsDsl is dsl for owner actions specification
type ExecuteOwnerActionsDsl interface {
	ExecuteQueryDsl
	ExecuteCallDsl
}

// TxQueryDsl is dsl for tx query specification
type TxQueryDsl interface {
	TxSuccess(ctx context.Context, txHash []byte) error
}

// ValidatorOpsDsl is a DSL for validator set updates specification such as join, leave, approve, etc.
type ValidatorOpsDsl interface {
	TxQueryDsl
	ValidatorNodeApprove(ctx context.Context, joinerPubKey []byte) ([]byte, error)
	ValidatorNodeJoin(ctx context.Context) ([]byte, error)
	ValidatorNodeLeave(ctx context.Context) ([]byte, error)
	ValidatorJoinStatus(ctx context.Context, pubKey []byte) (*validators.JoinRequest, error)
	ValidatorsList(ctx context.Context) ([]*validators.Validator, error)
}
