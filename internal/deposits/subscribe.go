package deposits

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/kwilteam/kwil-db/internal/config"
	"github.com/kwilteam/kwil-db/internal/logging"
	"github.com/rs/zerolog/log"
)

func Subscribe() error {
	// Get the config
	conf := &config.Conf

	// Dial the gateway service
	client, err := ethclient.Dial(conf.ClientChain.Endpoint)
	defer client.Close()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to client chain")
	}
	log.Printf("websocket connection established: %s", conf.ClientChain.Endpoint)

	// Find the deposit contract address
	contractAddr := common.HexToAddress(conf.ClientChain.DepositContract.Address)

	// Define the query to listen to
	query := ethereum.FilterQuery{
		Addresses: []common.Address{contractAddr},
	}

	// Channel for listening to ethereum events
	logs := make(chan types.Log)

	// Subscribe
	sub, err := client.SubscribeFilterLogs(context.Background(), query, logs)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to subscribe to client chain logs")
	}

	// Receive them here.  This loop will never end.
	for {
		select {
		case err := <-sub.Err():
			log.Fatal().Err(err).Msg("error on client chain log")
			logging.FileOutput("error reading chainlink")
		case v := <-logs:
			log.Printf("log: %+v", v)
			logging.FileOutput(string(v.Data))
			fmt.Println(v)
		}
	}
}
