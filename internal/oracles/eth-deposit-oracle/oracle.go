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
	"github.com/kwilteam/kwil-db/internal/events"
	"github.com/kwilteam/kwil-db/internal/voting"
)

const (
	oracleName = "eth_deposit_oracle"

	depositEventSignature = "Credit(address,uint256)"

	contractABIStr = `[{"anonymous":false,"inputs":[{"indexed":false,"internalType":"address","name":"_from","type":"address"},{"indexed":false,"internalType":"uint256","name":"_amount","type":"uint256"}],"name":"Credit","type":"event"}]`

	last_processed_block = "last_processed_block"
)

func init() {
	oracle := &EthDepositOracle{}

	// Register the oracle
	err := oracles.RegisterOracle(oracleName, oracle)
	if err != nil {
		panic(err)
	}

	// Register the AccountCredit payload with the Vote Processor
	payload := &AccountCredit{}
	err = voting.RegisterPayload(payload)
	if err != nil {
		panic(err)
	}
}

// EthDepositOracle listens for credit events on the escrow contract
// deployed on the ethereum blockchain and stores them in the eventstore.
// Requires ethereum endpoint, escrow address and chain id to be provided
type EthDepositOracle struct {
	cfg                  EthDepositOracleConfig
	eventstore           oracles.EventStore
	kvstore              *events.KV
	creditEventSignature common.Hash
	eventABI             abi.ABI
	ethclient            *ethereumClient.Client
	logger               log.Logger

	done chan bool
}

type EthDepositOracleConfig struct {
	endpoint      string
	chainID       string
	escrowAddress string
	// startingHeight is the block height to start processing events from.
	// Especially useful when node is catching up and don't want to process old events.
	startingHeight        int64
	requiredConfirmations int64
	reconnectInterval     time.Duration
	maxTotalRequests      int64
	maxRetries            uint64
}

func (do *EthDepositOracle) Start(ctx context.Context, eventstore oracles.EventStore, config map[string]string, logger log.Logger) error {
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

	do.done = make(chan bool)

	return do.listen(ctx)
}

func (do *EthDepositOracle) Stop() error {
	if do.ethclient != nil {
		do.ethclient.Close()
	}
	// Signal to the listener to stop
	do.done <- true
	return nil
}

func (do *EthDepositOracle) extractConfig(ctx context.Context, metadata map[string]string) error {
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

	// MaxRetries
	if maxRetries, ok := metadata["max_retries"]; ok {
		// convert maxRetries to uint64
		maxRetries64, err := strconv.ParseUint(maxRetries, 10, 64)
		if err != nil {
			return fmt.Errorf("failed to parse max retries: %w", err)
		}
		do.cfg.maxRetries = maxRetries64
	} else {
		do.cfg.maxRetries = 50
	}

	return nil
}
