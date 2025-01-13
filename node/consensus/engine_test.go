//go:build pglive

package consensus

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/config"
	"github.com/kwilteam/kwil-db/core/crypto"
	"github.com/kwilteam/kwil-db/core/crypto/auth"
	ktypes "github.com/kwilteam/kwil-db/core/types"
	blockprocessor "github.com/kwilteam/kwil-db/node/block_processor"
	"github.com/kwilteam/kwil-db/node/mempool"
	"github.com/kwilteam/kwil-db/node/meta"
	"github.com/kwilteam/kwil-db/node/migrations"
	"github.com/kwilteam/kwil-db/node/pg"
	dbtest "github.com/kwilteam/kwil-db/node/pg/test"
	"github.com/kwilteam/kwil-db/node/snapshotter"
	"github.com/kwilteam/kwil-db/node/store"
	"github.com/kwilteam/kwil-db/node/txapp"
	"github.com/kwilteam/kwil-db/node/types"
	"github.com/kwilteam/kwil-db/node/types/sql"
	"github.com/kwilteam/kwil-db/node/voting"
	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/kwilteam/kwil-db/core/log"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	broadcastFns = BroadcastFns{
		ProposalBroadcaster:     mockBlockPropBroadcaster,
		TxAnnouncer:             mockTxAnnouncer,
		BlkAnnouncer:            mockBlkAnnouncer,
		BlkRequester:            mockBlkRequester,
		AckBroadcaster:          mockVoteBroadcaster,
		RstStateBroadcaster:     mockResetStateBroadcaster,
		DiscoveryReqBroadcaster: mockDiscoveryBroadcaster,
		TxBroadcaster:           nil,
	}
)

// leaderDB is set assigns DB to the leader, else DB is assigned to the follower
// Most of these tests expect only one working node instance either a leader or the
// first validator and all other nodes interactions are mocked out.
func generateTestCEConfig(t *testing.T, nodes int, leaderDB bool) []*Config {
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
			PubKey:     types.HexBytes(pubKey.Bytes()),
			PubKeyType: pubKey.Type(),
			Power:      1,
		}
		validatorSet[hex.EncodeToString(pubKey.Bytes())] = *val
		valSet = append(valSet, val)
	}

	ctx := context.Background()

	// Account Store
	// accounts, err := accounts.InitializeAccountStore(ctx, db)
	// assert.NoError(t, err)
	accounts := &mockAccounts{}

	v := newValidatorStore(valSet)

	// txapp, err := txapp.NewTxApp(ctx, db, nil, signer, nil, service, accounts, v)
	// assert.NoError(t, err)
	txapp := newDummyTxApp(valSet)

	db := dbtest.NewTestDB(t, func(db *pg.DB) {
		db.AutoCommit(true)
		ctx := context.Background()
		db.Execute(ctx, `DROP SCHEMA IF EXISTS kwild_chain CASCADE;`)
		db.Execute(ctx, `DROP SCHEMA IF EXISTS kwild_voting CASCADE;`)
		db.Execute(ctx, `DROP SCHEMA IF EXISTS kwild_events CASCADE;`)
		db.Execute(ctx, `DROP SCHEMA IF EXISTS kwild_internal CASCADE;`)
		db.AutoCommit(false)
	})

	_, _, err := voting.NewResolutionStore(ctx, db) // create the voting resolution store
	require.NoError(t, err)

	func() {
		tx, err := db.BeginTx(ctx)
		require.NoError(t, err)
		defer tx.Rollback(ctx)

		err = meta.InitializeMetaStore(ctx, tx)
		require.NoError(t, err)

		require.NoError(t, tx.Commit(ctx))
	}()

	ss := &snapshotStore{}

	genCfg := config.DefaultGenesisConfig()
	genCfg.Leader = config.EncodePubKeyAndType(pubKeys[0].Bytes(), pubKeys[0].Type())

	for i := range nodes {
		nodeStr := fmt.Sprintf("NODE%d", i)
		nodeDir := filepath.Join(tempDir, nodeStr)

		logger := log.New(log.WithName(nodeStr), log.WithWriter(os.Stdout), log.WithLevel(log.LevelDebug), log.WithFormat(log.FormatUnstructured))
		// logger := log.DiscardLogger

		bs, err := store.NewBlockStore(nodeDir)
		assert.NoError(t, err)

		signer := auth.GetNodeSigner(privKeys[i])

		ev := &mockEventStore{}
		m := &mockMigrator{}

		bp, err := blockprocessor.NewBlockProcessor(ctx, db, txapp, accounts, v, ss, ev, m, bs, genCfg, signer, log.New(log.WithName("BP")))
		assert.NoError(t, err)
		bp.SetNetworkParameters(&common.NetworkParameters{
			MaxBlockSize:     genCfg.MaxBlockSize,
			JoinExpiry:       genCfg.JoinExpiry,
			VoteExpiry:       genCfg.VoteExpiry,
			DisabledGasCosts: true,
			MaxVotesPerTx:    genCfg.MaxVotesPerTx,
		})

		ceConfigs[i] = &Config{
			PrivateKey:            privKeys[i],
			Leader:                pubKeys[0],
			Mempool:               mempool.New(),
			BlockStore:            bs,
			BlockProcessor:        bp,
			ValidatorSet:          validatorSet,
			Logger:                logger,
			ProposeTimeout:        1 * time.Second,
			BlockProposalInterval: 1 * time.Second,
			BlockAnnInterval:      3 * time.Second,
			BroadcastTxTimeout:    10 * time.Second,
		}

		closers = append(closers, func() {
			bs.Close()
		})
	}

	if leaderDB {
		ceConfigs[0].DB = db
	} else {
		ceConfigs[1].DB = db
	}

	t.Cleanup(func() {
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

func TestMain(m *testing.M) {
	pg.UseLogger(log.New(log.WithName("DBS"), log.WithFormat(log.FormatUnstructured)))
	m.Run()
}

func addVotes(t *testing.T, blkHash, appHash ktypes.Hash, n1, n2 *ConsensusEngine) *ktypes.CommitInfo {
	ci := &ktypes.CommitInfo{
		AppHash: appHash,
		Votes:   make([]*ktypes.VoteInfo, 0),
	}

	sig1, err := ktypes.SignVote(blkHash, true, &appHash, n1.privKey)
	require.NoError(t, err)

	ci.Votes = append(ci.Votes, &ktypes.VoteInfo{
		Signature: *sig1,
		AckStatus: ktypes.AckStatusAgree,
	})

	sig2, err := ktypes.SignVote(blkHash, true, &appHash, n2.privKey)
	require.NoError(t, err)

	ci.Votes = append(ci.Votes, &ktypes.VoteInfo{
		Signature: *sig2,
		AckStatus: ktypes.AckStatusAgree,
	})

	return ci
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
				return generateTestCEConfig(t, 2, false)
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
						ci := addVotes(t, blkProp1.blkHash, blockAppHash, leader, val)
						val.NotifyBlockCommit(blkProp1.blk, ci)
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
				return generateTestCEConfig(t, 2, false)
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
						val.NotifyBlockCommit(blkProp1.blk, &ktypes.CommitInfo{AppHash: ktypes.Hash{}})
					},
					verify: func(t *testing.T, leader, val *ConsensusEngine) error {
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
				return generateTestCEConfig(t, 2, false)
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
						val.NotifyBlockProposal(blkProp2.blk)
					},
					verify: func(t *testing.T, leader, val *ConsensusEngine) error {
						return verifyStatus(t, val, Executed, 0, blkProp2.blkHash)
					},
				},
				{
					name: "commitNew",
					trigger: func(t *testing.T, leader, val *ConsensusEngine) {
						ci := addVotes(t, blkProp2.blkHash, blockAppHash, leader, val)
						val.NotifyBlockCommit(blkProp2.blk, ci)
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
				return generateTestCEConfig(t, 2, false)
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
						ci := addVotes(t, blkProp2.blkHash, blockAppHash, leader, val)
						val.NotifyBlockCommit(blkProp2.blk, ci)
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
				return generateTestCEConfig(t, 2, false)
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
						ci := addVotes(t, blkProp2.blkHash, blockAppHash, leader, val)
						val.NotifyBlockCommit(blkProp2.blk, ci)
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
				return generateTestCEConfig(t, 2, false)
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
						val.sendResetMsg(&resetMsg{height: 0})
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
						ci := addVotes(t, blkProp2.blkHash, blockAppHash, leader, val)
						val.NotifyBlockCommit(blkProp2.blk, ci)
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
				return generateTestCEConfig(t, 2, false)
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
						ci := addVotes(t, blkProp2.blkHash, blockAppHash, leader, val)
						val.NotifyBlockCommit(blkProp2.blk, ci)
					},
					verify: func(t *testing.T, leader, val *ConsensusEngine) error {
						return verifyStatus(t, val, Committed, 1, zeroHash)
					},
				},
				{
					name: "reset",
					trigger: func(t *testing.T, leader, val *ConsensusEngine) {
						val.sendResetMsg(&resetMsg{height: 0})
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
				return generateTestCEConfig(t, 2, false)
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
						val.sendResetMsg(&resetMsg{height: 0})
					},
					verify: func(t *testing.T, leader, val *ConsensusEngine) error {
						return verifyStatus(t, val, Committed, 0, zeroHash)
					},
				},
				{
					name: "reset",
					trigger: func(t *testing.T, leader, val *ConsensusEngine) {
						val.sendResetMsg(&resetMsg{height: 0})
					},
					verify: func(t *testing.T, leader, val *ConsensusEngine) error {
						return verifyStatus(t, val, Committed, 0, zeroHash)
					},
				},
				{
					name: "commitNew",
					trigger: func(t *testing.T, leader, val *ConsensusEngine) {
						ci := addVotes(t, blkProp2.blkHash, blockAppHash, leader, val)
						val.NotifyBlockCommit(blkProp2.blk, ci)
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
				return generateTestCEConfig(t, 2, false)
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
						val.sendResetMsg(&resetMsg{height: 1})
					},
					verify: func(t *testing.T, leader, val *ConsensusEngine) error {
						return verifyStatus(t, val, Executed, 0, blkProp1.blkHash)
					},
				},
				{
					name: "reset",
					trigger: func(t *testing.T, leader, val *ConsensusEngine) {
						val.sendResetMsg(&resetMsg{height: 2})
					},
					verify: func(t *testing.T, leader, val *ConsensusEngine) error {
						return verifyStatus(t, val, Executed, 0, blkProp1.blkHash)
					},
				},
				{
					name: "commitNew",
					trigger: func(t *testing.T, leader, val *ConsensusEngine) {
						ci := addVotes(t, blkProp2.blkHash, blockAppHash, leader, val)
						val.NotifyBlockCommit(blkProp2.blk, ci)
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

			ctxM := context.Background()
			blkProp1, err = leader.createBlockProposal(ctxM)
			assert.NoError(t, err)
			time.Sleep(300 * time.Millisecond) // just to ensure that the block hashes are different due to start time
			blkProp2, err = leader.createBlockProposal(ctxM)
			assert.NoError(t, err)
			t.Logf("blkProp1: %s, blkProp2: %s", blkProp1.blkHash.String(), blkProp2.blkHash.String())

			ctx, cancel := context.WithCancel(ctxM)
			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				defer wg.Done()
				val.Start(ctx, broadcastFns)
			}()

			t.Cleanup(func() {
				cancel()
				wg.Wait()
			})

			for _, act := range tc.actions {
				t.Log("action", act.name)
				act.trigger(t, leader, val)
				require.Eventually(t, func() bool {
					err := act.verify(t, leader, val)
					if err != nil {
						t.Log(err)
						return false
					}
					return true
				}, 6*time.Second, 100*time.Millisecond)
			}
		})
	}
}

func TestCELeaderSingleNode(t *testing.T) {
	// t.Parallel()
	ceConfigs := generateTestCEConfig(t, 1, true)

	// bring up the node
	leader := New(ceConfigs[0])

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		leader.Start(ctx, broadcastFns)
	}()

	t.Cleanup(func() {
		cancel()
		wg.Wait()
	})

	require.Eventually(t, func() bool {
		return leader.lastCommitHeight() >= 1 // Ensure that the leader mines a block
	}, 10*time.Second, 100*time.Millisecond)
}

func TestCELeaderTwoNodesMajorityAcks(t *testing.T) {
	// Majority > n/2 -> 2
	ceConfigs := generateTestCEConfig(t, 2, true)

	// bring up the nodes
	n1 := New(ceConfigs[0])
	// start node 1 (Leader)
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		n1.Start(ctx, broadcastFns)
	}()

	t.Cleanup(func() {
		cancel()
		wg.Wait()
	})

	n1.bestHeightCh <- &discoveryMsg{
		BestHeight: 0,
		Sender:     ceConfigs[1].PrivateKey.Public().Bytes(),
	}

	time.Sleep(500 * time.Millisecond)
	_, _, blProp := n1.info()
	// appHash := nextAppHash()

	require.NotNil(t, blProp)

	// node2 should send a vote to node1
	sig, err := ktypes.SignVote(blProp.blkHash, true, &blockAppHash, ceConfigs[1].PrivateKey)
	assert.NoError(t, err)

	vote := &vote{
		height:    1,
		ack:       true,
		blkHash:   blProp.blkHash,
		appHash:   &blockAppHash,
		signature: sig,
	}

	// Invalid sender
	err = n1.addVote(ctx, vote, "invalid")
	assert.Error(t, err)

	// Valid sender
	err = n1.addVote(ctx, vote, hex.EncodeToString(ceConfigs[1].PrivateKey.Public().Bytes()))
	assert.NoError(t, err)

	// ensure that the block is committed
	require.Eventually(t, func() bool {
		height := n1.lastCommitHeight()
		fmt.Printf("Height: %d\n", height)
		return height == 1
	}, 6*time.Second, 100*time.Millisecond)
}

func TestCELeaderTwoNodesMajorityNacks(t *testing.T) {
	// Majority > n/2 -> 2
	ceConfigs := generateTestCEConfig(t, 3, true)

	// bring up the nodes
	n1 := New(ceConfigs[0])
	// start node 1 (Leader)
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)

	t.Cleanup(func() {
		cancel()
		wg.Wait()
	})

	go func() {
		defer wg.Done()
		n1.Start(ctx, broadcastFns)
	}()

	n1.bestHeightCh <- &discoveryMsg{
		BestHeight: 0,
		Sender:     ceConfigs[1].PrivateKey.Public().Bytes(),
	}

	require.Eventually(t, func() bool {
		blockRes := n1.blockResult()
		return blockRes != nil && !blockRes.appHash.IsZero()
	}, 6*time.Second, 100*time.Millisecond)

	_, _, b := n1.info()
	assert.NotNil(t, b)
	nextAppHash := nextAppHash(nextAppHash(zeroHash))

	sig1, err := ktypes.SignVote(b.blkHash, true, &nextAppHash, ceConfigs[1].PrivateKey)
	assert.NoError(t, err)

	sig2, err := ktypes.SignVote(b.blkHash, true, &nextAppHash, ceConfigs[2].PrivateKey)
	assert.NoError(t, err)

	// node2 should send a vote to node1
	vote := &vote{
		height:    1,
		ack:       true,
		blkHash:   b.blkHash,
		appHash:   &nextAppHash,
		signature: sig1,
	}

	// Invalid sender -> vote ignored
	err = n1.addVote(ctx, vote, "invalid")
	assert.Error(t, err)

	// Valid sender
	err = n1.addVote(ctx, vote, hex.EncodeToString(ceConfigs[1].PrivateKey.Public().Bytes()))
	assert.NoError(t, err)

	vote.signature = sig2
	err = n1.addVote(ctx, vote, hex.EncodeToString(ceConfigs[2].PrivateKey.Public().Bytes()))
	assert.NoError(t, err)

	// node should not commit the block and halt
	time.Sleep(500 * time.Millisecond)
	assert.Equal(t, n1.lastCommitHeight(), int64(0))
}

// MockBroadcasters
func mockBlkRequester(ctx context.Context, height int64) (types.Hash, []byte, *ktypes.CommitInfo, error) {
	return types.Hash{}, nil, nil, fmt.Errorf("not implemented")
}

func mockBlockPropBroadcaster(_ context.Context, blk *ktypes.Block) {}

func mockVoteBroadcaster(ack bool, height int64, blkID types.Hash, appHash *types.Hash, sig []byte) error {
	return nil
}

func mockBlkAnnouncer(_ context.Context, blk *ktypes.Block, ci *ktypes.CommitInfo) {}

func mockTxAnnouncer(ctx context.Context, txHash types.Hash, rawTx []byte, from peer.ID) {}

func mockResetStateBroadcaster(_ int64, _ []ktypes.Hash) error {
	return nil
}

func mockDiscoveryBroadcaster() {}

func nextAppHash(prevHash types.Hash) types.Hash {
	hasher := sha256.New()
	txHash := types.Hash(hasher.Sum(nil))
	return sha256.Sum256(append(prevHash[:], txHash[:]...))
}

type dummyTxApp struct {
	vals []*ktypes.Validator
}

func newDummyTxApp(valset []*ktypes.Validator) *dummyTxApp {
	return &dummyTxApp{
		vals: valset,
	}
}
func (d *dummyTxApp) Begin(ctx context.Context, height int64) error {
	return nil
}

func (d *dummyTxApp) Execute(ctx *common.TxContext, db sql.DB, tx *ktypes.Transaction) *txapp.TxResponse {
	return &txapp.TxResponse{}
}

func (d *dummyTxApp) Finalize(ctx context.Context, db sql.DB, block *common.BlockContext) ([]*ktypes.Validator, error) {
	return d.vals, nil
}

func (d *dummyTxApp) Price(ctx context.Context, dbTx sql.DB, tx *ktypes.Transaction, chainContext *common.ChainContext) (*big.Int, error) {
	return big.NewInt(0), nil
}

func (d *dummyTxApp) Commit() error {
	return nil
}

func (d *dummyTxApp) Rollback() {}

func (d *dummyTxApp) GenesisInit(ctx context.Context, db sql.DB, validators []*ktypes.Validator, genesisAccounts []*ktypes.Account, initialHeight int64, dbOwner string, chain *common.ChainContext) error {
	return nil
}
func (d *dummyTxApp) AccountInfo(ctx context.Context, dbTx sql.DB, identifier string, pending bool) (balance *big.Int, nonce int64, err error) {
	return big.NewInt(0), 0, nil
}

func (d *dummyTxApp) ApplyMempool(ctx *common.TxContext, db sql.DB, tx *ktypes.Transaction) error {
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

func (v *validatorStore) LoadValidatorSet(ctx context.Context, db sql.Executor) error {
	return nil
}

type mockAccounts struct{}

func (m *mockAccounts) Updates() []*ktypes.Account {
	return nil
}

func (ce *ConsensusEngine) lastCommitHeight() int64 {
	ce.stateInfo.mtx.RLock()
	defer ce.stateInfo.mtx.RUnlock()

	return ce.stateInfo.height
}

func (ce *ConsensusEngine) info() (int64, Status, *blockProposal) {
	ce.stateInfo.mtx.RLock()
	defer ce.stateInfo.mtx.RUnlock()

	return ce.stateInfo.height, ce.stateInfo.status, ce.stateInfo.blkProp
}

func (ce *ConsensusEngine) blockResult() *blockResult {
	ce.state.mtx.RLock()
	defer ce.state.mtx.RUnlock()

	return ce.state.blockRes
}

// should satisfy the SnapshotModule interface
type snapshotStore struct {
}

func (s *snapshotStore) Enabled() bool {
	return false
}

func (s *snapshotStore) ListSnapshots() []*snapshotter.Snapshot {
	return nil
}

func (s *snapshotStore) CreateSnapshot(ctx context.Context, height uint64, snapshotID string, schemas, excludedTables []string, excludeTableData []string) error {
	return nil
}

func (s *snapshotStore) IsSnapshotDue(height uint64) bool {
	return false
}

func (s *snapshotStore) LoadSnapshotChunk(height uint64, format uint32, chunkID uint32) ([]byte, error) {
	return nil, nil
}

type mockEventStore struct {
	events []*ktypes.VotableEvent
}

func (m *mockEventStore) MarkBroadcasted(ctx context.Context, ids []*ktypes.UUID) error {
	return nil
}

func (m *mockEventStore) GetUnbroadcastedEvents(ctx context.Context) ([]*ktypes.UUID, error) {
	var ids []*ktypes.UUID
	for _, event := range m.events {
		ids = append(ids, event.ID())
	}
	return ids, nil
}

type mockMigrator struct{}

func (m *mockMigrator) NotifyHeight(ctx context.Context, block *common.BlockContext, db migrations.Database, tx sql.Executor) error {
	return nil
}

func (m *mockMigrator) StoreChangesets(height int64, changes <-chan any) error {
	return nil
}

func (m *mockMigrator) PersistLastChangesetHeight(ctx context.Context, tx sql.Executor, height int64) error {
	return nil
}

func (m *mockMigrator) GetMigrationMetadata(ctx context.Context, status ktypes.MigrationStatus) (*ktypes.MigrationMetadata, error) {
	return nil, nil
}
