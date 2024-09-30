package operator

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"

	"github.com/kwilteam/kwil-db/cmd/common/display"
	"github.com/kwilteam/kwil-db/core/types"
	admintypes "github.com/kwilteam/kwil-db/core/types/admin"
	"github.com/kwilteam/kwil-db/test/driver"
)

// OperatorCLIDriver is a driver for the operator using the kwil-admin cli.
type OperatorCLIDriver struct {
	Exec   ExecFn // execute a command
	RpcUrl string // rpc url (either unix socket path, host:port, or http/https URL)
}

// ExecFn executes a CLI command for the admin client.
type ExecFn func(ctx context.Context, args ...string) ([]byte, error)

var _ KwilOperatorDriver = (*OperatorCLIDriver)(nil)

// runCommand runs a kwil-admin command with the node's rpcserver.
// it returns the generic response.
// It will unmarshal the response into the provided result.
func (o *OperatorCLIDriver) runCommand(ctx context.Context, result any, args ...string) error {
	args = append(args, "--rpcserver", o.RpcUrl)
	args = append(args, "--output", "json")

	bts, err := o.Exec(ctx, args...)
	if err != nil {
		return err
	}

	// cli returns json response with an error field if there was an error
	resp := cliResponse{}
	err = json.Unmarshal(bts, &resp)
	if err != nil {
		return err
	}

	if resp.Error != "" {
		return errors.New(resp.Error)
	}

	// unmarshal the result into the provided result
	bts, err = json.Marshal(resp.Result)
	if err != nil {
		return err
	}

	return json.Unmarshal(bts, result)
}

func (o *OperatorCLIDriver) TxSuccess(ctx context.Context, txHash []byte) error {
	var res respTxQuery
	err := o.runCommand(ctx, &res, "utils", "query-tx", hex.EncodeToString(txHash))
	if err != nil {
		return err
	}

	// NOTE: this should not be considered a failure, should retry
	if res.Height < 0 {
		return driver.ErrTxNotConfirmed
	}

	if res.TxResult.Code != 0 {
		return fmt.Errorf("tx failed: %s", res.TxResult.Log)
	}
	return nil
}

func (o *OperatorCLIDriver) ValidatorJoinStatus(ctx context.Context, pubKey []byte) (*types.JoinRequest, error) {
	var res types.JoinRequest
	err := o.runCommand(ctx, &res, "validators", "join-status", hex.EncodeToString(pubKey))
	if err != nil {
		return nil, err
	}

	return &res, nil
}

// commands that return a tx hash return a hex encoded string
func (o *OperatorCLIDriver) ValidatorNodeApprove(ctx context.Context, joinerPubKey []byte) ([]byte, error) {
	var res display.TxHashResponse
	err := o.runCommand(ctx, &res, "validators", "approve", hex.EncodeToString(joinerPubKey))
	if err != nil {
		return nil, err
	}

	return hex.DecodeString(res.TxHash)
}

func (o *OperatorCLIDriver) ValidatorNodeJoin(ctx context.Context) ([]byte, error) {
	var res display.TxHashResponse
	err := o.runCommand(ctx, &res, "validators", "join")
	if err != nil {
		return nil, err
	}

	return hex.DecodeString(res.TxHash)
}

func (o *OperatorCLIDriver) ValidatorNodeLeave(ctx context.Context) ([]byte, error) {
	var res display.TxHashResponse
	err := o.runCommand(ctx, &res, "validators", "leave")
	if err != nil {
		return nil, err
	}

	return hex.DecodeString(res.TxHash)
}

func (o *OperatorCLIDriver) ValidatorNodeRemove(ctx context.Context, target []byte) ([]byte, error) {
	var res display.TxHashResponse
	err := o.runCommand(ctx, &res, "validators", "remove", hex.EncodeToString(target))
	if err != nil {
		return nil, err
	}

	return hex.DecodeString(res.TxHash)
}

func (o *OperatorCLIDriver) ValidatorsList(ctx context.Context) ([]*types.Validator, error) {
	var res []*types.Validator
	err := o.runCommand(ctx, &res, "validators", "list")
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (o *OperatorCLIDriver) AddPeer(ctx context.Context, peerID string) error {
	var peer string
	return o.runCommand(ctx, &peer, "whitelist", "add", peerID)
}

func (o *OperatorCLIDriver) RemovePeer(ctx context.Context, peerID string) error {
	var peer string
	return o.runCommand(ctx, &peer, "whitelist", "remove", peerID)
}

func (o *OperatorCLIDriver) ListPeers(ctx context.Context) ([]string, error) {
	var res []string
	err := o.runCommand(ctx, &res, "whitelist", "list")
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (o *OperatorCLIDriver) ConnectedPeers(ctx context.Context) ([]string, error) {
	var res []*admintypes.PeerInfo
	err := o.runCommand(ctx, &res, "node", "peers")
	if err != nil {
		return nil, err
	}

	var peers []string
	for _, p := range res {
		peers = append(peers, p.RemoteAddr)
	}
	return peers, nil
}

type cliResponse struct {
	Result any    `json:"result"`
	Error  string `json:"error"`
}

// respTxQuery represents the tx query response(json) from the cli response
type respTxQuery struct {
	Height   int64 `json:"height"`
	TxResult struct {
		Code uint32 `json:"code"`
		Log  string `json:"log"`
	} `json:"tx_result"`
}

func (o *OperatorCLIDriver) SubmitMigrationProposal(ctx context.Context, activationHeight, migrationDuration *big.Int) ([]byte, error) {
	var res display.TxHashResponse
	err := o.runCommand(ctx, &res, "migrate", "propose", "--activation-period", activationHeight.String(), "--duration", migrationDuration.String())
	if err != nil {
		return nil, err
	}

	return hex.DecodeString(res.TxHash)
}

func (o *OperatorCLIDriver) ApproveMigration(ctx context.Context, migrationResolutionID *types.UUID) ([]byte, error) {
	var res display.TxHashResponse
	err := o.runCommand(ctx, &res, "migrate", "approve", migrationResolutionID.String())
	if err != nil {
		return nil, err
	}

	return hex.DecodeString(res.TxHash)
}

func (o *OperatorCLIDriver) GenesisState(ctx context.Context) (*types.MigrationMetadata, error) {
	var res *types.MigrationMetadata
	err := o.runCommand(ctx, &res, "migrate", "genesis-state")
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (o *OperatorCLIDriver) ListMigrations(ctx context.Context) ([]*types.Migration, error) {
	var res []*types.Migration
	err := o.runCommand(ctx, &res, "migrate", "list")
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (o *OperatorCLIDriver) GenesisSnapshotChunk(ctx context.Context, height uint64, chunkIdx uint32) ([]byte, error) {
	panic("not implemented")
}
