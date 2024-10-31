package app

import (
	"fmt"

	"kwil/log"
	"kwil/version"

	// "github.com/knadh/koanf/parsers/json"

	"github.com/knadh/koanf/v2"
	"github.com/spf13/cobra"
)

var k = koanf.New(".")

const RootFlagName = "root"

func RootCmd() *cobra.Command {
	var rootDir string

	cmd := &cobra.Command{
		Use:               "kwil",
		Short:             "kwil v2 node app",
		Long:              "kwil main application (node and utilities)",
		DisableAutoGenTag: true,
		SilenceUsage:      true,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
		Version:           version.KwilVersion,
		Example:           "kwil -r .testnet",
		PersistentPreRunE: ChainPreRuns(PreRunBindConfigFile, PreRunBindFlags), // now k has all the settings in all (sub)command's RunE funcs
	}

	// "root" does not have config file analog, so binds with local var
	cmd.PersistentFlags().StringVarP(&rootDir, RootFlagName, "r", ".testnet", "root directory")

	cmd.AddCommand(StartCmd()) // default command
	cmd.AddCommand(SetupCmd())

	return cmd
}

func RootDir(cmd *cobra.Command) (string, error) {
	// fmt.Println("app.k's root:", k.String(RootFlagName))
	return cmd.Flags().GetString(RootFlagName)
}

func StartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "start",
		Short:             "start kwil node (default command)",
		Long:              "Start the v2 kwild node running",
		DisableAutoGenTag: true,
		SilenceUsage:      true,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
		Version: version.KwilVersion,
		Example: "kwil start -r .testnet",
		RunE: func(cmd *cobra.Command, args []string) error {
			logLevel, err := log.ParseLevel(k.String("log-level"))
			if err != nil {
				return fmt.Errorf("invalid log level: %w", err)
			}

			logFormat, err := log.ParseFormat(k.String("log-format"))
			if err != nil {
				return fmt.Errorf("invalid log format: %w", err)
			}
			rootDir, err := RootDir(cmd)
			if err != nil {
				return err // the parent command needs to set a persistent flag named "root"
			}
			return runNode(cmd.Context(), rootDir, logLevel, logFormat)
		},
	}

	// Other node flags have config file and env analogs, and will be loaded
	// into koanf where the values are merged.
	cmd.Flags().String("log-level", log.LevelInfo.String(), "log level")
	cmd.Flags().String("log-format", string(log.FormatUnstructured), "log format")

	cmd.SetVersionTemplate("kwil {{printf \"version %s\" .Version}}\n")

	return cmd
}

func SetupCmd() *cobra.Command {
	const setupLong = `The setup command provides functions for creating and managing node configuration and data, including:
	- performing quick setup of a standalone Kwil node (init) and Kwil test networks (testnet)
	- resetting node state and all data files (reset)`
	setupCmd := &cobra.Command{
		Use:   "setup",
		Short: "The setup command provides functions for creating and managing node configuration and data.",
		Long:  setupLong,
	}
	setupCmd.AddCommand(ResetCmd(), TestnetCmd(), KeyCmd())

	return setupCmd
}
