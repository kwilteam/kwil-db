package app

import (
	"os"
	"path/filepath"

	"github.com/kwilteam/kwil-db/app/block"
	"github.com/kwilteam/kwil-db/app/custom"
	"github.com/kwilteam/kwil-db/app/key"
	"github.com/kwilteam/kwil-db/app/migration"
	"github.com/kwilteam/kwil-db/app/node"
	"github.com/kwilteam/kwil-db/app/node/conf"
	"github.com/kwilteam/kwil-db/app/rpc"
	"github.com/kwilteam/kwil-db/app/seed"
	"github.com/kwilteam/kwil-db/app/setup"
	"github.com/kwilteam/kwil-db/app/shared/bind"
	"github.com/kwilteam/kwil-db/app/snapshot"
	"github.com/kwilteam/kwil-db/app/validator"
	"github.com/kwilteam/kwil-db/app/whitelist"
	"github.com/kwilteam/kwil-db/version"

	"github.com/spf13/cobra"
)

var defaultRoot = func() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".kwild")
}()

func RootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               custom.BinaryConfig.NodeCmd,
		Short:             custom.BinaryConfig.ProjectName + " daemon",
		Long:              custom.BinaryConfig.ProjectName + " node and utilities",
		DisableAutoGenTag: true,
		SilenceUsage:      true,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
		Version: version.KwilVersion,
		Example: custom.BinaryConfig.NodeCmd + " -r ~/.kwild",
		// PersistentPreRunE so k has all the settings in all (sub)command's RunE funcs
		PersistentPreRunE: bind.ChainPreRuns(bind.MaybeEnableCLIDebug, conf.PreRunBindConfigFile,
			conf.PreRunBindFlags, conf.PreRunBindEnvMatching, conf.PreRunPrintEffectiveConfig),
	}

	bind.BindDebugFlag(cmd) // --debug enabled CLI debug mode (shared.Debugf output)

	// conf.BindDefaults(struct {
	// 	RootDir        string `koanf:"root" toml:"root"`
	// 	*config.Config `koanf:",flatten"`
	// }{
	// 	RootDir: defaultRoot,
	// 	Config:  custom.DefaultConfig(), // not config.DefaultConfig(), so custom command config is used
	// })
	conf.BindDefaultsWithRootDir(custom.DefaultConfig(), defaultRoot)

	bind.BindRootDir(cmd, defaultRoot, "root directory") // --root/-r accessible with bind.RootDir from *any* subcommand

	// There is a virtual "node" command grouping, but no actual "node" command yet.
	cmd.AddCommand(node.StartCmd())
	cmd.AddCommand(node.PrintConfigCmd())
	cmd.AddCommand(rpc.NewAdminCmd())
	cmd.AddCommand(validator.NewValidatorsCmd())
	cmd.AddCommand(setup.SetupCmd())
	cmd.AddCommand(whitelist.WhitelistCmd())
	cmd.AddCommand(key.KeyCmd())
	cmd.AddCommand(snapshot.NewSnapshotCmd())
	cmd.AddCommand(migration.NewMigrationCmd())
	cmd.AddCommand(block.NewBlockExecCmd())
	cmd.AddCommand(seed.SeedCmd())

	return cmd
}
