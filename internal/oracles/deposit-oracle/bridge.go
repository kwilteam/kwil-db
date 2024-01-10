package deposit_oracle

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
	"github.com/kwilteam/kwil-db/extensions/oracles"
	"github.com/kwilteam/kwil-db/internal/sql"
	"github.com/kwilteam/kwil-db/internal/voting"
	"go.uber.org/zap"
)

const (
	oracleName = "deposit_oracle"

	depositEventSignature = "Credit(address,uint256)"

	contractABIStr = "[{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"_from\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"Credit\",\"type\":\"event\"}]"

	last_processed_block = "last_processed_block"
)

func init() {
	oracle := &DepositOracle{}

	// Register the oracle
	err := oracles.RegisterOracle(oracleName, oracle)
	if err != nil {
		fmt.Println("Failed to register oracle", zap.Error(err))
		panic(err)
	}

	// Register the AccountCredit payload with the Vote Processor
	payload := &AccountCredit{}
	err = voting.RegisterPaylod(payload)
	if err != nil {
		fmt.Println("Failed to register payload", zap.Error(err))
		panic(err)
	}
}

type DepositOracle struct {
	cfg                  DepositOracleConfig
	eventstore           oracles.EventStore
	kvstore              sql.KVStore
	creditEventSignature common.Hash
	eventABI             abi.ABI
	ethclient            *ethereumClient.Client
	logger               log.Logger
}

type DepositOracleConfig struct {
	endpoint              string
	chainID               string
	escrowAddress         string
	startingHeight        int64
	requiredConfirmations int64
	reconnectInterval     time.Duration
	maxTotalRequests      int64
}

func (do *DepositOracle) Start(ctx context.Context, eventstore oracles.EventStore, config map[string]string, logger log.Logger) error {
	do.logger = logger
	do.eventstore = eventstore
	do.kvstore = eventstore.KV([]byte(oracleName))

	if err := do.extractConfig(ctx, config); err != nil {
		return fmt.Errorf("failed to extract config: %w", err)
	}

	client, err := ethereumClient.DialContext(ctx, do.cfg.endpoint)
	if err != nil {
		return err
	}
	do.ethclient = client

	hash := crypto.Keccak256Hash([]byte(depositEventSignature))
	do.creditEventSignature = hash

	contractABI, err := abi.JSON(strings.NewReader(contractABIStr))
	if err != nil {
		return err
	}
	do.eventABI = contractABI

	// Start the listener
	return do.listen(ctx)
}

func (do *DepositOracle) Stop() error {
	do.ethclient.Close()
	return nil
}

func (do *DepositOracle) extractConfig(ctx context.Context, metadata map[string]string) error {
	// Endpoint
	if endpoint, ok := metadata["endpoint"]; ok {
		do.cfg.endpoint = endpoint
	} else {
		return fmt.Errorf("no endpoint provided")
	}

	// Escrow Address
	if escrowAddr, ok := metadata["escrow_address"]; ok {
		do.cfg.escrowAddress = escrowAddr
	} else {
		return fmt.Errorf("no escrow address provided")
	}

	// Chain ID
	if chainID, ok := metadata["chain_id"]; ok {
		do.cfg.chainID = chainID
	} else {
		return fmt.Errorf("no chain id provided")
	}

	// Required Confirmations
	if confirmations, ok := metadata["required_confirmations"]; ok {
		// convert confirmations to int64
		confirmations64, err := strconv.ParseInt(confirmations, 10, 64)
		if err != nil {
			return fmt.Errorf("failed to parse required confirmations: %w", err)
		}
		do.cfg.requiredConfirmations = confirmations64
	} else {
		do.cfg.requiredConfirmations = 12
	}

	// Starting Height
	if startingHeight, ok := metadata["starting_height"]; ok {
		// convert startingHeight to int64
		startingHeight64, err := strconv.ParseInt(startingHeight, 10, 64)
		if err != nil {
			return fmt.Errorf("failed to parse starting height: %w", err)
		}
		do.cfg.startingHeight = startingHeight64
	} else {
		do.cfg.startingHeight = 0
	}

	// Reconnect Interval
	if interval, ok := metadata["reconnect_interval"]; ok {
		// convert interval to float64
		interval64, err := strconv.ParseFloat(interval, 64)
		if err != nil {
			return fmt.Errorf("failed to parse reconnect interval: %w", err)
		}
		do.cfg.reconnectInterval = time.Duration(interval64) * time.Second
	}

	// MaxTotalRequests Size
	if maxTotalRequests, ok := metadata["max_total_requests"]; ok {
		// convert maxTotalRequests to int64
		maxTotalRequests64, err := strconv.ParseInt(maxTotalRequests, 10, 64)
		if err != nil {
			return fmt.Errorf("failed to parse max total requests: %w", err)
		}
		do.cfg.maxTotalRequests = maxTotalRequests64
	} else {
		do.cfg.maxTotalRequests = 1000
	}

	return nil
}
