package cli

import (
	"context"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/kwilteam/kwil-db/knode/internal/x/kwil/types"
	"github.com/spf13/cobra"
)

func CmdListDatabases() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-databases",
		Short: "list all databases",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryAllDatabasesRequest{
				Pagination: pageReq,
			}

			res, err := queryClient.DatabasesAll(context.Background(), params)
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

func CmdShowDatabases() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show-databases [index]",
		Short: "shows a databases",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			argIndex := args[0]

			params := &types.QueryGetDatabasesRequest{
				Index: argIndex,
			}

			res, err := queryClient.Databases(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
