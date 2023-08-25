package server

import (
	"github.com/kwilteam/kwil-db/internal/app/kwild/config"
	"github.com/spf13/cobra"
)

var (
	serverCmd = &cobra.Command{
		Use:   "server",
		Short: "Kwild Server commands",
		Long:  "Commands to run Kwild Server",
	}
)

func NewServerCmd(cfg *config.KwildConfig) *cobra.Command {
	serverCmd.AddCommand(
		NewStartCmd(cfg),
	)

	return serverCmd
}
