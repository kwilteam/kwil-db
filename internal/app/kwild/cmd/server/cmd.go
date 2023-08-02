package server

import (
	"github.com/spf13/cobra"
)

var (
	serverCmd = &cobra.Command{
		Use:   "start",
		Short: "kwil grpc server",
		Long:  "Starts node with Kwild and CometBFT services",
	}
)

func NewServerCmd() *cobra.Command {
	serverCmd.AddCommand(
		NewStartCmd(),
	)

	return serverCmd
}
