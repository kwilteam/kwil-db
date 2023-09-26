package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/kwilteam/kwil-db/internal/app/kwild/config"
	"github.com/kwilteam/kwil-db/internal/app/kwild/server"
	"github.com/kwilteam/kwil-db/internal/pkg/version"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	kwildCfg = config.DefaultConfig()
)

func main() {
	rootCmd.Version = version.KwilVersion

	flagSet := rootCmd.Flags()
	flagSet.SortFlags = false
	addKwildFlags(flagSet, kwildCfg)
	viper.BindPFlags(flagSet)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	os.Exit(0)
}

var rootCmd = &cobra.Command{
	Use:               "kwild",
	Short:             "kwild node and rpc server",
	Long:              "kwild: the Kwil blockchain node and RPC server",
	DisableAutoGenTag: true,
	Args:              cobra.NoArgs, // just flags
	RunE: func(cmd *cobra.Command, args []string) error {
		// command line flags are now (finally) parsed by cobra. Bind env vars,
		// load the config file, and unmarshal into kwildCfg.
		if err := kwildCfg.LoadKwildConfig(); err != nil {
			return fmt.Errorf("failed to load kwild config: %w", err)
		}

		if err := kwildCfg.InitPrivateKeyAndGenesis(); err != nil {
			return fmt.Errorf("failed to initialize private key and genesis: %w", err)
		}

		signalChan := make(chan os.Signal, 1)
		signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
		ctx, cancel := context.WithCancel(cmd.Context())

		go func() {
			<-signalChan
			cancel()
		}()

		svr, err := server.New(ctx, kwildCfg)
		if err != nil {
			return err
		}

		return svr.Start(ctx)
	},
}
