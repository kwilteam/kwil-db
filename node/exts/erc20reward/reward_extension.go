package erc20reward

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/extensions/precompiles"
	"github.com/kwilteam/kwil-db/node/exts/erc20reward/meta"
	"github.com/kwilteam/kwil-db/node/exts/erc20reward/reward"
	"github.com/kwilteam/kwil-db/node/types/sql"
)

// This Kwil extension makes the following assumptions:
// - All TXs to GnosisSafe wallet are through Kwil, so that this ext can easily
//   increment the nonce(GnosisSafe and Reward contract) instead of using Oracle
//   to sync from outside. TODO: maybe we can add extra methods to update nonce.

// This is how this extension works:
// 1. Extension will be used ba a Kwil App(A Kuneiform app). App calls 'issue_rewards' to issue certain number of Rewards to a user.
// 2. An Epoch will be proposed so all issued rewards will be aggregated, and a merkle tree will be generated.
// 3. The SignerService run by different operators will vote(sign) an Epoch whenever it sees one. If one vote reaches the quota,
//    a FinalizedReward will be created.
// 4. The PosterService will post FinalizedReward.

const (
	uint256Precision = 78

	txIssueRewardCounterKey = "issue_reward_counter"
)

type chainInfo struct {
	Name      string
	Etherscan string
}

func (c chainInfo) GetEtherscanAddr(contract string) string {
	return c.Etherscan + contract + "#writeContract"
}

var chainConvMap = map[string]chainInfo{
	"1": {
		Name:      "Ethereum",
		Etherscan: "https://etherscan.io/address/",
	}, // TODO: we should not keep these hard-coded here
	"11155111": {
		Name:      "Sepolia",
		Etherscan: "https://sepolia.etherscan.io/address/",
	},
}

func init() {
	err := precompiles.RegisterInitializer("erc20_rewards",
		func(ctx context.Context, service *common.Service, db sql.DB, alias string, metadata map[string]any) (p precompiles.Precompile, err error) {
			chainID, contractAddr, contractNonce, signers, threshold, safeAddr, safeNonce, decimals, err := getMetadata(metadata)
			if err != nil {
				return p, fmt.Errorf("parse ext configuration: %w", err)
			}

			ext := &Erc20RewardExt{
				contractID:    meta.GenRewardContractID(chainID, contractAddr),
				alias:         alias,
				decimals:      decimals,
				ContractAddr:  contractAddr,
				SafeAddr:      safeAddr,
				ChainID:       chainID,
				Signers:       signers,
				Threshold:     threshold,
				ContractNonce: contractNonce,
				SafeNonce:     safeNonce,
			}

			rewardAmtDecimal, err := types.NewNumericType(uint256Precision-decimals, decimals)
			if err != nil {
				return p, fmt.Errorf("failed to create decimal type: %w", err)
			}

			return precompiles.Precompile{
				Methods: []precompiles.Method{ // NOTE: engine ensures 'resultFn' in Handler is always non-nil
					{
						// Supposed to be called by App
						Name:            "issue_reward",
						AccessModifiers: []precompiles.Modifier{precompiles.PUBLIC}, // TODO: change to SYSTEM
						Handler: func(ctx *common.EngineContext, app *common.App, inputs []any, resultFn func([]any) error) error {
							return ext.issueReward(ctx, app, inputs, resultFn)
						},
						Parameters: []precompiles.PrecompileValue{
							{Name: "wallet_address", Type: types.TextType, Nullable: false},
							{Name: "amount", Type: rewardAmtDecimal, Nullable: false},
						},
					},
					{
						// Supposed to be called by Signer service
						// Returns epoch rewards after(non-include) after_height, in ASC order.
						Name:            "list_epochs",
						AccessModifiers: []precompiles.Modifier{precompiles.PUBLIC, precompiles.VIEW},
						Handler: func(ctx *common.EngineContext, app *common.App, inputs []any, resultFn func([]any) error) error {
							return ext.listEpochs(ctx, app, inputs, resultFn)
						},
						Parameters: []precompiles.PrecompileValue{
							{Name: "after_height", Type: types.IntType, Nullable: false}, // after height
							{Name: "limit", Type: types.IntType, Nullable: false},        // limit
						},
						Returns: &precompiles.MethodReturn{
							IsTable: true,
							Fields:  (&Epoch{}).UnpackTypes(rewardAmtDecimal),
						},
					},
					{
						// Supposed to be called by the SignerService, to verify the reward root.
						// Could be merged into 'list_epochs'
						// Returns pending rewards from(include) start_height to(include) end_height, in ASC order.
						// NOTE: Rewards of same address will be aggregated.
						Name:            "search_rewards", // maybe not useful for Signer Service.
						AccessModifiers: []precompiles.Modifier{precompiles.PUBLIC, precompiles.VIEW},
						Handler: func(ctx *common.EngineContext, app *common.App, inputs []any, resultFn func([]any) error) error {
							return ext.searchRewards(ctx, app, inputs, resultFn)
						},
						Parameters: []precompiles.PrecompileValue{
							{Name: "start_height", Type: types.IntType, Nullable: false}, // start height
							{Name: "end_height", Type: types.IntType, Nullable: false},   // end height
						},
						Returns: &precompiles.MethodReturn{
							IsTable: true,
							Fields:  (&Reward{}).UnpackTypes(rewardAmtDecimal),
						},
					},
					{
						// Supposed to be called by Kwil network.
						Name:            "propose_epoch",
						AccessModifiers: []precompiles.Modifier{precompiles.PUBLIC}, // TODO: make this SYSTEM or Private
						Handler: func(ctx *common.EngineContext, app *common.App, inputs []any, resultFn func([]any) error) error {
							return ext.proposeEpoch(ctx, app, inputs, resultFn)
						},
					},
					{
						// Supposed to be called by SignerService.
						Name:            "vote_epoch",
						AccessModifiers: []precompiles.Modifier{precompiles.PUBLIC},
						Handler: func(ctx *common.EngineContext, app *common.App, inputs []any, resultFn func([]any) error) error {
							return ext.voteEpochReward(ctx, app, inputs, resultFn)
						},
						Parameters: []precompiles.PrecompileValue{
							// TODO: change to uuid??
							{Name: "sign_hash", Type: types.ByteaType, Nullable: false}, // sign hash
							{Name: "signature", Type: types.ByteaType, Nullable: false}, // signature
						},
					},
					{
						// Supposed to be called by PosterService.
						// Returns finalized rewards after(non-include) start_height, in ASC order.
						Name:            "list_finalized",
						AccessModifiers: []precompiles.Modifier{precompiles.PUBLIC, precompiles.VIEW},
						Handler: func(ctx *common.EngineContext, app *common.App, inputs []any, resultFn func([]any) error) error {
							return ext.listFinalized(ctx, app, inputs, resultFn)
						},
						Parameters: []precompiles.PrecompileValue{
							{Name: "after_height", Type: types.IntType, Nullable: false}, // after height
							{Name: "limit", Type: types.IntType, Nullable: false},        // limit
						},
						Returns: &precompiles.MethodReturn{
							IsTable: true,
							Fields:  (&FinalizedReward{}).UnpackTypes(rewardAmtDecimal),
						},
					},
					{
						// Supposed to be called by PosterService ?? seems this is not PosterService wants
						// Returns finalized rewards from(include) latest, in DESC order.
						Name:            "latest_finalized",
						AccessModifiers: []precompiles.Modifier{precompiles.PUBLIC, precompiles.VIEW},
						Handler: func(ctx *common.EngineContext, app *common.App, inputs []any, resultFn func([]any) error) error {
							return ext.latestFinalized(ctx, app, inputs, resultFn)
						},
						Parameters: []precompiles.PrecompileValue{
							{Name: "limit", Type: types.IntType, Nullable: true}, // limit, default to 1
						},
						Returns: &precompiles.MethodReturn{
							IsTable: true,
							Fields:  (&FinalizedReward{}).UnpackTypes(rewardAmtDecimal),
						},
					},
					//{
					//	// TODO
					//	// Supposed to be called by PosterService,
					//	// Returns finalized rewards whose safeNonce are newer than afterSafeNonce, in DESC order.
					//	Name:            "newer_finalized",
					//	AccessModifiers: []pc.Modifier{pc.PUBLIC, pc.VIEW},
					//	Handler: func(ctx *kcommon.EngineContext, app *kcommon.App, inputs []any, resultFn func([]any) error) error {
					//		return ext.latestFinalized(ctx, app, inputs, resultFn)
					//	},
					//	Parameters: []pc.PrecompileValue{
					//		{Type: types.IntType, Nullable: true}, // afterSafeNonce
					//		{Type: types.IntType, Nullable: true}, // limit, default to 1
					//	},
					//	Returns: &pc.MethodReturn{
					//		IsTable:    true,
					//		Fields:     (&FinalizedReward{}).UnpackTypes(rewardAmtDecimal),
					//		FieldNames: (&FinalizedReward{}).UnpackColumns(),
					//	},
					//},
					{
						// Supposed to be called by App/User
						Name:            "list_wallet_rewards",
						AccessModifiers: []precompiles.Modifier{precompiles.PUBLIC, precompiles.VIEW},
						Parameters: []precompiles.PrecompileValue{
							{Name: "address", Type: types.TextType, Nullable: false}, // wallet address
						},
						Returns: &precompiles.MethodReturn{
							IsTable: true,
							Fields:  (&WalletReward{}).UnpackTypes(),
						},
						Handler: func(ctx *common.EngineContext, app *common.App, inputs []any, resultFn func([]any) error) error {
							return ext.listWalletRewards(ctx, app, inputs, resultFn)
						},
					},
					{
						// Supposed to be called by App/User
						Name:            "claim_param",
						AccessModifiers: []precompiles.Modifier{precompiles.PUBLIC, precompiles.VIEW},
						Handler: func(ctx *common.EngineContext, app *common.App, inputs []any, resultFn func([]any) error) error {
							return ext.getClaimParam(ctx, app, inputs, resultFn)
						},
						Parameters: []precompiles.PrecompileValue{
							// TODO: change to uuid??
							{Name: "sign_hash", Type: types.ByteaType, Nullable: false},     // sign hash
							{Name: "wallet_address", Type: types.TextType, Nullable: false}, // wallet address
						},
						Returns: &precompiles.MethodReturn{
							IsTable: true,
							Fields: []precompiles.PrecompileValue{
								{Name: "recipient", Type: types.TextType, Nullable: false},
								{Name: "amount", Type: types.TextType, Nullable: false},
								{Name: "block_hash", Type: types.TextType, Nullable: false},
								{Name: "root", Type: types.TextType, Nullable: false},
								{Name: "proofs", Type: types.TextArrayType, Nullable: true},
							},
						},
					},
					// TODO: modify posterFee, modify signers
				},
				OnStart: func(ctx context.Context, app *common.App) error {
					tx, err := app.DB.BeginTx(ctx)
					if err != nil {
						return err
					}
					defer tx.Rollback(ctx)

					// OnStart is not called at a certain block height, nor in a transaction
					emptyEngineCtx := &common.EngineContext{
						TxContext: &common.TxContext{
							Ctx: ctx,
						}}

					// check if the reward contract exists
					contract, err := meta.GetRewardContract(emptyEngineCtx, app.Engine, app.DB, ext.ChainID, ext.ContractAddr)
					if err != nil {
						if !errors.Is(err, sql.ErrNoRows) {
							return err
						}
						return nil
					}

					// exist, use values from DB
					ext.contractID = meta.GenRewardContractID(ext.ChainID, ext.ContractAddr)
					ext.Signers = contract.Signers
					ext.Threshold = contract.Threshold
					ext.ContractNonce = contract.Nonce
					ext.SafeNonce = contract.SafeNonce
					return nil
				},
				OnUse: func(ctx *common.EngineContext, app *common.App) error {
					// if the engine is not using OverrideAuthz, usage should fail.
					// This is because we want to prevent the DB owner from being able to
					// call USE on this extension. Instead, this extension should only
					// be created by the meta extension
					if !ctx.OverrideAuthz && !ctx.InvalidTxCtx {
						return errors.New("erc20_rewards extension can only be used by the meta extension")
					}

					app.Service.Logger.Info("Register a new erc20_rewards contract",
						"chainID", ext.ChainID, "contractAddr", ext.ContractAddr, "alias", ext.alias)
					initRewardMeta := "USE IF NOT EXISTS erc20_rewards_meta as " + meta.ExtAlias + ";"
					err := app.Engine.Execute(ctx, app.DB, initRewardMeta, nil, nil)
					if err != nil {
						return err
					}

					_, err = app.Engine.Call(ctx, app.DB, meta.ExtAlias, "register",
						[]any{
							ext.ChainID,
							ext.ContractAddr,
							ext.ContractNonce,
							strings.Join(ext.Signers, ","),
							ext.Threshold,
							ext.SafeAddr,
							ext.SafeNonce,
						},
						nil)
					if err != nil {
						return err
					}

					return initTables(ctx, app, ext.alias)
				},
				OnUnuse: func(ctx *common.EngineContext, app *common.App) error {
					// if the engine is not using OverrideAuthz, usage should fail.
					// This is because we want to prevent the DB owner from being able to
					// call UNUSE on this extension. Instead, this extension should only
					// be unused by the meta extension
					if !ctx.OverrideAuthz && !ctx.InvalidTxCtx {
						return errors.New("erc20_rewards extension can only be unused by the meta extension")
					}
					return nil
				},
			}, nil
		})
	if err != nil {
		panic(fmt.Errorf("failed to register erc20_rewards initializer: %w", err))
	}
}

func initTables(ctx *common.EngineContext, app *common.App, ns string) error {
	ctx.OverrideAuthz = true
	defer func() { ctx.OverrideAuthz = false }()

	err := app.Engine.Execute(ctx, app.DB, fmt.Sprintf(sqlInitTableErc20rwRewards, ns, meta.ExtAlias), nil, nil)
	if err != nil {
		return err
	}

	err = app.Engine.Execute(ctx, app.DB, fmt.Sprintf(sqlInitTableErc20rwEpochs, ns, meta.ExtAlias), nil, nil)
	if err != nil {
		return err
	}

	err = app.Engine.Execute(ctx, app.DB, fmt.Sprintf(sqlInitTableErc20rwEpochVotes, ns, ns, meta.ExtAlias), nil, nil)
	if err != nil {
		return err
	}

	err = app.Engine.Execute(ctx, app.DB, fmt.Sprintf(sqlInitTableErc20rwFinalizedRewards, ns, ns), nil, nil)
	if err != nil {
		return err
	}

	err = app.Engine.Execute(ctx, app.DB, fmt.Sprintf(sqlInitTableRecipientReward, ns, ns), nil, nil)
	if err != nil {
		return err
	}

	return nil
}

// Erc20RewardExt a struct that implements the precompiles.Instance interface
type Erc20RewardExt struct {
	// contractID identifies a reward extension Erc20RewardExt, it's the contractID in table
	// erc20rw_meta_contracts
	contractID *types.UUID
	// alias is the namespace/schema current extension resides
	alias string
	// BuiltinMode indicates whether this ext is initialized in builtin mode.
	//BuiltinMode bool
	// ContractAddr is the reward escrow EVM contract address.
	ContractAddr string
	// SafeAddr is the GnosisSafe wallet address that has permission to update
	// the reward escrow contract.
	SafeAddr string
	ChainID  int64
	decimals uint16 // the denotation of the reward token, most of the ERC20 are 18

	// Those are the parameters that could be updated over time.
	// Upon reload(node restart), should populate those values from DB.
	// Upon updates, those should also get updated.
	// NOTE: We assume they can only be modified through ext.
	Signers       []string
	Threshold     int64
	ContractNonce int64
	SafeNonce     int64
}

// erc20RewardMetadata is the metadata for erc20 reward extension.
// It is used in the USE statement.
type erc20RewardMetadata struct {
	ChainID       int64
	EscrowAddress string
	Threshold     int64
	Signers       [][]byte
	SafeAddress   []byte
	SafeNonce     int64
	Erc20Address  []byte
	Decimals      uint16
}

// use executes a use statement for the metadata.
func (m *erc20RewardMetadata) use(ctx context.Context, app *common.App, alias string) error {
	return app.Engine.ExecuteWithoutEngineCtx(ctx, app.DB, fmt.Sprintf(`
	USE erc20_rewards {
		chain_id: $chain_id,
		contract_address: $contract_address,
		threshold: $threshold,
		signers: $signers,
		safe_address: $safe_address,
		safe_nonce: $safe_nonce,
		erc20_address: $erc20_address,
		decimals: $decimals
	} AS %s
	`, alias), map[string]any{
		"chain_id":         m.ChainID,
		"contract_address": m.EscrowAddress,
		"threshold":        m.Threshold,
		"signers":          m.Signers,
		"safe_address":     m.SafeAddress,
		"safe_nonce":       m.SafeNonce,
		"erc20_address":    m.Erc20Address,
		"decimals":         m.Decimals,
	}, nil)
}

// getMetadata parses the metadata map and returns the chainID, contractAddr, contractNonce, signers, threshold, safeAddr, safeNonce, decimals.
func getMetadata(metadata map[string]any) (*erc20RewardMetadata, error) {
	chainID, ok := metadata["chain_id"].(int64)
	if !ok {
		return nil, fmt.Errorf("invalid chain_id")
	}

	contractAddr, ok := metadata["contract_address"].([]byte)
	if !ok {
		return nil, fmt.Errorf("invalid contract_address")
	}

	signersBts, ok := metadata["signers"].([]*[]byte)
	if !ok {
		return nil, fmt.Errorf("invalid signers")
	}

	if len(signersBts) == 0 {
		return nil, fmt.Errorf("signers is empty")
	}

	signers := make([][]byte, len(signersBts))
	for i, signer := range signersBts {
		if signer == nil {
			return nil, fmt.Errorf("received null signer")
		}

		if !ethcommon.IsHexAddress(hex.EncodeToString(*signer)) {
			return nil, fmt.Errorf("signer is not a valid address: %s", hex.EncodeToString(*signer))
		}

		signers[i] = *signer
	}

	threshold, ok := metadata["threshold"].(int64)
	if !ok {
		return nil, fmt.Errorf("invalid threshold")
	}

	if threshold == 0 {
		return nil, fmt.Errorf("threshold is 0")
	}

	if threshold > int64(len(signers)) {
		return nil, fmt.Errorf("threshold is larger than the number of signers")
	}

	safeAddr, ok := metadata["safe_address"].([]byte)
	if !ok {
		return nil, fmt.Errorf("invalid safe_address")
	}

	if !ethcommon.IsHexAddress(hex.EncodeToString(safeAddr)) {
		return nil, fmt.Errorf("invalid safe_address")
	}

	safeNonce, ok := metadata["safe_nonce"].(int64)
	if !ok {
		return nil, fmt.Errorf("invalid safe_nonce")
	}

	erc20Addr, ok := metadata["erc20_address"].([]byte)
	if !ok {
		return nil, fmt.Errorf("invalid erc20_address")
	}

	if !ethcommon.IsHexAddress(hex.EncodeToString(erc20Addr)) {
		return nil, fmt.Errorf("invalid erc20_address")
	}

	decimals, ok := metadata["decimals"].(int64)
	if !ok {
		return nil, fmt.Errorf("invalid decimals")
	}

	if decimals <= 0 {
		return nil, fmt.Errorf("decimals should be positive")
	}

	if decimals > math.MaxUint16 {
		return nil, fmt.Errorf("decimals too large")
	}

	return &erc20RewardMetadata{
		ChainID:       chainID,
		EscrowAddress: hex.EncodeToString(contractAddr),
		Threshold:     threshold,
		Signers:       signers,
		SafeAddress:   safeAddr,
		SafeNonce:     safeNonce,
		Erc20Address:  erc20Addr,
		Decimals:      uint16(decimals),
	}, nil
}

// scaleUpUint256 turns a decimal into uint256, i.e. (11.22, 4) -> 112200
func scaleUpUint256(amount *types.Decimal, decimals uint16) (*types.Decimal, error) {
	unit, err := types.ParseDecimal("1" + strings.Repeat("0", int(decimals)))
	if err != nil {
		return nil, fmt.Errorf("create decimal unit failed: %w", err)
	}

	n, err := types.DecimalMul(amount, unit)
	if err != nil {
		return nil, fmt.Errorf("expand amount decimal failed: %w", err)
	}

	err = n.SetPrecisionAndScale(uint256Precision, 0)
	if err != nil {
		return nil, fmt.Errorf("expand amount decimal failed: %w", err)
	}

	return n, nil
}

// scaleDownUint256 turns an uint256 to a decimal, i.e. (112200, 4) -> 11.22
func scaleDownUint256(amount *types.Decimal, decimals uint16) (*types.Decimal, error) {
	unit, err := types.ParseDecimal("1" + strings.Repeat("0", int(decimals)))
	if err != nil {
		return nil, fmt.Errorf("create decimal unit failed: %w", err)
	}

	n, err := types.DecimalDiv(amount, unit)
	if err != nil {
		return nil, fmt.Errorf("expand amount decimal failed: %w", err)
	}

	scale := n.Scale()
	if scale > decimals {
		scale = decimals
	}

	err = n.SetPrecisionAndScale(uint256Precision-decimals, scale)
	if err != nil {
		return nil, fmt.Errorf("expand amount decimal failed: %w", err)
	}

	return n, nil
}

func (h *Erc20RewardExt) issueReward(ctx *common.EngineContext, app *common.App, inputs []any, resultFn func([]any) error) error {
	wallet, ok := inputs[0].(string)
	if !ok {
		return fmt.Errorf("invalid wallet address")
	}

	if !ethcommon.IsHexAddress(wallet) {
		return fmt.Errorf("invalid wallet address")
	}

	amount, ok := inputs[1].(*types.Decimal)
	if !ok {
		return fmt.Errorf("invalid amount")
	}

	// require amount is positive
	zero, _ := types.ParseDecimal("0.0")
	r, err := types.DecimalCmp(zero, amount)
	if err != nil {
		return fmt.Errorf("invalid amount")
	}
	if r != -1 {
		return fmt.Errorf("invalid amount")
	}

	uint256Amount, err := scaleUpUint256(amount, h.decimals)
	if err != nil {
		return err
	}

	counter := 0
	c, exist := ctx.TxContext.Value(txIssueRewardCounterKey)
	if exist {
		counter = c.(int)
	}

	// not matter if db operations success, increase the counter
	defer func() {
		ctx.TxContext.SetValue(txIssueRewardCounterKey, counter+1)
	}()

	if err := IssueReward(ctx, app.Engine, app.DB, h.alias, &Reward{
		ID:         GenRewardID(wallet, uint256Amount.String(), ctx.TxContext.TxID, counter),
		Recipient:  wallet,
		Amount:     uint256Amount,
		CreatedAt:  ctx.TxContext.BlockContext.Height,
		ContractID: h.contractID,
	}); err != nil {
		return err
	}

	return nil
}

// searchRewards returns rewards between a starting height and ending height.
func (h *Erc20RewardExt) searchRewards(ctx *common.EngineContext, app *common.App, inputs []any, resultFn func([]any) error) error {
	startHeight, ok := inputs[0].(int64)
	if !ok {
		return fmt.Errorf("invalid start height")
	}

	if startHeight < 0 {
		return fmt.Errorf("invalid start height")
	}

	endHeight, ok := inputs[1].(int64)
	if !ok {
		return fmt.Errorf("invalid end height")
	}

	if endHeight < 0 {
		return fmt.Errorf("invalid end height")
	}

	if startHeight > endHeight {
		return fmt.Errorf("invalid start height")
	}

	if endHeight-startHeight > 10000 {
		return fmt.Errorf("search range too large")
	}

	rewards, err := SearchRewards(ctx, app.Engine, app.DB, h.alias, h.decimals, startHeight, endHeight)
	if err != nil {
		return err
	}

	for _, r := range rewards {
		err := resultFn(r.UnpackValues())
		if err != nil {
			return err
		}
	}

	return nil
}

// proposeEpoch proposes a new epoch of rewards that are pending from last batch
// to current height, it requires a correct Gnosis Safe wallet nonce.
// inputs[0] is the Gnosis Safe wallet nonce, which will be used to generate Gnosis Safe TX hash.
//
// This is supposed to be called by Network owner or Signer service??
// In either case, there should be just one proposer at a time.
// Seems reasonable to be called by KwilNetwork, as this action requires nothing
// but safeNonce(which could also be inferred inside the Extension, no need to be provided by caller)
// For simplicity, we just use safeNonce tracked by extension.
// NOTE: well, do we need to check permission? e.g., check the caller??
func (h *Erc20RewardExt) proposeEpoch(ctx *common.EngineContext, app *common.App, inputs []any, resultFn func([]any) error) error {
	endHeight := ctx.TxContext.BlockContext.Height
	blockHash := ctx.TxContext.BlockContext.Hash

	// get last finalized batch
	var lastEpochEndHeight int64
	finalizedReward, err := GetLatestFinalizedReward(ctx, app.Engine, app.DB, h.alias, h.decimals)
	if err != nil {
		return err
	}
	if finalizedReward != nil {
		lastEpochEndHeight = finalizedReward.EndHeight
	}

	epochStartHeight := lastEpochEndHeight + 1
	rewards, err := SearchRewards(ctx, app.Engine, app.DB, h.alias, h.decimals, epochStartHeight, endHeight)
	if err != nil {
		return err
	}

	if len(rewards) == 0 {
		return fmt.Errorf("no rewards")
	}

	recipients := make([]string, len(rewards))
	bigIntAmounts := make([]*big.Int, len(rewards))
	var totalAmount *types.Decimal // nil
	for i, r := range rewards {
		recipients[i] = r.Recipient

		if totalAmount == nil {
			totalAmount = r.Amount
		} else {
			totalAmount, err = types.DecimalAdd(totalAmount, r.Amount)
			if err != nil {
				return err
			}
		}
		bigIntAmounts[i] = r.Amount.BigInt()
	}

	// NOTE: since we don't have a limit on how many leafs(recipients) a tree can
	// have, this could be big
	jsonMtree, rootHash, err := reward.GenRewardMerkleTree(recipients, bigIntAmounts, h.ContractAddr, blockHash)
	if err != nil {
		return err
	}

	safeTxData, err := reward.GenPostRewardTxData(rootHash, totalAmount.BigInt())
	if err != nil {
		return err
	}

	// safeTxHash is the data that all signers will be signing(using personal_sign)
	_, safeTxHash, err := reward.GenGnosisSafeTx(h.ContractAddr, h.SafeAddr,
		0, safeTxData, h.ChainID, h.SafeNonce)
	if err != nil {
		return err
	}

	// NOTE: we save the digest of the msg, so it's fix length
	// well, safeTxHash should also be a fix length, we save the digest anyway.
	signHash := accounts.TextHash(safeTxHash)

	ctx.OverrideAuthz = true
	defer func() { ctx.OverrideAuthz = false }()

	uint256TotalAmount, err := scaleUpUint256(totalAmount, h.decimals)
	if err != nil {
		return err
	}

	return app.Engine.Execute(ctx, app.DB, fmt.Sprintf(sqlNewEpoch, h.alias), map[string]any{
		"$id":            GenBatchRewardID(endHeight, signHash),
		"$start_height":  epochStartHeight,
		"$end_height":    endHeight,
		"$total_rewards": uint256TotalAmount,
		"$mtree_json":    []byte(jsonMtree),
		"$reward_root":   rootHash,
		"$safe_nonce":    h.SafeNonce,
		"$sign_hash":     signHash,
		"$contract_id":   h.contractID,
		"$block_hash":    blockHash[:],
		"$created_at":    ctx.TxContext.BlockContext.Height, // TODO: seems we can remove this field, it's always the same as end_height
	}, nil)
}

// listEpochs returns reward epochs starting from a given height.
// inputs[0] is the starting height, inputs[1] is the return size and default to 10.
func (h *Erc20RewardExt) listEpochs(ctx *common.EngineContext, app *common.App, inputs []any, resultFn func([]any) error) error {
	afterHeight, ok := inputs[0].(int64)
	if !ok {
		return fmt.Errorf("invalid after height")
	}

	if afterHeight < 0 { // or set to default
		return fmt.Errorf("invalid after height")
	}

	limit, ok := inputs[1].(int64)
	if !ok {
		return fmt.Errorf("invalid limit type")
	}

	if limit < 0 { // or set to default
		return fmt.Errorf("invalid limit")
	}

	if limit == 0 {
		limit = 1 // default to 1
	}

	if limit > 10 {
		limit = 10 // max to 10
	}
	rewards, err := ListEpochs(ctx, app.Engine, app.DB, h.alias, h.decimals, afterHeight, limit)
	if err != nil {
		return err
	}

	for _, r := range rewards {
		err := resultFn(r.UnpackValues())
		if err != nil {
			return err
		}
	}

	return nil
}

// voteEpochReward votes one epoch of rewards by providing correspond signature.
// inputs[0] is the data digest, inputs[1] is the signature.
// This is supposed to be called by Signer service.
func (h *Erc20RewardExt) voteEpochReward(ctx *common.EngineContext, app *common.App, inputs []any, resultFn func([]any) error) error {
	// verify the caller is signer
	_voter, err := meta.GetSigner(ctx, app.Engine, app.DB, ctx.TxContext.Caller, h.contractID)
	if err != nil {
		return fmt.Errorf("check signer: %w", err)
	}
	if _voter == nil {
		return fmt.Errorf("voter not allowed")
	}

	digest, ok := inputs[0].([]byte)
	if !ok {
		return fmt.Errorf("invalid safe tx hash")
	}

	if len(digest) == 0 {
		return fmt.Errorf("invalid safe tx hash")
	}

	signature, ok := inputs[1].([]byte)
	if !ok {
		return fmt.Errorf("invalid signature")
	}

	caller := ethcommon.HexToAddress(ctx.TxContext.Caller)
	err = reward.EthGnosisVerifyDigest(signature, digest, caller.Bytes())
	if err != nil {
		return err
	}

	// if already voted, skip
	var voted bool
	err = app.Engine.Execute(ctx, app.DB, fmt.Sprintf(sqlGetVoteBySignHash, h.alias, h.alias, meta.ExtAlias),
		map[string]any{
			"$sign_hash":      digest,
			"$signer_address": ctx.TxContext.Caller,
			"$contract_id":    h.contractID,
		},
		func(row *common.Row) error {
			voted = true
			return nil
		})
	if err != nil {
		return err
	}
	if voted {
		return nil
	}

	ctx.OverrideAuthz = true
	defer func() { ctx.OverrideAuthz = false }()

	err = app.Engine.Execute(ctx, app.DB, fmt.Sprintf(sqlVoteEpochBySignHash, h.alias, h.alias, meta.ExtAlias),
		map[string]any{
			"$sign_hash":      digest,
			"$signer_address": ctx.TxContext.Caller,
			"$contract_id":    h.contractID,
			"$signature":      signature,
			"$created_at":     ctx.TxContext.BlockContext.Height,
		}, nil)
	if err != nil {
		return err
	}

	finalized, err := TryFinalizeEpochReward(ctx, app.Engine, app.DB, h.alias, h.contractID,
		digest, ctx.TxContext.BlockContext.Height, h.Threshold)
	if err != nil {
		return err
	}

	if finalized {
		h.SafeNonce += 1
	}

	return nil
}

// listFinalized returns finalized rewards.
// inputs[0] is the starting height, inputs[1] is the batch size and default to 10.
// This is supposed to be called by Poster service.
func (h *Erc20RewardExt) listFinalized(ctx *common.EngineContext, app *common.App, inputs []any, resultFn func([]any) error) error {
	afterHeight, ok := inputs[0].(int64)
	if !ok {
		return fmt.Errorf("invalid after height")
	}

	if afterHeight < 0 { // or set to default
		return fmt.Errorf("invalid after height")
	}

	limit, ok := inputs[1].(int64)
	if !ok {
		return fmt.Errorf("invalid limit type")
	}

	if limit < 0 { // or set to default
		return fmt.Errorf("invalid limit")
	}
	if limit == 0 {
		limit = 10 // default to 10
	}

	if limit > 50 {
		limit = 50 // max to 50
	}

	rewards, err := ListFinalizedRewards(ctx, app.Engine, app.DB, h.alias, h.decimals, afterHeight, limit)
	if err != nil {
		return err
	}

	for _, r := range rewards {
		err := resultFn(r.UnpackValues())
		if err != nil {
			return err
		}
	}

	return nil
}

// latestFinalized returns latest finalized rewards.
// inputs[0] is the size and default to 0, i.e. return the newest.
// This is supposed to be called by Poster service.
func (h *Erc20RewardExt) latestFinalized(ctx *common.EngineContext, app *common.App, inputs []any, resultFn func([]any) error) error {
	var limit int64 = 1
	if inputs[0] != nil {
		var ok bool
		limit, ok = inputs[0].(int64)
		if !ok {
			return fmt.Errorf("invalid limit type")
		}
	}

	if limit <= 0 { // or set to default
		limit = 1
	}

	if limit > 20 {
		limit = 20 // max to 20
	}

	rewards, err := ListLatestFinalizedRewards(ctx, app.Engine, app.DB, h.alias, h.decimals, limit)
	if err != nil {
		return err
	}

	for _, r := range rewards {
		err := resultFn(r.UnpackValues())
		if err != nil {
			return err
		}
	}

	return nil
}

func (h *Erc20RewardExt) listWalletRewards(ctx *common.EngineContext, app *common.App, inputs []any, resultFn func([]any) error) error {
	wallet, ok := inputs[0].(string)
	if !ok {
		return fmt.Errorf("invalid wallet address")
	}

	if !ethcommon.IsHexAddress(wallet) {
		return fmt.Errorf("invalid wallet address")
	}

	walletAddr := ethcommon.HexToAddress(wallet)

	partialWrs, err := GetWalletRewards(ctx, app.Engine, app.DB, h.alias, walletAddr.String())
	if err != nil {
		return err
	}

	wrs := make([]*WalletReward, len(partialWrs))

	for i, pwr := range partialWrs {
		treeRoot, proofs, _, bh, uint256AmtStr, err := reward.GetMTreeProof(pwr.mTreeJSON, walletAddr.String())
		if err != nil {
			return err
		}

		info, ok := chainConvMap[pwr.chainID]
		if !ok {
			return fmt.Errorf("internal bug: unknown chain id")
		}

		wrs[i] = &WalletReward{
			Chain:          info.Name,
			ChainID:        pwr.chainID,
			Contract:       pwr.contract,
			EtherScan:      info.GetEtherscanAddr(pwr.contract),
			CreatedAt:      pwr.createdAt,
			ParamRecipient: walletAddr.String(),
			ParamAmount:    uint256AmtStr,
			ParamBlockHash: toBytes32Str(bh),
			ParamRoot:      toBytes32Str(treeRoot),
			ParamProofs:    meta.Map(proofs, toBytes32Str),
		}
	}

	for _, r := range wrs {
		err := resultFn(r.UnpackValues())
		if err != nil {
			return err
		}
	}

	return nil
}

func toBytes32Str(bs []byte) string {
	return "0x" + hex.EncodeToString(bs)
}

// getClaimParam returns the claim parameters for given signHash and wallet address,
// User can use the parameters directly to call the `claimReward` method on reward contract.
// inputs[0] is the safeTxHash, inputs[1] is the user wallet address.
// This is supposed to be called by User.
func (h *Erc20RewardExt) getClaimParam(ctx *common.EngineContext, app *common.App, inputs []any, resultFn func([]any) error) error {
	signHash, ok := inputs[0].([]byte)
	if !ok {
		return fmt.Errorf("invalid signhash")
	}

	if len(signHash) == 0 {
		return fmt.Errorf("invalid signhash")
	}

	wallet, ok := inputs[1].(string)
	if !ok {
		return fmt.Errorf("invalid wallet address")
	}

	if !ethcommon.IsHexAddress(wallet) {
		return fmt.Errorf("invalid wallet address")
	}

	walletAddr := ethcommon.HexToAddress(wallet)

	mTreeJson, err := GetEpochMTreeBySignhash(ctx, app.Engine, app.DB, h.alias, signHash)
	if err != nil {
		return err
	}

	if mTreeJson == nil {
		return fmt.Errorf("no reward found")
	}

	treeRoot, proofs, _, bh, uint256Amt, err := reward.GetMTreeProof(string(mTreeJson), walletAddr.String())
	if err != nil {
		return err
	}

	return resultFn([]any{
		walletAddr.String(),
		uint256Amt,
		toBytes32Str(bh),
		toBytes32Str(treeRoot),
		meta.Map(proofs, toBytes32Str)})
}
