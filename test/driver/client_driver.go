package driver

import (
	"context"
	"encoding/hex"
	"fmt"
	"os"

	"github.com/kwilteam/kwil-db/core/client"
	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/types/transactions"
	"github.com/kwilteam/kwil-db/core/utils"

	"go.uber.org/zap"
)

func GetEnv(key, defaultValue string) string {
	v, ok := os.LookupEnv(key)
	if !ok {
		return defaultValue
	}
	return v
}

// KwildClientDriver is driver for tests using the `client` package
type KwildClientDriver struct {
	clt    *client.Client
	logger log.Logger
}

type GrpcDriverOpt func(*KwildClientDriver)

func WithLogger(logger log.Logger) GrpcDriverOpt {
	return func(d *KwildClientDriver) {
		d.logger = logger
	}
}

func NewKwildClientDriver(clt *client.Client, opts ...GrpcDriverOpt) *KwildClientDriver {
	driver := &KwildClientDriver{
		clt:    clt,
		logger: log.New(log.Config{}),
	}

	for _, opt := range opts {
		opt(driver)
	}

	return driver
}

func (d *KwildClientDriver) SupportBatch() bool {
	return true
}

func (d *KwildClientDriver) GetUserPublicKey() []byte {
	return d.clt.Signer.Identity()
}

// TxSuccess checks if the transaction was successful
func (d *KwildClientDriver) TxSuccess(ctx context.Context, txHash []byte) error {
	resp, err := d.clt.TxQuery(ctx, txHash)
	if err != nil {
		return fmt.Errorf("failed to query: %w", err)
	}

	d.logger.Info("tx info", zap.Int64("height", resp.Height),
		zap.String("txHash", hex.EncodeToString(txHash)),
		zap.Any("result", resp.TxResult))

	if resp.TxResult.Code != transactions.CodeOk.Uint32() {
		return fmt.Errorf("transaction not ok, %s", resp.TxResult.Log)
	}

	// NOTE: THIS should not be considered a failure, should retry
	if resp.Height < 0 {
		return ErrTxNotConfirmed
	}

	return nil
}

func (d *KwildClientDriver) DBID(name string) string {
	return utils.GenerateDBID(name, d.clt.Signer.Identity())
}

func (d *KwildClientDriver) DeployDatabase(ctx context.Context, db *transactions.Schema) ([]byte, error) {
	rec, err := d.clt.DeployDatabase(ctx, db)
	if err != nil {
		return nil, fmt.Errorf("error deploying database: %w", err)
	}

	d.logger.Debug("deployed database",
		zap.String("name", db.Name), zap.String("owner", hex.EncodeToString(d.clt.Signer.Identity())),
		zap.String("TxHash", rec.Hex()))
	return rec, nil
}

func (d *KwildClientDriver) DatabaseExists(ctx context.Context, dbid string) error {

	dbSchema, err := d.clt.GetSchema(ctx, dbid)
	if err != nil {
		return fmt.Errorf("failed to get database schema: %w", err)
	}

	if dbSchema == nil {
		return fmt.Errorf("database schema is nil")
	}

	return nil
}

func (d *KwildClientDriver) ExecuteAction(ctx context.Context, dbid string, actionName string, actionInputs ...[]any) ([]byte, error) {
	rec, err := d.clt.ExecuteAction(ctx, dbid, actionName, actionInputs)
	if err != nil {
		return nil, fmt.Errorf("error executing query: %w", err)
	}
	return rec, nil
}

func (d *KwildClientDriver) DropDatabase(ctx context.Context, dbName string) ([]byte, error) {
	rec, err := d.clt.DropDatabase(ctx, dbName)
	if err != nil {
		return nil, fmt.Errorf("error dropping database: %w", err)
	}

	d.logger.Info("drop database", zap.String("name", dbName), zap.String("owner", hex.EncodeToString(d.GetUserPublicKey())),
		zap.String("TxHash", rec.Hex()))
	return rec, nil
}

func (d *KwildClientDriver) QueryDatabase(ctx context.Context, dbid, query string) (*client.Records, error) {
	return d.clt.Query(ctx, dbid, query)
}

func (d *KwildClientDriver) Call(ctx context.Context, dbid, action string, inputs []any, withSignature bool) (*client.Records, error) {
	callOpts := make([]client.CallOpt, 0)
	callOpts = append(callOpts, client.Authenticated(withSignature))
	return d.clt.CallAction(ctx, dbid, action, inputs, callOpts...)
}

func (d *KwildClientDriver) ValidatorNodeApprove(ctx context.Context, joinerPubKey []byte) ([]byte, error) {
	return d.clt.ApproveValidator(ctx, joinerPubKey)
}

func (d *KwildClientDriver) ValidatorNodeRemove(ctx context.Context, target []byte) ([]byte, error) {
	return d.clt.RemoveValidator(ctx, target)
}

func (d *KwildClientDriver) ValidatorNodeJoin(ctx context.Context) ([]byte, error) {
	return d.clt.ValidatorJoin(ctx)
}

func (d *KwildClientDriver) ValidatorNodeLeave(ctx context.Context) ([]byte, error) {
	return d.clt.ValidatorLeave(ctx)
}

func (d *KwildClientDriver) ValidatorJoinStatus(ctx context.Context, pubKey []byte) (*types.JoinRequest, error) {
	return d.clt.ValidatorJoinStatus(ctx, pubKey)
}

func (d *KwildClientDriver) ValidatorsList(ctx context.Context) ([]*types.Validator, error) {
	return d.clt.CurrentValidators(ctx)
}
