package driver

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"
	"os"
	"strings"
	"time"

	ec "github.com/ethereum/go-ethereum/crypto"

	clientType "github.com/kwilteam/kwil-db/core/client/types"
	"github.com/kwilteam/kwil-db/core/crypto/auth"
	"github.com/kwilteam/kwil-db/core/log"
	rpcclient "github.com/kwilteam/kwil-db/core/rpc/client"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/utils"

	ethdeployer "github.com/kwilteam/kwil-db/test/integration/eth-deployer"
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

func NewKwildClientDriver(clt clientType.Client, signer auth.Signer,
	deployer *ethdeployer.Deployer, logger log.Logger) *KwildClientDriver {
	return &KwildClientDriver{
		clt:      clt,
		signer:   signer,
		logger:   logger,
		deployer: deployer,
	}
}

func (d *KwildClientDriver) Client() clientType.Client {
	return d.clt
}

func (d *KwildClientDriver) SupportBatch() bool {
	return true
}

func (d *KwildClientDriver) GetUserPublicKey() []byte {
	return d.signer.Identity()
}

// TxSuccess checks if the transaction was successful
func (d *KwildClientDriver) TxSuccess(ctx context.Context, txHash types.Hash) error {
	resp, err := d.clt.TxQuery(ctx, txHash)
	if err != nil {
		if errors.Is(err, rpcclient.ErrNotFound) {
			return ErrTxNotConfirmed // not quite, but for this driver it's a retry condition
		}
		return fmt.Errorf("failed to query: %w", err)
	}

	d.logger.Info("tx info", "height", resp.Height, "txHash", txHash, "result", resp.Result)

	if resp.Result.Code != uint32(types.CodeOk) {
		return fmt.Errorf("transaction not ok: %s", resp.Result.Log)
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

func (d *KwildClientDriver) TransferAmt(ctx context.Context, to []byte, amt *big.Int) (txHash types.Hash, err error) {
	return d.clt.Transfer(ctx, to, amt)
}

func (d *KwildClientDriver) DBID(name string) string {
	return utils.GenerateDBID(name, d.signer.Identity())
}

func (d *KwildClientDriver) Execute(ctx context.Context, dbid string, actionName string, actionInputs ...[]any) (types.Hash, error) {
	rec, err := d.clt.Execute(ctx, dbid, actionName, actionInputs)
	if err != nil {
		return types.Hash{}, fmt.Errorf("error executing query: %w", err)
	}
	return rec, nil
}

func (d *KwildClientDriver) ExecuteSQL(ctx context.Context, sql string, params map[string]any) (types.Hash, error) {
	rec, err := d.clt.ExecuteSQL(ctx, sql, params)
	if err != nil {
		return types.Hash{}, fmt.Errorf("error executing sql statement %s: error: %w", sql, err)
	}
	return rec, nil
}

func (d *KwildClientDriver) QueryDatabase(ctx context.Context, query string) (*types.QueryResult, error) {
	return d.clt.Query(ctx, query, nil)
}

func (d *KwildClientDriver) Call(ctx context.Context, dbid, action string, inputs []any) (*types.CallResult, error) {
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

func (d *KwildClientDriver) Signer() []byte {
	return d.signer.Identity()
}

func (d *KwildClientDriver) Identifier() (string, error) {
	return auth.EthSecp256k1Authenticator{}.Identifier(d.Signer())
}

func (d *KwildClientDriver) TxInfo(ctx context.Context, hash types.Hash) (*types.TxQueryResponse, error) {
	res, err := d.clt.TxQuery(ctx, hash)
	if err != nil {
		if strings.Contains(err.Error(), "transaction not found") {
			// try again, hacking around comet's mempool inconsistency
			time.Sleep(500 * time.Millisecond)
			return d.clt.TxQuery(ctx, hash)
		}
		return nil, err
	}

	return res, nil
}
