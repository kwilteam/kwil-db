package main

import (
	"context"
	"fmt"
	"os"

	"github.com/kwilteam/kwil-db/internal/api/rest"
	"github.com/kwilteam/kwil-db/internal/api/service"
	"github.com/kwilteam/kwil-db/internal/chain/auth"
	"github.com/kwilteam/kwil-db/internal/chain/config"
	"github.com/kwilteam/kwil-db/internal/chain/crypto"
	"github.com/kwilteam/kwil-db/internal/chain/deposits"
	"github.com/kwilteam/kwil-db/internal/chain/utils"
	"github.com/kwilteam/kwil-db/internal/common/logging"
	"github.com/kwilteam/kwil-db/pkg/types/chain/pricing"
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
	conf, err := config.LoadConfig("kwil_config.json")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
		os.Exit(1)
	}

	// Initialize the global logger
	logging.InitLogger(config.BuildInfo.Version, conf.Log.Debug, conf.Log.Human)

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
	d, err := deposits.Init(ctx, conf, client)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize deposits")
		os.Exit(1)
	}

	defer d.Store.Close()

	// Creating Account
	kr, err := crypto.NewKeyring(conf)
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
	a := auth.NewAuth(conf, acc)

	// Authenticate with peers
	a.Client.AuthAll()

	// Get the pricing config as bytes
	ppath := conf.GetPricePath()
	pbytes, err := utils.LoadFileFromRoot(ppath)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load pricing config")
		os.Exit(1)
	}

	// Initialize pricing
	p, err := pricing.New(pbytes)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize pricing")
		os.Exit(1)
	}
	// Creating Authenticator
	ath := auth.NewAuth(conf, acc)

	// Making a channel listening for interruptions or errors
	fmt.Println("Node is running properly!")

	// HTTP server
	serv := service.NewService(d.Store, p)
	httpHandler := rest.NewHandler(*serv, ath.Authenticator)
	if err := httpHandler.Serve(); err != nil {
		log.Fatal().Err(err).Msg("failed to start http server")
		os.Exit(1)
	}
}
