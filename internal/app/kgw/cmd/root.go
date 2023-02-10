package cmd

import (
	"fmt"
	"kwil/internal/app/kgw/config"
	"kwil/internal/app/kgw/server"
	"kwil/internal/pkg/gateway/middleware/cors"
	"kwil/pkg/log"
	"path/filepath"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "kwil-gateway",
	Short: "gateway to kwil-gateway service",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		cfg, err := config.LoadConfig()
		if err != nil {
			return err
		}

		logger := log.New(cfg.Log)

		gw := server.NewGWServer(runtime.NewServeMux(), *cfg, logger)

		if err := gw.SetupGrpcSvc(ctx); err != nil {
			return err
		}
		if err := gw.SetupHTTPSvc(ctx); err != nil {
			return err
		}

		gw.AddMiddlewares(
			// from innermost middleware
			//auth.MAuth(keyManager, logger),
			cors.MCors(cfg.Server.Cors),
		)

		return gw.Serve()
	},
}

func init() {
	defaultConfigPath := filepath.Join("$HOME", config.DefaultConfigDir,
		fmt.Sprintf("%s.%s", config.DefaultConfigName, config.DefaultConfigType))
	RootCmd.PersistentFlags().StringVar(&config.ConfigFile, "config", "", fmt.Sprintf("config file to use (default: '%s')", defaultConfigPath))

	config.BindGlobalFlags(RootCmd.PersistentFlags())
	config.BindGlobalEnv(RootCmd.PersistentFlags())
}
