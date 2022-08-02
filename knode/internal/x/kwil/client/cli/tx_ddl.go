package cli

import (
	"strconv"

	"github.com/kwilteam/kwil-db/knode/internal/x/kwil/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"
)

var _ = strconv.Itoa(0)

func CmdDDL() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ddl [ddl]",
		Short: "Broadcast message DDL",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argDbid := args[0]
			argDdl := args[1]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgDDL(
				clientCtx.GetFromAddress().String(),
				argDbid,
				argDdl,
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
