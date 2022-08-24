package main

import (
	"context"
	"fmt"
	"github.com/kwilteam/kwil-db/internal/api/rest"
	"github.com/kwilteam/kwil-db/internal/api/service"
	cosClient "github.com/kwilteam/kwil-db/internal/client"
	"github.com/kwilteam/kwil-db/internal/config"
	"github.com/kwilteam/kwil-db/internal/deposits"
	"github.com/kwilteam/kwil-db/internal/logging"
	"github.com/rs/zerolog/log"
	"os"
)

func main() {
	// Initialize build info
	err := config.InitBuildInfo()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize build info")
		os.Exit(1)
	}

	// Load Config
	err = config.LoadConfig(".")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
		os.Exit(1)
	}

	// Initialize the global logger
	logging.InitLogger(config.BuildInfo.Version, config.Conf.Log.Debug, config.Conf.Log.Human)

	ctx := context.Background()
	log.Debug().Msg("debug turned on")

	// Connect to the client chain
	// First attempt to connect to client
	client, err := config.ConnectChain()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to client chain")
		os.Exit(1)
	}

	/*// Create cosmos client
	cosmClient, err := cosClient.NewCosmosClient(ctx, &config.Conf)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create cosmos client")
		os.Exit(1)
	}

	cosmClient.Transfer(2, "kaddr-1jz2z9jtpza7a499cj4dpfmvzclwa0a5hva9ymq")*/

	// Initialize deposits
	cosmClient := cosClient.CosmosClient{}
	d, err := deposits.Init(ctx, &config.Conf, client, &cosmClient)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize deposits")
		os.Exit(1)
	}

	defer d.Store.Close()

	// Making a channel listening for interruptions or errors
	fmt.Println("Node is running properly!")

	// HTTP server
	serv := service.NewService(&config.Conf, d.Store)
	httpHandler := rest.NewHandler(*serv)
	if err := httpHandler.Serve(); err != nil {
		log.Fatal().Err(err).Msg("failed to start http server")
		os.Exit(1)
	}
}
