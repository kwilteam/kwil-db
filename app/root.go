package app

import (
	"github.com/kwilteam/kwil-db/app/custom"
	"github.com/kwilteam/kwil-db/app/key"
	"github.com/kwilteam/kwil-db/app/node"
	"github.com/kwilteam/kwil-db/app/setup"
	"github.com/kwilteam/kwil-db/app/shared"
	"github.com/kwilteam/kwil-db/config"
	"github.com/kwilteam/kwil-db/version"

	"github.com/spf13/cobra"
)

func RootCmd() *cobra.Command {
	var rootDir string
	shared.BindDefaults(struct {
		RootDir        string `koanf:"root" toml:"root"`
		*config.Config `koanf:",flatten"`
	}{
		RootDir: ".testnet",
		Config:  shared.DefaultConfig(), // not config.DefaultConfig(), so custom command config is used
	}, "koanf")

	// TODO: update to use app/custom.BinaryConfig

	cmd := &cobra.Command{
		Use:               custom.BinaryConfig.NodeCmd,
		Short:             custom.BinaryConfig.ProjectName + " daemon",
		Long:              custom.BinaryConfig.ProjectName + " main application (node and utilities)",
		DisableAutoGenTag: true,
		SilenceUsage:      true,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
		Version: version.KwilVersion,
		Example: "kwild -r .testnet",
		// PersistentPreRunE so k has all the settings in all (sub)command's RunE funcs
		PersistentPreRunE: shared.ChainPreRuns(shared.MaybeEnableCLIDebug, shared.PreRunBindConfigFile,
			shared.PreRunBindFlags, shared.PreRunBindEnvMatching, shared.PreRunPrintEffectiveConfig),
	}

	shared.BindDebugFlag(cmd) // --debug enabled CLI debug mode (shared.Debugf output)

	shared.BindRootDirVar(cmd, &rootDir, ".testnet", "root directory") // --root/-r accessible with shared.RootDir from any subcommand

	cmd.AddCommand(node.StartCmd())
	cmd.AddCommand(setup.SetupCmd())
	cmd.AddCommand(key.KeyCmd())
	cmd.AddCommand(PrintConfigCmd())

	return cmd
}
