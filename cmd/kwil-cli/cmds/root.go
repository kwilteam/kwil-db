package cmds

import (
	"fmt"

	"github.com/kwilteam/kwil-db/app/custom"
	"github.com/kwilteam/kwil-db/app/shared/bind"
	"github.com/kwilteam/kwil-db/app/shared/display"
	"github.com/kwilteam/kwil-db/app/shared/version"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/cmds/account"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/cmds/configure"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/cmds/database"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/cmds/utils"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/config"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/helpers"
	"github.com/spf13/cobra"
)

var longDesc = `Command line interface client for using %s.
	
` + "`" + `%s` + "`" + ` is a command line interface for interacting with %s. It can be used to deploy, update, and query databases.
	
` + "`" + `%s` + "`" + ` can be configured with a persistent configuration file. This file can be configured with the '%s configure' command.
` + "`" + `%s` + "`" + ` will look for a configuration file at ` + "`" + `$HOME/.kwil-cli/config.json` + "`" + `.`

func NewRootCmd() *cobra.Command {
	// The basis for ActiveConfig starts with defaults defined in DefaultKwilCliPersistedConfig.
	if err := config.BindDefaults(); err != nil {
		panic(err)
	}

	rootCmd := &cobra.Command{
		Use:   custom.BinaryConfig.ClientCmd,
		Short: fmt.Sprintf("Command line interface client for using %s.", custom.BinaryConfig.ProjectName),
		Long: fmt.Sprintf(longDesc, custom.BinaryConfig.ProjectName, custom.BinaryConfig.ClientUsage(),
			custom.BinaryConfig.ProjectName, custom.BinaryConfig.ClientUsage(), custom.BinaryConfig.ClientUsage(), custom.BinaryConfig.ClientUsage()),
		SilenceUsage:      true,
		DisableAutoGenTag: true,
		PersistentPreRunE: bind.ChainPreRuns(bind.MaybeEnableCLIDebug,
			// Config priority, highest to lowest: env, flags, config.json
			config.PreRunBindConfigFile, config.PreRunBindFlags, config.PreRunBindEnv,
			config.PreRunPrintEffectiveConfig),
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
	}

	// Define the --debug enabled CLI debug mode (shared.Debugf output)
	bind.BindDebugFlag(rootCmd)

	// Bind the --config flag, which informs PreRunBindConfigFile, as well as
	// PersistConfig and LoadPersistedConfig.
	config.BindConfigPath(rootCmd)

	// Automatically define flags for all of the fields of the config struct.
	config.SetFlags(rootCmd.PersistentFlags()) // share configs with all subcommands

	helpers.BindAssumeYesFlag(rootCmd) // --assume-yes/-Y

	display.BindOutputFormatFlag(rootCmd) // --output
	display.BindSilenceFlag(rootCmd)      // --silence/-S

	rootCmd.AddCommand(
		account.NewCmdAccount(),
		configure.NewCmdConfigure(),
		database.NewCmdDatabase(),
		utils.NewCmdUtils(),
		version.NewVersionCmd(),
		execSQLCmd(),
		execActionCmd(),
		callActionCmd(),
		queryCmd(),
	)

	return rootCmd
}
