// Package signersvc implements the SignerSvc of the Kwil reward system.
// It simply fetches the new Epoch from Kwil network and verify&sign it, then
// upload the signature back to the Kwil network. Each rewardSigner targets one registered
// erc20 Reward instance.
package signersvc

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	ethAccounts "github.com/ethereum/go-ethereum/accounts"
	ethCommon "github.com/ethereum/go-ethereum/common"
	ethCrypto "github.com/ethereum/go-ethereum/crypto"

	"github.com/kwilteam/kwil-db/config"
	"github.com/kwilteam/kwil-db/core/client"
	clientType "github.com/kwilteam/kwil-db/core/client/types"
	"github.com/kwilteam/kwil-db/core/crypto"
	"github.com/kwilteam/kwil-db/core/crypto/auth"
	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/node/exts/erc20-bridge/utils"
	"github.com/kwilteam/kwil-db/node/exts/evm-sync/chains"
)

// StateFilePath returns the state file.
func StateFilePath(dir string) string {
	return filepath.Join(dir, "erc20_signer_state.json")
}

// rewardSigner handles one registered erc20 reward instance.
type rewardSigner struct {
	target        string
	kwil          erc20ExtAPI
	lastVoteBlock int64
	escrowAddr    ethCommon.Address

	signerPk   *ecdsa.PrivateKey
	signerAddr ethCommon.Address
	safe       *Safe

	logger log.Logger
	state  *State
}

// newRewardSigner returns a new rewardSigner.
func newRewardSigner(kwil erc20ExtAPI, safe *Safe, target string, signerPk *ecdsa.PrivateKey, signerAddr ethCommon.Address, escrowAddr ethCommon.Address, state *State, logger log.Logger) (*rewardSigner, error) {
	if logger == nil {
		logger = log.DiscardLogger
	}

	// overwrite configured lastVoteBlock with the value from state if exist
	lastVoteBlock := int64(0)
	lastVote := state.LastVote(target)
	if lastVote != nil {
		lastVoteBlock = lastVote.BlockHeight
	}

	logger.Info("will sync after last vote epoch", "height", lastVoteBlock)

	return &rewardSigner{
		kwil:          kwil,
		signerPk:      signerPk,
		signerAddr:    signerAddr,
		state:         state,
		logger:        logger,
		target:        target,
		safe:          safe,
		escrowAddr:    escrowAddr,
		lastVoteBlock: lastVoteBlock,
	}, nil
}

// canSkip returns true if:
// - signer is not one of the safe owners
// - signer has voted this epoch with the same nonce as current safe nonce
func (s *rewardSigner) canSkip(epoch *Epoch, safeMeta *safeMetadata) bool {
	if !slices.Contains(safeMeta.owners, s.signerAddr) {
		s.logger.Warn("signer is not safe owner", "signer", s.signerAddr.String(), "owners", safeMeta.owners)
		return true
	}

	if epoch.Voters == nil {
		return false
	}

	for i, voter := range epoch.Voters {
		if voter == s.signerAddr.String() &&
			safeMeta.nonce.Cmp(big.NewInt(epoch.VoteNonces[i])) == 0 {
			return true
		}
	}

	return false
}

// verify verifies if the reward root is correct, and return the total amount.
func (s *rewardSigner) verify(ctx context.Context, epoch *Epoch, escrowAddr string) (*big.Int, error) {
	rewards, err := s.kwil.GetEpochRewards(ctx, epoch.ID)
	if err != nil {
		return nil, err
	}

	recipients := make([]string, len(rewards))
	amounts := make([]*big.Int, len(rewards))

	var ok bool
	total := big.NewInt(0)
	for i, r := range rewards {
		recipients[i] = r.Recipient

		amounts[i], ok = new(big.Int).SetString(r.Amount, 10)
		if !ok {
			return nil, fmt.Errorf("parse reward amount %s failed", r.Amount)
		}

		total = total.Add(total, amounts[i])
	}

	var b32 [32]byte
	copy(b32[:], epoch.EndBlockHash)

	_, root, err := utils.GenRewardMerkleTree(recipients, amounts, escrowAddr, b32)
	if err != nil {
		return nil, err
	}

	if !slices.Equal(root, epoch.RewardRoot) {
		return nil, fmt.Errorf("reward root mismatch: %s != %s", hex.EncodeToString(root), hex.EncodeToString(epoch.RewardRoot))
	}

	s.logger.Info("verified epoch", "id", epoch.ID.String(), "rewardRoot", hex.EncodeToString(epoch.RewardRoot))
	return total, nil
}

// vote votes an epoch reward, and updates the state.
// It will first fetch metadata from ETH, then generate the safeTx, then vote.
func (s *rewardSigner) vote(ctx context.Context, epoch *Epoch, safeMeta *safeMetadata, total *big.Int) error {
	safeTxData, err := utils.GenPostRewardTxData(epoch.RewardRoot, total)
	if err != nil {
		return err
	}

	// safeTxHash is the data that all signers will be signing(using personal_sign)
	_, safeTxHash, err := utils.GenGnosisSafeTx(s.escrowAddr.String(), s.safe.addr.String(),
		0, safeTxData, s.safe.chainID.Int64(), safeMeta.nonce.Int64())
	if err != nil {
		return err
	}

	signHash := ethAccounts.TextHash(safeTxHash)
	sig, err := utils.EthGnosisSignDigest(signHash, s.signerPk)
	if err != nil {
		return err
	}

	h, err := s.kwil.VoteEpoch(ctx, epoch.ID, safeMeta.nonce.Int64(), sig)
	if err != nil {
		return err
	}

	// NOTE: it's fine if s.kwil.VoteEpoch succeed, but s.state.UpdateLastVote failed,
	// as the epoch will be fetched again and skipped
	err = s.state.UpdateLastVote(s.target, &voteRecord{
		RewardRoot:  epoch.RewardRoot,
		BlockHeight: epoch.EndHeight,
		BlockHash:   hex.EncodeToString(epoch.EndBlockHash),
		SafeNonce:   safeMeta.nonce.Uint64(),
	})
	if err != nil {
		return err
	}

	s.logger.Info("vote epoch", "tx", h, "id", epoch.ID.String(),
		"nonce", safeMeta.nonce.Int64())

	return nil
}

// sync polls on newer epochs and try to vote/sign them.
// Since there could be the case that the target(namespace/or id) not exist for whatever reason,
// this function won't return Error, and also won't log at Error level.
func (s *rewardSigner) sync(ctx context.Context) {
	s.logger.Debug("polling epochs", "lastVoteBlock", s.lastVoteBlock)

	epochs, err := s.kwil.GetActiveEpochs(ctx)
	if err != nil {
		s.logger.Warn("fetch epoch", "error", err.Error())
		return
	}

	if len(epochs) == 0 {
		s.logger.Error("no epoch found")
		return
	}

	if len(epochs) == 1 {
		// the very first round of epoch, we wait until there are 2 active epochs
		return
	}

	if len(epochs) != 2 {
		s.logger.Error("unexpected number of epochs", "count", len(epochs))
		return
	}

	finalizedEpoch := epochs[0]

	safeMeta, err := s.safe.latestMetadata(ctx)
	if err != nil {
		s.logger.Warn("fetch safe metadata", "error", err.Error())
		return
	}

	if s.canSkip(finalizedEpoch, safeMeta) {
		s.logger.Info("skip epoch", "id", finalizedEpoch.ID.String(), "height", finalizedEpoch.EndHeight)
		s.lastVoteBlock = finalizedEpoch.EndHeight // update since we can skip it
		return
	}

	total, err := s.verify(ctx, finalizedEpoch, s.escrowAddr.String())
	if err != nil {
		s.logger.Warn("verify epoch", "id", finalizedEpoch.ID.String(), "height", finalizedEpoch.EndHeight, "error", err.Error())
		return
	}

	err = s.vote(ctx, finalizedEpoch, safeMeta, total)
	if err != nil {
		s.logger.Warn("vote epoch", "id", finalizedEpoch.ID.String(), "height", finalizedEpoch.EndHeight, "error", err.Error())
		return
	}

	s.lastVoteBlock = finalizedEpoch.EndHeight // update after all operations succeed
}

// getSigners verifies config and returns a list of signerSvc.
func getSigners(kwilRpc string, cfg config.ERC20BridgeConfig, state *State, logger log.Logger) ([]*rewardSigner, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	ctx := context.Background()

	clt, err := client.NewClient(ctx, kwilRpc, nil)
	if err != nil {
		return nil, fmt.Errorf("create erc20 bridge signer api client failed: %w", err)
	}

	signers := make([]*rewardSigner, 0, len(cfg.Signer))
	for target, pkPath := range cfg.Signer {
		// pkPath is validated already

		readOnlyAPi := newERC20RWExtAPI(clt, target)
		instanceInfo, err := readOnlyAPi.InstanceInfo(ctx)
		if err != nil {
			return nil, fmt.Errorf("get reward metadata failed: %w", err)
		}

		rawPkBytes, err := os.ReadFile(pkPath)
		if err != nil {
			return nil, fmt.Errorf("read private key file %s failed: %w", pkPath, err)
		}

		pkStr := strings.TrimSpace(string(rawPkBytes))
		pkBytes, err := hex.DecodeString(pkStr)
		if err != nil {
			return nil, fmt.Errorf("parse erc20 bridge signer private key failed: %w", err)
		}

		signerPk, err := ethCrypto.ToECDSA(pkBytes)
		if err != nil {
			return nil, fmt.Errorf("parse erc20 bridge signer private key failed: %w", err)
		}

		// Get the public key
		signerPubKey := signerPk.Public().(*ecdsa.PublicKey)

		// Get the Ethereum address from the public key
		signerAddr := ethCrypto.PubkeyToAddress(*signerPubKey)

		key, err := crypto.UnmarshalSecp256k1PrivateKey(pkBytes)
		if err != nil {
			return nil, fmt.Errorf("parse erc20 bridge signer private key failed: %w", err)
		}

		opts := &clientType.Options{Signer: &auth.EthPersonalSigner{Key: *key}}

		clt, err := client.NewClient(ctx, kwilRpc, opts)
		if err != nil {
			return nil, fmt.Errorf("create erc20 bridge signer api client failed: %w", err)
		}

		kwil := newERC20RWExtAPI(clt, target)

		chainRpc, ok := cfg.RPC[strings.ToLower(instanceInfo.Chain)]
		if !ok {
			return nil, fmt.Errorf("target '%s' chain '%s' not found in erc20_bridge.rpc config", target, instanceInfo.Chain)
		}

		safe, err := NewSafeFromEscrow(chainRpc, instanceInfo.Escrow)
		if err != nil {
			return nil, fmt.Errorf("create safe failed: %w", err)
		}

		chainInfo, ok := chains.GetChainInfo(chains.Chain(instanceInfo.Chain))
		if !ok {
			return nil, fmt.Errorf("chainID %s not supported", safe.chainID.String())
		}

		if safe.chainID.String() != chainInfo.ID {
			return nil, fmt.Errorf("chainID mismatch: configured %s != target %s", safe.chainID.String(), chainInfo.ID)
		}

		// wilRpc, target, chainRpc, strings.TrimSpace(string(pkBytes))
		svc, err := newRewardSigner(kwil, safe, target, signerPk, signerAddr, ethCommon.HexToAddress(instanceInfo.Escrow), state, logger.New("EVMRW."+target))
		if err != nil {
			return nil, fmt.Errorf("create erc20 bridge signer service failed: %w", err)
		}

		signers = append(signers, svc)
	}

	return signers, nil
}

// ServiceMgr manages multiple rewardSigner instances running in parallel.
type ServiceMgr struct {
	kwilRpc      string
	state        *State
	bridgeCfg    config.ERC20BridgeConfig
	syncInterval time.Duration
	logger       log.Logger
}

func NewServiceMgr(
	kwilRpc string,
	cfg config.ERC20BridgeConfig,
	state *State,
	logger log.Logger) (*ServiceMgr, error) {
	return &ServiceMgr{
		kwilRpc:      kwilRpc,
		state:        state,
		bridgeCfg:    cfg,
		logger:       logger,
		syncInterval: time.Minute, // default to 1m
	}, nil
}

// Start runs all rewardSigners. It returns error if there are issues initializing the rewardSigner;
// no errors are returned after the rewardSigner is running.
func (m *ServiceMgr) Start(ctx context.Context) error {
	// since we need to wait on RPC running, we move the initialization logic into `init`

	var err error
	var signers []*rewardSigner
	// To be able to run with docker, we need to apply a retry logic, because kwild
	// won't have erc20 instance when boot
	for { // naive way to keep retrying the init
		select {
		case <-ctx.Done():
			m.logger.Info("stop initializing erc20 bridge signer")
			return nil
		default:
		}

		signers, err = getSigners(m.kwilRpc, m.bridgeCfg, m.state, m.logger)
		if err == nil {
			break
		}

		m.logger.Warn("failed to initialize erc20 bridge signer, will retry", "error", err.Error())
		// any error, we try again
		time.Sleep(time.Second * 3)
	}

	wg := &sync.WaitGroup{}

	for _, s := range signers {
		wg.Add(1)
		go func() {
			defer wg.Done()

			s.logger.Info("start watching erc20 reward epoches")
			tick := time.NewTicker(m.syncInterval)

			for {
				s.sync(ctx)

				select {
				case <-ctx.Done():
					s.logger.Info("stop watching erc20 reward epoches")
					return
				case <-tick.C:
				}
			}
		}()
	}

	<-ctx.Done()
	wg.Wait()

	m.logger.Infof("Erc20 bridge signer service shutting down...")

	return nil
}
