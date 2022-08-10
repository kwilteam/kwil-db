package deposits

import (
	"fmt"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/kwilteam/kwil-db/internal/config"
	"github.com/rs/zerolog/log"
)

func DryConnectChain() error {
	// Get the config
	conf := &config.Conf

	// Dial the gateway service
	client, err := ethclient.Dial(conf.ClientChain.Endpoint)
	defer client.Close()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to client chain")
		return err
	}
	log.Printf("websocket connection established: %s", conf.ClientChain.Endpoint)
	fmt.Printf("websocket connection established: %s\n", conf.ClientChain.Endpoint)
	return nil
}

func ConnectChain() (*ethclient.Client, error) {
	// Get the config
	conf := &config.Conf

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
