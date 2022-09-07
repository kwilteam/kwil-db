package config

import (
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rs/zerolog/log"
)

func ConnectChain(c *Config) (*ethclient.Client, error) {

	// Dial the gateway service
	client, err := ethclient.Dial(c.ClientChain.Endpoint)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to client chain")
		return client, err
	}
	log.Printf("websocket connection established: %s", c.ClientChain.Endpoint)
	return client, err
}
