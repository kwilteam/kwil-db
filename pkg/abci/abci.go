package abci

import (
	"context"
	"fmt"
	"sync"

	abciTypes "github.com/cometbft/cometbft/abci/types"
	engineTypes "github.com/kwilteam/kwil-db/pkg/engine/types"
	"github.com/kwilteam/kwil-db/pkg/log"
	"github.com/kwilteam/kwil-db/pkg/modules/datasets"
	"github.com/kwilteam/kwil-db/pkg/transactions"
	"go.uber.org/zap"
)

func NewAbciApp(database DatasetsModule, validators ValidatorModule, committer AtomicCommitter,
	snapshotter SnapshotModule,
	bootstrapper DBBootstrapModule,
	opts ...AbciOpt) *AbciApp {
	app := &AbciApp{
		database:     database,
		validators:   validators,
		committer:    committer,
		snapshotter:  snapshotter,
		bootstrapper: bootstrapper,

		log: log.NewNoOp(),

		commitWaiter: sync.WaitGroup{},
	}

	for _, opt := range opts {
		opt(app)
	}

	return app
}

type AbciApp struct {
	// database is the database module that handles database deployment, dropping, and execution
	database DatasetsModule

	// validators is the validators module that handles joining and approving validators
	validators ValidatorModule

	// committer is the atomic committer that handles atomic commits across multiple stores
	committer AtomicCommitter

	snapshotter SnapshotModule

	bootstrapper DBBootstrapModule

	log log.Logger

	// commitWaiter is a waitgroup that waits for the commit to finish
	// when a block is begun, the commitWaiter waits until the previous commit is finished
	// it then increments and starts "begin block"
	// when a commit is finished, the commitWaiter is decremented
	commitWaiter sync.WaitGroup
}

func (a *AbciApp) ApplySnapshotChunk(p0 abciTypes.RequestApplySnapshotChunk) abciTypes.ResponseApplySnapshotChunk {
	refetch_chunks, err := a.bootstrapper.ApplySnapshotChunk(p0.Chunk, p0.Index)
	if err != nil {
		return abciTypes.ResponseApplySnapshotChunk{Result: abciTypes.ResponseApplySnapshotChunk_UNKNOWN, RefetchChunks: refetch_chunks}
	}

	if a.bootstrapper.IsDBRestored() {
		/*
			TODO: Update the app hash & app height here once we introduce app specific state.
			Comet uses ABCIInfo to query & verify the app hash and app height at the end of the state sync process.
			If the app hash and app height are not updated here, Comet will do block sync.

			TODO: Check how ABCI Init is called in state sync vs block sync.
		*/
		a.log.Info("Bootstrapped database successfully")
	}
	return abciTypes.ResponseApplySnapshotChunk{Result: abciTypes.ResponseApplySnapshotChunk_ACCEPT, RefetchChunks: nil}
}

// BeginBlock begins a block.
// If the previous commit is not finished, it will wait for the previous commit to finish.
func (a *AbciApp) BeginBlock(p0 abciTypes.RequestBeginBlock) abciTypes.ResponseBeginBlock {
	a.commitWaiter.Wait()
	a.commitWaiter.Add(1)

	err := a.committer.Begin(context.Background())
	if err != nil {
		a.log.Error("failed to begin atomic commit", zap.Error(err))
		return abciTypes.ResponseBeginBlock{}
	}

	return abciTypes.ResponseBeginBlock{}
}

func (a *AbciApp) CheckTx(p0 abciTypes.RequestCheckTx) abciTypes.ResponseCheckTx {
	panic("TODO")
}

// Commit commits a block.
// It will commit all changes to a wal, and then asynchronously apply the changes to the database.
func (a *AbciApp) Commit() abciTypes.ResponseCommit {
	ctx := context.Background()
	appHash, err := a.committer.Commit(ctx, func(err error) {
		if err != nil {
			a.log.Error("failed to apply atomic commit", zap.Error(err))
		}

		a.commitWaiter.Done()
	})
	if err != nil {
		a.log.Error("failed to commit atomic commit", zap.Error(err))
		return abciTypes.ResponseCommit{}
	}

	return abciTypes.ResponseCommit{
		// TODO: is this where appHash belongs? >>> Correct
		Data: appHash,
	}
}

func (a *AbciApp) DeliverTx(req abciTypes.RequestDeliverTx) abciTypes.ResponseDeliverTx {
	ctx := context.Background()

	tx := &transactions.Transaction{}
	err := tx.UnmarshalBinary(req.Tx)
	if err != nil {
		return abciTypes.ResponseDeliverTx{
			Code: 1,
			Log:  err.Error(),
		}
	}

	var res *transactions.TransactionStatus

	switch tx.Body.PayloadType {
	case transactions.PayloadTypeDeploySchema:
		var schemaPayload transactions.Schema
		err = schemaPayload.UnmarshalBinary(tx.Body.Payload)
		if err != nil {
			break
		}

		var schema *engineTypes.Schema
		schema, err = datasets.ConvertSchemaToEngine(&schemaPayload)
		if err != nil {
			break
		}

		res, err = a.database.Deploy(ctx, schema, tx)
	case transactions.PayloadTypeDropSchema:
		drop := &transactions.DropSchema{}
		err = drop.UnmarshalBinary(tx.Body.Payload)
		if err != nil {
			break
		}

		res, err = a.database.Drop(ctx, drop.DBID, tx)
	case transactions.PayloadTypeExecuteAction:
		execution := &transactions.ActionExecution{}
		err = execution.UnmarshalBinary(tx.Body.Payload)
		if err != nil {
			break
		}

		res, err = a.database.Execute(ctx, execution.DBID, execution.Action, convertArgs(execution.Arguments), tx)
	case transactions.PayloadTypeValidatorJoin:
		// TODO: update this with validator payload
		panic("TODO")
		/*
			validatorJoin := &PayloadValidatorJoin{}
			err = a.payloadEncoder.Decode(tx.Payload, validatorJoin)
			if err != nil {
				break
			}

			res, err = a.validators.ValidatorJoin(ctx, validatorJoin.Address, tx)
		*/
	case transactions.PayloadTypeValidatorApprove:
		// TODO: update this with validator payload
		panic("TODO")
	/*
		validatorApprove := &PayloadValidatorApprove{}
		err = a.payloadEncoder.Decode(tx.Payload, validatorApprove)
		if err != nil {
			break
		}

		res, err = a.validators.ValidatorApprove(ctx, validatorApprove.ValidatorToApprove, validatorApprove.ApprovedBy, tx)
	*/
	default:
		err = fmt.Errorf("unknown payload type: %s", tx.Body.PayloadType.String())
	}
	if err != nil {
		return abciTypes.ResponseDeliverTx{
			Code: 1,
			Log:  err.Error(),
		}
	}

	return abciTypes.ResponseDeliverTx{
		Code:    abciTypes.CodeTypeOK,
		GasUsed: res.Fee.Int64(),
	}
}

func (a *AbciApp) EndBlock(p0 abciTypes.RequestEndBlock) abciTypes.ResponseEndBlock {
	panic("TODO")
}

func (a *AbciApp) Info(p0 abciTypes.RequestInfo) abciTypes.ResponseInfo {
	panic("TODO")
}

func (a *AbciApp) InitChain(p0 abciTypes.RequestInitChain) abciTypes.ResponseInitChain {
	panic("TODO")
}

func (a *AbciApp) ListSnapshots(p0 abciTypes.RequestListSnapshots) abciTypes.ResponseListSnapshots {
	snapshots, err := a.snapshotter.ListSnapshots()
	if err != nil {
		return abciTypes.ResponseListSnapshots{Snapshots: nil}
	}

	var res []*abciTypes.Snapshot
	for _, snapshot := range snapshots {
		abcisnapshot, err := convertToABCISnapshot(&snapshot)
		if err != nil {
			return abciTypes.ResponseListSnapshots{Snapshots: nil}
		}
		res = append(res, abcisnapshot)
	}
	return abciTypes.ResponseListSnapshots{Snapshots: res}
}

func (a *AbciApp) LoadSnapshotChunk(p0 abciTypes.RequestLoadSnapshotChunk) abciTypes.ResponseLoadSnapshotChunk {
	chunk := a.snapshotter.LoadSnapshotChunk(p0.Height, p0.Format, p0.Chunk)
	return abciTypes.ResponseLoadSnapshotChunk{Chunk: chunk}
}

func (a *AbciApp) OfferSnapshot(p0 abciTypes.RequestOfferSnapshot) abciTypes.ResponseOfferSnapshot {
	snapshot := convertABCISnapshots(p0.Snapshot)
	if (a.bootstrapper.OfferSnapshot(snapshot)) != nil {
		return abciTypes.ResponseOfferSnapshot{Result: abciTypes.ResponseOfferSnapshot_REJECT}
	}
	return abciTypes.ResponseOfferSnapshot{Result: abciTypes.ResponseOfferSnapshot_ACCEPT}

}
func (a *AbciApp) PrepareProposal(p0 abciTypes.RequestPrepareProposal) abciTypes.ResponsePrepareProposal {
	panic("TODO")
}

func (a *AbciApp) ProcessProposal(p0 abciTypes.RequestProcessProposal) abciTypes.ResponseProcessProposal {
	panic("TODO")
}

func (a *AbciApp) Query(p0 abciTypes.RequestQuery) abciTypes.ResponseQuery {
	panic("TODO")
}

// convertArgs converts the string args to type any.
func convertArgs(args [][]string) [][]any {
	converted := make([][]any, len(args))
	for i, arg := range args {
		converted[i] = make([]any, len(arg))
		for j, a := range arg {
			converted[i][j] = a
		}
	}

	return converted
}
