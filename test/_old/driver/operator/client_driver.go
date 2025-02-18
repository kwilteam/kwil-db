package operator

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/kwilteam/kwil-db/core/crypto"
	"github.com/kwilteam/kwil-db/core/types"
	adminclient "github.com/kwilteam/kwil-db/node/admin"
	"github.com/kwilteam/kwil-db/node/migrations"
	"github.com/kwilteam/kwil-db/node/voting"

	"github.com/kwilteam/kwil-db/test/driver"
)

type AdminClientDriver struct {
	Client *adminclient.AdminClient
}

var _ KwilOperatorDriver = (*AdminClientDriver)(nil)

func (a *AdminClientDriver) TxSuccess(ctx context.Context, txHash types.Hash) error {
	resp, err := a.Client.TxQuery(ctx, txHash)
	if err != nil {
		return fmt.Errorf("failed to query: %w", err)
	}

	if resp.Result.Code != uint32(types.CodeOk) {
		return fmt.Errorf("transaction not ok, %s", resp.Result.Log)
	}

	// NOTE: THIS should not be considered a failure, should retry
	if resp.Height < 0 {
		return driver.ErrTxNotConfirmed
	}

	return nil
}

func (a *AdminClientDriver) ValidatorJoinStatus(ctx context.Context, pubKey []byte, pubKeyType crypto.KeyType) (*types.JoinRequest, error) {
	return a.Client.JoinStatus(ctx, pubKey, pubKeyType)
}

func (a *AdminClientDriver) ValidatorNodeApprove(ctx context.Context, joinerPubKey []byte, pubKeyType crypto.KeyType) (types.Hash, error) {
	return a.Client.Approve(ctx, joinerPubKey, pubKeyType)
}

func (a *AdminClientDriver) ValidatorNodeJoin(ctx context.Context) (types.Hash, error) {
	return a.Client.Join(ctx)
}

func (a *AdminClientDriver) ValidatorNodeLeave(ctx context.Context) (types.Hash, error) {
	return a.Client.Leave(ctx)
}

func (a *AdminClientDriver) ValidatorNodeRemove(ctx context.Context, target []byte, pubKeyType crypto.KeyType) (types.Hash, error) {
	return a.Client.Remove(ctx, target, pubKeyType)
}

func (a *AdminClientDriver) ValidatorsList(ctx context.Context) ([]*types.Validator, error) {
	return a.Client.ListValidators(ctx)
}

func (a *AdminClientDriver) AddPeer(ctx context.Context, peerID string) error {
	return a.Client.AddPeer(ctx, peerID)
}

func (a *AdminClientDriver) ListPeers(ctx context.Context) ([]string, error) {
	return a.Client.ListPeers(ctx)
}

func (a *AdminClientDriver) RemovePeer(ctx context.Context, peerID string) error {
	return a.Client.RemovePeer(ctx, peerID)
}

func (a *AdminClientDriver) ConnectedPeers(ctx context.Context) ([]string, error) {
	peersInfo, err := a.Client.Peers(ctx)
	if err != nil {
		return nil, err
	}

	peers := make([]string, 0, len(peersInfo))
	for _, peer := range peersInfo {
		peers = append(peers, peer.RemoteAddr)
	}

	return peers, nil
}

func (a *AdminClientDriver) SubmitMigrationProposal(ctx context.Context, activationPeriod, migrationDuration *big.Int) (types.Hash, error) {
	// return a.Client.SubmitMigrationProposal(ctx, activationHeight, migrationDuration, chainID)
	activationHeight := activationPeriod.Uint64()
	dur := migrationDuration.Uint64()

	res := migrations.MigrationDeclaration{
		ActivationPeriod: activationHeight,
		Duration:         dur,
		Timestamp:        time.Now().String(),
	}
	proposalBts, err := res.MarshalBinary()
	if err != nil {
		return types.Hash{}, err
	}

	return a.Client.CreateResolution(ctx, proposalBts, voting.StartMigrationEventType)
}

func (a *AdminClientDriver) ApproveMigration(ctx context.Context, migrationResolutionID *types.UUID) (types.Hash, error) {
	return a.Client.ApproveResolution(ctx, migrationResolutionID)
}

// func (a *AdminClientDriver) DeleteMigration(ctx context.Context, migrationResolutionID *types.UUID) ([]byte, error) {
// 	return a.Client.DeleteResolution(ctx, migrationResolutionID)
// }

func (a *AdminClientDriver) GenesisState(ctx context.Context) (*types.MigrationMetadata, error) {
	return a.Client.GenesisState(ctx)
}

func (a *AdminClientDriver) ListMigrations(ctx context.Context) ([]*types.Migration, error) {
	return a.Client.ListMigrations(ctx)
}

func (a *AdminClientDriver) GenesisSnapshotChunk(ctx context.Context, height uint64, chunkIdx uint32) ([]byte, error) {
	return a.Client.GenesisSnapshotChunk(ctx, height, chunkIdx)
}
