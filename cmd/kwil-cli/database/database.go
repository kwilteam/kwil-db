package database

import (
	"github.com/spf13/cobra"
	"kwil/cmd/kwil-cli/common"
)

var (
	rootCmd = &cobra.Command{
		Use:     "database",
		Aliases: []string{"db"},
		Short:   "manage databases",
		Long:    "Database is a command that contains subcommands for interacting with databases",
	}
	deployCmd = &cobra.Command{
		Use:   "deploy",
		Short: "Deploy a database",
		RunE:  cmdDeploy,
	}
	dropCmd = &cobra.Command{
		Use:   "drop",
		Short: "Drops a database",
		RunE:  cmdDrop,
	}
)

func NewCmdDatabase() *cobra.Command {
	deployCmd.Flags().StringP("path", "p", "", "Path to the database definition file")
	deployCmd.MarkFlagRequired("path")
	rootCmd.AddCommand(deployCmd)

	rootCmd.AddCommand(dropCmd)

	rootCmd.AddCommand(viewDatabaseCmd())

	common.BindKwilFlags(rootCmd)
	common.BindKwilEnv(rootCmd)
	common.BindChainFlags(rootCmd)
	common.BindChainEnv(rootCmd)

	return rootCmd
}
