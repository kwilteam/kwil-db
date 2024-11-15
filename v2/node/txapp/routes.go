package txapp

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"kwil/extensions/consensus"
	"kwil/extensions/resolutions"
	"kwil/node/accounts"
	"kwil/node/types/sql"
	"kwil/node/voting"
	"kwil/types"
	"math/big"
	"sync"
)

func init() {
	err := errors.Join(
		// RegisterRoute(types.PayloadTypeDeploySchema, NewRoute(&deployDatasetRoute{})),
		// RegisterRoute(types.PayloadTypeDropSchema, NewRoute(&dropDatasetRoute{})),
		// RegisterRoute(types.PayloadTypeExecute, NewRoute(&executeActionRoute{})),
		RegisterRoute(types.PayloadTypeTransfer, NewRoute(&transferRoute{})),
		RegisterRoute(types.PayloadTypeValidatorJoin, NewRoute(&validatorJoinRoute{})),
		RegisterRoute(types.PayloadTypeValidatorApprove, NewRoute(&validatorApproveRoute{})),
		RegisterRoute(types.PayloadTypeValidatorRemove, NewRoute(&validatorRemoveRoute{})),
		RegisterRoute(types.PayloadTypeValidatorLeave, NewRoute(&validatorLeaveRoute{})),
		RegisterRoute(types.PayloadTypeValidatorVoteIDs, NewRoute(&validatorVoteIDsRoute{})),
		RegisterRoute(types.PayloadTypeValidatorVoteBodies, NewRoute(&validatorVoteBodiesRoute{})),
		RegisterRoute(types.PayloadTypeCreateResolution, NewRoute(&createResolutionRoute{})),
		RegisterRoute(types.PayloadTypeApproveResolution, NewRoute(&approveResolutionRoute{})),
	)
	if err != nil {
		panic(fmt.Sprintf("failed to register routes: %s", err))
	}
}

// Route is a type that the router uses to handle a certain payload type.
type Route interface {
	Pricer
	// Execute is responsible for committing or rolling back types.
	// All transactions should spend, regardless of success or failure.
	// Therefore, a nested transaction should be used for all database
	// operations after the initial checkAndSpend.
	Execute(ctx *types.TxContext, router *TxApp, db sql.DB, tx *types.Transaction) *TxResponse
}

// NewRoute creates a complete Route for the TxApp from a consensus.Route.
func NewRoute(impl consensus.Route) Route {
	return &baseRoute{impl}
}

// RegisterRouteImpl associates a consensus.Route with a payload type. This is
// shorthand for RegisterRoute(payloadType, NewRoute(route)).
func RegisterRouteImpl(payloadType types.PayloadType, route consensus.Route) error {
	return RegisterRoute(payloadType, NewRoute(route))
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
	Price(ctx context.Context, router *TxApp, db sql.DB, tx *types.Transaction) (*big.Int, error)
}

func codeForEngineError(err error) types.TxCode {
	if err == nil {
		return types.CodeOk
	}
	return types.CodeUnknownError
}

// routes is a map of transaction payload types to their respective routes. This
// should be updated if a coordinated height-based update introduces new routes
// (or removes existing routes).
var (
	routeMtx sync.RWMutex // rare writes, frequent reads
	routes   = map[string]Route{}
)

func getRoute(name string) Route {
	routeMtx.RLock()
	defer routeMtx.RUnlock()
	return routes[name]
}

// RegisterRoute associates a Route with a payload type. See also
// RegisterRouteImpl to register a consensus.Route.
func RegisterRoute(payloadType types.PayloadType, route Route) error {
	typeName := payloadType.String()

	routeMtx.Lock()
	defer routeMtx.Unlock()
	_, ok := routes[typeName]
	if ok {
		return fmt.Errorf("route for payload type %s already exists", typeName)
	}

	routes[typeName] = route
	return nil
}

// baseRoute provides the Price and Execute methods used by TxApp, and embeds a
// consensus.Route, which provides the implementation for the route in a way
// that does not require a pointer to the TxApp instance as an input.
//
// The Execute method is essentially boilerplate code that creates a DB
// transaction, performs the pricing and spending using the Routes Price method,
// runs route-specific operations implemented in the PreTx method, creates a new
// nested DB transaction, and runs more route-specific operations in the InTx
// method inside this inner DB transaction. Finally, the transaction is either
// committed or rolled back.
type baseRoute struct {
	consensus.Route
}

func (d *baseRoute) Price(ctx context.Context, router *TxApp, db sql.DB, tx *types.Transaction) (*big.Int, error) {
	return d.Route.Price(ctx, &types.App{
		Service: router.service.NamedLogger("route_" + d.Name()),
		DB:      db,
		// Engine:     router.Engine,
		Accounts:   router.Accounts,
		Validators: router.Validators,
	}, tx)
}

func (d *baseRoute) Execute(ctx *types.TxContext, router *TxApp, db sql.DB, tx *types.Transaction) *TxResponse {
	dbTx, err := db.BeginTx(ctx.Ctx)
	if err != nil {
		return txRes(nil, types.CodeUnknownError, err)
	}

	spend, code, err := router.checkAndSpend(ctx, tx, d, dbTx)
	if err != nil {
		switch code {
		case types.CodeOk, types.CodeInsufficientBalance, types.CodeInsufficientFee:
			logErr(router.service.Logger, dbTx.Commit(ctx.Ctx))
		default:
			logErr(router.service.Logger, dbTx.Rollback(ctx.Ctx))
		}
		return txRes(spend, code, err)
	}
	defer func() {
		// Always Commit the outer transaction to ensure account updates.
		// Failures in route-specific queries are isolated with a nested
		// transaction (tx2 below).
		err := dbTx.Commit(ctx.Ctx) // must not fail this or user spend is reverted
		if err != nil {
			router.service.Logger.Error("failed to commit DB tx for the spend", err)
		}
	}()

	svc := router.service.NamedLogger("route_" + d.Name())

	code, err = d.PreTx(ctx, svc, tx)
	if err != nil {
		return txRes(spend, code, err)
	}

	tx2, err := dbTx.BeginTx(ctx.Ctx)
	if err != nil {
		return txRes(spend, types.CodeUnknownError, err)
	}
	defer tx2.Rollback(ctx.Ctx) // no-op if Commit succeeded

	app := &types.App{
		Service: svc,
		DB:      tx2,
		// Engine:     router.Engine,
		Accounts:   router.Accounts,
		Validators: router.Validators,
	}

	code, err = d.InTx(ctx, app, tx)
	if err != nil {
		return txRes(spend, code, err)
	}

	err = tx2.Commit(ctx.Ctx)
	if err != nil {
		return txRes(spend, types.CodeUnknownError, err)
	}

	return txRes(spend, types.CodeOk, nil)
}

// ========================== route implementations ==========================
// Each of the following route implementation satisfy the consensus.Route
// interface, which is embedded by the baseRoute for used by TxApp.

// How would we change price? The Price method would store the value in a field
// of the route, which is modified by the app. Alternatively, create a new
// route or replace the route entirely (same payload type, new impl).

type transferRoute struct {
	to  []byte
	amt *big.Int
}

var _ consensus.Route = (*transferRoute)(nil)

func (d *transferRoute) Name() string {
	return types.PayloadTypeTransfer.String()
}

func (d *transferRoute) Price(ctx context.Context, app *types.App, tx *types.Transaction) (*big.Int, error) {
	return big.NewInt(210_000), nil
}

func (d *transferRoute) PreTx(ctx *types.TxContext, svc *types.Service, tx *types.Transaction) (types.TxCode, error) {
	if ctx.BlockContext.ChainContext.NetworkParameters.MigrationStatus == types.MigrationInProgress ||
		ctx.BlockContext.ChainContext.NetworkParameters.MigrationStatus == types.MigrationCompleted {
		return types.CodeNetworkInMigration, errors.New("cannot transfer during migration")
	}

	transferBody := &types.Transfer{}
	err := transferBody.UnmarshalBinary(tx.Body.Payload)
	if err != nil {
		return types.CodeEncodingError, err
	}

	bigAmt, ok := new(big.Int).SetString(transferBody.Amount, 10)
	if !ok {
		return types.CodeInvalidAmount, fmt.Errorf("failed to parse amount: %s", transferBody.Amount)
	}

	// Negative send amounts should be blocked at various levels, so we should
	// never get this, but be extra defensive since we cannot allow thievery.
	if bigAmt.Sign() < 0 {
		return types.CodeInvalidAmount, fmt.Errorf("invalid transfer amount: %s", transferBody.Amount)
	}

	d.to = transferBody.To
	d.amt = bigAmt
	return 0, nil
}

func (d *transferRoute) InTx(ctx *types.TxContext, app *types.App, tx *types.Transaction) (types.TxCode, error) {
	err := app.Accounts.Transfer(ctx.Ctx, app.DB, tx.Sender, d.to, d.amt)
	if err != nil {
		if errors.Is(err, accounts.ErrInsufficientFunds) {
			return types.CodeInsufficientBalance, err
		}
		if errors.Is(err, accounts.ErrNegativeBalance) {
			return types.CodeInvalidAmount, err
		}
		return types.CodeUnknownError, err
	}
	return 0, nil
}

type validatorJoinRoute struct {
	power uint64
}

var _ consensus.Route = (*validatorJoinRoute)(nil)

func (d *validatorJoinRoute) Name() string {
	return types.PayloadTypeValidatorJoin.String()
}

func (d *validatorJoinRoute) Price(ctx context.Context, app *types.App, tx *types.Transaction) (*big.Int, error) {
	return big.NewInt(10000000000000), nil
}

func (d *validatorJoinRoute) PreTx(ctx *types.TxContext, svc *types.Service, tx *types.Transaction) (types.TxCode, error) {
	if ctx.BlockContext.ChainContext.NetworkParameters.MigrationStatus == types.MigrationInProgress ||
		ctx.BlockContext.ChainContext.NetworkParameters.MigrationStatus == types.MigrationCompleted {
		return types.CodeNetworkInMigration, errors.New("cannot join validator during migration")
	}

	join := &types.ValidatorJoin{}
	err := join.UnmarshalBinary(tx.Body.Payload)
	if err != nil {
		return types.CodeEncodingError, err
	}

	d.power = join.Power
	return 0, nil
}

func (d *validatorJoinRoute) InTx(ctx *types.TxContext, app *types.App, tx *types.Transaction) (types.TxCode, error) {
	// ensure this candidate is not already a validator
	power, err := app.Validators.GetValidatorPower(ctx.Ctx, app.DB, tx.Sender)
	if err != nil {
		return types.CodeUnknownError, err
	}
	if power > 0 {
		return types.CodeInvalidSender, ErrCallerIsValidator
	}

	// we first need to ensure that this validator does not have a pending join request
	// if it does, we should not allow it to join again
	pending, err := getResolutionsByTypeAndProposer(ctx.Ctx, app.DB, voting.ValidatorJoinEventType, tx.Sender)
	if err != nil {
		return types.CodeUnknownError, err
	}
	if len(pending) > 0 {
		return types.CodeInvalidSender, fmt.Errorf("validator already has a pending join request")
	}

	// there are no pending join requests, so we can create a new one
	joinReq := &voting.UpdatePowerRequest{
		PubKey: tx.Sender,
		Power:  int64(d.power),
	}
	bts, err := joinReq.MarshalBinary()
	if err != nil {
		return types.CodeUnknownError, err
	}

	event := &types.VotableEvent{
		Body: bts,
		Type: voting.ValidatorJoinEventType,
	}

	err = createResolution(ctx.Ctx, app.DB, event, ctx.BlockContext.Height+ctx.BlockContext.ChainContext.NetworkParameters.JoinExpiry, tx.Sender)
	if err != nil {
		return types.CodeUnknownError, err
	}
	// we do not approve, because a joiner is presumably not a voter
	return 0, nil
}

type validatorApproveRoute struct {
	candidate []byte
}

var _ consensus.Route = (*validatorApproveRoute)(nil)

func (d *validatorApproveRoute) Name() string {
	return types.PayloadTypeValidatorApprove.String()
}

func (d *validatorApproveRoute) Price(ctx context.Context, app *types.App, tx *types.Transaction) (*big.Int, error) {
	return big.NewInt(10000000000000), nil
}

func (d *validatorApproveRoute) PreTx(ctx *types.TxContext, svc *types.Service, tx *types.Transaction) (types.TxCode, error) {
	if ctx.BlockContext.ChainContext.NetworkParameters.MigrationStatus == types.MigrationInProgress ||
		ctx.BlockContext.ChainContext.NetworkParameters.MigrationStatus == types.MigrationCompleted {
		return types.CodeNetworkInMigration, errors.New("cannot approve validator join during migration")
	}

	approve := &types.ValidatorApprove{}
	err := approve.UnmarshalBinary(tx.Body.Payload)
	if err != nil {
		return types.CodeEncodingError, err
	}

	if bytes.Equal(approve.Candidate, tx.Sender) {
		return types.CodeInvalidSender, errors.New("cannot approve own join request")
	}

	d.candidate = approve.Candidate
	return 0, nil
}

func (d *validatorApproveRoute) InTx(ctx *types.TxContext, app *types.App, tx *types.Transaction) (types.TxCode, error) {
	// each pending validator can only have one active join request at a time
	// we need to retrieve the join request and ensure that it is still pending
	pending, err := getResolutionsByTypeAndProposer(ctx.Ctx, app.DB, voting.ValidatorJoinEventType, d.candidate)
	if err != nil {
		return types.CodeUnknownError, err
	}
	if len(pending) == 0 {
		return types.CodeInvalidSender, fmt.Errorf("validator does not have a pending join request")
	}
	if len(pending) > 1 {
		// this should never happen, but if it does, we should not allow it
		return types.CodeUnknownError, fmt.Errorf("validator has more than one pending join request. this is an internal bug")
	}

	// ensure that sender is a validator
	power, err := app.Validators.GetValidatorPower(ctx.Ctx, app.DB, tx.Sender)
	if err != nil {
		return types.CodeUnknownError, err
	}
	if power <= 0 {
		return types.CodeInvalidSender, ErrCallerNotValidator
	}

	err = approveResolution(ctx.Ctx, app.DB, pending[0], tx.Sender)
	if err != nil {
		return types.CodeUnknownError, err
	}

	return 0, nil
}

type validatorRemoveRoute struct {
	target []byte
}

var _ consensus.Route = (*validatorRemoveRoute)(nil)

func (d *validatorRemoveRoute) Name() string {
	return types.PayloadTypeValidatorRemove.String()
}

func (d *validatorRemoveRoute) Price(ctx context.Context, app *types.App, tx *types.Transaction) (*big.Int, error) {
	return big.NewInt(100_000), nil
}

func (d *validatorRemoveRoute) PreTx(ctx *types.TxContext, svc *types.Service, tx *types.Transaction) (types.TxCode, error) {
	if ctx.BlockContext.ChainContext.NetworkParameters.MigrationStatus == types.MigrationInProgress ||
		ctx.BlockContext.ChainContext.NetworkParameters.MigrationStatus == types.MigrationCompleted {
		return types.CodeNetworkInMigration, errors.New("cannot remove validator during migration")
	}

	remove := &types.ValidatorRemove{}
	err := remove.UnmarshalBinary(tx.Body.Payload)
	if err != nil {
		return types.CodeEncodingError, err
	}

	d.target = remove.Validator
	return 0, nil
}

func (d *validatorRemoveRoute) InTx(ctx *types.TxContext, app *types.App, tx *types.Transaction) (types.TxCode, error) {
	removeReq := &voting.UpdatePowerRequest{
		PubKey: d.target,
		Power:  0,
	}
	bts, err := removeReq.MarshalBinary()
	if err != nil {
		return types.CodeUnknownError, err
	}

	event := &types.VotableEvent{
		Body: bts,
		Type: voting.ValidatorRemoveEventType,
	}

	// ensure the sender is a validator
	power, err := app.Validators.GetValidatorPower(ctx.Ctx, app.DB, tx.Sender)
	if err != nil {
		return types.CodeUnknownError, err
	}
	if power <= 0 {
		return types.CodeInvalidSender, ErrCallerNotValidator
	}

	// ensure that the target is a validator
	power, err = app.Validators.GetValidatorPower(ctx.Ctx, app.DB, d.target)
	if err != nil {
		return types.CodeUnknownError, err
	}
	if power <= 0 {
		return types.CodeInvalidSender, ErrTargetNotValidator
	}

	// we should try to create the resolution, since validator removals are never
	// officially "started" by the user. Since we don't have a seperare process for
	// creating and approving Validator Removals, check if the resolution already exists
	// and if it does, approve it, otherwise create and approve it.
	exists, err := resolutionExists(ctx.Ctx, app.DB, event.ID())
	if err != nil {
		return types.CodeUnknownError, err
	}

	// if the resolution does not exist, create it
	if !exists {
		err = createResolution(ctx.Ctx, app.DB, event, ctx.BlockContext.Height+ctx.BlockContext.ChainContext.NetworkParameters.JoinExpiry, tx.Sender)
		if err != nil {
			return types.CodeUnknownError, err
		}
	}

	// we need to approve the resolution as well
	err = approveResolution(ctx.Ctx, app.DB, event.ID(), tx.Sender)
	if err != nil {
		return types.CodeUnknownError, err
	}

	return 0, nil
}

type validatorLeaveRoute struct{}

var _ consensus.Route = (*validatorLeaveRoute)(nil)

func (d *validatorLeaveRoute) Name() string {
	return types.PayloadTypeValidatorLeave.String()
}

func (d *validatorLeaveRoute) Price(ctx context.Context, app *types.App, tx *types.Transaction) (*big.Int, error) {
	return big.NewInt(10000000000000), nil
}

func (d *validatorLeaveRoute) PreTx(ctx *types.TxContext, svc *types.Service, tx *types.Transaction) (types.TxCode, error) {
	if ctx.BlockContext.ChainContext.NetworkParameters.MigrationStatus == types.MigrationInProgress ||
		ctx.BlockContext.ChainContext.NetworkParameters.MigrationStatus == types.MigrationCompleted {
		return types.CodeNetworkInMigration, errors.New("cannot leave validator during migration")
	}
	return 0, nil // no payload to decode or validate for this route
}

func (d *validatorLeaveRoute) InTx(ctx *types.TxContext, app *types.App, tx *types.Transaction) (types.TxCode, error) {
	power, err := app.Validators.GetValidatorPower(ctx.Ctx, app.DB, tx.Sender)
	if err != nil {
		return types.CodeUnknownError, err
	}
	if power <= 0 {
		return types.CodeInvalidSender, ErrCallerNotValidator
	}

	const noPower = 0
	err = app.Validators.SetValidatorPower(ctx.Ctx, app.DB, tx.Sender, noPower)
	if err != nil {
		return types.CodeUnknownError, err
	}

	return 0, nil
}

// validatorVoteIDsRoute is a route for approving a set of votes based on their IDs.
type validatorVoteIDsRoute struct{}

var _ consensus.Route = (*validatorVoteIDsRoute)(nil)

func (d *validatorVoteIDsRoute) Name() string {
	return types.PayloadTypeValidatorVoteIDs.String()
}

func (d *validatorVoteIDsRoute) Price(ctx context.Context, app *types.App, tx *types.Transaction) (*big.Int, error) {
	// VoteID pricing is based on the number of vote IDs.
	ids := &types.ValidatorVoteIDs{}
	err := ids.UnmarshalBinary(tx.Body.Payload)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal vote IDs: %w", err)
	}
	return big.NewInt(int64(len(ids.ResolutionIDs)) * ValidatorVoteIDPrice.Int64()), nil
}

func (d *validatorVoteIDsRoute) PreTx(ctx *types.TxContext, svc *types.Service, tx *types.Transaction) (types.TxCode, error) {
	if ctx.BlockContext.ChainContext.NetworkParameters.MigrationStatus == types.MigrationInProgress ||
		ctx.BlockContext.ChainContext.NetworkParameters.MigrationStatus == types.MigrationCompleted {
		return types.CodeNetworkInMigration, errors.New("cannot vote during migration")
	}
	return 0, nil
}

func (d *validatorVoteIDsRoute) InTx(ctx *types.TxContext, app *types.App, tx *types.Transaction) (types.TxCode, error) {
	// if the caller has 0 power, they are not a validator, and should not be able to vote
	power, err := app.Validators.GetValidatorPower(ctx.Ctx, app.DB, tx.Sender)
	if err != nil {
		return types.CodeUnknownError, err
	}
	if power == 0 {
		return types.CodeInvalidSender, ErrCallerNotValidator
	}

	approve := &types.ValidatorVoteIDs{}
	err = approve.UnmarshalBinary(tx.Body.Payload)
	if err != nil {
		return types.CodeEncodingError, err
	}

	// filter out the vote IDs that have already been processed
	ids, err := voting.FilterNotProcessed(ctx.Ctx, app.DB, approve.ResolutionIDs)
	if err != nil {
		return types.CodeUnknownError, err
	}

	fromLocalValidator := bytes.Equal(tx.Sender, app.Service.Identity)

	for _, voteID := range ids {
		err = approveResolution(ctx.Ctx, app.DB, voteID, tx.Sender)
		if err != nil {
			return types.CodeUnknownError, err
		}

		// if from local validator, delete the event now that we have voted on it and network already has the event body
		if fromLocalValidator {
			err = deleteEvent(ctx.Ctx, app.DB, voteID)
			if err != nil {
				return types.CodeUnknownError, err
			}
		}
	}

	if tooLate := len(approve.ResolutionIDs) - len(ids); tooLate > 0 {
		app.Service.Logger.Warn("vote contains resolution IDs that are already done. too late, no refund!", "numTooLate", tooLate)
	}

	return 0, nil
}

// validatorVoteIDsRoute is a route for approving a set of votes based on their IDs.
type validatorVoteBodiesRoute struct {
	events []*types.VotableEvent
}

var _ consensus.Route = (*validatorVoteBodiesRoute)(nil)

func (d *validatorVoteBodiesRoute) Name() string {
	return types.PayloadTypeValidatorVoteBodies.String()
}

func (d *validatorVoteBodiesRoute) Price(ctx context.Context, _ *types.App, tx *types.Transaction) (*big.Int, error) {
	// VoteBody pricing is based on the size of the vote bodies of all the events in the tx payload.
	votes := &types.ValidatorVoteBodies{}
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

func (d *validatorVoteBodiesRoute) PreTx(ctx *types.TxContext, _ *types.Service, tx *types.Transaction) (types.TxCode, error) {
	if ctx.BlockContext.ChainContext.NetworkParameters.MigrationStatus == types.MigrationInProgress ||
		ctx.BlockContext.ChainContext.NetworkParameters.MigrationStatus == types.MigrationCompleted {
		return types.CodeNetworkInMigration, errors.New("cannot vote during migration")
	}

	// Only proposer can issue a VoteBody transaction.
	if !bytes.Equal(tx.Sender, ctx.BlockContext.Proposer) {
		return types.CodeInvalidSender, ErrCallerNotProposer
	}

	vote := &types.ValidatorVoteBodies{}
	err := vote.UnmarshalBinary(tx.Body.Payload)
	if err != nil {
		return types.CodeEncodingError, err
	}

	d.events = vote.Events

	return 0, nil
}

func (d *validatorVoteBodiesRoute) InTx(ctx *types.TxContext, app *types.App, tx *types.Transaction) (types.TxCode, error) {
	fromLocalValidator := bytes.Equal(tx.Sender, app.Service.Identity)

	// Expectation:
	// 1. VoteBody should only include the events for which the resolutions are not yet created. Maybe filter out the events for which the resolutions are already created and ignore them.
	// 2. If the node is the proposer, delete the event from the event store
	for _, event := range d.events {
		resCfg, err := resolutions.GetResolution(event.Type)
		if err != nil {
			return types.CodeUnknownError, err
		}

		ev := &types.VotableEvent{
			Type: event.Type,
			Body: event.Body,
		}

		expiryHeight := ctx.BlockContext.Height + resCfg.ExpirationPeriod
		err = createResolution(ctx.Ctx, app.DB, ev, expiryHeight, tx.Sender)
		if err != nil {
			return types.CodeUnknownError, err
		}

		// since the vote body proposer is implicitly voting for the event,
		// we need to approve the newly created vote body here
		err = approveResolution(ctx.Ctx, app.DB, ev.ID(), tx.Sender)
		if err != nil {
			return types.CodeUnknownError, err
		}

		// If the local validator is the proposer, then we should delete the event from the event store.
		if fromLocalValidator {
			err = deleteEvent(ctx.Ctx, app.DB, ev.ID())
			if err != nil {
				return types.CodeUnknownError, err
			}
		}
	}

	return 0, nil
}

type createResolutionRoute struct {
	resolution *types.VotableEvent
	expiry     int64
}

var _ consensus.Route = (*createResolutionRoute)(nil)

func (d *createResolutionRoute) Name() string {
	return types.PayloadTypeCreateResolution.String()
}

func (d *createResolutionRoute) Price(ctx context.Context, app *types.App, tx *types.Transaction) (*big.Int, error) {
	res := &types.CreateResolution{}
	err := res.UnmarshalBinary(tx.Body.Payload)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal create resolution payload: %w", err)
	}

	if res.Resolution == nil {
		return nil, fmt.Errorf("resolution is nil")
	}

	// similar to the vote body route, pricing is based on the size of the resolution body
	return big.NewInt(int64(len(res.Resolution.Body)) * ValidatorVoteBodyBytePrice), nil
}

func (d *createResolutionRoute) PreTx(ctx *types.TxContext, svc *types.Service, tx *types.Transaction) (types.TxCode, error) {
	if ctx.BlockContext.ChainContext.NetworkParameters.MigrationStatus == types.MigrationInProgress ||
		ctx.BlockContext.ChainContext.NetworkParameters.MigrationStatus == types.MigrationCompleted {
		return types.CodeNetworkInMigration, errors.New("cannot create resolution during migration")
	}

	res := &types.CreateResolution{}
	err := res.UnmarshalBinary(tx.Body.Payload)
	if err != nil {
		return types.CodeEncodingError, err
	}

	if ctx.BlockContext.ChainContext.NetworkParameters.MigrationStatus != types.NoActiveMigration {
		if res.Resolution.Type == voting.StartMigrationEventType {
			return types.CodeNetworkInMigration, errors.New("migration is about to start, cannot accept new migration proposals")
		}
	}

	// Check if its a valid event type
	resCfg, err := resolutions.GetResolution(res.Resolution.Type)
	if err != nil {
		return types.CodeInvalidResolutionType, err
	}

	d.resolution = (*types.VotableEvent)(res.Resolution)
	d.expiry = resCfg.ExpirationPeriod + ctx.BlockContext.Height

	return 0, nil
}

func (d *createResolutionRoute) InTx(ctx *types.TxContext, app *types.App, tx *types.Transaction) (types.TxCode, error) {
	// ensure the sender is a validator
	// only validators can create resolutions
	power, err := app.Validators.GetValidatorPower(ctx.Ctx, app.DB, tx.Sender)
	if err != nil {
		return types.CodeUnknownError, err
	}
	if power <= 0 {
		return types.CodeInvalidSender, ErrCallerNotValidator
	}

	// create the resolution
	// if resolution already exists, it will return an error
	err = createResolution(ctx.Ctx, app.DB, d.resolution, d.expiry, tx.Sender)
	if err != nil {
		return types.CodeUnknownError, err
	}

	// approve the resolution
	err = approveResolution(ctx.Ctx, app.DB, d.resolution.ID(), tx.Sender)
	if err != nil {
		return types.CodeUnknownError, err
	}

	return 0, nil
}

type approveResolutionRoute struct {
	resolutionID *types.UUID
}

var _ consensus.Route = (*approveResolutionRoute)(nil)

func (d *approveResolutionRoute) Name() string {
	return types.PayloadTypeApproveResolution.String()
}

func (d *approveResolutionRoute) Price(ctx context.Context, app *types.App, tx *types.Transaction) (*big.Int, error) {
	return ValidatorVoteIDPrice, nil
}

func (d *approveResolutionRoute) PreTx(ctx *types.TxContext, svc *types.Service, tx *types.Transaction) (types.TxCode, error) {
	if ctx.BlockContext.ChainContext.NetworkParameters.MigrationStatus == types.MigrationInProgress ||
		ctx.BlockContext.ChainContext.NetworkParameters.MigrationStatus == types.MigrationCompleted {
		return types.CodeNetworkInMigration, fmt.Errorf("cannot approve a resolution during migration")
	}

	res := &types.ApproveResolution{}
	err := res.UnmarshalBinary(tx.Body.Payload)
	if err != nil {
		return types.CodeEncodingError, err
	}

	d.resolutionID = res.ResolutionID
	return 0, nil
}

func (d *approveResolutionRoute) InTx(ctx *types.TxContext, app *types.App, tx *types.Transaction) (types.TxCode, error) {
	// ensure the sender is a validator
	power, err := app.Validators.GetValidatorPower(ctx.Ctx, app.DB, tx.Sender)
	if err != nil {
		return types.CodeUnknownError, err
	}
	if power <= 0 {
		return types.CodeInvalidSender, ErrCallerNotValidator
	}

	// Check if the resolution exists and is still pending
	// You can only vote on a resolution that already exists
	resolution, err := resolutionByID(ctx.Ctx, app.DB, d.resolutionID)
	if err != nil {
		return types.CodeUnknownError, err
	}
	if resolution == nil {
		return types.CodeInvalidResolutionType, fmt.Errorf("resolution with ID %s does not exist", d.resolutionID)
	}

	if ctx.BlockContext.ChainContext.NetworkParameters.MigrationStatus != types.NoActiveMigration &&
		resolution.Type == voting.StartMigrationEventType {
		return types.CodeNetworkInMigration, fmt.Errorf("migration is about to start, cannot accept new migration proposals")
	}

	// vote on the resolution
	err = approveResolution(ctx.Ctx, app.DB, d.resolutionID, tx.Sender)
	if err != nil {
		return types.CodeUnknownError, err
	}

	return 0, nil
}

/* enable and test this in the future

type deleteResolutionRoute struct {
	resolutionID *types.UUID
}

var _ consensus.Route = (*deleteResolutionRoute)(nil)

func (d *deleteResolutionRoute) Name() string {
	return types.PayloadTypeDeleteResolution.String()
}

func (d *deleteResolutionRoute) Price(ctx context.Context, app *types.App, tx *types.Transaction) (*big.Int, error) {
	return ValidatorVoteIDPrice, nil
}

func (d *deleteResolutionRoute) PreTx(ctx *types.TxContext, svc *types.Service, tx *types.Transaction) (types.TxCode, error) {
	if ctx.BlockContext.ChainContext.NetworkParameters.MigrationStatus == types.MigrationInProgress ||
		ctx.BlockContext.ChainContext.NetworkParameters.MigrationStatus == types.MigrationCompleted {
		return types.CodeNetworkInMigration, errors.New("cannot delete resolution during migration")
	}

	res := &types.DeleteResolution{}
	err := res.UnmarshalBinary(tx.Body.Payload)
	if err != nil {
		return types.CodeEncodingError, err
	}

	d.resolutionID = res.ResolutionID
	return 0, nil
}

// deleteResolutionRoute is a route for deleting a resolution.
func (d *deleteResolutionRoute) InTx(ctx *types.TxContext, app *types.App, tx *types.Transaction) (types.TxCode, error) {
	// ensure the sender is a validator
	power, err := app.Validators.GetValidatorPower(ctx.Ctx, app.DB, tx.Sender)
	if err != nil {
		return types.CodeUnknownError, err
	}
	if power <= 0 {
		return types.CodeInvalidSender, ErrCallerNotValidator
	}

	// only the resolution proposer can delete the resolution
	resolution, err := resolutionByID(ctx.Ctx, app.DB, d.resolutionID)
	if err != nil {
		return types.CodeUnknownError, err
	}

	if !bytes.Equal(resolution.Proposer, tx.Sender) {
		return types.CodeInvalidSender, errors.New("only the resolution proposer can delete the resolution")
	}

	// delete the resolution
	err = deleteResolution(ctx.Ctx, app.DB, d.resolutionID)
	if err != nil {
		return types.CodeUnknownError, err
	}

	return 0, nil
}
*/
