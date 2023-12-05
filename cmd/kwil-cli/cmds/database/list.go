package database

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/kwilteam/kwil-db/cmd/common/display"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/cmds/common"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/config"
	"github.com/kwilteam/kwil-db/core/crypto/auth"
	"github.com/spf13/cobra"
)

var (
	listLong = `List databases owned by a wallet.

An owner can be specified with the ` + "`" + `--owner` + "`" + ` flag. If no owner is specified, then it will return all databases deployed on the network.
If the ` + "`" + `--self` + "`" + ` flag is specified, then the owner will be set to the current configured wallet.`

	listExample = `# list databases owned by the wallet "0x9228624C3185FCBcf24c1c9dB76D8Bef5f5DAd64"
kwil-cli database list --owner 0x9228624C3185FCBcf24c1c9dB76D8Bef5f5DAd64

# list all databases deployed on the network
kwil-cli database list

# list databases owned by the current configured wallet
kwil-cli database list --self`
)

func listCmd() *cobra.Command {
	var owner string
	var self bool

	cmd := &cobra.Command{
		Use:          "list",
		Short:        "List databases owned by a wallet.",
		Long:         listLong,
		Example:      listExample,
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return common.DialClient(cmd.Context(), cmd, common.WithoutPrivateKey, func(ctx context.Context, client common.Client, conf *config.KwilCliConfig) error {
				if owner != "" && self {
					return display.PrintErr(cmd, errors.New("cannot specify both --owner and --self"))
				}

				var ownerIdent []byte
				if self {
					if conf.PrivateKey == nil {
						return display.PrintErr(cmd, errors.New("must have a configured wallet to use --self"))
					}
					ownerIdent = (&auth.EthPersonalSigner{Key: *conf.PrivateKey}).Identity()
				} else if owner != "" {
					var err error
					ownerIdent, err = hex.DecodeString(owner)
					if err != nil {
						return display.PrintErr(cmd, fmt.Errorf("failed to decode hex owner: %w", err))
					}
				}

				dbs, err := client.ListDatabases(ctx, ownerIdent)
				if err != nil {
					return display.PrintErr(cmd, err)
				}

				return display.PrintCmd(cmd, &respDBList{
					Info:  dbs,
					owner: ownerIdent,
				})
			})
		},
	}

	cmd.Flags().StringVarP(&owner, ownerFlag, "o", "", "the owner of the database")
	cmd.Flags().BoolVar(&self, "self", false, "use the current configured wallet as the owner")

	return cmd
}
