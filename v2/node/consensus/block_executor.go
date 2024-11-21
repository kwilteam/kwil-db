package consensus

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"sort"

	"kwil/crypto/auth"
	"kwil/node/meta"
	"kwil/node/types"
	ktypes "kwil/types"
)

var (
	dirtyHash = types.HashBytes([]byte("0x42"))
)

// Block processing methods
func (ce *ConsensusEngine) validateBlock(blk *types.Block) error {
	// Validate if this is the correct block proposal to be processed.
	if blk.Header.Version != types.BlockVersion {
		return fmt.Errorf("block version mismatch, expected %d, got %d", types.BlockVersion, blk.Header.Version)
	}

	if ce.state.lc.height+1 != blk.Header.Height {
		return fmt.Errorf("block proposal for height %d does not follow %d", blk.Header.Height, ce.state.lc.height)
	}

	if ce.state.lc.blkHash != blk.Header.PrevHash {
		return fmt.Errorf("prevBlockHash mismatch, expected %v, got %v", ce.state.lc.blkHash, blk.Header.PrevHash)
	}

	if blk.Header.PrevAppHash != ce.state.lc.appHash {
		return fmt.Errorf("apphash mismatch, expected %v, got %v", ce.state.lc.appHash, blk.Header.PrevAppHash)
	}

	if blk.Header.NumTxns != uint32(len(blk.Txns)) {
		return fmt.Errorf("transaction count mismatch, expected %d, got %d", blk.Header.NumTxns, len(blk.Txns))
	}

	// Verify the merkle root of the block transactions
	merkleRoot := blk.MerkleRoot()
	if merkleRoot != blk.Header.MerkleRoot {
		return fmt.Errorf("merkleroot mismatch, expected %v, got %v", merkleRoot, blk.Header.MerkleRoot)
	}

	// Verify other stuff such as validatorsetHash, signature of the block etc.
	return nil
}

// executeBlock executes all the transactions in the block under a single pg consensus transaction,
// enforcing the atomicity of the block execution. It also calculates the appHash for the block and
// precommits the changeset to the pg database.
func (ce *ConsensusEngine) executeBlock() (err error) {
	defer func() {
		ce.stateInfo.mtx.Lock()
		ce.stateInfo.status = Executed
		ce.stateInfo.mtx.Unlock()
	}()

	ctx := context.Background() // TODO: Use block context with the chain params and stuff.

	blkProp := ce.state.blkProp

	// Begin the block execution session
	if err := ce.txapp.Begin(ctx, blkProp.height); err != nil {
		ce.log.Error("Failed to begin the block execution", "height", blkProp.height, "err", err)
	}

	ce.state.consensusTx, err = ce.db.BeginPreparedTx(ctx)
	if err != nil {
		return fmt.Errorf("begin outer tx failed: %w", err)
	}

	// TODO: log tracker

	var txResults []ktypes.TxResult

	for _, tx := range ce.state.blkProp.blk.Txns {
		decodedTx := &ktypes.Transaction{}
		if err := decodedTx.UnmarshalBinary(tx); err != nil {
			ce.log.Error("Failed to unmarshal the block tx", "err", err)
			return err
		}
		txHash := sha256.Sum256(tx)

		auth := auth.GetAuthenticator(decodedTx.Signature.Type)

		identifier, err := auth.Identifier(decodedTx.Sender)
		if err != nil {
			ce.log.Error("Failed to get identifier for the block tx", "err", err)
		}

		txCtx := &ktypes.TxContext{
			Ctx:           ctx,
			TxID:          hex.EncodeToString(txHash[:]),
			Signer:        decodedTx.Sender,
			Authenticator: decodedTx.Signature.Type,
			Caller:        identifier,
			// BlockContext: blkCtx,
		}

		select {
		case <-ctx.Done(): // is this the best way to abort the block execution?
			ce.state.blockRes.ack = false
			ce.log.Info("Block execution cancelled", "height", ce.state.blkProp.height)
			return nil // or error? or trigger resetState?
		default:
			res := ce.txapp.Execute(txCtx, ce.state.consensusTx, decodedTx)
			txResult := ktypes.TxResult{
				Code: uint16(res.ResponseCode),
				Gas:  res.Spend,
			}
			if res.Error != nil {
				txResult.Log = res.Error.Error()
				ce.log.Debug("Failed to execute transaction", "tx", txHash, "err", res.Error)
			} else {
				txResult.Log = "success"
			}

			txResults = append(txResults, txResult)
		}
	}

	// TODO: Any updates to the consensus params

	// TODO: Broadcast events

	// TODO: Notify the changesets to the migrator

	blockCtx := &ktypes.BlockContext{ // TODO: fill in the network params once we have them
		Height: ce.state.blkProp.height,
		ChainContext: &ktypes.ChainContext{
			NetworkParameters: &ktypes.NetworkParameters{
				DisabledGasCosts: true,
				JoinExpiry:       14400,
				MaxBlockSize:     6 * 1024 * 1024,
			},
		},
		Proposer: ce.leader.Bytes(),
	}
	_, err = ce.txapp.Finalize(ctx, ce.state.consensusTx, blockCtx)
	if err != nil {
		ce.log.Error("Failed to finalize txapp", "err", err)
		// send a nack?
		return err
	}

	if err := meta.SetChainState(ctx, ce.state.consensusTx, ce.state.lc.height+1, ce.state.lc.appHash[:], true); err != nil {
		ce.log.Error("Failed to set chain state", "err", err)
		return err
	}

	// Create a new changeset processor
	csp := newChangesetProcessor()
	// "migrator" module subscribes to the changeset processor to store changesets during the migration
	csErrChan := make(chan error, 1)
	defer close(csErrChan)
	// TODO: Subscribe to the changesets
	go csp.BroadcastChangesets(ctx)

	appHash, err := ce.state.consensusTx.Precommit(ctx, csp.csChan)
	if err != nil {
		ce.log.Error("Failed to precommit the changeset", "err", err)
	}

	valUpdates := ce.validators.ValidatorUpdates()
	valUpdatesHash := validatorUpdatesHash(valUpdates)
	for k, v := range valUpdates {
		if v.Power == 0 {
			delete(ce.validatorSet, k)
		} else {
			ce.validatorSet[k] = ktypes.Validator{
				PubKey: v.PubKey,
				Power:  v.Power,
			}
		}
	}

	accountsHash := ce.accountsHash()
	txResultsHash := txResultsHash(txResults)

	nextHash := ce.nextAppHash(ce.state.lc.appHash, types.Hash(appHash), valUpdatesHash, accountsHash, txResultsHash)

	ce.state.blockRes = &blockResult{
		txResults: txResults,
		appHash:   nextHash,
		ack:       true, // for reannounce
	}

	ce.log.Info("Executed Block", "height", ce.state.blkProp.blk.Header.Height, "blkHash", ce.state.blkProp.blkHash, "appHash", ce.state.blockRes.appHash)
	return nil
}

// nextAppHash calculates the appHash that encapsulates the state changes occurred during the block execution.
// sha256(prevAppHash || changesetHash || valUpdatesHash || accountsHash || txResultsHash)
func (ce *ConsensusEngine) nextAppHash(prevAppHash, changesetHash, valUpdatesHash, accountsHash, txResultsHash types.Hash) types.Hash {
	hasher := sha256.New()

	hasher.Write(prevAppHash[:])
	hasher.Write(changesetHash[:])
	hasher.Write(valUpdatesHash[:])
	hasher.Write(accountsHash[:])
	hasher.Write(txResultsHash[:])

	ce.log.Info("AppState updates: ", "prevAppHash", prevAppHash, "changesetsHash", changesetHash, "valUpdatesHash", valUpdatesHash, "accountsHash", accountsHash, "txResultsHash", txResultsHash)
	return types.Hash(hasher.Sum(nil))
}

func txResultsHash(results []ktypes.TxResult) types.Hash {
	hasher := sha256.New()
	for _, res := range results {
		binary.Write(hasher, binary.BigEndian, res.Code)
		binary.Write(hasher, binary.BigEndian, res.Gas)
	}

	return types.Hash(hasher.Sum(nil))
}

func (ce *ConsensusEngine) accountsHash() types.Hash {
	accounts := ce.accounts.Updates()
	hasher := sha256.New()
	for _, acc := range accounts {
		hasher.Write(acc.Identifier)
		binary.Write(hasher, binary.BigEndian, acc.Balance)
		binary.Write(hasher, binary.BigEndian, acc.Nonce)
	}

	return types.Hash(hasher.Sum(nil))
}

func validatorUpdatesHash(updates map[string]*ktypes.Validator) types.Hash {
	// sort the updates by the validator address
	// hash the validator address and the validator struct

	keys := make([]string, 0, len(updates))
	for k := range updates {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	hash := sha256.New()
	for _, k := range keys {
		// hash the validator address
		hash.Write(updates[k].PubKey)
		// hash the validator power
		binary.Write(hash, binary.BigEndian, updates[k].Power)
	}

	return types.Hash(hash.Sum(nil))
}

// Commit method commits the block to the blockstore and postgres database.
// It also updates the txIndexer and mempool with the transactions in the block.
func (ce *ConsensusEngine) commit() error {
	// TODO: Lock mempool and update the mempool to remove the transactions in the block
	// Mempool should not receive any new transactions until this Commit is done as
	// we are updating the state and the tx checks should be done against the new state.
	ctx := context.Background()
	blkProp := ce.state.blkProp
	height, appHash := ce.state.blkProp.height, ce.state.blockRes.appHash

	if err := ce.blockStore.Store(blkProp.blk, appHash); err != nil {
		return err
	}

	if err := ce.blockStore.StoreResults(blkProp.blkHash, ce.state.blockRes.txResults); err != nil {
		return err
	}

	// Commit the Postgres Consensus transaction
	if err := ce.state.consensusTx.Commit(ctx); err != nil {
		return err
	}

	// Update the chain meta store with the new height and the dirty
	ctxS := context.Background()
	tx, err := ce.db.BeginTx(ctxS)
	if err != nil {
		return err
	}

	if err := meta.SetChainState(ctx, tx, height, appHash[:], false); err != nil {
		err2 := tx.Rollback(ctxS)
		if err2 != nil {
			ce.log.Error("Failed to rollback the transaction", "err", err2)
			return err2
		}
		return err
	}

	if err := tx.Commit(ctxS); err != nil {
		return err
	}

	if err := ce.txapp.Commit(); err != nil {
		return err
	}

	// remove transactions from the mempool
	for _, txn := range blkProp.blk.Txns {
		txHash := types.HashBytes(txn)
		ce.mempool.Store(txHash, nil)
	}

	// TODO: set the role based on the final validators

	ce.log.Info("Committed Block", "height", height, "hash", blkProp.blkHash, "appHash", appHash.String())
	return nil
}

func (ce *ConsensusEngine) nextState() {
	ce.state.lc = &lastCommit{
		height:  ce.state.blkProp.height,
		blkHash: ce.state.blkProp.blkHash,
		appHash: ce.state.blockRes.appHash,
		blk:     ce.state.blkProp.blk,
	}

	ce.state.blkProp = nil
	ce.state.blockRes = nil
	ce.state.votes = make(map[string]*vote)
	ce.state.consensusTx = nil

	// update the stateInfo
	ce.stateInfo.mtx.Lock()
	ce.stateInfo.status = Committed
	ce.stateInfo.blkProp = nil
	ce.stateInfo.height = ce.state.lc.height
	ce.stateInfo.mtx.Unlock()
}

func (ce *ConsensusEngine) resetState(ctx context.Context) error {
	if ce.state.consensusTx != nil {
		err := ce.state.consensusTx.Rollback(ctx)
		if err != nil {
			return err
		}
	}

	ce.state.blkProp = nil
	ce.state.blockRes = nil
	ce.state.votes = make(map[string]*vote)
	ce.state.consensusTx = nil

	// update the stateInfo
	ce.stateInfo.mtx.Lock()
	ce.stateInfo.status = Committed
	ce.stateInfo.blkProp = nil
	ce.stateInfo.height = ce.state.lc.height
	ce.stateInfo.mtx.Unlock()

	return nil
}

// ChangesetProcessor is a PubSub that listens for changesets and broadcasts them to the receivers.
// Subscribers can be added and removed to listen for changesets.
// Statistics receiver might listen for changesets to update the statistics every block.
// Whereas Network migrations listen for the changesets only during the migration. (that's when you register)
// ABCI --> CS Processor ---> [Subscribers]
// Once all the changesets are processed, all the channels are closed [every block]
// The channels are reset for the next block.
type changesetProcessor struct {
	// channel to receive changesets
	// closed by the pgRepl layer after all the block changes have been pushed to the processor
	csChan chan any

	// subscribers to the changeset processor are the receivers of the changesets
	// Examples: Statistics receiver, Network migration receiver
	subscribers map[string]chan<- any
}

func newChangesetProcessor() *changesetProcessor {
	return &changesetProcessor{
		csChan:      make(chan any, 1), // buffered channel to avoid blocking
		subscribers: make(map[string]chan<- any),
	}
}

// Subscribe adds a subscriber to the changeset processor's subscribers list.
// The receiver can subscribe to the changeset processor using a unique id.
func (c *changesetProcessor) Subscribe(ctx context.Context, id string) (<-chan any, error) {
	_, ok := c.subscribers[id]
	if ok {
		return nil, fmt.Errorf("subscriber with id %s already exists", id)
	}

	ch := make(chan any, 1) // buffered channel to avoid blocking
	c.subscribers[id] = ch
	return ch, nil
}

// Unsubscribe removes the subscriber from the changeset processor.
func (c *changesetProcessor) Unsubscribe(ctx context.Context, id string) error {
	if ch, ok := c.subscribers[id]; ok {
		// close the channel to signal the subscriber to stop listening
		close(ch)
		delete(c.subscribers, id)
		return nil
	}

	return fmt.Errorf("subscriber with id %s does not exist", id)
}

// Broadcast sends changesets to all the subscribers through their channels.
func (c *changesetProcessor) BroadcastChangesets(ctx context.Context) {
	defer c.Close() // All the block changesets have been processed, signal subscribers to stop listening.

	// Listen on the csChan for changesets and broadcast them to all subscribers.
	for cs := range c.csChan {
		for _, ch := range c.subscribers {
			select {
			case ch <- cs:
			case <-ctx.Done():
				return
			}
		}
	}
}

func (c *changesetProcessor) Close() {
	// c.CsChan is closed by the repl layer (sender closes the channel)
	for _, ch := range c.subscribers {
		close(ch)
	}
}
