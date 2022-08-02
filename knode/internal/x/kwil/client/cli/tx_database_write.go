package cli

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/kwilteam/kwil-db/knode/internal/x/kwil/types"
	"github.com/spf13/cobra"
)

var _ = strconv.Itoa(0)

func CmdDatabaseWrite() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "database-write [database] [par-quer] [data]",
		Short: "Broadcast message DatabaseWrite",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argDatabase := args[0]
			argParQuer := args[1]
			argData := args[2]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgDatabaseWrite(
				clientCtx.GetFromAddress().String(),
				argDatabase,
				argParQuer,
				argData,
			)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
