package userjson

import (
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/types/transactions"
)

// This file defines the response types. There is one for each request type.
// As with the request types, many are based on types in the core module.

type VersionResponse struct {
	Service     string `json:"service"`
	Version     string `json:"api_ver"`
	Major       uint32 `json:"major"`
	Minor       uint32 `json:"minor"`
	Patch       uint32 `json:"patch"`
	KwilVersion string `json:"kwil_ver"`
}

// AccountResponse contains the response object for MethodAccount.
type AccountResponse struct {
	Identifier types.HexBytes `json:"identifier,omitempty"`
	Balance    string         `json:"balance"`
	Nonce      int64          `json:"nonce"`
}

// BroadcastResponse contains the response object for MethodBroadcast.
type BroadcastResponse struct {
	TxHash types.HexBytes `json:"tx_hash,omitempty"`
}

type Result struct { // for other types, but embedding it is kinda annoying when instantiating
	Result []byte `json:"result,omitempty"`
}

// CallResponse contains the response object for MethodCall.
type CallResponse Result

// QueryResponse contains the response object for MethodQuery.
type QueryResponse Result

// ChainInfoResponse contains the response object for MethodChainInfo.
type ChainInfoResponse = types.ChainInfo

// ServiceMode describes the operating mode of the user service. Namely, if the
// service is in private mode (where calls are authenticated, query is disabled,
// and raw transactions cannot be retrieved).
type ServiceMode string

const (
	ModeOpen    ServiceMode = "open"
	ModePrivate ServiceMode = "private"
)

// HealthResponse is the response for MethodHealth. This determines the
// serialized response for the Health method required by the rpcserver.Svc
// interface. This is the response with which most health checks will be concerned.
type HealthResponse struct {
	// Healthy is is based on several factors determined by the service and it's
	// configuration, such as the maximum age of the best block and if the node
	// is still syncing (in catch-up or replay).
	Healthy bool `json:"healthy"`

	// Version is the service API version.
	Version string `json:"version"`

	ChainInfoResponse
	BlockTimestamp int64          `json:"block_time"` // epoch millis
	BlockAge       int64          `json:"block_age"`  // milliseconds
	Syncing        bool           `json:"syncing"`
	AppHeight      int64          `json:"app_height"` // may be less than block store best block
	AppHash        types.HexBytes `json:"app_hash"`
	PeerCount      int            `json:"peer_count"`

	// Mode is an oddball field as it pertains to the service config rather than
	// state of the node. It is provided here as a convenience so applications
	// can discern node state and the mode of interaction with one request.
	Mode ServiceMode `json:"mode"` // e.g. "private"
}

// SchemaResponse contains the response object for MethodSchema.
type SchemaResponse struct {
	Schema *types.Schema `json:"schema,omitempty"`
}

// SchemaResponse contains the response object for MethodSchema.
type ListDatabasesResponse struct {
	Databases []*DatasetInfo `json:"databases,omitempty"`
}

// SchemaResponse contains the response object for MethodSchema.
type DatasetInfo = types.DatasetIdentifier

// SchemaResponse contains the response object for MethodSchema.
type PingResponse struct {
	Message string `json:"message,omitempty"`
}

// SchemaResponse contains the response object for MethodSchema.
type EstimatePriceResponse struct {
	Price string `json:"price,omitempty"`
}

// TxQueryResponse contains the response object for MethodTxQuery.
type TxQueryResponse struct { // transactions.TcTxQueryResponse but pointers
	Hash     types.HexBytes                  `json:"hash,omitempty"`
	Height   int64                           `json:"height,omitempty"`
	Tx       *transactions.Transaction       `json:"tx,omitempty"`
	TxResult *transactions.TransactionResult `json:"tx_result,omitempty"`
}
