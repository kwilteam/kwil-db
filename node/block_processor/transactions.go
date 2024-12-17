package blockprocessor

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"sort"

	"github.com/kwilteam/kwil-db/core/types"
	ktypes "github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/node/types/sql"
)

// txSubList implements sort.Interface to perform in-place sorting of a slice
// that is a subset of another slice, reordering in both while staying within
// the subsets positions in the parent slice.
//
// For example:
//
//	parent slice: {a0, b2, b0, a1, b1}
//	b's subset: {b2, b0, b1}
//	sorted subset: {b0, b1, b2}
//	parent slice: {a0, b0, b1, a1, b2}
//
// The set if locations used by b elements within the parent slice is unchanged,
// but the elements are sorted.
type txSubList struct {
	sub   []*indexedTxn // sort.Sort references only this with Len and Less
	super []*indexedTxn // sort.Sort also Swaps in super using the i field
}

func (txl txSubList) Len() int {
	return len(txl.sub)
}

func (txl txSubList) Less(i int, j int) bool {
	a, b := txl.sub[i], txl.sub[j]
	return a.Body.Nonce < b.Body.Nonce
}

func (txl txSubList) Swap(i int, j int) {
	// Swap elements in sub.
	txl.sub[i], txl.sub[j] = txl.sub[j], txl.sub[i]
	// Swap the elements in their positions in super.
	ip, jp := txl.sub[i].i, txl.sub[j].i
	txl.super[ip], txl.super[jp] = txl.super[jp], txl.super[ip]
}

// indexedTxn facilitates in-place sorting of transaction slices that are
// subsets of other larger slices using a txSubList. This is only used within
// prepareMempoolTxns, and is package-level rather than scoped to the function
// because we define methods to implement sort.Interface.
type indexedTxn struct {
	i int // index in superset slice
	*types.Transaction

	is int // not used for sorting, only referencing the marshalled txn slice
}

// prepareBlockTransactions is used by the leader to prepare block transactions.
// It ensures nonce ordering, removes transactions from unfunded accounts,
// enforces block size limits, and applies the maxVotesPerTx limit for voteID transactions.
// Additionally, it includes the ValidatorVoteBody transaction for unresolved events.
// The final transaction order is: MempoolProposerTxns, ValidatorVoteBodyTx, Other MempoolTxns (Nonce ordered, stable sorted).
func (bp *BlockProcessor) prepareBlockTransactions(ctx context.Context, txs [][]byte) (finalTxs [][]byte, invalidTxs [][]byte, err error) {
	// Unmarshal and index the transactions.
	var okTxns []*indexedTxn
	invalidTxs = make([][]byte, 0, len(txs))
	var i int

	for is, tx := range txs {
		txn := &types.Transaction{}
		if err = txn.UnmarshalBinary(tx); err != nil {
			invalidTxs = append(invalidTxs, tx)
			bp.log.Error("Failed to unmarshal transaction that was previously accepted to mempool", "error", err)
			continue
		}
		okTxns = append(okTxns, &indexedTxn{i, txn, is})
		i++
	}

	// Group by sender and stable sort each group by nonce.
	grouped := make(map[string][]*indexedTxn)
	for _, txn := range okTxns {
		key := string(txn.Sender)
		grouped[key] = append(grouped[key], txn)
	}

	for _, group := range grouped {
		sort.Stable(txSubList{group, okTxns})
	}

	nonces := make([]uint64, 0, len(okTxns))
	var propTxs, otherTxns []*indexedTxn
	i = 0
	proposerNonce := uint64(0)

	readTx, err := bp.db.BeginReadTx(ctx)
	if err != nil {
		bp.log.Error("Failed to begin read transaction", "error", err)
		return nil, nil, err
	}
	defer readTx.Rollback(ctx)

	// Enfore nonce ordering and remove transactions from the unfunded accounts
	for _, tx := range okTxns {
		if i > 0 && tx.Body.Nonce == nonces[i-1] && bytes.Equal(tx.Sender, okTxns[i-1].Sender) {
			invalidTxs = append(invalidTxs, txs[tx.is])
			bp.log.Warn("Transaction has a duplicate nonce", "tx", tx)
			continue
		}

		// Enforce maxVptesPerTx limit for voteID transactions
		if tx.Body.PayloadType == types.PayloadTypeValidatorVoteIDs {
			voteIDs := &types.ValidatorVoteIDs{}
			if err = voteIDs.UnmarshalBinary(tx.Body.Payload); err != nil {
				invalidTxs = append(invalidTxs, txs[tx.is])
				bp.log.Warn("Dropping voteID tx: failed to unmarshal ValidatorVoteIDs transaction", "error", err)
				continue
			}

			if len(voteIDs.ResolutionIDs) > int(bp.chainCtx.NetworkParameters.MaxVotesPerTx) {
				invalidTxs = append(invalidTxs, txs[tx.is])
				bp.log.Warn("Dropping voteID tx: exceeds max votes per tx", "numVotes", len(voteIDs.ResolutionIDs),
					"maxVotes", bp.chainCtx.NetworkParameters.MaxVotesPerTx)
				continue
			}
		}

		// Drop transactions from unfunded accounts in gasEnabled mode
		if !bp.chainCtx.NetworkParameters.DisabledGasCosts {
			balance, nonce, err := bp.AccountInfo(ctx, readTx, tx.Sender, false)
			if err != nil {
				bp.log.Error("failed to get account info", "error", err)
				continue
			}

			if nonce == 0 && balance.Sign() == 0 {
				invalidTxs = append(invalidTxs, txs[tx.is])
				bp.log.Warn("Dropping tx from unfunded account while preparing the block", "account", hex.EncodeToString(tx.Sender))
				continue
			}
		}

		if bytes.Equal(tx.Sender, bp.signer.Identity()) {
			proposerNonce = tx.Body.Nonce
			propTxs = append(propTxs, tx)
		} else {
			otherTxns = append(otherTxns, tx)
		}
		nonces = append(nonces, tx.Body.Nonce)
		i++
	}

	// Enforce block size limits
	// Txs order: MempoolProposerTxns, ProposerInjectedTxns, MempoolTxns

	finalTxs = make([][]byte, 0)
	maxTxBytes := bp.chainCtx.NetworkParameters.MaxBlockSize

	for _, tx := range propTxs {
		txBts := txs[tx.is]
		txSize := int64(len(txBts))
		if maxTxBytes < txSize {
			break
		}
		maxTxBytes -= txSize
		finalTxs = append(finalTxs, txBts)
	}

	var voteBodyTx []byte // TODO: check proposerNonce value again
	voteBodyTx, err = bp.prepareValidatorVoteBodyTx(ctx, int64(proposerNonce), maxTxBytes)
	if err != nil {
		bp.log.Error("Failed to prepare validator vote body transaction", "error", err)
		return nil, nil, err
	}
	if voteBodyTx != nil {
		finalTxs = append(finalTxs, voteBodyTx)
		maxTxBytes -= int64(len(voteBodyTx))
	}

	// senders tracks the sender of transactions that has pushed over the bytes limit for the block.
	// If a sender is in the senders, skip all subsequent transactions from the sender
	// because nonces need to be sequential.
	// Keep checking transactions for other senders that may be smaller and fit in the remaining space.
	senders := make(map[string]bool)
	for _, tx := range otherTxns {
		sender := string(tx.Sender)
		// if the sender is already in the skipped senders, skip the transaction
		if _, ok := senders[sender]; ok {
			continue
		}

		txSize := int64(len(txs[tx.is]))
		if maxTxBytes < txSize {
			// Ignore the transaction and all subsequent transactions from the sender
			senders[sender] = true
			continue
		}

		maxTxBytes -= txSize
		finalTxs = append(finalTxs, txs[tx.is])
	}

	return finalTxs, invalidTxs, err
}

// prepareValidatorVoteBodyTx authors the ValidatorVoteBody transaction to be included by the leader in the block.
// It fetches the events which does not have resolutions yet and creates a validator vote body transaction.
// The number of events to be included in a single transaction is limited either by MaxVotesPerTx or the maxTxSize
// whichever is reached first. The estimated fee for validatorVOteBodies transaction is directly proportional to
// the size of the event body. The transaction is signed by the leader and returned.
func (bp *BlockProcessor) prepareValidatorVoteBodyTx(ctx context.Context, nonce int64, maxTxSize int64) ([]byte, error) {
	readTx, err := bp.db.BeginReadTx(ctx)
	if err != nil {
		return nil, err
	}
	defer readTx.Rollback(ctx)

	bal, n, err := bp.AccountInfo(ctx, readTx, bp.signer.Identity(), false)
	if err != nil {
		return nil, err
	}

	if nonce == 0 {
		nonce = n
	}

	events, err := getEvents(ctx, readTx)
	if err != nil {
		return nil, err
	}

	if len(events) == 0 {
		bp.log.Debug("No events to propose for voting")
		return nil, nil
	}

	// If gas costs are enabled, ensure that the node has sufficient funds to include this transaction
	if !bp.chainCtx.NetworkParameters.DisabledGasCosts && n == 0 && bal.Sign() == 0 {
		bp.log.Debug("Leader account has no balance, not allowed to propose any transactions")
		return nil, nil
	}

	ids := make([]*types.UUID, 0, len(events))
	for _, e := range events {
		ids = append(ids, e.ID())
	}

	// Limit only upto MaxVoteIDsPerTx events to be included in a single transaction
	if len(ids) > int(bp.chainCtx.NetworkParameters.MaxVotesPerTx) {
		ids = ids[:bp.chainCtx.NetworkParameters.MaxVotesPerTx]
	}

	eventMap := make(map[types.UUID]*types.VotableEvent)
	for _, evt := range events {
		eventMap[*evt.ID()] = evt
	}

	emptyTxSz, err := bp.emptyVoteBodyTxSize()
	if err != nil {
		return nil, err
	}
	maxTxSize -= emptyTxSz

	var finalEvents []*types.VotableEvent
	estimatedTxSize := 0

	for _, id := range ids {
		evt, ok := eventMap[*id]
		if !ok {
			bp.log.Error("Event not found in event map", "eventID", id)
			return nil, fmt.Errorf("event  %s not found in event map", id.String())
		}

		evtSz := 4 + 4 + len(evt.Type) + len(evt.Body)
		estimatedTxSize = evtSz
		if int64(evtSz) > maxTxSize {
			bp.log.Debug("reached maximum proposer tx size", "maxTxSize", maxTxSize, "evtSz", evtSz)
			break
		}

		maxTxSize -= int64(evtSz)
		finalEvents = append(finalEvents, &types.VotableEvent{
			Type: evt.Type,
			Body: evt.Body,
		})
	}

	if len(finalEvents) == 0 && len(ids) > 0 {
		bp.log.Warn("found proposer events to propose, but cannot fit them in a block", "maxTxSize", maxTxSize,
			"numEvents", len(ids), "maxVotesPerTx", bp.chainCtx.NetworkParameters.MaxVotesPerTx,
			"estimatedTxSize", estimatedTxSize)
		return nil, nil
	}

	tx, err := types.CreateTransaction(&types.ValidatorVoteBodies{
		Events: finalEvents,
	}, bp.chainCtx.ChainID, uint64(nonce)+1)
	if err != nil {
		return nil, err
	}

	// Fee estimation
	fee, err := bp.Price(ctx, readTx, tx)
	if err != nil {
		return nil, err
	}
	tx.Body.Fee = fee

	if err = tx.Sign(bp.signer); err != nil {
		return nil, err
	}

	bts, err := tx.MarshalBinary()
	if err != nil {
		return nil, err
	}

	bp.log.Info("Created a ValidatorVoteBody transaction", "events", len(finalEvents), "nonce", n, "GasPrice", fee.String())

	return bts, nil
}

// emptyVodeBodyTxSize returns the size of an empty validator vote body transaction.
// used to estimate the size of the validator vote body transactions with events as the
// size is directly proportional to the events size.
func (bp *BlockProcessor) emptyVoteBodyTxSize() (int64, error) {
	payload := &types.ValidatorVoteBodies{
		Events: []*types.VotableEvent{},
	}
	tx, err := types.CreateTransaction(payload, bp.chainCtx.ChainID, 0)
	if err != nil {
		return -1, err
	}

	err = tx.Sign(bp.signer)
	if err != nil {
		return -1, err
	}

	bts, err := tx.MarshalBinary()
	if err != nil {
		return -1, err
	}

	return int64(len(bts)), nil
}

func (bp *BlockProcessor) BroadcastVoteIDTx(ctx context.Context, db sql.DB) error {
	tx, ids, err := bp.PrepareValidatorVoteIDTx(ctx, db)
	if err != nil {
		return err
	}

	if tx == nil || len(ids) == 0 { // no voteIDs to broadcast
		return nil
	}

	_, err = bp.broadcastTxFn(ctx, tx, 0)
	if err != nil {
		return err
	}

	return bp.events.MarkBroadcasted(ctx, ids)
}

func (bp *BlockProcessor) PrepareValidatorVoteIDTx(ctx context.Context, db sql.DB) (*ktypes.Transaction, []*types.UUID, error) {
	readTx, err := bp.db.BeginReadTx(ctx)
	if err != nil {
		bp.log.Error("Failed to begin read transaction while preparing voteID Tx", "error", err)
		return nil, nil, err
	}
	defer readTx.Rollback(ctx)

	// Only validators can issue voteID transactions not the leader or sentry nodes

	// check if the node is a leader
	if bytes.Equal(bp.signer.Identity(), bp.genesisParams.Leader) {
		bp.log.Debug("Leader node is not allowed to propose voteID transactions")
		return nil, nil, nil
	}

	// check if the node is a sentry node
	vals := bp.GetValidators()
	found := false
	for _, val := range vals {
		if bytes.Equal(val.PubKey, bp.signer.Identity()) {
			found = true
		}
	}

	if !found {
		bp.log.Debug("Sentry node is not allowed to propose voteID transactions")
		return nil, nil, nil
	}

	// Vote only if the voter observed the event corresponding to the resolution.
	// ids are the resolution ids that the validator witnessed the events for and can vote on.
	ids, err := bp.events.GetUnbroadcastedEvents(ctx)
	if err != nil {
		return nil, nil, err
	}

	if len(ids) == 0 {
		bp.log.Debug("no voteIDs to broadcast")
		return nil, nil, nil
	}

	// consider only the first maxVoteIDsPerTx events, to limit the postgres access roundtrips per block execution.
	if len(ids) > int(bp.chainCtx.NetworkParameters.MaxVotesPerTx) {
		ids = ids[:bp.chainCtx.NetworkParameters.MaxVotesPerTx]
	}

	bal, nonce, err := bp.AccountInfo(ctx, readTx, bp.signer.Identity(), true)
	if err != nil {
		return nil, nil, err
	}

	tx, err := types.CreateTransaction(&types.ValidatorVoteIDs{ResolutionIDs: ids}, bp.chainCtx.ChainID, uint64(nonce)+1)
	if err != nil {
		return nil, nil, err
	}

	// Fee estimation
	fee, err := bp.Price(ctx, readTx, tx)
	if err != nil {
		return nil, nil, err
	}
	tx.Body.Fee = fee

	// check if the node has enough balance to propose the transaction
	if bal.Cmp(fee) < 0 {
		bp.log.Warnf("skipping voteID broadcast: not enough balance to pay for the tx fee, balance: %s, fee: %s", bal.String(), fee.String())
		return nil, nil, nil
	}

	if err = tx.Sign(bp.signer); err != nil {
		return nil, nil, err
	}

	return tx, ids, nil
}
