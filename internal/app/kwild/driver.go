package kwild

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/kwilteam/kwil-db/pkg/client"
	"github.com/kwilteam/kwil-db/pkg/engine/utils"
	"github.com/kwilteam/kwil-db/pkg/log"
	"github.com/kwilteam/kwil-db/pkg/serialize"
	kTx "github.com/kwilteam/kwil-db/pkg/tx"

	types "github.com/cometbft/cometbft/abci/types"
	ec "github.com/ethereum/go-ethereum/crypto"
	"go.uber.org/zap"
)

// KwildDriver is a grpc driver for  integration tests
type KwildDriver struct {
	clt         *client.Client
	pk          *ecdsa.PrivateKey
	gatewayAddr string // to ignore the gatewayAddr returned by the config.service

	logger log.Logger
}

func NewKwildDriver(clt *client.Client, pk *ecdsa.PrivateKey, gatewayAddr string, logger log.Logger) *KwildDriver {
	return &KwildDriver{
		clt:         clt,
		pk:          pk,
		gatewayAddr: gatewayAddr,
		logger:      logger,
	}
}

func (d *KwildDriver) GetUserAddress() string {
	return ec.PubkeyToAddress(d.pk.PublicKey).Hex()
}

// TODO: this likely needs to change; the old Kwild driver is not compatible, since deploy, drop, and execute are asynchronous

func (d *KwildDriver) DeployDatabase(ctx context.Context, db *serialize.Schema) error {
	rec, err := d.clt.DeployDatabase(ctx, db)
	if err != nil {
		fmt.Println("Error deploying database: ", err.Error())
		return fmt.Errorf("error deploying database: %w", err)
	}

	res, err := d.clt.CometBftClient.Tx(ctx, rec.TxHash, false)
	if err != nil {
		return fmt.Errorf("error getting transaction: %w", err)
	}

	if !GetTransactionResult(res.TxResult.Events[0].Attributes) {
		return fmt.Errorf("failed to deploy database")
	}

	d.logger.Debug("deployed database", zap.String("name", db.Name), zap.String("owner", db.Owner))
	return nil
}

func (d *KwildDriver) DatabaseShouldExists(ctx context.Context, owner string, dbName string) error {
	dbid := utils.GenerateDBID(dbName, owner)

	dbSchema, err := d.clt.GetSchema(ctx, dbid)
	if err != nil {
		return fmt.Errorf("failed to get database schema: %w", err)
	}

	if strings.EqualFold(dbSchema.Owner, owner) && strings.EqualFold(dbSchema.Name, dbName) {
		return nil
	}
	return fmt.Errorf("database does not exist")
}

func (d *KwildDriver) ExecuteAction(ctx context.Context, dbid string, actionName string, actionInputs []map[string]any) (*kTx.Receipt, []map[string]any, error) {
	rec, err := d.clt.ExecuteAction(ctx, dbid, actionName, actionInputs)
	if err != nil {
		return nil, nil, fmt.Errorf("error executing query: %w", err)
	}

	res, err := d.clt.CometBftClient.Tx(ctx, rec.TxHash, false)
	if err != nil {
		return nil, nil, fmt.Errorf("error getting transaction: %w", err)
	}

	data := res.TxResult.Data
	var updated_rec *kTx.Receipt
	err = json.Unmarshal(data, &updated_rec)
	if err != nil {
		return nil, nil, err
	}

	outputs, err := client.DecodeOutputs(updated_rec.Body)
	if err != nil {
		return nil, nil, err
	}

	d.logger.Debug("execute action", zap.String("database", dbid), zap.String("action", actionName))
	return rec, outputs, nil
}

func (d *KwildDriver) DropDatabase(ctx context.Context, dbName string) error {
	rec, err := d.clt.DropDatabase(ctx, dbName)
	if err != nil {
		return fmt.Errorf("error dropping database: %w", err)
	}

	res, err := d.clt.CometBftClient.Tx(ctx, rec.TxHash, false)
	if err != nil {
		return fmt.Errorf("error getting transaction: %w", err)
	}

	if !GetTransactionResult(res.TxResult.Events[0].Attributes) {
		return fmt.Errorf("failed to drop database")
	}

	d.logger.Debug("drop database", zap.String("name", dbName), zap.String("owner", d.GetUserAddress()))
	return nil
}

func (d *KwildDriver) QueryDatabase(ctx context.Context, dbid, query string) (*client.Records, error) {
	return d.clt.Query(ctx, dbid, query)
}

func (d *KwildDriver) Call(ctx context.Context, dbid, action string, inputs map[string]any, opts ...client.CallOpt) ([]map[string]any, error) {
	return d.clt.CallAction(ctx, dbid, action, inputs, opts...)
}

func GetTransactionResult(attributes []types.EventAttribute) bool {
	for _, attr := range attributes {
		if attr.Key == "Result" {
			return attr.Value == "Success"
		}
	}
	return false
}

func (d *KwildDriver) ApproveNode(ctx context.Context, joinerPubKey string, approverPrivKey string) error {
	_, err := d.clt.ApproveValidator(ctx, approverPrivKey, joinerPubKey)
	return err
}

func (d *KwildDriver) ValidatorSetCount(ctx context.Context) (int, error) {
	vals, err := d.clt.CometBftClient.Validators(ctx, nil, nil, nil)
	if err != nil {
		return -1, err
	}

	return vals.Count, nil
}

func (d *KwildDriver) ValidatorNodeJoin(ctx context.Context, joiner string, power int64) error {
	_, err := d.clt.ValidatorJoin(ctx, string(joiner), power)
	if err != nil {
		return fmt.Errorf("error joining validator: %w", err)
	}

	return nil
}

func (d *KwildDriver) ValidatorNodeLeave(ctx context.Context, joiner string) error {
	hash, err := d.clt.ValidatorLeave(ctx, joiner, 0)
	if err != nil {
		return fmt.Errorf("error joining validator: %w", err)
	}

	res, err := d.clt.CometBftClient.Tx(ctx, hash, false)
	if err != nil {
		return fmt.Errorf("error getting transaction: %w", err)
	}

	if !GetTransactionResult(res.TxResult.Events[0].Attributes) {
		return fmt.Errorf("failed to join as a validator")
	}

	return nil
}
