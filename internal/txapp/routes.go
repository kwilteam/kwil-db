package txapp

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math/big"

	"github.com/kwilteam/kwil-db/core/types/transactions"
	engineTypes "github.com/kwilteam/kwil-db/internal/engine/types"
	"github.com/kwilteam/kwil-db/internal/ident"
	"github.com/kwilteam/kwil-db/internal/voting"
)

func init() {
	err := errors.Join(
		registerRoute(transactions.PayloadTypeDeploySchema.String(), &deployDatasetRoute{}),
		registerRoute(transactions.PayloadTypeDropSchema.String(), &dropDatasetRoute{}),
		registerRoute(transactions.PayloadTypeExecuteAction.String(), &executeActionRoute{}),
		registerRoute(transactions.PayloadTypeTransfer.String(), &transferRoute{}),
		registerRoute(transactions.PayloadTypeValidatorJoin.String(), &validatorJoinRoute{}),
		registerRoute(transactions.PayloadTypeValidatorApprove.String(), &validatorApproveRoute{}),
		registerRoute(transactions.PayloadTypeValidatorRemove.String(), &validatorRemoveRoute{}),
		registerRoute(transactions.PayloadTypeValidatorLeave.String(), &validatorLeaveRoute{}),
		registerRoute(transactions.PayloadTypeValidatorVoteIDs.String(), &validatorVoteIDsRoute{}),
		registerRoute(transactions.PayloadTypeValidatorVoteBodies.String(), &validatorVoteBodiesRoute{}),
	)
	if err != nil {
		panic(fmt.Sprintf("failed to register routes: %s", err))
	}
}

type Route interface {
	Pricer
	// Execute is responsible for committing or rolling back transactions.
	// All transactions should spend, regardless of success or failure.
	// Therefore, a nested transaction should be used for all database
	// operations after the initial checkAndSpend.
	Execute(ctx TxContext, router *TxApp, tx *transactions.Transaction) *TxResponse
}

// TxContext is the context for transaction execution.
type TxContext struct {
	Ctx context.Context
	// BlockHeight gets the height of the current block.
	BlockHeight int64
	// Proposer gets the proposer public key of the current block.
	Proposer []byte
	// VotingPeriod is the maximum length of a voting period.
	// It is measured in blocks, and is applied additively.
	// e.g. if the current block is 50, and VotingPeriod is 100,
	// then the current voting period ends at block 150.
	VotingPeriod int64
}

type Pricer interface {
	Price(ctx context.Context, router *TxApp, tx *transactions.Transaction) (*big.Int, error)
}

// routes is a map of transaction payload types to their respective routes
var routes = map[string]Route{}

func registerRoute(payloadType string, route Route) error {
	_, ok := routes[payloadType]
	if ok {
		return fmt.Errorf("route for payload type %s already exists", payloadType)
	}

	routes[payloadType] = route
	return nil
}

type deployDatasetRoute struct{}

func (d *deployDatasetRoute) Execute(ctx TxContext, router *TxApp, tx *transactions.Transaction) *TxResponse {
	dbTx, err := router.currentTx.BeginTx(ctx.Ctx)
	if err != nil {
		return txRes(nil, transactions.CodeUnknownError, err)
	}

	spend, code, err := router.checkAndSpend(ctx, tx, d, dbTx)
	if err != nil {
		// if insufficient balance / spend amount, still commit the tx
		// otherwise, it is some internal database error, and should fail.
		switch code {
		case transactions.CodeOk, transactions.CodeInsufficientBalance, transactions.CodeInsufficientFee:
			logErr(router.log, dbTx.Commit(ctx.Ctx))
		default:
			logErr(router.log, dbTx.Rollback(ctx.Ctx))
		}
		return txRes(spend, code, err)
	}
	defer dbTx.Commit(ctx.Ctx)

	schemaPayload := &transactions.Schema{}
	err = schemaPayload.UnmarshalBinary(tx.Body.Payload)
	if err != nil {
		return txRes(spend, transactions.CodeEncodingError, err)
	}

	var schema *engineTypes.Schema
	schema, err = convertSchemaToEngine(schemaPayload)
	if err != nil {
		return txRes(spend, transactions.CodeUnknownError, err)
	}

	tx2, err := dbTx.BeginTx(ctx.Ctx)
	if err != nil {
		return txRes(spend, transactions.CodeUnknownError, err)
	}
	defer tx2.Rollback(ctx.Ctx)

	err = router.Engine.CreateDataset(ctx.Ctx, tx2, schema, tx.Sender)
	if err != nil {
		return txRes(spend, transactions.CodeUnknownError, err)
	}

	err = tx2.Commit(ctx.Ctx)
	if err != nil {
		return txRes(spend, transactions.CodeUnknownError, err)
	}

	return txRes(spend, transactions.CodeOk, nil)
}

func (d *deployDatasetRoute) Price(ctx context.Context, router *TxApp, tx *transactions.Transaction) (*big.Int, error) {
	return big.NewInt(1000000000000000000), nil
}

type dropDatasetRoute struct{}

func (d *dropDatasetRoute) Execute(ctx TxContext, router *TxApp, tx *transactions.Transaction) *TxResponse {
	dbTx, err := router.currentTx.BeginTx(ctx.Ctx)
	if err != nil {
		return txRes(nil, transactions.CodeUnknownError, err)
	}

	spend, code, err := router.checkAndSpend(ctx, tx, d, dbTx)
	if err != nil {
		switch code {
		case transactions.CodeOk, transactions.CodeInsufficientBalance, transactions.CodeInsufficientFee:
			logErr(router.log, dbTx.Commit(ctx.Ctx))
		default:
			logErr(router.log, dbTx.Rollback(ctx.Ctx))
		}
		return txRes(spend, code, err)
	}
	defer dbTx.Commit(ctx.Ctx)

	drop := &transactions.DropSchema{}
	err = drop.UnmarshalBinary(tx.Body.Payload)
	if err != nil {
		return txRes(spend, transactions.CodeEncodingError, err)
	}

	tx2, err := dbTx.BeginTx(ctx.Ctx)
	if err != nil {
		return txRes(spend, transactions.CodeUnknownError, err)
	}
	defer tx2.Rollback(ctx.Ctx)

	err = router.Engine.DeleteDataset(ctx.Ctx, tx2, drop.DBID, tx.Sender)
	if err != nil {
		return txRes(spend, transactions.CodeUnknownError, err)
	}

	err = tx2.Commit(ctx.Ctx)
	if err != nil {
		return txRes(spend, transactions.CodeUnknownError, err)
	}

	return txRes(spend, transactions.CodeOk, nil)
}

func (d *dropDatasetRoute) Price(ctx context.Context, router *TxApp, tx *transactions.Transaction) (*big.Int, error) {
	return big.NewInt(10000000000000), nil
}

type executeActionRoute struct{}

func (e *executeActionRoute) Execute(ctx TxContext, router *TxApp, tx *transactions.Transaction) *TxResponse {
	dbTx, err := router.currentTx.BeginTx(ctx.Ctx)
	if err != nil {
		return txRes(nil, transactions.CodeUnknownError, err)
	}

	spend, code, err := router.checkAndSpend(ctx, tx, e, dbTx)
	if err != nil {
		switch code {
		case transactions.CodeOk, transactions.CodeInsufficientBalance, transactions.CodeInsufficientFee:
			logErr(router.log, dbTx.Commit(ctx.Ctx))
		default:
			logErr(router.log, dbTx.Rollback(ctx.Ctx))
		}
		return txRes(spend, code, err)
	}
	defer dbTx.Commit(ctx.Ctx)

	action := &transactions.ActionExecution{}
	err = action.UnmarshalBinary(tx.Body.Payload)
	if err != nil {
		return txRes(spend, transactions.CodeEncodingError, err)
	}

	identifier, err := ident.Identifier(tx.Signature.Type, tx.Sender)
	if err != nil {
		return txRes(spend, transactions.CodeUnknownError, err)
	}

	args := make([][]any, len(action.Arguments))
	for i, arg := range action.Arguments {
		args[i] = make([]any, len(arg))

		for j, val := range arg {
			args[i][j] = val
		}
	}

	// we want to execute the tx for as many arg arrays exist
	// if there are no arg arrays, we want to execute it once
	if len(args) == 0 {
		args = make([][]any, 1)
	}

	tx2, err := dbTx.BeginTx(ctx.Ctx)
	if err != nil {
		return txRes(spend, transactions.CodeUnknownError, err)
	}
	defer tx2.Rollback(ctx.Ctx)

	for i := range action.Arguments {
		_, err = router.Engine.Execute(ctx.Ctx, tx2, &engineTypes.ExecutionData{
			Dataset:   action.DBID,
			Procedure: action.Action,
			Args:      args[i],
			Signer:    tx.Sender,
			Caller:    identifier,
		})
		if err != nil {
			return txRes(spend, transactions.CodeUnknownError, err)
		}
	}

	err = tx2.Commit(ctx.Ctx)
	if err != nil {
		return txRes(spend, transactions.CodeUnknownError, err)
	}

	return txRes(spend, transactions.CodeOk, nil)
}

func (e *executeActionRoute) Price(ctx context.Context, router *TxApp, tx *transactions.Transaction) (*big.Int, error) {
	return big.NewInt(2000000000000000), nil
}

type transferRoute struct{}

var bigZero = big.NewInt(0)

func (t *transferRoute) Execute(ctx TxContext, router *TxApp, tx *transactions.Transaction) *TxResponse {
	dbTx, err := router.currentTx.BeginTx(ctx.Ctx)
	if err != nil {
		return txRes(nil, transactions.CodeUnknownError, err)
	}

	spend, code, err := router.checkAndSpend(ctx, tx, t, dbTx)
	if err != nil {
		switch code {
		case transactions.CodeOk, transactions.CodeInsufficientBalance, transactions.CodeInsufficientFee:
			logErr(router.log, dbTx.Commit(ctx.Ctx))
		default:
			logErr(router.log, dbTx.Rollback(ctx.Ctx))
		}
		return txRes(spend, code, err)
	}
	defer dbTx.Commit(ctx.Ctx)

	transfer := &transactions.Transfer{}
	err = transfer.UnmarshalBinary(tx.Body.Payload)
	if err != nil {
		return txRes(spend, transactions.CodeEncodingError, err)
	}

	bigAmt, ok := new(big.Int).SetString(transfer.Amount, 10)
	if !ok {
		return txRes(spend, transactions.CodeInvalidAmount, fmt.Errorf("failed to parse amount: %s", transfer.Amount))
	}

	// Negative send amounts should be blocked at various levels, so we should
	// never get this, but be extra defensive since we cannot allow thievery.
	if bigAmt.Cmp(bigZero) < 0 {
		return txRes(spend, transactions.CodeInvalidAmount, fmt.Errorf("invalid transfer amount: %s", transfer.Amount))
	}

	tx2, err := dbTx.BeginTx(ctx.Ctx)
	if err != nil {
		return txRes(spend, transactions.CodeUnknownError, err)
	}
	defer tx2.Rollback(ctx.Ctx)

	err = router.Accounts.Transfer(ctx.Ctx, tx2, transfer.To, tx.Sender, bigAmt)
	if err != nil {
		return txRes(spend, transactions.CodeUnknownError, err)
	}

	err = tx2.Commit(ctx.Ctx)
	if err != nil {
		return txRes(spend, transactions.CodeUnknownError, err)
	}

	return txRes(spend, transactions.CodeOk, nil)
}

func (t *transferRoute) Price(ctx context.Context, router *TxApp, tx *transactions.Transaction) (*big.Int, error) {
	return big.NewInt(210_000), nil
}

type validatorJoinRoute struct{}

func (v *validatorJoinRoute) Execute(ctx TxContext, router *TxApp, tx *transactions.Transaction) *TxResponse {
	dbTx, err := router.currentTx.BeginTx(ctx.Ctx)
	if err != nil {
		return txRes(nil, transactions.CodeUnknownError, err)
	}

	spend, code, err := router.checkAndSpend(ctx, tx, v, dbTx)
	if err != nil {
		switch code {
		case transactions.CodeOk, transactions.CodeInsufficientBalance, transactions.CodeInsufficientFee:
			logErr(router.log, dbTx.Commit(ctx.Ctx))
		default:
			logErr(router.log, dbTx.Rollback(ctx.Ctx))
		}
		return txRes(spend, code, err)
	}
	defer dbTx.Commit(ctx.Ctx)

	join := &transactions.ValidatorJoin{}
	err = join.UnmarshalBinary(tx.Body.Payload)
	if err != nil {
		return txRes(spend, transactions.CodeEncodingError, err)
	}

	tx2, err := dbTx.BeginTx(ctx.Ctx)
	if err != nil {
		return txRes(spend, transactions.CodeUnknownError, err)
	}
	defer tx2.Rollback(ctx.Ctx)

	err = router.Validators.Join(ctx.Ctx, tx2, tx.Sender, int64(join.Power))
	if err != nil {
		return txRes(spend, transactions.CodeUnknownError, err)
	}

	err = tx2.Commit(ctx.Ctx)
	if err != nil {
		return txRes(spend, transactions.CodeUnknownError, err)
	}

	return txRes(spend, transactions.CodeOk, nil)
}

func (v *validatorJoinRoute) Price(ctx context.Context, router *TxApp, tx *transactions.Transaction) (*big.Int, error) {
	return big.NewInt(10000000000000), nil
}

type validatorApproveRoute struct{}

func (v *validatorApproveRoute) Execute(ctx TxContext, router *TxApp, tx *transactions.Transaction) *TxResponse {
	dbTx, err := router.currentTx.BeginTx(ctx.Ctx)
	if err != nil {
		return txRes(nil, transactions.CodeUnknownError, err)
	}

	spend, code, err := router.checkAndSpend(ctx, tx, v, dbTx)
	if err != nil {
		switch code {
		case transactions.CodeOk, transactions.CodeInsufficientBalance, transactions.CodeInsufficientFee:
			logErr(router.log, dbTx.Commit(ctx.Ctx))
		default:
			logErr(router.log, dbTx.Rollback(ctx.Ctx))
		}
		return txRes(spend, code, err)
	}
	defer dbTx.Commit(ctx.Ctx)

	approve := &transactions.ValidatorApprove{}
	err = approve.UnmarshalBinary(tx.Body.Payload)
	if err != nil {
		return txRes(spend, transactions.CodeEncodingError, err)
	}

	tx2, err := dbTx.BeginTx(ctx.Ctx)
	if err != nil {
		return txRes(spend, transactions.CodeUnknownError, err)
	}
	defer tx2.Rollback(ctx.Ctx)

	err = router.Validators.Approve(ctx.Ctx, tx2, approve.Candidate, tx.Sender)
	if err != nil {
		return txRes(spend, transactions.CodeUnknownError, err)
	}

	err = tx2.Commit(ctx.Ctx)
	if err != nil {
		return txRes(spend, transactions.CodeUnknownError, err)
	}

	return txRes(spend, transactions.CodeOk, nil)
}

func (v *validatorApproveRoute) Price(ctx context.Context, router *TxApp, tx *transactions.Transaction) (*big.Int, error) {
	return big.NewInt(10000000000000), nil
}

type validatorRemoveRoute struct{}

func (v *validatorRemoveRoute) Execute(ctx TxContext, router *TxApp, tx *transactions.Transaction) *TxResponse {
	dbTx, err := router.currentTx.BeginTx(ctx.Ctx)
	if err != nil {
		return txRes(nil, transactions.CodeUnknownError, err)
	}

	spend, code, err := router.checkAndSpend(ctx, tx, v, dbTx)
	if err != nil {
		switch code {
		case transactions.CodeOk, transactions.CodeInsufficientBalance, transactions.CodeInsufficientFee:
			logErr(router.log, dbTx.Commit(ctx.Ctx))
		default:
			logErr(router.log, dbTx.Rollback(ctx.Ctx))
		}
		return txRes(spend, code, err)
	}
	defer dbTx.Commit(ctx.Ctx)

	remove := &transactions.ValidatorRemove{}
	err = remove.UnmarshalBinary(tx.Body.Payload)
	if err != nil {
		return txRes(spend, transactions.CodeEncodingError, err)
	}

	tx2, err := dbTx.BeginTx(ctx.Ctx)
	if err != nil {
		return txRes(spend, transactions.CodeUnknownError, err)
	}
	defer tx2.Rollback(ctx.Ctx)

	err = router.Validators.Remove(ctx.Ctx, tx2, remove.Validator, tx.Sender)
	if err != nil {
		return txRes(spend, transactions.CodeUnknownError, err)
	}

	err = tx2.Commit(ctx.Ctx)
	if err != nil {
		return txRes(spend, transactions.CodeUnknownError, err)
	}

	return txRes(spend, transactions.CodeOk, nil)
}

func (v *validatorRemoveRoute) Price(ctx context.Context, router *TxApp, tx *transactions.Transaction) (*big.Int, error) {
	return bigZero, nil
}

type validatorLeaveRoute struct{}

func (v *validatorLeaveRoute) Execute(ctx TxContext, router *TxApp, tx *transactions.Transaction) *TxResponse {
	dbTx, err := router.currentTx.BeginTx(ctx.Ctx)
	if err != nil {
		return txRes(nil, transactions.CodeUnknownError, err)
	}

	spend, code, err := router.checkAndSpend(ctx, tx, v, dbTx)
	if err != nil {
		switch code {
		case transactions.CodeOk, transactions.CodeInsufficientBalance, transactions.CodeInsufficientFee:
			logErr(router.log, dbTx.Commit(ctx.Ctx))
		default:
			logErr(router.log, dbTx.Rollback(ctx.Ctx))
		}
		return txRes(spend, code, err)
	}
	defer dbTx.Commit(ctx.Ctx)

	// doing this b/c the old version did, but it seems there is no reason to do this
	leave := &transactions.ValidatorLeave{}
	err = leave.UnmarshalBinary(tx.Body.Payload)
	if err != nil {
		return txRes(spend, transactions.CodeEncodingError, err)
	}

	tx2, err := dbTx.BeginTx(ctx.Ctx)
	if err != nil {
		return txRes(spend, transactions.CodeUnknownError, err)
	}
	defer tx2.Rollback(ctx.Ctx)

	err = router.Validators.Leave(ctx.Ctx, tx2, tx.Sender)
	if err != nil {
		return txRes(spend, transactions.CodeUnknownError, err)
	}

	err = tx2.Commit(ctx.Ctx)
	if err != nil {
		return txRes(spend, transactions.CodeUnknownError, err)
	}

	return txRes(spend, transactions.CodeOk, nil)
}

func (v *validatorLeaveRoute) Price(ctx context.Context, router *TxApp, tx *transactions.Transaction) (*big.Int, error) {
	return big.NewInt(10000000000000), nil
}

// validatorVoteIDsRoute is a route for approving a set of votes based on their IDs.
type validatorVoteIDsRoute struct{}

// Execute will approve the votes for the given IDs.
// If the event already has a body in the event store, and the vote
// is from the local validator, the event will be deleted from the event store.
func (v *validatorVoteIDsRoute) Execute(ctx TxContext, router *TxApp, tx *transactions.Transaction) *TxResponse {
	dbTx, err := router.currentTx.BeginTx(ctx.Ctx)
	if err != nil {
		return txRes(nil, transactions.CodeUnknownError, err)
	}

	spend, code, err := router.checkAndSpend(ctx, tx, v, dbTx)
	if err != nil {
		switch code {
		case transactions.CodeOk, transactions.CodeInsufficientBalance, transactions.CodeInsufficientFee:
			logErr(router.log, dbTx.Commit(ctx.Ctx))
		default:
			logErr(router.log, dbTx.Rollback(ctx.Ctx))
		}
		return txRes(spend, code, err)
	}
	defer dbTx.Commit(ctx.Ctx)

	tx2, err := dbTx.BeginTx(ctx.Ctx)
	if err != nil {
		return txRes(spend, transactions.CodeUnknownError, err)
	}
	defer tx2.Rollback(ctx.Ctx)

	isValidator, err := router.Validators.IsCurrent(ctx.Ctx, tx2, tx.Sender)
	if err != nil {
		return txRes(spend, transactions.CodeUnknownError, err)
	}
	if !isValidator {
		return txRes(spend, transactions.CodeInvalidSender, ErrCallerNotValidator)
	}

	approve := &transactions.ValidatorVoteIDs{}
	err = approve.UnmarshalBinary(tx.Body.Payload)
	if err != nil {
		return txRes(spend, transactions.CodeEncodingError, err)
	}

	fromLocalValidator := bytes.Equal(tx.Sender, router.signer.Identity())
	expiryHeight := ctx.BlockHeight + ctx.VotingPeriod

	for _, voteID := range approve.ResolutionIDs {
		err = router.VoteStore.Approve(ctx.Ctx, tx2, voteID, expiryHeight, tx.Sender)
		if err != nil {
			return txRes(spend, transactions.CodeUnknownError, err)
		}

		// if from local validator, we should mark that it is committed,
		// so that we do not rebroadcast. We do not want to delete,
		// since we may be the proposer later, and will need the body
		// If the network already has the body, then we can just delete.
		if fromLocalValidator {
			containsBody, err := router.VoteStore.ContainsBodyOrFinished(ctx.Ctx, tx2, voteID) // should be uncommitted queries internally?
			if err != nil {
				return txRes(spend, transactions.CodeUnknownError, err)
			}

			if containsBody {
				err = deleteEvent(ctx.Ctx, tx2, voteID)
				if err != nil {
					return txRes(spend, transactions.CodeUnknownError, err)
				}
			} else {
				err = markReceived(ctx.Ctx, tx2, voteID)
				if err != nil {
					return txRes(spend, transactions.CodeUnknownError, err)
				}
			}
		}
	}

	err = tx2.Commit(ctx.Ctx)
	if err != nil {
		return txRes(spend, transactions.CodeUnknownError, err)
	}

	return txRes(spend, transactions.CodeOk, nil)
}

func (v *validatorVoteIDsRoute) Price(ctx context.Context, router *TxApp, tx *transactions.Transaction) (*big.Int, error) {
	// Check if gas costs are enabled

	// VoteID pricing is based on the number of vote IDs.
	ids := &transactions.ValidatorVoteIDs{}
	err := ids.UnmarshalBinary(tx.Body.Payload)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal vote IDs: %w", err)
	}

	return big.NewInt(int64(len(ids.ResolutionIDs)) * voting.ValidatorVoteIDPrice), nil
}

// validatorVoteBodiesRoute is a route for handling votes for a set of vote bodies.
type validatorVoteBodiesRoute struct{}

// Execute will add the event bodies to the event store.
// For each event, if the local validator has already voted on the event,
// the event will be deleted from the event store.
func (v *validatorVoteBodiesRoute) Execute(ctx TxContext, router *TxApp, tx *transactions.Transaction) *TxResponse {
	dbTx, err := router.currentTx.BeginTx(ctx.Ctx)
	if err != nil {
		return txRes(nil, transactions.CodeUnknownError, err)
	}

	spend, code, err := router.checkAndSpend(ctx, tx, v, dbTx)
	if err != nil {
		switch code {
		case transactions.CodeOk, transactions.CodeInsufficientBalance, transactions.CodeInsufficientFee:
			logErr(router.log, dbTx.Commit(ctx.Ctx))
		default:
			logErr(router.log, dbTx.Rollback(ctx.Ctx))
		}
		return txRes(spend, code, err)
	}
	defer dbTx.Commit(ctx.Ctx)

	if !bytes.Equal(tx.Sender, ctx.Proposer) {
		return txRes(spend, transactions.CodeInvalidSender, ErrCallerNotProposer)
	}

	vote := &transactions.ValidatorVoteBodies{}
	err = vote.UnmarshalBinary(tx.Body.Payload)
	if err != nil {
		return txRes(spend, transactions.CodeEncodingError, err)
	}

	localValidator := router.signer.Identity()
	expiryHeight := ctx.BlockHeight + ctx.VotingPeriod

	tx2, err := dbTx.BeginTx(ctx.Ctx)
	if err != nil {
		return txRes(spend, transactions.CodeUnknownError, err)
	}
	defer tx2.Rollback(ctx.Ctx)

	for _, event := range vote.Events {
		err = router.VoteStore.CreateResolution(ctx.Ctx, tx2, event, expiryHeight, tx.Sender)
		if err != nil {
			return txRes(spend, transactions.CodeUnknownError, err)
		}

		// since the vote body proposer is implicitly voting for the event,
		// we need to approve the newly created vote body here
		err = router.VoteStore.Approve(ctx.Ctx, tx2, event.ID(), expiryHeight, tx.Sender)
		if err != nil {
			return txRes(spend, transactions.CodeUnknownError, err)
		}

		// If the local validator has already voted on the event, then we should delete the event.
		hasVoted, err := router.VoteStore.HasVoted(ctx.Ctx, tx2, event.ID(), localValidator)
		if err != nil {
			return txRes(spend, transactions.CodeUnknownError, err)
		}
		if hasVoted {
			err = deleteEvent(ctx.Ctx, tx2, event.ID())
			if err != nil {
				return txRes(spend, transactions.CodeUnknownError, err)
			}
		}
	}

	if err = tx2.Commit(ctx.Ctx); err != nil {
		return txRes(spend, transactions.CodeUnknownError, err)
	}

	return txRes(spend, transactions.CodeOk, nil)
}
func (v *validatorVoteBodiesRoute) Price(ctx context.Context, router *TxApp, tx *transactions.Transaction) (*big.Int, error) {
	// Check if gas costs are enabled

	// VoteBody pricing is based on the size of the vote bodies of all the events in the tx payload.
	votes := &transactions.ValidatorVoteBodies{}
	err := votes.UnmarshalBinary(tx.Body.Payload)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal vote bodies: %w", err)
	}

	var totalSize int64

	for _, event := range votes.Events {
		totalSize += int64(len(event.Body))
	}

	return big.NewInt(totalSize * voting.ValidatorVoteBodyBytePrice), nil
}
