package main

import (
	"context"
	"fmt"
	"github.com/kwilteam/kwil-db/internal/config"
	"github.com/kwilteam/kwil-db/internal/deposits"
	"github.com/kwilteam/kwil-db/internal/events"
	"github.com/kwilteam/kwil-db/internal/logging"
	"github.com/kwilteam/kwil-db/internal/processing"
	"github.com/kwilteam/kwil-db/internal/store"
	"github.com/rs/zerolog/log"
	"os"
	"os/signal"
)

func main() {

	ctx := context.Background()

	// Initialize build info
	err := config.InitBuildInfo()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize build info")
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

	// Connect to the client chain
	// First attempt to connect to client
	client, err := events.ConnectChain()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to client chain")
		os.Exit(1)
	}

	// Print that the node is running
	ef, err := events.New(&config.Conf, client)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize event feed")
		os.Exit(1)
	}

	// Get the event channcel from the eventfeed
	evChan, err := ef.Start(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to start event feed")
		os.Exit(1)
	}

	// Initialize KV store
	kv, err := store.New(&config.Conf)
	defer kv.Close()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize KV store")
		os.Exit(1)
	}
	// Initialize the deposit store
	ds := deposits.New(kv)

	// Initialize the event processor
	ep := processing.New(&config.Conf, evChan, ds)
	ep.ProcessEvents(ctx, evChan) // Listening

	// Making a channel listening for interruptions or errors
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, os.Kill)
	fmt.Println("Node is running properly!")
	// Block until a signal is received.
	sig := <-c
	fmt.Println("\nGot signal:", sig)
}
