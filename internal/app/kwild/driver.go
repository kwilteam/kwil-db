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
	"github.com/kwilteam/kwil-db/pkg/validators"

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

	d.logger.Info("tx info", zap.Int64("height", resp.Height),
		zap.String("txHash", strings.ToUpper(hex.EncodeToString(txHash))),
		zap.Any("result", resp.TxResult))

	if resp.Height < 0 {
		return fmt.Errorf("transaction unconfirmed")
	}

	if resp.TxResult.Code != abci.CodeOk.Uint32() {
		return fmt.Errorf("transaction not ok, %s", resp.TxResult.Log)
	}

	return nil
}

func (d *KwildDriver) DBID(name string) string {
	return utils.GenerateDBID(name, d.clt.Signer.PubKey().Bytes())
}

func (d *KwildDriver) DeployDatabase(ctx context.Context, db *transactions.Schema) ([]byte, error) {
	rec, err := d.clt.DeployDatabase(ctx, db)
	if err != nil {
		return nil, fmt.Errorf("error deploying database: %w", err)
	}

	d.logger.Debug("deployed database",
		zap.String("name", db.Name), zap.Binary("owner", d.clt.Signer.PubKey().Bytes()),
		zap.String("TxHash", rec.Hex()))
	return rec, nil
}

func (d *KwildDriver) DatabaseExists(ctx context.Context, dbid string) error {

	dbSchema, err := d.clt.GetSchema(ctx, dbid)
	if err != nil {
		return fmt.Errorf("failed to get database schema: %w", err)
	}

	if dbSchema == nil {
		return fmt.Errorf("database schema is nil")
	}

	return nil
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

func (d *KwildDriver) ValidatorNodeApprove(ctx context.Context, joinerPubKey []byte) ([]byte, error) {
	return d.clt.ApproveValidator(ctx, joinerPubKey)
}

func (d *KwildDriver) ValidatorNodeJoin(ctx context.Context) ([]byte, error) {
	return d.clt.ValidatorJoin(ctx)
}

func (d *KwildDriver) ValidatorNodeLeave(ctx context.Context) ([]byte, error) {
	return d.clt.ValidatorLeave(ctx)
}

func (d *KwildDriver) ValidatorJoinStatus(ctx context.Context, pubKey []byte) (*validators.JoinRequest, error) {
	return d.clt.ValidatorJoinStatus(ctx, pubKey)
}

func (d *KwildDriver) ValidatorsList(ctx context.Context) ([]*validators.Validator, error) {
	return d.clt.CurrentValidators(ctx)
}
