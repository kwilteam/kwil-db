package kwild

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/kwilteam/kwil-db/pkg/abci"

	"github.com/kwilteam/kwil-db/pkg/client"
	"github.com/kwilteam/kwil-db/pkg/engine/utils"
	"github.com/kwilteam/kwil-db/pkg/log"
	"github.com/kwilteam/kwil-db/pkg/transactions"

	types "github.com/cometbft/cometbft/abci/types"
	"go.uber.org/zap"
)

// KwildDriver is a grpc driver for  integration tests
type KwildDriver struct {
	clt    *client.Client
	logger log.Logger
}

type GrpcDriverOpt func(*KwildDriver)

func WithLogger(logger log.Logger) GrpcDriverOpt {
	return func(d *KwildDriver) {
		d.logger = logger
	}
}

func NewKwildDriver(clt *client.Client, opts ...GrpcDriverOpt) *KwildDriver {
	driver := &KwildDriver{
		clt:    clt,
		logger: log.New(log.Config{}),
	}

	for _, opt := range opts {
		opt(driver)
	}

	return driver
}

func (d *KwildDriver) GetUserAddress() string {
	return d.clt.Signer.PubKey().Address().String()
}

func (d *KwildDriver) TxSuccess(ctx context.Context, txHash []byte) error {
	resp, err := d.clt.TxQuery(ctx, txHash)
	if err != nil {
		return fmt.Errorf("failed to query: %w", err)
	}

	d.logger.Info("tx info", zap.Uint64("height", resp.Height),
		zap.String("txHash", strings.ToUpper(hex.EncodeToString(txHash))),
		zap.Any("result", resp.TxResult))

	if resp.TxResult.Code != abci.CodeOk.Uint32() {
		return fmt.Errorf("transaction not ok, %s", resp.TxResult.Log)
	}

	return nil
}

func (d *KwildDriver) DeployDatabase(ctx context.Context, db *transactions.Schema) ([]byte, error) {
	db.Owner = d.GetUserAddress()
	rec, err := d.clt.DeployDatabase(ctx, db)
	if err != nil {
		return nil, fmt.Errorf("error deploying database: %w", err)
	}

	d.logger.Info("deployed database",
		zap.String("name", db.Name), zap.String("owner", db.Owner),
		zap.String("TxHash", rec.Hex()))
	return rec, nil
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

	// TODO: verify more than just existence, check schema structure
	return fmt.Errorf("database does not exist")
}

func (d *KwildDriver) ExecuteAction(ctx context.Context, dbid string, actionName string, actionInputs ...[]any) ([]byte, error) {
	rec, err := d.clt.ExecuteAction(ctx, dbid, actionName, actionInputs...)
	if err != nil {
		return nil, fmt.Errorf("error executing query: %w", err)
	}
	return rec, nil
}

func (d *KwildDriver) DropDatabase(ctx context.Context, dbName string) ([]byte, error) {
	rec, err := d.clt.DropDatabase(ctx, dbName)
	if err != nil {
		return nil, fmt.Errorf("error dropping database: %w", err)
	}

	d.logger.Info("drop database", zap.String("name", dbName), zap.String("owner", d.GetUserAddress()),
		zap.String("TxHash", rec.Hex()))
	return rec, nil
}

func (d *KwildDriver) QueryDatabase(ctx context.Context, dbid, query string) (*client.Records, error) {
	return d.clt.Query(ctx, dbid, query)
}

func (d *KwildDriver) Call(ctx context.Context, dbid, action string, inputs []any, opts ...client.CallOpt) ([]map[string]any, error) {
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

//// Validator related methods

// NOTE: The keys in these validator related methods are base64-encoded.

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
	_, err := d.clt.ValidatorJoin(ctx, joiner, power)
	if err != nil {
		return fmt.Errorf("error joining validator: %w", err)
	}

	return nil
}

func (d *KwildDriver) ValidatorNodeLeave(ctx context.Context, leaver string) error {
	hash, err := d.clt.ValidatorLeave(ctx, leaver)
	if err != nil {
		return fmt.Errorf("error joining validator: %w", err)
	}
	// how come this one goes through the cometBFT client?
	res, err := d.clt.CometBftClient.Tx(ctx, hash, false)
	if err != nil {
		return fmt.Errorf("error getting transaction: %w", err)
	}

	if !GetTransactionResult(res.TxResult.Events[0].Attributes) {
		return fmt.Errorf("failed to join as a validator")
	}

	return nil
}
