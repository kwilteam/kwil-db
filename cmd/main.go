package main

import (
	"context"
	"fmt"
	"github.com/kwilteam/kwil-db/internal/config"
	"github.com/kwilteam/kwil-db/internal/deposits"
	"github.com/kwilteam/kwil-db/internal/logging"
	"github.com/rs/zerolog/log"
	"os"
	"os/signal"
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

	// Initialize deposits
	d, err := deposits.Init(ctx, &config.Conf, client)
	defer d.Store.Close()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize deposits")
		os.Exit(1)
	}

	// Making a channel listening for interruptions or errors
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, os.Kill)
	fmt.Println("Node is running properly!")
	// Block until a signal is received.
	sig := <-c
	fmt.Println("\nGot signal:", sig)
}
