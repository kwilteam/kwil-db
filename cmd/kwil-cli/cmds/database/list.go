package database

import (
	"context"

	"github.com/kwilteam/kwil-db/cmd/internal/display"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/cmds/common"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/config"
	"github.com/spf13/cobra"
)

var (
	listLong = `List databases owned by a wallet.

An owner can be specified with the --owner flag. If no owner is specified, the locally configured wallet is used.`
	listExample = `# list databases owned by the wallet "0x9228624C3185FCBcf24c1c9dB76D8Bef5f5DAd64"
kwil-cli database list --owner 0x9228624C3185FCBcf24c1c9dB76D8Bef5f5DAd64`
)

func listCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "list",
		Short:        "List databases owned by a wallet.",
		Long:         listLong,
		Example:      listExample,
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return common.DialClient(cmd.Context(), cmd, common.WithoutPrivateKey, func(ctx context.Context, client common.Client, conf *config.KwilCliConfig) error {
				owner, err := getSelectedOwner(cmd, conf)
				if err != nil {
					return err
				}

				dbs, err := client.ListDatabases(ctx, owner)
				if err != nil {
					return err
				}

				resp := &respDBList{
					Databases: dbs,
					Owner:     owner,
				}

				return display.PrintCmd(cmd, resp)
			})
		},
	}

	cmd.Flags().StringP(ownerFlag, "o", "", "The owner of the database")
	return cmd
}
