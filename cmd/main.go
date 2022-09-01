package main

import (
	"context"
	"fmt"
	"os"

	"github.com/kwilteam/kwil-db/internal/api/rest"
	"github.com/kwilteam/kwil-db/internal/api/service"
	"github.com/kwilteam/kwil-db/internal/auth"
	"github.com/kwilteam/kwil-db/internal/config"
	"github.com/kwilteam/kwil-db/internal/crypto"
	"github.com/kwilteam/kwil-db/internal/deposits"
	"github.com/kwilteam/kwil-db/internal/logging"
	"github.com/rs/zerolog/log"
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
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize deposits")
		os.Exit(1)
	}

	defer d.Store.Close()

	// Creating Account
	kr, err := crypto.NewKeyring(&config.Conf)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create keyring")
		os.Exit(1)
	}
	acc, err := kr.GetDefaultAccount()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to get default account")
		os.Exit(1)
	}
	// Creating Authenticator
	a := auth.NewAuth(&config.Conf, acc)

	// Authenticate with peers
	a.Client.AuthAll()

	// Making a channel listening for interruptions or errors
	fmt.Println("Node is running properly!")

	// HTTP server
	serv := service.NewService(&config.Conf, d.Store)
	httpHandler := rest.NewHandler(*serv, a.Authenticator)
	if err := httpHandler.Serve(); err != nil {
		log.Fatal().Err(err).Msg("failed to start http server")
		os.Exit(1)
	}
}
