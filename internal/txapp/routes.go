package txapp

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math/big"
	"sync"

	"github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/common/ident"
	"github.com/kwilteam/kwil-db/common/sql"
	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/types/transactions"
	"github.com/kwilteam/kwil-db/extensions/consensus"
	"github.com/kwilteam/kwil-db/extensions/resolutions"
	"github.com/kwilteam/kwil-db/internal/accounts"
	"github.com/kwilteam/kwil-db/internal/engine/execution"
	"github.com/kwilteam/kwil-db/internal/voting"
)

func init() {
	err := errors.Join(
		RegisterRoute(transactions.PayloadTypeDeploySchema, NewRoute(&deployDatasetRoute{})),
		RegisterRoute(transactions.PayloadTypeDropSchema, NewRoute(&dropDatasetRoute{})),
		RegisterRoute(transactions.PayloadTypeExecute, NewRoute(&executeActionRoute{})),
		RegisterRoute(transactions.PayloadTypeTransfer, NewRoute(&transferRoute{})),
		RegisterRoute(transactions.PayloadTypeValidatorJoin, NewRoute(&validatorJoinRoute{})),
		RegisterRoute(transactions.PayloadTypeValidatorApprove, NewRoute(&validatorApproveRoute{})),
		RegisterRoute(transactions.PayloadTypeValidatorRemove, NewRoute(&validatorRemoveRoute{})),
		RegisterRoute(transactions.PayloadTypeValidatorLeave, NewRoute(&validatorLeaveRoute{})),
		RegisterRoute(transactions.PayloadTypeValidatorVoteIDs, NewRoute(&validatorVoteIDsRoute{})),
		RegisterRoute(transactions.PayloadTypeValidatorVoteBodies, NewRoute(&validatorVoteBodiesRoute{})),
		RegisterRoute(transactions.PayloadTypeCreateResolution, NewRoute(&createResolutionRoute{})),
		RegisterRoute(transactions.PayloadTypeApproveResolution, NewRoute(&approveResolutionRoute{})),
	)
	if err != nil {
		panic(fmt.Sprintf("failed to register routes: %s", err))
	}
}

// Route is a type that the router uses to handle a certain payload type.
type Route interface {
	Pricer
	// Execute is responsible for committing or rolling back transactions.
	// All transactions should spend, regardless of success or failure.
	// Therefore, a nested transaction should be used for all database
	// operations after the initial checkAndSpend.
	Execute(ctx *common.TxContext, router *TxApp, db sql.DB, tx *transactions.Transaction) *TxResponse
}

// NewRoute creates a complete Route for the TxApp from a consensus.Route.
func NewRoute(impl consensus.Route) Route {
	return &baseRoute{impl}
}

// RegisterRouteImpl associates a consensus.Route with a payload type. This is
// shorthand for RegisterRoute(payloadType, NewRoute(route)).
func RegisterRouteImpl(payloadType transactions.PayloadType, route consensus.Route) error {
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
	Price(ctx context.Context, router *TxApp, db sql.DB, tx *transactions.Transaction) (*big.Int, error)
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
func RegisterRoute(payloadType transactions.PayloadType, route Route) error {
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

func (d *baseRoute) Price(ctx context.Context, router *TxApp, db sql.DB, tx *transactions.Transaction) (*big.Int, error) {
	return d.Route.Price(ctx, &common.App{
		Service: &common.Service{
			Logger:           router.log.Named("route_" + d.Name()).Sugar(),
			ExtensionConfigs: router.extensionConfigs,
			Identity:         router.signer.Identity(),
		},
		DB:     db,
		Engine: router.Engine,
	}, tx)
}

func (d *baseRoute) Execute(ctx *common.TxContext, router *TxApp, db sql.DB, tx *transactions.Transaction) *TxResponse {
	dbTx, err := db.BeginTx(ctx.Ctx)
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
	defer func() {
		// Always Commit the outer transaction to ensure account updates.
		// Failures in route-specific queries are isolated with a nested
		// transaction (tx2 below).
		err := dbTx.Commit(ctx.Ctx) // must not fail this or user spend is reverted
		if err != nil {
			router.log.Error("failed to commit DB tx for the spend", log.Error(err))
		}
	}()

	svc := &common.Service{
		Logger:           router.log.Named("route_" + d.Name()).Sugar(),
		ExtensionConfigs: router.extensionConfigs,
		Identity:         router.signer.Identity(),
	}

	code, err = d.PreTx(ctx, svc, tx)
	if err != nil {
		return txRes(spend, code, err)
	}

	tx2, err := dbTx.BeginTx(ctx.Ctx)
	if err != nil {
		return txRes(spend, transactions.CodeUnknownError, err)
	}
	defer tx2.Rollback(ctx.Ctx) // no-op if Commit succeeded

	app := &common.App{
		Service: svc,
		DB:      tx2,
		Engine:  router.Engine,
	}

	code, err = d.InTx(ctx, app, tx)
	if err != nil {
		return txRes(spend, code, err)
	}

	err = tx2.Commit(ctx.Ctx)
	if err != nil {
		return txRes(spend, transactions.CodeUnknownError, err)
	}

	return txRes(spend, transactions.CodeOk, nil)
}

// ========================== route implementations ==========================
// Each of the following route implementation satisfy the consensus.Route
// interface, which is embedded by the baseRoute for used by TxApp.

// How would we change price? The Price method would store the value in a field
// of the route, which is modified by the app. Alternatively, create a new
// route or replace the route entirely (same payload type, new impl).

type deployDatasetRoute struct {
	schema     *types.Schema // set by PreTx
	identifier string
	authType   string
}

var _ consensus.Route = (*deployDatasetRoute)(nil)

func (d *deployDatasetRoute) Name() string {
	return transactions.PayloadTypeDeploySchema.String()
}

func (d *deployDatasetRoute) Price(ctx context.Context, app *common.App, tx *transactions.Transaction) (*big.Int, error) {
	return big.NewInt(1000000000000000000), nil
}

func (d *deployDatasetRoute) PreTx(ctx *common.TxContext, svc *common.Service, tx *transactions.Transaction) (transactions.TxCode, error) {
	if ctx.BlockContext.ChainContext.NetworkParameters.InMigration {
		return transactions.CodeNetworkInMigration, fmt.Errorf("cannot deploy dataset during migration")
	}

	schemaPayload := &transactions.Schema{}
	err := schemaPayload.UnmarshalBinary(tx.Body.Payload)
	if err != nil {
		return transactions.CodeEncodingError, err
	}

	d.schema, err = schemaPayload.ToTypes()
	if err != nil {
		return transactions.CodeUnknownError, err
	}

	d.identifier, err = ident.Identifier(tx.Signature.Type, tx.Sender)
	if err != nil {
		return transactions.CodeUnknownError, err
	}

	d.authType = tx.Signature.Type

	return 0, nil
}

func (d *deployDatasetRoute) InTx(ctx *common.TxContext, app *common.App, tx *transactions.Transaction) (transactions.TxCode, error) {
	err := app.Engine.CreateDataset(ctx, app.DB, d.schema)
	if err != nil {
		return transactions.CodeUnknownError, err
	}
	return 0, nil
}

type dropDatasetRoute struct {
	dbid string
}

var _ consensus.Route = (*dropDatasetRoute)(nil)

func (d *dropDatasetRoute) Name() string {
	return transactions.PayloadTypeDropSchema.String()
}

func (d *dropDatasetRoute) Price(ctx context.Context, app *common.App, tx *transactions.Transaction) (*big.Int, error) {
	return big.NewInt(10000000000000), nil
}

func (d *dropDatasetRoute) PreTx(ctx *common.TxContext, svc *common.Service, tx *transactions.Transaction) (transactions.TxCode, error) {
	if ctx.BlockContext.ChainContext.NetworkParameters.InMigration {
		return transactions.CodeNetworkInMigration, fmt.Errorf("cannot drop dataset during migration")
	}

	drop := &transactions.DropSchema{}
	err := drop.UnmarshalBinary(tx.Body.Payload)
	if err != nil {
		return transactions.CodeEncodingError, err
	}

	d.dbid = drop.DBID
	return 0, nil
}

func (d *dropDatasetRoute) InTx(ctx *common.TxContext, app *common.App, tx *transactions.Transaction) (transactions.TxCode, error) {
	err := app.Engine.DeleteDataset(ctx, app.DB, d.dbid)
	if err != nil {
		return transactions.CodeUnknownError, err
	}
	return 0, nil
}

type executeActionRoute struct {
	dbid   string
	action string
	args   [][]any
}

var _ consensus.Route = (*executeActionRoute)(nil)

func (d *executeActionRoute) Name() string {
	return transactions.PayloadTypeExecute.String()
}

func (d *executeActionRoute) Price(ctx context.Context, app *common.App, tx *transactions.Transaction) (*big.Int, error) {
	return big.NewInt(2000000000000000), nil
}

func (d *executeActionRoute) PreTx(ctx *common.TxContext, svc *common.Service, tx *transactions.Transaction) (transactions.TxCode, error) {
	action := &transactions.ActionExecution{}
	err := action.UnmarshalBinary(tx.Body.Payload)
	if err != nil {
		return transactions.CodeEncodingError, err
	}

	d.action = action.Action
	d.dbid = action.DBID

	// here, we decode the [][]transactions.EncodedTypes into [][]any
	args := make([][]any, len(action.Arguments))
	for i, arg := range action.Arguments {
		args[i] = make([]any, len(arg))
		for j, val := range arg {
			decoded, err := val.Decode()
			if err != nil {
				return transactions.CodeEncodingError, err
			}
			args[i][j] = decoded
		}
	}

	// we want to execute the tx for as many arg arrays exist
	// if there are no arg arrays, we want to execute it once
	if len(args) == 0 {
		args = make([][]any, 1)
	}

	d.args = args

	return 0, nil
}

func (d *executeActionRoute) InTx(ctx *common.TxContext, app *common.App, tx *transactions.Transaction) (transactions.TxCode, error) {
	for i := range d.args {
		_, err := app.Engine.Procedure(ctx, app.DB, &common.ExecutionData{
			Dataset:   d.dbid,
			Procedure: d.action,
			Args:      d.args[i],
		})
		if err != nil {
			return codeForEngineError(err), err
		}
	}
	return 0, nil
}

type transferRoute struct {
	to  []byte
	amt *big.Int
}

var _ consensus.Route = (*transferRoute)(nil)

func (d *transferRoute) Name() string {
	return transactions.PayloadTypeTransfer.String()
}

func (d *transferRoute) Price(ctx context.Context, app *common.App, tx *transactions.Transaction) (*big.Int, error) {
	return big.NewInt(210_000), nil
}

func (d *transferRoute) PreTx(ctx *common.TxContext, svc *common.Service, tx *transactions.Transaction) (transactions.TxCode, error) {
	if ctx.BlockContext.ChainContext.NetworkParameters.InMigration {
		return transactions.CodeNetworkInMigration, fmt.Errorf("cannot transfer during migration")
	}

	transferBody := &transactions.Transfer{}
	err := transferBody.UnmarshalBinary(tx.Body.Payload)
	if err != nil {
		return transactions.CodeEncodingError, err
	}

	bigAmt, ok := new(big.Int).SetString(transferBody.Amount, 10)
	if !ok {
		return transactions.CodeInvalidAmount, fmt.Errorf("failed to parse amount: %s", transferBody.Amount)
	}

	// Negative send amounts should be blocked at various levels, so we should
	// never get this, but be extra defensive since we cannot allow thievery.
	if bigAmt.Sign() < 0 {
		return transactions.CodeInvalidAmount, fmt.Errorf("invalid transfer amount: %s", transferBody.Amount)
	}

	d.to = transferBody.To
	d.amt = bigAmt
	return 0, nil
}

func (d *transferRoute) InTx(ctx *common.TxContext, app *common.App, tx *transactions.Transaction) (transactions.TxCode, error) {
	err := transfer(ctx.Ctx, app.DB, tx.Sender, d.to, d.amt)
	if err != nil {
		if errors.Is(err, accounts.ErrInsufficientFunds) {
			return transactions.CodeInsufficientBalance, err
		}
		if errors.Is(err, accounts.ErrNegativeBalance) {
			return transactions.CodeInvalidAmount, err
		}
		return transactions.CodeUnknownError, err
	}
	return 0, nil
}

type validatorJoinRoute struct {
	power uint64
}

var _ consensus.Route = (*validatorJoinRoute)(nil)

func (d *validatorJoinRoute) Name() string {
	return transactions.PayloadTypeValidatorJoin.String()
}

func (d *validatorJoinRoute) Price(ctx context.Context, app *common.App, tx *transactions.Transaction) (*big.Int, error) {
	return big.NewInt(10000000000000), nil
}

func (d *validatorJoinRoute) PreTx(ctx *common.TxContext, svc *common.Service, tx *transactions.Transaction) (transactions.TxCode, error) {
	if ctx.BlockContext.ChainContext.NetworkParameters.InMigration {
		return transactions.CodeNetworkInMigration, fmt.Errorf("cannot join validator during migration")
	}

	join := &transactions.ValidatorJoin{}
	err := join.UnmarshalBinary(tx.Body.Payload)
	if err != nil {
		return transactions.CodeEncodingError, err
	}

	d.power = join.Power
	return 0, nil
}

func (d *validatorJoinRoute) InTx(ctx *common.TxContext, app *common.App, tx *transactions.Transaction) (transactions.TxCode, error) {
	// ensure this candidate is not already a validator
	power, err := getVoterPower(ctx.Ctx, app.DB, tx.Sender)
	if err != nil {
		return transactions.CodeUnknownError, err
	}
	if power > 0 {
		return transactions.CodeInvalidSender, ErrCallerIsValidator
	}

	// we first need to ensure that this validator does not have a pending join request
	// if it does, we should not allow it to join again
	pending, err := getResolutionsByTypeAndProposer(ctx.Ctx, app.DB, voting.ValidatorJoinEventType, tx.Sender)
	if err != nil {
		return transactions.CodeUnknownError, err
	}
	if len(pending) > 0 {
		return transactions.CodeInvalidSender, fmt.Errorf("validator already has a pending join request")
	}

	// there are no pending join requests, so we can create a new one
	joinReq := &voting.UpdatePowerRequest{
		PubKey: tx.Sender,
		Power:  int64(d.power),
	}
	bts, err := joinReq.MarshalBinary()
	if err != nil {
		return transactions.CodeUnknownError, err
	}

	event := &types.VotableEvent{
		Body: bts,
		Type: voting.ValidatorJoinEventType,
	}

	err = createResolution(ctx.Ctx, app.DB, event, ctx.BlockContext.Height+ctx.BlockContext.ChainContext.NetworkParameters.JoinExpiry, tx.Sender)
	if err != nil {
		return transactions.CodeUnknownError, err
	}
	// we do not approve, because a joiner is presumably not a voter
	return 0, nil
}

type validatorApproveRoute struct {
	candidate []byte
}

var _ consensus.Route = (*validatorApproveRoute)(nil)

func (d *validatorApproveRoute) Name() string {
	return transactions.PayloadTypeValidatorApprove.String()
}

func (d *validatorApproveRoute) Price(ctx context.Context, app *common.App, tx *transactions.Transaction) (*big.Int, error) {
	return big.NewInt(10000000000000), nil
}

func (d *validatorApproveRoute) PreTx(ctx *common.TxContext, svc *common.Service, tx *transactions.Transaction) (transactions.TxCode, error) {
	if ctx.BlockContext.ChainContext.NetworkParameters.InMigration {
		return transactions.CodeNetworkInMigration, fmt.Errorf("cannot approve validator join during migration")
	}

	approve := &transactions.ValidatorApprove{}
	err := approve.UnmarshalBinary(tx.Body.Payload)
	if err != nil {
		return transactions.CodeEncodingError, err
	}

	if bytes.Equal(approve.Candidate, tx.Sender) {
		return transactions.CodeInvalidSender, errors.New("cannot approve own join request")
	}

	d.candidate = approve.Candidate
	return 0, nil
}

func (d *validatorApproveRoute) InTx(ctx *common.TxContext, app *common.App, tx *transactions.Transaction) (transactions.TxCode, error) {
	// each pending validator can only have one active join request at a time
	// we need to retrieve the join request and ensure that it is still pending
	pending, err := getResolutionsByTypeAndProposer(ctx.Ctx, app.DB, voting.ValidatorJoinEventType, d.candidate)
	if err != nil {
		return transactions.CodeUnknownError, err
	}
	if len(pending) == 0 {
		return transactions.CodeInvalidSender, fmt.Errorf("validator does not have a pending join request")
	}
	if len(pending) > 1 {
		// this should never happen, but if it does, we should not allow it
		return transactions.CodeUnknownError, fmt.Errorf("validator has more than one pending join request. this is an internal bug")
	}

	// ensure that sender is a validator
	power, err := getVoterPower(ctx.Ctx, app.DB, tx.Sender)
	if err != nil {
		return transactions.CodeUnknownError, err
	}
	if power <= 0 {
		return transactions.CodeInvalidSender, ErrCallerNotValidator
	}

	err = approveResolution(ctx.Ctx, app.DB, pending[0], tx.Sender)
	if err != nil {
		return transactions.CodeUnknownError, err
	}

	return 0, nil
}

type validatorRemoveRoute struct {
	target []byte
}

var _ consensus.Route = (*validatorRemoveRoute)(nil)

func (d *validatorRemoveRoute) Name() string {
	return transactions.PayloadTypeValidatorRemove.String()
}

func (d *validatorRemoveRoute) Price(ctx context.Context, app *common.App, tx *transactions.Transaction) (*big.Int, error) {
	return big.NewInt(100_000), nil
}

func (d *validatorRemoveRoute) PreTx(ctx *common.TxContext, svc *common.Service, tx *transactions.Transaction) (transactions.TxCode, error) {
	if ctx.BlockContext.ChainContext.NetworkParameters.InMigration {
		return transactions.CodeNetworkInMigration, fmt.Errorf("cannot remove validator during migration")
	}

	remove := &transactions.ValidatorRemove{}
	err := remove.UnmarshalBinary(tx.Body.Payload)
	if err != nil {
		return transactions.CodeEncodingError, err
	}

	d.target = remove.Validator
	return 0, nil
}

func (d *validatorRemoveRoute) InTx(ctx *common.TxContext, app *common.App, tx *transactions.Transaction) (transactions.TxCode, error) {
	removeReq := &voting.UpdatePowerRequest{
		PubKey: d.target,
		Power:  0,
	}
	bts, err := removeReq.MarshalBinary()
	if err != nil {
		return transactions.CodeUnknownError, err
	}

	event := &types.VotableEvent{
		Body: bts,
		Type: voting.ValidatorRemoveEventType,
	}

	// ensure the sender is a validator
	power, err := getVoterPower(ctx.Ctx, app.DB, tx.Sender)
	if err != nil {
		return transactions.CodeUnknownError, err
	}
	if power <= 0 {
		return transactions.CodeInvalidSender, ErrCallerNotValidator
	}

	// we should try to create the resolution, since validator removals are never
	// officially "started" by the user. If it fails because it already exists,
	// then we should do nothing

	err = createResolution(ctx.Ctx, app.DB, event, ctx.BlockContext.Height+ctx.BlockContext.ChainContext.NetworkParameters.JoinExpiry, tx.Sender)
	if errors.Is(err, voting.ErrResolutionAlreadyHasBody) {
		app.Service.Logger.Debug("validator removal resolution already exists")
	} else if err != nil {
		return transactions.CodeUnknownError, err
	}

	// we need to approve the resolution as well
	err = approveResolution(ctx.Ctx, app.DB, event.ID(), tx.Sender)
	if err != nil {
		return transactions.CodeUnknownError, err
	}

	return 0, nil
}

type validatorLeaveRoute struct{}

var _ consensus.Route = (*validatorLeaveRoute)(nil)

func (d *validatorLeaveRoute) Name() string {
	return transactions.PayloadTypeValidatorLeave.String()
}

func (d *validatorLeaveRoute) Price(ctx context.Context, app *common.App, tx *transactions.Transaction) (*big.Int, error) {
	return big.NewInt(10000000000000), nil
}

func (d *validatorLeaveRoute) PreTx(ctx *common.TxContext, svc *common.Service, tx *transactions.Transaction) (transactions.TxCode, error) {
	if ctx.BlockContext.ChainContext.NetworkParameters.InMigration {
		return transactions.CodeNetworkInMigration, fmt.Errorf("cannot leave validator during migration")
	}
	return 0, nil // no payload to decode or validate for this route
}

func (d *validatorLeaveRoute) InTx(ctx *common.TxContext, app *common.App, tx *transactions.Transaction) (transactions.TxCode, error) {
	power, err := getVoterPower(ctx.Ctx, app.DB, tx.Sender)
	if err != nil {
		return transactions.CodeUnknownError, err
	}
	if power <= 0 {
		return transactions.CodeInvalidSender, ErrCallerNotValidator
	}

	const noPower = 0
	err = setVoterPower(ctx.Ctx, app.DB, tx.Sender, noPower)
	if err != nil {
		return transactions.CodeUnknownError, err
	}

	return 0, nil
}

// validatorVoteIDsRoute is a route for approving a set of votes based on their IDs.
type validatorVoteIDsRoute struct{}

var _ consensus.Route = (*validatorVoteIDsRoute)(nil)

func (d *validatorVoteIDsRoute) Name() string {
	return transactions.PayloadTypeValidatorVoteIDs.String()
}

func (d *validatorVoteIDsRoute) Price(ctx context.Context, app *common.App, tx *transactions.Transaction) (*big.Int, error) {
	// VoteID pricing is based on the number of vote IDs.
	ids := &transactions.ValidatorVoteIDs{}
	err := ids.UnmarshalBinary(tx.Body.Payload)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal vote IDs: %w", err)
	}
	return big.NewInt(int64(len(ids.ResolutionIDs)) * ValidatorVoteIDPrice.Int64()), nil
}

func (d *validatorVoteIDsRoute) PreTx(ctx *common.TxContext, svc *common.Service, tx *transactions.Transaction) (transactions.TxCode, error) {
	if ctx.BlockContext.ChainContext.NetworkParameters.InMigration {
		return transactions.CodeNetworkInMigration, fmt.Errorf("cannot vote during migration")
	}
	return 0, nil
}

func (d *validatorVoteIDsRoute) InTx(ctx *common.TxContext, app *common.App, tx *transactions.Transaction) (transactions.TxCode, error) {
	// if the caller has 0 power, they are not a validator, and should not be able to vote
	power, err := getVoterPower(ctx.Ctx, app.DB, tx.Sender)
	if err != nil {
		return transactions.CodeUnknownError, err
	}
	if power == 0 {
		return transactions.CodeInvalidSender, ErrCallerNotValidator
	}

	approve := &transactions.ValidatorVoteIDs{}
	err = approve.UnmarshalBinary(tx.Body.Payload)
	if err != nil {
		return transactions.CodeEncodingError, err
	}

	// filter out the vote IDs that have already been processed
	ids, err := voting.FilterNotProcessed(ctx.Ctx, app.DB, approve.ResolutionIDs)
	if err != nil {
		return transactions.CodeUnknownError, err
	}

	fromLocalValidator := bytes.Equal(tx.Sender, app.Service.Identity)

	for _, voteID := range ids {
		err = approveResolution(ctx.Ctx, app.DB, voteID, tx.Sender)
		if err != nil {
			return transactions.CodeUnknownError, err
		}

		// if from local validator, delete the event now that we have voted on it and network already has the event body
		if fromLocalValidator {
			err = deleteEvent(ctx.Ctx, app.DB, voteID)
			if err != nil {
				return transactions.CodeUnknownError, err
			}
		}
	}

	if tooLate := len(approve.ResolutionIDs) - len(ids); tooLate > 0 {
		app.Service.Logger.Warn("vote contains resolution IDs that are already done. too late, no refund!", log.Int("num", tooLate))
	}

	return 0, nil
}

// validatorVoteIDsRoute is a route for approving a set of votes based on their IDs.
type validatorVoteBodiesRoute struct {
	events []*transactions.VotableEvent
}

var _ consensus.Route = (*validatorVoteBodiesRoute)(nil)

func (d *validatorVoteBodiesRoute) Name() string {
	return transactions.PayloadTypeValidatorVoteBodies.String()
}

func (d *validatorVoteBodiesRoute) Price(ctx context.Context, _ *common.App, tx *transactions.Transaction) (*big.Int, error) {
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

func (d *validatorVoteBodiesRoute) PreTx(ctx *common.TxContext, _ *common.Service, tx *transactions.Transaction) (transactions.TxCode, error) {
	if ctx.BlockContext.ChainContext.NetworkParameters.InMigration {
		return transactions.CodeNetworkInMigration, fmt.Errorf("cannot vote during migration")

	}

	// Only proposer can issue a VoteBody transaction.
	if !bytes.Equal(tx.Sender, ctx.BlockContext.Proposer) {
		return transactions.CodeInvalidSender, ErrCallerNotProposer
	}

	vote := &transactions.ValidatorVoteBodies{}
	err := vote.UnmarshalBinary(tx.Body.Payload)
	if err != nil {
		return transactions.CodeEncodingError, err
	}

	d.events = vote.Events

	return 0, nil
}

func (d *validatorVoteBodiesRoute) InTx(ctx *common.TxContext, app *common.App, tx *transactions.Transaction) (transactions.TxCode, error) {
	fromLocalValidator := bytes.Equal(tx.Sender, app.Service.Identity)

	// Expectation:
	// 1. VoteBody should only include the events for which the resolutions are not yet created. Maybe filter out the events for which the resolutions are already created and ignore them.
	// 2. If the node is the proposer, delete the event from the event store
	for _, event := range d.events {
		resCfg, err := resolutions.GetResolution(event.Type)
		if err != nil {
			return transactions.CodeUnknownError, err
		}

		ev := &types.VotableEvent{
			Type: event.Type,
			Body: event.Body,
		}

		expiryHeight := ctx.BlockContext.Height + resCfg.ExpirationPeriod
		err = createResolution(ctx.Ctx, app.DB, ev, expiryHeight, tx.Sender)
		if err != nil {
			return transactions.CodeUnknownError, err
		}

		// since the vote body proposer is implicitly voting for the event,
		// we need to approve the newly created vote body here
		err = approveResolution(ctx.Ctx, app.DB, ev.ID(), tx.Sender)
		if err != nil {
			return transactions.CodeUnknownError, err
		}

		// If the local validator is the proposer, then we should delete the event from the event store.
		if fromLocalValidator {
			err = deleteEvent(ctx.Ctx, app.DB, ev.ID())
			if err != nil {
				return transactions.CodeUnknownError, err
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
	return transactions.PayloadTypeCreateResolution.String()
}

func (d *createResolutionRoute) Price(ctx context.Context, app *common.App, tx *transactions.Transaction) (*big.Int, error) {
	res := &transactions.CreateResolution{}
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

func (d *createResolutionRoute) PreTx(ctx *common.TxContext, svc *common.Service, tx *transactions.Transaction) (transactions.TxCode, error) {
	if ctx.BlockContext.ChainContext.NetworkParameters.InMigration {
		return transactions.CodeNetworkInMigration, errors.New("cannot create resolution during migration")
	}

	res := &transactions.CreateResolution{}
	err := res.UnmarshalBinary(tx.Body.Payload)
	if err != nil {
		return transactions.CodeEncodingError, err
	}

	// Check if its a valid event type
	resCfg, err := resolutions.GetResolution(res.Resolution.Type)
	if err != nil {
		return transactions.CodeInvalidResolutionType, err
	}

	d.resolution = (*types.VotableEvent)(res.Resolution)
	d.expiry = resCfg.ExpirationPeriod + ctx.BlockContext.Height

	return 0, nil
}

func (d *createResolutionRoute) InTx(ctx *common.TxContext, app *common.App, tx *transactions.Transaction) (transactions.TxCode, error) {
	// ensure the sender is a validator
	// only validators can create resolutions
	power, err := getVoterPower(ctx.Ctx, app.DB, tx.Sender)
	if err != nil {
		return transactions.CodeUnknownError, err
	}
	if power <= 0 {
		return transactions.CodeInvalidSender, ErrCallerNotValidator
	}

	// create the resolution
	err = createResolution(ctx.Ctx, app.DB, d.resolution, d.expiry, tx.Sender)
	if err != nil {
		return transactions.CodeUnknownError, err
	}

	// approve the resolution
	err = approveResolution(ctx.Ctx, app.DB, d.resolution.ID(), tx.Sender)
	if err != nil {
		return transactions.CodeUnknownError, err
	}

	return 0, nil
}

type approveResolutionRoute struct {
	resolutionID *types.UUID
}

var _ consensus.Route = (*approveResolutionRoute)(nil)

func (d *approveResolutionRoute) Name() string {
	return transactions.PayloadTypeApproveResolution.String()
}

func (d *approveResolutionRoute) Price(ctx context.Context, app *common.App, tx *transactions.Transaction) (*big.Int, error) {
	return ValidatorVoteIDPrice, nil
}

func (d *approveResolutionRoute) PreTx(ctx *common.TxContext, svc *common.Service, tx *transactions.Transaction) (transactions.TxCode, error) {
	if ctx.BlockContext.ChainContext.NetworkParameters.InMigration {
		return transactions.CodeNetworkInMigration, errors.New("cannot approve a resolution during migration")
	}

	res := &transactions.ApproveResolution{}
	err := res.UnmarshalBinary(tx.Body.Payload)
	if err != nil {
		return transactions.CodeEncodingError, err
	}

	d.resolutionID = res.ResolutionID
	return 0, nil
}

func (d *approveResolutionRoute) InTx(ctx *common.TxContext, app *common.App, tx *transactions.Transaction) (transactions.TxCode, error) {
	// ensure the sender is a validator
	power, err := getVoterPower(ctx.Ctx, app.DB, tx.Sender)
	if err != nil {
		return transactions.CodeUnknownError, err
	}
	if power <= 0 {
		return transactions.CodeInvalidSender, ErrCallerNotValidator
	}

	// Check if the resolution exists and is still pending
	// You can only vote on a resolution that already exists
	exists, err := resolutionExists(ctx.Ctx, app.DB, d.resolutionID)
	if err != nil {
		return transactions.CodeUnknownError, err
	}
	if !exists {
		return transactions.CodeUnknownError, fmt.Errorf("resolution with ID %s does not exist", d.resolutionID)
	}

	// vote on the resolution
	err = approveResolution(ctx.Ctx, app.DB, d.resolutionID, tx.Sender)
	if err != nil {
		return transactions.CodeUnknownError, err
	}

	return 0, nil
}

type deleteResolutionRoute struct {
	resolutionID *types.UUID
}

var _ consensus.Route = (*deleteResolutionRoute)(nil)

func (d *deleteResolutionRoute) Name() string {
	return transactions.PayloadTypeDeleteResolution.String()
}

func (d *deleteResolutionRoute) Price(ctx context.Context, app *common.App, tx *transactions.Transaction) (*big.Int, error) {
	return ValidatorVoteIDPrice, nil
}

func (d *deleteResolutionRoute) PreTx(ctx *common.TxContext, svc *common.Service, tx *transactions.Transaction) (transactions.TxCode, error) {
	if ctx.BlockContext.ChainContext.NetworkParameters.InMigration {
		return transactions.CodeNetworkInMigration, errors.New("cannot vote during migration")
	}

	res := &transactions.DeleteResolution{}
	err := res.UnmarshalBinary(tx.Body.Payload)
	if err != nil {
		return transactions.CodeEncodingError, err
	}

	d.resolutionID = res.ResolutionID
	return 0, nil
}

// deleteResolutionRoute is a route for deleting a resolution.
func (d *deleteResolutionRoute) InTx(ctx *common.TxContext, app *common.App, tx *transactions.Transaction) (transactions.TxCode, error) {
	// ensure the sender is a validator
	power, err := getVoterPower(ctx.Ctx, app.DB, tx.Sender)
	if err != nil {
		return transactions.CodeUnknownError, err
	}
	if power <= 0 {
		return transactions.CodeInvalidSender, ErrCallerNotValidator
	}

	// only the resolution proposer can delete the resolution
	resolution, err := resolutionByID(ctx.Ctx, app.DB, d.resolutionID)
	if err != nil {
		return transactions.CodeUnknownError, err
	}

	if !bytes.Equal(resolution.Proposer, tx.Sender) {
		return transactions.CodeInvalidSender, errors.New("only the resolution proposer can delete the resolution")
	}

	// delete the resolution
	err = deleteResolution(ctx.Ctx, app.DB, d.resolutionID)
	if err != nil {
		return transactions.CodeUnknownError, err
	}

	return 0, nil
}
