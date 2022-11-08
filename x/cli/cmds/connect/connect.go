package connect

import (
	"context"

	"kwil/x/cli/util"
	"kwil/x/proto/apipb"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

func NewCmdConnect() *cobra.Command {
	var save bool

	cmd := &cobra.Command{
		Use:   "connect",
		Short: "Connect tests a connection with the Kwil node.",
		Long: `Connect tests a connection with the specified Kwil node. It also
		exchanges relevant information regarding the node's capabilities, keys, etc..`,

		RunE: func(cmd *cobra.Command, args []string) error {
			return util.ConnectKwil(cmd.Context(), viper.GetViper(), func(ctx context.Context, cc *grpc.ClientConn) error {
				client := apipb.NewKwilServiceClient(cc)
				res, err := client.Connect(ctx, &apipb.ConnectRequest{})
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
