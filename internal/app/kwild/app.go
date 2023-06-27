package kwild

import (
	"github.com/kwilteam/kwil-db/internal/app/kwild/cmd"
	"github.com/kwilteam/kwil-db/internal/app/kwild/cmd/validator"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:               "kwild",
	Short:             "kwild command line interface",
	Long:              "kwild allows you to configure Kwild services",
	SilenceUsage:      true,
	DisableAutoGenTag: true,
}

func Execute() error {
	rootCmd.AddCommand(
		validator.NewCmdValidator(),
		cmd.NewStartCmd(),
	)
	return rootCmd.Execute()
}
