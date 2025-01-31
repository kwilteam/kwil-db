package meta

import (
	"context"
	"errors"
	"fmt"

	ethCommon "github.com/ethereum/go-ethereum/common"

	kcommon "github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/node/engine"
	"github.com/kwilteam/kwil-db/node/types/sql"
)

// ExtAlias is the namespace/schema of rewards meta-extension.
// All extensions use rewards meta extension should use this alias.
const ExtAlias = "erc20_rewards_meta_ns"

// sql
var (
	sqlInitTableErc20rwContracts = fmt.Sprintf(`
-- erc20rw_meta_contracts holds all contracts that the network knows of.
-- It is set when the extension is imported using USE
-- NOTE: We might need to store GnosisSafe address here as well.
{%s}CREATE TABLE IF NOT EXISTS erc20rw_meta_contracts (
	id UUID PRIMARY KEY,
    chain_id INT8 NOT NULL, -- the chain ID of the contract. -- TODO: use Text for bigint
	address TEXT NOT NULL, -- the reward escrow contract address.
	nonce INT8 NOT NULL, -- the last known nonce of the contract
	threshold INT8 NOT NULL,
    safe_address TEXT NOT NULL, -- the GnosisSafe address.
    safe_nonce INT8 NOT NULL, -- the last known nonce of the safe. NOTE: unless we force the safe can only be updated through KWIL, which is not practical, so the nonce may change without the ext knowing.
    unique (chain_id, address) -- unique per chain and address
);`, ExtAlias)

	sqlInitTableErc20rwSigners = fmt.Sprintf(`
-- erc20rw_meta_signers holds all signers that of a reward contract.
{%s}CREATE TABLE IF NOT EXISTS erc20rw_meta_signers (
	id UUID PRIMARY KEY,
	address TEXT NOT NULL, -- eth address
	contract_id UUID NOT NULL REFERENCES %s.erc20rw_meta_contracts(id) ON UPDATE CASCADE ON DELETE CASCADE,
	UNIQUE (address, contract_id)
);`, ExtAlias, ExtAlias)

	sqlGetRewardContractByAddress = fmt.Sprintf(`SELECT * FROM %s.erc20rw_meta_contracts WHERE chain_id = $chain_id AND address = $address`, ExtAlias)

	sqlIncrementSafeNonce = fmt.Sprintf(`{%s}UPDATE erc20rw_meta_contracts SET safe_nonce = safe_nonce + 1 WHERE id = $contract_id`, ExtAlias)

	//sqlNewSigner          = `INSERT INTO %s.erc20rw_meta_signers (id, address, contract_id) VALUES ($id, $address, $contract_id)`
	sqlListSigners        = fmt.Sprintf(`SELECT * FROM %s.erc20rw_meta_signers WHERE contract_id = $contract_id`, ExtAlias)
	sqlGetSignerByAddress = fmt.Sprintf(`SELECT * FROM %s.erc20rw_meta_signers WHERE address = $address AND contract_id = $contract_id`, ExtAlias)

	sqlCreateRewardContract = fmt.Sprintf(`{%s}INSERT INTO erc20rw_meta_contracts (id, chain_id, address, nonce, threshold, safe_address, safe_nonce)
VALUES ($contract_id,$chain_id,$address,$nonce,$threshold,$safe_address,$safe_nonce);`, ExtAlias)
)

type EngineExecutor interface {
	Execute(ctx *kcommon.EngineContext, db sql.DB, statement string, params map[string]any, fn func(*kcommon.Row) error) error
	ExecuteWithoutEngineCtx(ctx context.Context, db sql.DB, statement string, params map[string]any, fn func(row *kcommon.Row) error) error
}

type RewardContract struct {
	ID          *types.UUID
	ChainID     int64
	Address     string
	Nonce       int64
	Threshold   int64
	SafeAddress string
	SafeNonce   int64 // probably should use big int?

	Signers []string
}

type RewardSigner struct {
	ID         *types.UUID
	Address    string
	ContractID *types.UUID
}

func GenRewardContractID(chainID int64, address string) *types.UUID {
	return types.NewUUIDV5([]byte(fmt.Sprintf("erc20rw_meta_contracts_%v_%v", chainID, address)))
}

func GenSignerID(contractID *types.UUID, address string) *types.UUID {
	return types.NewUUIDV5([]byte(fmt.Sprintf("erc20rw_meta_signers_%v_%v", contractID.String(), address)))
}

// GetRewardContract returns correspond escrow contract and its signers.
// Will create a new record if it is not found.
func GetRewardContract(ctx *kcommon.EngineContext, ee EngineExecutor, db sql.DB, chainID int64,
	contractAddress string) (*RewardContract, error) {

	var rc *RewardContract
	err := ee.Execute(ctx, db, sqlGetRewardContractByAddress,
		map[string]any{
			"$chain_id": chainID,
			"$address":  contractAddress,
		},
		func(row *kcommon.Row) error {
			var err error
			rc, err = rowToRewardContract(row.Values)
			if err != nil {
				return err
			}
			return nil
		})

	if err != nil {
		if errors.Is(err, engine.ErrNamespaceNotFound) {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}

	err = ee.Execute(ctx, db, sqlListSigners,
		map[string]any{"$contract_id": rc.ID},
		func(row *kcommon.Row) error {
			rc.Signers = append(rc.Signers, row.Values[1].(string))
			return nil
		})
	if err != nil {
		return nil, err
	}

	return rc, nil
}

func rowToRewardContract(row []any) (*RewardContract, error) {
	id, ok := row[0].(*types.UUID)
	if !ok {
		return nil, fmt.Errorf("failed to convert id to UUID")
	}

	chainID, ok := row[1].(int64)
	if !ok {
		return nil, fmt.Errorf("failed to convert chain_id to int64")
	}

	address, ok := row[2].(string)
	if !ok {
		return nil, fmt.Errorf("failed to convert address to string")
	}

	nonce, ok := row[3].(int64)
	if !ok {
		return nil, fmt.Errorf("failed to convert nonce to int64")
	}

	threshold, ok := row[4].(int64)
	if !ok {
		return nil, fmt.Errorf("failed to convert threshold to int64")
	}

	safeAddress, ok := row[5].(string)
	if !ok {
		return nil, fmt.Errorf("failed to convert safeAddress to string")
	}

	safeNonce, ok := row[6].(int64)
	if !ok {
		return nil, fmt.Errorf("failed to convert safeNonce to int64")
	}

	return &RewardContract{
		ID:          id,
		ChainID:     chainID,
		Address:     address,
		Nonce:       nonce,
		Threshold:   threshold,
		SafeAddress: safeAddress,
		SafeNonce:   safeNonce,
	}, nil
}

func GetSigner(ctx *kcommon.EngineContext, engine EngineExecutor, db sql.DB, address string, contractID *types.UUID) (*RewardSigner, error) {
	var signer *RewardSigner
	err := engine.Execute(ctx, db, sqlGetSignerByAddress,
		map[string]any{
			"$address":     address,
			"$contract_id": contractID,
		},
		func(row *kcommon.Row) error {
			var err error
			signer, err = rowToSigner(row.Values)
			return err
		})
	if err != nil {
		return nil, err
	}
	return signer, nil
}

func rowToSigner(row []any) (*RewardSigner, error) {
	id, ok := row[0].(*types.UUID)
	if !ok {
		return nil, fmt.Errorf("failed to convert id to UUID")
	}
	address, ok := row[1].(string)
	if !ok {
		return nil, fmt.Errorf("failed to convert address to string")
	}
	if !ethCommon.IsHexAddress(address) { // todo on other queries
		return nil, fmt.Errorf("internal bug: invalid address: %s", address)
	}

	contractID, ok := row[2].(*types.UUID)
	if !ok {
		return nil, fmt.Errorf("failed to convert contract_id to UUID")
	}

	return &RewardSigner{
		ID:         id,
		Address:    address,
		ContractID: contractID,
	}, nil
}

func IncrementSafeNonce(ctx *kcommon.EngineContext, engine EngineExecutor, db sql.DB, contractID *types.UUID) error {
	return engine.Execute(ctx, db, sqlIncrementSafeNonce,
		map[string]any{"$contract_id": contractID}, nil)
}
