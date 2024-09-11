package userjson

import (
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/types/transactions"
)

// This file defines the structured parameter types used in the "params" field
// of the request. Many of them build on or alias the types in core/types and
// core/types/transactions to avoid duplication and make conversion nearly
// transparent. Those types MUST remain json tagged. If the RPC API must diverge
// from the Go APIs that use those types, they can be cloned and versioned here.

// NOTE: Any field of type []byte will marshal to/from base64 strings. This is
// established by the convention of the encoding/json standard library packages.
// For values that we would like to marshal as hexadecimal, a type such as
// HexBytes should be used, or the field should be a string and converted by the
// application. For instance, "owner" would be friendlier as hexadecimal so that
// it can look like an address if the signature type permits it (e.g. secp256k1).

type VersionRequest struct{}

// SchemaRequest contains the request parameters for MethodSchema.
type SchemaRequest struct {
	DBID string `json:"dbid"`
}

// AccountRequest contains the request parameters for MethodAccount.
type AccountRequest struct {
	Identifier types.HexBytes `json:"identifier" desc:"account identifier"`
	Status     *AccountStatus `json:"status,omitempty" desc:"blockchain status (confirmed or unconfirmed)"` // Mapped to URL query parameter `status`.
}

// AccountStatus is the type used to enumerate the different account status
// options recognized in AccountRequest.
type AccountStatus = types.AccountStatus

// These are the recognized AccountStatus values used with AccountRequest.
// AccountStatusLatest reflects confirmed state, while AccountStatusPending
// includes changes in mempool.
var (
	AccountStatusLatest  = types.AccountStatusLatest
	AccountStatusPending = types.AccountStatusPending
)

// BroadcastRequest contains the request parameters for MethodBroadcast.
type BroadcastRequest struct {
	Tx   *transactions.Transaction `json:"tx"`
	Sync *BroadcastSync            `json:"sync,omitempty"`
}

// BroadcastSync is the type used to enumerate the broadcast request
// synchronization options available to BroadcastRequest.
type BroadcastSync uint8

// These are the recognized BroadcastSync values used with BroadcastRequest.
const (
	// BroadcastSyncAsync does not wait for acceptance into mempool, only
	// computing the transaction hash.
	BroadcastSyncAsync BroadcastSync = 0
	// BroadcastSyncSync ensures the transaction is accepted to mempool before
	// responding. Ths should be preferred to BroadcastSyncAsync in most cases.
	BroadcastSyncSync BroadcastSync = 1
	// BroadcastSyncCommit will wait for the transaction to be included in a
	// block.
	BroadcastSyncCommit BroadcastSync = 2
)

// CallRequest contains the request parameters for MethodCall.
type CallRequest = transactions.CallMessage

// ChainInfoRequest contains the request parameters for MethodChainInfo.
type ChainInfoRequest struct{}

// ListDatabasesRequest contains the request parameters for MethodDatabases.
type ListDatabasesRequest struct {
	Owner types.HexBytes `json:"owner,omitempty"`
}

// PingRequest contains the request parameters for MethodPing.
type PingRequest struct {
	Message string `json:"message"`
}

// EstimatePriceRequest contains the request parameters for MethodPrice.
type EstimatePriceRequest struct {
	Tx *transactions.Transaction `json:"tx"`
}

// QueryRequest contains the request parameters for MethodQuery.
type QueryRequest struct {
	DBID  string `json:"dbid"`
	Query string `json:"query"`
}

// TxQueryRequest contains the request parameters for MethodTxQuery.
type TxQueryRequest struct {
	TxHash types.HexBytes `json:"tx_hash"`
}

type HealthRequest struct{}
