package server

import (
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
