package config

import (
	"fmt"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rs/zerolog/log"
)

func ConnectChain() (*ethclient.Client, error) {
	// Get the config
	conf, err := LoadConfig("kwil_config.json")
	if err != nil {
		return nil, err
	}

	// Dial the gateway service
	client, err := ethclient.Dial(conf.ClientChain.Endpoint)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to client chain")
		return client, err
	}
	log.Printf("websocket connection established: %s", conf.ClientChain.Endpoint)
	fmt.Printf("websocket connection established: %s\n", conf.ClientChain.Endpoint)
	return client, err
}
