package driver

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"os"

	ec "github.com/ethereum/go-ethereum/crypto"
	"github.com/kwilteam/kwil-db/core/crypto/auth"
	"github.com/kwilteam/kwil-db/core/log"
	rpcclient "github.com/kwilteam/kwil-db/core/rpc/client"
	"github.com/kwilteam/kwil-db/core/types"
	clientType "github.com/kwilteam/kwil-db/core/types/client"
	"github.com/kwilteam/kwil-db/core/types/transactions"
	"github.com/kwilteam/kwil-db/core/utils"
	ethdeployer "github.com/kwilteam/kwil-db/test/integration/eth-deployer"
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
	clt      clientType.Client
	signer   auth.Signer
	deployer *ethdeployer.Deployer
	logger   log.Logger
}

func NewKwildClientDriver(clt clientType.Client, signer auth.Signer, deployer *ethdeployer.Deployer, logger log.Logger) *KwildClientDriver {
	driver := &KwildClientDriver{
		clt:      clt,
		signer:   signer,
		logger:   logger,
		deployer: deployer,
	}

	return driver
}

func (d *KwildClientDriver) SupportBatch() bool {
	return true
}

func (d *KwildClientDriver) GetUserPublicKey() []byte {
	return d.signer.Identity()
}

// TxSuccess checks if the transaction was successful
func (d *KwildClientDriver) TxSuccess(ctx context.Context, txHash []byte) error {
	resp, err := d.clt.TxQuery(ctx, txHash)
	if err != nil {
		if errors.Is(err, rpcclient.ErrNotFound) {
			return ErrTxNotConfirmed // not quite, but for this driver it's a retry condition
		}
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

func (d *KwildClientDriver) AccountBalance(ctx context.Context, acctID []byte) (*big.Int, error) {
	acct, err := d.clt.GetAccount(ctx, acctID, types.AccountStatusLatest) // confirmed
	if err != nil {
		return nil, err
	}
	return acct.Balance, nil
}

func (d *KwildClientDriver) TransferAmt(ctx context.Context, to []byte, amt *big.Int) (txHash []byte, err error) {
	return d.clt.Transfer(ctx, to, amt)
}

func (d *KwildClientDriver) DBID(name string) string {
	return utils.GenerateDBID(name, d.signer.Identity())
}

func (d *KwildClientDriver) DeployDatabase(ctx context.Context, db *types.Schema) ([]byte, error) {
	rec, err := d.clt.DeployDatabase(ctx, db)
	if err != nil {
		return nil, fmt.Errorf("error deploying database: %w", err)
	}

	d.logger.Debug("deployed database",
		zap.String("name", db.Name), zap.String("owner", hex.EncodeToString(d.signer.Identity())),
		zap.String("TxHash", rec.Hex()))
	return rec, nil
}

func (d *KwildClientDriver) DatabaseExists(ctx context.Context, dbid string) error {
	// check GetSchema
	dbSchema, err := d.clt.GetSchema(ctx, dbid)
	if err != nil {
		return fmt.Errorf("failed to get database schema: %w", err)
	}

	if dbSchema == nil {
		return fmt.Errorf("database schema is nil")
	}

	// check ListDatabases
	dbs, err := d.clt.ListDatabases(ctx, dbSchema.Owner)
	if err != nil {
		return fmt.Errorf("failed to get database list: %w", err)
	}

	found := false
	for _, db := range dbs {
		if db.DBID == dbid {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("ListDatabase: database not found: %s", dbid)
	}

	return nil
}

func (d *KwildClientDriver) Execute(ctx context.Context, dbid string, actionName string, actionInputs ...[]any) ([]byte, error) {
	rec, err := d.clt.Execute(ctx, dbid, actionName, actionInputs)
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

func (d *KwildClientDriver) QueryDatabase(ctx context.Context, dbid, query string) (*clientType.Records, error) {
	return d.clt.Query(ctx, dbid, query)
}

func (d *KwildClientDriver) Call(ctx context.Context, dbid, action string, inputs []any) (*clientType.Records, error) {
	return d.clt.Call(ctx, dbid, action, inputs)
}

func (d *KwildClientDriver) ChainInfo(ctx context.Context) (*types.ChainInfo, error) {
	return d.clt.ChainInfo(ctx)
}

func (d *KwildClientDriver) Approve(ctx context.Context, sender *ecdsa.PrivateKey, amount *big.Int) error {
	if d.deployer == nil {
		return fmt.Errorf("deployer is nil")
	}
	return d.deployer.Approve(ctx, sender, amount)
}

func (d *KwildClientDriver) Deposit(ctx context.Context, sender *ecdsa.PrivateKey, amount *big.Int) error {
	if d.deployer == nil {
		return fmt.Errorf("deployer is nil")
	}

	return d.deployer.Deposit(ctx, sender, amount)
}

func (d *KwildClientDriver) EscrowBalance(ctx context.Context, senderAddress *ecdsa.PrivateKey) (*big.Int, error) {
	if d.deployer == nil {
		return nil, fmt.Errorf("deployer is nil")
	}

	senderAddr := ec.PubkeyToAddress(senderAddress.PublicKey)
	return d.deployer.EscrowBalance(ctx, senderAddr)
}

func (d *KwildClientDriver) UserBalance(ctx context.Context, sender *ecdsa.PrivateKey) (*big.Int, error) {
	if d.deployer == nil {
		return nil, fmt.Errorf("deployer is nil")
	}

	senderAddr := ec.PubkeyToAddress(sender.PublicKey)

	return d.deployer.UserBalance(ctx, senderAddr)
}

func (d *KwildClientDriver) Allowance(ctx context.Context, sender *ecdsa.PrivateKey) (*big.Int, error) {
	if d.deployer == nil {
		return nil, fmt.Errorf("deployer is nil")
	}

	senderAddr := ec.PubkeyToAddress(sender.PublicKey)
	return d.deployer.Allowance(ctx, senderAddr)
}
