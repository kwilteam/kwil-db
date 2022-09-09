/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package fund

import (
	"github.com/kwilteam/kwil-db/cli/cmd"
	"github.com/spf13/cobra"
)

// fundCmd represents the fund command
var fundCmd = &cobra.Command{
	Use:   "fund",
	Short: "fund contains subcommands for funding",
	Long: `"fund" contains subcommands for funding.
	With "fund" you can deposit, withdraw, and check your allowance.`,
}

func init() {
	cmd.RootCmd.AddCommand(fundCmd)
}
