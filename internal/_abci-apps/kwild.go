package abci_apps

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kwilteam/kwil-db/internal/controller/grpc/txsvc/v1"
	"github.com/kwilteam/kwil-db/pkg/serialize"
	"github.com/kwilteam/kwil-db/pkg/tx"

	kTx "github.com/kwilteam/kwil-db/pkg/tx"

	abcitypes "github.com/cometbft/cometbft/abci/types"
	cryptoenc "github.com/cometbft/cometbft/crypto/encoding"
	node "github.com/kwilteam/kwil-db/internal/_node"

	// "github.com/kwilteam/kwil-db/internal/usecases/datasets"
	"github.com/kwilteam/kwil-db/pkg/crypto"
	"github.com/kwilteam/kwil-db/pkg/engine/utils"
	"github.com/kwilteam/kwil-db/pkg/log"
	utilpkg "github.com/kwilteam/kwil-db/pkg/utils"
	gowal "github.com/tidwall/wal"
	"go.uber.org/zap"
)

type KwildState struct {
	PrevBlockHeight int64             `json:"prev_block_height"`
	PrevAppHash     []byte            `json:"prev_app_hash"`
	CurValidatorSet map[string][]byte `json:"cur_validator_set"`
	ExecState       string            `json:"exec_state"`
	// "initChain", "precommit", "postcommit", "delivertx"
}

// KwilExecutor is the interface to a Kwil dataset execution engine. This is a
// subset of the full DatasetUseCase method set.
//
// TODO: KwilDbApplication methods themselves don't need this *directly*; Kwil
// database business can be encapsulated in a separate type.
type KwilExecutor interface {
	txsvc.EngineReader

	// methods for state, apphash, consensus...
	StartBlockSession() error
	EndBlockSession() ([]byte, error)
	InitializeAppHash(appHash []byte)

	// "wrong place" methods for a future encapsulated tx executor type
	Spend(ctx context.Context, address string, amount string, nonce int64) error

	Deploy(ctx context.Context, schema *serialize.Schema, tx *tx.Transaction) (*tx.ExecutionResponse, error)
	Drop(ctx context.Context, dbid string, tx *tx.Transaction) (*tx.ExecutionResponse, error)
	Execute(ctx context.Context, dbid string, action string, params []map[string]any, tx *tx.Transaction) (*tx.ExecutionResponse, error)
}

type KwilDbApplication struct {
	state KwildState

	log      log.Logger
	executor KwilExecutor

	ValUpdates  []abcitypes.ValidatorUpdate
	valInfo     *node.ValidatorsInfo
	joinReqPool *node.JoinRequestPool

	BlockWal *gowal.Log
	StateWal *utilpkg.Wal

	recoveryMode bool
}

var _ abcitypes.Application = (*KwilDbApplication)(nil)

func NewKwilDbApplication(log log.Logger, executor KwilExecutor) (*KwilDbApplication, error) {
	CometHomeDir := os.Getenv("COMET_BFT_HOME")
	blockWalPath := filepath.Join(CometHomeDir, "data", "Block.wal")
	wal, err := gowal.Open(blockWalPath, nil)
	if err != nil {
		return nil, err
	}
	// This is done to reset the indexes to 1, [need this to avoid a bug with the truncation of tidwal]
	wal.Write(1, []byte("no-op"))
	wal.TruncateBack(1)

	stateWalPath := filepath.Join(CometHomeDir, "data", "AppState.wal")
	stateWal, err := utilpkg.NewWal(stateWalPath)
	if err != nil {
		return nil, err
	}

	kwild := &KwilDbApplication{
		log:          log,
		executor:     executor,
		valInfo:      node.NewValidatorsInfo(),
		joinReqPool:  node.NewJoinRequestPool(),
		BlockWal:     wal,
		StateWal:     stateWal,
		recoveryMode: false,
		state: KwildState{
			PrevBlockHeight: 0,
			PrevAppHash:     crypto.Sha256([]byte("")), // TODO: Initialize this with the genesis hash
			CurValidatorSet: make(map[string][]byte),
			ExecState:       "init",
		},
	}

	if !stateWal.IsEmpty() {
		fmt.Println("Crash Recovery Mode")
		kwild.recoveryMode = true
		kwild.state = kwild.RetrieveState()
	}
	kwild.executor.InitializeAppHash(kwild.state.PrevAppHash)
	return kwild, nil
}

func (app *KwilDbApplication) Info(info abcitypes.RequestInfo) abcitypes.ResponseInfo {
	return abcitypes.ResponseInfo{
		LastBlockHeight:  app.state.PrevBlockHeight,
		LastBlockAppHash: app.state.PrevAppHash,
	}
}

func (app *KwilDbApplication) Query(query abcitypes.RequestQuery) abcitypes.ResponseQuery {
	return abcitypes.ResponseQuery{}
}

func (app *KwilDbApplication) CheckTx(req_tx abcitypes.RequestCheckTx) abcitypes.ResponseCheckTx {
	var tx kTx.Transaction
	err := json.Unmarshal(req_tx.Tx, &tx)
	if err != nil {
		app.log.Error("failed to unmarshal CheckTx transaction with ", zap.String("error", err.Error()))
		return abcitypes.ResponseCheckTx{Code: 1, Log: err.Error()}
	}
	err = tx.Verify()
	if err != nil {
		app.log.Error("failed to verify CheckTx transaction with ", zap.String("error", err.Error()))
		return abcitypes.ResponseCheckTx{Code: 1, Log: err.Error()}
	}
	//TODO: Move the accounts and nonce verification here:

	app.log.Info("transaction verified", zap.String("tx hash", string(tx.Hash)))
	return abcitypes.ResponseCheckTx{Code: 0}
}

func (app *KwilDbApplication) DeliverTx(req_tx abcitypes.RequestDeliverTx) abcitypes.ResponseDeliverTx {
	if app.recoveryMode && (app.state.ExecState == "precommit" || app.state.ExecState == "postcommit") {
		fmt.Println("Replay Mode: Skipping TX", app.state.ExecState)
		return abcitypes.ResponseDeliverTx{Code: 0, Log: "Replay mode, Txs already executed"}
	}

	var tx kTx.Transaction
	err := json.Unmarshal(req_tx.Tx, &tx)
	if err != nil {
		app.log.Error("failed to unmarshal DeliverTx transaction with ", zap.String("error", err.Error()))
		return abcitypes.ResponseDeliverTx{Code: 1, Log: err.Error()}
	}

	switch tx.PayloadType {
	case kTx.DEPLOY_DATABASE:
		return app.deploy_database(&tx)
	case kTx.DROP_DATABASE:
		return app.drop_database(&tx)
	case kTx.EXECUTE_ACTION:
		return app.execute_action(&tx)
	case kTx.VALIDATOR_JOIN:
		return app.validator_join(&tx)
	case kTx.VALIDATOR_LEAVE:
		return app.validator_leave(&tx)
	case kTx.VALIDATOR_APPROVE:
		return app.validator_approve(&tx)
	default:
		err = fmt.Errorf("unknown payload type: %s", tx.PayloadType)
	}

	app.log.Error("failed to deliver transaction with ", zap.String("error", err.Error()))
	return abcitypes.ResponseDeliverTx{Code: 1, Log: err.Error()}
}

func addFailedEvent(eventType string, err error, owner string, sender string) abcitypes.Event {
	return abcitypes.Event{
		Type: eventType,
		Attributes: []abcitypes.EventAttribute{
			{Key: "Result", Value: "Failed", Index: true},
			{Key: "DbOwner", Value: owner, Index: true},
			{Key: "TxSender", Value: sender, Index: true},
			{Key: "Error", Value: err.Error(), Index: true},
		},
	}
}

func (app *KwilDbApplication) deploy_database(tx *kTx.Transaction) abcitypes.ResponseDeliverTx {
	var events []abcitypes.Event
	ctx := context.Background()
	schema, err := serialize.UnmarshalSchema(tx.Payload)
	if err != nil {
		app.log.Error("ABCI: failed to unmarshal database schema ", zap.String("error", err.Error()))
		return abcitypes.ResponseDeliverTx{Code: 1, Log: err.Error(), Events: append(events, addFailedEvent("deploy", err, "", tx.Sender))}
	}

	if schema.Owner != tx.Sender {
		err = fmt.Errorf("sender is not the owner of the dataset")
		app.log.Error("ABCI: failed to deploy database with ", zap.String("error", err.Error()))
		return abcitypes.ResponseDeliverTx{Code: 1, Log: "Sender is not the owner of the dataset", Events: append(events, addFailedEvent("deploy", err, schema.Owner, tx.Sender))}
	}

	resp, err := app.executor.Deploy(ctx, schema, tx)
	if err != nil {
		app.log.Error("ABCI: failed to deploy database with ", zap.String("error", err.Error()))
		return abcitypes.ResponseDeliverTx{Code: 1, Log: err.Error(), Events: append(events, addFailedEvent("deploy", err, schema.Owner, tx.Sender))}
	}

	data, err := json.Marshal(resp)
	if err != nil {
		app.log.Error("ABCI: failed to marshal deploy database response with ", zap.String("error", err.Error()))
		return abcitypes.ResponseDeliverTx{Code: 1, Log: err.Error(), Events: append(events, addFailedEvent("deploy", err, schema.Owner, tx.Sender))}
	}

	dbid := utils.GenerateDBID(schema.Owner, schema.Name)
	events = []abcitypes.Event{
		{
			Type: "deploy",
			Attributes: []abcitypes.EventAttribute{
				{Key: "Result", Value: "Success", Index: true},
				{Key: "DbOwner", Value: schema.Owner, Index: true},
				{Key: "TxSender", Value: tx.Sender, Index: true},
				{Key: "DbName", Value: schema.Name, Index: true},
				{Key: "DbId", Value: dbid, Index: true},
				{Key: "GasUsed", Value: resp.Fee.String(), Index: true},
			},
		},
	}

	app.log.Info("ABCI: deployed database", zap.String("db id", dbid), zap.String("db name", schema.Name), zap.String("db owner", schema.Owner), zap.String("tx sender", tx.Sender))
	return abcitypes.ResponseDeliverTx{Code: 0, Data: data, Log: "Deployed", Events: events}
}

func (app *KwilDbApplication) drop_database(tx *kTx.Transaction) abcitypes.ResponseDeliverTx {
	var events []abcitypes.Event
	ctx := context.Background()
	dsIdent, err := serialize.UnmarshalDatasetIdentifier(tx.Payload)
	if err != nil {
		app.log.Error("ABCI Drop database: failed to unmarshal dataset identifier ", zap.String("error", err.Error()))
		return abcitypes.ResponseDeliverTx{Code: 1, Log: err.Error(), Events: append(events, addFailedEvent("drop", err, "", tx.Sender))}
	}
	app.log.Info("ABCI Drop database: dropping database", zap.String("db name", dsIdent.Name), zap.String("db owner", dsIdent.Owner), zap.String("tx sender", tx.Sender))

	if dsIdent.Owner != tx.Sender {
		err = fmt.Errorf("sender is not the owner of the dataset")
		app.log.Error("ABCI Drop database: failed to drop database with ", zap.String("error", err.Error()))
		return abcitypes.ResponseDeliverTx{Code: 1, Log: "Sender is not the owner of the dataset", Events: append(events, addFailedEvent("drop", err, dsIdent.Owner, tx.Sender))}
	}

	dbid := utils.GenerateDBID(dsIdent.Name, dsIdent.Owner)
	resp, err := app.executor.Drop(ctx, dbid, tx)
	if err != nil {
		app.log.Error("ABCI Drop database: failed to drop database with ", zap.String("error", err.Error()))
		return abcitypes.ResponseDeliverTx{Code: 1, Log: err.Error(), Events: append(events, addFailedEvent("drop", err, dsIdent.Owner, tx.Sender))}
	}

	data, err := json.Marshal(resp)
	if err != nil {
		app.log.Error("ABCI Drop database: failed to marshal drop database response with ", zap.String("error", err.Error()))
		return abcitypes.ResponseDeliverTx{Code: 1, Log: err.Error(), Events: append(events, addFailedEvent("drop", err, dsIdent.Owner, tx.Sender))}
	}

	events = []abcitypes.Event{
		{
			Type: "drop",
			Attributes: []abcitypes.EventAttribute{
				{Key: "Result", Value: "Success", Index: true},
				{Key: "DbOwner", Value: dsIdent.Owner, Index: true},
				{Key: "DbName", Value: dsIdent.Name, Index: true},
				{Key: "TxSender", Value: tx.Sender, Index: true},
				{Key: "DbId", Value: dbid, Index: true},
				{Key: "GasUsed", Value: resp.Fee.String(), Index: true},
			},
		},
	}
	app.log.Info("ABCI: dropped database", zap.String("db id", dbid), zap.String("db name", dsIdent.Name), zap.String("db owner", dsIdent.Owner), zap.String("tx sender", tx.Sender))
	return abcitypes.ResponseDeliverTx{Code: 0, Data: data, Events: events}
}

func (app *KwilDbApplication) execute_action(tx *kTx.Transaction) abcitypes.ResponseDeliverTx {
	var events []abcitypes.Event
	ctx := context.Background()
	action, err := kTx.UnmarshalExecuteAction(tx.Payload)
	if err != nil {
		app.log.Error("ABCI execute action: failed to unmarshal action execution ", zap.String("error", err.Error()))
		return abcitypes.ResponseDeliverTx{Code: 1, Log: err.Error(), Events: append(events, addFailedEvent("execute", err, "", tx.Sender))}
	}

	resp, err := app.executor.Execute(ctx, action.DBID, action.Action, action.Params, tx)
	if err != nil {
		app.log.Error("ABCI execute action: failed to execute ", zap.String("action", action.Action), zap.String("error", err.Error()))
		return abcitypes.ResponseDeliverTx{Code: 1, Log: err.Error(), Events: append(events, addFailedEvent("execute", err, "", tx.Sender))}
	}

	data, err := json.Marshal(resp)
	if err != nil {
		app.log.Error("ABCI execute action: failed to marshal execute action response with ", zap.String("error", err.Error()))
		return abcitypes.ResponseDeliverTx{Code: 1, Log: err.Error(), Events: append(events, addFailedEvent("execute", err, "", tx.Sender))}
	}

	params := ""
	for _, m := range action.Params {
		for k, v := range m {
			params += fmt.Sprintf("%v:%v,", k, v)
		}
	}
	app.log.Info("ABCI: executed action", zap.String("db id", action.DBID), zap.String("action", action.Action), zap.String("params", params), zap.String("tx sender", tx.Sender))
	events = []abcitypes.Event{
		{
			Type: "execute",
			Attributes: []abcitypes.EventAttribute{
				{Key: "Result", Value: "Success", Index: true},
				{Key: "TxSender", Value: tx.Sender, Index: true},
				{Key: "DbId", Value: action.DBID, Index: true},
				{Key: "Action", Value: action.Action, Index: true},
				{Key: "Params", Value: params, Index: true},
				{Key: "Fee", Value: resp.Fee.String(), Index: true},
				// {Key: "TxHash", Value: string(resp.TxHash), Index: true},
			},
		},
	}
	return abcitypes.ResponseDeliverTx{Code: 0, Data: data, Events: events}
}

func (app *KwilDbApplication) validator_approve(tx *kTx.Transaction) abcitypes.ResponseDeliverTx {
	ctx := context.Background()
	/*
		Tx Sender: Approver Pubkey
		Payload: Joiner PubKey
	*/
	PrintTx(tx)
	approver, err := node.UnmarshalPublicKey([]byte(tx.Sender))
	if err != nil {
		return abcitypes.ResponseDeliverTx{Code: 1, Log: err.Error()}
	}

	joiner, err := node.UnmarshalPublicKey(tx.Payload)
	if err != nil {
		return abcitypes.ResponseDeliverTx{Code: 1, Log: err.Error()}
	}
	joinerAddr := joiner.Address().String()
	approverAddr := approver.Address().String()

	// TODO: this should not spend here.  This is something that should be included in use case, or use case should include here
	err = app.executor.Spend(ctx, approverAddr, tx.Fee, tx.Nonce)
	if err != nil {
		return abcitypes.ResponseDeliverTx{Code: 1, Log: err.Error()}
	}

	// Update the Approved Validators List in the DB
	app.valInfo.AddApprovedValidator(joinerAddr, approverAddr)

	// Add approval vote to the Join request
	app.joinReqPool.AddVote(joinerAddr, approverAddr)
	fmt.Println("Approve Validator: Vote added ", approverAddr, " -> ", joinerAddr)
	if app.joinReqPool.AddToValUpdates(joinerAddr) {
		power, err := app.joinReqPool.GetJoinerPower(joinerAddr)
		if err != nil {
			return abcitypes.ResponseDeliverTx{Code: 1, Log: err.Error()}
		}
		valUpdates := abcitypes.Ed25519ValidatorUpdate(joiner.Bytes(), power)
		app.ValUpdates = append(app.ValUpdates, valUpdates)
		app.joinReqPool.RemoveJoinRequest(joinerAddr)
		app.state.CurValidatorSet[joinerAddr] = tx.Payload
		fmt.Println("Approve Validator: Validator added to ValUpdates ", joinerAddr, ": ", power)
	}
	return abcitypes.ResponseDeliverTx{Code: 0}
}

func (app *KwilDbApplication) validator_update(tx *kTx.Transaction, is_join bool) (*serialize.Validator, error) {
	ctx := context.Background()

	validator, err := node.UnmarshalValidator(tx.Payload)
	if err != nil {
		app.log.Error("ABCI validator update: failed to unmarshal validator request ", zap.String("error", err.Error()))
		return nil, err
	}

	fmt.Println("Validator Info:", validator.PubKey, validator)

	joiner, err := node.UnmarshalPublicKey([]byte(validator.PubKey))
	if err != nil {
		return nil, err
	}
	joinerAddr := joiner.Address().String()

	err = app.executor.Spend(ctx, joinerAddr, tx.Fee, tx.Nonce)
	if err != nil {
		return nil, err
	}

	if !is_join || app.valInfo.FinalizedValidators[joinerAddr] {
		fmt.Println("Validator Update: Validator already finalized or not a joiner", joinerAddr, validator.Power)
		valUpdates := abcitypes.Ed25519ValidatorUpdate(joiner.Bytes(), validator.Power)
		if is_join {
			fmt.Println("Join class")
			if _, ok := app.state.CurValidatorSet[joinerAddr]; !ok {
				app.ValUpdates = append(app.ValUpdates, valUpdates)
				app.state.CurValidatorSet[joinerAddr] = []byte(validator.PubKey)
			}
		} else {
			fmt.Println("Leave class")
			if _, ok := app.state.CurValidatorSet[joinerAddr]; ok {
				app.ValUpdates = append(app.ValUpdates, valUpdates)
				delete(app.state.CurValidatorSet, joinerAddr)
			}
		}
	} else {
		// Create a Join Request
		req := app.joinReqPool.GetJoinRequest(joinerAddr, validator.Power)
		fmt.Println("Join Request created for: ", joinerAddr, validator.Power)
		fmt.Println("Join Request:", req)
		fmt.Println("Validators info", validator)
		// Add votes if any of the validators have already approved the joiner
		for val := range req.ValidatorSet {
			if app.valInfo.IsJoinerApproved(joinerAddr, val) {
				fmt.Println("Validator Update: Validator already approved", val, " -> ", joinerAddr)
				app.joinReqPool.AddVote(joinerAddr, val)
				if app.joinReqPool.AddToValUpdates(joinerAddr) {
					valUpdates := abcitypes.Ed25519ValidatorUpdate(joiner.Bytes(), validator.Power)
					app.ValUpdates = append(app.ValUpdates, valUpdates)
					app.joinReqPool.RemoveJoinRequest(joinerAddr)
					app.valInfo.FinalizedValidators[joinerAddr] = true
					app.state.CurValidatorSet[joinerAddr] = []byte(validator.PubKey)
					fmt.Println("Validator Update: Validator added to ValUpdates ", joinerAddr, ": ", validator.Power)
				}
			}
		}
	}
	return validator, nil
}

func (app *KwilDbApplication) validator_join(tx *kTx.Transaction) abcitypes.ResponseDeliverTx {
	var events []abcitypes.Event

	validator, err := app.validator_update(tx, true)
	if err != nil {
		app.log.Error("ABCI validator leave: failed to update validator ", zap.String("error", err.Error()))
		if validator != nil {
			return abcitypes.ResponseDeliverTx{Code: 1, Log: err.Error(), Events: append(events, addFailedEvent("leave_validator", err, string(validator.PubKey), fmt.Sprintf("%d", validator.Power)))}
		} else {
			return abcitypes.ResponseDeliverTx{Code: 1, Log: err.Error(), Events: append(events, addFailedEvent("leave_validator", err, "", ""))}
		}
	}

	// TODO:  Persist these changes to the disk: only if a validator is removed or added (ignore the power updates )

	events = []abcitypes.Event{
		{
			Type: "validator_join",
			Attributes: []abcitypes.EventAttribute{
				{Key: "Result", Value: "Success", Index: true},
				{Key: "ValidatorPubKey", Value: validator.PubKey, Index: true},
				{Key: "ValidatorPower", Value: fmt.Sprintf("%d", validator.Power), Index: true},
			},
		},
	}
	return abcitypes.ResponseDeliverTx{Code: 0, Events: events}
}

func (app *KwilDbApplication) validator_leave(tx *kTx.Transaction) abcitypes.ResponseDeliverTx {
	var events []abcitypes.Event
	validator, err := app.validator_update(tx, false)
	if err != nil {
		app.log.Error("ABCI validator leave: failed to update validator ", zap.String("error", err.Error()))
		if validator != nil {
			return abcitypes.ResponseDeliverTx{Code: 1, Log: err.Error(), Events: append(events, addFailedEvent("leave_validator", err, string(validator.PubKey), fmt.Sprintf("%d", validator.Power)))}
		} else {
			return abcitypes.ResponseDeliverTx{Code: 1, Log: err.Error(), Events: append(events, addFailedEvent("leave_validator", err, "", ""))}
		}
	}

	// TODO:  Persist these changes to the disk: only if a validator is removed or added (ignore the power updates )

	events = []abcitypes.Event{
		{
			Type: "remove_validator",
			Attributes: []abcitypes.EventAttribute{
				{Key: "Result", Value: "Success", Index: true},
				{Key: "ValidatorPubKey", Value: validator.PubKey, Index: true},
				{Key: "ValidatorPower", Value: fmt.Sprintf("%d", validator.Power), Index: true},
			},
		},
	}
	return abcitypes.ResponseDeliverTx{Code: 0, Events: events}
}

func PrintTx(tx *kTx.Transaction) {
	fmt.Println("Payload type: ", tx.PayloadType)
	fmt.Println("Tx Sender: ", tx.Sender)
	fmt.Println("Tx Payload: ", tx.Payload)
	fmt.Println("Tx Signature: ", tx.Signature)
}

func (app *KwilDbApplication) InitChain(req abcitypes.RequestInitChain) abcitypes.ResponseInitChain {

	app.ValUpdates = append(app.ValUpdates, req.Validators...)
	for _, val := range req.Validators {
		fmt.Println("val.Pubkey: pc/PublicKey : ", val.PubKey)
		pubkey, err := cryptoenc.PubKeyFromProto(val.PubKey)
		if err != nil {
			fmt.Println("can't decode public key: %w", err)
		}
		fmt.Println("Pubkey: crypto.PubKey : ", pubkey)

		app.state.CurValidatorSet[pubkey.Address().String()] = pubkey.Bytes()
	}
	return abcitypes.ResponseInitChain{}
}

func (app *KwilDbApplication) PrepareProposal(proposal abcitypes.RequestPrepareProposal) abcitypes.ResponsePrepareProposal {
	return abcitypes.ResponsePrepareProposal{Txs: proposal.Txs}
}

func (app *KwilDbApplication) ProcessProposal(proposal abcitypes.RequestProcessProposal) abcitypes.ResponseProcessProposal {
	return abcitypes.ResponseProcessProposal{Status: abcitypes.ResponseProcessProposal_ACCEPT}
}

func (app *KwilDbApplication) BeginBlock(req abcitypes.RequestBeginBlock) abcitypes.ResponseBeginBlock {
	if app.recoveryMode {
		return abcitypes.ResponseBeginBlock{}
	}
	app.ValUpdates = make([]abcitypes.ValidatorUpdate, 0)
	// Punish bad validators
	for _, ev := range req.ByzantineValidators {
		if ev.Type == abcitypes.MisbehaviorType_DUPLICATE_VOTE {
			addr := string(ev.Validator.Address)
			if pubKey, ok := app.state.CurValidatorSet[addr]; ok {
				PublicKey, err := node.UnmarshalPublicKey([]byte(pubKey))
				if err != nil {
					fmt.Println("can't decode public key: %w", err)
					return abcitypes.ResponseBeginBlock{}
				}
				key, err := cryptoenc.PubKeyToProto(PublicKey)
				if err != nil {
					fmt.Println("can't encode public key: %w", err)
					return abcitypes.ResponseBeginBlock{}
				}
				app.ValUpdates = append(app.ValUpdates, abcitypes.ValidatorUpdate{PubKey: key, Power: ev.Validator.Power - 1})
				app.log.Info("Decreased val power by 1 because of the equivocation", zap.String("val", addr))
				if (ev.Validator.Power - 1) == 0 {
					app.log.Info("Val power is 0, removing it from the validator set", zap.String("val", addr))
					delete(app.state.CurValidatorSet, addr)
				}
			} else {
				app.log.Error("Wanted to punish val, but can't find it", zap.String("val", addr))
			}
		}
	}

	app.executor.StartBlockSession()
	app.state.ExecState = "delivertx"
	app.UpdateState()

	// Create an accounts savepoint
	return abcitypes.ResponseBeginBlock{}
}

func (app *KwilDbApplication) EndBlock(block abcitypes.RequestEndBlock) abcitypes.ResponseEndBlock {
	return abcitypes.ResponseEndBlock{ValidatorUpdates: app.ValUpdates}
}

func (app *KwilDbApplication) Commit() abcitypes.ResponseCommit {
	// Update state
	app.state.ExecState = "precommit"
	app.UpdateState()

	appHash, err := app.executor.EndBlockSession()
	if err != nil {
		app.log.Error("ABCI: failed to end block session with ", zap.String("error", err.Error()))
		return abcitypes.ResponseCommit{Data: app.state.PrevAppHash}
	}

	app.state.PrevBlockHeight += 1
	app.state.PrevAppHash = appHash
	app.state.ExecState = "postcommit"
	app.recoveryMode = false
	app.UpdateState()

	return abcitypes.ResponseCommit{Data: app.state.PrevAppHash}
}

func (app *KwilDbApplication) UpdateState() {
	stateBts, err := json.Marshal(app.state)
	if err != nil {
		app.log.Error("ABCI: failed to marshal state with ", zap.String("error", err.Error()))
	}
	app.StateWal.OverwriteSync(stateBts)
}

func (app *KwilDbApplication) RetrieveState() KwildState {
	if app.StateWal.IsEmpty() {
		fmt.Println("State is empty")
		return KwildState{}
	}

	state := app.StateWal.Read()
	var stateObj KwildState
	err := json.Unmarshal(state, &stateObj)
	if err != nil {
		app.log.Error("ABCI: failed to unmarshal state with ", zap.String("error", err.Error()))
		return KwildState{}
	}
	return stateObj
}

func (app *KwilDbApplication) ListSnapshots(snapshots abcitypes.RequestListSnapshots) abcitypes.ResponseListSnapshots {
	return abcitypes.ResponseListSnapshots{}
}

func (app *KwilDbApplication) OfferSnapshot(snapshot abcitypes.RequestOfferSnapshot) abcitypes.ResponseOfferSnapshot {
	return abcitypes.ResponseOfferSnapshot{}
}

func (app *KwilDbApplication) LoadSnapshotChunk(chunk abcitypes.RequestLoadSnapshotChunk) abcitypes.ResponseLoadSnapshotChunk {
	return abcitypes.ResponseLoadSnapshotChunk{}
}

func (app *KwilDbApplication) ApplySnapshotChunk(chunk abcitypes.RequestApplySnapshotChunk) abcitypes.ResponseApplySnapshotChunk {
	return abcitypes.ResponseApplySnapshotChunk{}
}
