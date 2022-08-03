package cli

import (
	"strconv"

	"github.com/kwilteam/kwil-db/cmd/kwil-cosmos/x/kwil/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cast"
	"github.com/spf13/cobra"
)

var _ = strconv.Itoa(0)

func CmdDefineQuery() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "define-query [db-id] [par-quer] [publicity]",
		Short: "Broadcast message DefineQuery",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argDbid := args[0]
			argParQuer := args[1]
			argPublicity, err := cast.ToBoolE(args[2])
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgDefineQuery(
				clientCtx.GetFromAddress().String(),
				argDbid,
				argParQuer,
				argPublicity,
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
