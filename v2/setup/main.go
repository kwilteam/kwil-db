package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func setupCmd() *cobra.Command {
	desc := "Setup is a tool to generate configuration for multiple nodes and generate keys for testing purposes"
	cmd := &cobra.Command{
		Use:               "setup",
		Short:             desc,
		Long:              desc,
		SilenceUsage:      true,
		DisableAutoGenTag: true,
	}

	cmd.AddCommand(testnetCmd())
	cmd.AddCommand(keygenCmd())
	cmd.AddCommand(resetCmd())

	return cmd
}

func main() {
	if err := setupCmd().Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
