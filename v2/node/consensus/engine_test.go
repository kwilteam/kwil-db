package consensus

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"kwil/crypto"
	"kwil/log"
	"kwil/node/mempool"
	"kwil/node/meta"
	dbtest "kwil/node/pg/test"
	"kwil/node/store"
	"kwil/node/txapp"
	"kwil/node/types"
	"kwil/node/types/sql"
	ktypes "kwil/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func generateTestCEConfig(t *testing.T, nodes int) []*Config {
	ceConfigs := make([]*Config, nodes)
	tempDir := t.TempDir()

	closers := make([]func(), 0)

	privKeys := make([]crypto.PrivateKey, nodes)
	pubKeys := make([]crypto.PublicKey, nodes)

	for i := range nodes {
		// generate a secp256k1 key pair
		privKey, pubKey, err := crypto.GenerateSecp256k1Key(nil)
		assert.NoError(t, err)

		privKeys[i] = privKey
		pubKeys[i] = pubKey
	}

	validatorSet := make(map[string]ktypes.Validator)
	var valSet []*ktypes.Validator
	for _, pubKey := range pubKeys {
		val := &ktypes.Validator{
			PubKey: types.HexBytes(pubKey.Bytes()),
			Power:  1,
		}
		validatorSet[hex.EncodeToString(pubKey.Bytes())] = *val
		valSet = append(valSet, val)
	}

	db, err := dbtest.NewTestDB(t)
	require.NoError(t, err)

	ctx := context.Background()

	prepTx, err := db.BeginPreparedTx(ctx)
	require.NoError(t, err)

	err = meta.InitializeMetaStore(ctx, prepTx)
	assert.NoError(t, err)

	assert.NoError(t, prepTx.Commit(ctx))
	// Account Store
	// accounts, err := accounts.InitializeAccountStore(ctx, db)
	// assert.NoError(t, err)
	accounts := &mockAccounts{}

	v := newValidatorStore(valSet)
	// txapp, err := txapp.NewTxApp(ctx, db, nil, signer, nil, service, accounts, v)
	// assert.NoError(t, err)
	txapp := newDummyTxApp(valSet)

	for i := range nodes {
		nodeStr := fmt.Sprintf("NODE%d", i)
		nodeDir := filepath.Join(tempDir, nodeStr)

		logger := log.New(log.WithName(nodeStr), log.WithWriter(os.Stdout), log.WithLevel(log.LevelDebug), log.WithFormat(log.FormatUnstructured))
		// logger := log.DiscardLogger

		bs, err := store.NewBlockStore(nodeDir)
		assert.NoError(t, err)

		ceConfigs[i] = &Config{
			DB:             db,
			PrivateKey:     privKeys[i],
			Dir:            nodeDir,
			Leader:         pubKeys[0],
			Mempool:        mempool.New(),
			BlockStore:     bs,
			TxApp:          txapp,
			Accounts:       accounts,
			ValidatorSet:   validatorSet,
			ValidatorStore: v,
			Logger:         logger,
			ProposeTimeout: 1 * time.Second,
		}

		closers = append(closers, func() {
			bs.Close()
		})
	}

	t.Cleanup(func() {
		db.AutoCommit(true)
		defer db.AutoCommit(false)
		defer db.Close()
		ctx := context.Background()
		db.Execute(ctx, `DROP SCHEMA IF EXISTS kwild_chain CASCADE;`)
		db.Execute(ctx, `DROP SCHEMA IF EXISTS kwild_internal CASCADE;`)

		for _, closerFn := range closers {
			closerFn()
		}
	})

	return ceConfigs
}

var blockAppHash = nextHash()

func nextHash() types.Hash {
	newHash, err := ktypes.NewHashFromString("2bf6a0d3cd2cce6a2ff0d67f2f252842aa5541eb2b870792b29f5bac699ac7ec")
	if err != nil {
		panic(err)
	}
	return newHash
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
	// t.Parallel()
	type action struct {
		name    string
		trigger triggerFn
		verify  verifyFn
	}

	var blkProp1, blkProp2 *blockProposal
	var err error

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
						// appHash := val.blockResult().appHash
						// val.NotifyBlockCommit(blkProp1.blk, appHash)
						val.NotifyBlockCommit(blkProp1.blk, blockAppHash)
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
						// appHash := val.blockResult().appHash
						val.NotifyBlockCommit(blkProp2.blk, blockAppHash)
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
						// appHash := val.blockResult().appHash
						// val.NotifyBlockCommit(blkProp2.blk, appHash)
						val.NotifyBlockCommit(blkProp2.blk, blockAppHash)
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
						// appHash := val.blockResult().appHash
						// val.NotifyBlockCommit(blkProp2.blk, appHash)
						val.NotifyBlockCommit(blkProp2.blk, blockAppHash)
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
						// appHash := val.blockResult().appHash
						// val.NotifyBlockCommit(blkProp2.blk, appHash)
						val.NotifyBlockCommit(blkProp2.blk, blockAppHash)
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
						// appHash := val.blockResult().appHash
						// val.NotifyBlockCommit(blkProp2.blk, appHash)
						val.NotifyBlockCommit(blkProp2.blk, blockAppHash)
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
						// appHash := val.blockResult().appHash
						// val.NotifyBlockCommit(blkProp2.blk, appHash)
						val.NotifyBlockCommit(blkProp2.blk, blockAppHash)
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
						// appHash := val.blockResult().appHash
						// val.NotifyBlockCommit(blkProp2.blk, appHash)
						val.NotifyBlockCommit(blkProp2.blk, blockAppHash)
					},
					verify: func(t *testing.T, leader, val *ConsensusEngine) error {
						return verifyStatus(t, val, Committed, 1, zeroHash)
					},
				},
			},
		},
	}

	for _, tc := range testcases {
		t.Log(tc.name)
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
	// t.Parallel()
	ceConfigs := generateTestCEConfig(t, 1)

	// bring up the node
	leader := New(ceConfigs[0])

	ctx := context.Background()
	go leader.Start(ctx, mockBlockPropBroadcaster, mockBlkAnnouncer, mockVoteBroadcaster, mockBlockRequester, mockResetStateBroadcaster)

	require.Eventually(t, func() bool {
		return leader.lastCommitHeight() >= 1 // Ensure that the leader mines a block
	}, 2*time.Second, 100*time.Millisecond)

	ctx.Done()
}

func TestCELeaderTwoNodesMajorityAcks(t *testing.T) {
	// Majority > n/2 -> 2
	ceConfigs := generateTestCEConfig(t, 2)

	// bring up the nodes
	n1 := New(ceConfigs[0])
	// start node 1 (Leader)
	ctx := context.Background()
	go n1.Start(ctx, mockBlockPropBroadcaster, mockBlkAnnouncer, mockVoteBroadcaster, mockBlockRequester, mockResetStateBroadcaster)

	time.Sleep(500 * time.Millisecond)
	_, _, blProp := n1.info()
	// appHash := nextAppHash()

	// node2 should send a vote to node1
	vote := &vote{
		height:  1,
		ack:     true,
		blkHash: blProp.blkHash,
		appHash: &blockAppHash,
	}

	// Invalid sender
	err := n1.addVote(ctx, vote, "invalid")
	assert.Error(t, err)

	// Valid sender
	err = n1.addVote(ctx, vote, hex.EncodeToString(ceConfigs[1].PrivateKey.Public().Bytes()))
	assert.NoError(t, err)

	// ensure that the block is committed
	require.Eventually(t, func() bool {
		height := n1.lastCommitHeight()
		fmt.Printf("Height: %d\n", height)
		return height == 1
	}, 2*time.Second, 100*time.Millisecond)

	ctx.Done()
}

func TestCELeaderTwoNodesMajorityNacks(t *testing.T) {
	// Majority > n/2 -> 2
	ceConfigs := generateTestCEConfig(t, 3)

	// bring up the nodes
	n1 := New(ceConfigs[0])
	// start node 1 (Leader)
	ctx := context.Background()
	go n1.Start(ctx, mockBlockPropBroadcaster, mockBlkAnnouncer, mockVoteBroadcaster, mockBlockRequester, mockResetStateBroadcaster)

	require.Eventually(t, func() bool {
		blockRes := n1.blockResult()
		return blockRes != nil && !blockRes.appHash.IsZero()
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
	err = n1.addVote(ctx, vote, hex.EncodeToString(ceConfigs[1].PrivateKey.Public().Bytes()))
	assert.NoError(t, err)
	err = n1.addVote(ctx, vote, hex.EncodeToString(ceConfigs[2].PrivateKey.Public().Bytes()))
	assert.NoError(t, err)

	// node should not commit the block and halt
	time.Sleep(500 * time.Millisecond)
	assert.Equal(t, n1.lastCommitHeight(), int64(0))
	fmt.Println("is Halt channel closed")
	_, ok := <-n1.haltChan
	assert.False(t, ok)
	fmt.Println("Halt channel closed")
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

type dummyTxApp struct {
	changesets []types.Hash
	vals       []*ktypes.Validator
}

func newDummyTxApp(valset []*ktypes.Validator) *dummyTxApp {
	return &dummyTxApp{
		vals: valset,
	}
}
func (d *dummyTxApp) Begin(ctx context.Context, height int64) error {
	return nil
}

func (d *dummyTxApp) Execute(ctx *ktypes.TxContext, db sql.DB, tx *ktypes.Transaction) *txapp.TxResponse {
	return &txapp.TxResponse{}
}

func (d *dummyTxApp) Finalize(ctx context.Context, db sql.DB, block *ktypes.BlockContext) ([]*ktypes.Validator, error) {
	return d.vals, nil
}

func (d *dummyTxApp) Commit() error {
	return nil
}

type validatorStore struct {
	valSet []*ktypes.Validator
}

func newValidatorStore(valSet []*ktypes.Validator) *validatorStore {
	return &validatorStore{
		valSet: valSet,
	}
}

func (v *validatorStore) GetValidators() []*ktypes.Validator {
	return v.valSet
}

func (v *validatorStore) ValidatorUpdates() map[string]*ktypes.Validator {
	return nil
}

type mockAccounts struct{}

func (m *mockAccounts) Updates() []*ktypes.Account {
	return nil
}
