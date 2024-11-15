package txapp

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"kwil/node/types/sql"
	"kwil/node/voting"
	"kwil/types"
	"math/big"
	"sync"
)

type mempool struct {
	accountMgr   Accounts
	validatorMgr Validators

	accounts map[string]*types.Account
	acctsMtx sync.Mutex // protects accounts

	nodeAddr []byte
}

// accountInfo retrieves the account info from the mempool state or the account store.
func (m *mempool) accountInfo(ctx context.Context, tx sql.Executor, acctID []byte) (*types.Account, error) {
	if acctInfo, ok := m.accounts[string(acctID)]; ok {
		return acctInfo, nil // there is an unconfirmed tx for this account
	}

	// get account from account store
	acct, err := m.accountMgr.GetAccount(ctx, tx, acctID)
	if err != nil {
		return nil, err
	}

	m.accounts[string(acctID)] = acct

	return acct, nil
}

// accountInfoSafe is wraps accountInfo in a mutex lock.
func (m *mempool) accountInfoSafe(ctx context.Context, tx sql.Executor, acctID []byte) (*types.Account, error) {
	m.acctsMtx.Lock()
	defer m.acctsMtx.Unlock()

	return m.accountInfo(ctx, tx, acctID)
}

// applyTransaction validates account specific info and applies valid transactions to the mempool state.
func (m *mempool) applyTransaction(ctx *types.TxContext, tx *types.Transaction, dbTx sql.Executor, rebroadcaster Rebroadcaster) error {
	m.acctsMtx.Lock()
	defer m.acctsMtx.Unlock()

	// if the network is in a migration, there are numerous
	// transaction types we must disallow.
	// see [internal/migrations/migrations.go] for more info
	status := ctx.BlockContext.ChainContext.NetworkParameters.MigrationStatus
	inMigration := status == types.MigrationInProgress || status == types.MigrationCompleted
	activeMigration := status != types.NoActiveMigration
	genesisMigration := status == types.GenesisMigration

	if inMigration {
		switch tx.Body.PayloadType {
		case types.PayloadTypeValidatorJoin:
			return fmt.Errorf("validator joins are not allowed during migration")
		case types.PayloadTypeValidatorLeave:
			return fmt.Errorf("validator leaves are not allowed during migration")
		case types.PayloadTypeValidatorApprove:
			return fmt.Errorf("validator approvals are not allowed during migration")
		case types.PayloadTypeValidatorRemove:
			return fmt.Errorf("validator removals are not allowed during migration")
		case types.PayloadTypeValidatorVoteIDs:
			return fmt.Errorf("validator vote ids are not allowed during migration")
		case types.PayloadTypeValidatorVoteBodies:
			return fmt.Errorf("validator vote bodies are not allowed during migration")
		// case types.PayloadTypeDeploySchema:
		// 	return fmt.Errorf("deploy schema transactions are not allowed during migration")
		// case types.PayloadTypeDropSchema:
		// 	return fmt.Errorf("drop schema transactions are not allowed during migration")
		case types.PayloadTypeTransfer:
			return fmt.Errorf("transfer transactions are not allowed during migration")
		}
	}

	// Migration proposals and its approvals are not allowed once the migration is approved
	if tx.Body.PayloadType == types.PayloadTypeCreateResolution {
		res := &types.CreateResolution{}
		if err := res.UnmarshalBinary(tx.Body.Payload); err != nil {
			return err
		}
		if (activeMigration || genesisMigration) && res.Resolution.Type == voting.StartMigrationEventType {
			return fmt.Errorf(" migration resolutions are not allowed during migration")
		}
	}

	if tx.Body.PayloadType == types.PayloadTypeApproveResolution {
		res := &types.ApproveResolution{}
		if err := res.UnmarshalBinary(tx.Body.Payload); err != nil {
			return err
		}

		// check if resolution is a migration resolution
		resolution, err := resolutionByID(ctx.Ctx, dbTx, res.ResolutionID)
		if err != nil {
			return errors.New("migration proposal not found")
		}

		if (activeMigration || genesisMigration) && resolution.Type == voting.StartMigrationEventType {
			return fmt.Errorf("approving migration resolutions are not allowed during migration")
		}
	}

	// seems like maybe this should go in the switch statement below,
	// but I put it here to avoid extra db call for account info
	if tx.Body.PayloadType == types.PayloadTypeValidatorVoteIDs {
		power, err := m.validatorMgr.GetValidatorPower(ctx.Ctx, dbTx, tx.Sender)
		if err != nil {
			return err
		}

		if power == 0 {
			return fmt.Errorf("only validators can submit validator vote transactions")
		}

		// reject the transaction if the number of voteIDs exceeds the limit
		voteID := &types.ValidatorVoteIDs{}
		err = voteID.UnmarshalBinary(tx.Body.Payload)
		if err != nil {
			return err
		}
		if maxVotes := ctx.BlockContext.ChainContext.NetworkParameters.MaxVotesPerTx; (int64)(len(voteID.ResolutionIDs)) > maxVotes {
			return fmt.Errorf("number of voteIDs exceeds the limit of %d", maxVotes)
		}
	}

	if tx.Body.PayloadType == types.PayloadTypeValidatorVoteBodies {
		// not sure if this is the right error code
		return fmt.Errorf("validator vote bodies can not enter the mempool, and can only be submitted during block proposal")
	}

	// get account info from mempool state or account store
	acct, err := m.accountInfo(ctx.Ctx, dbTx, tx.Sender)
	if err != nil {
		return err
	}

	// reject the transactions from unfunded user accounts in gasEnabled mode
	if !ctx.BlockContext.ChainContext.NetworkParameters.DisabledGasCosts && acct.Nonce == 0 && acct.Balance.Sign() == 0 {
		delete(m.accounts, string(tx.Sender))
		return types.ErrInsufficientBalance
	}

	// It is normally permissible to accept a transaction with the same nonce as
	// a tx already in mempool (but not in a block), however without gas we
	// would not want to allow that since there is no criteria for selecting the
	// one to mine (normally higher fee).
	if tx.Body.Nonce != uint64(acct.Nonce)+1 {
		// If the transaction with invalid nonce is a ValidatorVoteIDs transaction,
		// then mark the events for rebroadcast before discarding the transaction
		// as the votes for these events are not yet received by the network.

		fromLocalValidator := bytes.Equal(tx.Sender, m.nodeAddr) // Check if the transaction is from the local node

		if tx.Body.PayloadType == types.PayloadTypeValidatorVoteIDs && fromLocalValidator {
			// Mark these ids for rebroadcast
			voteID := &types.ValidatorVoteIDs{}
			err = voteID.UnmarshalBinary(tx.Body.Payload)
			if err != nil {
				return err
			}

			err = rebroadcaster.MarkRebroadcast(ctx.Ctx, voteID.ResolutionIDs)
			if err != nil {
				return err
			}
		}
		return fmt.Errorf("%w for account %s: got %d, expected %d", types.ErrInvalidNonce,
			hex.EncodeToString(tx.Sender), tx.Body.Nonce, acct.Nonce+1)
	}

	spend := big.NewInt(0).Set(tx.Body.Fee) // NOTE: this could be the fee *limit*, but it depends on how the modules work

	switch tx.Body.PayloadType {
	case types.PayloadTypeTransfer:
		transfer := &types.Transfer{}
		err = transfer.UnmarshalBinary(tx.Body.Payload)
		if err != nil {
			return err
		}

		amt, ok := big.NewInt(0).SetString(transfer.Amount, 10)
		if !ok {
			return types.ErrInvalidAmount
		}

		if amt.Cmp(&big.Int{}) < 0 {
			return errors.Join(types.ErrInvalidAmount, errors.New("negative transfer not permitted"))
		}

		if amt.Cmp(acct.Balance) > 0 {
			return types.ErrInsufficientBalance
		}

		spend.Add(spend, amt)
	}

	// We'd check balance against the total spend (fees plus value sent) if we
	// know gas is enabled. Transfers must be funded regardless of transaction
	// gas requirement:

	// if spend.Cmp(acct.balance) > 0 {
	// 	return errors.New("insufficient funds")
	// }

	// Since we're not yet operating with different policy depending on whether
	// gas is enabled for the chain, we're just going to reduce the account's
	// pending balance, but no lower than zero. Tx execution will handle it.
	if spend.Cmp(acct.Balance) > 0 {
		acct.Balance.SetUint64(0)
	} else {
		acct.Balance.Sub(acct.Balance, spend)
	}

	// Account nonces and spends tracked by mempool should be incremented only for the
	// valid transactions. This is to avoid the case where mempool rejects a transaction
	// due to insufficient balance, but the account nonce and spend are already incremented.
	// Due to which it accepts the next transaction with nonce+1, instead of nonce
	// (but Tx with nonce is never pushed to the consensus pool).
	acct.Nonce = int64(tx.Body.Nonce)

	return nil
}

// reset clears the in-memory unconfirmed account states.
// This should be done at the end of block commit.
func (m *mempool) reset() {
	m.acctsMtx.Lock()
	defer m.acctsMtx.Unlock()

	m.accounts = make(map[string]*types.Account)
}
