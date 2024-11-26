package cmds

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/kwilteam/kwil-db/app/custom"
	"github.com/kwilteam/kwil-db/app/shared"
	"github.com/kwilteam/kwil-db/app/shared/display"
	"github.com/kwilteam/kwil-db/app/shared/version"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/cmds/account"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/cmds/configure"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/cmds/database"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/cmds/utils"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/config"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/helpers"
)

var longDesc = `Command line interface client for using %s.
	
` + "`" + `%s` + "`" + ` is a command line interface for interacting with %s. It can be used to deploy, update, and query databases.
	
` + "`" + `%s` + "`" + ` can be configured with a persistent configuration file. This file can be configured with the '%s configure' command.
` + "`" + `%s` + "`" + ` will look for a configuration file at ` + "`" + `$HOME/.kwil-cli/config.json` + "`" + `.`

func NewRootCmd() *cobra.Command {
	// The basic for ActiveConfig starts with defaults defined in DefaultKwilCliPersistedConfig.
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
		PersistentPreRunE: shared.ChainPreRuns(shared.MaybeEnableCLIDebug,
			config.PreRunBindFlags, config.PreRunBindConfigFile,
			config.PreRunPrintEffectiveConfig),
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
	}

	// Define the --debug enabled CLI debug mode (shared.Debugf output)
	shared.BindDebugFlag(rootCmd)

	// Bind the --config flag, which informs PreRunBindConfigFile, as well as
	// PersistConfig and LoadPersistedConfig.
	config.BindConfigPath(rootCmd)

	// Automatically define flags for all of the fields of the config struct.
	config.SetFlags(rootCmd.Flags())

	helpers.BindAssumeYesFlag(rootCmd) // --assume-yes/-Y

	display.BindOutputFormatFlag(rootCmd) // --output
	display.BindSilenceFlag(rootCmd)      // --silence/-S

	rootCmd.AddCommand(
		account.NewCmdAccount(),
		configure.NewCmdConfigure(),
		database.NewCmdDatabase(),
		utils.NewCmdUtils(),
		version.NewVersionCmd(),
	)

	return rootCmd
}
