package server

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/kwilteam/kwil-db/internal/app/kwild/config"

	// shorthand for chain client service

	"github.com/spf13/cobra"

	"github.com/kwilteam/kwil-db/internal/app/kwild/server"
)

func NewStartCmd() *cobra.Command {
	return startCmd
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "kwil grpc server",
	Long:  "Starts node with Kwild and CometBFT services",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		signalChan := make(chan os.Signal, 1)
		signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
		ctx, cancel := context.WithCancel(ctx)

		go func() {
			<-signalChan
			cancel()
		}()

		svr, err := server.BuildKwildServer(ctx)
		if err != nil {
			return err
		}

		return svr.Start(ctx)

	},
}

func init() {
	/*
		defaultConfigPath := filepath.Join("$HOME", config.DefaultConfigDir,
			fmt.Sprintf("%s.%s", config.DefaultConfigName, config.DefaultConfigType))
		RootCmd.PersistentFlags().StringVar(&config.ConfigFile, "config", "", fmt.Sprintf("config file to use (default: '%s')", defaultConfigPath))
	*/

	config.BindFlagsAndEnv(startCmd.PersistentFlags())
}
