package database

import (
	"github.com/spf13/cobra"
)

var (
	dbCmd = &cobra.Command{
		Use:     "database",
		Aliases: []string{"db"},
		Short:   "The database command is a parent command containing subcommands for interacting with databases.",
		Long:    "The database command is a parent command containing subcommands for interacting with databases.",
	}

	nonceOverride int64
	syncBcast     bool
)

func NewCmdDatabase() *cobra.Command {
	// readOnlyCmds do not create a transaction.
	readOnlyCmds := []*cobra.Command{
		listCmd(),
		readSchemaCmd(),
		queryCmd(),
		callCmd(), // no tx, but may required key for signature, for now
	}
	dbCmd.AddCommand(readOnlyCmds...)

	// writeCmds create a transactions, requiring a private key for signing/
	writeCmds := []*cobra.Command{
		deployCmd(),
		dropCmd(),
		executeCmd(),
		batchCmd(),
	}
	dbCmd.AddCommand(writeCmds...)

	// The write commands may also specify a nonce to use instead of asking the
	// node for the latest confirmed nonce.
	for _, cmd := range writeCmds {
		cmd.Flags().Int64VarP(&nonceOverride, "nonce", "N", -1, "nonce override (-1 means request from server)")
		cmd.Flags().BoolVar(&syncBcast, "sync", false, "synchronous broadcast (wait for it to be included in a block)")
	}

	return dbCmd
}
