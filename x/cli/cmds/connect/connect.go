package connect

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	v0 "kwil/x/api/v0"
	"kwil/x/cli/util"
)

func NewCmdConnect() *cobra.Command {
	var save bool

	cmd := &cobra.Command{
		Use:   "connect",
		Short: "Connect tests a connection with the Kwil node.",
		Long: `Connect tests a connection with the specified Kwil node. It also
		exchanges relevant information regarding the node's capabilities, keys, etc..`,

		RunE: func(cmd *cobra.Command, args []string) error {
			return util.ConnectKwil(cmd.Context(), viper.GetViper(), func(ctx context.Context, client v0.KwilServiceClient) error {
				res, err := client.Connect(ctx, &v0.ConnectRequest{})
				if err != nil {
					return err
				}
				cmd.Println(res.Address)

				if save {
					return util.WriteConfig(map[string]any{"node-address": res.Address})
				}
				return nil
			})
		},
	}

	util.BindKwilFlags(cmd.PersistentFlags())
	cmd.Flags().BoolVar(&save, "save", false, "save the node address to the config file")

	return cmd
}
