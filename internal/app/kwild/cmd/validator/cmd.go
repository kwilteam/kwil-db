package validator

import (
	"github.com/kwilteam/kwil-db/internal/app/kwild/config"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:     "validator",
		Aliases: []string{"val"},
		Short:   "manage validators",
		Long:    "Validator is a command that contains subcommands for handling the validators",
	}
)

func NewCmdValidator(cfg *config.KwildConfig) *cobra.Command {
	rootCmd.AddCommand(
		approveCmd(cfg),
		joinCmd(cfg),
		leaveCmd(cfg),
		statusCmd(cfg),
	)

	return rootCmd
}
