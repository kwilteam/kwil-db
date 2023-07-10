package server

import (
	"github.com/spf13/cobra"
)

var (
	serverCmd = &cobra.Command{
		Use:   "server",
		Short: "Kwild Server commands",
		Long:  "Commands to run Kwild Server",
	}
)

func NewServerCmd() *cobra.Command {
	serverCmd.AddCommand(
		NewStartCmd(),
	)

	return serverCmd
}
