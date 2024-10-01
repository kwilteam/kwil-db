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

type PeerRequest struct {
	PeerID string `json:"peerid"`
}

type ListPeersRequest struct{}

type CreateResolutionRequest struct {
	Resolution     []byte `json:"resolution"`
	ResolutionType string `json:"resolution_type"`
}

type ApproveResolutionRequest struct {
	ResolutionID *types.UUID `json:"resolution_id"` // Id is the resolution ID
}

// type DeleteResolutionRequest struct {
// 	ResolutionID *types.UUID `json:"resolution_id"` // Id is the resolution ID
// }

type ResolutionStatusRequest struct {
	ResolutionID *types.UUID `json:"resolution_id"` // Id is the resolution ID
}
