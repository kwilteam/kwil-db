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
<<<<<<< HEAD
	"github.com/kwilteam/kwil-db/internal/utils/files"
	"github.com/kwilteam/kwil-db/pkg/pricing"
=======
>>>>>>> b64dc94cf02f1f9d814336627f167ff5d29bb7d5
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

	ctx := context.Background()
	log.Debug().Msg("debug turned on")

	// Connect to the client chain
	// First attempt to connect to client
<<<<<<< HEAD
	client, err := config.ConnectChain(conf)
=======
	client, err := config.ConnectChain()
>>>>>>> b64dc94cf02f1f9d814336627f167ff5d29bb7d5
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to client chain")
		os.Exit(1)
	}

	// Initialize deposits
<<<<<<< HEAD
	d, err := deposits.Init(ctx, conf, client)
=======
	d, err := deposits.Init(ctx, &config.Conf, client)
>>>>>>> b64dc94cf02f1f9d814336627f167ff5d29bb7d5
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize deposits")
		os.Exit(1)
	}
<<<<<<< HEAD

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
	pbytes, err := files.LoadFileFromRoot(ppath)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load pricing config")
		os.Exit(1)
	}

	// Initialize pricing
	p, err := pricing.New(pbytes)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize pricing")
=======

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
>>>>>>> b64dc94cf02f1f9d814336627f167ff5d29bb7d5
		os.Exit(1)
	}
	// Creating Authenticator
	a := auth.NewAuth(&config.Conf, acc)

<<<<<<< HEAD
=======
	// Authenticate with peers
	a.Client.AuthAll()

>>>>>>> b64dc94cf02f1f9d814336627f167ff5d29bb7d5
	// Making a channel listening for interruptions or errors
	fmt.Println("Node is running properly!")

	// HTTP server
<<<<<<< HEAD
	serv := service.NewService(d.Store, p)
=======
	serv := service.NewService(&config.Conf, d.Store)
>>>>>>> b64dc94cf02f1f9d814336627f167ff5d29bb7d5
	httpHandler := rest.NewHandler(*serv, a.Authenticator)
	if err := httpHandler.Serve(); err != nil {
		log.Fatal().Err(err).Msg("failed to start http server")
		os.Exit(1)
	}
}
