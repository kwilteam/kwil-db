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
	"github.com/kwilteam/kwil-db/app/params"
	"github.com/kwilteam/kwil-db/app/rpc"
	"github.com/kwilteam/kwil-db/app/seed"
	"github.com/kwilteam/kwil-db/app/setup"
	"github.com/kwilteam/kwil-db/app/shared"
	"github.com/kwilteam/kwil-db/app/shared/bind"
	"github.com/kwilteam/kwil-db/app/shared/display"
	verCmd "github.com/kwilteam/kwil-db/app/shared/version"
	"github.com/kwilteam/kwil-db/app/snapshot"
	"github.com/kwilteam/kwil-db/app/utils"
	"github.com/kwilteam/kwil-db/app/validator"
	"github.com/kwilteam/kwil-db/app/whitelist"
	"github.com/kwilteam/kwil-db/config"
	_ "github.com/kwilteam/kwil-db/extensions" // a base location where all extensions can be registered
	_ "github.com/kwilteam/kwil-db/extensions/auth"
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
		// PersistentPreRunE so k has all the settings in all (sub)command's RunE funcs
		PersistentPreRunE: bind.ChainPreRuns(bind.MaybeEnableCLIDebug,
			conf.PreRunBindEarlyRootDirEnv, conf.PreRunBindEarlyRootDirFlag, // bind root from env and flag first
			conf.PreRunBindConfigFileStrict[config.Config], // then config file
			conf.PreRunBindFlags,                           // then flags
			conf.PreRunBindEnvMatching,                     // then env vars
			conf.PreRunPrintEffectiveConfig),
	}

	bind.BindDebugFlag(cmd) // --debug enabled CLI debug mode (shared.Debugf output)

	conf.BindDefaultsWithRootDir(custom.DefaultConfig(), defaultRoot)

	bind.BindRootDir(cmd, defaultRoot, "root directory") // --root/-r accessible with bind.RootDir from *any* subcommand

	display.BindOutputFormatFlag(cmd) // --output/-o

	// There is a virtual "node" command grouping, but no actual "node" command yet.
	cmd.AddCommand(node.StartCmd())       // needs merged config
	cmd.AddCommand(node.PrintConfigCmd()) // needs merged config

	// This group of command uses the merged config for fallback admin listen
	// addr if the --rpcserver flag is not set.
	cmd.AddCommand(rpc.NewAdminCmd())
	cmd.AddCommand(validator.NewValidatorsCmd())
	cmd.AddCommand(params.NewConsensusCmd())
	cmd.AddCommand(whitelist.WhitelistCmd())
	cmd.AddCommand(block.NewBlockExecCmd())
	cmd.AddCommand(migration.NewMigrationCmd())

	cmd.AddCommand(setup.SetupCmd()) // only kinda needs merged config for `setup reset`

	cmd.AddCommand(key.KeyCmd())
	cmd.AddCommand(snapshot.NewSnapshotCmd())

	cmd.AddCommand(seed.SeedCmd())
	cmd.AddCommand(utils.NewCmdUtils())
	cmd.AddCommand(verCmd.NewVersionCmd())

	// Apply the custom help function to the current command
	shared.ApplySanitizedHelpFuncRecursively(cmd)

	return cmd
}
