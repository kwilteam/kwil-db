package ethbridge

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	ethereumClient "github.com/ethereum/go-ethereum/ethclient"
	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/internal/sql"
	"github.com/kwilteam/kwil-db/internal/voting"
	"github.com/kwilteam/kwil-db/oracles"
	"go.uber.org/zap"
)

const (
	oracleName = "ethBridge"

	depositEventSignature = "Credit(address,uint256)"

	contractABIStr = "[{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"_from\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"Credit\",\"type\":\"event\"}]"
)

// RegisterOracle registers the ethBridge oracle with the oracles package.
func init() {
	oracle := &EthBridge{}

	err := oracles.RegisterOracle(oracleName, oracle)
	if err != nil {
		fmt.Println("Failed to register oracle", zap.Error(err))
		panic(err)
	}

	payload := &AccountCredit{}
	err = voting.RegisterPaylod(payload)
	if err != nil {
		fmt.Println("Failed to register payload", zap.Error(err))
		panic(err)
	}
}

type EthBridge struct {
	cfg                  EthBridgeConfig
	eventstore           oracles.EventStore
	kvstore              sql.KVStore
	creditEventSignature common.Hash
	eventABI             abi.ABI
	ethclient            *ethereumClient.Client
	logger               log.Logger
}

type EthBridgeConfig struct {
	endpoint              string
	chainID               string
	escrowAddress         string
	startingHeight        int64
	requiredConfirmations int64
	ReconnectInterval     time.Duration
}

func (eb *EthBridge) Initialize(ctx context.Context, eventstore oracles.EventStore, config map[string]string, logger log.Logger) error {
	eb.logger = logger
	eb.eventstore = eventstore
	eb.kvstore = eventstore.KV([]byte(oracleName))

	if err := eb.extractConfig(ctx, config); err != nil {
		return fmt.Errorf("failed to extract config: %w", err)
	}

	client, err := ethereumClient.DialContext(ctx, eb.cfg.endpoint)
	if err != nil {
		return err
	}
	eb.ethclient = client

	hash := crypto.Keccak256Hash([]byte(depositEventSignature))
	eb.creditEventSignature = hash

	contractABI, err := abi.JSON(strings.NewReader(contractABIStr))
	if err != nil {
		panic(err)
	}
	eb.eventABI = contractABI

	return nil
}

func (o *EthBridge) Start(ctx context.Context) error {
	return o.listen(ctx)
}

func (o *EthBridge) Stop() error {
	return nil
}

func (tb *EthBridge) extractConfig(ctx context.Context, metadata map[string]string) error {
	// Endpoint, EscrowAddress, ChainCode
	if endpoint, ok := metadata["endpoint"]; ok {
		tb.cfg.endpoint = endpoint
	} else {
		return fmt.Errorf("no endpoint provided")
	}

	if escrowAddr, ok := metadata["escrow_address"]; ok {
		tb.cfg.escrowAddress = escrowAddr
	} else {
		return fmt.Errorf("no escrow address provided")
	}

	if chainID, ok := metadata["chain_id"]; ok {
		tb.cfg.chainID = chainID
	} else {
		return fmt.Errorf("no chain id provided")
	}

	if confirmations, ok := metadata["required_confirmations"]; ok {
		// convert confirmations to int64
		confirmations64, err := strconv.ParseInt(confirmations, 10, 64)
		if err != nil {
			return fmt.Errorf("failed to parse required confirmations: %w", err)
		}
		tb.cfg.requiredConfirmations = confirmations64
	} else {
		tb.cfg.requiredConfirmations = 12
	}

	if startingHeight, ok := metadata["starting_height"]; ok {
		// convert startingHeight to int64
		startingHeight64, err := strconv.ParseInt(startingHeight, 10, 64)
		if err != nil {
			return fmt.Errorf("failed to parse starting height: %w", err)
		}
		tb.cfg.startingHeight = startingHeight64
	} else {
		tb.cfg.startingHeight = 0
	}

	if interval, ok := metadata["reconnect_interval"]; ok {
		// convert interval to float64
		interval64, err := strconv.ParseFloat(interval, 64)
		if err != nil {
			return fmt.Errorf("failed to parse reconnect interval: %w", err)
		}
		tb.cfg.ReconnectInterval = time.Duration(interval64) * time.Second
	}

	return nil
}
