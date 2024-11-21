package app

import (
	"fmt"

	"github.com/kwilteam/kwil-db/config"
	"github.com/kwilteam/kwil-db/version"

	"github.com/knadh/koanf/v2"
	gotoml "github.com/pelletier/go-toml/v2"
	"github.com/spf13/cobra"
)

var k = koanf.New(".")

const RootFlagName = "root"

func RootCmd() *cobra.Command {
	var rootDir string
	BindDefaults(struct {
		RootDir        string `koanf:"root" toml:"root"`
		*config.Config `koanf:",flatten"`
	}{
		RootDir: ".testnet",
		Config:  config.DefaultConfig(),
	}, "koanf")

	// TODO: update to use app/custom.BinaryConfig

	cmd := &cobra.Command{
		Use:               "kwil",
		Short:             "kwil v2 node app",
		Long:              "kwil main application (node and utilities)",
		DisableAutoGenTag: true,
		SilenceUsage:      true,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
		Version: version.KwilVersion,
		Example: "kwil -r .testnet",
		// PersistentPreRunE so k has all the settings in all (sub)command's RunE funcs
		PersistentPreRunE: ChainPreRuns(maybeEnableCLIDebug, PreRunBindConfigFile,
			PreRunBindFlags, PreRunBindEnvMatching, PreRunPrintEffectiveConfig),
	}

	// --debug enabled CLI debug mode (debugf output)
	cmd.PersistentFlags().Bool("debug", false, "enable debugging, will print debug logs")
	cmd.Flag("debug").Hidden = true

	// "root" does not have config file analog, so binds with local var, which
	// is then available to all subcommands via RootDir(cmd).
	cmd.PersistentFlags().StringVarP(&rootDir, RootFlagName, "r", ".testnet", "root directory")

	cmd.AddCommand(StartCmd()) // default command
	cmd.AddCommand(SetupCmd())
	cmd.AddCommand(PrintConfigCmd())

	return cmd
}

func maybeEnableCLIDebug(cmd *cobra.Command, args []string) error {
	debugFlag := cmd.Flag("debug")
	if !debugFlag.Changed {
		return nil
	}
	debug, err := cmd.Flags().GetBool("debug")
	if err != nil {
		return err
	}
	if debug {
		enableCLIDebugging()
		k.Set("log_level", "debug")
	}
	return nil
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
			rootDir, err := RootDir(cmd)
			if err != nil {
				return err // the parent command needs to set a persistent flag named "root"
			}

			// k => config.Config
			var cfg config.Config
			err = k.UnmarshalWithConf("", &cfg, koanf.UnmarshalConf{Tag: "koanf"})
			if err != nil {
				return fmt.Errorf("failed to unmarshal config: %w", err)
			}

			debugf("effective node config (toml):\n%s", lazyPrinter(func() string {
				rawToml, err := gotoml.Marshal(&cfg)
				if err != nil {
					return fmt.Errorf("failed to marshal config to toml: %w", err).Error()
				}
				return string(rawToml)
			}))

			return runNode(cmd.Context(), rootDir, &cfg)
		},
	}

	// Other node flags have config file and env analogs, and will be loaded
	// into koanf where the values are merged.
	// SetNodeFlags(cmd)
	defaultCfg := config.DefaultConfig()
	SetNodeFlagsFromStruct(cmd, defaultCfg)

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
