package abci_apps

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kwilteam/kwil-db/internal/app/kwild/server"
	"github.com/kwilteam/kwil-db/internal/entity"

	//"github.com/kwilteam/kwil-db/pkg/engine/models"
	kTx "github.com/kwilteam/kwil-db/pkg/tx"
	//"github.com/kwilteam/kwil-db/pkg/utils/serialize"

	// shorthand for chain client service

	abcitypes "github.com/cometbft/cometbft/abci/types"
	cmtCrypto "github.com/cometbft/cometbft/crypto"
	cryptoenc "github.com/cometbft/cometbft/crypto/encoding"
	cmtjson "github.com/cometbft/cometbft/libs/json"
	txsvc "github.com/kwilteam/kwil-db/internal/controller/grpc/txsvc/v1"
	"github.com/kwilteam/kwil-db/internal/node"
	"github.com/kwilteam/kwil-db/internal/usecases/datasets"
	"github.com/kwilteam/kwil-db/pkg/engine/utils"
	"go.uber.org/zap"
)

type KwilDbApplication struct {
	server      *server.Server
	executor    datasets.DatasetUseCaseInterface
	ValUpdates  []abcitypes.ValidatorUpdate
	valInfo     *node.ValidatorsInfo
	joinReqPool *node.JoinRequestPool
}

var _ abcitypes.Application = (*KwilDbApplication)(nil)

func NewKwilDbApplication(srv *server.Server, executor datasets.DatasetUseCaseInterface) (*KwilDbApplication, error) {

	return &KwilDbApplication{
		server:      srv,
		executor:    executor,
		valInfo:     node.NewValidatorsInfo(),
		joinReqPool: node.NewJoinRequestPool(),
	}, nil
}

func (app *KwilDbApplication) Start(ctx context.Context) error {
	return app.server.Start(ctx)
}

func (app *KwilDbApplication) Info(info abcitypes.RequestInfo) abcitypes.ResponseInfo {
	return abcitypes.ResponseInfo{}
}

func (app *KwilDbApplication) Query(query abcitypes.RequestQuery) abcitypes.ResponseQuery {
	return abcitypes.ResponseQuery{}
}

func (app *KwilDbApplication) CheckTx(req_tx abcitypes.RequestCheckTx) abcitypes.ResponseCheckTx {
	var tx kTx.Transaction
	err := json.Unmarshal(req_tx.Tx, &tx)
	if err != nil {
		app.server.Log.Error("failed to unmarshal CheckTx transaction with ", zap.String("error", err.Error()))
		return abcitypes.ResponseCheckTx{Code: 1, Log: err.Error()}
	}
	err = tx.Verify()
	if err != nil {
		app.server.Log.Error("failed to verify CheckTx transaction with ", zap.String("error", err.Error()))
		return abcitypes.ResponseCheckTx{Code: 1, Log: err.Error()}
	}
	app.server.Log.Info("transaction verified", zap.String("tx hash", string(tx.Hash)))
	return abcitypes.ResponseCheckTx{Code: 0}
}

func (app *KwilDbApplication) DeliverTx(req_tx abcitypes.RequestDeliverTx) abcitypes.ResponseDeliverTx {
	var tx kTx.Transaction
	err := json.Unmarshal(req_tx.Tx, &tx)
	if err != nil {
		app.server.Log.Error("failed to unmarshal DeliverTx transaction with ", zap.String("error", err.Error()))
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

	app.server.Log.Error("failed to deliver transaction with ", zap.String("error", err.Error()))
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
	schema, err := txsvc.UnmarshalSchema(tx.Payload)
	if err != nil {
		app.server.Log.Error("ABCI: failed to unmarshal database schema ", zap.String("error", err.Error()))
		return abcitypes.ResponseDeliverTx{Code: 1, Log: err.Error(), Events: append(events, addFailedEvent("deploy", err, "", tx.Sender))}
	}

	if schema.Owner != tx.Sender {
		err = fmt.Errorf("sender is not the owner of the dataset")
		app.server.Log.Error("ABCI: failed to deploy database with ", zap.String("error", err.Error()))
		return abcitypes.ResponseDeliverTx{Code: 1, Log: "Sender is not the owner of the dataset", Events: append(events, addFailedEvent("deploy", err, schema.Owner, tx.Sender))}
	}

	resp, err := app.executor.Deploy(ctx, &entity.DeployDatabase{
		Schema: schema,
		Tx:     tx,
	})
	if err != nil {
		app.server.Log.Error("ABCI: failed to deploy database with ", zap.String("error", err.Error()))
		return abcitypes.ResponseDeliverTx{Code: 1, Log: err.Error(), Events: append(events, addFailedEvent("deploy", err, schema.Owner, tx.Sender))}
	}

	data, err := json.Marshal(resp)
	if err != nil {
		app.server.Log.Error("ABCI: failed to marshal deploy database response with ", zap.String("error", err.Error()))
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
				{Key: "GasUsed", Value: resp.Fee, Index: true},
			},
		},
	}

	app.server.Log.Info("ABCI: deployed database", zap.String("db id", dbid), zap.String("db name", schema.Name), zap.String("db owner", schema.Owner), zap.String("tx sender", tx.Sender))
	return abcitypes.ResponseDeliverTx{Code: 0, Data: data, Log: "Deployed", Events: events}
}

func (app *KwilDbApplication) drop_database(tx *kTx.Transaction) abcitypes.ResponseDeliverTx {
	var events []abcitypes.Event
	ctx := context.Background()
	dsIdent, err := txsvc.UnmarshalDatasetIdentifier(tx.Payload)
	if err != nil {
		app.server.Log.Error("ABCI Drop database: failed to unmarshal dataset identifier ", zap.String("error", err.Error()))
		return abcitypes.ResponseDeliverTx{Code: 1, Log: err.Error(), Events: append(events, addFailedEvent("drop", err, "", tx.Sender))}
	}
	app.server.Log.Info("ABCI Drop database: dropping database", zap.String("db name", dsIdent.Name), zap.String("db owner", dsIdent.Owner), zap.String("tx sender", tx.Sender))

	if dsIdent.Owner != tx.Sender {
		err = fmt.Errorf("sender is not the owner of the dataset")
		app.server.Log.Error("ABCI Drop database: failed to drop database with ", zap.String("error", err.Error()))
		return abcitypes.ResponseDeliverTx{Code: 1, Log: "Sender is not the owner of the dataset", Events: append(events, addFailedEvent("drop", err, dsIdent.Owner, tx.Sender))}
	}

	dbid := utils.GenerateDBID(dsIdent.Name, dsIdent.Owner)
	resp, err := app.executor.Drop(ctx, &entity.DropDatabase{
		DBID: dbid,
		Tx:   tx,
	})
	if err != nil {
		app.server.Log.Error("ABCI Drop database: failed to drop database with ", zap.String("error", err.Error()))
		return abcitypes.ResponseDeliverTx{Code: 1, Log: err.Error(), Events: append(events, addFailedEvent("drop", err, dsIdent.Owner, tx.Sender))}
	}

	data, err := json.Marshal(resp)
	if err != nil {
		app.server.Log.Error("ABCI Drop database: failed to marshal drop database response with ", zap.String("error", err.Error()))
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
				{Key: "GasUsed", Value: resp.Fee, Index: true},
			},
		},
	}
	app.server.Log.Info("ABCI: dropped database", zap.String("db id", dbid), zap.String("db name", dsIdent.Name), zap.String("db owner", dsIdent.Owner), zap.String("tx sender", tx.Sender))
	return abcitypes.ResponseDeliverTx{Code: 0, Data: data, Events: events}
}

func (app *KwilDbApplication) execute_action(tx *kTx.Transaction) abcitypes.ResponseDeliverTx {
	var events []abcitypes.Event
	ctx := context.Background()
	action, err := txsvc.UnmarshalActionExecution(tx.Payload)
	if err != nil {
		app.server.Log.Error("ABCI execute action: failed to unmarshal action execution ", zap.String("error", err.Error()))
		return abcitypes.ResponseDeliverTx{Code: 1, Log: err.Error(), Events: append(events, addFailedEvent("execute", err, "", tx.Sender))}
	}

	resp, err := app.executor.Execute(ctx, &entity.ExecuteAction{
		Tx:            tx,
		ExecutionBody: action,
	})
	if err != nil {
		app.server.Log.Error("ABCI execute action: failed to execute ", zap.String("action", action.Action), zap.String("error", err.Error()))
		return abcitypes.ResponseDeliverTx{Code: 1, Log: err.Error(), Events: append(events, addFailedEvent("execute", err, "", tx.Sender))}
	}

	data, err := json.Marshal(resp)
	if err != nil {
		app.server.Log.Error("ABCI execute action: failed to marshal execute action response with ", zap.String("error", err.Error()))
		return abcitypes.ResponseDeliverTx{Code: 1, Log: err.Error(), Events: append(events, addFailedEvent("execute", err, "", tx.Sender))}
	}

	params := ""
	for _, m := range action.Params {
		for k, v := range m {
			params += fmt.Sprintf("%v:%v,", k, v)
		}
	}
	app.server.Log.Info("ABCI: executed action", zap.String("db id", action.DBID), zap.String("action", action.Action), zap.String("params", params), zap.String("tx sender", tx.Sender))
	events = []abcitypes.Event{
		{
			Type: "execute",
			Attributes: []abcitypes.EventAttribute{
				{Key: "Result", Value: "Success", Index: true},
				{Key: "TxSender", Value: tx.Sender, Index: true},
				{Key: "DbId", Value: action.DBID, Index: true},
				{Key: "Action", Value: action.Action, Index: true},
				{Key: "Params", Value: params, Index: true},
				{Key: "Fee", Value: resp.Fee, Index: true},
				{Key: "TxHash", Value: string(resp.TxHash), Index: true},
			},
		},
	}
	return abcitypes.ResponseDeliverTx{Code: 0, Data: data, Events: events}
}

func (app *KwilDbApplication) validator_approve(tx *kTx.Transaction) abcitypes.ResponseDeliverTx {
	/*
		Tx Sender: Approver Pubkey
		Payload: Joiner PubKey
	*/
	PrintTx(tx)
	approver, err := UnmarshalPublicKey([]byte(tx.Sender))
	if err != nil {
		return abcitypes.ResponseDeliverTx{Code: 1, Log: err.Error()}
	}

	joiner, err := UnmarshalPublicKey(tx.Payload)
	if err != nil {
		return abcitypes.ResponseDeliverTx{Code: 1, Log: err.Error()}
	}
	joinerAddr := joiner.Address().String()
	approverAddr := approver.Address().String()

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
		joinPublicKey, err := cryptoenc.PubKeyToProto(joiner)
		if err != nil {
			fmt.Println("can't encode public key: %w", err)
		}
		app.valInfo.ValAddrToPubKeyMap[joinerAddr] = joinPublicKey
		fmt.Println("Approve Validator: Validator added to ValUpdates ", joinerAddr, ": ", power)
	}
	fmt.Println("AddrMap:", app.valInfo.ValAddrToPubKeyMap)
	return abcitypes.ResponseDeliverTx{Code: 0}
}

func (app *KwilDbApplication) validator_update(tx *kTx.Transaction, is_join bool) (*entity.Validator, error) {
	validator, err := UnmarshalValidator(tx.Payload)
	if err != nil {
		app.server.Log.Error("ABCI validator update: failed to unmarshal validator request ", zap.String("error", err.Error()))
		return nil, err
	}

	fmt.Println("Validator Info:", validator.PubKey, validator)

	joiner, err := UnmarshalPublicKey([]byte(validator.PubKey))
	// var joiner cmtCrypto.PubKey
	// err = cmtjson.Unmarshal([]byte(validator.PubKey), &joiner)
	if err != nil {
		return nil, err
	}
	joinerAddr := joiner.Address().String()
	joinPublicKey, err := cryptoenc.PubKeyToProto(joiner)
	if err != nil {
		fmt.Println("can't encode public key: %w", err)
	}

	if !is_join || app.valInfo.FinalizedValidators[joinerAddr] {
		fmt.Println("Validator Update: Validator already finalized or not a joiner", joinerAddr, validator.Power)
		fmt.Println("AddrMap:", app.valInfo.ValAddrToPubKeyMap)
		valUpdates := abcitypes.Ed25519ValidatorUpdate(joiner.Bytes(), validator.Power)
		if is_join {
			fmt.Println("Join class")
			if _, ok := app.valInfo.ValAddrToPubKeyMap[joinerAddr]; !ok {
				app.ValUpdates = append(app.ValUpdates, valUpdates)
				app.valInfo.ValAddrToPubKeyMap[joinerAddr] = joinPublicKey
			}
		} else {
			fmt.Println("Leave class")
			if _, ok := app.valInfo.ValAddrToPubKeyMap[joinerAddr]; ok {
				app.ValUpdates = append(app.ValUpdates, valUpdates)
				delete(app.valInfo.ValAddrToPubKeyMap, joinerAddr)
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
					app.valInfo.ValAddrToPubKeyMap[joinerAddr] = joinPublicKey
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
		app.server.Log.Error("ABCI validator leave: failed to update validator ", zap.String("error", err.Error()))
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
				{Key: "ValidatorPubKey", Value: string(validator.PubKey), Index: true},
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
		app.server.Log.Error("ABCI validator leave: failed to update validator ", zap.String("error", err.Error()))
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
				{Key: "ValidatorPubKey", Value: string(validator.PubKey), Index: true},
				{Key: "ValidatorPower", Value: fmt.Sprintf("%d", validator.Power), Index: true},
			},
		},
	}
	return abcitypes.ResponseDeliverTx{Code: 0, Events: events}
}

func UnmarshalValidator(payload []byte) (*entity.Validator, error) {
	validator := entity.Validator{}

	err := json.Unmarshal(payload, &validator)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal validator: %w", err)
	}

	return &validator, nil
}

func UnmarshalPublicKey(addr []byte) (cmtCrypto.PubKey, error) {
	var publicKey cmtCrypto.PubKey
	key := fmt.Sprintf(`{"type":"tendermint/PubKeyEd25519","value":%s}`, string(addr))
	fmt.Println("Key:", key)

	err := cmtjson.Unmarshal([]byte(key), &publicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal validator public key: %w", err)
	}
	fmt.Println("publicKey: ", publicKey)
	return publicKey, nil
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
		pubkey, err := cryptoenc.PubKeyFromProto(val.PubKey)
		if err != nil {
			fmt.Println("can't decode public key: %w", err)
		}
		publicKey, err := cryptoenc.PubKeyToProto(pubkey)
		if err != nil {
			fmt.Println("can't encode public key: %w", err)
		}

		app.valInfo.ValAddrToPubKeyMap[pubkey.Address().String()] = publicKey
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
	app.ValUpdates = make([]abcitypes.ValidatorUpdate, 0)
	// Punish bad validators
	for _, ev := range req.ByzantineValidators {
		if ev.Type == abcitypes.MisbehaviorType_DUPLICATE_VOTE {
			addr := string(ev.Validator.Address)
			if pubKey, ok := app.valInfo.ValAddrToPubKeyMap[addr]; ok {

				app.ValUpdates = append(app.ValUpdates, abcitypes.ValidatorUpdate{PubKey: pubKey, Power: ev.Validator.Power - 1})
				app.server.Log.Info("Decreased val power by 1 because of the equivocation", zap.String("val", addr))
				if (ev.Validator.Power - 1) == 0 {
					app.server.Log.Info("Val power is 0, removing it from the validator set", zap.String("val", addr))
					delete(app.valInfo.ValAddrToPubKeyMap, addr)
					// TODO: Persist these updates to the disk ==> Is it possible to save it in the kwil db sql db?
				}
			} else {
				app.server.Log.Error("Wanted to punish val, but can't find it", zap.String("val", addr))
			}

		}
	}
	return abcitypes.ResponseBeginBlock{}
}

func (app *KwilDbApplication) EndBlock(block abcitypes.RequestEndBlock) abcitypes.ResponseEndBlock {
	return abcitypes.ResponseEndBlock{ValidatorUpdates: app.ValUpdates}
}

func (app *KwilDbApplication) Commit() abcitypes.ResponseCommit {
	return abcitypes.ResponseCommit{Data: []byte{}}
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
