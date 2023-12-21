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
	Execute(ctx TxContext, router *TxApp, tx *transactions.Transaction) *TxResponse
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
	spend, code, err := router.checkAndSpend(ctx, tx)
	if err != nil {
		return txRes(spend, code, err)
	}

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

	err = router.Database.CreateDataset(ctx.Ctx(), schema, tx.Sender)
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
	spend, code, err := router.checkAndSpend(ctx, tx)
	if err != nil {
		return txRes(spend, code, err)
	}

	drop := &transactions.DropSchema{}
	err = drop.UnmarshalBinary(tx.Body.Payload)
	if err != nil {
		return txRes(spend, transactions.CodeEncodingError, err)
	}

	err = router.Database.DeleteDataset(ctx.Ctx(), drop.DBID, tx.Sender)
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
	spend, code, err := router.checkAndSpend(ctx, tx)
	if err != nil {
		return txRes(spend, code, err)
	}

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

	for i := range action.Arguments {
		_, err = router.Database.Execute(ctx.Ctx(), &engineTypes.ExecutionData{
			Dataset:   action.DBID,
			Procedure: action.Action,
			Mutative:  true, // transaction execution is always mutative
			Args:      args[i],
			Signer:    tx.Sender,
			Caller:    identifier,
		})
		if err != nil {
			return txRes(spend, transactions.CodeUnknownError, err)
		}
	}

	return txRes(spend, transactions.CodeOk, nil)
}

func (e *executeActionRoute) Price(ctx context.Context, router *TxApp, tx *transactions.Transaction) (*big.Int, error) {
	return big.NewInt(2000000000000000), nil
}

type transferRoute struct{}

var bigZero = big.NewInt(0)

func (t *transferRoute) Execute(ctx TxContext, router *TxApp, tx *transactions.Transaction) *TxResponse {
	spend, code, err := router.checkAndSpend(ctx, tx)
	if err != nil {
		return txRes(spend, code, err)
	}

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

	err = router.Accounts.Transfer(ctx.Ctx(), transfer.To, tx.Sender, bigAmt)
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
	spend, code, err := router.checkAndSpend(ctx, tx)
	if err != nil {
		return txRes(spend, code, err)
	}

	join := &transactions.ValidatorJoin{}
	err = join.UnmarshalBinary(tx.Body.Payload)
	if err != nil {
		return txRes(spend, transactions.CodeEncodingError, err)
	}

	err = router.Validators.Join(ctx.Ctx(), tx.Sender, int64(join.Power))
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
	spend, code, err := router.checkAndSpend(ctx, tx)
	if err != nil {
		return txRes(spend, code, err)
	}

	approve := &transactions.ValidatorApprove{}
	err = approve.UnmarshalBinary(tx.Body.Payload)
	if err != nil {
		return txRes(spend, transactions.CodeEncodingError, err)
	}

	err = router.Validators.Approve(ctx.Ctx(), approve.Candidate, tx.Sender)
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
	spend, code, err := router.checkAndSpend(ctx, tx)
	if err != nil {
		return txRes(spend, code, err)
	}

	remove := &transactions.ValidatorRemove{}
	err = remove.UnmarshalBinary(tx.Body.Payload)
	if err != nil {
		return txRes(spend, transactions.CodeEncodingError, err)
	}

	err = router.Validators.Remove(ctx.Ctx(), remove.Validator, tx.Sender)
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
	spend, code, err := router.checkAndSpend(ctx, tx)
	if err != nil {
		return txRes(spend, code, err)
	}

	// doing this b/c the old version did, but it seems there is no reason to do this
	leave := &transactions.ValidatorLeave{}
	err = leave.UnmarshalBinary(tx.Body.Payload)
	if err != nil {
		return txRes(spend, transactions.CodeEncodingError, err)
	}

	err = router.Validators.Leave(ctx.Ctx(), tx.Sender)
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
	spend, code, err := router.checkAndSpend(ctx, tx)
	if err != nil {
		return txRes(spend, code, err)
	}

	isValidator, err := router.Validators.IsCurrent(ctx.Ctx(), tx.Sender)
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

	isLocalValidator := bytes.Equal(tx.Sender, router.LocalValidator.Signer().Identity())
	expiryHeight := ctx.BlockHeight() + ctx.ConsensusParams().VotingPeriod

	for _, voteID := range approve.ResolutionIDs {
		err = router.VoteStore.Approve(ctx.Ctx(), voteID, expiryHeight, tx.Sender)
		if err != nil {
			return txRes(spend, transactions.CodeUnknownError, err)
		}

		containsBody, err := router.VoteStore.ContainsBodyOrFinished(ctx.Ctx(), voteID)
		if err != nil {
			return txRes(spend, transactions.CodeUnknownError, err)
		}

		if isLocalValidator && containsBody {
			err = router.EventStore.DeleteEvent(ctx.Ctx(), voteID)
			if err != nil {
				return txRes(spend, transactions.CodeUnknownError, err)
			}
		}
	}

	return txRes(spend, transactions.CodeOk, nil)
}

func (v *validatorVoteIDsRoute) Price(ctx context.Context, router *TxApp, tx *transactions.Transaction) (*big.Int, error) {
	return bigZero, nil
}

// validatorVoteBodiesRoute is a route for handling votes for a set of vote bodies.
type validatorVoteBodiesRoute struct{}

// Execute will add the event bodies to the event store.
// For each event, if the local validator has already voted on the event,
// the event will be deleted from the event store.
func (v *validatorVoteBodiesRoute) Execute(ctx TxContext, router *TxApp, tx *transactions.Transaction) *TxResponse {
	spend, code, err := router.checkAndSpend(ctx, tx)
	if err != nil {
		return txRes(spend, code, err)
	}

	if !bytes.Equal(tx.Sender, ctx.Proposer()) {
		return txRes(spend, transactions.CodeInvalidSender, ErrCallerNotProposer)
	}

	vote := &transactions.ValidatorVoteBodies{}
	err = vote.UnmarshalBinary(tx.Body.Payload)
	if err != nil {
		return txRes(spend, transactions.CodeEncodingError, err)
	}

	localValidator := router.LocalValidator.Signer().Identity()
	expiryHeight := ctx.BlockHeight() + ctx.ConsensusParams().VotingPeriod

	for _, event := range vote.Events {
		err = router.VoteStore.CreateResolution(ctx.Ctx(), event, expiryHeight)
		if err != nil {
			return txRes(spend, transactions.CodeUnknownError, err)
		}

		// since the vote body proposer is implicitly voting for the event,
		// we need to approve the newly created vote body here
		err = router.VoteStore.Approve(ctx.Ctx(), event.ID(), expiryHeight, tx.Sender)
		if err != nil {
			return txRes(spend, transactions.CodeUnknownError, err)
		}

		hasVoted, err := router.VoteStore.HasVoted(ctx.Ctx(), event.ID(), localValidator)
		if err != nil {
			return txRes(spend, transactions.CodeUnknownError, err)
		}
		if hasVoted {
			err = router.EventStore.DeleteEvent(ctx.Ctx(), event.ID())
			if err != nil {
				return txRes(spend, transactions.CodeUnknownError, err)
			}
		}
	}

	return txRes(spend, transactions.CodeOk, nil)
}

func (v *validatorVoteBodiesRoute) Price(ctx context.Context, router *TxApp, tx *transactions.Transaction) (*big.Int, error) {
	return bigZero, nil
}
