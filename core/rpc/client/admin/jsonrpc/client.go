package adminclient

import (
	"context"
	"net/url"
	"time"

	rpcclient "github.com/kwilteam/kwil-db/core/rpc/client"
	"github.com/kwilteam/kwil-db/core/rpc/client/admin"
	"github.com/kwilteam/kwil-db/core/rpc/client/user"
	userClient "github.com/kwilteam/kwil-db/core/rpc/client/user/jsonrpc"
	jsonrpc "github.com/kwilteam/kwil-db/core/rpc/json"
	adminjson "github.com/kwilteam/kwil-db/core/rpc/json/admin"
	"github.com/kwilteam/kwil-db/core/types"
	adminTypes "github.com/kwilteam/kwil-db/core/types/admin"
)

// Client is an admin RPC client. It provides all methods of the user RPC
// service, plus methods that are specific to the admin service.
type Client struct {
	*userClient.Client // expose all user service methods, and CallMethod for admin svc
}

// NewClient constructs a new admin Client.
func NewClient(u *url.URL, opts ...rpcclient.RPCClientOpts) *Client {
	// alt: jsonclient.NewBaseClient() ... WrapBaseClient() ...
	userClient := userClient.NewClient(u, opts...)
	return WrapUserClient(userClient)
}

// WrapUserClient can be used to construct a new admin Client from an existing
// user RPC client.
func WrapUserClient(cl *userClient.Client) *Client {
	return &Client{
		Client: cl,
	}
}

var _ user.TxSvcClient = (*Client)(nil)  // via embedded userClient.Client
var _ admin.AdminClient = (*Client)(nil) // with extra methods

// Approve approves a validator join request for the validator identified by a
// public key. The transaction hash for the broadcasted approval transaction is
// returned.
func (cl *Client) Approve(ctx context.Context, publicKey []byte) ([]byte, error) {
	cmd := &adminjson.ApproveRequest{
		PubKey: publicKey,
	}
	res := &jsonrpc.BroadcastResponse{}
	err := cl.CallMethod(ctx, string(adminjson.MethodValApprove), cmd, res)
	if err != nil {
		return nil, err
	}
	return res.TxHash, err
}

// Join makes a validator join request for the node being administered. The
// transaction hash for the broadcasted join transaction is returned.
func (cl *Client) Join(ctx context.Context) ([]byte, error) {
	cmd := &adminjson.JoinRequest{}
	res := &jsonrpc.BroadcastResponse{}
	err := cl.CallMethod(ctx, string(adminjson.MethodValJoin), cmd, res)
	if err != nil {
		return nil, err
	}
	return res.TxHash, err
}

// JoinStatus returns the status of an active join request for the validator
// identified by the public key.
func (cl *Client) JoinStatus(ctx context.Context, pubkey []byte) (*types.JoinRequest, error) {
	cmd := &adminjson.JoinStatusRequest{
		PubKey: pubkey,
	}
	res := &adminjson.JoinStatusResponse{}
	err := cl.CallMethod(ctx, string(adminjson.MethodValJoinStatus), cmd, res)
	if err != nil {
		return nil, err
	}
	return res.JoinRequest, nil
}

// Leave makes a validator leave request for the node being administered. The
// transaction hash for the broadcasted leave transaction is returned.
func (cl *Client) Leave(ctx context.Context) ([]byte, error) {
	cmd := &adminjson.LeaveRequest{}
	res := &jsonrpc.BroadcastResponse{}
	err := cl.CallMethod(ctx, string(adminjson.MethodValLeave), cmd, res)
	if err != nil {
		return nil, err
	}
	return res.TxHash, err
}

// ListValidators gets the current validator set.
func (cl *Client) ListValidators(ctx context.Context) ([]*types.Validator, error) {
	cmd := &adminjson.ListValidatorsRequest{}
	res := &adminjson.ListValidatorsResponse{}
	err := cl.CallMethod(ctx, string(adminjson.MethodValList), cmd, res)
	if err != nil {
		return nil, err
	}
	return res.Validators, err
}

// Peers lists the nodes current peers (p2p node connections).
func (cl *Client) Peers(ctx context.Context) ([]*adminTypes.PeerInfo, error) {
	cmd := &adminjson.PeersRequest{}
	res := &adminjson.PeersResponse{}
	err := cl.CallMethod(ctx, string(adminjson.MethodPeers), cmd, res)
	if err != nil {
		return nil, err
	}
	return res.Peers, err
}

// Remove votes to remove the validator specified by the given public key.
func (cl *Client) Remove(ctx context.Context, publicKey []byte) ([]byte, error) {
	cmd := &adminjson.RemoveRequest{
		PubKey: publicKey,
	}
	res := &jsonrpc.BroadcastResponse{}
	err := cl.CallMethod(ctx, string(adminjson.MethodValRemove), cmd, res)
	if err != nil {
		return nil, err
	}
	return res.TxHash, err
}

// Status gets the node's status, such as it's name, chain ID, versions, sync
// status, best block info, and validator identity.
func (cl *Client) Status(ctx context.Context) (*adminTypes.Status, error) {
	cmd := &adminjson.StatusRequest{}
	res := &adminjson.StatusResponse{}
	err := cl.CallMethod(ctx, string(adminjson.MethodStatus), cmd, res)
	if err != nil {
		return nil, err
	}
	// TODO: convert!
	return &adminTypes.Status{
		Node: res.Node,
		Sync: &adminTypes.SyncInfo{
			AppHash:         res.Sync.AppHash,
			BestBlockHash:   res.Sync.BestBlockHash,
			BestBlockHeight: res.Sync.BestBlockHeight,
			BestBlockTime:   time.UnixMilli(res.Sync.BestBlockTime),
			Syncing:         res.Sync.Syncing,
		},
		Validator: &adminTypes.ValidatorInfo{
			PubKey: res.Validator.PubKey,
			Power:  res.Validator.Power,
		},
	}, nil
}

// Version reports the version of the running node.
func (cl *Client) Version(ctx context.Context) (string, error) {
	cmd := &jsonrpc.VersionRequest{}
	res := &jsonrpc.VersionResponse{}
	err := cl.CallMethod(ctx, string(adminjson.MethodVersion), cmd, res)
	if err != nil {
		return "", err
	}
	return res.KwilVersion, err
}

// ListPendingJoins lists all active validator join requests.
func (cl *Client) ListPendingJoins(ctx context.Context) ([]*types.JoinRequest, error) {
	cmd := &adminjson.ListJoinRequestsRequest{}
	res := &adminjson.ListJoinRequestsResponse{}
	err := cl.CallMethod(ctx, string(adminjson.MethodValListJoins), cmd, res)
	if err != nil {
		return nil, err
	}
	return res.JoinRequests, err
}

// GetConfig gets the current config from the node.
// It returns the config serialized as JSON.
func (cl *Client) GetConfig(ctx context.Context) ([]byte, error) {
	cmd := &adminjson.GetConfigRequest{}
	res := &adminjson.GetConfigResponse{}
	err := cl.CallMethod(ctx, string(adminjson.MethodConfig), cmd, res)
	if err != nil {
		return nil, err
	}
	return res.Config, err
}

// Ping just tests RPC connectivity. The expected response is "pong".
func (cl *Client) Ping(ctx context.Context) (string, error) {
	cmd := &jsonrpc.PingRequest{
		Message: "ping",
	}
	res := &jsonrpc.PingResponse{}
	err := cl.CallMethod(ctx, string(jsonrpc.MethodPing), cmd, res)
	if err != nil {
		return "", err
	}
	return res.Message, nil
}
