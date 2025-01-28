// This file contains reward extension related types and database operations.
package erc20reward

import (
	"context"
	"fmt"

	kcommon "github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/core/types"
	pc "github.com/kwilteam/kwil-db/extensions/precompiles"
	"github.com/kwilteam/kwil-db/node/exts/erc20reward/meta"
	"github.com/kwilteam/kwil-db/node/types/sql"
)

var (
	sqlInitTableErc20rwPendingRewards = `
-- erc20rw_pending_rewards holds pending rewards that have not yet been finalized.
-- The id is generated based on the recipient_amount_contractID.
{%s}CREATE TABLE IF NOT EXISTS erc20rw_pending_rewards (
	id UUID PRIMARY KEY,
	recipient TEXT NOT NULL,
	amount NUMERIC(78,0) NOT NULL, -- allows uint256
	contract_id UUID NOT NULL references %s.erc20rw_meta_contracts(id) ON UPDATE CASCADE ON DELETE CASCADE,
	created_at INT8 NOT NULL -- kwil block height
);`

	// TODO: maybe add safeTxHash? it'll give us a way to find the related EVM tx.
	sqlInitTableErc20rwEpochRewards = `
-- erc20rw_epoch_rewards holds the finalized rewards for a batch of rewards.
-- If no one votes for a batch, it's fine, since each batch using a unique safe_nonce, e.g., even
-- the batch got votes later for some reason, it won't succeed. BUT seems a trouble for Poster service
{%s}CREATE TABLE IF NOT EXISTS erc20rw_epoch_rewards (
	id UUID PRIMARY KEY,
    start_height int8 NOT NULL, -- the height of the first reward in this batch. we need this in case a proposed batch is wrong and no one votes for it.
	end_height INT8 NOT NULL, -- the height of the last reward in this batch.
	total_rewards NUMERIC(78,0) NOT NULL, -- the total rewards issued in this batch, calculated automatically. Allow uint256.
    mtree_json BYTEA NOT NULL, -- the merkle tree of rewards, serialized as JSON. so later we can generate proof for user.
	reward_root BYTEA UNIQUE NOT NULL, -- the root of the merkle tree of rewards, it's unique per contract
    safe_nonce INT8 NOT NULL, -- the nonce of the Gnosis Safe wallet; used to generate the sign_hash
	sign_hash BYTEA UNIQUE NOT NULL, -- the hash(safeTxHash) that should be signed, and it's chain aware(because safeTX is). it's unique otherwise it can be replayed. Multisig nodes should independently calculate rewards themselves. This is simply stored here for quicker verification of signatures.
	contract_id UUID NOT NULL references %s.erc20rw_meta_contracts(id) ON UPDATE CASCADE ON DELETE CASCADE,
    block_hash BYTEA NOT NULL, -- the hash of the block that is used in merkle tree leaf
	created_at INT8 NOT NULL -- kwil block height
);`

	sqlInitTableErc20rwPendingSignatures = `
-- erc20rw_pending_signatures holds signatures for a reward epoch that has
-- not yet received enough signatures.
{%s}CREATE TABLE IF NOT EXISTS erc20rw_pending_signatures (
	epoch_id UUID NOT NULL REFERENCES %s.erc20rw_epoch_rewards(id) ON UPDATE CASCADE ON DELETE CASCADE,
	signer_id UUID NOT NULL REFERENCES %s.erc20rw_meta_signers(id) ON UPDATE CASCADE ON DELETE CASCADE,
	signature BYTEA NOT NULL,
	created_at INT8 NOT NULL, -- kwil block height
	PRIMARY KEY (epoch_id, signer_id)
);`

	sqlInitTableErc20rwFinalizedRewards = `
-- erc20rw_finalized_rewards holds finalized rewards that have been finalized.
-- A finalized reward is considered finalized on chain.
{%s}CREATE TABLE IF NOT EXISTS erc20rw_finalized_rewards (
    id UUID PRIMARY KEY,
    voters TEXT[] NOT NULL, -- snapshot of the voters of the epoch
	signatures BYTEA[] NOT NULL, -- snapshot of the signatures of the epoch
    epoch_id UUID NOT NULL REFERENCES %s.erc20rw_epoch_rewards(id) ON UPDATE CASCADE ON DELETE CASCADE,
    created_at INT8 NOT NULL
);`

	sqlNewPendingReward = `{%s}INSERT INTO erc20rw_pending_rewards (id, recipient, amount, contract_id, created_at) VALUES ($id, $recipient, $amount, $contract_id, $created_at);`

	sqlListRangePendingRewards = `SELECT * FROM %s.erc20rw_pending_rewards WHERE created_at >= $start_height and created_at <= $end_height ORDER BY created_at ASC`

	sqlNewEpochReward = `{%s}INSERT INTO erc20rw_epoch_rewards
(id, start_height, end_height, total_rewards, mtree_json, reward_root, safe_nonce, sign_hash, contract_id, block_hash, created_at)
VALUES ($id, $start_height, $end_height, $total_rewards, $mtree_json, $reward_root, $safe_nonce, $sign_hash, $contract_id, $block_hash, $created_at)`
	sqlGetEpochMtreeBySignhash    = `SELECT mtree_json FROM %s.erc20rw_epoch_rewards WHERE sign_hash = $sign_hash`
	sqlGetEpochBySignhash         = `SELECT * FROM %s.erc20rw_epoch_rewards WHERE sign_hash = $sign_hash`
	sqlListEpochRewardsWithVoters = `select er.*, array_agg(s.address) as voters
from %s.erc20rw_epoch_rewards as er
left join %s.erc20rw_pending_signatures as ps on er.id = ps.epoch_id
left join %s.erc20rw_meta_signers as s on ps.signer_id = s.id
WHERE er.end_height > $after_height
group by er.id, er.start_height, er.end_height, er.total_rewards, er.mtree_json, er.reward_root, er.safe_nonce, er.sign_hash, er.contract_id, er.block_hash, er.created_at
ORDER BY er.end_height ASC limit $limit`

	sqlVoteEpochBySignHash = `{%s}INSERT INTO erc20rw_pending_signatures (epoch_id, signer_id, signature, created_at)
VALUES ((SELECT id FROM %s.erc20rw_epoch_rewards WHERE sign_hash = $sign_hash),
        (SELECT id FROM %s.erc20rw_meta_signers WHERE address = $signer_address and contract_id = $contract_id),
        $signature, $created_at)`
	sqlCountSignatures   = `SELECT COUNT(*) FROM %s.erc20rw_pending_signatures WHERE epoch_id = (select id from %s.erc20rw_epoch_rewards WHERE sign_hash = $sign_hash)`
	sqlGetVoteBySignHash = `SELECT * from %s.erc20rw_pending_signatures where
    epoch_id = (SELECT id FROM %s.erc20rw_epoch_rewards WHERE sign_hash = $sign_hash)
    and signer_id = (SELECT id FROM %s.erc20rw_meta_signers WHERE address = $signer_address and contract_id = $contract_id)`

	sqlCreateFinalizedReward = `{%s}WITH
epoch AS (SELECT * FROM %s.erc20rw_epoch_rewards WHERE sign_hash = $sign_hash),
votes as (SELECT * FROM %s.erc20rw_pending_signatures WHERE epoch_id = (select id from epoch)),
sigs AS (SELECT ARRAY_AGG(signature) as signatures FROM votes),
voters As (SELECT ARRAY_AGG(s.address) as voters FROM votes as vs join %s.erc20rw_meta_signers as s on vs.signer_id = s.id)
INSERT INTO erc20rw_finalized_rewards (id, voters, signatures, epoch_id, created_at)
VALUES ($rid,(SELECT voters from voters),(SELECT signatures from sigs),(select id from epoch), $created_at)`
	sqlListFinalizedRewards = `SELECT fr.*, er.start_height, er.end_height, er.total_rewards, er.reward_root, er.safe_nonce, er.sign_hash, er.contract_id, er.block_hash
FROM %s.erc20rw_finalized_rewards as fr
join %s.erc20rw_epoch_rewards as er on er.id = fr.epoch_id
WHERE end_height > $after_height ORDER BY end_height ASC limit $limit`
	sqlListLatestFinalizedRewards = `SELECT fr.*, er.start_height, er.end_height, er.total_rewards, er.reward_root, er.safe_nonce, er.sign_hash, er.contract_id, er.block_hash
FROM %s.erc20rw_finalized_rewards as fr
join %s.erc20rw_epoch_rewards as er on er.id = fr.epoch_id
ORDER by end_height DESC LIMIT $limit`
	sqlGetFinalizedRewardByHash = `SELECT fr.*, er.start_height, er.end_height, er.total_rewards, er.reward_root, er.safe_nonce, er.sign_hash, er.contract_id, er.block_hash
FROM %s.erc20rw_finalized_rewards as fr
join %s.erc20rw_epoch_rewards as er on er.id = fr.epoch_id
WHERE er.sign_hash = $sign_hash`
)

type EngineExecutor interface {
	Execute(ctx *kcommon.EngineContext, db sql.DB, statement string, params map[string]any, fn func(*kcommon.Row) error) error
	ExecuteWithoutEngineCtx(ctx context.Context, db sql.DB, statement string, params map[string]any, fn func(*kcommon.Row) error) error
}

// PendingReward is the data model of table erc20rw_pending_rewards.
type PendingReward struct {
	ID         *types.UUID
	Recipient  string
	Amount     *types.Decimal
	ContractID *types.UUID
	CreatedAt  int64
}

func (pr *PendingReward) UnpackColumns() []string {
	return []string{
		"id",
		"recipient",
		"amount",
		"contract_id",
		"created_at",
	}
}

func (pr *PendingReward) UnpackValues() []any {
	return []any{
		pr.ID,
		pr.Recipient,
		pr.Amount,
		pr.ContractID,
		pr.CreatedAt,
	}
}

func (pr *PendingReward) UnpackTypes(decimalType *types.DataType) []pc.PrecompileValue {
	return []pc.PrecompileValue{
		{Type: types.UUIDType, Nullable: false},
		{Type: types.TextType, Nullable: false},
		{Type: decimalType, Nullable: false},
		{Type: types.UUIDType, Nullable: false},
		{Type: types.IntType, Nullable: false},
	}
}

// EpochReward is the data model of table erc20rw_epoch_rewards.
type EpochReward struct {
	ID           *types.UUID
	StartHeight  int64
	EndHeight    int64
	TotalRewards *types.Decimal
	MtreeJson    string
	RewardRoot   []byte
	SafeNonce    int64
	SignHash     []byte
	ContractID   *types.UUID
	BlockHash    []byte
	CreatedAt    int64
	Voters       []string //
}

func (br *EpochReward) UnpackColumns() []string {
	return []string{
		"id",
		"start_height",
		"end_height",
		"total_rewards",
		//"mtree_json", // we don't want user to access this
		"reward_root",
		"safe_nonce",
		"sign_hash",
		"contract_id",
		"block_hash",
		"created_at",
		"voters",
	}
}

func (br *EpochReward) UnpackValues() []any {
	return []any{
		br.ID,
		br.StartHeight,
		br.EndHeight,
		br.TotalRewards,
		br.RewardRoot,
		br.SafeNonce,
		br.SignHash,
		br.ContractID,
		br.BlockHash,
		br.CreatedAt,
		br.Voters,
	}
}

func (br *EpochReward) UnpackTypes(decimalType *types.DataType) []pc.PrecompileValue {
	return []pc.PrecompileValue{
		{Type: types.UUIDType, Nullable: false},
		{Type: types.IntType, Nullable: false},
		{Type: types.IntType, Nullable: false},
		{Type: decimalType, Nullable: false},
		{Type: types.ByteaType, Nullable: false},
		{Type: types.IntType, Nullable: false},
		{Type: types.ByteaType, Nullable: false},
		{Type: types.UUIDType, Nullable: false},
		{Type: types.ByteaType, Nullable: false},
		{Type: types.IntType, Nullable: false},
		{Type: types.TextArrayType, Nullable: false},
	}
}

// FinalizedReward is the data model of table erc20rw_finalized_rewards.
type FinalizedReward struct {
	ID         *types.UUID
	Voters     []string
	Signatures [][]byte
	EpochID    *types.UUID
	CreatedAt  int64
	//
	StartHeight  int64
	EndHeight    int64
	TotalRewards *types.Decimal
	RewardRoot   []byte
	SafeNonce    int64
	SignHash     []byte
	ContractID   *types.UUID
	BlockHash    []byte
}

func (fr *FinalizedReward) UnpackColumns() []string {
	return []string{
		"id",
		"voters",
		"signatures",
		"epoch_id",
		"created_at",
		"start_height",
		"end_height",
		"total_rewards",
		"reward_root",
		"safe_nonce",
		"sign_hash",
		"contract_id",
		"block_hash",
	}
}

func (fr *FinalizedReward) UnpackValues() []any {
	return []any{
		fr.ID,
		fr.Voters,
		fr.Signatures,
		fr.EpochID,
		fr.CreatedAt,
		fr.StartHeight,
		fr.EndHeight,
		fr.TotalRewards,
		fr.RewardRoot,
		fr.SafeNonce,
		fr.SignHash,
		fr.ContractID,
		fr.BlockHash,
	}
}

func (fr *FinalizedReward) UnpackTypes(decimalType *types.DataType) []pc.PrecompileValue {
	return []pc.PrecompileValue{
		{Type: types.UUIDType, Nullable: false},
		{Type: types.TextArrayType, Nullable: false},
		{Type: types.ByteaArrayType, Nullable: false},
		{Type: types.UUIDType, Nullable: false},
		{Type: types.IntType, Nullable: false},
		{Type: types.IntType, Nullable: false},
		{Type: types.IntType, Nullable: false},
		{Type: decimalType, Nullable: false},
		{Type: types.ByteaType, Nullable: false},
		{Type: types.IntType, Nullable: false},
		{Type: types.ByteaType, Nullable: false},
		{Type: types.UUIDType, Nullable: false},
		{Type: types.ByteaType, Nullable: false},
	}
}

// GenPendingRewardID generates a unique UUID for a reward. We need special handling
// here because there could be multiple rewards to the same user with the same amount.
func GenPendingRewardID(recipient string, amount string, txID string, idx int) *types.UUID {
	return types.NewUUIDV5([]byte(fmt.Sprintf("erc20rw_pending_rewards_%v_%v_%v_%v", recipient, amount, txID, idx)))
}

func GenBatchRewardID(endHeight int64, signHash []byte) *types.UUID {
	return types.NewUUIDV5([]byte(fmt.Sprintf("erc20rw_epoch_rewards_%v_%x", endHeight, signHash)))
}

func GenFinalizedRewardID(contractID *types.UUID, digest []byte) *types.UUID {
	return types.NewUUIDV5([]byte(fmt.Sprintf("erc20rw_finalized_rewards_%v_%x", contractID.String(), digest)))
}

func IssueReward(ctx *kcommon.EngineContext, engine EngineExecutor, db sql.DB, ns string, pr *PendingReward) error {
	ctx.OverrideAuthz = true
	defer func() { ctx.OverrideAuthz = false }()

	// Need access to block height.
	query := fmt.Sprintf(sqlNewPendingReward, ns)
	return engine.Execute(ctx, db, query, map[string]any{
		"$id":          pr.ID,
		"$recipient":   pr.Recipient,
		"$amount":      pr.Amount,
		"$created_at":  pr.CreatedAt,
		"$contract_id": pr.ContractID,
	}, nil)
}

// ListPendingRewards returns all pending rewards from the last batch till given height.
func ListPendingRewards(ctx *kcommon.EngineContext, engine EngineExecutor, db sql.DB, ns string,
	rewardDecimals uint16, startHeight int64, endHeight int64) ([]*PendingReward, error) {
	query := fmt.Sprintf(sqlListRangePendingRewards, ns)

	var rewards []*PendingReward
	idx := 0
	// if same recipient get issued multiple times, we need to aggregate it
	seenRecipients := make(map[string]int) // recipient to idx
	err := engine.Execute(ctx, db, query, map[string]any{
		"$start_height": startHeight,
		"$end_height":   endHeight,
	}, func(row *kcommon.Row) error {
		pr, err := rowToPendingReward(row.Values, rewardDecimals)
		if err != nil {
			return err
		}

		if _, ok := seenRecipients[pr.Recipient]; ok {
			rewards[seenRecipients[pr.Recipient]].Amount, _ = types.DecimalAdd(rewards[seenRecipients[pr.Recipient]].Amount, pr.Amount)
			rewards[seenRecipients[pr.Recipient]].Amount.SetPrecisionAndScale(pr.Amount.Precision(), pr.Amount.Scale()) // keep precision/scale
			rewards[seenRecipients[pr.Recipient]].CreatedAt = pr.CreatedAt                                              // use the latest issuance
			return nil
		}

		rewards = append(rewards, pr)

		seenRecipients[pr.Recipient] = idx

		idx += 1
		return nil
	})
	if err != nil {
		return nil, err
	}

	return rewards, nil
}

func rowToPendingReward(row []any, decimals uint16) (*PendingReward, error) {
	if len(row) != 5 {
		return nil, fmt.Errorf("internal bug, expected 5 columns from pending rewards, got %d", len(row))
	}

	id, ok := row[0].(*types.UUID)
	if !ok {
		return nil, fmt.Errorf("failed to convert id to UUID")
	}
	recipient, ok := row[1].(string)
	if !ok {
		return nil, fmt.Errorf("failed to convert recipient to string")
	}

	uint256Amount, ok := row[2].(*types.Decimal)
	if !ok {
		return nil, fmt.Errorf("failed to convert amount to types.Decimal")
	}
	amount, err := scaleDownUint256(uint256Amount, decimals)
	if err != nil {
		return nil, fmt.Errorf("failed to scale down uint256 amount: %w", err)
	}

	contractID, ok := row[3].(*types.UUID)
	if !ok {
		return nil, fmt.Errorf("failed to convert contract_id to UUID")
	}

	createdAt, ok := row[4].(int64)
	if !ok {
		return nil, fmt.Errorf("failed to convert created_at to int64")
	}
	return &PendingReward{
		ID:         id,
		Recipient:  recipient,
		Amount:     amount,
		CreatedAt:  createdAt,
		ContractID: contractID,
	}, nil
}

func ListEpochs(ctx *kcommon.EngineContext, engine EngineExecutor, db sql.DB, ns string,
	rewardDecimals uint16, afterBlockHeight, limit int64) ([]*EpochReward, error) {
	query := fmt.Sprintf(sqlListEpochRewardsWithVoters, ns, ns, meta.ExtAlias)

	var epochRewards []*EpochReward
	err := engine.Execute(ctx, db, query, map[string]any{
		"$after_height": afterBlockHeight,
		"$limit":        limit,
	}, func(row *kcommon.Row) error {
		er, err := rowToEpochReward(row.Values, rewardDecimals)
		if err != nil {
			return err
		}
		epochRewards = append(epochRewards, er)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return epochRewards, nil
}

func rowToEpochReward(row []any, decimals uint16) (*EpochReward, error) {
	if len(row) != 12 {
		return nil, fmt.Errorf("internal bug, expected 12 columns from epoch rewards, got %d", len(row))
	}

	id, ok := row[0].(*types.UUID)
	if !ok {
		return nil, fmt.Errorf("failed to convert id to UUID")
	}
	startHeight, ok := row[1].(int64)
	if !ok {
		return nil, fmt.Errorf("failed to convert start_height to int64")
	}
	endHeight, ok := row[2].(int64)
	if !ok {
		return nil, fmt.Errorf("failed to convert end_height to int64")
	}
	uint256TotalRewards, ok := row[3].(*types.Decimal)
	if !ok {
		return nil, fmt.Errorf("failed to convert total_rewards to *types.Decimal")
	}
	totalRewards, err := scaleDownUint256(uint256TotalRewards, decimals)
	if err != nil {
		return nil, fmt.Errorf("failed to scale down uint256 amount: %w", err)
	}
	// NOTE: we don't want to return merkle tree, skip it
	rewardRoot, ok := row[5].([]byte)
	if !ok {
		return nil, fmt.Errorf("failed to convert reward_root to []byte")
	}
	safeNonce, ok := row[6].(int64)
	if !ok {
		return nil, fmt.Errorf("failed to convert safe_nonce to int64")
	}
	signHash, ok := row[7].([]byte)
	if !ok {
		return nil, fmt.Errorf("failed to convert sign_hash to []byte")
	}
	contractID, ok := row[8].(*types.UUID)
	if !ok {
		return nil, fmt.Errorf("failed to convert contract_id to UUID")
	}
	blockHash, ok := row[9].([]byte)
	if !ok {
		return nil, fmt.Errorf("failed to convert block to []byte")
	}
	createdAt, ok := row[10].(int64)
	if !ok {
		return nil, fmt.Errorf("failed to convert created_at to int64")
	}
	voters, ok := row[11].([]*string)
	if !ok {
		return nil, fmt.Errorf("failed to convert voters to []*string")
	}

	return &EpochReward{
		ID:           id,
		StartHeight:  startHeight,
		EndHeight:    endHeight,
		TotalRewards: totalRewards,
		//MtreeJson:    mtreeJson,
		RewardRoot: rewardRoot,
		SafeNonce:  safeNonce,
		SignHash:   signHash,
		ContractID: contractID,
		BlockHash:  blockHash,
		CreatedAt:  createdAt,
		Voters: meta.Map(voters, func(v *string) string {
			if v == nil {
				return ""
			} else {
				return *v
			}
		}),
	}, nil
}

func GetEpochRewardMTreeBySignhash(ctx *kcommon.EngineContext, engine EngineExecutor, db sql.DB, ns string, signHash []byte) ([]byte, error) {
	var mtreeJson []byte
	err := engine.Execute(ctx, db, fmt.Sprintf(sqlGetEpochMtreeBySignhash, ns),
		map[string]any{"$sign_hash": signHash},
		func(row *kcommon.Row) error {
			var ok bool
			mtreeJson, ok = row.Values[0].([]byte)
			if !ok {
				return fmt.Errorf("failed to convert mtree_json to []byte")
			}
			return nil
		})
	if err != nil {
		return nil, err
	}

	return mtreeJson, nil
}

func GetEpochRewardBySignhash(ctx *kcommon.EngineContext, engine EngineExecutor, db sql.DB, ns string, signHash []byte, rewardDecimals uint16) (*EpochReward, error) {
	var er *EpochReward
	err := engine.Execute(ctx, db, fmt.Sprintf(sqlGetEpochBySignhash, ns),
		map[string]any{"$sign_hash": signHash},
		func(row *kcommon.Row) error {
			var err error
			er, err = rowToEpochReward(row.Values, rewardDecimals)
			return err
		})
	if err != nil {
		return nil, err
	}

	return er, nil
}

func TryFinalizeEpochReward(ctx *kcommon.EngineContext, engine EngineExecutor, db sql.DB, ns string,
	contractID *types.UUID, digest []byte, height int64, threshold int64) (bool, error) {
	// TODO: might be able to implement the following using only SQL or Procedure, less round-trip

	voteCount := int64(0)
	err := engine.Execute(ctx, db, fmt.Sprintf(sqlCountSignatures, ns, ns),
		map[string]any{"$sign_hash": digest},
		func(row *kcommon.Row) error {
			var ok bool
			voteCount, ok = row.Values[0].(int64)
			if !ok {
				return fmt.Errorf("failed to convert count to int64")
			}
			return nil
		})
	if err != nil {
		return false, err
	}
	if voteCount < threshold {
		return false, nil
	}

	// if already finalized, skip
	var finalized bool
	err = engine.Execute(ctx, db, fmt.Sprintf(sqlGetFinalizedRewardByHash, ns, ns),
		map[string]any{"$sign_hash": digest},
		func(row *kcommon.Row) error {
			finalized = true
			return nil
		})
	if err != nil {
		return false, err
	}
	if finalized {
		return false, nil
	}

	// create finalized reward
	err = engine.Execute(ctx, db, fmt.Sprintf(sqlCreateFinalizedReward, ns, ns, ns, meta.ExtAlias),
		map[string]any{
			"$rid":        GenFinalizedRewardID(contractID, digest),
			"$sign_hash":  digest,
			"$created_at": height,
		}, nil)
	if err != nil {
		return false, err
	}

	// NOTE: should call through engine.Call???
	return true, meta.IncrementSafeNonce(ctx, engine, db, contractID)
}

func ListFinalizedRewards(ctx *kcommon.EngineContext, engine EngineExecutor, db sql.DB, ns string,
	rewardDecimals uint16, afterBlockHeight, limit int64) ([]*FinalizedReward, error) {
	query := fmt.Sprintf(sqlListFinalizedRewards, ns, ns)
	var rewards []*FinalizedReward
	err := engine.Execute(ctx, db, query, map[string]any{
		"$after_height": afterBlockHeight,
		"$limit":        limit,
	}, func(row *kcommon.Row) error {
		reward, err := rowToFinalizedReward(row.Values, rewardDecimals)
		if err != nil {
			return err
		}
		rewards = append(rewards, reward)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return rewards, nil
}

func ListLatestFinalizedRewards(ctx *kcommon.EngineContext, engine EngineExecutor, db sql.DB, ns string,
	rewardDecimals uint16, limit int64) ([]*FinalizedReward, error) {
	query := fmt.Sprintf(sqlListLatestFinalizedRewards, ns, ns)
	var rewards []*FinalizedReward
	err := engine.Execute(ctx, db, query, map[string]any{
		"$limit": limit,
	}, func(row *kcommon.Row) error {
		reward, err := rowToFinalizedReward(row.Values, rewardDecimals)
		if err != nil {
			return err
		}

		rewards = append(rewards, reward)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return rewards, nil
}

func GetLatestFinalizedReward(ctx *kcommon.EngineContext, engine EngineExecutor, db sql.DB, namespace string, rewardDecimals uint16) (*FinalizedReward, error) {
	frs, err := ListLatestFinalizedRewards(ctx, engine, db, namespace, rewardDecimals, 1)
	if err != nil {
		return nil, err
	}
	if len(frs) == 0 {
		return nil, nil
	}
	return frs[0], nil
}

func rowToFinalizedReward(row []any, decimals uint16) (*FinalizedReward, error) {
	if len(row) != 13 {
		return nil, fmt.Errorf("internal bug, expected 13 columns from epoch rewards, got %d", len(row))
	}

	id, ok := row[0].(*types.UUID)
	if !ok {
		return nil, fmt.Errorf("failed to convert id to UUID")
	}

	voters, ok := row[1].([]*string)
	if !ok {
		return nil, fmt.Errorf("failed to convert voters to []*string")
	}

	signatures, ok := row[2].([]*[]byte)
	if !ok {
		return nil, fmt.Errorf("failed to convert signatures to []*[]byte")
	}

	epochID, ok := row[3].(*types.UUID)
	if !ok {
		return nil, fmt.Errorf("failed to convert epoch_id to UUID")
	}

	createdAt, ok := row[4].(int64)
	if !ok {
		return nil, fmt.Errorf("failed to convert created_at to int64")
	}

	startHeight, ok := row[5].(int64)
	if !ok {
		return nil, fmt.Errorf("failed to convert start_height to int64")
	}

	endHeight, ok := row[6].(int64)
	if !ok {
		return nil, fmt.Errorf("failed to convert end_height to int64")
	}

	uint256TotalRewards, ok := row[7].(*types.Decimal)
	if !ok {
		return nil, fmt.Errorf("failed to convert total_rewards to *types.Decimal")
	}
	totalRewards, err := scaleDownUint256(uint256TotalRewards, decimals)
	if err != nil {
		return nil, fmt.Errorf("failed to scale down uint256 amount: %w", err)
	}

	rewardRoot, ok := row[8].([]byte)
	if !ok {
		return nil, fmt.Errorf("failed to convert reward_root to []byte")
	}

	safeNonce, ok := row[9].(int64)
	if !ok {
		return nil, fmt.Errorf("failed to convert safe_nonce to int64")
	}

	signHash, ok := row[10].([]byte)
	if !ok {
		return nil, fmt.Errorf("failed to convert sign_hash to []byte")
	}

	contactID, ok := row[11].(*types.UUID)
	if !ok {
		return nil, fmt.Errorf("failed to convert contract_id to UUID")
	}

	blockHash, ok := row[12].([]byte)
	if !ok {
		return nil, fmt.Errorf("failed to convert block hash to []byte")
	}

	return &FinalizedReward{
		ID: id,
		Voters: meta.Map(voters, func(v *string) string {
			if v == nil {
				return ""
			} else {
				return *v
			}
		}),
		Signatures: meta.Map(signatures, func(v *[]byte) []byte {
			if v == nil {
				return nil
			} else {
				return *v
			}
		}),
		EpochID:      epochID,
		CreatedAt:    createdAt,
		StartHeight:  startHeight,
		EndHeight:    endHeight,
		TotalRewards: totalRewards,
		RewardRoot:   rewardRoot,
		SafeNonce:    safeNonce,
		SignHash:     signHash,
		ContractID:   contactID,
		BlockHash:    blockHash,
	}, nil
}
