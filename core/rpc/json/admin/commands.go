// Package adminjson defines the admin service's method names, request objects,
// and response objects.
package adminjson

import "github.com/kwilteam/kwil-db/core/types"

type StatusRequest struct{}
type PeersRequest struct{}
type GetConfigRequest struct{}
type ApproveRequest struct {
	PubKey []byte `json:"pubkey"`
}
type JoinRequest struct{}
type LeaveRequest struct{}
type RemoveRequest struct {
	PubKey []byte `json:"pubkey"`
}
type JoinStatusRequest struct {
	PubKey []byte `json:"pubkey"`
}
type ListValidatorsRequest struct{}
type ListJoinRequestsRequest struct{}

// LoadChangesetsRequest contains the request parameters for MethodLoadChangesets.
type ChangesetMetadataRequest struct {
	Height int64 `json:"height"`
}

type ChangesetRequest struct {
	Height int64 `json:"height"`
	Index  int64 `json:"index"`
}

type MigrationSnapshotChunkRequest struct {
	Height     uint64 `json:"height"`
	ChunkIndex uint32 `json:"chunk_index"`
}

type MigrationMetadataRequest struct{}

type TriggerMigrationRequest struct {
	Migration types.Migration `json:"migration"`
}

type ApproveMigrationRequest struct {
	Id string `json:"id"`
}

type ListMigrationsRequest struct{}
