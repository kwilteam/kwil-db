package connect

import (
	"kwil/cmd/kwil-cli/util"

	"github.com/spf13/cobra"
)

func NewCmdConnect() *cobra.Command {
	var save bool

	cmd := &cobra.Command{
		Use:   "connect",
		Short: "Connect tests a connection with the Kwil node.",
		Long: `Connect tests a connection with the specified Kwil node. It also
		exchanges relevant information regarding the node's capabilities, keys, etc..`,

		RunE: func(cmd *cobra.Command, args []string) error {
			/*return util.ConnectKwil(cmd.Context(), viper.GetViper(), func(ctx context.Context, cc *grpc.ClientConn) error {
				client := apipb.NewKwilServiceClient(cc)

				res, err := client.Connect(ctx, &apipb.ConnectRequest{})
				if err != nil {
					return err
				}
				color.Set(color.FgGreen)
				cmd.Println("Connection successful ")
				color.Unset()

				if save {
					return util.WriteConfig(map[string]any{"node-address": res.Address})
				}
				return nil
			})*/
			return nil
		},
	}

	util.BindKwilFlags(cmd.PersistentFlags())
	cmd.Flags().BoolVar(&save, "save", false, "save the node address to the config file")

	return cmd
}
