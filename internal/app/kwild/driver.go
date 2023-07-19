package kwild

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"time"

	schema "github.com/kwilteam/kwil-db/internal/entity"
	"github.com/kwilteam/kwil-db/pkg/client"
	"github.com/kwilteam/kwil-db/pkg/engine/utils"
	grpc "github.com/kwilteam/kwil-db/pkg/grpc/client/v1"
	"github.com/kwilteam/kwil-db/pkg/log"
	kTx "github.com/kwilteam/kwil-db/pkg/tx"

	types "github.com/cometbft/cometbft/abci/types"
	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	ec "github.com/ethereum/go-ethereum/crypto"
	"go.uber.org/zap"
)

// KwildDriver is a grpc driver for  integration tests
type KwildDriver struct {
	clt         *client.Client
	bcClt       *rpchttp.HTTP
	pk          *ecdsa.PrivateKey
	gatewayAddr string // to ignore the gatewayAddr returned by the config.service

	logger log.Logger
}

func NewKwildDriver(clt *client.Client, bcClt *rpchttp.HTTP, pk *ecdsa.PrivateKey, gatewayAddr string, logger log.Logger) *KwildDriver {
	return &KwildDriver{
		clt:         clt,
		bcClt:       bcClt,
		pk:          pk,
		gatewayAddr: gatewayAddr,
		logger:      logger,
	}
}

func (d *KwildDriver) GetUserAddress() string {
	return ec.PubkeyToAddress(d.pk.PublicKey).Hex()
}

func (d *KwildDriver) GetServiceConfig(ctx context.Context) (*grpc.SvcConfig, error) {
	return d.clt.GetConfig(ctx)
}

func (d *KwildDriver) DepositFund(ctx context.Context, amount *big.Int) error {
	_, err := d.clt.Deposit(ctx, amount)
	if err != nil {
		return fmt.Errorf("failed to send deposit transaction: %w", err)
	}

	d.logger.Debug("deposit fund", zap.String("from", d.GetUserAddress()),
		zap.String("to", d.clt.ProviderAddress), zap.String("amount", amount.String()))
	return nil
}

func (d *KwildDriver) GetDepositBalance(ctx context.Context) (*big.Int, error) {
	bal, err := d.clt.GetDepositedAmount(ctx)
	if err != nil {
		return nil, err
	}
	return bal, nil
}

func (d *KwildDriver) ApproveToken(ctx context.Context, amount *big.Int) error {
	txHash, err := d.clt.ApproveDeposit(ctx, amount)
	if err != nil {
		return err
	}

	fmt.Println("Cherry: approve token txHash", txHash)
	d.logger.Debug("approve token", zap.String("from", ec.PubkeyToAddress(d.pk.PublicKey).Hex()),
		zap.String("spender", d.clt.PoolAddress), zap.String("amount", amount.String()))
	return nil
}

func (d *KwildDriver) GetAllowance(ctx context.Context) (*big.Int, error) {
	amount, err := d.clt.GetApprovedAmount(ctx)
	if err != nil {
		return nil, err
	}

	return amount, nil
}

func (d *KwildDriver) DeployDatabase(ctx context.Context, db *schema.Schema) error {
	rec, err := d.clt.DeployDatabase(ctx, db)
	if err != nil {
		fmt.Println("Error deploying database: ", err.Error())
		return fmt.Errorf("error deploying database: %w", err)
	}
	time.Sleep(15 * time.Second)
	fmt.Printf("Cherry: rec.TxHash %v\n", rec.TxHash)
	res, err := d.bcClt.Tx(ctx, rec.TxHash, false)
	if err != nil {
		fmt.Println("Error getting transaction: ", err.Error())
		return fmt.Errorf("error getting transaction: %w", err)
	}

	if !GetTransactionResult(res.TxResult.Events[0].Attributes) {
		return fmt.Errorf("failed to deploy database")
	}

	fmt.Println("Deployed database", res.TxResult.Events[0])
	d.logger.Debug("deploy database", zap.String("name", db.Name), zap.String("owner", db.Owner))
	return nil
}

func (d *KwildDriver) DatabaseShouldExists(ctx context.Context, owner string, dbName string) error {
	dbid := utils.GenerateDBID(dbName, owner)
	fmt.Println("Cherry: dbid", dbid)
	dbSchema, err := d.clt.GetSchema(ctx, dbid)
	if err != nil {
		return fmt.Errorf("failed to get database schema: %w", err)
	}
	fmt.Println("Cherry: dbSchema", dbSchema, err)
	if strings.EqualFold(dbSchema.Owner, owner) && strings.EqualFold(dbSchema.Name, dbName) {
		return nil
	}
	return fmt.Errorf("database does not exist")
}

func (d *KwildDriver) ExecuteAction(ctx context.Context, dbid string, actionName string, actionInputs []map[string]any) (*kTx.Receipt, []map[string]any, error) {
	rec, err := d.clt.ExecuteAction(ctx, dbid, actionName, actionInputs)
	if err != nil {
		fmt.Println("Error executing action: ", err.Error())
		return nil, nil, fmt.Errorf("error executing query: %w", err)
	}
	time.Sleep(15 * time.Second)

	res, err := d.bcClt.Tx(ctx, rec.TxHash, false)
	if err != nil {
		fmt.Println("Error getting transaction: ", err.Error())
		return nil, nil, fmt.Errorf("error getting transaction: %w", err)
	}

	data := res.TxResult.Data
	var updated_rec *kTx.Receipt
	err = json.Unmarshal(data, &updated_rec)
	if err != nil {
		return nil, nil, err
	}
	fmt.Println("Updated receipt: ", updated_rec)

	outputs, err := client.DecodeOutputs(updated_rec.Body)
	if err != nil {
		return nil, nil, err
	}
	fmt.Println("Outputs: ", outputs)
	//fmt.Println("Executed actions on the database", res.TxResult.Events[0])
	//events := res.Txs[numTx_post-1].TxResult.Events[0].Attributes
	d.logger.Debug("execute action", zap.String("database", dbid), zap.String("action", actionName))
	return rec, outputs, nil
}

func (d *KwildDriver) DropDatabase(ctx context.Context, dbName string) error {
	rec, err := d.clt.DropDatabase(ctx, dbName)
	if err != nil {
		return fmt.Errorf("error dropping database: %w", err)
	}
	time.Sleep(15 * time.Second)
	res, err := d.bcClt.Tx(ctx, rec.TxHash, false)
	if err != nil {
		fmt.Println("Error getting transaction: ", err.Error())
		return fmt.Errorf("error getting transaction: %w", err)
	}

	if !GetTransactionResult(res.TxResult.Events[0].Attributes) {
		return fmt.Errorf("failed to drop database")
	}

	fmt.Println("Dropped database", res.TxResult.Events[0].Attributes)
	d.logger.Debug("drop database", zap.String("name", dbName), zap.String("owner", d.GetUserAddress()))
	return nil
}

func (d *KwildDriver) QueryDatabase(ctx context.Context, dbid, query string) (*client.Records, error) {
	return d.clt.Query(ctx, dbid, query)
}

func GetTransactionResult(attributes []types.EventAttribute) bool {
	for _, attr := range attributes {
		if attr.Key == "Result" {
			return attr.Value == "Success"
		}
	}
	return false
}

// func (d *KwildDriver) ApproveNode(ctx context.Context, pubKey []byte) error {
// 	_, err := d.clt.ApproveValidator(ctx, "", "")
// 	return err
// }

// func (d *KwildDriver) ValidatorSetCount(ctx context.Context) (int, error) {
// 	vals, err := d.bcClt.Validators(ctx, nil, nil, nil)
// 	if err != nil {
// 		return -1, err
// 	}
// 	fmt.Println("ValidatorSet count: ", vals.Count)
// 	return vals.Count, nil
// }

// func (d *KwildDriver) ValidatorNodeJoin(ctx context.Context, pubKey []byte, power int64) error {
// 	_, err := d.clt.ValidatorJoin(ctx, pubKey, power)
// 	if err != nil {
// 		return fmt.Errorf("error joining validator: %w", err)
// 	}
// 	return nil
// }

// func (d *KwildDriver) ValidatorNodeLeave(ctx context.Context, pubKey []byte) error {
// 	rec, err := d.clt.ValidatorLeave(ctx, pubKey)
// 	if err != nil {
// 		return fmt.Errorf("error joining validator: %w", err)
// 	}

// 	time.Sleep(15 * time.Second)
// 	res, err := d.bcClt.Tx(ctx, rec.TxHash, false)
// 	if err != nil {
// 		fmt.Println("Error getting transaction: ", err.Error())
// 		return fmt.Errorf("error getting transaction: %w", err)
// 	}

// 	if !GetTransactionResult(res.TxResult.Events[0].Attributes) {
// 		return fmt.Errorf("failed to join as a validator")
// 	}

// 	fmt.Println("Join as Validator", res.TxResult.Events[0].Attributes)
// 	return nil
// }
