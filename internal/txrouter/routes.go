package txrouter

import (
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
	)
	if err != nil {
		panic(fmt.Sprintf("failed to register routes: %s", err))
	}
}

type Route interface {
	PayloadType() string // not sure if we need this
	Execute(ctx context.Context, router *Router, tx *transactions.Transaction) *TxResponse
	Price(ctx context.Context, router *Router, tx *transactions.Transaction) (*big.Int, error)
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

func (d *deployDatasetRoute) Execute(ctx context.Context, router *Router, tx *transactions.Transaction) *TxResponse {
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

	err = router.Database.CreateDataset(ctx, schema, tx.Sender)
	if err != nil {
		return txRes(spend, transactions.CodeUnknownError, err)
	}

	return txRes(spend, transactions.CodeOk, nil)
}

func (d *deployDatasetRoute) PayloadType() string {
	return transactions.PayloadTypeDeploySchema.String()
}

func (d *deployDatasetRoute) Price(ctx context.Context, router *Router, tx *transactions.Transaction) (*big.Int, error) {
	return big.NewInt(1000000000000000000), nil
}

type dropDatasetRoute struct{}

func (d *dropDatasetRoute) Execute(ctx context.Context, router *Router, tx *transactions.Transaction) *TxResponse {
	spend, code, err := router.checkAndSpend(ctx, tx)
	if err != nil {
		return txRes(spend, code, err)
	}

	drop := &transactions.DropSchema{}
	err = drop.UnmarshalBinary(tx.Body.Payload)
	if err != nil {
		return txRes(spend, transactions.CodeEncodingError, err)
	}

	err = router.Database.DeleteDataset(ctx, drop.DBID, tx.Sender)
	if err != nil {
		return txRes(spend, transactions.CodeUnknownError, err)
	}

	return txRes(spend, transactions.CodeOk, nil)
}

func (d *dropDatasetRoute) PayloadType() string {
	return transactions.PayloadTypeDropSchema.String()
}

func (d *dropDatasetRoute) Price(ctx context.Context, router *Router, tx *transactions.Transaction) (*big.Int, error) {
	return big.NewInt(10000000000000), nil
}

type executeActionRoute struct{}

func (e *executeActionRoute) Execute(ctx context.Context, router *Router, tx *transactions.Transaction) *TxResponse {
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
		_, err = router.Database.Execute(ctx, &engineTypes.ExecutionData{
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

func (e *executeActionRoute) PayloadType() string {
	return transactions.PayloadTypeExecuteAction.String()
}

func (e *executeActionRoute) Price(ctx context.Context, router *Router, tx *transactions.Transaction) (*big.Int, error) {
	return big.NewInt(2000000000000000), nil
}

type transferRoute struct{}

func (t *transferRoute) Execute(ctx context.Context, router *Router, tx *transactions.Transaction) *TxResponse {
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

	// check if the sender has enough tokens to transfer
	acct, err := router.Accounts.GetAccount(ctx, tx.Sender)
	if err != nil {
		return txRes(spend, transactions.CodeUnknownError, err)
	}

	if acct.Balance.Cmp(bigAmt) < 0 {
		return txRes(spend, transactions.CodeInsufficientBalance, fmt.Errorf("account %s does not have enough tokens to transfer. account balance: %s, required balance: %s", tx.Sender, acct.Balance.String(), bigAmt.String()))
	}

	err = router.Accounts.Transfer(ctx, transfer.To, tx.Sender, bigAmt)
	if err != nil {
		return txRes(spend, transactions.CodeUnknownError, err)
	}

	return txRes(spend, transactions.CodeOk, nil)
}

func (t *transferRoute) PayloadType() string {
	return transactions.PayloadTypeTransfer.String()
}

func (t *transferRoute) Price(ctx context.Context, router *Router, tx *transactions.Transaction) (*big.Int, error) {
	return big.NewInt(210_000), nil
}

type validatorJoinRoute struct{}

func (v *validatorJoinRoute) Execute(ctx context.Context, router *Router, tx *transactions.Transaction) *TxResponse {
	spend, code, err := router.checkAndSpend(ctx, tx)
	if err != nil {
		return txRes(spend, code, err)
	}

	join := &transactions.ValidatorJoin{}
	err = join.UnmarshalBinary(tx.Body.Payload)
	if err != nil {
		return txRes(spend, transactions.CodeEncodingError, err)
	}

	err = router.Validators.Join(ctx, tx.Sender, int64(join.Power))
	if err != nil {
		return txRes(spend, transactions.CodeUnknownError, err)
	}

	return txRes(spend, transactions.CodeOk, nil)
}

func (v *validatorJoinRoute) PayloadType() string {
	return transactions.PayloadTypeValidatorJoin.String()
}

func (v *validatorJoinRoute) Price(ctx context.Context, router *Router, tx *transactions.Transaction) (*big.Int, error) {
	return big.NewInt(10000000000000), nil
}

type validatorApproveRoute struct{}

func (v *validatorApproveRoute) Execute(ctx context.Context, router *Router, tx *transactions.Transaction) *TxResponse {
	spend, code, err := router.checkAndSpend(ctx, tx)
	if err != nil {
		return txRes(spend, code, err)
	}

	approve := &transactions.ValidatorApprove{}
	err = approve.UnmarshalBinary(tx.Body.Payload)
	if err != nil {
		return txRes(spend, transactions.CodeEncodingError, err)
	}

	err = router.Validators.Approve(ctx, approve.Candidate, tx.Sender)
	if err != nil {
		return txRes(spend, transactions.CodeUnknownError, err)
	}

	return txRes(spend, transactions.CodeOk, nil)
}

func (v *validatorApproveRoute) PayloadType() string {
	return transactions.PayloadTypeValidatorApprove.String()
}

func (v *validatorApproveRoute) Price(ctx context.Context, router *Router, tx *transactions.Transaction) (*big.Int, error) {
	return big.NewInt(10000000000000), nil
}

type validatorRemoveRoute struct{}

func (v *validatorRemoveRoute) Execute(ctx context.Context, router *Router, tx *transactions.Transaction) *TxResponse {
	spend, code, err := router.checkAndSpend(ctx, tx)
	if err != nil {
		return txRes(spend, code, err)
	}

	remove := &transactions.ValidatorRemove{}
	err = remove.UnmarshalBinary(tx.Body.Payload)
	if err != nil {
		return txRes(spend, transactions.CodeEncodingError, err)
	}

	err = router.Validators.Remove(ctx, remove.Validator, tx.Sender)
	if err != nil {
		return txRes(spend, transactions.CodeUnknownError, err)
	}

	return txRes(spend, transactions.CodeOk, nil)
}

func (v *validatorRemoveRoute) PayloadType() string {
	return transactions.PayloadTypeValidatorRemove.String()
}

func (v *validatorRemoveRoute) Price(ctx context.Context, router *Router, tx *transactions.Transaction) (*big.Int, error) {
	return big.NewInt(10000000000000), nil
}

type validatorLeaveRoute struct{}

func (v *validatorLeaveRoute) Execute(ctx context.Context, router *Router, tx *transactions.Transaction) *TxResponse {
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

	err = router.Validators.Leave(ctx, tx.Sender)
	if err != nil {
		return txRes(spend, transactions.CodeUnknownError, err)
	}

	return txRes(spend, transactions.CodeOk, nil)
}

func (v *validatorLeaveRoute) PayloadType() string {
	return transactions.PayloadTypeValidatorLeave.String()
}

func (v *validatorLeaveRoute) Price(ctx context.Context, router *Router, tx *transactions.Transaction) (*big.Int, error) {
	return big.NewInt(10000000000000), nil
}
