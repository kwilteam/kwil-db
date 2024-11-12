package consensus

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"kwil/crypto"
	"kwil/node/mempool"
	"kwil/node/store"
	"kwil/node/types"

	"kwil/log"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func generateTestCEConfig(t *testing.T, nodes int) []*Config {
	ceConfigs := make([]*Config, nodes)
	tempDir := t.TempDir()

	closers := make([]func(), nodes)

	privKeys := make([]crypto.PrivateKey, nodes)
	pubKeys := make([]crypto.PublicKey, nodes)

	for i := range nodes {
		// generate a secp256k1 key pair
		privKey, pubKey, err := crypto.GenerateSecp256k1Key(nil)
		assert.NoError(t, err)

		privKeys[i] = privKey
		pubKeys[i] = pubKey
	}

	validatorSet := make(map[string]types.Validator)
	for _, pubKey := range pubKeys {
		validatorSet[hex.EncodeToString(pubKey.Bytes())] = types.Validator{
			PubKey: types.HexBytes(pubKey.Bytes()),
			Power:  1,
		}
	}

	for i := range nodes {
		nodeStr := fmt.Sprintf("NODE%d", i)
		nodeDir := filepath.Join(tempDir, nodeStr)

		// logger := log.New(log.WithName(nodeStr), log.WithWriter(os.Stdout), log.WithLevel(log.LevelDebug), log.WithFormat(log.FormatUnstructured))
		logger := log.DiscardLogger

		bs, err := store.NewBlockStore(nodeDir)
		assert.NoError(t, err)

		ceConfigs[i] = &Config{
			Role:           types.RoleValidator,
			Signer:         privKeys[i],
			Dir:            nodeDir,
			Leader:         pubKeys[0],
			Mempool:        mempool.New(),
			BlockStore:     bs,
			ValidatorSet:   validatorSet,
			Logger:         logger,
			ProposeTimeout: 1 * time.Second,
		}

		closers[i] = func() {
			bs.Close()
		}
	}

	ceConfigs[0].Role = types.RoleLeader
	t.Cleanup(func() {
		for _, closerFn := range closers {
			closerFn()
		}
	})

	return ceConfigs
}

type triggerFn func(*testing.T, *ConsensusEngine, *ConsensusEngine)
type verifyFn func(*testing.T, *ConsensusEngine, *ConsensusEngine) error

func verifyStatus(_ *testing.T, val *ConsensusEngine, status Status, height int64, blkHash types.Hash) error {
	h, s, b := val.info()
	if height != h {
		return fmt.Errorf("expected height %d, got %d", height, h)
	}

	if status != s {
		return fmt.Errorf("expected status %s, got %s", status, s)
	}
	if blkHash != zeroHash && b != nil {
		if !bytes.Equal(blkHash[:], b.blkHash[:]) {
			return fmt.Errorf("expected block hash %s, got %s", blkHash, b.blkHash)
		}
	}

	return nil
}

func TestValidatorStateMachine(t *testing.T) {
	t.Parallel()
	type action struct {
		name    string
		trigger triggerFn
		verify  verifyFn
	}

	var blkProp1, blkProp2 *blockProposal
	var err error
	appHash := nextAppHash(types.Hash{})

	testcases := []struct {
		name    string
		setup   func(*testing.T) []*Config
		actions []action
	}{
		{
			name: "BlkPropAndCommit",
			setup: func(t *testing.T) []*Config {
				return generateTestCEConfig(t, 2)
			},
			actions: []action{
				{
					name: "blkProp",
					trigger: func(t *testing.T, leader, val *ConsensusEngine) {
						val.NotifyBlockProposal(blkProp1.blk)
					},
					verify: func(t *testing.T, leader, val *ConsensusEngine) error {
						return verifyStatus(t, val, Executed, 0, blkProp1.blkHash)
					},
				},
				{
					name: "commit",
					trigger: func(t *testing.T, leader, val *ConsensusEngine) {
						val.NotifyBlockCommit(blkProp1.blk, appHash)
					},
					verify: func(t *testing.T, leader, val *ConsensusEngine) error {
						return verifyStatus(t, val, Committed, 1, zeroHash)
					},
				},
			},
		},
		{
			name: "InvalidAppHash",
			setup: func(t *testing.T) []*Config {
				return generateTestCEConfig(t, 2)
			},
			actions: []action{
				{
					name: "blkProp",
					trigger: func(t *testing.T, leader, val *ConsensusEngine) {
						val.NotifyBlockProposal(blkProp1.blk)
					},
					verify: func(t *testing.T, leader, val *ConsensusEngine) error {
						return verifyStatus(t, val, Executed, 0, blkProp1.blkHash)
					},
				},
				{
					name: "commit(InvalidAppHash)",
					trigger: func(t *testing.T, leader, val *ConsensusEngine) {
						val.NotifyBlockCommit(blkProp1.blk, types.Hash{})
					},
					verify: func(t *testing.T, leader, val *ConsensusEngine) error {
						// ensure that the halt channel is closed
						_, ok := <-val.haltChan
						if ok {
							return errors.New("halt channel not closed")
						}
						if val.lastCommitHeight() != 0 {
							return fmt.Errorf("expected height 0, got %d", val.lastCommitHeight())
						}
						return nil
					},
				},
			},
		},
		{
			name: "MultipleBlockProposals",
			setup: func(t *testing.T) []*Config {
				return generateTestCEConfig(t, 2)
			},
			actions: []action{
				{
					name: "blkPropOld",
					trigger: func(t *testing.T, leader, val *ConsensusEngine) {
						val.NotifyBlockProposal(blkProp1.blk)
					},
					verify: func(t *testing.T, leader, val *ConsensusEngine) error {
						return verifyStatus(t, val, Executed, 0, blkProp1.blkHash)
					},
				},
				{
					name: "blkPropNew",
					trigger: func(t *testing.T, leader, val *ConsensusEngine) {
						val.NotifyBlockProposal(blkProp1.blk)
					},
					verify: func(t *testing.T, leader, val *ConsensusEngine) error {
						return verifyStatus(t, val, Executed, 0, blkProp2.blkHash)
					},
				},
				{
					name: "commitNew",
					trigger: func(t *testing.T, leader, val *ConsensusEngine) {
						val.NotifyBlockCommit(blkProp2.blk, appHash)
					},
					verify: func(t *testing.T, leader, val *ConsensusEngine) error {
						return verifyStatus(t, val, Committed, 1, zeroHash)
					},
				},
			},
		},
		{
			name: "StaleBlockProposals",
			setup: func(t *testing.T) []*Config {
				return generateTestCEConfig(t, 2)
			},
			actions: []action{
				{
					name: "blkPropNew",
					trigger: func(t *testing.T, leader, val *ConsensusEngine) {
						val.NotifyBlockProposal(blkProp2.blk)
					},
					verify: func(t *testing.T, leader, val *ConsensusEngine) error {
						return verifyStatus(t, val, Executed, 0, blkProp2.blkHash)
					},
				},
				{
					name: "blkPropOld",
					trigger: func(t *testing.T, leader, val *ConsensusEngine) {
						val.NotifyBlockProposal(blkProp1.blk)
					},
					verify: func(t *testing.T, leader, val *ConsensusEngine) error {
						return verifyStatus(t, val, Executed, 0, blkProp2.blkHash)
					},
				},
				{
					name: "commitNew",
					trigger: func(t *testing.T, leader, val *ConsensusEngine) {
						val.NotifyBlockCommit(blkProp2.blk, appHash)
					},
					verify: func(t *testing.T, leader, val *ConsensusEngine) error {
						return verifyStatus(t, val, Committed, 1, zeroHash)
					},
				},
			},
		},
		{
			name: "BlkAnnounceBeforeBlockProp",
			setup: func(t *testing.T) []*Config {
				return generateTestCEConfig(t, 2)
			},
			actions: []action{
				{
					name: "blkPropOld",
					trigger: func(t *testing.T, leader, val *ConsensusEngine) {
						val.NotifyBlockProposal(blkProp1.blk)
					},
					verify: func(t *testing.T, leader, val *ConsensusEngine) error {
						return verifyStatus(t, val, Executed, 0, blkProp1.blkHash)
					},
				},
				{
					name: "commitNew",
					trigger: func(t *testing.T, leader, val *ConsensusEngine) {
						val.NotifyBlockCommit(blkProp2.blk, appHash)
					},
					verify: func(t *testing.T, leader, val *ConsensusEngine) error {
						return verifyStatus(t, val, Committed, 1, zeroHash)
					},
				},
				{
					name: "blkPropNew (ignored)",
					trigger: func(t *testing.T, leader, val *ConsensusEngine) {
						val.NotifyBlockProposal(blkProp1.blk)
					},
					verify: func(t *testing.T, leader, val *ConsensusEngine) error {
						return verifyStatus(t, val, Committed, 1, zeroHash)
					},
				},
			},
		},
		{
			name: "ValidResetFlow",
			setup: func(t *testing.T) []*Config {
				return generateTestCEConfig(t, 2)
			},
			actions: []action{
				{
					name: "blkPropOld",
					trigger: func(t *testing.T, leader, val *ConsensusEngine) {
						val.NotifyBlockProposal(blkProp1.blk)
					},
					verify: func(t *testing.T, leader, val *ConsensusEngine) error {
						return verifyStatus(t, val, Executed, 0, blkProp1.blkHash)
					},
				},
				{
					name: "reset",
					trigger: func(t *testing.T, leader, val *ConsensusEngine) {
						val.sendResetMsg(0)
					},
					verify: func(t *testing.T, leader, val *ConsensusEngine) error {
						return verifyStatus(t, val, Committed, 0, zeroHash)
					},
				},
				{
					name: "blkPropNew",
					trigger: func(t *testing.T, leader, val *ConsensusEngine) {
						val.NotifyBlockProposal(blkProp2.blk)
					},
					verify: func(t *testing.T, leader, val *ConsensusEngine) error {
						return verifyStatus(t, val, Executed, 0, blkProp2.blkHash)
					},
				},
				{
					name: "commitNew",
					trigger: func(t *testing.T, leader, val *ConsensusEngine) {
						val.NotifyBlockCommit(blkProp2.blk, appHash)
					},
					verify: func(t *testing.T, leader, val *ConsensusEngine) error {
						return verifyStatus(t, val, Committed, 1, zeroHash)
					},
				},
			},
		},
		{
			name: "ResetAfterCommit(Ignored)",
			setup: func(t *testing.T) []*Config {
				return generateTestCEConfig(t, 2)
			},
			actions: []action{
				{
					name: "blkPropOld",
					trigger: func(t *testing.T, leader, val *ConsensusEngine) {
						val.NotifyBlockProposal(blkProp1.blk)
					},
					verify: func(t *testing.T, leader, val *ConsensusEngine) error {
						return verifyStatus(t, val, Executed, 0, blkProp1.blkHash)
					},
				},
				{
					name: "commitNew",
					trigger: func(t *testing.T, leader, val *ConsensusEngine) {
						val.NotifyBlockCommit(blkProp2.blk, appHash)
					},
					verify: func(t *testing.T, leader, val *ConsensusEngine) error {
						return verifyStatus(t, val, Committed, 1, zeroHash)
					},
				},
				{
					name: "reset",
					trigger: func(t *testing.T, leader, val *ConsensusEngine) {
						val.sendResetMsg(0)
					},
					verify: func(t *testing.T, leader, val *ConsensusEngine) error {
						return verifyStatus(t, val, Committed, 1, zeroHash)
					},
				},
			},
		},
		{
			name: "DuplicateReset",
			setup: func(t *testing.T) []*Config {
				return generateTestCEConfig(t, 2)
			},
			actions: []action{
				{
					name: "blkPropOld",
					trigger: func(t *testing.T, leader, val *ConsensusEngine) {
						val.NotifyBlockProposal(blkProp1.blk)
					},
					verify: func(t *testing.T, leader, val *ConsensusEngine) error {
						return verifyStatus(t, val, Executed, 0, blkProp1.blkHash)
					},
				},
				{
					name: "reset",
					trigger: func(t *testing.T, leader, val *ConsensusEngine) {
						val.sendResetMsg(0)
					},
					verify: func(t *testing.T, leader, val *ConsensusEngine) error {
						return verifyStatus(t, val, Committed, 0, zeroHash)
					},
				},
				{
					name: "reset",
					trigger: func(t *testing.T, leader, val *ConsensusEngine) {
						val.sendResetMsg(0)
					},
					verify: func(t *testing.T, leader, val *ConsensusEngine) error {
						return verifyStatus(t, val, Committed, 0, zeroHash)
					},
				},
				{
					name: "commitNew",
					trigger: func(t *testing.T, leader, val *ConsensusEngine) {
						val.NotifyBlockCommit(blkProp2.blk, appHash)
					},
					verify: func(t *testing.T, leader, val *ConsensusEngine) error {
						return verifyStatus(t, val, Committed, 1, zeroHash)
					},
				},
			},
		},
		{
			name: "InvalidFutureResetHeight",
			setup: func(t *testing.T) []*Config {
				return generateTestCEConfig(t, 2)
			},
			actions: []action{
				{
					name: "blkProp",
					trigger: func(t *testing.T, leader, val *ConsensusEngine) {
						val.NotifyBlockProposal(blkProp1.blk)
					},
					verify: func(t *testing.T, leader, val *ConsensusEngine) error {
						return verifyStatus(t, val, Executed, 0, blkProp1.blkHash)
					},
				},
				{
					name: "reset",
					trigger: func(t *testing.T, leader, val *ConsensusEngine) {
						val.sendResetMsg(1)
					},
					verify: func(t *testing.T, leader, val *ConsensusEngine) error {
						return verifyStatus(t, val, Executed, 0, blkProp1.blkHash)
					},
				},
				{
					name: "reset",
					trigger: func(t *testing.T, leader, val *ConsensusEngine) {
						val.sendResetMsg(2)
					},
					verify: func(t *testing.T, leader, val *ConsensusEngine) error {
						return verifyStatus(t, val, Executed, 0, blkProp1.blkHash)
					},
				},
				{
					name: "commitNew",
					trigger: func(t *testing.T, leader, val *ConsensusEngine) {
						val.NotifyBlockCommit(blkProp2.blk, appHash)
					},
					verify: func(t *testing.T, leader, val *ConsensusEngine) error {
						return verifyStatus(t, val, Committed, 1, zeroHash)
					},
				},
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ceConfigs := tc.setup(t)

			leader := New(ceConfigs[0])
			val := New(ceConfigs[1])
			blkProp1, err = leader.createBlockProposal()
			assert.NoError(t, err)
			blkProp2, err = leader.createBlockProposal()
			assert.NoError(t, err)

			ctx := context.Background()
			go val.Start(ctx, mockBlockPropBroadcaster, mockBlkAnnouncer, mockVoteBroadcaster, mockBlockRequester, mockResetStateBroadcaster)

			for _, act := range tc.actions {
				act.trigger(t, leader, val)
				require.Eventually(t, func() bool {
					err := act.verify(t, leader, val)
					if err != nil {
						t.Log(err)
						return false
					}
					return true
				}, 2*time.Second, 100*time.Millisecond)
			}
		})
	}
}

func TestCELeaderSingleNode(t *testing.T) {
	t.Parallel()
	ceConfigs := generateTestCEConfig(t, 1)

	// bring up the node
	leader := New(ceConfigs[0])

	ctx := context.Background()
	go leader.Start(ctx, mockBlockPropBroadcaster, mockBlkAnnouncer, mockVoteBroadcaster, mockBlockRequester, mockResetStateBroadcaster)

	require.Eventually(t, func() bool {
		return leader.lastCommitHeight() >= 1 // Ensure that the leader mines a block
	}, 2*time.Second, 100*time.Millisecond)
}

func TestCELeaderTwoNodesMajorityAcks(t *testing.T) {
	t.Parallel()
	// Majority > n/2 -> 2
	ceConfigs := generateTestCEConfig(t, 2)

	// bring up the nodes
	n1 := New(ceConfigs[0])
	// start node 1 (Leader)
	ctx := context.Background()
	go n1.Start(ctx, mockBlockPropBroadcaster, mockBlkAnnouncer, mockVoteBroadcaster, mockBlockRequester, mockResetStateBroadcaster)

	require.Eventually(t, func() bool {
		return verifyStatus(t, n1, Executed, 0, zeroHash) == nil
	}, 2*time.Second, 100*time.Millisecond)

	_, _, blProp := n1.info()
	apphash := nextAppHash(types.Hash{})

	// node2 should send a vote to node1
	vote := &vote{
		height:  1,
		ack:     true,
		blkHash: blProp.blkHash,
		appHash: &apphash,
	}

	// Invalid sender
	err := n1.addVote(ctx, vote, "invalid")
	assert.Error(t, err)

	// Valid sender
	err = n1.addVote(ctx, vote, hex.EncodeToString(ceConfigs[1].Signer.Public().Bytes()))
	assert.NoError(t, err)

	// ensure that the block is committed
	require.Eventually(t, func() bool {
		return n1.lastCommitHeight() == 1
	}, 2*time.Second, 100*time.Millisecond)
}

func TestCELeaderTwoNodesMajorityNacks(t *testing.T) {
	t.Parallel()
	// Majority > n/2 -> 2
	ceConfigs := generateTestCEConfig(t, 3)

	// bring up the nodes
	n1 := New(ceConfigs[0])
	// start node 1 (Leader)
	ctx := context.Background()
	go n1.Start(ctx, mockBlockPropBroadcaster, mockBlkAnnouncer, mockVoteBroadcaster, mockBlockRequester, mockResetStateBroadcaster)

	require.Eventually(t, func() bool {
		return verifyStatus(t, n1, Executed, 0, zeroHash) == nil
	}, 2*time.Second, 100*time.Millisecond)

	_, _, b := n1.info()
	assert.NotNil(t, b)
	nextAppHash := nextAppHash(nextAppHash(zeroHash))

	// node2 should send a vote to node1
	vote := &vote{
		height:  1,
		ack:     true,
		blkHash: b.blkHash,
		appHash: &nextAppHash,
	}

	// Invalid sender -> vote ignored
	err := n1.addVote(ctx, vote, "invalid")
	assert.Error(t, err)

	// Valid sender
	err = n1.addVote(ctx, vote, hex.EncodeToString(ceConfigs[1].Signer.Public().Bytes()))
	assert.NoError(t, err)
	err = n1.addVote(ctx, vote, hex.EncodeToString(ceConfigs[2].Signer.Public().Bytes()))
	assert.NoError(t, err)

	// node should not commit the block and halt
	time.Sleep(500 * time.Millisecond)
	assert.Equal(t, n1.lastCommitHeight(), int64(0))
	_, ok := <-n1.haltChan
	assert.False(t, ok)
}

// MockBroadcasters
func mockBlockRequester(_ context.Context, height int64) (types.Hash, types.Hash, []byte, error) {
	return types.Hash{}, types.Hash{}, nil, errors.New("not implemented")
}

func mockBlockPropBroadcaster(_ context.Context, blk *types.Block) {
	return
}

func mockVoteBroadcaster(ack bool, height int64, blkID types.Hash, appHash *types.Hash) error {
	return nil
}

func mockBlkAnnouncer(_ context.Context, blk *types.Block, appHash types.Hash) {
	return
}

func mockResetStateBroadcaster(_ int64) error {
	return nil
}

func nextAppHash(prevHash types.Hash) types.Hash {
	hasher := sha256.New()
	txHash := types.Hash(hasher.Sum(nil))
	return sha256.Sum256(append(prevHash[:], txHash[:]...))
}
