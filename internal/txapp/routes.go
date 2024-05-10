package txapp

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"

	"github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/types/transactions"
	"github.com/kwilteam/kwil-db/extensions/resolutions"
	"github.com/kwilteam/kwil-db/internal/accounts"
	"github.com/kwilteam/kwil-db/internal/engine/execution"
	"github.com/kwilteam/kwil-db/internal/ident"
	"github.com/kwilteam/kwil-db/internal/voting"
)

func init() {
	err := errors.Join(
		registerRoute(transactions.PayloadTypeDeploySchema.String(), &deployDatasetRoute{}),
		registerRoute(transactions.PayloadTypeDropSchema.String(), &dropDatasetRoute{}),
		registerRoute(transactions.PayloadTypeExecute.String(), &executeActionRoute{}),
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
	// ConsensusParams holds network level parameters that can be evolved
	// over the lifetime of a network.
	ConsensusParams ConsensusParams
	// TxID is the ID of the current transaction.
	// It is defined by CometBFT.
	TxID []byte
}

// ConsensusParams holds network level parameters that may evolve over time.
type ConsensusParams struct {
	// VotingPeriod is the maximum length of a voting period.
	// It is measured in blocks, and is applied additively.
	// e.g. if the current block is 50, and VotingPeriod is 100,
	// then the current voting period ends at block 150.
	VotingPeriod int64
	// JoinVoteExpiration is the voting period for any validator
	// join or removal vote. It is measured in blocks, and is applied additively.
	// e.g. if the current block is 50, and JoinVoteExpiration is 100,
	// then the current voting period ends at block 150.
	JoinVoteExpiration int64
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

func codeForEngineError(err error) transactions.TxCode {
	if err == nil {
		return transactions.CodeOk
	}
	if errors.Is(err, execution.ErrDatasetExists) {
		return transactions.CodeDatasetExists
	}
	if errors.Is(err, execution.ErrDatasetNotFound) {
		return transactions.CodeDatasetMissing
	}
	if errors.Is(err, execution.ErrInvalidSchema) {
		return transactions.CodeInvalidSchema
	}

	return transactions.CodeUnknownError
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

	schema, err := schemaPayload.ToTypes()
	if err != nil {
		return txRes(spend, transactions.CodeUnknownError, err)
	}

	identifier, err := ident.Identifier(tx.Signature.Type, tx.Sender)
	if err != nil {
		return txRes(spend, transactions.CodeUnknownError, err)
	}

	tx2, err := dbTx.BeginTx(ctx.Ctx)
	if err != nil {
		return txRes(spend, transactions.CodeUnknownError, err)
	}
	defer tx2.Rollback(ctx.Ctx)

	err = router.Engine.CreateDataset(ctx.Ctx, tx2, schema, &common.TransactionData{
		Signer: tx.Sender,
		Caller: identifier,
		TxID:   hex.EncodeToString(ctx.TxID),
	})
	if err != nil {
		return txRes(spend, codeForEngineError(err), err)
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

	identifier, err := ident.Identifier(tx.Signature.Type, tx.Sender)
	if err != nil {
		return txRes(spend, transactions.CodeUnknownError, err)
	}

	tx2, err := dbTx.BeginTx(ctx.Ctx)
	if err != nil {
		return txRes(spend, transactions.CodeUnknownError, err)
	}
	defer tx2.Rollback(ctx.Ctx)

	err = router.Engine.DeleteDataset(ctx.Ctx, tx2, drop.DBID, &common.TransactionData{
		Signer: tx.Sender,
		Caller: identifier,
		TxID:   hex.EncodeToString(ctx.TxID),
	})
	if err != nil {
		return txRes(spend, codeForEngineError(err), err)
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

	if len(action.NilArg) > len(action.Arguments) {
		return txRes(nil, transactions.CodeEncodingError, errors.New("too many NilArgs"))
	}

	// [][]string => [][]any
	args := make([][]any, len(action.Arguments))
	for i, arg := range action.Arguments {
		args[i] = make([]any, len(arg))
		for j, val := range arg {
			args[i][j] = val
		}
	}

	// Set args to nil as specified.
	for i, nils := range action.NilArg {
		if len(nils) > len(args[i]) {
			return txRes(nil, transactions.CodeEncodingError, errors.New("too many NilArgs"))
		}
		for j, isNil := range nils {
			if isNil {
				args[i][j] = nil
			}
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

	for i := range args {
		_, err = router.Engine.Procedure(ctx.Ctx, tx2, &common.ExecutionData{
			Dataset:   action.DBID,
			Procedure: action.Action,
			Args:      args[i],
			TransactionData: common.TransactionData{
				Signer: tx.Sender,
				Caller: identifier,
				TxID:   hex.EncodeToString(ctx.TxID),
			},
		})
		if err != nil {
			return txRes(spend, codeForEngineError(err), err)
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

	transferBody := &transactions.Transfer{}
	err = transferBody.UnmarshalBinary(tx.Body.Payload)
	if err != nil {
		return txRes(spend, transactions.CodeEncodingError, err)
	}

	bigAmt, ok := new(big.Int).SetString(transferBody.Amount, 10)
	if !ok {
		return txRes(spend, transactions.CodeInvalidAmount, fmt.Errorf("failed to parse amount: %s", transferBody.Amount))
	}

	// Negative send amounts should be blocked at various levels, so we should
	// never get this, but be extra defensive since we cannot allow thievery.
	if bigAmt.Sign() < 0 {
		return txRes(spend, transactions.CodeInvalidAmount, fmt.Errorf("invalid transfer amount: %s", transferBody.Amount))
	}

	tx2, err := dbTx.BeginTx(ctx.Ctx)
	if err != nil {
		return txRes(spend, transactions.CodeUnknownError, err)
	}
	defer tx2.Rollback(ctx.Ctx)

	err = transfer(ctx.Ctx, tx2, tx.Sender, transferBody.To, bigAmt)
	if err != nil {
		if errors.Is(err, accounts.ErrInsufficientFunds) {
			return txRes(spend, transactions.CodeInsufficientBalance, err)
		}
		if errors.Is(err, accounts.ErrNegativeBalance) {
			return txRes(spend, transactions.CodeInvalidAmount, err)
		}
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

	// ensure this candidate is not already a validator
	power, err := getVoterPower(ctx.Ctx, tx2, tx.Sender)
	if err != nil {
		return txRes(spend, transactions.CodeUnknownError, err)
	}
	if power > 0 {
		return txRes(spend, transactions.CodeInvalidSender, ErrCallerIsValidator)
	}

	// we first need to ensure that this validator does not have a pending join request
	// if it does, we should not allow it to join again
	pending, err := getResolutionsByTypeAndProposer(ctx.Ctx, tx2, voting.ValidatorJoinEventType, tx.Sender)
	if err != nil {
		return txRes(spend, transactions.CodeUnknownError, err)
	}
	if len(pending) > 0 {
		return txRes(spend, transactions.CodeInvalidSender, fmt.Errorf("validator already has a pending join request"))
	}

	// there are no pending join requests, so we can create a new one
	joinReq := &voting.UpdatePowerRequest{
		PubKey: tx.Sender,
		Power:  int64(join.Power),
	}
	bts, err := joinReq.MarshalBinary()
	if err != nil {
		return txRes(spend, transactions.CodeUnknownError, err)
	}

	event := &types.VotableEvent{
		Body: bts,
		Type: voting.ValidatorJoinEventType,
	}

	err = createResolution(ctx.Ctx, tx2, event, ctx.BlockHeight+ctx.ConsensusParams.JoinVoteExpiration, tx.Sender)
	if err != nil {
		return txRes(spend, transactions.CodeUnknownError, err)
	}
	// we do not approve, because a joiner is presumably not a voter

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

	if bytes.Equal(approve.Candidate, tx.Sender) {
		return txRes(spend, transactions.CodeInvalidSender, errors.New("cannot approve own join request"))
	}

	// each pending validator can only have one active join request at a time
	// we need to retrieve the join request and ensure that it is still pending
	pending, err := getResolutionsByTypeAndProposer(ctx.Ctx, tx2, voting.ValidatorJoinEventType, approve.Candidate)
	if err != nil {
		return txRes(spend, transactions.CodeUnknownError, err)
	}
	if len(pending) == 0 {
		return txRes(spend, transactions.CodeInvalidSender, fmt.Errorf("validator does not have a pending join request"))
	}
	if len(pending) > 1 {
		// this should never happen, but if it does, we should not allow it
		return txRes(spend, transactions.CodeUnknownError, fmt.Errorf("validator has more than one pending join request. this is an internal bug"))
	}

	// ensure that sender is a validator
	power, err := getVoterPower(ctx.Ctx, tx2, tx.Sender)
	if err != nil {
		return txRes(spend, transactions.CodeUnknownError, err)
	}
	if power <= 0 {
		return txRes(spend, transactions.CodeInvalidSender, ErrCallerNotValidator)
	}

	err = approveResolution(ctx.Ctx, tx2, pending[0], tx.Sender)
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

	removeReq := &voting.UpdatePowerRequest{
		PubKey: remove.Validator,
		Power:  0,
	}
	bts, err := removeReq.MarshalBinary()
	if err != nil {
		return txRes(spend, transactions.CodeUnknownError, err)
	}

	event := &types.VotableEvent{
		Body: bts,
		Type: voting.ValidatorRemoveEventType,
	}

	// ensure the sender is a validator
	power, err := getVoterPower(ctx.Ctx, tx2, tx.Sender)
	if err != nil {
		return txRes(spend, transactions.CodeUnknownError, err)
	}
	if power <= 0 {
		return txRes(spend, transactions.CodeInvalidSender, ErrCallerNotValidator)
	}

	// we should try to create the resolution, since validator removals are never
	// officially "started" by the user. If it fails because it already exists,
	// then we should do nothing

	err = createResolution(ctx.Ctx, tx2, event, ctx.BlockHeight+ctx.ConsensusParams.JoinVoteExpiration, tx.Sender)
	if errors.Is(err, voting.ErrResolutionAlreadyHasBody) {
		router.log.Debug("validator removal resolution already exists")
	} else if err != nil {
		return txRes(spend, transactions.CodeUnknownError, err)
	}

	// we need to approve the resolution as well
	err = approveResolution(ctx.Ctx, tx2, event.ID(), tx.Sender)
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
	return big.NewInt(100_000), nil
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

	// don't touch the DB or start another RW tx is sender isn't a validator
	power, err := getVoterPower(ctx.Ctx, dbTx, tx.Sender)
	if err != nil {
		return txRes(spend, transactions.CodeUnknownError, err)
	}
	if power <= 0 {
		return txRes(spend, transactions.CodeInvalidSender, ErrCallerNotValidator)
	}

	tx2, err := dbTx.BeginTx(ctx.Ctx)
	if err != nil {
		return txRes(spend, transactions.CodeUnknownError, err)
	}
	defer tx2.Rollback(ctx.Ctx)

	err = setVoterPower(ctx.Ctx, tx2, tx.Sender, 0)
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

	// if the caller has 0 power, they are not a validator, and should not be able to vote
	power, err := getVoterPower(ctx.Ctx, tx2, tx.Sender)
	if err != nil {
		return txRes(spend, transactions.CodeUnknownError, err)
	}
	if power == 0 {
		return txRes(spend, transactions.CodeInvalidSender, ErrCallerNotValidator)
	}

	approve := &transactions.ValidatorVoteIDs{}
	err = approve.UnmarshalBinary(tx.Body.Payload)
	if err != nil {
		return txRes(spend, transactions.CodeEncodingError, err)
	}

	// filter out the vote IDs that have already been processed
	ids, err := voting.FilterNotProcessed(ctx.Ctx, tx2, approve.ResolutionIDs)
	if err != nil {
		return txRes(spend, transactions.CodeUnknownError, err)
	}

	fromLocalValidator := bytes.Equal(tx.Sender, router.signer.Identity())

	for _, voteID := range ids {
		err = approveResolution(ctx.Ctx, tx2, voteID, tx.Sender)
		if err != nil {
			return txRes(spend, transactions.CodeUnknownError, err)
		}

		// if from local validator, delete the event now that we have voted on it and network already has the event body
		if fromLocalValidator {
			err = deleteEvent(ctx.Ctx, tx2, voteID)
			if err != nil {
				return txRes(spend, transactions.CodeUnknownError, err)
			}
		}
	}

	if tooLate := len(approve.ResolutionIDs) - len(ids); tooLate > 0 {
		// TODO: probably refund these votes??
		router.Log().Warnf("vote contains %d resolution IDs that are already done. too late, no refund!", tooLate)
	}

	err = tx2.Commit(ctx.Ctx)
	if err != nil {
		return txRes(spend, transactions.CodeUnknownError, err)
	}

	return txRes(spend, transactions.CodeOk, nil)
}

func (v *validatorVoteIDsRoute) Price(ctx context.Context, router *TxApp, tx *transactions.Transaction) (*big.Int, error) {
	// VoteID pricing is based on the number of vote IDs.
	ids := &transactions.ValidatorVoteIDs{}
	err := ids.UnmarshalBinary(tx.Body.Payload)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal vote IDs: %w", err)
	}
	return big.NewInt(int64(len(ids.ResolutionIDs)) * ValidatorVoteIDPrice.Int64()), nil
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

	// Only proposer can issue a VoteBody transaction.
	if !bytes.Equal(tx.Sender, ctx.Proposer) {
		return txRes(spend, transactions.CodeInvalidSender, ErrCallerNotProposer)
	}

	vote := &transactions.ValidatorVoteBodies{}
	err = vote.UnmarshalBinary(tx.Body.Payload)
	if err != nil {
		return txRes(spend, transactions.CodeEncodingError, err)
	}

	fromLocalValidator := bytes.Equal(tx.Sender, router.signer.Identity())

	tx2, err := dbTx.BeginTx(ctx.Ctx)
	if err != nil {
		return txRes(spend, transactions.CodeUnknownError, err)
	}
	defer tx2.Rollback(ctx.Ctx)

	// Expectation:
	// 1. VoteBody should only include the events for which the resolutions are not yet created. Maybe filter out the events for which the resolutions are already created and ignore them.
	// 2. If the node is the proposer, delete the event from the event store
	for _, event := range vote.Events {
		ev := &types.VotableEvent{
			Type: event.Type,
			Body: event.Body,
		}

		resCfg, err := resolutions.GetResolution(event.Type)
		if err != nil {
			return txRes(spend, transactions.CodeUnknownError, err)
		}

		expiryHeight := ctx.BlockHeight + resCfg.ExpirationPeriod
		err = createResolution(ctx.Ctx, tx2, ev, expiryHeight, tx.Sender)
		if err != nil {
			return txRes(spend, transactions.CodeUnknownError, err)
		}

		// since the vote body proposer is implicitly voting for the event,
		// we need to approve the newly created vote body here
		err = approveResolution(ctx.Ctx, tx2, ev.ID(), tx.Sender)
		if err != nil {
			return txRes(spend, transactions.CodeUnknownError, err)
		}

		// If the local validator is the proposer, then we should delete the event from the event store.
		if fromLocalValidator {
			err = deleteEvent(ctx.Ctx, tx2, ev.ID())
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

	return big.NewInt(totalSize * ValidatorVoteBodyBytePrice), nil
}
