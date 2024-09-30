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
type CallResponse struct {
	Result []byte   `json:"result,omitempty"`
	Logs   []string `json:"logs,omitempty"`
}

// QueryResponse contains the response object for MethodQuery.
type QueryResponse Result

// ChainInfoResponse contains the response object for MethodChainInfo.
type ChainInfoResponse = types.ChainInfo

// HealthResponse is the response for MethodHealth. This determines the
// serialized response for the Health method required by the rpcserver.Svc
// interface. This is the response with which most health checks will be concerned.
type HealthResponse = types.Health

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

type ChangesetsResponse struct {
	Changesets []byte `json:"changesets"`
}

type ChangesetMetadataResponse struct {
	Height     int64   `json:"height"`
	Changesets int64   `json:"changesets"`
	ChunkSizes []int64 `json:"chunk_sizes"`
}

type MigrationMetadataResponse struct {
	Metadata *types.MigrationMetadata `json:"metadata"`
}

type MigrationSnapshotChunkResponse struct {
	Chunk []byte `json:"chunk"`
}

type ListMigrationsResponse struct {
	Migrations []*types.Migration `json:"migrations"`
}

type MigrationStatusResponse struct {
	Status *types.MigrationState `json:"status"`
}

type ChallengeResponse struct {
	Challenge types.HexBytes `json:"challenge"`
}
