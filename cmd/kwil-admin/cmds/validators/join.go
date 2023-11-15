package validators

import (
	"context"

	"github.com/kwilteam/kwil-db/cmd/internal/display"
	"github.com/kwilteam/kwil-db/cmd/kwil-admin/cmds/common"
	"github.com/spf13/cobra"
)

var (
	joinLong = "A node may request to join the validator set by submitting a join request using the `join` command. The key used to sign the join request will be the treated as the node request to join the validator set. The node will be added to the validator set if the join request is approved by the current validator set. The status of a join request can be queried using the `join-status` command."

	joinExample = `$ kwil-admin validators join --key-file "~/.kwild/private_key"
Joining the network as a validator...
Node PublicKey: d692a88b6fee8d1399f7aab70db25001080dbf2fe8ca4f345e0fdca6b853713a
tx hash: 125ae9009a5056a152542aa84e97a2a49c547e492dffdfea50eb44623fbb3efb
tx sender: 1pKoi2/ujROZ96q3DbJQAQgNvy/oyk80Xg/cprhTcTo=
tx signature (ed25519): ce56440b1ecd529b5345a6662f3a30831b7d85ddc714036c4e7b124dd0481429478c1459c1833a8e9fa338329b0f217ca04c4381bb53679e3d72dec3317e7b09
tx payload (validator_join): 0001e2a0d692a88b6fee8d1399f7aab70db25001080dbf2fe8ca4f345e0fdca6b853713a01`
)

func joinCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "join",
		Short:   "A node may request to join the validator set by submitting a join request using the `join` command.",
		Long:    joinLong,
		Example: joinExample,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			clt, err := common.GetAdminSvcClient(ctx, cmd)
			if err != nil {
				return err
			}

			txHash, err := clt.Join(ctx)
			if err != nil {
				return err
			}

			return display.PrintCmd(cmd, display.RespTxHash(txHash))
		},
	}

	return cmd
}
