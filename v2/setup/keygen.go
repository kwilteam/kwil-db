package main

import "github.com/spf13/cobra"

func keygenCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "keygen",
		Short: "Generate keys for testing purposes",
		Run: func(cmd *cobra.Command, args []string) {
			// Logic to generate keys
		},
	}
}
