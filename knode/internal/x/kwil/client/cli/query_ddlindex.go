package cli

import (
	"context"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/kwilteam/kwil-db/knode/internal/x/kwil/types"
	"github.com/spf13/cobra"
)

func CmdListDdlindex() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-ddlindex",
		Short: "list all ddlindex",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryAllDdlindexRequest{
				Pagination: pageReq,
			}

			res, err := queryClient.DdlindexAll(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddPaginationFlagsToCmd(cmd, cmd.Use)
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdShowDdlindex() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show-ddlindex [index]",
		Short: "shows a ddlindex",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			argIndex := args[0]

			params := &types.QueryGetDdlindexRequest{
				Index: argIndex,
			}

			res, err := queryClient.Ddlindex(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
