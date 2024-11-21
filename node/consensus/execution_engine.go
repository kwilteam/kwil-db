package consensus

import (
	"context"
	"crypto/sha256"
	"fmt"

	"github.com/kwilteam/kwil-db/core/crypto/auth"
	ktypes "github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/node/types"
)

// potentially our txApp
// - ConsensusTx
// 		- ExecuteTx(ctx, tx, nestedTx) -> TxResult
// - CommitTx
// - RollbackTx
// Cancellable context for tx execution

// Block Atomicity

// placeholder for our pg module constructs
// state.json will save the last commit info to tag the app state
// therefore blockExecutor should only track the intermediate state changes
// resulting from the txs within a block
type blockExecutor struct {
	changesets []types.Hash // changesets for each tx in a block, commit will give a hash of these changesets in a deterministic order
}

//	func (be *blockExecutor) BeginBlock() error {
//		return nil
//	}
func newBlockExecutor() *blockExecutor {
	return &blockExecutor{}
}

func (be *blockExecutor) Execute(_ context.Context, tx []byte) (ktypes.TxResult, error) {
	// unmashal tx
	var transaction ktypes.Transaction
	err := transaction.UnmarshalBinary(tx)
	if err != nil {
		return ktypes.TxResult{}, err
	}

	// validate tx
	if err := validateTransaction(&transaction); err != nil {
		return ktypes.TxResult{}, fmt.Errorf("invalid transaction: %w", err)
	}

	// execute tx
	hash := sha256.Sum256(tx)
	be.changesets = append(be.changesets, hash)
	return ktypes.TxResult{
		Code: 0,
		Log:  fmt.Sprintf("Success: %x", hash),
	}, nil
}

func validateTransaction(tx *ktypes.Transaction) error {
	// Signature validation
	authenticator := auth.GetAuthenticator(tx.Signature.Type)
	if authenticator == nil {
		return fmt.Errorf("unknown authenticator: %s", tx.Signature.Type)
	}

	txMsg, err := tx.SerializeMsg()
	if err != nil {
		return err
	}

	if err := authenticator.Verify(tx.Sender, txMsg, tx.Signature.Data); err != nil {
		return err
	}

	// Payload validation
	if !tx.Body.PayloadType.Valid() {
		return fmt.Errorf("invalid payload type: %s", tx.Body.PayloadType)
	}

	return nil
}

// Precommit gives a deterministic hash based on the changesets resulting from the txs in a block
func (be *blockExecutor) Precommit() (types.Hash, error) {
	hasher := sha256.New()
	for _, changeset := range be.changesets {
		hasher.Write(changeset[:])
	}
	return types.Hash(hasher.Sum(nil)), nil
}

func (be *blockExecutor) Commit(commitFn func() error) error {
	// updates to the state should be done in the commitFn
	if err := commitFn(); err != nil {
		return err
	}

	be.changesets = nil
	return nil
}

// TODO: not much of any use yet.
func (be *blockExecutor) Rollback() error {
	be.changesets = nil
	return nil
}
