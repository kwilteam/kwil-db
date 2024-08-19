// Package adminjson defines the admin service's method names, request objects,
// and response objects.
package adminjson

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
