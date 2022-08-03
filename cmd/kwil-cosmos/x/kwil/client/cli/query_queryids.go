package cli

import (
	"context"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/kwilteam/kwil-db/cmd/kwil-cosmos/x/kwil/types"
	"github.com/spf13/cobra"
)

func CmdListQueryids() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-queryids",
		Short: "list all queryids",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryAllQueryidsRequest{
				Pagination: pageReq,
			}

			res, err := queryClient.QueryidsAll(context.Background(), params)
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

func CmdShowQueryids() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show-queryids [index]",
		Short: "shows a queryids",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			argIndex := args[0]

			params := &types.QueryGetQueryidsRequest{
				Index: argIndex,
			}

			res, err := queryClient.Queryids(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
