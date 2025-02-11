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
	"path/filepath"
	"slices"
	"sync"
	"time"

	ethAccounts "github.com/ethereum/go-ethereum/accounts"
	ethCommon "github.com/ethereum/go-ethereum/common"
	ethMath "github.com/ethereum/go-ethereum/common/math"
	ethCrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/kwilteam/kwil-db/core/client"
	clientType "github.com/kwilteam/kwil-db/core/client/types"
	"github.com/kwilteam/kwil-db/core/crypto"
	"github.com/kwilteam/kwil-db/core/crypto/auth"
	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/node/exts/erc20reward/reward"
)

// StateFilePath returns the state file.
func StateFilePath(dir string) string {
	return filepath.Join(dir, "erc20reward_signer_state.json")
}

// rewardSigner handles one registered erc20 reward instance.
type rewardSigner struct {
	kwilRpc       string
	target        string
	kwil          erc20ExtAPI
	lastVoteBlock int64
	escrowAddr    ethCommon.Address

	ethRpc      string
	signerPkStr string
	signerPk    *ecdsa.PrivateKey
	signerAddr  ethCommon.Address
	safe        *Safe

	logger log.Logger
	every  time.Duration
	state  *State
}

// newRewardSigner returns a new rewardSigner.
func newRewardSigner(kwilRpc string, target string, ethRpc string, pkStr string,
	every time.Duration, state *State, logger log.Logger) (*rewardSigner, error) {
	if logger == nil {
		logger = log.DiscardLogger
	}

	privateKey, err := ethCrypto.HexToECDSA(pkStr)
	if err != nil {
		return nil, err
	}

	// Get the public key
	publicKey := privateKey.Public().(*ecdsa.PublicKey)

	// Get the Ethereum address from the public key
	address := ethCrypto.PubkeyToAddress(*publicKey)

	return &rewardSigner{
		kwilRpc:     kwilRpc,
		ethRpc:      ethRpc,
		signerPkStr: pkStr,
		signerPk:    privateKey,
		signerAddr:  address,
		state:       state,
		logger:      logger,
		every:       every,
		target:      target,
	}, nil
}

func (s *rewardSigner) init() error {
	ctx := context.Background()

	pkBytes, err := hex.DecodeString(s.signerPkStr)
	if err != nil {
		return fmt.Errorf("decode erc20 reward signer private key failed: %w", err)
	}

	key, err := crypto.UnmarshalSecp256k1PrivateKey(pkBytes)
	if err != nil {
		return fmt.Errorf("parse erc20 reward signer private key failed: %w", err)
	}

	opts := &clientType.Options{Signer: &auth.EthPersonalSigner{Key: *key}}

	clt, err := client.NewClient(ctx, s.kwilRpc, opts)
	if err != nil {
		return fmt.Errorf("create erc20 reward signer api client failed: %w", err)
	}

	s.kwil = newERC20RWExtAPI(clt, s.target)

	info, err := s.kwil.InstanceInfo(ctx)
	if err != nil {
		return fmt.Errorf("get reward metadata failed: %w", err)
	}

	s.safe, err = NewSafeFromEscrow(s.ethRpc, info.Escrow)
	if err != nil {
		return fmt.Errorf("create safe failed: %w", err)
	}

	if s.safe.chainID.String() != info.ChainID {
		return fmt.Errorf("chainID mismatch: %s != %s", s.safe.chainID.String(), info.ChainID)
	}

	s.escrowAddr = ethCommon.HexToAddress(info.Escrow)

	// overwrite configured lastVoteBlock with the value from state if exist
	lastVote := s.state.LastVote(s.target)
	if lastVote != nil {
		s.lastVoteBlock = lastVote.BlockHeight
	}

	s.logger.Info("will sync after last vote epoch", "height", s.lastVoteBlock)

	return nil
}

// canSkip returns true if the epoch:
// a) is not the owner b) is finalized c) is not finalized already voted from this signer;
func (s *rewardSigner) canSkip(epoch *Epoch, safeMeta *safeMetadata) bool {
	// TODO: if no rewards in epoch, skip

	if !slices.Contains(safeMeta.owners, s.signerAddr) {
		s.logger.Warn("signer is not safe owner", "signer", s.signerAddr.String(), "owners", safeMeta.owners)
		return true
	}

	//for _, voter := range epoch.Voters {
	//	if voter == s.signerAddr.String() {
	//		return true
	//	}
	//}

	return false
}

// verify verifies if the reward root is correct, and return the total amount.
func (s *rewardSigner) verify(ctx context.Context, epoch *Epoch, safeAddr string) (*big.Int, error) {
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
	copy(b32[:], epoch.BlockHash)

	_, root, err := reward.GenRewardMerkleTree(recipients, amounts, safeAddr, b32)
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
	safeTxData, err := reward.GenPostRewardTxData(epoch.RewardRoot, total)
	if err != nil {
		return err
	}

	// safeTxHash is the data that all signers will be signing(using personal_sign)
	_, safeTxHash, err := reward.GenGnosisSafeTx(s.escrowAddr.String(), s.safe.addr.String(),
		0, safeTxData, ethMath.HexOrDecimal256(*s.safe.chainID), *safeMeta.nonce)
	if err != nil {
		return err
	}

	signHash := ethAccounts.TextHash(safeTxHash)
	sig, err := reward.EthGnosisSignDigest(signHash, s.signerPk)
	if err != nil {
		return err
	}

	h, err := s.kwil.VoteEpoch(ctx, epoch.RewardRoot, sig)
	if err != nil {
		return err
	}

	// NOTE: it's fine if s.kwil.VoteEpoch succeed, but s.state.UpdateLastVote failed,
	// as the epoch will be fetched again and skipped
	err = s.state.UpdateLastVote(s.target, &voteRecord{
		RewardRoot:  epoch.RewardRoot,
		BlockHeight: epoch.EndHeight,
		BlockHash:   hex.EncodeToString(epoch.BlockHash),
		SafeNonce:   safeMeta.nonce.Uint64(),
	})
	if err != nil {
		return err
	}

	s.logger.Info("vote epoch", "tx", h, "id", epoch.ID.String(),
		"signHash", hex.EncodeToString(signHash))

	return nil
}

// watch polls on newer epochs and try to vote/sign them.
// Since there could be the case that the target(namespace/or id) not exist for whatever reason,
// this function won't return Error, and also won't log at Error level.
func (s *rewardSigner) watch(ctx context.Context) {
	s.logger.Info("start watching erc20 reward epoches")

	tick := time.NewTicker(s.every)

	for {
		s.logger.Debug("polling epochs", "lastVoteBlock", s.lastVoteBlock)
		// fetch next batch rewards to be voted, and vote them.
		// NOTE: we use ListUnconfirmedEpochs (not FetchLatestRewards) so we don't accidently SKIP epoch.
		epochs, err := s.kwil.ListUnconfirmedEpochs(ctx, s.lastVoteBlock, 10)
		if err != nil {
			s.logger.Warn("fetch epoch", "error", err.Error())
			continue
		}

		if len(epochs) == 0 {
			s.logger.Debug("no epoch found")
			continue
		}

		safeMeta, err := s.safe.latestMetadata(ctx)
		if err != nil {
			s.logger.Warn("fetch safe metadata", "error", err.Error())
			continue
		}

		for _, epoch := range epochs {
			voteRecord := s.state.LastVote(s.target)
			if voteRecord != nil && voteRecord.SafeNonce == safeMeta.nonce.Uint64() {
				continue
			}

			if s.canSkip(epoch, safeMeta) {
				s.logger.Debug("skip epoch", "id", epoch.ID.String(), "height", epoch.EndHeight)
				s.lastVoteBlock = epoch.EndHeight // update since we can skip it
				continue
			}

			total, err := s.verify(ctx, epoch, s.safe.addr.String())
			if err != nil {
				s.logger.Warn("verify epoch", "id", epoch.ID.String(), "height", epoch.EndHeight, "error", err.Error())
				break
			}

			err = s.vote(ctx, epoch, safeMeta, total)
			if err != nil {
				s.logger.Warn("vote epoch", "id", epoch.ID.String(), "height", epoch.EndHeight, "error", err.Error())
				break
			}

			s.lastVoteBlock = epoch.EndHeight // update after all operations succeed
		}

		select {
		case <-ctx.Done():
			s.logger.Info("stop watching erc20 reward epoches")
			return
		case <-tick.C:
			continue
		}
	}
}

// ServiceMgr manages multiple rewardSigner instances running in parallel.
type ServiceMgr struct {
	signers []*rewardSigner
	logger  log.Logger
}

func NewServiceMgr(
	kwilRpc string,
	targets []string,
	ethRpcs []string,
	pkStrs []string,
	syncEvery time.Duration,
	state *State,
	logger log.Logger) (*ServiceMgr, error) {

	signers := make([]*rewardSigner, len(targets))
	for i, target := range targets {
		pk := pkStrs[i]
		svc, err := newRewardSigner(kwilRpc, target, ethRpcs[i], pk,
			syncEvery, state, logger.New("EVMRW."+target))
		if err != nil {
			return nil, fmt.Errorf("create erc20 reward signer service failed: %w", err)
		}

		signers[i] = svc
	}

	return &ServiceMgr{
		signers: signers,
		logger:  logger,
	}, nil
}

// Start runs all rewardSigners. It returns error if there are issues initializing the rewardSigner;
// no errors are returned after the rewardSigner is running.
func (s *ServiceMgr) Start(ctx context.Context) error {
	// since we need to wait on RPC running, we move the initialization logic into `init`
	for _, s := range s.signers {
		err := s.init()
		if err != nil {
			return err
		}
	}

	wg := &sync.WaitGroup{}

	for _, s := range s.signers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.watch(ctx)
		}()
	}

	<-ctx.Done()
	wg.Wait()

	s.logger.Infof("Erc20 reward signer service shutting down...")

	return nil
}
